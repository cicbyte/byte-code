package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/service"
)

var ProjectCtrl = projectController{}

type projectController struct {
	BaseController
}

// ==================== 项目 CRUD ====================

func (c *projectController) CreateProject(ctx context.Context, req *api.ProjectCreateReq) (res *api.ProjectCreateRes, err error) {
	id, err := service.Project().CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.ProjectCreateRes{Id: id}, nil
}

func (c *projectController) UpdateProject(ctx context.Context, req *api.ProjectUpdateReq) (res *api.ProjectUpdateRes, err error) {
	err = service.Project().UpdateProject(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.ProjectUpdateRes{}, nil
}

func (c *projectController) DeleteProject(ctx context.Context, req *api.ProjectDeleteReq) (res *api.ProjectDeleteRes, err error) {
	err = service.Project().DeleteProject(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.ProjectDeleteRes{}, nil
}

func (c *projectController) ProjectDetail(ctx context.Context, req *api.ProjectDetailReq) (res *api.ProjectDetailRes, err error) {
	return service.Project().GetProject(ctx, req.Id)
}

func (c *projectController) ProjectList(ctx context.Context, req *api.ProjectListReq) (res *api.ProjectListRes, err error) {
	return service.Project().ListProjects(ctx, req)
}

// ==================== 项目成员 ====================

func (c *projectController) AddMember(ctx context.Context, req *api.MemberAddReq) (res *api.MemberAddRes, err error) {
	err = service.Project().AddMember(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.MemberAddRes{}, nil
}

func (c *projectController) RemoveMember(ctx context.Context, req *api.MemberRemoveReq) (res *api.MemberRemoveRes, err error) {
	err = service.Project().RemoveMember(ctx, req.ProjectId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &api.MemberRemoveRes{}, nil
}

func (c *projectController) MemberList(ctx context.Context, req *api.MemberListReq) (res *api.MemberListRes, err error) {
	return service.Project().ListMembers(ctx, req.ProjectId)
}

// ==================== 任务 CRUD ====================

func (c *projectController) CreateTask(ctx context.Context, req *api.TaskCreateReq) (res *api.TaskCreateRes, err error) {
	id, err := service.Project().CreateTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskCreateRes{Id: id}, nil
}

func (c *projectController) UpdateTask(ctx context.Context, req *api.TaskUpdateReq) (res *api.TaskUpdateRes, err error) {
	err = service.Project().UpdateTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskUpdateRes{}, nil
}

func (c *projectController) DeleteTask(ctx context.Context, req *api.TaskDeleteReq) (res *api.TaskDeleteRes, err error) {
	err = service.Project().DeleteTask(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.TaskDeleteRes{}, nil
}

func (c *projectController) TaskDetail(ctx context.Context, req *api.TaskDetailReq) (res *api.TaskDetailRes, err error) {
	return service.Project().GetTask(ctx, req.Id)
}

func (c *projectController) TaskList(ctx context.Context, req *api.TaskListReq) (res *api.TaskListRes, err error) {
	return service.Project().ListTasks(ctx, req)
}

// ==================== 任务特殊操作 ====================

func (c *projectController) ClaimTask(ctx context.Context, req *api.TaskClaimReq) (res *api.TaskClaimRes, err error) {
	err = service.Project().ClaimTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskClaimRes{}, nil
}

func (c *projectController) CompleteTask(ctx context.Context, req *api.TaskCompleteReq) (res *api.TaskCompleteRes, err error) {
	err = service.Project().CompleteTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskCompleteRes{}, nil
}

func (c *projectController) ReviewTask(ctx context.Context, req *api.TaskReviewReq) (res *api.TaskReviewRes, err error) {
	err = service.Project().ReviewTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskReviewRes{}, nil
}

func (c *projectController) ImportTasks(ctx context.Context, req *api.TaskImportReq) (res *api.TaskImportRes, err error) {
	taskIds, err := service.Project().ImportTasks(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskImportRes{TaskIds: taskIds}, nil
}

// ==================== 评论 ====================

func (c *projectController) CreateComment(ctx context.Context, req *api.CommentCreateReq) (res *api.CommentCreateRes, err error) {
	id, err := service.Project().CreateComment(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.CommentCreateRes{Id: id}, nil
}

func (c *projectController) CommentList(ctx context.Context, req *api.CommentListReq) (res *api.CommentListRes, err error) {
	return service.Project().ListComments(ctx, req.TaskId)
}

// ==================== AI 执行日志 ====================

func (c *projectController) CreateAiLog(ctx context.Context, req *api.AiLogCreateReq) (res *api.AiLogCreateRes, err error) {
	id, err := service.Project().CreateAiLog(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.AiLogCreateRes{Id: id}, nil
}

func (c *projectController) AiLogList(ctx context.Context, req *api.AiLogListReq) (res *api.AiLogListRes, err error) {
	return service.Project().ListAiLogs(ctx, req.TaskId)
}

// ==================== Sprint CRUD ====================

func (c *projectController) CreateSprint(ctx context.Context, req *api.SprintCreateReq) (res *api.SprintCreateRes, err error) {
	id, err := service.Project().CreateSprint(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.SprintCreateRes{Id: id}, nil
}

func (c *projectController) UpdateSprint(ctx context.Context, req *api.SprintUpdateReq) (res *api.SprintUpdateRes, err error) {
	err = service.Project().UpdateSprint(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.SprintUpdateRes{}, nil
}

func (c *projectController) DeleteSprint(ctx context.Context, req *api.SprintDeleteReq) (res *api.SprintDeleteRes, err error) {
	err = service.Project().DeleteSprint(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.SprintDeleteRes{}, nil
}

func (c *projectController) SprintDetail(ctx context.Context, req *api.SprintDetailReq) (res *api.SprintDetailRes, err error) {
	return service.Project().GetSprint(ctx, req.Id)
}

func (c *projectController) SprintList(ctx context.Context, req *api.SprintListReq) (res *api.SprintListRes, err error) {
	return service.Project().ListSprints(ctx, req)
}

// ==================== Sprint 任务管理 ====================

func (c *projectController) AddTaskToSprint(ctx context.Context, req *api.SprintTaskAddReq) (res *api.SprintTaskAddRes, err error) {
	err = service.Project().AddTaskToSprint(ctx, req.SprintId, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &api.SprintTaskAddRes{}, nil
}

func (c *projectController) RemoveTaskFromSprint(ctx context.Context, req *api.SprintTaskRemoveReq) (res *api.SprintTaskRemoveRes, err error) {
	err = service.Project().RemoveTaskFromSprint(ctx, req.SprintId, req.TaskId)
	if err != nil {
		return nil, err
	}
	return &api.SprintTaskRemoveRes{}, nil
}

// ==================== 燃尽图 ====================

func (c *projectController) Burndown(ctx context.Context, req *api.BurndownReq) (res *api.BurndownRes, err error) {
	return service.Project().GetBurndown(ctx, req.Id)
}
