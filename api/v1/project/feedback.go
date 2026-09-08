// Package api project 跨项目反馈：A 投递线索给 B，B 的 agent 自主决定建不建任务。
package project

import "github.com/gogf/gf/v2/frame/g"

type FeedbackCreateReq struct {
	g.Meta       `path:"/projects/{projectId}/feedbacks" method:"post" tags:"跨项目反馈" summary:"投递反馈（需目标项目已关联本账号可访问的项目）"`
	ProjectId    int    `json:"projectId" v:"required" in:"path" dc:"目标（接收方）项目 id"`
	Title        string `json:"title" v:"required|max-length:255#反馈标题不能为空|上限255字"`
	Content      string `json:"content" dc:"markdown：现象/线索/怀疑点"`
	SourceTaskId int `json:"sourceTaskId" dc:"来源任务 id（可空，血缘可溯）"`
	// 显式来源项目（可空）：多项目 agent 场景下推导链取最早一条 binding
	// 可能不准，显式指定优先于一切推导（bcode-cli 反馈建议）
	SourceProjectId int `json:"sourceProjectId" dc:"显式来源项目 id（可空；优先于任务/成员/bindings 推导）"`
}

type FeedbackCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type FeedbackListReq struct {
	g.Meta    `path:"/projects/{projectId}/feedbacks" method:"get" tags:"跨项目反馈" summary:"反馈收件箱（默认 open；all=全部）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Status    string `json:"status" in:"query" d:"open"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"20"`
}

type FeedbackItem struct {
	Id               int    `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	Status           string `json:"status"`
	SourceProjectId  int    `json:"sourceProjectId"`
	SourceProjectName string `json:"sourceProjectName"`
	SourceTaskId     int    `json:"sourceTaskId"`
	ConvertedTaskId  int    `json:"convertedTaskId"`
	DismissReason    string `json:"dismissReason"`
	HandledBy        int    `json:"handledBy"`
	CreatedAt        string `json:"createdAt"`
}

type FeedbackListRes struct {
	g.Meta `mime:"application/json"`
	List   []FeedbackItem `json:"list"`
	Total  int            `json:"total"`
}

type FeedbackConvertReq struct {
	g.Meta    `path:"/projects/{projectId}/feedbacks/{id}/convert" method:"post" tags:"跨项目反馈" summary:"反馈转任务（B 侧成员/agent；建任务并回填 converted_task_id）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Id        int    `json:"id" v:"required" in:"path"`
	Title     string `json:"title" dc:"任务标题（缺省用反馈标题）"`
	SprintId  int    `json:"sprintId"`
}

type FeedbackConvertRes struct {
	g.Meta `mime:"application/json"`
	TaskId int `json:"taskId"`
}

type FeedbackDismissReq struct {
	g.Meta    `path:"/projects/{projectId}/feedbacks/{id}/dismiss" method:"post" tags:"跨项目反馈" summary:"忽略反馈（必填理由，回告发起方）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Id        int    `json:"id" v:"required" in:"path"`
	Reason    string `json:"reason" v:"required#忽略理由不能为空"`
}

type FeedbackDismissRes struct {
	g.Meta `mime:"application/json"`
}
