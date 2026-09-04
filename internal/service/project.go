package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/project"
)

type IProject interface {
	// 项目 CRUD
	CreateProject(ctx context.Context, req *api.ProjectCreateReq) (id int, err error)
	UpdateProject(ctx context.Context, req *api.ProjectUpdateReq) (err error)
	DeleteProject(ctx context.Context, id int) (err error)
	GetProject(ctx context.Context, id int) (res *api.ProjectDetailRes, err error)
	ListProjects(ctx context.Context, req *api.ProjectListReq) (res *api.ProjectListRes, err error)

	// 项目成员
	AddMember(ctx context.Context, req *api.MemberAddReq) (err error)
	RemoveMember(ctx context.Context, projectId, userId int) (err error)
	ListMembers(ctx context.Context, projectId int) (res *api.MemberListRes, err error)

	// 任务 CRUD
	CreateTask(ctx context.Context, req *api.TaskCreateReq) (id int, err error)
	UpdateTask(ctx context.Context, req *api.TaskUpdateReq) (err error)
	DeleteTask(ctx context.Context, id int) (err error)
	GetTask(ctx context.Context, id int) (res *api.TaskDetailRes, err error)
	ListTasks(ctx context.Context, req *api.TaskListReq) (res *api.TaskListRes, err error)
	MyTaskList(ctx context.Context, req *api.MyTaskListReq) (res *api.MyTaskListRes, err error)

	// 任务特殊操作
	ClaimTask(ctx context.Context, req *api.TaskClaimReq) (err error)
	CompleteTask(ctx context.Context, req *api.TaskCompleteReq) (err error)
	ReviewTask(ctx context.Context, req *api.TaskReviewReq) (err error)
	ImportTasks(ctx context.Context, req *api.TaskImportReq) (taskIds []int, err error)

	// 评论
	CreateComment(ctx context.Context, req *api.CommentCreateReq) (id int, err error)
	ListComments(ctx context.Context, taskId int) (res *api.CommentListRes, err error)
	UpdateComment(ctx context.Context, req *api.CommentUpdateReq) (err error)
	DeleteComment(ctx context.Context, id int) (err error)

	// AI 执行日志
	CreateAiLog(ctx context.Context, req *api.AiLogCreateReq) (id int, err error)
	ListAiLogs(ctx context.Context, taskId int) (res *api.AiLogListRes, err error)

	// Sprint CRUD
	CreateSprint(ctx context.Context, req *api.SprintCreateReq) (id int, err error)
	UpdateSprint(ctx context.Context, req *api.SprintUpdateReq) (err error)
	DeleteSprint(ctx context.Context, id int) (err error)
	GetSprint(ctx context.Context, id int) (res *api.SprintDetailRes, err error)
	ListSprints(ctx context.Context, req *api.SprintListReq) (res *api.SprintListRes, err error)

	// Sprint 任务管理
	AddTaskToSprint(ctx context.Context, sprintId, taskId int) (err error)
	RemoveTaskFromSprint(ctx context.Context, sprintId, taskId int) (err error)

	// 燃尽图
	GetBurndown(ctx context.Context, sprintId int) (res *api.BurndownRes, err error)

	// 需求 CRUD
	CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (id int, err error)
	UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (err error)
	DeleteRequirement(ctx context.Context, id int) (err error)
	GetRequirement(ctx context.Context, id int) (res *api.RequirementDetailRes, err error)
	ListRequirements(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error)

	// 里程碑
	CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (id int, err error)
	ListMilestones(ctx context.Context, projectId int) (res *api.MilestoneListRes, err error)
	UpdateMilestone(ctx context.Context, req *api.MilestoneUpdateReq) (err error)
	DeleteMilestone(ctx context.Context, id int) (err error)
}

var localProject IProject

func Project() IProject {
	if localProject == nil {
		panic("implement not found for interface IProject, forgot register?")
	}
	return localProject
}

func RegisterProject(i IProject) {
	localProject = i
}
