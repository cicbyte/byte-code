package project

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 工作日志（Worklogs） ====================
// 项目演化叙事：成员与 agent（worklog 能力位）撰写的总结性记录，
// 日期分组时间轴呈现；支持从已完成任务/发布半自动生成草稿（人工润色后发布）。
// 区别于活动流（动词级系统事件）与 git commit（代码快照）——这里记
// 「做了什么、结果如何、为什么」的决策叙事。

type WorklogCreateReq struct {
	g.Meta    `path:"/projects/{projectId}/worklogs" method:"post" tags:"工作日志" summary:"记录工作日志"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Content   string `json:"content" v:"required|max-length:65535#内容不能为空|内容上限64K"`
	Source    string `json:"source" d:"manual" v:"in:,manual,tasks#来源不合法" dc:"manual 手写 | tasks 任务草稿生成"`
}

type WorklogCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type WorklogListReq struct {
	g.Meta    `path:"/projects/{projectId}/worklogs" method:"get" tags:"工作日志" summary:"日志列表（倒序分页，前端按日期分组）"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	commonApi.PageReq
}

type WorklogListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []WorklogItem `json:"list"`
}

type WorklogItem struct {
	Id         int    `json:"id"`
	ProjectId  int    `json:"projectId"`
	AuthorId   int    `json:"authorId"`
	AuthorName string `json:"authorName,omitempty"`
	AuthorType string `json:"authorType" dc:"human/ai"`
	Content    string `json:"content"`
	Source     string `json:"source" dc:"manual 手写 | tasks 草稿生成"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type WorklogUpdateReq struct {
	g.Meta  `path:"/worklogs/{id}" method:"put" tags:"工作日志" summary:"编辑日志（作者或 maintainer）"`
	Id      int    `json:"-" in:"path" v:"required#日志ID不能为空"`
	Content string `json:"content" v:"required|max-length:65535#内容不能为空|内容上限64K"`
}

type WorklogUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type WorklogDeleteReq struct {
	g.Meta `path:"/worklogs/{id}" method:"delete" tags:"工作日志" summary:"删除日志（作者或 maintainer）"`
	Id     int `json:"-" in:"path" v:"required#日志ID不能为空"`
}

type WorklogDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type WorklogDraftReq struct {
	g.Meta    `path:"/projects/{projectId}/worklogs/draft" method:"get" tags:"工作日志" summary:"从已完成任务/发布生成草稿（不落库，人工润色后另发）"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	From      string `json:"from" in:"query" v:"required#起始日期不能为空" dc:"YYYY-MM-DD（含）"`
	To        string `json:"to" in:"query" v:"required#结束日期不能为空" dc:"YYYY-MM-DD（含）"`
}

type WorklogDraftRes struct {
	g.Meta       `mime:"application/json"`
	Content      string `json:"content" dc:"草稿文本（markdown 列表）"`
	TaskCount    int    `json:"taskCount"`
	ReleaseCount int    `json:"releaseCount"`
}
