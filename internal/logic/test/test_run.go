package test

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	api "github.com/cicbyte/byte-code/api/v1/test"
	service "github.com/cicbyte/byte-code/internal/service"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 测试执行记录（Run，#504 pytest P1） ====================

// runMessageMax 单条用例失败信息上限（字符）：traceback 全量进库会让单条
// 记录膨胀到 MB 级，列表/详情渲染与库体积都受伤；插件侧已截断，这里是
// 服务端兜底（不信任上报方）
const runMessageMax = 8000

// runCasesMax 单次上报用例数上限：防滥用（正常 pytest session 远低于此）
const runCasesMax = 5000

// runTimeLayout 与库内时间列约定一致（VARCHAR(19)；gtime 微秒超长教训 #484）
const runTimeLayout = "2006-01-02 15:04:05"

// isDuplicateKeyErr 唯一索引冲突判定（SQLite "UNIQUE constraint failed" / MySQL 1062）
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Error 1062") ||
		strings.Contains(msg, "Duplicate entry")
}

func truncateRunMessage(s string) string {
	if utf8.RuneCountInString(s) <= runMessageMax {
		return s
	}
	r := []rune(s)
	return string(r[:runMessageMax]) + "\n...[truncated]"
}

// parseRunTime 宽松解析客户端时间：仅接受 YYYY-MM-DD HH:MM:SS，非法返回零值
func parseRunTime(s string) (time.Time, bool) {
	if len(s) != 19 {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation(runTimeLayout, s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (s *sTest) ReportRun(ctx context.Context, req *api.TestRunReportReq) (id int, err error) {
	// 上报走 agent 身份（CI 机器即 agent）：能力门禁 + 会话项目约束；
	// 人类成员直通（AgentRequire 对人类放行）
	if err := perm.AgentTaskGate(ctx, req.ProjectId, "test_execute"); err != nil {
		return 0, err
	}
	if cnt, _ := g.DB().Model("projects").Ctx(ctx).Where("id", req.ProjectId).Count(); cnt == 0 {
		return 0, fmt.Errorf("项目不存在")
	}
	if len(req.Cases) == 0 || len(req.Cases) > runCasesMax {
		return 0, fmt.Errorf("用例结果数量须在 1~%d 之间", runCasesMax)
	}

	// 幂等键命中：同项目内已有同键 run 直接返回（插件/CI 网络重试安全）；
	// 并发竞态由迁移 80 的 (project_id, idempotency_key) 唯一索引兜底
	if req.IdempotencyKey != "" {
		if v, qerr := g.DB().Model("test_runs").Ctx(ctx).
			Where("project_id", req.ProjectId).
			Where("idempotency_key", req.IdempotencyKey).
			Fields("id").Value(); qerr == nil && v != nil {
			return 0, &service.IdempotentHitError{RunId: v.Int()}
		}
	}

	uid := perm.UserId(ctx)
	now := time.Now()

	// 起止/耗时推导：优先客户端值，缺失按服务端时钟补齐
	finished := now
	if t, ok := parseRunTime(req.FinishedAt); ok {
		finished = t
	}
	durationMs := req.DurationMs
	if durationMs <= 0 {
		if st, ok := parseRunTime(req.StartedAt); ok && !st.After(finished) {
			durationMs = int(finished.Sub(st).Milliseconds())
		}
	}
	started := finished
	if t, ok := parseRunTime(req.StartedAt); ok && !t.After(finished) {
		started = t
	} else if durationMs > 0 {
		started = finished.Add(-time.Duration(durationMs) * time.Millisecond)
	}
	startedStr, finishedStr := started.Format(runTimeLayout), finished.Format(runTimeLayout)

	// 汇总在服务端重算：不信任上报方计数（口径/篡改都归一）
	var counts struct{ total, pass, fail, skip, errs int }
	caseRows := make([]g.Map, 0, len(req.Cases))
	for i := range req.Cases {
		c := &req.Cases[i]
		counts.total++
		switch c.Status {
		case "pass":
			counts.pass++
		case "fail":
			counts.fail++
		case "skip":
			counts.skip++
		case "error":
			counts.errs++
		}
		if c.Title == "" {
			c.Title = c.ExternalKey
		}
		caseRows = append(caseRows, g.Map{
			// 待回填 test_run_id
			"test_case_id": c.TestCaseId,
			"external_key": c.ExternalKey,
			"title":        c.Title,
			"status":       c.Status,
			"duration_ms":  c.DurationMs,
			"message":      truncateRunMessage(c.Message),
		})
	}

	// 用例映射防御：跨项目/不存在的 test_case_id 静默置 0（降级为仅
	// external_key 记录）——CI 上报不应因用例被删/映射过期而整批失败
	if mapped := collectMappedIds(req.Cases); len(mapped) > 0 {
		validRows, verr := g.DB().Model("test_cases").Ctx(ctx).
			WhereIn("id", mapped).Where("project_id", req.ProjectId).
			Fields("id").All()
		if verr == nil {
			valid := map[int]bool{}
			for _, r := range validRows {
				valid[r["id"].Int()] = true
			}
			for i := range req.Cases {
				if req.Cases[i].TestCaseId > 0 && !valid[req.Cases[i].TestCaseId] {
					req.Cases[i].TestCaseId = 0
					caseRows[i]["test_case_id"] = 0
				}
			}
		}
	}

	actorType := "human"
	// activities.actor_type CHECK 白名单是 ('human','ai','system')——agent 上报
	// 落 'ai'（#504 曾误传 'agent' 被 CHECK 静默拒绝，活动流缺记录）
	if v, _ := g.DB().Model("sys_users").Where("id", uid).Fields("type").Value(); v != nil && v.String() == "ai" {
		actorType = "ai"
	}

	txErr := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		insertData := g.Map{
			"project_id":   req.ProjectId,
			"source":       req.Source,
			"branch":       req.Branch,
			"git_sha":      req.GitSha,
			"env":          req.Env,
			"triggered_by": uid,
			"total":        counts.total,
			"passed":       counts.pass,
			"failed":       counts.fail,
			"skipped":      counts.skip,
			"errors":       counts.errs,
			"duration_ms":  durationMs,
			"started_at":   startedStr,
			"finished_at":  finishedStr,
			"created_at":   now.Format(runTimeLayout),
		}
		if req.IdempotencyKey != "" {
			insertData["idempotency_key"] = req.IdempotencyKey
		}
		result, ierr := tx.Ctx(ctx).Model("test_runs").Insert(insertData)
		if ierr != nil {
			// 唯一索引冲突（并发同键双写）：回查既有 run 幂等返回
			if req.IdempotencyKey != "" && isDuplicateKeyErr(ierr) {
				if v, qerr := tx.Ctx(ctx).Model("test_runs").
					Where("project_id", req.ProjectId).
					Where("idempotency_key", req.IdempotencyKey).
					Fields("id").Value(); qerr == nil && v != nil {
					return &service.IdempotentHitError{RunId: v.Int()}
				}
			}
			return ierr
		}
		lastId, _ := result.LastInsertId()
		id = int(lastId)
		for _, row := range caseRows {
			row["test_run_id"] = id
		}
		if _, ierr = tx.Ctx(ctx).Model("test_run_cases").Insert(caseRows); ierr != nil {
			return ierr
		}
		return nil
	})
	if txErr != nil {
		return 0, liberr.WrapDb(ctx, txErr, "上报测试执行记录失败")
	}

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    uid,
		ActorType:  actorType,
		Action:     "test_run.reported",
		TargetType: "test_run",
		TargetID:   id,
		ProjectID:  req.ProjectId,
		Detail: fmt.Sprintf("上报测试执行: %d 用例（%d 失败/%d 错误），来源 %s",
			counts.total, counts.fail, counts.errs, req.Source),
	})
	return id, nil
}

