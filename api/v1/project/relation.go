// Package api project 项目关联（项目级协作）：
// A 关联 B 后，A 可向 B 投递跨项目反馈（线索），由 B 的准入 agent 阅读
// 分析是否建任务——A 不直接向 B 的任务池写入（决定权在对方）。owner 级管理。
package project

import "github.com/gogf/gf/v2/frame/g"

type RelationListReq struct {
	g.Meta    `path:"/projects/{projectId}/relations" method:"get" tags:"项目关联" summary:"关联项目列表"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type RelationItem struct {
	Id        int    `json:"id"`
	ProjectId int    `json:"projectId"` // 关联对方项目
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type RelationListRes struct {
	g.Meta `mime:"application/json"`
	List   []RelationItem `json:"list"`
}

type RelationAddReq struct {
	g.Meta       `path:"/projects/{projectId}/relations" method:"post" tags:"项目关联" summary:"添加关联项目（owner；对方须为本账号可访问项目）"`
	ProjectId    int `json:"projectId" v:"required" in:"path"`
	RelatedProjectId int `json:"relatedProjectId" v:"required#对方项目不能为空"`
}

type RelationAddRes struct {
	g.Meta `mime:"application/json"`
}

type RelationRemoveReq struct {
	g.Meta    `path:"/projects/{projectId}/relations/{relationId}" method:"delete" tags:"项目关联" summary:"移除关联项目（owner）"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	RelationId int `json:"relationId" v:"required" in:"path"`
}

type RelationRemoveRes struct {
	g.Meta `mime:"application/json"`
}
