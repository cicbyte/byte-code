package aiengine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/frame/g"
)

// 类型别名：vaulttools.go 中新工具的签名缩写
type (
	ToolInfo      = schema.ToolInfo
	ParameterInfo = schema.ParameterInfo
	ToolOption    = tool.Option
)

// ==================== ByteCode 工具定义 ====================
// 将后端业务操作封装为 eino Tool，供 AI Agent 调用

func paramInfo(typ, desc string) *schema.ParameterInfo {
	return &schema.ParameterInfo{Type: schema.DataType(typ), Desc: desc}
}

func toolInfo(name, desc string, required map[string]*schema.ParameterInfo, optional map[string]*schema.ParameterInfo) *schema.ToolInfo {
	all := make(map[string]*schema.ParameterInfo)
	for k, v := range required {
		all[k] = v
	}
	for k, v := range optional {
		all[k] = v
	}
	reqNames := make([]string, 0, len(required))
	for k := range required {
		reqNames = append(reqNames, k)
	}
	return &schema.ToolInfo{
		Name:        name,
		Desc:        desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(all),
	}
}

// ---- 查询任务列表 ----

type ListTasksArgs struct {
	ProjectId int    `json:"projectId"`
	Status    string `json:"status,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
}

type ListTasksTool struct{}

func (t *ListTasksTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return toolInfo("list_tasks", "查询项目下的任务列表，可按状态和关键词过滤",
		map[string]*schema.ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
		},
		map[string]*schema.ParameterInfo{
			"status":  paramInfo("string", "状态过滤(open/in_progress/review/done/closed)"),
			"keyword": paramInfo("string", "标题关键词"),
		},
	), nil
}

func (t *ListTasksTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...tool.Option) (string, error) {
	var p ListTasksArgs
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	m := g.DB().Model("tasks t").Ctx(ctx).
		LeftJoin("sys_users u", "t.assignee_id = u.id").
		Fields("t.id, t.title, t.status, t.priority, COALESCE(u.real_name, u.username, '') as assignee, t.created_at").
		Where("t.project_id", p.ProjectId)

	if p.Status != "" {
		m = m.Where("t.status", p.Status)
	}
	if p.Keyword != "" {
		m = m.Where("t.title LIKE ?", "%"+p.Keyword+"%")
	}

	var tasks []map[string]interface{}
	err := m.Order("t.id DESC").Limit(50).Scan(&tasks)
	if err != nil {
		return "", fmt.Errorf("查询失败: %v", err)
	}

	b, _ := json.Marshal(tasks)
	return string(b), nil
}

// ---- 查询项目信息 ----

type GetProjectTool struct{}

func (t *GetProjectTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return toolInfo("get_project", "获取项目基本信息（名称/描述/状态/创建者）",
		map[string]*schema.ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
		}, nil,
	), nil
}

func (t *GetProjectTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...tool.Option) (string, error) {
	var p struct{ ProjectId int }
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	record, err := g.DB().Model("projects p").Ctx(ctx).
		LeftJoin("sys_users u", "p.created_by = u.id").
		Fields("p.id, p.name, p.description, p.status, COALESCE(u.real_name, u.username, '') as creator").
		Where("p.id", p.ProjectId).One()
	if err != nil || record.IsEmpty() {
		return "", fmt.Errorf("项目不存在")
	}

	b, _ := json.Marshal(record.Map())
	return string(b), nil
}

// ---- 搜索项目文档（vault 索引） ----

type SearchDocsTool struct{}

func (t *SearchDocsTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return toolInfo("search_docs", "在项目文档（磁盘 vault 索引）中搜索标题/标签/路径，返回 path 列表后用 read_doc 读正文",
		map[string]*schema.ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
			"keyword":   paramInfo("string", "搜索关键词"),
		}, nil,
	), nil
}

func (t *SearchDocsTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...tool.Option) (string, error) {
	var p struct {
		ProjectId int
		Keyword   string
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	like := "%" + p.Keyword + "%"
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Fields("path, title, space, type, tags").
		Where("project_id", p.ProjectId).
		Where("title LIKE ? OR tags LIKE ? OR path LIKE ?", like, like, like).
		Order("space DESC, updated_at DESC"). // knowledge 优先
		Limit(20).All()
	if err != nil {
		return "", fmt.Errorf("搜索失败: %v", err)
	}
	if len(rows) == 0 {
		return marshalString(g.Map{"items": []string{}, "hint": "未搜到文档，可换关键词"}), nil
	}
	b, _ := json.Marshal(rows)
	return string(b), nil
}

// GetTools 返回所有可用工具（实现 tool.InvokableTool 接口）。
// 第三组为记忆/文档中枢工具：AI 开工先读记忆与知识库，产出后沉淀记忆
func GetTools() []tool.InvokableTool {
	return []tool.InvokableTool{
		&ListTasksTool{},
		&GetProjectTool{},
		&SearchDocsTool{},
		&ReadDocTool{},
		&KbConventionsTool{},
		&DocLinkedTool{},
		&MemListTool{},
		&MemGetTool{},
		&MemSetTool{},
		&MemVerifyTool{},
	}
}
