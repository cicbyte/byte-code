package test

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/test"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 用例执行统计 / 历史 ====================

// 匹配口径（见 CaseStats/CaseRunsHistory 的 join/where）：显式映射
// trc.test_case_id 命中，或 external_key 相同——后者兼容 --bcode-sync
// 之前仅按 nodeid 落库的旧行；join 侧限定 tr.project_id = tc.project_id，
// 同 nodeid 在他项目的执行不计入

const caseStatFields = "COUNT(*) AS total, " +
	"SUM(CASE WHEN trc.status = 'pass' THEN 1 ELSE 0 END) AS p, " +
	"SUM(CASE WHEN trc.status = 'fail' THEN 1 ELSE 0 END) AS f, " +
	"SUM(CASE WHEN trc.status = 'error' THEN 1 ELSE 0 END) AS e, " +
	"SUM(CASE WHEN trc.status = 'skip' THEN 1 ELSE 0 END) AS sk"

// CaseStats 项目内用例执行统计（批量）：cases 页「执行」列数据源
func (s *sTest) CaseStats(ctx context.Context, req *api.TestCaseStatsReq) (res *api.TestCaseStatsRes, err error) {
	res = &api.TestCaseStatsRes{List: []api.TestCaseStatItem{}}
	rows, err := g.DB().Model("test_cases tc").Ctx(ctx).
		Where("tc.project_id", req.ProjectId).
		InnerJoin("test_run_cases trc", "(trc.test_case_id = tc.id OR (tc.external_key != '' AND trc.external_key = tc.external_key))").
		InnerJoin("test_runs tr", "tr.id = trc.test_run_id AND tr.project_id = tc.project_id").
		Group("tc.id").
		Fields("tc.id AS case_id, MAX(tc.external_key) AS ext_key, " + caseStatFields + ", " +
			"MAX(tr.finished_at) AS last_run_at, MAX(tr.id) AS last_run_id").
		All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计用例执行失败")
	}
	if len(rows) == 0 {
		return res, nil
	}

	// 各用例最近一次状态：按 last_run_id 批量回查该 run 内的执行行，
	// 优先 test_case_id 命中，回退 external_key（与主匹配口径一致）
	extOf := map[int]string{}
	lastIds := make([]int, 0, len(rows))
	for _, r := range rows {
		extOf[r["case_id"].Int()] = r["ext_key"].String()
		lastIds = append(lastIds, r["last_run_id"].Int())
	}
	byCase, byKey := map[int]string{}, map[string]string{}
	if statusRows, serr := g.DB().Model("test_run_cases").Ctx(ctx).
		WhereIn("test_run_id", lastIds).
		Order("id ASC").
		Fields("test_case_id, external_key, status").All(); serr == nil {
		for _, sr := range statusRows {
			if cid := sr["test_case_id"].Int(); cid > 0 {
				byCase[cid] = sr["status"].String()
			}
			if k := sr["external_key"].String(); k != "" {
				byKey[k] = sr["status"].String()
			}
		}
	}

	for _, r := range rows {
		cid := r["case_id"].Int()
		item := api.TestCaseStatItem{
			CaseId:    cid,
			Total:     r["total"].Int(),
			Pass:      r["p"].Int(),
			Fail:      r["f"].Int(),
			Error:     r["e"].Int(),
			Skip:      r["sk"].Int(),
			LastRunAt: r["last_run_at"].String(),
		}
		if v, ok := byCase[cid]; ok {
			item.LastStatus = v
		} else if extOf[cid] != "" {
			item.LastStatus = byKey[extOf[cid]]
		}
		res.List = append(res.List, item)
	}
	return res, nil
}

// CaseRunsHistory 单用例历史执行记录：汇总（全量行）+ 最近 N 条明细
func (s *sTest) CaseRunsHistory(ctx context.Context, req *api.TestCaseRunsReq) (res *api.TestCaseRunsRes, err error) {
	crow, qerr := g.DB().Model("test_cases").Ctx(ctx).
		Where("id", req.Id).
		Fields("id, project_id, external_key").One()
	if qerr != nil {
		return nil, liberr.WrapDb(ctx, qerr, "查询测试用例失败")
	}
	if crow.IsEmpty() {
		return nil, fmt.Errorf("测试用例不存在")
	}
	pid := crow["project_id"].Int()
	extKey := crow["external_key"].String()

	applyMatch := func(m *gdb.Model) *gdb.Model {
		m = m.InnerJoin("test_runs tr", "tr.id = trc.test_run_id").
			Where("tr.project_id", pid)
		if extKey != "" {
			return m.Where("(trc.test_case_id = ? OR trc.external_key = ?)", req.Id, extKey)
		}
		return m.Where("trc.test_case_id", req.Id)
	}

	agg, aerr := applyMatch(g.DB().Model("test_run_cases trc").Ctx(ctx)).
		Fields(caseStatFields + ", MAX(tr.finished_at) AS last_run_at").
		One()
	if aerr != nil {
		return nil, liberr.WrapDb(ctx, aerr, "统计用例执行失败")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var list []api.TestCaseRunItem
	if err = applyMatch(g.DB().Model("test_run_cases trc").Ctx(ctx)).
		Fields("trc.id AS run_case_id, trc.status, trc.duration_ms, trc.message, " +
			"tr.id AS run_id, tr.source, tr.branch, tr.git_sha, tr.started_at, tr.finished_at").
		Order("tr.id DESC, trc.id DESC").
		Limit(limit).
		Scan(&list); err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询用例执行记录失败")
	}
	if list == nil {
		list = []api.TestCaseRunItem{}
	}
	summary := api.TestCaseStatItem{
		CaseId:    req.Id,
		Total:     agg["total"].Int(),
		Pass:      agg["p"].Int(),
		Fail:      agg["f"].Int(),
		Error:     agg["e"].Int(),
		Skip:      agg["sk"].Int(),
		LastRunAt: agg["last_run_at"].String(),
	}
	if len(list) > 0 {
		summary.LastStatus = list[0].Status
	}
	return &api.TestCaseRunsRes{Summary: summary, List: list}, nil
}
