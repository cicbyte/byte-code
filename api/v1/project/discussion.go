package project

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 讨论区（Discussions） ====================
// 论坛式想法/议题线程：不进任务工作流，成员与 agent（discuss 能力位）均可
// 发起与回复，为 AI 提供决策背景；成熟后 convert 转任务（血缘互链）。

type DiscussionCreateReq struct {
	g.Meta    `path:"/projects/{projectId}/discussions" method:"post" tags:"讨论区" summary:"发起讨论"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Title     string `json:"title" v:"required|max-length:191#标题不能为空|标题上限191字符"`
	Body      string `json:"body" v:"max-length:65535#正文上限64K" dc:"想法背景（markdown）"`
}

type DiscussionCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type DiscussionListReq struct {
	g.Meta    `path:"/projects/{projectId}/discussions" method:"get" tags:"讨论区" summary:"讨论列表"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Status    string `json:"status" in:"query" v:"in:,open,converted,archived#状态不合法" dc:"状态过滤（空=全部）"`
	commonApi.PageReq
}

type DiscussionListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []DiscussionItem `json:"list"`
}

type DiscussionItem struct {
	Id              int    `json:"id"`
	ProjectId       int    `json:"projectId"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	Status          string `json:"status" dc:"open/converted/archived"`
	AuthorId        int    `json:"authorId"`
	AuthorName      string `json:"authorName,omitempty"`
	AuthorType      string `json:"authorType" dc:"human/ai"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	ReplyCount      int    `json:"replyCount"`
	ConvertedTaskId int    `json:"convertedTaskId" dc:"转任务后的任务 id（未转为 0）"`
}

type DiscussionDetailReq struct {
	g.Meta `path:"/discussions/{id}" method:"get" tags:"讨论区" summary:"讨论详情（含回复）"`
	Id     int `json:"-" in:"path" v:"required#讨论ID不能为空"`
}

type DiscussionReplyItem struct {
	Id         int    `json:"id"`
	DiscussId  int    `json:"discussionId"`
	UserId     int    `json:"userId"`
	UserName   string `json:"userName,omitempty"`
	UserType   string `json:"userType" dc:"human/ai"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
}

type DiscussionDetailRes struct {
	g.Meta      `mime:"application/json"`
	Id          int                   `json:"id"`
	ProjectId   int                   `json:"projectId"`
	Title       string                `json:"title"`
	Body        string                `json:"body"`
	Status      string                `json:"status"`
	AuthorId    int                   `json:"authorId"`
	AuthorName  string                `json:"authorName,omitempty"`
	AuthorType  string                `json:"authorType"`
	CreatedAt   string                `json:"createdAt"`
	UpdatedAt   string                `json:"updatedAt"`
	ConvertedTaskId int               `json:"convertedTaskId"`
	Replies     []DiscussionReplyItem `json:"replies"`
}

type DiscussionUpdateReq struct {
	g.Meta `path:"/discussions/{id}" method:"put" tags:"讨论区" summary:"编辑讨论（作者或 maintainer）"`
	Id     int    `json:"-" in:"path" v:"required#讨论ID不能为空"`
	Title  string `json:"title" v:"max-length:191#标题上限191字符"`
	Body   string `json:"body" v:"max-length:65535#正文上限64K"`
}

type DiscussionUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type DiscussionDeleteReq struct {
	g.Meta `path:"/discussions/{id}" method:"delete" tags:"讨论区" summary:"删除讨论（作者或 maintainer，级联删回复）"`
	Id     int `json:"-" in:"path" v:"required#讨论ID不能为空"`
}

type DiscussionDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type DiscussionArchiveReq struct {
	g.Meta `path:"/discussions/{id}/archive" method:"post" tags:"讨论区" summary:"归档/恢复（作者或 maintainer）"`
	Id     int `json:"-" in:"path" v:"required#讨论ID不能为空"`
}

type DiscussionArchiveRes struct {
	g.Meta `mime:"application/json"`
	Status string `json:"status" dc:"归档后的状态（archived/open）"`
}

type DiscussionReplyCreateReq struct {
	g.Meta    `path:"/discussions/{id}/replies" method:"post" tags:"讨论区" summary:"回复讨论（成员直通；agent 需 discuss 能力）"`
	Id        int    `json:"-" in:"path" v:"required#讨论ID不能为空"`
	Content   string `json:"content" v:"required|max-length:10000#回复内容不能为空|回复上限1万字符"`
}

type DiscussionReplyCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type DiscussionConvertReq struct {
	g.Meta `path:"/discussions/{id}/convert" method:"post" tags:"讨论区" summary:"转为任务（作者或 maintainer；血缘互链）"`
	Id     int    `json:"-" in:"path" v:"required#讨论ID不能为空"`
	Title  string `json:"title" v:"max-length:191#任务标题上限191字符" dc:"缺省同讨论标题"`
	Type   string `json:"type" d:"feature" v:"in:feature,bug,chore,test,docs#任务类型不合法"`
}

type DiscussionConvertRes struct {
	g.Meta  `mime:"application/json"`
	TaskId  int `json:"taskId"`
	DiscussId int `json:"discussionId"`
}
