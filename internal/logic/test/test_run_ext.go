package test

import (
	"context"
	"fmt"
	"strings"

	projApi "github.com/cicbyte/byte-code/api/v1/project"
	api "github.com/cicbyte/byte-code/api/v1/test"
	service "github.com/cicbyte/byte-code/internal/service"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 失败闭环 / Flaky / 趋势（#506 P3） ====================

// CaseToBug 失败用例转缺陷：缺省自动建 bug 任务（描述带 traceback 与执行
// 上下文），复用 CreateTask 的未指派广播——绑定 agent 即刻收到可认领通知，
// 修复后 `bcode test --run` 重跑形成验证闭环。taskId 非 0 时仅挂接既有任务
func (s *sTest) CaseToBug(ctx context.Context, req *api.TestRunCaseBugReq) (res *api.TestRunCaseBugRes, err error) {
	row, rerr := g.DB().Model("test_run_cases trc").Ctx(ctx).
		InnerJoin("test_runs tr", "trc.test_run_id = tr.id").
		Fields("trc.id, trc.title, trc.external_key, trc.status, trc.message, trc.bug_task_id, "+
			"tr.project_id, tr.branch, tr.git_sha, tr.finished_at, tr.id AS run_id").
		Where("trc.id", req.Id).
		One()
	if rerr != nil {
		return nil, liberr.WrapDb(ctx, rerr, "查询执行用例失败")
	}
	if row.IsEmpty() {
		return nil, fmt.Errorf("执行用例不存在")
	}
	pid := row["project_id"].Int()
	// 建任务侧门禁与任务域一致（tasks_write；人类直通）
	if err := perm.AgentTaskGate(ctx, pid, "tasks_write"); err != nil {
		return nil, err
	}
	if b := row["bug_task_id"].Int(); b > 0 {
		return nil, fmt.Errorf("已挂接缺陷任务 #%d（先解挂再转）", b)
	}

	taskId := req.TaskId
	created := false
	if taskId > 0 {
		// 挂接既有任务：必须同项目
		tp, terr := g.DB().Model("tasks").Ctx(ctx).Where("id", taskId).Fields("project_id").Value()
		if terr != nil {
			return nil, liberr.WrapDb(ctx, terr, "查询任务失败")
		}
		if tp == nil {
			return nil, fmt.Errorf("任务 #%d 不存在", taskId)
		}
		if tp.Int() != pid {
			return nil, fmt.Errorf("任务 #%d 不属于该项目", taskId)
		}
	} else {
		title := req.Title
		if title == "" {
			title = bugTitleFrom(row["title"].String(), row["external_key"].String())
		}
		prio := req.Priority
		if prio < 1 || prio > 4 {
			prio = 2
		}
		created = true
		taskId, err = s.createBugTask(ctx, pid, title, prio, row)
		if err != nil {
			return nil, err
		}
	}

	if _, uerr := g.DB().Model("test_run_cases").Ctx(ctx).
		WherePri(req.Id).
		Data(g.Map{"bug_task_id": taskId}).
		Update(); uerr != nil {
		return nil, liberr.WrapDb(ctx, uerr, "回填缺陷挂钩失败")
	}
	return &api.TestRunCaseBugRes{TaskId: taskId, Created: created}, nil
}

func bugTitleFrom(title, externalKey string) string {
	name := title
	if name == "" {
		name = externalKey
	}
	if len([]rune(name)) > 60 {
		name = string([]rune(name)[:60]) + "…"
	}
	return "测试失败：" + name
}

// createBugTask 描述拼装：traceback + 执行上下文（分支/commit/时间/记录入口），
// 让 agent 拿到任务即可定位，无需回查平台
func (s *sTest) createBugTask(ctx context.Context, pid int, title string, prio int, row gdb.Record) (int, error) {
	var desc strings.Builder
	fmt.Fprintf(&desc, "- 失败用例：%s\n", row["title"].String())
	if k := row["external_key"].String(); k != "" {
		fmt.Fprintf(&desc, "- 外部键：%s\n", k)
	}
	fmt.Fprintf(&desc, "- 执行记录：run #%d（%s@%s %s）\n",
		row["run_id"].Int(), row["branch"].String(), row["git_sha"].String(), row["finished_at"].String())
	fmt.Fprintf(&desc, "- 平台入口：项目 %d → 测试 → 执行记录 → run #%d\n", pid, row["run_id"].Int())
	fmt.Fprintf(&desc, "- 修复验证：修复后执行 `bcode test --run -- <测试命令>` 重跑并上报\n")
	if msg := strings.TrimSpace(row["message"].String()); msg != "" {
		fmt.Fprintf(&desc, "\n失败信息：\n```\n%s\n```\n", msg)
	}
	return serviceCreateTask(ctx, pid, title, desc.String(), prio)
}

// serviceCreateTask 经 service 层建任务：拿到 CreateTask 自带的未指派
// agent 广播（事件驱动认领）与统一校验
func serviceCreateTask(ctx context.Context, pid int, title, desc string, prio int) (int, error) {
	return service.Project().CreateTask(ctx, &projApi.TaskCreateReq{
		ProjectId:   pid,
		Title:       title,
		Description: desc,
		Type:        "bug",
		Priority:    prio,
	})
}

// recentRunIds 项目最近 N 次 run id（id 倒序；flaky/趋势共用窗口）
func recentRunIds(ctx context.Context, pid, limit int) ([]int, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := g.DB().Model("test_runs").Ctx(ctx).
		Where("project_id", pid).
		Order("id DESC").
		Limit(limit).
		Fields("id").Array()
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(rows))
	for _, v := range rows {
		ids = append(ids, v.Int())
	}
	return ids, nil
}

