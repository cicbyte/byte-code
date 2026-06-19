package project

import "github.com/gogf/gf/v2/frame/g"

// ==================== 项目 CRUD ====================

type ProjectCreateReq struct {
	g.Meta      `path:"/projects" method:"post" tags:"项目管理" summary:"创建项目"`
	ProductId   int    `json:"productId"`
	Name        string `json:"name" v:"required#项目名称不能为空"`
	Description string `json:"description"`
}

type ProjectCreateRes struct {
	Id int `json:"id"`
}

type ProjectUpdateReq struct {
	g.Meta      `path:"/projects/{id}" method:"put" tags:"项目管理" summary:"更新项目"`
	Id          int    `json:"id" v:"required" in:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	ProductId   int    `json:"productId"`
}

type ProjectUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type ProjectDeleteReq struct {
	g.Meta `path:"/projects/{id}" method:"delete" tags:"项目管理" summary:"删除项目"`
	Id     int `json:"id" v:"required" in:"path"`
}

type ProjectDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type ProjectListReq struct {
	g.Meta    `path:"/projects" method:"get" tags:"项目管理" summary:"项目列表"`
	Status    int    `json:"status" in:"query"`
	ProductId int    `json:"productId" in:"query"`
	Keyword   string `json:"keyword" in:"query"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"20"`
}

type ProjectListRes struct {
	List  []ProjectItem `json:"list"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type ProjectItem struct {
	Id          int    `json:"id"`
	ProductId   int    `json:"productId"`
	ProductName string `json:"productName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   int    `json:"createdBy"`
	CreatorName string `json:"creatorName"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ProjectDetailReq struct {
	g.Meta `path:"/projects/{id}" method:"get" tags:"项目管理" summary:"项目详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type ProjectDetailRes struct {
	ProjectItem
}

// ==================== 项目成员 ====================

type MemberAddReq struct {
	g.Meta    `path:"/projects/{projectId}/members" method:"post" tags:"项目成员" summary:"添加成员"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	UserId    int    `json:"userId" v:"required#用户ID不能为空"`
	Role      string `json:"role" d:"member"`
}

type MemberAddRes struct {
	g.Meta `mime:"application/json"`
}

type MemberRemoveReq struct {
	g.Meta    `path:"/projects/{projectId}/members/{userId}" method:"delete" tags:"项目成员" summary:"移除成员"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
	UserId    int `json:"userId" v:"required" in:"path"`
}

type MemberRemoveRes struct {
	g.Meta `mime:"application/json"`
}

type MemberListReq struct {
	g.Meta    `path:"/projects/{projectId}/members" method:"get" tags:"项目成员" summary:"成员列表"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type MemberItem struct {
	Id       int    `json:"id"`
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type MemberListRes struct {
	List []MemberItem `json:"list"`
}

// ==================== 任务 CRUD ====================

type TaskCreateReq struct {
	g.Meta         `path:"/projects/{projectId}/tasks" method:"post" tags:"任务管理" summary:"创建任务"`
	ProjectId      int    `json:"projectId" v:"required" in:"path"`
	RequirementId  int    `json:"requirementId"`
	SprintId       int    `json:"sprintId"`
	Title          string `json:"title" v:"required#任务标题不能为空"`
	Description    string `json:"description"`
	Type           string `json:"type" d:"feature"`
	Priority       int    `json:"priority" d:"3"`
	AssigneeId     int    `json:"assigneeId"`
	ParentTaskId   int    `json:"parentTaskId"`
}

type TaskCreateRes struct {
	Id int `json:"id"`
}

type TaskUpdateReq struct {
	g.Meta         `path:"/tasks/{id}" method:"put" tags:"任务管理" summary:"更新任务"`
	Id             int    `json:"id" v:"required" in:"path"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Priority       int    `json:"priority"`
	AssigneeId     int    `json:"assigneeId"`
	SprintId       int    `json:"sprintId"`
	ParentTaskId   int    `json:"parentTaskId"`
	SortOrder      int    `json:"sortOrder"`
}

type TaskUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type TaskDeleteReq struct {
	g.Meta `path:"/tasks/{id}" method:"delete" tags:"任务管理" summary:"删除任务"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TaskDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type TaskListReq struct {
	g.Meta    `path:"/projects/{projectId}/tasks" method:"get" tags:"任务管理" summary:"任务列表"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Status    string `json:"status" in:"query"`
	Type      string `json:"type" in:"query"`
	SprintId  int    `json:"sprintId" in:"query"`
	AssigneeId int   `json:"assigneeId" in:"query"`
	Keyword   string `json:"keyword" in:"query"`
	Page      int    `json:"page" in:"query" d:"1"`
	Size      int    `json:"size" in:"query" d:"50"`
}

type TaskListRes struct {
	List  []TaskItem `json:"list"`
	Total int        `json:"total"`
}

type TaskItem struct {
	Id                  int    `json:"id"`
	ProjectId           int    `json:"projectId"`
	RequirementId       int    `json:"requirementId"`
	SprintId            int    `json:"sprintId"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Type                string `json:"type"`
	Status              string `json:"status"`
	Priority            int    `json:"priority"`
	AssigneeId          int    `json:"assigneeId"`
	AssigneeName        string `json:"assigneeName"`
	CreatorId           int    `json:"creatorId"`
	CreatorName         string `json:"creatorName"`
	ParentTaskId        int    `json:"parentTaskId"`
	Artifacts           string `json:"artifacts"`
	RequiresHumanReview int    `json:"requiresHumanReview"`
	HumanReviewStatus   string `json:"humanReviewStatus"`
	SortOrder           int    `json:"sortOrder"`
	Source              string `json:"source"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type TaskDetailReq struct {
	g.Meta `path:"/tasks/{id}" method:"get" tags:"任务管理" summary:"任务详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TaskDetailRes struct {
	TaskItem
}

// ==================== 任务特殊操作 ====================

type TaskClaimReq struct {
	g.Meta `path:"/tasks/{id}/claim" method:"post" tags:"任务管理" summary:"AI认领任务"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TaskClaimRes struct {
	g.Meta `mime:"application/json"`
}

type TaskCompleteReq struct {
	g.Meta    `path:"/tasks/{id}/complete" method:"post" tags:"任务管理" summary:"AI完成任务"`
	Id        int    `json:"id" v:"required" in:"path"`
	Artifacts string `json:"artifacts"`
}

type TaskCompleteRes struct {
	g.Meta `mime:"application/json"`
}

type TaskReviewReq struct {
	g.Meta   `path:"/tasks/{id}/review" method:"post" tags:"任务管理" summary:"审核任务"`
	Id       int    `json:"id" v:"required" in:"path"`
	Status   string `json:"status" v:"required|in:approved,rejected#审核状态不能为空|状态不合法"`
	Comment  string `json:"comment"`
}

type TaskReviewRes struct {
	g.Meta `mime:"application/json"`
}

type TaskImportReq struct {
	g.Meta        `path:"/projects/{projectId}/tasks/import" method:"post" tags:"任务管理" summary:"从需求导入任务"`
	ProjectId     int  `json:"projectId" v:"required" in:"path"`
	RequirementId int  `json:"requirementId" v:"required#需求ID不能为空"`
	SprintId      int  `json:"sprintId"`
}

type TaskImportRes struct {
	TaskIds []int `json:"taskIds"`
}

// ==================== 评论 ====================

type CommentCreateReq struct {
	g.Meta   `path:"/tasks/{taskId}/comments" method:"post" tags:"任务评论" summary:"创建评论"`
	TaskId   int    `json:"taskId" v:"required" in:"path"`
	Content  string `json:"content" v:"required#评论内容不能为空"`
	UserType string `json:"userType" d:"human"`
}

type CommentCreateRes struct {
	Id int `json:"id"`
}

type CommentListReq struct {
	g.Meta `path:"/tasks/{taskId}/comments" method:"get" tags:"任务评论" summary:"评论列表"`
	TaskId int `json:"taskId" v:"required" in:"path"`
}

type CommentItem struct {
	Id        int    `json:"id"`
	TaskId    int    `json:"taskId"`
	UserId    int    `json:"userId"`
	Username  string `json:"username"`
	RealName  string `json:"realName"`
	Content   string `json:"content"`
	UserType  string `json:"userType"`
	CreatedAt string `json:"createdAt"`
}

type CommentListRes struct {
	List []CommentItem `json:"list"`
}

// ==================== AI 执行日志 ====================

type AiLogCreateReq struct {
	g.Meta    `path:"/tasks/{taskId}/ai-logs" method:"post" tags:"AI执行日志" summary:"创建AI执行日志"`
	TaskId    int    `json:"taskId" v:"required" in:"path"`
	AiUserId  int    `json:"aiUserId" v:"required#AI用户ID不能为空"`
	Action    string `json:"action" v:"required#操作不能为空"`
	Detail    string `json:"detail"`
	Status    string `json:"status" d:"success"`
}

type AiLogCreateRes struct {
	Id int `json:"id"`
}

type AiLogListReq struct {
	g.Meta `path:"/tasks/{taskId}/ai-logs" method:"get" tags:"AI执行日志" summary:"AI执行日志列表"`
	TaskId int `json:"taskId" v:"required" in:"path"`
}

type AiLogItem struct {
	Id        int    `json:"id"`
	TaskId    int    `json:"taskId"`
	AiUserId  int    `json:"aiUserId"`
	Action    string `json:"action"`
	Detail    string `json:"detail"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type AiLogListRes struct {
	List []AiLogItem `json:"list"`
}

// ==================== Sprint CRUD ====================

type SprintCreateReq struct {
	g.Meta    `path:"/projects/{projectId}/sprints" method:"post" tags:"Sprint" summary:"创建Sprint"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Name      string `json:"name" v:"required#名称不能为空"`
	Goal      string `json:"goal"`
	StartDate string `json:"startDate" v:"required#开始日期不能为空"`
	EndDate   string `json:"endDate" v:"required#结束日期不能为空"`
}

type SprintCreateRes struct {
	Id int `json:"id"`
}

type SprintUpdateReq struct {
	g.Meta    `path:"/sprints/{id}" method:"put" tags:"Sprint" summary:"更新Sprint"`
	Id        int    `json:"id" v:"required" in:"path"`
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Status    string `json:"status"`
}

type SprintUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type SprintDeleteReq struct {
	g.Meta `path:"/sprints/{id}" method:"delete" tags:"Sprint" summary:"删除Sprint"`
	Id     int `json:"id" v:"required" in:"path"`
}

type SprintDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type SprintListReq struct {
	g.Meta    `path:"/projects/{projectId}/sprints" method:"get" tags:"Sprint" summary:"Sprint列表"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Status    string `json:"status" in:"query"`
}

type SprintItem struct {
	Id        int    `json:"id"`
	ProjectId int    `json:"projectId"`
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type SprintListRes struct {
	List []SprintItem `json:"list"`
}

type SprintDetailReq struct {
	g.Meta `path:"/sprints/{id}" method:"get" tags:"Sprint" summary:"Sprint详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type SprintDetailRes struct {
	SprintItem
}

// Sprint 任务管理
type SprintTaskAddReq struct {
	g.Meta    `path:"/sprints/{sprintId}/tasks" method:"post" tags:"Sprint" summary:"添加任务到Sprint"`
	SprintId  int `json:"sprintId" v:"required" in:"path"`
	TaskId    int `json:"taskId" v:"required#任务ID不能为空"`
}

type SprintTaskAddRes struct {
	g.Meta `mime:"application/json"`
}

type SprintTaskRemoveReq struct {
	g.Meta    `path:"/sprints/{sprintId}/tasks/{taskId}" method:"delete" tags:"Sprint" summary:"从Sprint移除任务"`
	SprintId  int `json:"sprintId" v:"required" in:"path"`
	TaskId    int `json:"taskId" v:"required" in:"path"`
}

type SprintTaskRemoveRes struct {
	g.Meta `mime:"application/json"`
}

// 燃尽图数据
type BurndownReq struct {
	g.Meta   `path:"/sprints/{id}/burndown" method:"get" tags:"Sprint" summary:"燃尽图数据"`
	Id       int `json:"id" v:"required" in:"path"`
}

type BurndownItem struct {
	Date      string `json:"date"`
	Remaining int    `json:"remaining"`
	Completed int    `json:"completed"`
}

type BurndownRes struct {
	Items []BurndownItem `json:"items"`
}
