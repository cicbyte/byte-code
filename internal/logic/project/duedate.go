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
			Where("EXISTS (SELECT 1 FROM sys_users u WHERE u.id = t.assignee_id AND u.type = 'human')")
	}
	overdueRows, err := baseQuery().
		Where("t.due_date < ?", today).
		Order("t.due_date ASC").
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

	sent := 0
	for _, r := range rows {
		due := r["due_date"].String()
		taskTitle := r["title"].String()
		projectName := r["project_name"].String()
		scope := taskTitle
		if projectName != "" {
			scope = fmt.Sprintf("[%s] %s", projectName, taskTitle)
		}

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

		// 当天同任务同类提醒已发过则跳过（定时+启动双跑会重复触发）
		cnt, cerr := g.DB().Model("notifications").Ctx(ctx).
			Where("source_type", "task").
			Where("source_id", r["id"].Int64()).
			Where("title", title).
			Where("created_at >= ?", dayStart).
			Count()
		if cerr != nil || cnt > 0 {
			continue
		}
		notify.Send(ctx, r["assignee_id"].Int(), title, content, ntype, "task", r["id"].Int())
		sent++
	}
	if sent > 0 {
		g.Log().Infof(ctx, "task due scan: %d notified", sent)
	}
	return nil
}
