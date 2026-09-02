package consts

const (
	PageSize = 10 //分页长度
)

// ==================== 任务状态 ====================

const (
	TaskStatusOpen       = "open"
	TaskStatusInProgress = "in_progress"
	TaskStatusReview     = "review"
	TaskStatusDone       = "done"
	TaskStatusClosed     = "closed"
)

// TaskTerminalStatuses 终态（进入时记录 completed_at，从终态切回时清空）
var TaskTerminalStatuses = []string{TaskStatusDone, TaskStatusClosed}

// TaskActiveStatuses 非终态
var TaskActiveStatuses = []string{TaskStatusOpen, TaskStatusInProgress, TaskStatusReview}

// ==================== 项目状态 ====================

const (
	ProjectStatusActive   = 1
	ProjectStatusArchived = 2
)

// ==================== Sprint 状态 ====================

const (
	SprintStatusPlanning = "planning"
	SprintStatusActive    = "active"
	SprintStatusCompleted = "completed"
)

// ==================== 需求状态 ====================

const (
	ReqStatusDraft    = "draft"
	ReqStatusActive   = "active"
	ReqStatusArchived = "archived"
)

// ==================== 测试状态 ====================

const (
	TestPlanStatusPlanning = "planning"
	TestPlanStatusActive   = "active"
	TestPlanStatusDone     = "completed"

	TestCaseStatusDraft     = "draft"
	TestCaseStatusActive    = "active"
	TestCaseStatusDeprecate = "deprecated"

	TestPlanCasePending = "pending"
	TestPlanCasePass    = "pass"
	TestPlanCaseFail    = "fail"
	TestPlanCaseBlocked = "blocked"
	TestPlanCaseSkip    = "skip"
)

// ==================== 文档状态 ====================

const (
	DocStatusActive   = "active"
	DocStatusArchived = "archived"
)

// ==================== 成员角色 ====================

const (
	MemberRoleOwner  = "owner"
	MemberRoleMember = "member"
)