// flakyKeys 窗口内既有 pass 又有 fail/error 的 external_key 集合
func flakyKeys(ctx context.Context, runIds []int) (map[string]bool, error) {
	if len(runIds) == 0 {
		return map[string]bool{}, nil
	}
	rows, err := g.DB().Model("test_run_cases").Ctx(ctx).
		WhereIn("test_run_id", runIds).
		Group("external_key").
		Having("SUM(CASE WHEN status = 'pass' THEN 1 ELSE 0 END) > 0 " +
			"AND SUM(CASE WHEN status IN ('fail','error') THEN 1 ELSE 0 END) > 0").
		Fields("external_key").Array()
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(rows))
	for _, v := range rows {
		if k := v.String(); k != "" {
			out[k] = true
		}
	}
	return out, nil
}

// ListFlaky Flaky 用例：窗口内 pass/fail 混现 + 最近一次状态
func (s *sTest) ListFlaky(ctx context.Context, req *api.TestFlakyListReq) (res *api.TestFlakyListRes, err error) {
	res = &api.TestFlakyListRes{Window: req.Window, List: []api.TestFlakyItem{}}
	ids, err := recentRunIds(ctx, req.ProjectId, req.Window)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询执行窗口失败")
	}
	if len(ids) == 0 {
		return res, nil
	}
	// SUM CASE 聚合双方言可用（无窗口函数依赖）
	rows, err := g.DB().Model("test_run_cases").Ctx(ctx).
		WhereIn("test_run_id", ids).
		Where("external_key != ''").
		Group("external_key").
		Having("SUM(CASE WHEN status = 'pass' THEN 1 ELSE 0 END) > 0 " +
			"AND SUM(CASE WHEN status IN ('fail','error') THEN 1 ELSE 0 END) > 0").
		Fields("external_key, MAX(title) AS title, " +
			"SUM(CASE WHEN status = 'pass' THEN 1 ELSE 0 END) AS p, " +
			"SUM(CASE WHEN status IN ('fail','error') THEN 1 ELSE 0 END) AS f, " +
			"MAX(test_run_id) AS last_run_id").
		Order("f DESC, p DESC").
		Limit(100).
		All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计 Flaky 用例失败")
	}
	if len(rows) == 0 {
		return res, nil
	}
	// 最近一次状态：按各 key 的 last_run_id 批量回查（两次查询替代逐 key 循环）
	lastIds := make([]int, 0, len(rows))
	for _, r := range rows {
		lastIds = append(lastIds, r["last_run_id"].Int())
	}
	statusRows, serr := g.DB().Model("test_run_cases").Ctx(ctx).
		WhereIn("test_run_id", lastIds).
		Where("external_key != ''").
		Order("id ASC").
		Fields("external_key, status, test_run_id").All()
	lastStatus := map[string]string{}
	if serr == nil {
		for _, sr := range statusRows {
			lastStatus[sr["external_key"].String()] = sr["status"].String()
		}
	}
	for _, r := range rows {
		key := r["external_key"].String()
		res.List = append(res.List, api.TestFlakyItem{
			ExternalKey: key,
			Title:       r["title"].String(),
			PassCount:   r["p"].Int(),
			FailCount:   r["f"].Int(),
			LastStatus:  lastStatus[key],
			LastRunId:   r["last_run_id"].Int(),
		})
	}
	return res, nil
}

// ListTrends 近期执行 + 失败 Top（跨 run 按 external_key 聚合）
func (s *sTest) ListTrends(ctx context.Context, req *api.TestTrendReq) (res *api.TestTrendRes, err error) {
	res = &api.TestTrendRes{Runs: []api.TestTrendRun{}, TopFailed: []api.TestFailTop{}}
	ids, err := recentRunIds(ctx, req.ProjectId, req.Limit)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询执行窗口失败")
	}
	if len(ids) == 0 {
		return res, nil
	}
	var runs []api.TestTrendRun
	if err = g.DB().Model("test_runs").Ctx(ctx).
		WhereIn("id", ids).
		Order("id ASC").
		Fields("id, source, branch, total, passed, failed, errors, skipped, duration_ms, finished_at").
		Scan(&runs); err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询执行趋势失败")
	}
	res.Runs = runs

	var tops []struct {
		ExternalKey string `json:"externalKey"`
		Title       string
		F           int
		T           int
	}
	if err = g.DB().Model("test_run_cases").Ctx(ctx).
		WhereIn("test_run_id", ids).
		Where("external_key != ''").
		Group("external_key").
		Fields("external_key, MAX(title) AS title, " +
			"SUM(CASE WHEN status IN ('fail','error') THEN 1 ELSE 0 END) AS F, " +
			"COUNT(*) AS T").
		Order("F DESC, T DESC").
		Limit(10).
		Scan(&tops); err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计失败 Top 失败")
	}
	for _, t := range tops {
		res.TopFailed = append(res.TopFailed, api.TestFailTop{
			ExternalKey: t.ExternalKey, Title: t.Title, FailCount: t.F, TotalCount: t.T,
		})
	}
	return res, nil
}
