// Package review 审核中心：跨项目聚合待审任务 + 批量审核。
// 审核动作语义与单任务 ReviewTask 完全一致（通过=done、驳回=in_progress），
// 本包只做聚合视图与批量编排，不重复实现审核规则。
package review

import "github.com/gogf/gf/v2/frame/g"

// PendingListReq 跨项目待审列表（登录即可见自己有权限的范围）
type PendingListReq struct {
	g.Meta `path:"/reviews/pending" method:"get" tags:"审核中心" summary:"跨项目待审任务聚合"`
	Size   int `json:"size" dc:"单页条数，缺省 200（审核队列不是无限流）"`
}

type PendingItem struct {
	Id                 int    `json:"id"`
	ProjectId          int    `json:"projectId"`
	ProjectName        string `json:"projectName"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Type               string `json:"type"`
	Priority           int    `json:"priority"`
	Source             string `json:"source"`
	Status             string `json:"status"`
	AssigneeName       string `json:"assigneeName"`
	RequiresHumanReview int   `json:"requiresHumanReview"`
	UpdatedAt          string `json:"updatedAt"`
}

type PendingListRes struct {
	List  []PendingItem `json:"list" dc:"待审任务（含项目名，按更新时间倒序）"`
	Total int           `json:"total"`
}

// BatchReviewReq 批量审核（逐条复用单任务审核规则，含 agent 禁审与人审门禁）
type BatchReviewReq struct {
	g.Meta `path:"/reviews/batch" method:"post" tags:"审核中心" summary:"批量通过/驳回"`
	Ids    []int  `json:"ids" v:"required#任务ID不能为空#min:1"`
	Status string `json:"status" v:"required#审核结果不能为空|in:approved,rejected#审核结果仅支持 approved/rejected"`
	// 批量驳回时的统一驳回理由（逐条落为任务评论，与单条驳回口径一致）
	Comment string `json:"comment"`
}

type BatchFailItem struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Error string `json:"error"`
}

type BatchReviewRes struct {
	Succeeded int              `json:"succeeded" dc:"成功条数"`
	Failed    []BatchFailItem  `json:"failed" dc:"失败明细（如任务已被他人审掉/不需要人审）"`
}
