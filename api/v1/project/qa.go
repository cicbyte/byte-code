// Package api project QA 库：常见问答沉淀（agent 维护 + 检索）。
package project

import "github.com/gogf/gf/v2/frame/g"

type QaUpsertReq struct {
	g.Meta    `path:"/projects/{projectId}/qas" method:"post" tags:"QA库" summary:"创建/更新 QA（按问题去重：同问题存在则更新答案）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Question  string `json:"question" v:"required|max-length:500#问题不能为空|上限500字"`
	Answer    string `json:"answer" v:"required#答案不能为空" dc:"markdown"`
	Tags      string `json:"tags" dc:"逗号分隔标签"`
}

type QaUpsertRes struct {
	g.Meta `mime:"application/json"`
	Id     int  `json:"id"`
	Updated bool `json:"updated" dc:"true=更新了已有条目"`
}

type QaListReq struct {
	g.Meta    `path:"/projects/{projectId}/qas" method:"get" tags:"QA库" summary:"QA 列表（关键词命中问题/答案/标签；按 hits 降序）"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Keyword   string `json:"keyword" in:"query"`
	Tag       string `json:"tag" in:"query"`
	Size      int    `json:"size" in:"query" d:"50"`
}

type QaItem struct {
	Id       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Tags     string `json:"tags"`
	Hits     int    `json:"hits"`
	Status   string `json:"status"`
	Updater  string `json:"updater"`
	UpdatedAt string `json:"updatedAt"`
}

type QaListRes struct {
	g.Meta `mime:"application/json"`
	List   []QaItem `json:"list"`
}

type QaHitReq struct {
	g.Meta    `path:"/projects/{projectId}/qas/{id}/hit" method:"post" tags:"QA库" summary:"命中计数(agent/人查阅时调用; 开工包排序依据)"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	Id        int `json:"id" v:"required" in:"path"`
}

type QaHitRes struct {
	g.Meta `mime:"application/json"`
}

type QaArchiveReq struct {
	g.Meta    `path:"/projects/{projectId}/qas/{id}/archive" method:"post" tags:"QA库" summary:"归档(过期/失效的 QA)"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	Id        int `json:"id" v:"required" in:"path"`
}

type QaArchiveRes struct {
	g.Meta `mime:"application/json"`
}
