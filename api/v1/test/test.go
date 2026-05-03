package test

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 测试用例 ====================

type TestCaseCreateReq struct {
	g.Meta        `path:"/projects/{projectId}/test-cases" method:"post" tags:"测试管理" summary:"创建测试用例"`
	ProjectId     int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	RequirementId int    `json:"requirementId" dc:"关联需求ID"`
	TaskId        int    `json:"taskId" dc:"关联任务ID"`
	Title         string `json:"title" v:"required#用例标题不能为空"`
	Preconditions string `json:"preconditions" dc:"前置条件"`
	Steps         string `json:"steps" dc:"测试步骤"`
	ExpectedResult string `json:"expectedResult" dc:"预期结果"`
	Category      string `json:"category" dc:"分类:功能/性能/安全/兼容性"`
	Module        string `json:"module" dc:"所属模块"`
	Priority      string `json:"priority" v:"required|in:P0,P1,P2,P3#优先级不能为空|优先级必须是P0/P1/P2/P3"`
	Source        string `json:"source" dc:"来源:manual/ai"`
}

type TestCaseCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type TestCaseUpdateReq struct {
	g.Meta         `path:"/test-cases/{id}" method:"put" tags:"测试管理" summary:"更新测试用例"`
	Id             int    `json:"-" in:"path" v:"required#用例ID不能为空"`
	RequirementId  int    `json:"requirementId" dc:"关联需求ID"`
	TaskId         int    `json:"taskId" dc:"关联任务ID"`
	Title          string `json:"title"`
	Preconditions  string `json:"preconditions"`
	Steps          string `json:"steps"`
	ExpectedResult string `json:"expectedResult"`
	Category       string `json:"category"`
	Module         string `json:"module"`
	Priority       string `json:"priority" v:"in:P0,P1,P2,P3#优先级必须是P0/P1/P2/P3"`
	Status         string `json:"status" v:"in:draft,active,deprecated#状态必须是draft/active/deprecated"`
}

type TestCaseUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TestCaseDeleteReq struct {
	g.Meta `path:"/test-cases/{id}" method:"delete" tags:"测试管理" summary:"删除测试用例"`
	Id     int `json:"-" in:"path" v:"required#用例ID不能为空"`
}

type TestCaseDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type TestCaseListReq struct {
	g.Meta    `path:"/projects/{projectId}/test-cases" method:"get" tags:"测试管理" summary:"用例列表"`
	ProjectId int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	commonApi.PageReq
	Category string `json:"category" dc:"分类筛选"`
	Module   string `json:"module" dc:"模块筛选"`
	Status   string `json:"status" dc:"状态筛选"`
	Keyword  string `json:"keyword" dc:"关键字搜索"`
}

type TestCaseListRes struct {
	g.Meta      `mime:"application/json"`
	commonApi.ListRes
	List []TestCaseItem `json:"list"`
}

type TestCaseItem struct {
	Id             int    `json:"id"`
	ProjectId      int    `json:"projectId"`
	RequirementId  int    `json:"requirementId"`
	TaskId         int    `json:"taskId"`
	Title          string `json:"title"`
	Preconditions  string `json:"preconditions"`
	Steps          string `json:"steps"`
	ExpectedResult string `json:"expectedResult"`
	Category       string `json:"category"`
	Module         string `json:"module"`
	Priority       string `json:"priority"`
	Source         string `json:"source"`
	CreatorId      int    `json:"creatorId"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	CreatorName    string `json:"creatorName,omitempty"`
}

type TestCaseDetailReq struct {
	g.Meta `path:"/test-cases/{id}" method:"get" tags:"测试管理" summary:"用例详情"`
	Id     int `json:"-" in:"path" v:"required#用例ID不能为空"`
}

type TestCaseDetailRes struct {
	g.Meta `mime:"application/json"`
	TestCaseItem
}

// ==================== 测试计划 ====================

type TestPlanCreateReq struct {
	g.Meta      `path:"/projects/{projectId}/test-plans" method:"post" tags:"测试管理" summary:"创建测试计划"`
	ProjectId   int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	Name        string `json:"name" v:"required#计划名称不能为空"`
	Description string `json:"description" dc:"计划描述"`
	MilestoneId int    `json:"milestoneId" dc:"关联里程碑ID"`
}

type TestPlanCreateRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type TestPlanUpdateReq struct {
	g.Meta      `path:"/test-plans/{id}" method:"put" tags:"测试管理" summary:"更新测试计划"`
	Id          int    `json:"-" in:"path" v:"required#计划ID不能为空"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MilestoneId int    `json:"milestoneId"`
	Status      string `json:"status" v:"in:draft,active,completed#状态必须是draft/active/completed"`
}

type TestPlanUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TestPlanDeleteReq struct {
	g.Meta `path:"/test-plans/{id}" method:"delete" tags:"测试管理" summary:"删除测试计划"`
	Id     int `json:"-" in:"path" v:"required#计划ID不能为空"`
}

type TestPlanDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type TestPlanListReq struct {
	g.Meta    `path:"/projects/{projectId}/test-plans" method:"get" tags:"测试管理" summary:"计划列表"`
	ProjectId int `json:"-" in:"path" v:"required#项目ID不能为空"`
	commonApi.PageReq
	Status string `json:"status" dc:"状态筛选"`
}

type TestPlanListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []TestPlanItem `json:"list"`
}

type TestPlanItem struct {
	Id          int    `json:"id"`
	ProjectId   int    `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MilestoneId int    `json:"milestoneId"`
	Status      string `json:"status"`
	CreatorId   int    `json:"creatorId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	CreatorName string `json:"creatorName,omitempty"`
}

type TestPlanDetailReq struct {
	g.Meta `path:"/test-plans/{id}" method:"get" tags:"测试管理" summary:"计划详情"`
	Id     int `json:"-" in:"path" v:"required#计划ID不能为空"`
}

type TestPlanDetailRes struct {
	g.Meta `mime:"application/json"`
	TestPlanItem
}

// ==================== 测试执行 ====================

type TestPlanAddCaseReq struct {
	g.Meta     `path:"/test-plans/{id}/cases" method:"post" tags:"测试管理" summary:"添加用例到计划"`
	Id         int   `json:"-" in:"path" v:"required#计划ID不能为空" dc:"测试计划ID"`
	CaseIds    []int `json:"caseIds" v:"required#用例ID列表不能为空"`
	AssigneeId int   `json:"assigneeId" dc:"指派执行人ID"`
}

type TestPlanAddCaseRes struct {
	g.Meta `mime:"application/json"`
}

type TestCaseExecuteReq struct {
	g.Meta       `path:"/test-plan-cases/{id}/execute" method:"put" tags:"测试管理" summary:"执行用例"`
	Id           int    `json:"-" in:"path" v:"required#计划用例ID不能为空"`
	Status       string `json:"status" v:"required|in:passed,failed,blocked,skipped#执行状态不能为空|状态必须是passed/failed/blocked/skipped"`
	ActualResult string `json:"actualResult" dc:"实际结果"`
	BugTaskId    int    `json:"bugTaskId" dc:"关联缺陷任务ID"`
}

type TestCaseExecuteRes struct {
	g.Meta `mime:"application/json"`
}

type TestPlanResultsReq struct {
	g.Meta `path:"/test-plans/{id}/results" method:"get" tags:"测试管理" summary:"计划执行结果"`
	Id     int `json:"-" in:"path" v:"required#计划ID不能为空"`
}

type TestPlanResultsRes struct {
	g.Meta     `mime:"application/json"`
	Total      int                    `json:"total"`
	Passed     int                    `json:"passed"`
	Failed     int                    `json:"failed"`
	Blocked    int                    `json:"blocked"`
	Skipped    int                    `json:"skipped"`
	Pending    int                    `json:"pending"`
	Results    []TestPlanCaseResult   `json:"results"`
}

type TestPlanCaseResult struct {
	Id             int    `json:"id"`
	TestCaseId     int    `json:"testCaseId"`
	TestCaseTitle  string `json:"testCaseTitle"`
	AssigneeId     int    `json:"assigneeId"`
	AssigneeName   string `json:"assigneeName,omitempty"`
	Status         string `json:"status"`
	ActualResult   string `json:"actualResult"`
	BugTaskId      int    `json:"bugTaskId"`
	ExecutedAt     string `json:"executedAt"`
}
