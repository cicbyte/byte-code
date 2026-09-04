package aiengine

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== eino Agent 执行器 ====================

// TaskContext 传给 AI 的任务上下文
type TaskContext struct {
	Id        int
	ProjectId int
	Title     string
	Desc      string
}

// ExecuteConfig 执行配置
type ExecuteConfig struct {
	BaseURL  string
	ApiKey   string
	Model    string
	AIUserId int
	Task     TaskContext
}

const systemPrompt = `你是 ByteCode 项目管理平台的 AI 助手。你已被分配了一个任务，请分析并执行它。

工作流（记忆/文档中枢）：
1. 开工先取上下文：mem_list 扫描项目记忆（约定/偏好/结论），kb_get_conventions 读已发布知识库，doc_linked 拉取与本任务关联的文档（targetType=task, targetId=任务ID）
2. 调用 get_project / list_tasks 了解项目与任务状况
3. 需要更多资料时 search_docs 搜索 + read_doc 读正文
4. 过程中发现的稳定结论：mem_set 沉淀为项目记忆（推测未确认用 status=pending）；确认某条记忆仍正确时 mem_verify 保鲜
5. 完成后输出工作结果（简洁报告，包含关键发现和建议）

记忆使用注意：mem_get/mem_list 返回的 hint 提示（过期/腐化/待验证）必须考虑，过期与腐化记忆不要作为依据。

输出格式：
## 分析结果
（对任务的理解和分析）

## 执行内容
（你做了什么、查了什么、发现了什么、沉淀了哪些记忆）

## 结论
（最终结论或建议）`

// ExecuteTask 执行单个任务（ReAct 循环）
func ExecuteTask(ctx context.Context, cfg *ExecuteConfig) (string, error) {
	// AI 身份与任务项目归属注入 ctx：mem_set/mem_verify 的 updated_by 来源标识；
	// projectId 是安全边界——工具层强制取 ctx 值，忽略 LLM 参数传入的任意 projectId，
	// 防止任务描述里诱导 AI 读写其他项目的 vault/记忆（跨项目数据外泄/投毒）
	ctx = context.WithValue(ctx, ctxAIUserId, cfg.AIUserId)
	ctx = context.WithValue(ctx, ctxTaskProjectId, cfg.Task.ProjectId)

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.ApiKey,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return "", fmt.Errorf("创建模型客户端失败: %v", err)
	}

	// 绑定工具
	toolInfos := make([]*schema.ToolInfo, 0)
	for _, t := range GetTools() {
		info, err := t.Info(ctx)
		if err != nil {
			return "", fmt.Errorf("获取工具信息失败: %v", err)
		}
		toolInfos = append(toolInfos, info)
	}
	boundModel, err := chatModel.WithTools(toolInfos)
		if err != nil {
			return "", fmt.Errorf("绑定工具失败: %v", err)
		}

	// 构造消息
	taskMsg := fmt.Sprintf(
		"请处理以下任务：\n\n任务 ID: %d\n项目 ID: %d\n标题: %s\n描述: %s",
		cfg.Task.Id, cfg.Task.ProjectId, cfg.Task.Title, defaultStr(cfg.Task.Desc, "无描述"),
	)
	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: taskMsg},
	}

	// ReAct 循环
	const maxIterations = 10
	// 上下文治理：超过 keepRecentToolResults 轮之前的工具结果替换为占位符。
	// read_doc/kb 工具单次可达 8KB，10 轮迭代累积轻松几十 K token（成本翻倍且
	// 逼近上下文窗口），而 ReAct 决策只依赖近期结果
	const keepRecentToolResults = 2
	tools := GetTools()

	for i := 0; i < maxIterations; i++ {
		resp, err := boundModel.Generate(ctx, messages)
		if err != nil {
			return "", fmt.Errorf("模型调用失败(iter %d): %v", i, err)
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		messages = append(messages, resp)

		for _, tc := range resp.ToolCalls {
			g.Log().Debugf(ctx, "AI 工具调用: %s(%s)", tc.Function.Name, tc.Function.Arguments)
			result := executeToolCall(ctx, tools, tc)
			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// 本轮结束后，把 keepRecentToolResults 轮之前的工具结果替换为截断占位
		evictOldToolResults(messages, keepRecentToolResults)
	}

	return "", fmt.Errorf("AI 执行超过最大迭代次数 %d", maxIterations)
}

// evictOldToolResults 就地截断旧工具结果：从后往前保留最近 keepN 轮，
// 更早的 Tool 消息内容替换为占位（保留 ToolCallID 关联，不破坏消息结构）
func evictOldToolResults(messages []*schema.Message, keepN int) {
	// 从尾部数 Tool 消息所在的“轮”——同一轮多个工具结果视为一批
	cutoff := len(messages) // 此下标之后的 Tool 消息保留
	seenRounds := 0
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i] != nil && messages[i].Role == schema.Tool {
			seenRounds++
			if seenRounds > keepN {
				cutoff = i + 1 // 含首个超限轮自身
				break
			}
		}
	}
	if cutoff == len(messages) {
		return
	}
	for i := 0; i < cutoff; i++ {
		if messages[i] != nil && messages[i].Role == schema.Tool {
			messages[i].Content = "[older tool result truncated for context budget]"
		}
	}
}

// executeToolCall 执行工具调用并返回结果字符串
func executeToolCall(ctx context.Context, tools []tool.InvokableTool, tc schema.ToolCall) string {
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil || info.Name != tc.Function.Name {
			continue
		}
		result, err := t.InvokableRun(ctx, tc.Function.Arguments)
		if err != nil {
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}
		return result
	}
	return fmt.Sprintf(`{"error":"tool %s not found"}`, tc.Function.Name)
}

// GetModel 返回 ChatModel 实例（供测试/调试）
func GetModel(ctx context.Context, cfg *ExecuteConfig) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.ApiKey,
		BaseURL: cfg.BaseURL,
	})
}

func defaultStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
