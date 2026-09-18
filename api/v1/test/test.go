package test

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 测试用例 ====================

type TestCaseCreateReq struct {
	g.Meta         `path:"/projects/{projectId}/test-cases" method:"post" tags:"测试管理" summary:"创建测试用例"`
	ProjectId      int    `json:"-" in:"path" v:"required#项目ID不能为空"`
	RequirementId  int    `json:"requirementId" dc:"关联需求ID"`
	TaskId         int    `json:"taskId" dc:"关联任务ID"`
	Title          string `json:"title" v:"required#用例标题不能为空"`
	Preconditions  string `json:"preconditions" dc:"前置条件"`
	Steps          string `json:"steps" dc:"测试步骤"`
	ExpectedResult string `json:"expectedResult" dc:"预期结果"`
	Category       string `json:"category" dc:"分类:功能/性能/安全/兼容性"`
	Module         string `json:"module" dc:"所属模块"`
	Priority       string `json:"priority" v:"required|in:P0,P1,P2,P3#优先级不能为空|优先级必须是P0/P1/P2/P3"`
	Source         string `json:"source" dc:"来源:human/ai_generated" d:"human" v:"in:human,ai_generated#来源必须是human/ai_generated"`
	ExternalKey    string `json:"externalKey" dc:"外部键（pytest nodeid），--bcode-sync 幂等依据"`
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
	ExternalKey    string `json:"externalKey" dc:"外部键（pytest nodeid）——CLI cases push 同步依据"`
	Status         string `json:"status" v:"in:active,deprecated#状态必须是active/deprecated"`
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
	ProjectId int `json:"-" in:"path" v:"required#项目ID不能为空"`
	commonApi.PageReq
	Category    string `json:"category" dc:"分类筛选"`
	Module      string `json:"module" dc:"模块筛选"`
	Status      string `json:"status" dc:"状态筛选"`
	Keyword     string `json:"keyword" dc:"关键字搜索"`
	ExternalKey string `json:"externalKey" dc:"外部键精确匹配（pytest nodeid，--bcode-sync 幂等查找）"`
}

type TestCaseListRes struct {
	g.Meta `mime:"application/json"`
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
	ExternalKey    string `json:"externalKey,omitempty" dc:"外部键（pytest nodeid）"`
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
	Status      string `json:"status" v:"in:draft,running,completed#状态必须是draft/running/completed"`
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
	Status       string `json:"status" v:"required|in:pass,fail,blocked,skip#执行状态不能为空|状态必须是pass/fail/blocked/skip"`
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
	g.Meta  `mime:"application/json"`
	Total   int                  `json:"total"`
	Passed  int                  `json:"passed"`
	Failed  int                  `json:"failed"`
	Blocked int                  `json:"blocked"`
	Skipped int                  `json:"skipped"`
	Pending int                  `json:"pending"`
	Results []TestPlanCaseResult `json:"results"`
}

type TestPlanCaseResult struct {
	Id            int    `json:"id"`
	TestCaseId    int    `json:"testCaseId"`
	TestCaseTitle string `json:"testCaseTitle"`
	AssigneeId    int    `json:"assigneeId"`
	AssigneeName  string `json:"assigneeName,omitempty"`
	Status        string `json:"status"`
	ActualResult  string `json:"actualResult"`
	BugTaskId     int    `json:"bugTaskId"`
	ExecutedAt    string `json:"executedAt"`
}

// ==================== 测试执行记录（Run，#504 pytest P1） ====================

// TestRunCaseReport 单用例上报项：pytest 侧逐条结果
type TestRunCaseReport struct {
	TestCaseId  int    `json:"testCaseId" dc:"映射的平台用例ID（@pytest.mark.bytecode(case=N)，0=未映射"`
	ExternalKey string `json:"externalKey" dc:"外部键：pytest nodeid"`
	Title       string `json:"title" dc:"用例标题（缺省同 externalKey）"`
	Status      string `json:"status" v:"required|in:pass,fail,error,skip#状态不能为空|状态必须是pass/fail/error/skip"`
	DurationMs  int    `json:"durationMs" dc:"耗时毫秒"`
	Message     string `json:"message" dc:"失败信息（截断 traceback），服务端限长"`
}

