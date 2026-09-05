// Package api agent 定义外部 Agent 接入协议端点（dev-docs/agent-protocol.md）：
// 注册（公开）/ 项目接入 / 工作会话 / 免参任务与文档别名
package agent

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 注册（公开端点：纯身份，无任何权限） ====================

type RegisterReq struct {
	g.Meta `path:"/agent/register" method:"post" tags:"Agent接入" summary:"Agent 自助注册（纯身份标识，需项目接入码才可获得权限）"`
	Name   string `json:"name" v:"required|length:2,64#名称不能为空|名称长度2-64" dc:"全局唯一，如 claude-code@zhang-dev"`
	// 与现有 AiUserCreateReq.Capabilities 同语义：逗号分隔能力声明
	Capabilities string `json:"capabilities" dc:"能力声明（可选）"`
}

type RegisterRes struct {
	g.Meta `mime:"application/json"`
	AgentId int    `json:"agentId"`
	ApiKey  string `json:"apiKey" dc:"bc_ 前缀，只此一次返回，客户端自行保存"`
}

// ==================== 项目接入 ====================

type JoinCodeCreateReq struct {
	g.Meta `path:"/projects/{projectId}/agent-codes" method:"post" tags:"Agent接入" summary:"生成项目接入码（owner/超管；一次性/24h）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type JoinCodeCreateRes struct {
	g.Meta `mime:"application/json"`
	Code string `json:"code" dc:"bcg_ 前缀，展示一次即焚"`
	ExpiresAt string `json:"expiresAt"`
}

type JoinReq struct {
	g.Meta `path:"/agent/projects/join" method:"post" tags:"Agent接入" summary:"Agent 加入项目（bc key 认证 + 接入码）"`
	Code   string `json:"code" v:"required#接入码不能为空"`
}

type JoinRes struct {
	g.Meta `mime:"application/json"`
	ProjectId int    `json:"projectId"`
	ProjectName string `json:"projectName"`
}

// ==================== 工作会话 ====================

type SessionCreateReq struct {
	g.Meta `path:"/agent/sessions" method:"post" tags:"Agent接入" summary:"建立工作会话（开工包；键=agent+project，复用续期）"`
	ProjectId int `json:"projectId" v:"required#项目ID不能为空" dc:"来自 join 返回或 .bc/project 指向"`
}

type SessionCreateRes struct {
	g.Meta `mime:"application/json"`
	SessionId string `json:"sessionId"`
	Project   ProjectBrief `json:"project"`
	// 开工包：全局+项目记忆（conventions.* 优先），与 kb_get_conventions 同口径
	Conventions    []ConventionItem `json:"conventions"`
	MyTasks        []TaskBrief      `json:"myTasks"`
	PendingReviews []TaskBrief      `json:"pendingReviews"`
}

type ProjectBrief struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ConventionItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Scope string `json:"scope" dc:"global/project"`
}

type TaskBrief struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
	DueDate  string `json:"dueDate,omitempty"`
}

// ==================== 免参别名（X-Session 推导 agent+project） ====================

type AgentTasksReq struct {
	g.Meta `path:"/agent/tasks" method:"get" tags:"Agent接入" summary:"项目任务列表（会话推导，免 projectId）"`
	Status string `json:"status" in:"query" dc:"缺省=未完成三态；all=全部"`
	Keyword string `json:"keyword" in:"query"`
}

type AgentTasksRes struct {
	g.Meta `mime:"application/json"`
	List  []TaskBrief `json:"list"`
	Total int         `json:"total"`
}
