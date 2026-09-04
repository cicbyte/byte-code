package project

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 任务 CRUD 与列表

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
		return 0, liberr.WrapDb(ctx, err, "创建任务失败")
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
	// 指针语义：nil=不更新，非nil=更新（含零值/空串）
	if req.Title != nil {
		data["title"] = *req.Title
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.Type != nil {
		data["type"] = *req.Type
	}
	if req.Status != nil {
		data["status"] = *req.Status
		switch *req.Status {
		case consts.TaskStatusDone, consts.TaskStatusClosed:
			// 仅首次进入完成态时记录；已完成的更新不重写
			if cur, _ := g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Fields("completed_at").Value(); cur == nil || cur.String() == "" {
				data["completed_at"] = time.Now().Format("2006-01-02 15:04:05")
			}
		case consts.TaskStatusOpen, consts.TaskStatusInProgress, consts.TaskStatusReview:
			// 从完成态切回非完成态时才清空（重开）；首次设定非完成态不误清
			if cur, _ := g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Fields("completed_at").Value(); cur != nil && cur.String() != "" {
				data["completed_at"] = ""
			}
		}
	}
	if req.Priority != nil {
		data["priority"] = *req.Priority
	}
	if req.AssigneeId != nil {
		data["assignee_id"] = *req.AssigneeId
	}
	if req.SprintId != nil {
		data["sprint_id"] = *req.SprintId
	}
	if req.ParentTaskId != nil {
		data["parent_task_id"] = *req.ParentTaskId
	}
	if req.SortOrder != nil {
		data["sort_order"] = *req.SortOrder
	}
	if len(data) == 0 {
		return nil
	}
	data["updated_at"] = time.Now().Format("2006-01-02 15:04:05")

	_, err = g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新任务失败")
	}
	return nil
}

func (s *sProject) DeleteTask(ctx context.Context, id int) (err error) {
	// 事务内清理任务的评论、AI 日志、标签与附件关联，最后删任务本身
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Delete("comments", "task_id", id); err != nil {
			return err
		}
		if _, err := tx.Delete("ai_execution_logs", "task_id", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM entity_tags WHERE entity_type = 'task' AND entity_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM attachments WHERE entity_type = 'task' AND entity_id = ?", id); err != nil {
			return err
		}
		_, err := tx.Delete("tasks", "id", id)
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除任务失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询任务失败")
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
		kw := "%" + escape.Like(req.Keyword) + "%"
		countM = countM.Where("(t.title LIKE ? ESCAPE '\\' OR t.description LIKE ? ESCAPE '\\')", kw, kw)
	}

	total, err := countM.Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务数量失败")
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
		kw := "%" + escape.Like(req.Keyword) + "%"
		m = m.Where("(t.title LIKE ? ESCAPE '\\' OR t.description LIKE ? ESCAPE '\\')", kw, kw)
	}

	var list []api.TaskItem
	err = m.Page(req.Page, req.Size).Order("t.sort_order ASC, t.id DESC").Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务列表失败")
	}
	res.List = list
	return res, nil
}