// collectMappedIds 取上报里非零的用例映射 id（去重）
func collectMappedIds(cases []api.TestRunCaseReport) []int {
	seen := map[int]bool{}
	ids := make([]int, 0, len(cases))
	for i := range cases {
		if cases[i].TestCaseId > 0 && !seen[cases[i].TestCaseId] {
			seen[cases[i].TestCaseId] = true
			ids = append(ids, cases[i].TestCaseId)
		}
	}
	return ids
}

func (s *sTest) ListRuns(ctx context.Context, req *api.TestRunListReq) (total int, list []api.TestRunItem, err error) {
	// failed+errors>0 即整批 fail；两列都是本表列，纯 WHERE 可表达（方言无关）
	applyFilters := func(m *gdb.Model) *gdb.Model {
		if req.Source != "" {
			m = m.Where("test_runs.source", req.Source)
		}
		if req.Branch != "" {
			m = m.Where("test_runs.branch", req.Branch)
		}
		switch req.Status {
		case "fail":
			m = m.Where("(test_runs.failed + test_runs.errors) > 0")
		case "pass":
			m = m.Where("test_runs.failed = 0 AND test_runs.errors = 0")
		}
		return m
	}
	countM := applyFilters(g.DB().Model("test_runs").Ctx(ctx))
	total, err = countM.Count()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询执行记录数量失败")
	}

	m := applyFilters(g.DB().Model("test_runs").Ctx(ctx)).
		LeftJoin("sys_users u", "test_runs.triggered_by = u.id").
		Fields("test_runs.*, COALESCE(NULLIF(u.real_name, ''), u.username) AS triggered_by_name").
		Order("test_runs.id DESC")
	pageNum, pageSize := req.PageNum, req.PageSize
	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if err = m.Page(pageNum, pageSize).Scan(&list); err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询执行记录列表失败")
	}
	if list == nil {
		list = []api.TestRunItem{}
	}
	return total, list, nil
}

