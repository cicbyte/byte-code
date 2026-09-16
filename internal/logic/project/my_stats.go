package project

import (
	"context"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 我的任务统计（个人效率视图）：作用域与 MyTaskList 一致——assignee=当前
// 用户，非管理员再限定其仍是成员的项目（被移出项目后的历史指派不进统计）。
// 交付周期与管理员效率报告同口径：created_at → completed_at 的小时差。
func (s *sProject) MyTaskStats(ctx context.Context, req *api.MyTaskStatsReq) (res *api.MyTaskStatsRes, err error) {
	res = &api.MyTaskStatsRes{
		StatusCounts: []api.MyStatusCount{},
		Trend:        []api.MyTrendPoint{},
		ByProject:    []api.MyProjectActive{},
	}
	uid := perm.UserId(ctx)
	base := func() *gdb.Model {
		m := g.DB().Model("tasks t").Ctx(ctx).Where("t.assignee_id", uid)
		if !perm.IsAdmin(ctx, uid) {
			m = m.Where("EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = t.project_id AND pm.user_id = ?)", uid)
		}
		return m
	}
	active := []string{"open", "in_progress", "blocked", "review"}
	done := []string{"done", "closed"}
	today := time.Now().Format("2006-01-02")

	// 全状态计数 + 活跃合计
	rows, err := base().Fields("t.status, COUNT(*) AS cnt").Group("t.status").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计任务状态失败")
	}
	activeSet := map[string]bool{}
	for _, a := range active {
		activeSet[a] = true
	}
	for _, r := range rows {
		st, n := r["status"].String(), r["cnt"].Int()
		res.StatusCounts = append(res.StatusCounts, api.MyStatusCount{Status: st, Count: n})
		if activeSet[st] {
			res.ActiveTotal += n
		}
	}

	// 逾期：活跃四态 + due_date 非空且早于今天（due_date 归一为 YYYY-MM-DD，串比较即日期比较）
	if res.Overdue, err = base().
		Where("t.status IN (?)", active).
		Where("t.due_date != ''").
		Where("t.due_date < ?", today).
		Count(); err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计逾期任务失败")
	}

	// 近 30 天逐日完成趋势（零填充；DATE() 双方言通用，见 usage 物化同款）
	from := time.Now().AddDate(0, 0, -29).Format("2006-01-02")
	trows, err := base().
		Fields("DATE(t.completed_at) AS day, COUNT(*) AS cnt").
		Where("t.status IN (?)", done).
		Where("t.completed_at >= ?", from+" 00:00:00").
		Group("day").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计完成趋势失败")
	}
	doneByDay := map[string]int{}
	for _, r := range trows {
		doneByDay[r["day"].String()] = r["cnt"].Int()
	}
	for i := 29; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		n := doneByDay[d]
		res.Trend = append(res.Trend, api.MyTrendPoint{Day: d, Done: n})
		res.Completed30d += n
	}

	// 近 90 天完成任务的创建→完成平均小时数（Go 侧算，julianday 是 SQLite 方言）
	leadFrom := time.Now().AddDate(0, 0, -89).Format("2006-01-02") + " 00:00:00"
	lrows, err := base().
		Fields("t.created_at, t.completed_at").
		Where("t.status IN (?)", done).
		Where("t.completed_at >= ?", leadFrom).
		All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计交付周期失败")
	}
	var leadSum float64
	leadN := 0
	for _, r := range lrows {
		if h := myLeadHours(r["created_at"].String(), r["completed_at"].String()); h >= 0 {
			leadSum += h
			leadN++
		}
	}
	if leadN > 0 {
		res.AvgLeadHours = float64(int64(leadSum/float64(leadN)*100+0.5)) / 100
	}

	// 活跃任务按项目分布（含项目名，倒序）
	prows, err := base().
		Fields("t.project_id, p.name AS pname, COUNT(*) AS cnt").
		LeftJoin("projects p", "p.id = t.project_id").
		Where("t.status IN (?)", active).
		Group("t.project_id, p.name").
		Order("cnt DESC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计项目分布失败")
	}
	for _, r := range prows {
		res.ByProject = append(res.ByProject, api.MyProjectActive{
			ProjectId:   r["project_id"].Int(),
			ProjectName: r["pname"].String(),
			Active:      r["cnt"].Int(),
		})
	}
	return res, nil
}

// myLeadHours 与 platform.hoursBetween 同语义（created→completed 小时差，
// 解析失败返回 -1 不计入均值）；跨包不导出，改口径时两处同步
func myLeadHours(start, end string) float64 {
	const layout = "2006-01-02 15:04:05"
	s, err := time.Parse(layout, start)
	if err != nil {
		return -1
	}
	e, err := time.Parse(layout, end)
	if err != nil {
		return -1
	}
	return e.Sub(s).Hours()
}
