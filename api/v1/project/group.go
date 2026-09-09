package project

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 项目分组（多对多） ====================
// 同 group 的项目自动互为关联（反馈/引用免准入），替代手动 project_relations

type GroupCreateReq struct {
	g.Meta      `path:"/groups" method:"post" tags:"项目分组" summary:"创建分组"`
	Name        string `json:"name" v:"required|length:2,64#分组名不能为空|分组名长度2-64"`
	Description string `json:"description"`
}

type GroupCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type GroupListReq struct {
	g.Meta `path:"/groups" method:"get" tags:"项目分组" summary:"分组列表（含成员项目摘要）"`
}

type GroupListItem struct {
	Id          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedBy   int          `json:"createdBy"`
	CreatedAt   string       `json:"createdAt"`
	Projects    []GroupProjectBrief `json:"projects"`
}

type GroupProjectBrief struct {
	Id   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type GroupListRes struct {
	g.Meta `mime:"application/json"`
	List   []GroupListItem `json:"list"`
}

type GroupUpdateReq struct {
	g.Meta      `path:"/groups/{id}" method:"put" tags:"项目分组" summary:"更新分组"`
	Id          int     `json:"id" v:"required" in:"path"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type GroupUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type GroupDeleteReq struct {
	g.Meta `path:"/groups/{id}" method:"delete" tags:"项目分组" summary:"删除分组（成员关系级联清除）"`
	Id     int `json:"id" v:"required" in:"path"`
}

type GroupDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type GroupMemberAddReq struct {
	g.Meta    `path:"/groups/{id}/projects" method:"post" tags:"项目分组" summary:"添加项目到分组"`
	Id        int `json:"id" v:"required" in:"path"`
	ProjectId int `json:"projectId" v:"required#项目不能为空"`
}

type GroupMemberAddRes struct {
	g.Meta `mime:"application/json"`
}

type GroupMemberRemoveReq struct {
	g.Meta    `path:"/groups/{id}/projects/{projectId}" method:"delete" tags:"项目分组" summary:"从分组移除项目"`
	Id        int `json:"id" v:"required" in:"path"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type GroupMemberRemoveRes struct {
	g.Meta `mime:"application/json"`
}
