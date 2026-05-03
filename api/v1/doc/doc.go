package doc

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 文档 CRUD ====================

type DocCreateReq struct {
	g.Meta    `path:"/v1/docs" method:"post" tags:"知识库" summary:"创建文档"`
	ProjectId int    `json:"projectId" v:"required#项目ID不能为空"`
	ParentId  int    `json:"parentId" dc:"父文档ID"`
	Title     string `json:"title" v:"required#文档标题不能为空"`
	Content   string `json:"content" dc:"文档内容"`
	Type      string `json:"type" v:"in:document,wiki,api,design|default:document#类型必须是document/wiki/api/design"`
	SortOrder int    `json:"sortOrder" dc:"排序"`
}

type DocCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type DocUpdateReq struct {
	g.Meta    `path:"/v1/docs/{id}" method:"put" tags:"知识库" summary:"更新文档"`
	Id        int    `json:"-" in:"path" v:"required#文档ID不能为空"`
	ParentId  int    `json:"parentId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type" v:"in:document,wiki,api,design#类型必须是document/wiki/api/design"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status" v:"in:active,archived,draft#状态必须是active/archived/draft"`
}

type DocUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type DocDeleteReq struct {
	g.Meta `path:"/v1/docs/{id}" method:"delete" tags:"知识库" summary:"删除文档"`
	Id     int `json:"-" in:"path" v:"required#文档ID不能为空"`
}

type DocDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type DocListReq struct {
	g.Meta    `path:"/v1/docs" method:"get" tags:"知识库" summary:"文档列表"`
	ProjectId int    `json:"projectId" dc:"项目ID"`
	ParentId  int    `json:"parentId" dc:"父文档ID"`
	Type      string `json:"type" dc:"文档类型"`
	Status    string `json:"status" dc:"状态筛选"`
	Keyword   string `json:"keyword" dc:"关键字搜索"`
	commonApi.PageReq
}

type DocListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []DocItem `json:"list"`
}

type DocItem struct {
	Id          int    `json:"id"`
	ProjectId   int    `json:"projectId"`
	ParentId    int    `json:"parentId"`
	Title       string `json:"title"`
	Content     string `json:"content,omitempty"`
	Type        string `json:"type"`
	SortOrder   int    `json:"sortOrder"`
	CreatorId   int    `json:"creatorId"`
	LastEditorId int   `json:"lastEditorId"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	Source      string `json:"source"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	CreatorName string `json:"creatorName,omitempty"`
	EditorName  string `json:"editorName,omitempty"`
}

type DocDetailReq struct {
	g.Meta `path:"/v1/docs/{id}" method:"get" tags:"知识库" summary:"文档详情"`
	Id     int `json:"-" in:"path" v:"required#文档ID不能为空"`
}

type DocDetailRes struct {
	g.Meta `mime:"application/json"`
	DocItem
}

// ==================== 文档移动 ====================

type DocMoveReq struct {
	g.Meta   `path:"/v1/docs/{id}/move" method:"put" tags:"知识库" summary:"移动文档"`
	Id       int `json:"-" in:"path" v:"required#文档ID不能为空"`
	ParentId int `json:"parentId" v:"required#目标父文档ID不能为空"`
}

type DocMoveRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 文档树 ====================

type DocTreeReq struct {
	g.Meta    `path:"/v1/projects/{projectId}/docs/tree" method:"get" tags:"知识库" summary:"项目文档树"`
	ProjectId int `json:"-" in:"path" v:"required#项目ID不能为空"`
}

type DocTreeRes struct {
	g.Meta `mime:"application/json"`
	Tree   []DocTreeNode `json:"tree"`
}

type DocTreeNode struct {
	Id        int           `json:"id"`
	ParentId  int           `json:"parentId"`
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	SortOrder int           `json:"sortOrder"`
	Children  []DocTreeNode `json:"children,omitempty"`
}

// ==================== 文档版本 ====================

type DocVersionListReq struct {
	g.Meta `path:"/v1/docs/{id}/versions" method:"get" tags:"知识库" summary:"版本历史"`
	Id     int `json:"-" in:"path" v:"required#文档ID不能为空"`
	commonApi.PageReq
}

type DocVersionListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []DocVersionItem `json:"list"`
}

type DocVersionItem struct {
	Id            int    `json:"id"`
	DocId         int    `json:"docId"`
	Version       int    `json:"version"`
	Content       string `json:"content,omitempty"`
	EditorId      int    `json:"editorId"`
	EditorName    string `json:"editorName,omitempty"`
	ChangeSummary string `json:"changeSummary"`
	CreatedAt     string `json:"createdAt"`
}

type DocRevertReq struct {
	g.Meta    `path:"/v1/docs/{id}/revert" method:"post" tags:"知识库" summary:"回退版本"`
	Id        int    `json:"-" in:"path" v:"required#文档ID不能为空"`
	Version   int    `json:"version" v:"required#目标版本号不能为空"`
	Summary   string `json:"summary" dc:"回退说明"`
}

type DocRevertRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 文档关联 ====================

type DocRelationCreateReq struct {
	g.Meta     `path:"/v1/docs/{id}/relations" method:"post" tags:"知识库" summary:"创建关联"`
	Id         int    `json:"-" in:"path" v:"required#文档ID不能为空"`
	TargetType string `json:"targetType" v:"required|in:task,requirement,test_case,doc#目标类型不能为空"`
	TargetId   int    `json:"targetId" v:"required#目标ID不能为空"`
}

type DocRelationCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type DocRelationDeleteReq struct {
	g.Meta     `path:"/v1/docs/{id}/relations/{targetType}/{targetId}" method:"delete" tags:"知识库" summary:"删除关联"`
	Id         int    `json:"-" in:"path" v:"required#文档ID不能为空"`
	TargetType string `json:"-" in:"path" v:"required|in:task,requirement,test_case,doc#目标类型不能为空"`
	TargetId   int    `json:"-" in:"path" v:"required#目标ID不能为空"`
}

type DocRelationDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type DocRelationListReq struct {
	g.Meta `path:"/v1/docs/{id}/relations" method:"get" tags:"知识库" summary:"关联列表"`
	Id     int `json:"-" in:"path" v:"required#文档ID不能为空"`
}

type DocRelationListRes struct {
	g.Meta `mime:"application/json"`
	List   []DocRelationItem `json:"list"`
}

type DocRelationItem struct {
	Id         int    `json:"id"`
	DocId      int    `json:"docId"`
	TargetType string `json:"targetType"`
	TargetId   int    `json:"targetId"`
	TargetTitle string `json:"targetTitle,omitempty"`
}

// ==================== 全局搜索 ====================

type SearchReq struct {
	g.Meta `path:"/v1/search" method:"get" tags:"知识库" summary:"全局搜索"`
	Keyword string `json:"keyword" v:"required#搜索关键字不能为空"`
	commonApi.PageReq
}

type SearchRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []SearchResult `json:"list"`
}

type SearchResult struct {
	Module  string `json:"module"`
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
