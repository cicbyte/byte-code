package controller

import (
	"github.com/gogf/gf/v2/errors/gerror"
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

func (c *projectController) ExportProject(ctx context.Context, req *api.ProjectExportReq) (*api.ProjectExportRes, error) {
	return service.Project().ExportProject(ctx, req.ProjectId)
}

func (c *projectController) MyTaskList(ctx context.Context, req *api.MyTaskListReq) (res *api.MyTaskListRes, err error) {
	return service.Project().MyTaskList(ctx, req)
}

// ==================== 任务特殊操作 ====================

func (c *projectController) ClaimTask(ctx context.Context, req *api.TaskClaimReq) (res *api.TaskClaimRes, err error) {
	err = service.Project().ClaimTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskClaimRes{}, nil
}

func (c *projectController) ReleaseTask(ctx context.Context, req *api.TaskReleaseReq) (res *api.TaskReleaseRes, err error) {
	err = service.Project().ReleaseTask(ctx, req)
	res = new(api.TaskReleaseRes)
	return
}

func (c *projectController) ReopenTask(ctx context.Context, req *api.TaskReopenReq) (res *api.TaskReopenRes, err error) {
	err = service.Project().ReopenTask(ctx, req)
	res = new(api.TaskReopenRes)
	return
}

func (c *projectController) CreateGroup(ctx context.Context, req *api.GroupCreateReq) (*api.GroupCreateRes, error) {
	id, err := service.Project().CreateGroup(ctx, req)
	if err != nil {
		return nil, gerror.New(err.Error())
	}
	return &api.GroupCreateRes{Id: id}, nil
}

func (c *projectController) ListGroups(ctx context.Context, req *api.GroupListReq) (*api.GroupListRes, error) {
	return service.Project().ListGroups(ctx)
}

func (c *projectController) UpdateGroup(ctx context.Context, req *api.GroupUpdateReq) (*api.GroupUpdateRes, error) {
	err := service.Project().UpdateGroup(ctx, req)
	return &api.GroupUpdateRes{}, err
}

func (c *projectController) DeleteGroup(ctx context.Context, req *api.GroupDeleteReq) (*api.GroupDeleteRes, error) {
	err := service.Project().DeleteGroup(ctx, req.Id)
	return &api.GroupDeleteRes{}, err
}

func (c *projectController) AddProjectToGroup(ctx context.Context, req *api.GroupMemberAddReq) (*api.GroupMemberAddRes, error) {
	err := service.Project().AddProjectToGroup(ctx, req)
	return &api.GroupMemberAddRes{}, err
}

func (c *projectController) RemoveProjectFromGroup(ctx context.Context, req *api.GroupMemberRemoveReq) (*api.GroupMemberRemoveRes, error) {
	err := service.Project().RemoveProjectFromGroup(ctx, req.Id, req.ProjectId)
	return &api.GroupMemberRemoveRes{}, err
}

func (c *projectController) CompleteTask(ctx context.Context, req *api.TaskCompleteReq) (res *api.TaskCompleteRes, err error) {
	err = service.Project().CompleteTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskCompleteRes{}, nil
}

func (c *projectController) ListRelations(ctx context.Context, req *api.RelationListReq) (res *api.RelationListRes, err error) {
	return service.Project().ListRelations(ctx, req.ProjectId)
}

func (c *projectController) AddRelation(ctx context.Context, req *api.RelationAddReq) (res *api.RelationAddRes, err error) {
	err = service.Project().AddRelation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.RelationAddRes{}, nil
}

func (c *projectController) RemoveRelation(ctx context.Context, req *api.RelationRemoveReq) (res *api.RelationRemoveRes, err error) {
	err = service.Project().RemoveRelation(ctx, req.ProjectId, req.RelationId)
	if err != nil {
		return nil, err
	}
	return &api.RelationRemoveRes{}, nil
}

func (c *projectController) CreateFeedback(ctx context.Context, req *api.FeedbackCreateReq) (res *api.FeedbackCreateRes, err error) {
	id, err := service.Project().CreateFeedback(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.FeedbackCreateRes{Id: id}, nil
}

func (c *projectController) ListFeedbacks(ctx context.Context, req *api.FeedbackListReq) (res *api.FeedbackListRes, err error) {
	return service.Project().ListFeedbacks(ctx, req)
}

func (c *projectController) ConvertFeedback(ctx context.Context, req *api.FeedbackConvertReq) (res *api.FeedbackConvertRes, err error) {
	taskId, err := service.Project().ConvertFeedback(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.FeedbackConvertRes{TaskId: taskId}, nil
}

func (c *projectController) DismissFeedback(ctx context.Context, req *api.FeedbackDismissReq) (res *api.FeedbackDismissRes, err error) {
	err = service.Project().DismissFeedback(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.FeedbackDismissRes{}, nil
}

func (c *projectController) CreateTopic(ctx context.Context, req *api.TopicCreateReq) (res *api.TopicCreateRes, err error) {
	id, err := service.Project().CreateTopic(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicCreateRes{Id: id}, nil
}

func (c *projectController) ListTopics(ctx context.Context, req *api.TopicListReq) (res *api.TopicListRes, err error) {
	return service.Project().ListTopics(ctx, req)
}

func (c *projectController) GetTopic(ctx context.Context, req *api.TopicDetailReq) (res *api.TopicDetailRes, err error) {
	return service.Project().GetTopic(ctx, req.ProjectId, req.Id)
}

func (c *projectController) UpsertTopicPhases(ctx context.Context, req *api.TopicPhaseUpsertReq) (res *api.TopicPhaseUpsertRes, err error) {
	err = service.Project().UpsertTopicPhases(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseUpsertRes{}, nil
}

func (c *projectController) AppendTopicPhase(ctx context.Context, req *api.TopicPhaseAddReq) (res *api.TopicPhaseAddRes, err error) {
	id, err := service.Project().AppendTopicPhase(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseAddRes{Id: id}, nil
}

func (c *projectController) UpdateTopicPhase(ctx context.Context, req *api.TopicPhaseUpdateReq) (res *api.TopicPhaseUpdateRes, err error) {
	err = service.Project().UpdateTopicPhase(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseUpdateRes{}, nil
}

func (c *projectController) DeleteTopicPhase(ctx context.Context, req *api.TopicPhaseDeleteReq) (res *api.TopicPhaseDeleteRes, err error) {
	err = service.Project().DeleteTopicPhase(ctx, req.ProjectId, req.TopicId, req.PhaseId)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseDeleteRes{}, nil
}

func (c *projectController) ToggleTopicPhase(ctx context.Context, req *api.TopicPhaseToggleReq) (res *api.TopicPhaseToggleRes, err error) {
	err = service.Project().ToggleTopicPhase(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseToggleRes{}, nil
}

func (c *projectController) ConvertTopicPhase(ctx context.Context, req *api.TopicPhaseConvertReq) (res *api.TopicPhaseConvertRes, err error) {
	taskId, err := service.Project().ConvertTopicPhase(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicPhaseConvertRes{TaskId: taskId}, nil
}

func (c *projectController) LogTopic(ctx context.Context, req *api.TopicLogReq) (res *api.TopicLogRes, err error) {
	err = service.Project().LogTopic(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicLogRes{}, nil
}

func (c *projectController) FinishTopic(ctx context.Context, req *api.TopicFinishReq) (res *api.TopicFinishRes, err error) {
	err = service.Project().FinishTopic(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TopicFinishRes{}, nil
}

func (c *projectController) UpsertQa(ctx context.Context, req *api.QaUpsertReq) (res *api.QaUpsertRes, err error) {
	id, updated, err := service.Project().UpsertQa(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.QaUpsertRes{Id: id, Updated: updated}, nil
}

func (c *projectController) ListQas(ctx context.Context, req *api.QaListReq) (res *api.QaListRes, err error) {
	return service.Project().ListQas(ctx, req)
}

func (c *projectController) HitQa(ctx context.Context, req *api.QaHitReq) (res *api.QaHitRes, err error) {
	err = service.Project().HitQa(ctx, req.ProjectId, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.QaHitRes{}, nil
}

func (c *projectController) ArchiveQa(ctx context.Context, req *api.QaArchiveReq) (res *api.QaArchiveRes, err error) {
	err = service.Project().ArchiveQa(ctx, req.ProjectId, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.QaArchiveRes{}, nil
}

func (c *projectController) BlockTask(ctx context.Context, req *api.TaskBlockReq) (res *api.TaskBlockRes, err error) {
	err = service.Project().BlockTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskBlockRes{}, nil
}

func (c *projectController) UnblockTask(ctx context.Context, req *api.TaskUnblockReq) (res *api.TaskUnblockRes, err error) {
	err = service.Project().UnblockTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.TaskUnblockRes{}, nil
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

// ==================== 需求 ====================

func (c *projectController) CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (res *api.RequirementCreateRes, err error) {
	id, err := service.Project().CreateRequirement(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.RequirementCreateRes{Id: id}, nil
}

func (c *projectController) UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (res *api.RequirementUpdateRes, err error) {
	err = service.Project().UpdateRequirement(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.RequirementUpdateRes{}, nil
}

func (c *projectController) DeleteRequirement(ctx context.Context, req *api.RequirementDeleteReq) (res *api.RequirementDeleteRes, err error) {
	err = service.Project().DeleteRequirement(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.RequirementDeleteRes{}, nil
}

func (c *projectController) RequirementDetail(ctx context.Context, req *api.RequirementDetailReq) (res *api.RequirementDetailRes, err error) {
	return service.Project().GetRequirement(ctx, req.Id)
}

func (c *projectController) RequirementList(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error) {
	return service.Project().ListRequirements(ctx, req)
}

// ==================== 里程碑 ====================

func (c *projectController) CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (res *api.MilestoneCreateRes, err error) {
	id, err := service.Project().CreateMilestone(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.MilestoneCreateRes{Id: id}, nil
}

func (c *projectController) MilestoneList(ctx context.Context, req *api.MilestoneListReq) (res *api.MilestoneListRes, err error) {
	return service.Project().ListMilestones(ctx, req.ProjectId)
}

func (c *projectController) UpdateMilestone(ctx context.Context, req *api.MilestoneUpdateReq) (res *api.MilestoneUpdateRes, err error) {
	err = service.Project().UpdateMilestone(ctx, req)
	res = new(api.MilestoneUpdateRes)
	return
}

func (c *projectController) DeleteMilestone(ctx context.Context, req *api.MilestoneDeleteReq) (res *api.MilestoneDeleteRes, err error) {
	err = service.Project().DeleteMilestone(ctx, req.Id)
	res = new(api.MilestoneDeleteRes)
	return
}

func (c *projectController) UpdateComment(ctx context.Context, req *api.CommentUpdateReq) (res *api.CommentUpdateRes, err error) {
	err = service.Project().UpdateComment(ctx, req)
	res = new(api.CommentUpdateRes)
	return
}

func (c *projectController) DeleteComment(ctx context.Context, req *api.CommentDeleteReq) (res *api.CommentDeleteRes, err error) {
	err = service.Project().DeleteComment(ctx, req.Id)
	res = new(api.CommentDeleteRes)
	return
}
