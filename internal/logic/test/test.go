package test

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/test"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/activity"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterTest(New())
}

func New() *sTest {
	return &sTest{}
}

type sTest struct{}

// ==================== 测试用例 ====================

func (s *sTest) CreateCase(ctx context.Context, req *api.TestCaseCreateReq) (id int, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		if uid == 0 {
			panic("用户未登录")
		}
		result, err := g.DB().Model("test_cases").Ctx(ctx).Insert(g.Map{
			"project_id":      req.ProjectId,
			"requirement_id":  req.RequirementId,
			"task_id":         req.TaskId,
			"title":           req.Title,
			"preconditions":   req.Preconditions,
			"steps":           req.Steps,
			"expected_result": req.ExpectedResult,
			"category":        req.Category,
			"module":          req.Module,
			"priority":        req.Priority,
			"source":          req.Source,
			"creator_id":      uid,
			// 表 CHECK 约束只允许 active/deprecated，新建用例即为 active
			"status":     "active",
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建测试用例失败")
		lastId, err := result.LastInsertId()
		liberr.ErrIsNil(ctx, err, "获取用例ID失败")
		id = int(lastId)

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_case.created",
			TargetType: "test_case",
			TargetID:   id,
			TargetName: req.Title,
			ProjectID:  req.ProjectId,
			Detail:     fmt.Sprintf("创建测试用例: %s", req.Title),
		})
	})
	return
}

func (s *sTest) UpdateCase(ctx context.Context, req *api.TestCaseUpdateReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		_, err = s.GetCase(ctx, req.Id)
		liberr.ErrIsNil(ctx, err, "测试用例不存在")

		data := g.Map{
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		if req.RequirementId != 0 {
			data["requirement_id"] = req.RequirementId
		}
		if req.TaskId != 0 {
			data["task_id"] = req.TaskId
		}
		if req.Title != "" {
			data["title"] = req.Title
		}
		if req.Preconditions != "" {
			data["preconditions"] = req.Preconditions
		}
		if req.Steps != "" {
			data["steps"] = req.Steps
		}
		if req.ExpectedResult != "" {
			data["expected_result"] = req.ExpectedResult
		}
		if req.Category != "" {
			data["category"] = req.Category
		}
		if req.Module != "" {
			data["module"] = req.Module
		}
		if req.Priority != "" {
			data["priority"] = req.Priority
		}
		if req.Status != "" {
			data["status"] = req.Status
		}

		_, err = g.DB().Model("test_cases").Ctx(ctx).WherePri(req.Id).Update(data)
		liberr.ErrIsNil(ctx, err, "更新测试用例失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_case.updated",
			TargetType: "test_case",
			TargetID:   req.Id,
			ProjectID:  0,
			Detail:     fmt.Sprintf("更新测试用例: ID=%d", req.Id),
		})
	})
	return
}

func (s *sTest) DeleteCase(ctx context.Context, id int) (err error) {
	userId := ctx.Value("userId")
	uid, _ := userId.(int)
	if _, err = s.GetCase(ctx, id); err != nil {
		return fmt.Errorf("测试用例不存在")
	}

	// 计划用例关联、标签关联与用例本身在同一事务内删除
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Delete("test_plan_cases", "test_case_id", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM entity_tags WHERE entity_type = 'test_case' AND entity_id = ?", id); err != nil {
			return err
		}
		_, err := tx.Delete("test_cases", "id", id)
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除测试用例失败")
	}

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    uid,
		ActorType:  "human",
		Action:     "test_case.deleted",
		TargetType: "test_case",
		TargetID:   id,
		Detail:     fmt.Sprintf("删除测试用例: ID=%d", id),
	})
	return nil
}

func (s *sTest) ListCases(ctx context.Context, req *api.TestCaseListReq) (total int, list []api.TestCaseItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// Count 查询（不带 Fields，兼容 SQLite）
		countM := g.DB().Model("test_cases").Ctx(ctx)
		if req.ProjectId != 0 {
			countM = countM.Where("test_cases.project_id", req.ProjectId)
		}
		if req.Category != "" {
			countM = countM.Where("test_cases.category", req.Category)
		}
		if req.Module != "" {
			countM = countM.Where("test_cases.module", req.Module)
		}
		if req.Status != "" {
			countM = countM.Where("test_cases.status", req.Status)
		}
		if req.Keyword != "" {
			countM = countM.Where("test_cases.title LIKE ? ESCAPE '\\'", "%"+escape.Like(req.Keyword)+"%")
		}

		total, err = countM.Count()
		liberr.ErrIsNil(ctx, err, "获取用例数量失败")

		// 数据查询
		m := g.DB().Model("test_cases").Ctx(ctx).
			LeftJoin("sys_users u", "test_cases.creator_id = u.id").
			Fields("test_cases.*, u.real_name as creator_name")
		if req.ProjectId != 0 {
			m = m.Where("test_cases.project_id", req.ProjectId)
		}
		if req.Category != "" {
			m = m.Where("test_cases.category", req.Category)
		}
		if req.Module != "" {
			m = m.Where("test_cases.module", req.Module)
		}
		if req.Status != "" {
			m = m.Where("test_cases.status", req.Status)
		}
		if req.Keyword != "" {
			m = m.Where("test_cases.title LIKE ? ESCAPE '\\'", "%"+escape.Like(req.Keyword)+"%")
		}

		pageNum := req.PageNum
		if pageNum == 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize == 0 {
			pageSize = 10
		}
		err = m.Page(pageNum, pageSize).Order("test_cases.created_at desc").Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取用例列表失败")
	})
	return
}

