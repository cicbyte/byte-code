package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
)

func init() {
	service.RegisterProject(New())
}

func New() *sProject {
	return &sProject{}
}

type sProject struct{}

// ==================== 项目 CRUD ====================

func (s *sProject) CreateProject(ctx context.Context, req *api.ProjectCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	// 开启事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 创建项目
		result, err := tx.Insert("projects", g.Map{
			"name":        req.Name,
			"description": req.Description,
			"created_by":  uid,
			"status":      1,
		})
		if err != nil {
			return fmt.Errorf("创建项目失败: %v", err)
		}
		lastId, _ := result.LastInsertId()
		id = int(lastId)

		// 自动将创建者添加为 owner
		_, err = tx.Insert("project_members", g.Map{
			"project_id": id,
			"user_id":    uid,
			"role":       "owner",
		})
		if err != nil {
			return fmt.Errorf("添加项目成员失败: %v", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	// 记录活动
	s.recordActivity(ctx, uid, "project.created", "project", id, req.Name, id, "")
	return id, nil
}

func (s *sProject) UpdateProject(ctx context.Context, req *api.ProjectUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Status > 0 {
		data["status"] = req.Status
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("projects").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新项目失败: %v", err)
	}
	return nil
}

func (s *sProject) DeleteProject(ctx context.Context, id int) (err error) {
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 删除项目成员
		_, err := tx.Delete("project_members", "project_id", id)
		if err != nil {
			return err
		}
		// 删除任务
		_, err = tx.Delete("tasks", "project_id", id)
		if err != nil {
			return err
		}
		// 删除 Sprint
		_, err = tx.Delete("sprints", "project_id", id)
		if err != nil {
			return err
		}
		// 删除项目
		_, err = tx.Delete("projects", "id", id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("删除项目失败: %v", err)
	}
	return nil
}

func (s *sProject) GetProject(ctx context.Context, id int) (res *api.ProjectDetailRes, err error) {
	var item api.ProjectItem
	err = g.DB().Model("projects p").Ctx(ctx).
		LeftJoin("sys_users u", "p.created_by = u.id").
		Fields("p.id, p.name, p.description, p.created_by, COALESCE(u.real_name, u.username) as creator_name, p.status, p.created_at, p.updated_at").
		Where("p.id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询项目失败: %v", err)
	}
	if item.Id == 0 {
		return nil, fmt.Errorf("项目不存在")
	}
	return &api.ProjectDetailRes{ProjectItem: item}, nil
}

func (s *sProject) ListProjects(ctx context.Context, req *api.ProjectListReq) (res *api.ProjectListRes, err error) {
	res = &api.ProjectListRes{
		Page: req.Page,
		Size: req.Size,
	}

	// 非管理员只能看到自己所在的项目
	uid := perm.UserId(ctx)
	memberOnly := uid > 0 && !perm.IsAdmin(ctx, uid)

	// Count 查询（不带 Fields，兼容 SQLite）
	countM := g.DB().Model("projects p").Ctx(ctx)
	if memberOnly {
		countM = countM.Where(
			"EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = ?)", uid)
	}
	if req.Status > 0 {
		countM = countM.Where("p.status", req.Status)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		countM = countM.Where("(p.name LIKE ? OR p.description LIKE ?)", kw, kw)
	}

	total, err := countM.Count()
	if err != nil {
		return nil, fmt.Errorf("查询项目数量失败: %v", err)
	}
	res.Total = total

	// 数据查询
	m := g.DB().Model("projects p").Ctx(ctx).
		LeftJoin("sys_users u", "p.created_by = u.id").
		Fields("p.id, p.name, p.description, p.created_by, COALESCE(u.real_name, u.username) as creator_name, p.status, p.created_at, p.updated_at")
	if memberOnly {
		m = m.Where(
			"EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = ?)", uid)
	}
	if req.Status > 0 {
		m = m.Where("p.status", req.Status)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		m = m.Where("(p.name LIKE ? OR p.description LIKE ?)", kw, kw)
	}

	var list []api.ProjectItem
	err = m.Page(req.Page, req.Size).Order("p.id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询项目列表失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== 项目成员 ====================

func (s *sProject) AddMember(ctx context.Context, req *api.MemberAddReq) (err error) {
	_, err = g.DB().Model("project_members").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"user_id":    req.UserId,
		"role":       req.Role,
	})
	if err != nil {
		return fmt.Errorf("添加成员失败: %v", err)
	}

	// 发送通知
	notify.Send(ctx, req.UserId, "加入项目", fmt.Sprintf("您已被添加到项目中"), "info", "project", req.ProjectId)
	return nil
}

func (s *sProject) RemoveMember(ctx context.Context, projectId, userId int) (err error) {
	_, err = g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).
		Where("user_id", userId).
		Delete()
	if err != nil {
		return fmt.Errorf("移除成员失败: %v", err)
	}
	return nil
}

