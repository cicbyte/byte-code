package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/test"
)

type ITest interface {
	// 测试用例
	CreateCase(ctx context.Context, req *api.TestCaseCreateReq) (id int, err error)
	UpdateCase(ctx context.Context, req *api.TestCaseUpdateReq) (err error)
	DeleteCase(ctx context.Context, id int) (err error)
	ListCases(ctx context.Context, req *api.TestCaseListReq) (total int, list []api.TestCaseItem, err error)
	GetCase(ctx context.Context, id int) (res *api.TestCaseDetailRes, err error)
	// 测试计划
	CreatePlan(ctx context.Context, req *api.TestPlanCreateReq) (id int, err error)
	UpdatePlan(ctx context.Context, req *api.TestPlanUpdateReq) (err error)
	DeletePlan(ctx context.Context, id int) (err error)
	ListPlans(ctx context.Context, req *api.TestPlanListReq) (total int, list []api.TestPlanItem, err error)
	GetPlan(ctx context.Context, id int) (res *api.TestPlanDetailRes, err error)
	// 测试执行
	AddCasesToPlan(ctx context.Context, req *api.TestPlanAddCaseReq) (err error)
	ExecuteCase(ctx context.Context, req *api.TestCaseExecuteReq) (err error)
	GetPlanResults(ctx context.Context, req *api.TestPlanResultsReq) (res *api.TestPlanResultsRes, err error)
}

var localTest ITest

func Test() ITest {
	if localTest == nil {
		panic("implement not found for interface ITest, forgot register?")
	}
	return localTest
}

func RegisterTest(i ITest) {
	localTest = i
}