type TestRunReportReq struct {
	g.Meta     `path:"/projects/{projectId}/test-runs" method:"post" tags:"测试管理" summary:"上报测试执行记录"`
	ProjectId  int                 `json:"-" in:"path" v:"required#项目ID不能为空"`
	Source     string              `json:"source" d:"pytest" v:"in:manual,pytest,ci,junit#来源必须是manual/pytest/ci/junit"`
	Branch     string              `json:"branch" dc:"git 分支"`
	GitSha     string              `json:"gitSha" dc:"git commit"`
	Env        string              `json:"env" dc:"环境标识（local/ci 等）"`
	StartedAt  string              `json:"startedAt" dc:"开始时间 YYYY-MM-DD HH:MM:SS，空则由服务端推导"`
	FinishedAt string              `json:"finishedAt" dc:"结束时间，空则取当前"`
	DurationMs int                 `json:"durationMs" dc:"总耗时毫秒（缺省用起止差推导）"`
	Cases      []TestRunCaseReport `json:"cases" v:"required#用例结果不能为空"`
}

type TestRunReportRes struct {
	g.Meta `mime:"application/json"`
	Id     int `json:"id"`
}

type TestRunListReq struct {
	g.Meta    `path:"/projects/{projectId}/test-runs" method:"get" tags:"测试管理" summary:"执行记录列表"`
	ProjectId int `json:"-" in:"path" v:"required#项目ID不能为空"`
	commonApi.PageReq
	Source string `json:"source" dc:"来源筛选"`
	Status string `json:"status" dc:"结果筛选:pass/fail（failed+errors>0 即 fail）"`
	Branch string `json:"branch" dc:"分支筛选"`
}

type TestRunListRes struct {
	g.Meta `mime:"application/json"`
	commonApi.ListRes
	List []TestRunItem `json:"list"`
}

type TestRunItem struct {
	Id              int    `json:"id"`
	ProjectId       int    `json:"projectId"`
	Source          string `json:"source"`
	Branch          string `json:"branch"`
	GitSha          string `json:"gitSha"`
	Env             string `json:"env"`
	TriggeredBy     int    `json:"triggeredBy"`
	TriggeredByName string `json:"triggeredByName,omitempty"`
	Total           int    `json:"total"`
	Passed          int    `json:"passed"`
	Failed          int    `json:"failed"`
	Skipped         int    `json:"skipped"`
	Errors          int    `json:"errors"`
	DurationMs      int    `json:"durationMs"`
	StartedAt       string `json:"startedAt"`
	FinishedAt      string `json:"finishedAt"`
	CreatedAt       string `json:"createdAt"`
}

type TestRunDetailReq struct {
	g.Meta `path:"/test-runs/{id}" method:"get" tags:"测试管理" summary:"执行记录详情"`
	Id     int `json:"-" in:"path" v:"required#记录ID不能为空"`
}

type TestRunDetailRes struct {
	g.Meta `mime:"application/json"`
	TestRunItem
	Cases []TestRunCaseItem `json:"cases"`
}

type TestRunCaseItem struct {
	Id            int    `json:"id"`
	TestRunId     int    `json:"testRunId"`
	TestCaseId    int    `json:"testCaseId"`
	TestCaseTitle string `json:"testCaseTitle,omitempty" dc:"映射的平台用例标题（未映射为空）"`
	ExternalKey   string `json:"externalKey"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	DurationMs    int    `json:"durationMs"`
	Message       string `json:"message"`
}

type TestRunDeleteReq struct {
	g.Meta `path:"/test-runs/{id}" method:"delete" tags:"测试管理" summary:"删除执行记录"`
	Id     int `json:"-" in:"path" v:"required#记录ID不能为空"`
}

type TestRunDeleteRes struct {
	g.Meta `mime:"application/json"`
}
