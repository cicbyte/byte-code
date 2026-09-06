package project

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 任务操作：认领/完成/审核/需求导入

// actorIsAgent 判定当前操作者是否 agent 账号（sys_users.type='ai'）。
// 查询失败返回错误供调用方 fail-closed：人审等安全门禁在操作者身份
// 无法确认时必须拒绝而非放行（DB 抖动窗口可被 agent 趁机过审）
func actorIsAgent(ctx context.Context) (bool, error) {
	t, err := g.DB().Model("sys_users").Ctx(ctx).
		Where("id", perm.UserId(ctx)).Fields("type").Value()
	if err != nil {
		return false, liberr.WrapDb(ctx, err, "校验操作者身份失败")
	}
	return t != nil && t.String() == "ai", nil
}

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

	// 完成限任务的 assignee 本人或项目管理员：任意项目成员/绑定 agent
	// 可完成任意任务会破坏执行归属（claim 的意义）
	uid := perm.UserId(ctx)
	if task.AssigneeId != uid && !perm.IsProjectOwner(ctx, uid, task.ProjectId) {
		return fmt.Errorf("仅任务负责人或项目管理员可完成任务")
	}

	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	data := g.Map{
		"status": consts.TaskStatusReview,
		// 完成时间独立记录：燃尽图按此统计，任务后续编辑不会重写历史
		"completed_at": time.Now().Format("2006-01-02 15:04:05"),
		// 进 review 即待人审：此前无人设置 rh=1，审核按钮（前端要求
		// rh=1）与审核接口（后端同样校验）对完成态任务全部失效
		"requires_human_review": 1,
		"human_review_status":   "pending",
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

// BlockTask 上报阻塞：assignee 执行中遇到外部阻碍（等信息补充/环境问题/
// 依赖未就绪）主动举手。条件更新保证只有 in_progress 可进 blocked；
// 原因进执行日志留痕（任务详情的时间线上分人可见），并通知任务创建者。
// blocked 豁免租约回收（ReleaseStaleClaims 只扫 in_progress）——
// 阻塞是明确的"等人"，不是失联
func (s *sProject) BlockTask(ctx context.Context, req *api.TaskBlockReq) (err error) {
	task, err := s.GetTask(ctx, req.Id)
	if err != nil {
		return err
	}
	if task.AssigneeId != perm.UserId(ctx) && !perm.IsProjectOwner(ctx, perm.UserId(ctx), task.ProjectId) {
		return fmt.Errorf("仅任务负责人或项目管理员可上报阻塞")
	}
	result, err := g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Where("status", consts.TaskStatusInProgress).
		Data(g.Map{"status": consts.TaskStatusBlocked}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "上报阻塞失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("仅进行中的任务可上报阻塞")
	}
	if _, e := g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"task_id": req.Id, "ai_user_id": perm.UserId(ctx),
		"action": "blocked", "detail": req.Reason, "status": "failed",
	}); e != nil {
		g.Log().Warningf(ctx, "block log insert failed: %v", e)
	}
	notify.Send(ctx, task.CreatorId, "任务已阻塞",
		fmt.Sprintf("任务「%s」被阻塞：%s（请补充信息或处理环境问题）", task.Title, req.Reason),
		"warning", "task", req.Id)
	s.recordActivity(ctx, perm.UserId(ctx), "task.blocked", "task", req.Id, task.Title, task.ProjectId, req.Reason)
	return nil
}

// UnblockTask 解除阻塞：信息补齐后回进行中，继续由原 assignee 执行
func (s *sProject) UnblockTask(ctx context.Context, req *api.TaskUnblockReq) (err error) {
	task, err := s.GetTask(ctx, req.Id)
	if err != nil {
		return err
	}
	if task.AssigneeId != perm.UserId(ctx) && !perm.IsProjectOwner(ctx, perm.UserId(ctx), task.ProjectId) {
		return fmt.Errorf("仅任务负责人或项目管理员可解除阻塞")
	}
	result, err := g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Where("status", consts.TaskStatusBlocked).
		Data(g.Map{"status": consts.TaskStatusInProgress}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "解除阻塞失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("任务不在阻塞状态")
	}
	if _, e := g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"task_id": req.Id, "ai_user_id": perm.UserId(ctx),
		"action": "unblocked", "detail": "阻塞解除，恢复执行", "status": "success",
	}); e != nil {
		g.Log().Warningf(ctx, "unblock log insert failed: %v", e)
	}
	notify.Send(ctx, task.CreatorId, "阻塞已解除",
		fmt.Sprintf("任务「%s」的阻塞已解除，恢复进行中", task.Title),
		"success", "task", req.Id)
	s.recordActivity(ctx, perm.UserId(ctx), "task.unblocked", "task", req.Id, task.Title, task.ProjectId, "")
	return nil
}

func (s *sProject) ReviewTask(ctx context.Context, req *api.TaskReviewReq) (err error) {
	// 人审门禁：agent 不能自审通过自己提交的任务（CanAccessProject 对绑定
	// agent 放行，此处是审核语义的最后防线）；身份查询失败同样拒绝
	isAgent, aerr := actorIsAgent(ctx)
	if aerr != nil {
		return aerr
	}
	if isAgent {
		return fmt.Errorf("任务审核仅限人类用户")
	}
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
	if req.SprintId > 0 {
		if sp, _ := g.DB().Model("sprints").Ctx(ctx).Where("id", req.SprintId).Value("project_id"); sp == nil || sp.Int() != req.ProjectId {
			return nil, fmt.Errorf("Sprint 不属于当前项目")
		}
	}
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
