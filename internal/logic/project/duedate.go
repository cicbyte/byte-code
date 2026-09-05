package project

// 任务截止日期：格式归一化 + 到期扫描提醒

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// normalizeDueDate 校验并归一化截止日期为 Y-m-d；空串合法（语义为清除）
func normalizeDueDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	// 兼容带时间的输入（Y-m-d H:i:s）：取日期部分
	if len(s) > 10 {
		s = s[:10]
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return "", gerror.New("截止日期格式应为 Y-m-d")
	}
	return s, nil
}

// ScanDueTasks 到期扫描：今天/明天截止与已逾期的未完成任务通知指派人。
// 仅通知 human 指派（AI 用户由引擎自调度，无需提醒）；
// 同一任务每天每类别至多一条（notifications 按 标题+source 查重）。
// 定时任务与启动各调用一次——失败只告警不阻塞
func ScanDueTasks(ctx context.Context) error {
	today := gtime.Now().Format("Y-m-d")
	tomorrow := gtime.Now().AddDate(0, 0, 1).Format("Y-m-d")
	dayStart := today + " 00:00:00"

	// 分两段查询：逾期任务长期堆积会占满单一 Limit 配额，把今天/明天到期
	// 的挤出结果集（越堆积越只提醒最老的）。今明段全量，逾期段封顶防失控
	const overdueCap = 200
	baseQuery := func() *gdb.Model {
		return g.DB().Model("tasks t").Ctx(ctx).
			Fields("t.id, t.title, t.due_date, t.assignee_id, p.name AS project_name").
			LeftJoin("projects p", "p.id = t.project_id").
			Where("t.status IN (?)", consts.TaskActiveStatuses).
			Where("t.due_date IS NOT NULL").Where("t.due_date != ''").
			Where("t.assignee_id > 0").
			Where("EXISTS (SELECT 1 FROM sys_users u WHERE u.id = t.assignee_id AND u.type = 'human' AND u.status = 1)")
	}
	// 逾期段取【最近逾期】的 200 条（DESC）：ASC 会天天重发最老的一批，
	// 把较新（往往更相关）的逾期任务永久挤出结果集
	overdueRows, err := baseQuery().
		Where("t.due_date < ?", today).
		Order("t.due_date DESC").
		Limit(overdueCap).
		All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "到期任务扫描失败")
	}
	todayRows, err := baseQuery().
		Where("t.due_date >= ? AND t.due_date <= ?", today, tomorrow).
		Order("t.due_date ASC").
		All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "到期任务扫描失败")
	}
	rows := append(overdueRows, todayRows...)

	// 当天已发提醒集合一次取回：去重键为 (source_id, user_id, title)——
	// 含 user_id 使任务当天改派后新负责人仍能收到自己的提醒；
	// 也消除了逐行 COUNT 的 N+1（最多 500 次额外查询）
	type dedupKey struct {
		sourceId int64
		userId   int
		title    string
	}
	sentToday := make(map[dedupKey]bool)
	if notifiedRows, nerr := g.DB().Model("notifications").Ctx(ctx).
		Fields("source_id, user_id, title").
		Where("source_type", "task").
		Where("title IN (?)", g.Slice{"任务已逾期", "任务今天到期", "任务明天到期"}).
		Where("created_at >= ?", dayStart).
		All(); nerr == nil {
		for _, n := range notifiedRows {
			sentToday[dedupKey{n["source_id"].Int64(), n["user_id"].Int(), n["title"].String()}] = true
		}
	}

	sent := 0
	for _, r := range rows {
		due := r["due_date"].String()
		taskTitle := r["title"].String()
		projectName := r["project_name"].String()
		scope := taskTitle
		if projectName != "" {
			scope = fmt.Sprintf("[%s] %s", projectName, taskTitle)
		}
		assignee := r["assignee_id"].Int()

		var title, content, ntype string
		switch {
		case due < today:
			days, _ := time.Parse("2006-01-02", today)
			d, _ := time.Parse("2006-01-02", due)
			overdue := int(days.Sub(d).Hours() / 24)
			title = "任务已逾期"
			content = fmt.Sprintf("%s 已逾期 %d 天（截止 %s），请处理或调整截止日期", scope, overdue, due)
			ntype = "warning"
		case due == today:
			title = "任务今天到期"
			content = fmt.Sprintf("%s 今天到期，请关注进度", scope)
			ntype = "warning"
		default: // due == tomorrow
			title = "任务明天到期"
			content = fmt.Sprintf("%s 明天到期，请提前安排", scope)
			ntype = "info"
		}

		if sentToday[dedupKey{r["id"].Int64(), assignee, title}] {
			continue
		}
		sentToday[dedupKey{r["id"].Int64(), assignee, title}] = true
		notify.Send(ctx, assignee, title, content, ntype, "task", r["id"].Int())
		sent++
	}
	if sent > 0 {
		g.Log().Infof(ctx, "task due scan: %d notified", sent)
	}
	return nil
}

// ReleaseStaleClaims 任务租约超时释放：agent 认领后失联（无心跳/未完成）的
// in_progress 任务超时回到 open 并解除指派，人或其他 agent 可重新认领。
// 判定依据：updated_at 距今超过阈值且 assignee 是 agent（human 的进行中任务
// 不自动释放——人有沟通渠道，agent 崩溃无人知晓）。与到期扫描同挂定时框架。
func ReleaseStaleClaims(ctx context.Context) error {
	// 阈值 2 小时：claim 后正常执行（读开工包→写代码→complete）远短于此；
	// 过短会把慢任务误释放，过长则失联任务卡死更久
	threshold := gtime.Now().Add(-2 * time.Hour)
	res, err := g.DB().Model("tasks t").Ctx(ctx).
		Fields("t.id").
		Where("t.status", "in_progress").
		Where("t.updated_at < ?", threshold).
		Where("EXISTS (SELECT 1 FROM sys_users u WHERE u.id = t.assignee_id AND u.type = 'ai')").
		LockUpdate().
		All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "租约释放查询失败")
	}
	if len(res) == 0 {
		return nil
	}
	ids := make([]int, 0, len(res))
	for _, r := range res {
		ids = append(ids, r["id"].Int())
	}
	// 回 open + 解除指派 + 记日志（释放动作本身入 ai_execution_logs 供追溯）
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先取原 assignee 供日志留痕，再释放
		rows, qe := tx.Ctx(ctx).Model("tasks").Fields("id, assignee_id").WhereIn("id", ids).All()
		if qe != nil {
			return qe
		}
		if _, e := tx.Ctx(ctx).Exec(
			"UPDATE tasks SET status = 'open', assignee_id = 0, updated_at = ? WHERE id IN (?)",
			gtime.Now(), ids); e != nil {
			return e
		}
		for _, r := range rows {
			if _, e := tx.Ctx(ctx).Model("ai_execution_logs").Insert(g.Map{
				"task_id": r["id"].Int(), "ai_user_id": r["assignee_id"].Int(),
				"action": "lease_expired", "detail": "认领超时未完成，系统自动释放（2 小时无进展）", "status": "failed",
			}); e != nil {
				break // 日志失败不阻断释放
			}
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "租约释放失败")
	}
	g.Log().Infof(ctx, "lease release: %d tasks returned to open", len(ids))
	return nil
}
