package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/test"
	"github.com/cicbyte/byte-code/internal/consts"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Test = testController{}

type testController struct {
	BaseController
}

// ==================== 测试用例 ====================

func (c *testController) TestCaseCreate(ctx context.Context, req *api.TestCaseCreateReq) (res *api.TestCaseCreateRes, err error) {
	res = new(api.TestCaseCreateRes)
	id, err := service.Test().CreateCase(ctx, req)
	res.Id = id
	return
}

func (c *testController) TestCaseUpdate(ctx context.Context, req *api.TestCaseUpdateReq) (res *api.TestCaseUpdateRes, err error) {
	res = new(api.TestCaseUpdateRes)
	err = service.Test().UpdateCase(ctx, req)
	return
}

func (c *testController) TestCaseDelete(ctx context.Context, req *api.TestCaseDeleteReq) (res *api.TestCaseDeleteRes, err error) {
	res = new(api.TestCaseDeleteRes)
	err = service.Test().DeleteCase(ctx, req.Id)
	return
}

func (c *testController) TestCaseList(ctx context.Context, req *api.TestCaseListReq) (res *api.TestCaseListRes, err error) {
	res = new(api.TestCaseListRes)
	if req.PageSize == 0 {
		req.PageSize = consts.PageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, list, err := service.Test().ListCases(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.List = list
	return
}

func (c *testController) TestCaseDetail(ctx context.Context, req *api.TestCaseDetailReq) (res *api.TestCaseDetailRes, err error) {
	return service.Test().GetCase(ctx, req.Id)
}

// ==================== 测试计划 ====================

func (c *testController) TestPlanCreate(ctx context.Context, req *api.TestPlanCreateReq) (res *api.TestPlanCreateRes, err error) {
	res = new(api.TestPlanCreateRes)
	id, err := service.Test().CreatePlan(ctx, req)
	res.Id = id
	return
}

func (c *testController) TestPlanUpdate(ctx context.Context, req *api.TestPlanUpdateReq) (res *api.TestPlanUpdateRes, err error) {
	res = new(api.TestPlanUpdateRes)
	err = service.Test().UpdatePlan(ctx, req)
	return
}

func (c *testController) TestPlanDelete(ctx context.Context, req *api.TestPlanDeleteReq) (res *api.TestPlanDeleteRes, err error) {
	res = new(api.TestPlanDeleteRes)
	err = service.Test().DeletePlan(ctx, req.Id)
	return
}

func (c *testController) TestPlanList(ctx context.Context, req *api.TestPlanListReq) (res *api.TestPlanListRes, err error) {
	res = new(api.TestPlanListRes)
	if req.PageSize == 0 {
		req.PageSize = consts.PageSize
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	total, list, err := service.Test().ListPlans(ctx, req)
	res.Total = total
	res.CurrentPage = req.PageNum
	res.List = list
	return
}

func (c *testController) TestPlanDetail(ctx context.Context, req *api.TestPlanDetailReq) (res *api.TestPlanDetailRes, err error) {
	return service.Test().GetPlan(ctx, req.Id)
}

// ==================== 测试执行 ====================

func (c *testController) TestPlanAddCase(ctx context.Context, req *api.TestPlanAddCaseReq) (res *api.TestPlanAddCaseRes, err error) {
	res = new(api.TestPlanAddCaseRes)
	err = service.Test().AddCasesToPlan(ctx, req)
	return
}

func (c *testController) TestCaseExecute(ctx context.Context, req *api.TestCaseExecuteReq) (res *api.TestCaseExecuteRes, err error) {
	res = new(api.TestCaseExecuteRes)
	err = service.Test().ExecuteCase(ctx, req)
	return
}

func (c *testController) TestPlanResults(ctx context.Context, req *api.TestPlanResultsReq) (res *api.TestPlanResultsRes, err error) {
	return service.Test().GetPlanResults(ctx, req)
}
