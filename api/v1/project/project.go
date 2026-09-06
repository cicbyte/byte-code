package project

import "github.com/gogf/gf/v2/frame/g"

// ==================== 项目 CRUD ====================

type ProjectCreateReq struct {
	g.Meta      `path:"/projects" method:"post" tags:"项目管理" summary:"创建项目"`
	Name        string `json:"name" v:"required#项目名称不能为空"`
	Description string `json:"description"`
}

type ProjectCreateRes struct {
	Id int `json:"id"`
}

type ProjectUpdateReq struct {
	g.Meta      `path:"/projects/{id}" method:"put" tags:"项目管理" summary:"更新项目"`
	Id          int     `json:"id" v:"required" in:"path"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Status      *int    `json:"status"`
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
	g.Meta  `path:"/projects" method:"get" tags:"项目管理" summary:"项目列表"`
	Status  int    `json:"status" in:"query"`
	Keyword string `json:"keyword" in:"query"`
	Page    int    `json:"page" in:"query" d:"1" v:"min:1#页码从1开始"`
	Size    int    `json:"size" in:"query" d:"20" v:"max:100#每页上限100"`
}

type ProjectListRes struct {
	List  []ProjectItem `json:"list"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type ProjectItem struct {
	Id          int    `json:"id"`
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
	Id         int    `json:"id"`
	UserId     int    `json:"userId"`
	Username   string `json:"username"`
	RealName   string `json:"realName"`
	Role       string `json:"role"`
	JoinedAt   string `json:"joinedAt"`
	UserType   string `json:"userType" dc:"human/ai"`
	ViaBinding int    `json:"viaBinding" dc:"1=来自 agent 项目准入（移除走 DeleteAgentProject）；0=project_members"`
}

type MemberListRes struct {
	List []MemberItem `json:"list"`
}

// ==================== 任务 CRUD ====================

type TaskCreateReq struct {
	g.Meta        `path:"/projects/{projectId}/tasks" method:"post" tags:"任务管理" summary:"创建任务"`
	ProjectId     int    `json:"projectId" v:"required" in:"path"`
	RequirementId int    `json:"requirementId"`
	SprintId      int    `json:"sprintId"`
	Title         string `json:"title" v:"required|max-length:255#任务标题不能为空|上限255字"`
	Description   string `json:"description"`
	Type          string `json:"type" d:"feature"`
	Priority      int    `json:"priority" d:"3"`
	AssigneeId    int    `json:"assigneeId"`
	ParentTaskId  int    `json:"parentTaskId"`
	DueDate       string `json:"dueDate" dc:"截止日期（Y-m-d），缺省无截止；格式在 logic 层校验"`
	Checklist     string `json:"checklist" dc:"步骤清单 JSON 数组 [{text,done}]，缺省空清单"`
}

type TaskCreateRes struct {
	Id int `json:"id"`
}

type TaskUpdateReq struct {
	g.Meta      `path:"/tasks/{id}" method:"put" tags:"任务管理" summary:"更新任务"`
	Id          int     `json:"id" v:"required" in:"path"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	// 枚举校验防脏值入库；指针为 nil 时 gf 跳过校验（不更新语义不受影响）
	Status *string `json:"status" v:"in:open,in_progress,blocked,review,done,closed"`
	// 指针字段：nil=不更新，非nil零值=显式清空
	Priority     *int    `json:"priority"`
	AssigneeId   *int    `json:"assigneeId"`
	SprintId     *int    `json:"sprintId"`
	ParentTaskId *int    `json:"parentTaskId"`
	SortOrder    *int    `json:"sortOrder"`
	DueDate      *string `json:"dueDate" dc:"截止日期（Y-m-d）；空串=清除"`
	Checklist    *string `json:"checklist" dc:"步骤清单 JSON [{text,done}]；打勾=进展（顺带续租约）"`
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
	g.Meta     `path:"/projects/{projectId}/tasks" method:"get" tags:"任务管理" summary:"任务列表"`
	ProjectId  int    `json:"projectId" v:"required" in:"path"`
	Status     string `json:"status" in:"query"`
	Type       string `json:"type" in:"query"`
	SprintId   int    `json:"sprintId" in:"query"`
	AssigneeId int    `json:"assigneeId" in:"query"`
	Keyword    string `json:"keyword" in:"query"`
	TagId      int    `json:"tagId" in:"query" dc:"按标签筛选"`
	Page       int    `json:"page" in:"query" d:"1" v:"min:1#页码从1开始"`
	Size       int    `json:"size" in:"query" d:"50" v:"max:200#每页上限200"`
}

type TaskListRes struct {
	List  []TaskItem `json:"list"`
	Total int        `json:"total"`
}

// ==================== 我的任务（跨项目聚合） ====================

type MyTaskListReq struct {
	g.Meta    `path:"/my-tasks" method:"get" tags:"任务管理" summary:"我的任务（当前用户跨项目聚合）"`
	Status    string `json:"status" in:"query" dc:"缺省=活跃四态(open,in_progress,blocked,review)；all=全部；或逗号分隔状态列表"`
	ProjectId int    `json:"projectId" in:"query" dc:"按项目过滤"`
	Keyword   string `json:"keyword" in:"query"`
	Page      int    `json:"page" in:"query" d:"1" v:"min:1#页码从1开始"`
	Size      int    `json:"size" in:"query" d:"50" v:"max:200#每页上限200"`
}

type MyTaskItem struct {
	TaskItem
	ProjectName string `json:"projectName"`
}

type MyTaskListRes struct {
	List  []MyTaskItem `json:"list"`
	Total int          `json:"total"`
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
	DueDate             string `json:"dueDate"`
	RequiresHumanReview int    `json:"requiresHumanReview"`
	HumanReviewStatus   string `json:"humanReviewStatus"`
	// 步骤清单 JSON：[{"text":"...","done":false}]；打勾走 UpdateTask
	// （触发 updated_at，构成租约心跳）
	Checklist string   `json:"checklist"`
	SortOrder int      `json:"sortOrder"`
	Source    string   `json:"source"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
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

type TaskBlockReq struct {
	g.Meta `path:"/tasks/{id}/block" method:"post" tags:"任务管理" summary:"上报阻塞（assignee/owner；in_progress→blocked）"`
	Id     int    `json:"id" v:"required" in:"path"`
	Reason string `json:"reason" v:"required#阻塞原因不能为空"`
}

type TaskBlockRes struct {
	g.Meta `mime:"application/json"`
}

type TaskUnblockReq struct {
	g.Meta `path:"/tasks/{id}/unblock" method:"post" tags:"任务管理" summary:"解除阻塞（assignee/owner；blocked→in_progress）"`
	Id     int `json:"id" v:"required" in:"path"`
}

type TaskUnblockRes struct {
	g.Meta `mime:"application/json"`
}

type TaskCompleteRes struct {
	g.Meta `mime:"application/json"`
}

type TaskReviewReq struct {
	g.Meta  `path:"/tasks/{id}/review" method:"post" tags:"任务管理" summary:"审核任务"`
	Id      int    `json:"id" v:"required" in:"path"`
	Status  string `json:"status" v:"required|in:approved,rejected#审核状态不能为空|状态不合法"`
	Comment string `json:"comment"`
}

type TaskReviewRes struct {
	g.Meta `mime:"application/json"`
}

type TaskImportReq struct {
	g.Meta        `path:"/projects/{projectId}/tasks/import" method:"post" tags:"任务管理" summary:"从需求导入任务"`
	ProjectId     int `json:"projectId" v:"required" in:"path"`
	RequirementId int `json:"requirementId" v:"required#需求ID不能为空"`
	SprintId      int `json:"sprintId"`
}

type TaskImportRes struct {
	TaskIds []int `json:"taskIds"`
}

// ==================== 评论 ====================

type CommentCreateReq struct {
	g.Meta   `path:"/tasks/{taskId}/comments" method:"post" tags:"任务评论" summary:"创建评论"`
	TaskId   int    `json:"taskId" v:"required" in:"path"`
	ParentId int    `json:"parentId"`
	Content  string `json:"content" v:"required|max-length:16384#评论内容不能为空|上限16KB"`
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
	ParentId  int    `json:"parentId"`
	UserId    int    `json:"userId"`
	Username  string `json:"username"`
	RealName  string `json:"realName"`
	Content   string `json:"content"`
	UserType  string `json:"userType"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type CommentListRes struct {
	List []CommentItem `json:"list"`
}

type CommentUpdateReq struct {
	g.Meta  `path:"comments/{id}" method:"put" tags:"任务评论" summary:"编辑评论"`
	Id      int     `json:"id" v:"required" in:"path"`
	Content *string `json:"content" v:"required|max-length:16384#评论内容不能为空|上限16KB"`
}

type CommentUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type CommentDeleteReq struct {
	g.Meta `path:"comments/{id}" method:"delete" tags:"任务评论" summary:"删除评论"`
	Id     int `json:"id" v:"required" in:"path"`
}

type CommentDeleteRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== AI 执行日志 ====================

type AiLogCreateReq struct {
	g.Meta   `path:"/tasks/{taskId}/ai-logs" method:"post" tags:"AI执行日志" summary:"创建AI执行日志"`
	TaskId   int    `json:"taskId" v:"required" in:"path"`
	AiUserId int    `json:"aiUserId" v:"required#AI用户ID不能为空"`
	Action   string `json:"action" v:"required#操作不能为空"`
	Detail   string `json:"detail"`
	Status   string `json:"status" d:"success"`
}

type AiLogCreateRes struct {
	Id int `json:"id"`
}

type AiLogListReq struct {
	g.Meta `path:"/tasks/{taskId}/ai-logs" method:"get" tags:"AI执行日志" summary:"AI执行日志列表"`
	TaskId int `json:"taskId" v:"required" in:"path"`
}

type AiLogItem struct {
	Id         int    `json:"id"`
	TaskId     int    `json:"taskId"`
	AiUserId   int    `json:"aiUserId"`
	AiUsername string `json:"aiUsername" dc:"执行 AI 用户名（回填）"`
	Action     string `json:"action"`
	Detail     string `json:"detail"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
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
	Id        int     `json:"id" v:"required" in:"path"`
	Name      *string `json:"name"`
	Goal      *string `json:"goal"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	Status    *string `json:"status"`
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
	g.Meta   `path:"/sprints/{sprintId}/tasks" method:"post" tags:"Sprint" summary:"添加任务到Sprint"`
	SprintId int `json:"sprintId" v:"required" in:"path"`
	TaskId   int `json:"taskId" v:"required#任务ID不能为空"`
}

type SprintTaskAddRes struct {
	g.Meta `mime:"application/json"`
}

type SprintTaskRemoveReq struct {
	g.Meta   `path:"/sprints/{sprintId}/tasks/{taskId}" method:"delete" tags:"Sprint" summary:"从Sprint移除任务"`
	SprintId int `json:"sprintId" v:"required" in:"path"`
	TaskId   int `json:"taskId" v:"required" in:"path"`
}

type SprintTaskRemoveRes struct {
	g.Meta `mime:"application/json"`
}

// 燃尽图数据
type BurndownReq struct {
	g.Meta `path:"/sprints/{id}/burndown" method:"get" tags:"Sprint" summary:"燃尽图数据"`
	Id     int `json:"id" v:"required" in:"path"`
}

type BurndownItem struct {
	Date      string `json:"date"`
	Remaining int    `json:"remaining"`
	Completed int    `json:"completed"`
}

type BurndownRes struct {
	Items []BurndownItem `json:"items"`
}

// ==================== 需求 CRUD ====================

type RequirementCreateReq struct {
	g.Meta             `path:"/projects/{projectId}/requirements" method:"post" tags:"需求管理" summary:"创建需求"`
	ProjectId          int    `json:"projectId" v:"required" in:"path"`
	ParentId           int    `json:"parentId"`
	Type               string `json:"type" v:"required|in:epic,story,task#类型不能为空|类型不合法" d:"story"`
	Title              string `json:"title" v:"required#标题不能为空"`
	Description        string `json:"description"`
	Priority           int    `json:"priority" d:"3"`
	AssigneeId         int    `json:"assigneeId"`
	MilestoneId        int    `json:"milestoneId"`
	AcceptanceCriteria string `json:"acceptanceCriteria"`
}

type RequirementCreateRes struct {
	Id int `json:"id"`
}

type RequirementUpdateReq struct {
	g.Meta             `path:"/requirements/{id}" method:"put" tags:"需求管理" summary:"更新需求"`
	Id                 int     `json:"id" v:"required" in:"path"`
	Title              *string `json:"title"`
	Description        *string `json:"description"`
	Status             *string `json:"status"`
	Priority           *int    `json:"priority"`
	AssigneeId         *int    `json:"assigneeId"`
	MilestoneId        *int    `json:"milestoneId"`
	AcceptanceCriteria *string `json:"acceptanceCriteria"`
	SortOrder          *int    `json:"sortOrder"`
}

type RequirementUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type RequirementDeleteReq struct {
	g.Meta `path:"/requirements/{id}" method:"delete" tags:"需求管理" summary:"删除需求"`
	Id     int `json:"id" v:"required" in:"path"`
}

type RequirementDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type RequirementListReq struct {
	g.Meta    `path:"/projects/{projectId}/requirements" method:"get" tags:"需求管理" summary:"需求列表"`
	ProjectId int    `json:"projectId" v:"required" in:"path"`
	Type      string `json:"type" in:"query"`
	Status    string `json:"status" in:"query"`
	ParentId  int    `json:"parentId" in:"query"`
	Page      int    `json:"page" in:"query" d:"1" v:"min:1#页码从1开始"`
	Size      int    `json:"size" in:"query" d:"50" v:"max:200#每页上限200"`
}

type RequirementListRes struct {
	List  []RequirementItem `json:"list"`
	Total int               `json:"total"`
}

type RequirementItem struct {
	Id                 int               `json:"id"`
	ProjectId          int               `json:"projectId"`
	ParentId           int               `json:"parentId"`
	Type               string            `json:"type"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	Status             string            `json:"status"`
	Priority           int               `json:"priority"`
	AssigneeId         int               `json:"assigneeId"`
	AssigneeName       string            `json:"assigneeName"`
	CreatorId          int               `json:"creatorId"`
	CreatorName        string            `json:"creatorName"`
	MilestoneId        int               `json:"milestoneId"`
	SortOrder          int               `json:"sortOrder"`
	AcceptanceCriteria string            `json:"acceptanceCriteria"`
	Source             string            `json:"source"`
	Children           []RequirementItem `json:"children,omitempty"`
	CreatedAt          string            `json:"createdAt"`
	UpdatedAt          string            `json:"updatedAt"`
}

type RequirementDetailReq struct {
	g.Meta `path:"/requirements/{id}" method:"get" tags:"需求管理" summary:"需求详情"`
	Id     int `json:"id" v:"required" in:"path"`
}

type RequirementDetailRes struct {
	RequirementItem
}

// ==================== 里程碑 ====================

type MilestoneCreateReq struct {
	g.Meta      `path:"/projects/{projectId}/milestones" method:"post" tags:"里程碑" summary:"创建里程碑"`
	ProjectId   int    `json:"projectId" v:"required" in:"path"`
	Name        string `json:"name" v:"required#名称不能为空"`
	Description string `json:"description"`
	TargetDate  string `json:"targetDate"`
}

type MilestoneCreateRes struct {
	Id int `json:"id"`
}

type MilestoneListReq struct {
	g.Meta    `path:"/projects/{projectId}/milestones" method:"get" tags:"里程碑" summary:"里程碑列表"`
	ProjectId int `json:"projectId" v:"required" in:"path"`
}

type MilestoneUpdateReq struct {
	g.Meta      `path:"/milestones/{id}" method:"put" tags:"里程碑" summary:"更新里程碑"`
	Id          int     `json:"id" v:"required" in:"path"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TargetDate  *string `json:"targetDate"`
	Status      *string `json:"status" v:"in:planning,in_progress,released#状态必须是planning/in_progress/released"`
}

type MilestoneUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type MilestoneDeleteReq struct {
	g.Meta `path:"/milestones/{id}" method:"delete" tags:"里程碑" summary:"删除里程碑"`
	Id     int `json:"id" v:"required" in:"path"`
}

type MilestoneDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type MilestoneItem struct {
	Id          int    `json:"id"`
	ProjectId   int    `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TargetDate  string `json:"targetDate"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type MilestoneListRes struct {
	List []MilestoneItem `json:"list"`
}
