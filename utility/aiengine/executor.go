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

规则：
1. 首先调用 get_project 了解项目信息
2. 调用 list_tasks 了解项目当前任务状况
3. 如需查阅参考资料，调用 search_docs 搜索项目文档
4. 完成分析后，输出你的工作结果（简洁报告，包含关键发现和建议）

输出格式：
## 分析结果
（对任务的理解和分析）

## 执行内容
（你做了什么、查了什么、发现了什么）

## 结论
（最终结论或建议）`

// ExecuteTask 执行单个任务（ReAct 循环）
func ExecuteTask(ctx context.Context, cfg *ExecuteConfig) (string, error) {
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
	}

	return "", fmt.Errorf("AI 执行超过最大迭代次数 %d", maxIterations)
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
