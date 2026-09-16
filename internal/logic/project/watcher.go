package project

import (
	"context"

	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

// 任务关注（watcher 订阅）：assignee/creator 之外的旁观点显式订阅任务动态。
// 关注是读订阅：读能力即可 watch；通知扇出见 fanOutWatchers

func (s *sProject) WatchTask(ctx context.Context, id int) (err error) {
	if err := perm.AgentTaskGate(ctx, perm.EntityProjectId(ctx, "tasks", id), "tasks_read"); err != nil {
		return err
	}
	uid := perm.UserId(ctx)
	// 先查后插 + 唯一约束兜底：重复关注幂等返回成功（并发双击最坏撞唯一键报错，可接受）
	n, err := g.DB().Model("task_watchers").Ctx(ctx).
		Where("task_id", id).Where("user_id", uid).Count()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询关注状态失败")
	}
	if n > 0 {
		return nil
	}
	if _, err := g.DB().Model("task_watchers").Ctx(ctx).Insert(g.Map{
		"task_id": id, "user_id": uid,
	}); err != nil {
		return liberr.WrapDb(ctx, err, "关注任务失败")
	}
	return nil
}

func (s *sProject) UnwatchTask(ctx context.Context, id int) (err error) {
	if err := perm.AgentTaskGate(ctx, perm.EntityProjectId(ctx, "tasks", id), "tasks_read"); err != nil {
		return err
	}
	if _, err := g.DB().Model("task_watchers").Ctx(ctx).
		Where("task_id", id).Where("user_id", perm.UserId(ctx)).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "取消关注失败")
	}
	return nil
}

// fanOutWatchers 向任务的关注者广播事件通知。notified 是本次事件已通知的
// 收件人去重表（含操作者本人）——调用方先把定向收件人（assignee/creator/
// 被回复者/被 @者）填进去，本函数只补漏网的 watcher，并在发送后回填去重表
func fanOutWatchers(ctx context.Context, taskId int, notified map[int]bool, title, content, notifyType string) {
	rows, err := g.DB().Model("task_watchers").Ctx(ctx).
		Fields("user_id").Where("task_id", taskId).All()
	if err != nil {
		return
	}
	for _, r := range rows {
		uid := r["user_id"].Int()
		if uid <= 0 || notified[uid] {
			continue
		}
		notified[uid] = true
		notify.Send(ctx, uid, title, content, notifyType, "task", taskId)
	}
}

// taskWatcherSummary 详情回填用：返回 (名称列表, 当前用户是否在关注中)
func taskWatcherSummary(ctx context.Context, taskId int) (names []string, watching bool) {
	rows, err := g.DB().Model("task_watchers w").Ctx(ctx).
		LeftJoin("sys_users u", "u.id = w.user_id").
		Fields("w.user_id, COALESCE(NULLIF(u.real_name, ''), u.username) as name").
		Where("w.task_id", taskId).
		Order("w.id ASC").Limit(200).
		All()
	if err != nil {
		return nil, false
	}
	uid := perm.UserId(ctx)
	names = make([]string, 0, len(rows))
	for _, r := range rows {
		if r["user_id"].Int() == uid {
			watching = true
		}
		names = append(names, r["name"].String())
	}
	return names, watching
}
