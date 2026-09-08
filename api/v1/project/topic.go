// Package api project 专题（long-task）：专注一个 topic 的长时间自动执行工程。
package project

import "github.com/gogf/gf/v2/frame/g"

type TopicCreateReq struct {
	g.Meta     `path:"/projects/{projectId}/topics" method:"post" tags:"专题" summary:"创建专题（goal+验收+可选 PRD 文档关联+执行 agent）"`
	ProjectId  int    `json:"projectId" v:"required" in:"path"`
	Title      string `json:"title" v:"required|max-length:255#专题标题不能为空|上限255字"`
	Goal       string `json:"goal" dc:"markdown：目标与范围"`
	Acceptance string `json:"acceptance" dc:"完成判据（人验收依据）"`
	DocPath    string `json:"docPath" dc:"关联文档（通常是 PRD）：项目文档库内的 path"`
	AssigneeId int    `json:"assigneeId" dc:"执行 agent（一期单 agent）"`
}

type TopicCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type TopicListReq struct {
	g.Meta    `path:"/projects/{projectId}/topics" method:"get" tags:"专题" summary:"专题列表（默认 active；all=全部）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Status    string `json:"status" in:"query" d:"active"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"20" v:"max:100#单页上限100"`
}

type TopicPhaseBrief struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status string `json:"status"`
	// 迷你任务结构：负责人 + 产出（专题自闭环）
	AssigneeId   int    `json:"assigneeId"`
	AssigneeName string `json:"assigneeName"`
	Artifacts    string `json:"artifacts"`
	TaskId       int    `json:"taskId"`
	// 历史转出任务的实况（转任务已下线，仅存血缘展示）
	TaskTitle    string `json:"taskTitle"`
	TaskStatus   string `json:"taskStatus"`
	TaskAssignee string `json:"taskAssignee"`
	SortOrder    int    `json:"sortOrder"`
}

type TopicItem struct {
	Id           int               `json:"id"`
	Title        string            `json:"title"`
	Goal         string            `json:"goal"`
	Acceptance   string            `json:"acceptance"`
	DocPath      string            `json:"docPath"`
	AssigneeId   int               `json:"assigneeId"`
	AssigneeName string            `json:"assigneeName"`
	Status       string            `json:"status"`
	CreatedAt    string            `json:"createdAt"`
	CompletedAt  string            `json:"completedAt"`
	Phases       []TopicPhaseBrief `json:"phases"`
	PhaseTotal   int               `json:"phaseTotal"`
	PhaseDone    int               `json:"phaseDone"`
	LastHandoff  string            `json:"lastHandoff" dc:"最近一次交接摘要（下个会话恢复点）"`
}

type TopicListRes struct {
	g.Meta `mime:"application/json"`
	List   []TopicItem `json:"list"`
	Total  int         `json:"total"`
	Page   int         `json:"page"`
	Size   int         `json:"size"`
}

type TopicDetailReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{id}" method:"get" tags:"专题" summary:"专题详情（含阶段与最近 handoff）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	Id        int `json:"id" v:"required" in:"path"`
}

type TopicDetailRes struct {
	g.Meta `mime:"application/json"`
	TopicItem
}

type TopicPhaseUpsertReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{id}/phases" method:"post" tags:"专题" summary:"批量写入阶段（整体替换；agent 从 PRD 拆解后导入）"`
	ProjectId int            `json:"projectId" v:"required" in:"path"`
	Id        int            `json:"id" v:"required" in:"path"`
	Phases    []TopicPhaseIn `json:"phases" v:"required#阶段不能为空"`
}

type TopicPhaseIn struct {
	Title  string `json:"title" v:"required#阶段标题不能为空"`
	Detail string `json:"detail"`
}

type TopicPhaseUpsertRes struct {
	g.Meta `mime:"application/json"`
}

type TopicPhaseAddReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{id}/phases/add" method:"post" tags:"专题" summary:"追加单个阶段（手动补充清单）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Id        int    `json:"id" v:"required" in:"path"`
	Title     string `json:"title" v:"required|max-length:255#阶段标题不能为空|上限255字"`
	Detail    string `json:"detail"`
}

type TopicPhaseAddRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type TopicPhaseUpdateReq struct {
	g.Meta     `path:"/projects/{projectId}/topics/{topicId}/phases/{phaseId}" method:"put" tags:"专题" summary:"编辑阶段（标题/描述/负责人/产出）"`
	ProjectId  int     `json:"projectId" v:"required" in:"path"`
	TopicId    int     `json:"topicId" v:"required" in:"path"`
	PhaseId    int     `json:"phaseId" v:"required" in:"path"`
	Title      *string `json:"title"`
	Detail     *string `json:"detail"`
	AssigneeId *int    `json:"assigneeId"`
	Artifacts  *string `json:"artifacts"`
}

type TopicPhaseUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TopicPhaseDeleteReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{topicId}/phases/{phaseId}" method:"delete" tags:"专题" summary:"删除阶段（已转出任务的历史阶段保血缘不可删）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	TopicId   int `json:"topicId" v:"required" in:"path"`
	PhaseId   int `json:"phaseId" v:"required" in:"path"`
}

type TopicPhaseDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type TopicPhaseToggleReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{topicId}/phases/{phaseId}/toggle" method:"post" tags:"专题" summary:"阶段状态推进（pending→in_progress→done；agent/成员）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	TopicId   int    `json:"topicId" v:"required" in:"path"`
	PhaseId   int    `json:"phaseId" v:"required" in:"path"`
	Status    string `json:"status" v:"required|in:pending,in_progress,done#状态不能为空|状态不合法"`
}

type TopicPhaseToggleRes struct {
	g.Meta `mime:"application/json"`
}

type TopicPhaseConvertReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{topicId}/phases/{phaseId}/convert" method:"post" tags:"专题" summary:"阶段转日常任务（落任务池可被认领，反向关联）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	TopicId   int `json:"topicId" v:"required" in:"path"`
	PhaseId   int `json:"phaseId" v:"required" in:"path"`
}

type TopicPhaseConvertRes struct {
	g.Meta `mime:"application/json"`
	TaskId int `json:"taskId"`
}

type TopicLogReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{id}/log" method:"post" tags:"专题" summary:"专题留痕（执行进展 / handoff 交接摘要）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Id        int    `json:"id" v:"required" in:"path"`
	Action    string `json:"action" v:"required|in:progress,handoff#动作不能为空|仅支持 progress/handoff"`
	Detail    string `json:"detail" v:"required#内容不能为空"`
}

type TopicLogRes struct {
	g.Meta `mime:"application/json"`
}

type TopicFinishReq struct {
	g.Meta    `path:"/projects/{projectId}/topics/{id}/finish" method:"post" tags:"专题" summary:"终验收（人；completed 或 abandoned）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Id        int    `json:"id" v:"required" in:"path"`
	Result    string `json:"result" v:"required|in:completed,abandoned#结论不能为空|仅支持 completed/abandoned"`
}

type TopicFinishRes struct {
	g.Meta `mime:"application/json"`
}
