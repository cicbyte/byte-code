package project

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 任务操作：认领/完成/审核/需求导入

func (s *sProject) ClaimTask(ctx context.Context, req *api.TaskClaimReq) (err error) {
	// 预取任务信息用于存在性校验与活动记录
	task, err := s.GetTask(ctx, req.Id)
	if err != nil {
		return err
	}

	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	// 条件更新防并发认领：仅当任务仍为 open 时生效，以影响行数判断成败。
	// 先查后改在并发下会让两个认领者同时成功（后者覆盖前者的 assignee）
	result, err := g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Where("status", consts.TaskStatusOpen).
		Data(g.Map{
			"status":      consts.TaskStatusInProgress,
			"assignee_id": userId,
		}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "认领任务失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("任务已被认领或状态不可认领")
	}

	// 记录活动
	s.recordActivity(ctx, userId, "task.claimed", "task", req.Id, task.Title, task.ProjectId, "")
	return nil
}

func (s *sProject) CompleteTask(ctx context.Context, req *api.TaskCompleteReq) (err error) {
	task, err := s.GetTask(ctx, req.Id)
	if err != nil {
		return err
	}

	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	data := g.Map{
		"status": consts.TaskStatusReview,
		// 完成时间独立记录：燃尽图按此统计，任务后续编辑不会重写历史
		"completed_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	if req.Artifacts != "" {
		data["artifacts"] = req.Artifacts
	}

	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "完成任务失败")
	}

	// 记录活动
	s.recordActivity(ctx, userId, "task.completed", "task", req.Id, task.Title, task.ProjectId, "")

	// 通知创建者审核
	notify.Send(ctx, task.CreatorId, "任务待审核", fmt.Sprintf("任务 '%s' 已完成，等待审核", task.Title), "info", "task", req.Id)
	return nil
}

func (s *sProject) ReviewTask(ctx context.Context, req *api.TaskReviewReq) (err error) {
	task, err := s.GetTask(ctx, req.Id)
	if err != nil {
		return err
	}
	if task.RequiresHumanReview != 1 {
		return fmt.Errorf("该任务不需要人工审核")
	}

	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	newStatus := "done"
	if req.Status == "rejected" {
		newStatus = "in_progress"
	}

	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Data(g.Map{
			"status":              newStatus,
			"human_review_status": req.Status,
		}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "审核任务失败")
	}

	// 如果有评论，添加审核评论
	if req.Comment != "" {
		_, _ = g.DB().Model("comments").Ctx(ctx).Insert(g.Map{
			"task_id":   req.Id,
			"user_id":   userId,
			"content":   req.Comment,
			"user_type": "human",
		})
	}

	// 记录活动
	s.recordActivity(ctx, userId, "task.reviewed", "task", req.Id, task.Title, task.ProjectId, fmt.Sprintf("审核结果: %s", req.Status))

	// 通知任务执行者
	notify.Send(ctx, task.AssigneeId, "任务审核结果", fmt.Sprintf("任务 '%s' 审核结果: %s", task.Title, req.Status), "info", "task", req.Id)
	return nil
}

func (s *sProject) ImportTasks(ctx context.Context, req *api.TaskImportReq) (taskIds []int, err error) {
	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	// 查询需求及其子需求
	var requirements []struct {
		Id          int
		ParentId    int
		Title       string
		Description string
		Type        string
		Priority    int
	}
	err = g.DB().Model("requirements").Ctx(ctx).
		Where("project_id", req.ProjectId).
		Where("id = ? OR parent_id = ?", req.RequirementId, req.RequirementId).
		Order("parent_id ASC, sort_order ASC").
		Scan(&requirements)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询需求失败")
	}
	if len(requirements) == 0 {
		return nil, fmt.Errorf("未找到需求或子需求")
	}

	taskIds = make([]int, 0)
	// 需求 ID 到任务 ID 的映射（用于处理父子关系）
	reqToTask := make(map[int]int)

	// 整批导入包事务：中途失败整体回滚，不留半截导入（父子映射也会断）
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, r := range requirements {
			parentTaskId := 0
			if r.ParentId > 0 {
				if tid, ok := reqToTask[r.ParentId]; ok {
					parentTaskId = tid
				}
			}

			result, err := tx.Ctx(ctx).Model("tasks").Insert(g.Map{
				"project_id":     req.ProjectId,
				"requirement_id": r.Id,
				"sprint_id":      req.SprintId,
				"title":          r.Title,
				"description":    r.Description,
				"type":           s.mapReqTypeToTaskType(r.Type),
				"status":         "open",
				"priority":       r.Priority,
				"creator_id":     userId,
				"parent_task_id": parentTaskId,
				"source":         "pm_import",
			})
			if err != nil {
				return liberr.WrapDb(ctx, err, "导入任务失败")
			}
			lastId, _ := result.LastInsertId()
			tid := int(lastId)
			taskIds = append(taskIds, tid)
			reqToTask[r.Id] = tid

			// 记录活动（事务外补记也不影响一致性，这里顺路同事务写入）
			s.recordActivity(ctx, userId, "task.imported", "task", tid, r.Title, req.ProjectId, fmt.Sprintf("从需求 #%d 导入", r.Id))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return taskIds, nil
}

func (s *sProject) mapReqTypeToTaskType(reqType string) string {
	switch reqType {
	case "bug":
		return "bug"
	case "task":
		return "chore"
	default:
		return "feature"
	}
}