func (s *sTest) GetRun(ctx context.Context, id int) (res *api.TestRunDetailRes, err error) {
	res = &api.TestRunDetailRes{Cases: []api.TestRunCaseItem{}}
	err = g.DB().Model("test_runs").Ctx(ctx).
		LeftJoin("sys_users u", "test_runs.triggered_by = u.id").
		Fields("test_runs.*, COALESCE(NULLIF(u.real_name, ''), u.username) AS triggered_by_name").
		Where("test_runs.id", id).
		Scan(&res.TestRunItem)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询执行记录失败")
	}
	if res.TestRunItem.Id == 0 {
		return nil, fmt.Errorf("执行记录不存在")
	}
	var cases []api.TestRunCaseItem
	err = g.DB().Model("test_run_cases trc").Ctx(ctx).
		LeftJoin("test_cases tc", "trc.test_case_id = tc.id").
		LeftJoin("tasks bt", "trc.bug_task_id = bt.id").
		Fields("trc.id, trc.test_run_id, trc.test_case_id, tc.title AS test_case_title, "+
			"trc.external_key, trc.title, trc.status, trc.duration_ms, trc.message, "+
			"trc.bug_task_id, bt.title AS bug_task_title").
		Where("trc.test_run_id", id).
		Order("trc.id ASC").
		Scan(&cases)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询执行记录用例失败")
	}
	if cases != nil {
		res.Cases = cases
	}
	// flaky 富化：窗口内同 key 既有 pass 又有 fail/error 即标（#506）
	if flaky, ferr := flakyKeysForRun(ctx, res.TestRunItem.ProjectId); ferr == nil {
		for i := range res.Cases {
			res.Cases[i].Flaky = flaky[res.Cases[i].ExternalKey]
		}
	}
	return res, nil
}

// flakyKeysForRun 详情页富化用窗口（固定最近 10 次，与 Flaky 面板口径一致）
func flakyKeysForRun(ctx context.Context, pid int) (map[string]bool, error) {
	ids, err := recentRunIds(ctx, pid, 10)
	if err != nil {
		return nil, err
	}
	return flakyKeys(ctx, ids)
}

func (s *sTest) DeleteRun(ctx context.Context, id int) (err error) {
	// 执行记录属治理数据：owner/maintainer 可删（清误报/脏数据）
	pid := perm.EntityProjectId(ctx, "test_runs", id)
	if pid == 0 {
		return fmt.Errorf("执行记录不存在")
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), pid) {
		return fmt.Errorf("仅项目管理员可删除执行记录")
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Model("test_run_cases").Where("test_run_id", id).Delete(); err != nil {
			return err
		}
		_, err := tx.Ctx(ctx).Model("test_runs").WherePri(id).Delete()
		return err
	})
}