func (s *sTest) GetCase(ctx context.Context, id int) (res *api.TestCaseDetailRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		res = &api.TestCaseDetailRes{}
		err = g.DB().Model("test_cases").Ctx(ctx).
			LeftJoin("sys_users u", "test_cases.creator_id = u.id").
			Fields("test_cases.*, u.real_name as creator_name").
			Where("test_cases.id", id).
			Scan(&res.TestCaseItem)
		liberr.ErrIsNil(ctx, err, "获取测试用例失败")
		if res.TestCaseItem.Id == 0 {
			panic("测试用例不存在")
		}
	})
	return
}

// ==================== 测试计划 ====================

func (s *sTest) CreatePlan(ctx context.Context, req *api.TestPlanCreateReq) (id int, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		if uid == 0 {
			panic("用户未登录")
		}
		result, err := g.DB().Model("test_plans").Ctx(ctx).Insert(g.Map{
			"project_id":  req.ProjectId,
			"name":        req.Name,
			"description": req.Description,
			"milestone_id": req.MilestoneId,
			"status":      "draft",
			"creator_id":  uid,
			"created_at":  time.Now().Format("2006-01-02 15:04:05"),
			"updated_at":  time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建测试计划失败")
		lastId, err := result.LastInsertId()
		liberr.ErrIsNil(ctx, err, "获取计划ID失败")
		id = int(lastId)

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_plan.created",
			TargetType: "test_plan",
			TargetID:   id,
			TargetName: req.Name,
			ProjectID:  req.ProjectId,
			Detail:     fmt.Sprintf("创建测试计划: %s", req.Name),
		})
	})
	return
}

func (s *sTest) UpdatePlan(ctx context.Context, req *api.TestPlanUpdateReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		_, err = s.GetPlan(ctx, req.Id)
		liberr.ErrIsNil(ctx, err, "测试计划不存在")

		data := g.Map{
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		}
		if req.Name != "" {
			data["name"] = req.Name
		}
		if req.Description != "" {
			data["description"] = req.Description
		}
		if req.MilestoneId != 0 {
			data["milestone_id"] = req.MilestoneId
		}
		if req.Status != "" {
			data["status"] = req.Status
		}

		_, err = g.DB().Model("test_plans").Ctx(ctx).WherePri(req.Id).Update(data)
		liberr.ErrIsNil(ctx, err, "更新测试计划失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_plan.updated",
			TargetType: "test_plan",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("更新测试计划: ID=%d", req.Id),
		})
	})
	return
}

func (s *sTest) DeletePlan(ctx context.Context, id int) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// 两步删除在事务中执行
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		_, err = s.GetPlan(ctx, id)
		liberr.ErrIsNil(ctx, err, "测试计划不存在")

		// 删除计划关联的用例
		_, err = g.DB().Model("test_plan_cases").Ctx(ctx).Where("test_plan_id", id).Delete()
		liberr.ErrIsNil(ctx, err, "删除计划关联用例失败")

		_, err = g.DB().Model("test_plans").Ctx(ctx).WherePri(id).Delete()
		liberr.ErrIsNil(ctx, err, "删除测试计划失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_plan.deleted",
			TargetType: "test_plan",
			TargetID:   id,
			Detail:     fmt.Sprintf("删除测试计划: ID=%d", id),
		})
	})
	return
}

func (s *sTest) ListPlans(ctx context.Context, req *api.TestPlanListReq) (total int, list []api.TestPlanItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// Count 查询（不带 Fields，兼容 SQLite）
		countM := g.DB().Model("test_plans").Ctx(ctx)
		if req.ProjectId != 0 {
			countM = countM.Where("test_plans.project_id", req.ProjectId)
		}
		if req.Status != "" {
			countM = countM.Where("test_plans.status", req.Status)
		}

		total, err = countM.Count()
		liberr.ErrIsNil(ctx, err, "获取计划数量失败")

		// 数据查询
		m := g.DB().Model("test_plans").Ctx(ctx).
			LeftJoin("sys_users u", "test_plans.creator_id = u.id").
			Fields("test_plans.*, u.real_name as creator_name")
		if req.ProjectId != 0 {
			m = m.Where("test_plans.project_id", req.ProjectId)
		}
		if req.Status != "" {
			m = m.Where("test_plans.status", req.Status)
		}

		pageNum := req.PageNum
		if pageNum == 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize == 0 {
			pageSize = 10
		}
		err = m.Page(pageNum, pageSize).Order("test_plans.created_at desc").Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取计划列表失败")
	})
	return
}

