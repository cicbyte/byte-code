package platform

import (
	"context"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/platform"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 使用分析 P2：日汇总物化 + 报表（趋势/漏斗/版本/效率）

// MaterializeUsageDaily 全量重建 usage_daily（幂等）。明细只有 30 天保留，
// 全量重聚合的行数上限 = 30 天 × 客户端数，代价可忽略；今日部分数据
// 每次重跑自然刷新。维护链排在 dbclean 之前，报表查询时惰性补跑兜底
func MaterializeUsageDaily(ctx context.Context) error {
	rows, err := g.DB().Model("usage_events").Ctx(ctx).
		Fields(`DATE(created_at) AS day, client, COUNT(*) AS calls,
			SUM(CASE WHEN error_code != 0 OR status_code >= 400 THEN 1 ELSE 0 END) AS err_calls,
			COUNT(DISTINCT actor_id) AS active_actors`).
		Group("day, client").
		Order("day ASC").
		All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "usage_daily 聚合失败")
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model("usage_daily").Ctx(ctx).Where("1=1").Delete(); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		inserts := make(g.List, 0, len(rows))
		for _, r := range rows {
			inserts = append(inserts, g.Map{
				"day":           r["day"].String(),
				"client":        r["client"].String(),
				"calls":         r["calls"].Int(),
				"err_calls":     r["err_calls"].Int(),
				"active_actors": r["active_actors"].Int(),
			})
		}
		_, err := tx.Model("usage_daily").Ctx(ctx).Insert(inserts)
		return err
	})
}

func (s *sPlatform) UsageReport(ctx context.Context, req *api.UsageReportReq) (res *api.UsageReportRes, err error) {
	days := req.Days
	if days <= 0 || days > 90 {
		days = 30
	}
	res = &api.UsageReportRes{
		WindowDays: days,
		Daily:      []api.UsageDailyPoint{},
		Funnel:     []api.UsageFunnelStep{},
		Versions:   []api.UsageVersionItem{},
	}
	before := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	beforeDay := before[:10]

	// 惰性物化兜底（维护链之外的重启缺口）；失败继续出报表（尽力而为）
	if err := MaterializeUsageDaily(ctx); err != nil {
		g.Log().Warningf(ctx, "usage report lazy materialize failed: %v", err)
	}

	// ---- 每日趋势（usage_daily 长期保留，可覆盖超过明细保留期的窗口） ----
	// activeActors 是 cli/web 两行活跃数的合计（同一天跨端重复的账号会被
	// 计两次——按端去重已够趋势语义，精确跨端去重需回明细表，不划算）
	dailyRows, err := g.DB().Model("usage_daily").Ctx(ctx).
		Fields("day, client, calls, err_calls, active_actors").
		Where("day >= ?", beforeDay).
		Order("day ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "每日趋势查询失败")
	}
	byDay := map[string]*api.UsageDailyPoint{}
	for _, r := range dailyRows {
		d := r["day"].String()
		p, ok := byDay[d]
		if !ok {
			p = &api.UsageDailyPoint{Day: d}
			byDay[d] = p
		}
		if r["client"].String() == "cli" {
			p.CliCalls += r["calls"].Int()
		} else {
			p.WebCalls += r["calls"].Int()
		}
		p.ErrCalls += r["err_calls"].Int()
		p.ActiveActors += r["active_actors"].Int()
	}
	for _, p := range byDay {
		res.Daily = append(res.Daily, *p)
	}
	sortDailyPoints(res.Daily)

	// ---- CLI 工作流漏斗（去重账号数） ----
	steps := []struct {
		label    string
		endpoint string
	}{
		{"建立会话", "/v1/agent/sessions"},
		{"查看任务", "/v1/agent/tasks"},
		{"认领任务", "/v1/tasks/{id}/claim"},
		{"完成任务", "/v1/tasks/{id}/complete"},
	}
	for i, st := range steps {
		v, err := g.DB().Model("usage_events").Ctx(ctx).
			Where("client = 'cli'").Where("created_at >= ?", before).
			Where("endpoint = ?", st.endpoint).
			Distinct().Fields("actor_id").Count()
		if err != nil {
			return nil, liberr.WrapDb(ctx, err, "漏斗查询失败")
		}
		res.Funnel = append(res.Funnel, api.UsageFunnelStep{
			Step: i + 1, Label: st.label, Actors: v,
		})
	}

	// ---- CLI 版本分布（空版本=未识别：UA 未带版本号） ----
	verRows, err := g.DB().Model("usage_events").Ctx(ctx).
		Fields("COALESCE(NULLIF(client_version, ''), '未识别') AS version, COUNT(*) AS calls, COUNT(DISTINCT actor_id) AS actors, MAX(created_at) AS last_seen").
		Where("client = 'cli'").Where("created_at >= ?", before).
		Group("version").Order("calls DESC").Limit(20).
		All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "版本分布查询失败")
	}
	for _, r := range verRows {
		res.Versions = append(res.Versions, api.UsageVersionItem{
			Version: r["version"].String(), Calls: r["calls"].Int(),
			Actors: r["actors"].Int(), LastSeen: r["last_seen"].String(),
		})
	}

	// ---- 任务效率报告 ----
	if err := s.fillEfficiency(ctx, res, before); err != nil {
		return nil, err
	}
	return res, nil
}