func (s *sProject) ListMembers(ctx context.Context, projectId int) (res *api.MemberListRes, err error) {
	res = &api.MemberListRes{}
	var list []api.MemberItem
	err = g.DB().Model("project_members pm").Ctx(ctx).
		LeftJoin("sys_users u", "pm.user_id = u.id").
		Fields("pm.id, pm.user_id, u.username, COALESCE(u.real_name, '') as real_name, pm.role, pm.created_at as joined_at").
		Where("pm.project_id", projectId).
		Order("pm.id ASC").
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询成员列表失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== 任务 CRUD ====================

func (s *sProject) CreateTask(ctx context.Context, req *api.TaskCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	result, err := g.DB().Model("tasks").Ctx(ctx).Insert(g.Map{
		"project_id":     req.ProjectId,
		"requirement_id": req.RequirementId,
		"sprint_id":      req.SprintId,
		"title":          req.Title,
		"description":    req.Description,
		"type":           req.Type,
		"status":         "open",
		"priority":       req.Priority,
		"assignee_id":    req.AssigneeId,
		"creator_id":     uid,
		"parent_task_id": req.ParentTaskId,
		"source":         "human",
	})
	if err != nil {
		return 0, fmt.Errorf("创建任务失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	taskId := int(lastId)

	// 记录活动
	s.recordActivity(ctx, uid, "task.created", "task", taskId, req.Title, req.ProjectId, "")

	// 如果有指派人，发送通知
	if req.AssigneeId > 0 {
		notify.Send(ctx, req.AssigneeId, "新任务分配", fmt.Sprintf("您有一个新任务: %s", req.Title), "info", "task", taskId)
	}

	return taskId, nil
}

func (s *sProject) UpdateTask(ctx context.Context, req *api.TaskUpdateReq) (err error) {
	data := g.Map{}
	if req.Title != "" {
		data["title"] = req.Title
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Type != "" {
		data["type"] = req.Type
	}
	if req.Status != "" {
		data["status"] = req.Status
	}
	if req.Priority > 0 {
		data["priority"] = req.Priority
	}
	if req.AssigneeId > 0 {
		data["assignee_id"] = req.AssigneeId
	}
	if req.SprintId > 0 {
		data["sprint_id"] = req.SprintId
	}
	if req.ParentTaskId > 0 {
		data["parent_task_id"] = req.ParentTaskId
	}
	if req.SortOrder > 0 {
		data["sort_order"] = req.SortOrder
	}
	if len(data) == 0 {
		return nil
	}

	_, err = g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新任务失败: %v", err)
	}

	// 状态变更时记录活动
	if req.Status != "" {
		userId := 0
		if uid := ctx.Value("userId"); uid != nil {
			userId = uid.(int)
		}
		// 获取任务信息用于活动记录
		task, _ := s.GetTask(ctx, req.Id)
		taskTitle := ""
		projectId := 0
		if task != nil {
			taskTitle = task.Title
			projectId = task.ProjectId
		}
		s.recordActivity(ctx, userId, "task.status_changed", "task", req.Id, taskTitle, projectId, fmt.Sprintf("状态变更为: %s", req.Status))
	}

	return nil
}

func (s *sProject) DeleteTask(ctx context.Context, id int) (err error) {
	// 先删除关联的评论和日志
	_, _ = g.DB().Model("comments").Ctx(ctx).Where("task_id", id).Delete()
	_, _ = g.DB().Model("ai_execution_logs").Ctx(ctx).Where("task_id", id).Delete()
	_, err = g.DB().Model("tasks").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除任务失败: %v", err)
	}
	return nil
}

func (s *sProject) GetTask(ctx context.Context, id int) (res *api.TaskDetailRes, err error) {
	var item api.TaskItem
	err = g.DB().Model("tasks t").Ctx(ctx).
		LeftJoin("sys_users au", "t.assignee_id = au.id").
		LeftJoin("sys_users cu", "t.creator_id = cu.id").
		Fields("t.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		Where("t.id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	if item.Id == 0 {
		return nil, fmt.Errorf("任务不存在")
	}
	return &api.TaskDetailRes{TaskItem: item}, nil
}

func (s *sProject) ListTasks(ctx context.Context, req *api.TaskListReq) (res *api.TaskListRes, err error) {
	res = &api.TaskListRes{}

	// Count 查询（不带 Fields，兼容 SQLite）
	countM := g.DB().Model("tasks t").Ctx(ctx).
		Where("t.project_id", req.ProjectId)
	if req.Status != "" {
		countM = countM.Where("t.status", req.Status)
	}
	if req.Type != "" {
		countM = countM.Where("t.type", req.Type)
	}
	if req.SprintId > 0 {
		countM = countM.Where("t.sprint_id", req.SprintId)
	}
	if req.AssigneeId > 0 {
		countM = countM.Where("t.assignee_id", req.AssigneeId)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		countM = countM.Where("(t.title LIKE ? OR t.description LIKE ?)", kw, kw)
	}

	total, err := countM.Count()
	if err != nil {
		return nil, fmt.Errorf("查询任务数量失败: %v", err)
	}
	res.Total = total

	// 数据查询
	m := g.DB().Model("tasks t").Ctx(ctx).
		LeftJoin("sys_users au", "t.assignee_id = au.id").
		LeftJoin("sys_users cu", "t.creator_id = cu.id").
		Fields("t.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		Where("t.project_id", req.ProjectId)
	if req.Status != "" {
		m = m.Where("t.status", req.Status)
	}
	if req.Type != "" {
		m = m.Where("t.type", req.Type)
	}
	if req.SprintId > 0 {
		m = m.Where("t.sprint_id", req.SprintId)
	}
	if req.AssigneeId > 0 {
		m = m.Where("t.assignee_id", req.AssigneeId)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		m = m.Where("(t.title LIKE ? OR t.description LIKE ?)", kw, kw)
	}

	var list []api.TaskItem
	err = m.Page(req.Page, req.Size).Order("t.sort_order ASC, t.id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== 任务特殊操作 ====================

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
		Where("status", "open").
		Data(g.Map{
			"status":      "in_progress",
			"assignee_id": userId,
		}).Update()
	if err != nil {
		return fmt.Errorf("认领任务失败: %v", err)
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
		"status": "review",
	}
	if req.Artifacts != "" {
		data["artifacts"] = req.Artifacts
	}

	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", req.Id).
		Data(data).Update()
	if err != nil {
		return fmt.Errorf("完成任务失败: %v", err)
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
		return fmt.Errorf("审核任务失败: %v", err)
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
		return nil, fmt.Errorf("查询需求失败: %v", err)
	}
	if len(requirements) == 0 {
		return nil, fmt.Errorf("未找到需求或子需求")
	}

	taskIds = make([]int, 0)
	// 需求 ID 到任务 ID 的映射（用于处理父子关系）
	reqToTask := make(map[int]int)

	for _, r := range requirements {
		parentTaskId := 0
		if r.ParentId > 0 {
			if tid, ok := reqToTask[r.ParentId]; ok {
				parentTaskId = tid
			}
		}

		result, err := g.DB().Model("tasks").Ctx(ctx).Insert(g.Map{
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
			return taskIds, fmt.Errorf("导入任务失败: %v", err)
		}
		lastId, _ := result.LastInsertId()
		tid := int(lastId)
		taskIds = append(taskIds, tid)
		reqToTask[r.Id] = tid

		// 记录活动
		s.recordActivity(ctx, userId, "task.imported", "task", tid, r.Title, req.ProjectId, fmt.Sprintf("从需求 #%d 导入", r.Id))
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

// ==================== 评论 ====================

func (s *sProject) CreateComment(ctx context.Context, req *api.CommentCreateReq) (id int, err error) {
	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	result, err := g.DB().Model("comments").Ctx(ctx).Insert(g.Map{
		"task_id":   req.TaskId,
		"user_id":   userId,
		"content":   req.Content,
		"user_type": req.UserType,
	})
	if err != nil {
		return 0, fmt.Errorf("创建评论失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) ListComments(ctx context.Context, taskId int) (res *api.CommentListRes, err error) {
	res = &api.CommentListRes{}
	var list []api.CommentItem
	err = g.DB().Model("comments c").Ctx(ctx).
		LeftJoin("sys_users u", "c.user_id = u.id").
		Fields("c.id, c.task_id, c.user_id, u.username, COALESCE(u.real_name, '') as real_name, c.content, c.user_type, c.created_at").
		Where("c.task_id", taskId).
		Order("c.id ASC").
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询评论失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== AI 执行日志 ====================

func (s *sProject) CreateAiLog(ctx context.Context, req *api.AiLogCreateReq) (id int, err error) {
	result, err := g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"task_id":    req.TaskId,
		"ai_user_id": req.AiUserId,
		"action":     req.Action,
		"detail":     req.Detail,
		"status":     req.Status,
	})
	if err != nil {
		return 0, fmt.Errorf("创建AI执行日志失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) ListAiLogs(ctx context.Context, taskId int) (res *api.AiLogListRes, err error) {
	res = &api.AiLogListRes{}
	var list []api.AiLogItem
	err = g.DB().Model("ai_execution_logs").Ctx(ctx).
		Where("task_id", taskId).
		Order("id DESC").
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询AI执行日志失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== Sprint CRUD ====================

func (s *sProject) CreateSprint(ctx context.Context, req *api.SprintCreateReq) (id int, err error) {
	result, err := g.DB().Model("sprints").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"name":       req.Name,
		"goal":       req.Goal,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"status":     "planning",
	})
	if err != nil {
		return 0, fmt.Errorf("创建Sprint失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	sprintId := int(lastId)

	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}
	s.recordActivity(ctx, userId, "sprint.created", "sprint", sprintId, req.Name, req.ProjectId, "")
	return sprintId, nil
}

func (s *sProject) UpdateSprint(ctx context.Context, req *api.SprintUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Goal != "" {
		data["goal"] = req.Goal
	}
	if req.StartDate != "" {
		data["start_date"] = req.StartDate
	}
	if req.EndDate != "" {
		data["end_date"] = req.EndDate
	}
	if req.Status != "" {
		data["status"] = req.Status
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("sprints").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新Sprint失败: %v", err)
	}
	return nil
}

func (s *sProject) DeleteSprint(ctx context.Context, id int) (err error) {
	// 将 Sprint 下的任务的 sprint_id 置 0
	_, _ = g.DB().Model("tasks").Ctx(ctx).Where("sprint_id", id).Data(g.Map{"sprint_id": 0}).Update()
	_, err = g.DB().Model("sprints").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除Sprint失败: %v", err)
	}
	return nil
}

func (s *sProject) GetSprint(ctx context.Context, id int) (res *api.SprintDetailRes, err error) {
	var item api.SprintItem
	err = g.DB().Model("sprints").Ctx(ctx).
		Where("id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询Sprint失败: %v", err)
	}
	if item.Id == 0 {
		return nil, fmt.Errorf("Sprint不存在")
	}
	return &api.SprintDetailRes{SprintItem: item}, nil
}

func (s *sProject) ListSprints(ctx context.Context, req *api.SprintListReq) (res *api.SprintListRes, err error) {
	res = &api.SprintListRes{}
	m := g.DB().Model("sprints").Ctx(ctx).
		Where("project_id", req.ProjectId)

	if req.Status != "" {
		m = m.Where("status", req.Status)
	}

	var list []api.SprintItem
	err = m.Order("id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询Sprint列表失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== Sprint 任务管理 ====================

func (s *sProject) AddTaskToSprint(ctx context.Context, sprintId, taskId int) (err error) {
	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", taskId).
		Data(g.Map{"sprint_id": sprintId}).
		Update()
	if err != nil {
		return fmt.Errorf("添加任务到Sprint失败: %v", err)
	}
	return nil
}

func (s *sProject) RemoveTaskFromSprint(ctx context.Context, sprintId, taskId int) (err error) {
	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", taskId).
		Where("sprint_id", sprintId).
		Data(g.Map{"sprint_id": 0}).
		Update()
	if err != nil {
		return fmt.Errorf("从Sprint移除任务失败: %v", err)
	}
	return nil
}

// ==================== 燃尽图 ====================

func (s *sProject) GetBurndown(ctx context.Context, sprintId int) (res *api.BurndownRes, err error) {
	res = &api.BurndownRes{}

	// 获取 Sprint 信息
	var sprint api.SprintItem
	err = g.DB().Model("sprints").Ctx(ctx).Where("id", sprintId).Scan(&sprint)
	if err != nil || sprint.Id == 0 {
		return nil, fmt.Errorf("Sprint不存在")
	}

	// Sprint 下的总任务数
	totalCount, err := g.DB().Model("tasks").Ctx(ctx).
		Where("sprint_id", sprintId).
		Count()
	if err != nil {
		return nil, fmt.Errorf("查询任务数量失败: %v", err)
	}

	// 按天统计完成的任务数
	type dayCount struct {
		Date      string
		Completed int
	}
	var completedByDay []dayCount
	err = g.DB().Model("tasks").Ctx(ctx).
		Fields("DATE(updated_at) as date, COUNT(*) as completed").
		Where("sprint_id", sprintId).
		WhereIn("status", g.Slice{"done", "closed"}).
		Group("DATE(updated_at)").
		Order("date ASC").
		Scan(&completedByDay)
	if err != nil {
		return nil, fmt.Errorf("统计燃尽图数据失败: %v", err)
	}

	// 构建按天的数据
	completedMap := make(map[string]int)
	for _, dc := range completedByDay {
		completedMap[dc.Date] = dc.Completed
	}

	startDate, err := gtime.StrToTime(sprint.StartDate)
	if err != nil {
		startDate = gtime.Now()
	}
	endDate, err := gtime.StrToTime(sprint.EndDate)
	if err != nil {
		endDate = gtime.Now()
	}

	items := make([]api.BurndownItem, 0)
	cumulativeCompleted := 0
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("Y-m-d")
		if c, ok := completedMap[dateStr]; ok {
			cumulativeCompleted += c
		}
		items = append(items, api.BurndownItem{
			Date:      dateStr,
			Remaining: totalCount - cumulativeCompleted,
			Completed: cumulativeCompleted,
		})
	}
	res.Items = items
	return res, nil
}

// ==================== 辅助方法 ====================

func (s *sProject) recordActivity(ctx context.Context, actorId int, action, targetType string, targetId int, targetName string, projectId int, detail string) {
	actorName := ""
	if actorId > 0 {
		var user struct {
			Username string
			RealName string
		}
		err := g.DB().Model("sys_users").Ctx(ctx).Where("id", actorId).Scan(&user)
		if err == nil && user.RealName != "" {
			actorName = user.RealName
		} else if err == nil {
			actorName = user.Username
		}
	}

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    actorId,
		ActorType:  "human",
		ActorName:  actorName,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetId,
		TargetName: targetName,
		ProjectID:  projectId,
		Detail:     detail,
	})
}

// ==================== 需求 ====================

func (s *sProject) CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	result, err := g.DB().Model("requirements").Ctx(ctx).Insert(g.Map{
		"project_id":          req.ProjectId,
		"parent_id":           req.ParentId,
		"type":                req.Type,
		"title":               req.Title,
		"description":         req.Description,
		"status":              "draft",
		"priority":            req.Priority,
		"assignee_id":         req.AssigneeId,
		"creator_id":          uid,
		"milestone_id":        req.MilestoneId,
		"acceptance_criteria": req.AcceptanceCriteria,
		"source":              "human",
	})
	if err != nil {
		return 0, fmt.Errorf("创建需求失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (err error) {
	data := g.Map{}
	if req.Title != "" {
		data["title"] = req.Title
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Status != "" {
		data["status"] = req.Status
	}
	if req.Priority > 0 {
		data["priority"] = req.Priority
	}
	if req.AssigneeId > 0 {
		data["assignee_id"] = req.AssigneeId
	}
	if req.MilestoneId > 0 {
		data["milestone_id"] = req.MilestoneId
	}
	if req.AcceptanceCriteria != "" {
		data["acceptance_criteria"] = req.AcceptanceCriteria
	}
	if req.SortOrder > 0 {
		data["sort_order"] = req.SortOrder
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("requirements").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新需求失败: %v", err)
	}
	return nil
}

func (s *sProject) DeleteRequirement(ctx context.Context, id int) (err error) {
	_, err = g.DB().Model("requirements").Ctx(ctx).Where("parent_id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除子需求失败: %v", err)
	}
	_, err = g.DB().Model("requirements").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除需求失败: %v", err)
	}
	return nil
}

func (s *sProject) GetRequirement(ctx context.Context, id int) (res *api.RequirementDetailRes, err error) {
	var item api.RequirementItem
	err = g.DB().Model("requirements r").Ctx(ctx).
		LeftJoin("sys_users au", "r.assignee_id = au.id").
		LeftJoin("sys_users cu", "r.creator_id = cu.id").
		Fields("r.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		Where("r.id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询需求失败: %v", err)
	}
	if item.Id == 0 {
		return nil, fmt.Errorf("需求不存在")
	}
	return &api.RequirementDetailRes{RequirementItem: item}, nil
}

func (s *sProject) ListRequirements(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error) {
	res = &api.RequirementListRes{}

	countM := g.DB().Model("requirements r").Ctx(ctx).
		Where("r.project_id", req.ProjectId)
	if req.Type != "" {
		countM = countM.Where("r.type", req.Type)
	}
	if req.Status != "" {
		countM = countM.Where("r.status", req.Status)
	}
	if req.ParentId > 0 {
		countM = countM.Where("r.parent_id", req.ParentId)
	}

	total, err := countM.Count()
	if err != nil {
		return nil, fmt.Errorf("查询需求数量失败: %v", err)
	}
	res.Total = total

	m := g.DB().Model("requirements r").Ctx(ctx).
		LeftJoin("sys_users au", "r.assignee_id = au.id").
		LeftJoin("sys_users cu", "r.creator_id = cu.id").
		Fields("r.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		Where("r.project_id", req.ProjectId)
	if req.Type != "" {
		m = m.Where("r.type", req.Type)
	}
	if req.Status != "" {
		m = m.Where("r.status", req.Status)
	}
	if req.ParentId > 0 {
		m = m.Where("r.parent_id", req.ParentId)
	}

	var list []api.RequirementItem
	err = m.Page(req.Page, req.Size).Order("r.sort_order ASC, r.id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询需求列表失败: %v", err)
	}

	if req.ParentId == 0 {
		res.List = s.buildRequirementTree(ctx, list)
	} else {
		res.List = list
	}
	return res, nil
}

func (s *sProject) buildRequirementTree(ctx context.Context, list []api.RequirementItem) []api.RequirementItem {
	if len(list) == 0 {
		return list
	}
	parentIds := make([]int, 0, len(list))
	for _, item := range list {
		parentIds = append(parentIds, item.Id)
	}

	var children []api.RequirementItem
	err := g.DB().Model("requirements r").Ctx(ctx).
		LeftJoin("sys_users au", "r.assignee_id = au.id").
		LeftJoin("sys_users cu", "r.creator_id = cu.id").
		Fields("r.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		WhereIn("r.parent_id", parentIds).
		Order("r.sort_order ASC, r.id ASC").
		Scan(&children)
	if err != nil {
		return list
	}

	childrenMap := make(map[int][]api.RequirementItem)
	for _, child := range children {
		childrenMap[child.ParentId] = append(childrenMap[child.ParentId], child)
	}
	for i := range list {
		if childs, ok := childrenMap[list[i].Id]; ok {
			list[i].Children = childs
		}
	}
	return list
}

// ==================== 里程碑 ====================

func (s *sProject) CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (id int, err error) {
	result, err := g.DB().Model("milestones").Ctx(ctx).Insert(g.Map{
		"project_id":  req.ProjectId,
		"name":        req.Name,
		"description": req.Description,
		"target_date": req.TargetDate,
		"status":      "planning",
	})
	if err != nil {
		return 0, fmt.Errorf("创建里程碑失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) ListMilestones(ctx context.Context, projectId int) (res *api.MilestoneListRes, err error) {
	res = &api.MilestoneListRes{}
	var list []api.MilestoneItem
	err = g.DB().Model("milestones").Ctx(ctx).
		Where("project_id", projectId).
		Order("id DESC").
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询里程碑列表失败: %v", err)
	}
	res.List = list
	return res, nil
}