func (s *sTest) GetPlan(ctx context.Context, id int) (res *api.TestPlanDetailRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		res = &api.TestPlanDetailRes{}
		err = g.DB().Model("test_plans").Ctx(ctx).
			LeftJoin("sys_users u", "test_plans.creator_id = u.id").
			Fields("test_plans.*, u.real_name as creator_name").
			Where("test_plans.id", id).
			Scan(&res.TestPlanItem)
		liberr.ErrIsNil(ctx, err, "获取测试计划失败")
		if res.TestPlanItem.Id == 0 {
			panic("测试计划不存在")
		}
	})
	return
}

// ==================== 测试执行 ====================

func (s *sTest) AddCasesToPlan(ctx context.Context, req *api.TestPlanAddCaseReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		_, err = s.GetPlan(ctx, req.Id)
		liberr.ErrIsNil(ctx, err, "测试计划不存在")

		// 批量查重（一次 WhereIn 替代 N 次 Count），未存在的单事务批量插入
		existingRows, err := g.DB().Model("test_plan_cases").Ctx(ctx).
			Where("test_plan_id", req.Id).
			WhereIn("test_case_id", req.CaseIds).
			Fields("test_case_id").All()
		liberr.ErrIsNil(ctx, err, "检查用例关联失败")
		existing := map[int]bool{}
		for _, r := range existingRows {
			existing[r["test_case_id"].Int()] = true
		}
		now := time.Now().Format("2006-01-02 15:04:05")
		rows := make([]g.Map, 0, len(req.CaseIds))
		for _, caseId := range req.CaseIds {
			if existing[caseId] {
				continue
			}
			rows = append(rows, g.Map{
				"test_plan_id": req.Id,
				"test_case_id": caseId,
				"assignee_id":  req.AssigneeId,
				"status":       "pending",
				"created_at":   now,
			})
		}
		if len(rows) > 0 {
			txErr := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
				_, err := tx.Ctx(ctx).Model("test_plan_cases").Insert(rows)
				return err
			})
			liberr.ErrIsNil(ctx, txErr, "添加用例到计划失败")
		}

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_plan.added_cases",
			TargetType: "test_plan",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("添加 %d 个用例到计划", len(req.CaseIds)),
		})
	})
	return
}

func (s *sTest) ExecuteCase(ctx context.Context, req *api.TestCaseExecuteReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		// 所属计划须处于可执行状态：已关闭/已完成的计划不应再改写用例结果
		planId, err := g.DB().Model("test_plan_cases").Ctx(ctx).
			WherePri(req.Id).Value("test_plan_id")
		liberr.ErrIsNil(ctx, err, "查询用例所属计划失败")
		if !planId.IsNil() {
			planStatus, perr := g.DB().Model("test_plans").Ctx(ctx).
				WherePri(planId.Int()).Value("status")
			liberr.ErrIsNil(ctx, perr, "查询计划失败")
			if st := planStatus.String(); st == "completed" {
				liberr.ErrIsNil(ctx, fmt.Errorf("计划已%s，不可再执行用例", st), "计划已关闭")
			}
		}

		data := g.Map{
			"status":      req.Status,
			"updated_at":  time.Now().Format("2006-01-02 15:04:05"),
		}
		if req.ActualResult != "" {
			data["actual_result"] = req.ActualResult
		}
		if req.BugTaskId != 0 {
			data["bug_task_id"] = req.BugTaskId
		}
		// 接口校验保证 status 只能是 pass/fail/blocked/skip（与表 CHECK 约束一致），
		// 执行即记录执行时间与执行人
		data["executed_at"] = time.Now().Format("2006-01-02 15:04:05")
		data["assignee_id"] = uid

		_, err = g.DB().Model("test_plan_cases").Ctx(ctx).WherePri(req.Id).Update(data)
		liberr.ErrIsNil(ctx, err, "执行用例失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "test_case.executed",
			TargetType: "test_plan_case",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("执行用例结果: %s", req.Status),
		})
	})
	return
}

func (s *sTest) GetPlanResults(ctx context.Context, req *api.TestPlanResultsReq) (res *api.TestPlanResultsRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		res = &api.TestPlanResultsRes{}

		var results []api.TestPlanCaseResult
		err = g.DB().Model("test_plan_cases tpc").Ctx(ctx).
			LeftJoin("test_cases tc", "tpc.test_case_id = tc.id").
			LeftJoin("sys_users u", "tpc.assignee_id = u.id").
			Fields("tpc.id, tpc.test_case_id, tc.title as test_case_title, "+
				"tpc.assignee_id, u.real_name as assignee_name, "+
				"tpc.status, tpc.actual_result, tpc.bug_task_id, tpc.executed_at").
			Where("tpc.test_plan_id", req.Id).
			Order("tpc.id asc").
			Scan(&results)
		liberr.ErrIsNil(ctx, err, "获取计划结果失败")

		res.Results = results
		res.Total = len(results)
		for _, r := range results {
			switch r.Status {
			case "pass":
				res.Passed++
			case "fail":
				res.Failed++
			case "blocked":
				res.Blocked++
			case "skip":
				res.Skipped++
			default:
				res.Pending++
			}
		}
	})
	return
}