// fillEfficiency 窗口内完成的任务：总量/平均交付周期（创建→完成，小时）/
// 阻塞次数 + 人均效率表。周期在 Go 侧算（julianday 是 SQLite 方言）
func (s *sPlatform) fillEfficiency(ctx context.Context, res *api.UsageReportRes, before string) error {
	taskRows, err := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, assignee_id, created_at, completed_at").
		Where("completed_at != ''").Where("completed_at >= ?", before).
		Order("completed_at DESC").Limit(5000).All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "效率报告任务查询失败")
	}
	type actorAgg struct {
		completed int
		leadSum   float64
		leadN     int
	}
	byActor := map[int]*actorAgg{}
	var leadSum float64
	var leadN int
	for _, r := range taskRows {
		lead := hoursBetween(r["created_at"].String(), r["completed_at"].String())
		aid := r["assignee_id"].Int()
		a := byActor[aid]
		if a == nil {
			a = &actorAgg{}
			byActor[aid] = a
		}
		a.completed++
		if lead >= 0 {
			a.leadSum += lead
			a.leadN++
			leadSum += lead
			leadN++
		}
	}
	res.Efficiency.CompletedTotal = len(taskRows)
	if leadN > 0 {
		res.Efficiency.AvgLeadHours = round2(leadSum / float64(leadN))
	}

	// 窗口内阻塞事件（ai_execution_logs.action='blocked'，行为者可以是
	// agent 或人类 maintainer）按人聚合
	blockedRows, err := g.DB().Model("ai_execution_logs").Ctx(ctx).
		Fields("ai_user_id, COUNT(*) AS cnt").
		Where("action = 'blocked'").Where("created_at >= ?", before).
		Group("ai_user_id").All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "效率报告阻塞查询失败")
	}
	blockedBy := map[int]int{}
	blockedTotal := 0
	for _, r := range blockedRows {
		blockedBy[r["ai_user_id"].Int()] = r["cnt"].Int()
		blockedTotal += r["cnt"].Int()
	}
	res.Efficiency.BlockedTotal = blockedTotal

	// 人均表：有完成或有阻塞的账号都列（限 Top 15 按完成数）
	userRows, err := g.DB().Model("sys_users u").Ctx(ctx).
		Fields("u.id, COALESCE(NULLIF(u.real_name, ''), u.username) AS name, u.type").All()
	if err != nil {
		return liberr.WrapDb(ctx, err, "效率报告账号查询失败")
	}
	nameOf := map[int][2]string{}
	for _, r := range userRows {
		nameOf[r["id"].Int()] = [2]string{r["name"].String(), r["type"].String()}
	}
	list := make([]api.UsageActorEfficiency, 0, len(byActor))
	for aid, a := range byActor {
		if a.completed == 0 {
			continue
		}
		name, utype := nameOf[aid][0], nameOf[aid][1]
		if aid == 0 {
			name, utype = "未指派", ""
		}
		item := api.UsageActorEfficiency{
			AssigneeId: aid, AssigneeName: name, ActorType: utype,
			Completed: a.completed, Blocked: blockedBy[aid],
		}
		if a.leadN > 0 {
			item.AvgLeadHours = round2(a.leadSum / float64(a.leadN))
		}
		list = append(list, item)
	}
	// 完成数降序，同名次按 id 升序（输出稳定）
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Completed > list[i].Completed ||
				(list[j].Completed == list[i].Completed && list[j].AssigneeId < list[i].AssigneeId) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if len(list) > 15 {
		list = list[:15]
	}
	res.Efficiency.ByActor = list
	return nil
}

// hoursBetween 两个 'YYYY-MM-DD HH:MM:SS' 的小时差（end-start）；
// 解析失败返回 -1（不计入均值）
func hoursBetween(start, end string) float64 {
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

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// sortDailyPoints 按日期升序（字符串比较即字典序，YYYY-MM-DD 天然正确）
func sortDailyPoints(pts []api.UsageDailyPoint) {
	for i := 0; i < len(pts); i++ {
		for j := i + 1; j < len(pts); j++ {
			if pts[j].Day < pts[i].Day {
				pts[i], pts[j] = pts[j], pts[i]
			}
		}
	}
}
