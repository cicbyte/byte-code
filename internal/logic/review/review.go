package review

import (
	"context"

	reviewApi "github.com/cicbyte/byte-code/api/v1/review"
	projectApi "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/service"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

type sReview struct{}

func init() {
	service.RegisterReview(New())
}

func New() *sReview {
	return &sReview{}
}

// ListPending 跨项目待审聚合：管理员看全部，普通成员看所在项目
//（口径与 DashboardStats 的范围过滤一致，避免跨项目任务标题泄露）
func (s *sReview) ListPending(ctx context.Context, req *reviewApi.PendingListReq) (res *reviewApi.PendingListRes, err error) {
	res = &reviewApi.PendingListRes{List: []reviewApi.PendingItem{}, Total: 0}
	size := req.Size
	if size <= 0 || size > 500 {
		size = 200
	}

	uid := perm.UserId(ctx)
	memberOnly := uid > 0 && !perm.IsAdmin(ctx, uid)
	memberScope := "EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = t.project_id AND pm.user_id = ?)"

	m := g.DB().Model("tasks t").Ctx(ctx).
		Fields("t.id, t.project_id, t.title, t.description, t.type, t.priority, t.source, t.status, t.requires_human_review, t.updated_at, p.name AS project_name, au.real_name AS assignee_name").
		LeftJoin("projects p", "p.id = t.project_id").
		LeftJoin("sys_users au", "au.id = t.assignee_id").
		Where("t.status", "review").
		Order("t.updated_at DESC").
		Limit(size)
	cm := g.DB().Model("tasks t").Ctx(ctx).Where("t.status", "review")
	if memberOnly {
		m = m.Where(memberScope, uid)
		cm = cm.Where(memberScope, uid)
	}

	list := []reviewApi.PendingItem{}
	if err = m.Scan(&list); err != nil {
		return nil, liberr.WrapDb(ctx, err, "加载待审任务失败")
	}
	total, err := cm.Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计待审任务失败")
	}
	res.List = list
	res.Total = int(total)
	return
}

// BatchReview 批量审核：逐条走单任务 ReviewTask（人审门禁/agent 禁审/
// 状态校验/评论/活动/通知语义全部保留），单条失败不阻断整批。
// 失败明细带任务标题，前端逐条呈现而非整批报错
func (s *sReview) BatchReview(ctx context.Context, req *reviewApi.BatchReviewReq) (res *reviewApi.BatchReviewRes, err error) {
	res = &reviewApi.BatchReviewRes{Succeeded: 0, Failed: []reviewApi.BatchFailItem{}}
	comment := req.Comment
	if runes := []rune(comment); len(runes) > 2000 {
		comment = string(runes[:2000])
	}

	for _, id := range req.Ids {
		title := ""
		if task, terr := service.Project().GetTask(ctx, id); terr == nil && task != nil {
			title = task.Title
		}
		rerr := service.Project().ReviewTask(ctx, &projectApi.TaskReviewReq{Id: id, Status: req.Status, Comment: comment})
		if rerr != nil {
			res.Failed = append(res.Failed, reviewApi.BatchFailItem{Id: id, Title: title, Error: rerr.Error()})
			continue
		}
		res.Succeeded++
	}
	return res, nil
}
