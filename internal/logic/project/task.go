package project

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
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

	// 截止日期校验前置：非法格式在插入前报错，避免"报错但任务已落库，
	// 重试产生重复任务"（normalizeDueDate 是纯函数，无 IO 代价）
	due := ""
	if req.DueDate != "" {
		if due, err = normalizeDueDate(req.DueDate); err != nil {
			return 0, err
		}
	}
	// Sprint 必须属于当前项目：跨项目 sprint_id 会污染他项目燃尽统计
	if req.SprintId > 0 {
		if sp, _ := g.DB().Model("sprints").Ctx(ctx).Where("id", req.SprintId).Value("project_id"); sp == nil || sp.Int() != req.ProjectId {
			return 0, fmt.Errorf("Sprint 不属于当前项目")
		}
	}
	insert := g.Map{
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
	}
	if due != "" {
		insert["due_date"] = due
	}
	// 步骤清单：空串按空清单落库；非 JSON 数组拒绝（防脏数据进列）
	if req.Checklist != "" && req.Checklist != "[]" {
		var probe []interface{}
		if err := json.Unmarshal([]byte(req.Checklist), &probe); err != nil {
			return 0, fmt.Errorf("checklist 必须是 JSON 数组")
		}
		insert["checklist"] = req.Checklist
	}
	result, err := g.DB().Model("tasks").Ctx(ctx).Insert(insert)
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建任务失败")
	}
	lastId, _ := result.LastInsertId()
	taskId := int(lastId)

	// 未指派的 open 任务广播给项目内已准入的外部 agents（事件驱动认领，
	// 替代盲轮询任务池）；已指派任务的触达走既有指派通知
	if req.AssigneeId == 0 {
		agentRows, aerr := g.DB().Model("agent_project_bindings b").Ctx(ctx).
			Fields("b.agent_id, u.username").
			LeftJoin("sys_users u", "u.id = b.agent_id").
			Where("b.project_id", req.ProjectId).
			Where("u.status", 1).
			All()
		if aerr == nil {
			for _, ar := range agentRows {
				notify.Send(ctx, ar["agent_id"].Int(), "新任务可认领",
					fmt.Sprintf("项目新增未指派任务「%s」，可凭 bc key 认领（POST /v1/tasks/%d/claim）", req.Title, taskId),
					"info", "task", taskId)
			}
		}
	}

	// 记录活动
	s.recordActivity(ctx, uid, "task.created", "task", taskId, req.Title, req.ProjectId, "")

	// 如果有指派人，发送通知
	if req.AssigneeId > 0 {
		notify.Send(ctx, req.AssigneeId, "新任务分配", fmt.Sprintf("您有一个新任务: %s", req.Title), "info", "task", taskId)
	}

	return taskId, nil
}

func (s *sProject) UpdateTask(ctx context.Context, req *api.TaskUpdateReq) (err error) {
	// 门禁：agent 禁改 status/assignee——任务状态流转必须走 claim/complete
	// 专用端点（条件更新防并发、完成限 assignee/owner、人审拒 ai），
	// PUT 直改 done 可整体绕过这三道防线（CanAccessProject 对绑定 agent 放行）。
	// human 不受限：任意成员本就有 ReviewTask 权，直接改 done 不构成越权
	isAgent, aerr := actorIsAgent(ctx)
	if aerr != nil {
		return aerr
	}
	if isAgent {
		if req.Status != nil {
			return fmt.Errorf("Agent 不能直接修改任务状态，请通过认领/完成接口流转")
		}
		if req.AssigneeId != nil {
			return fmt.Errorf("Agent 不能改派任务")
		}
	}
	notifyAssigneeChange := 0
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
		case consts.TaskStatusOpen, consts.TaskStatusInProgress, consts.TaskStatusBlocked, consts.TaskStatusReview:
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
		// 指派变更通知新负责人（改派不通知等于改派无效传播）；
		// 需旧值对比：与自己相同/清空（0）不发
		notifyAssigneeChange = *req.AssigneeId
	}
	if req.SprintId != nil {
		if *req.SprintId > 0 {
			taskProject := perm.EntityProjectId(ctx, "tasks", req.Id)
			if sp, _ := g.DB().Model("sprints").Ctx(ctx).Where("id", *req.SprintId).Value("project_id"); sp == nil || sp.Int() != taskProject {
				return fmt.Errorf("Sprint 与任务不属于同一项目")
			}
		}
		data["sprint_id"] = *req.SprintId
	}
	if req.ParentTaskId != nil {
		data["parent_task_id"] = *req.ParentTaskId
	}
	if req.SortOrder != nil {
		data["sort_order"] = *req.SortOrder
	}
	if req.Checklist != nil {
		var probe []interface{}
		if e := json.Unmarshal([]byte(*req.Checklist), &probe); e != nil {
			return fmt.Errorf("checklist 必须是 JSON 数组")
		}
		data["checklist"] = *req.Checklist
	}
	if req.DueDate != nil {
		due, derr := normalizeDueDate(*req.DueDate)
		if derr != nil {
			return derr
		}
		data["due_date"] = due
	}
	if len(data) == 0 {
		return nil
	}
	data["updated_at"] = time.Now().Format("2006-01-02 15:04:05")

	// 旧指派用于对比（改派才通知，原值相同/清空不发）
	oldAssignee := 0
	if v, _ := g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Fields("assignee_id").Value(); v != nil {
		oldAssignee = v.Int()
	}
	_, err = g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新任务失败")
	}
	if notifyAssigneeChange > 0 && notifyAssigneeChange != oldAssignee {
		title, _ := g.DB().Model("tasks").Ctx(ctx).Where("id", req.Id).Fields("title").Value()
		notify.Send(ctx, notifyAssigneeChange, "任务转派给你",
			fmt.Sprintf("任务「%s」已指派给你", title.String()), "info", "task", req.Id)
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
	// 标签回填（与列表口径一致）
	item.Tags = taskTagNames(ctx, id)
	return &api.TaskDetailRes{TaskItem: item}, nil
}

// taskTagNames 查单个任务的标签名列表
func taskTagNames(ctx context.Context, taskId int) []string {
	rows, err := g.DB().Model("entity_tags et").Ctx(ctx).
		LeftJoin("tags g", "g.id = et.tag_id").
		Fields("g.name").
		Where("et.entity_type", "task").
		Where("et.entity_id", taskId).
		All()
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r["name"].String())
	}
	return out
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
	if req.TagId > 0 {
		tagCond := "EXISTS (SELECT 1 FROM entity_tags et WHERE et.entity_type = 'task' AND et.entity_id = t.id AND et.tag_id = ?)"
		countM = countM.Where(tagCond, req.TagId)
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
	if req.TagId > 0 {
		tagCond := "EXISTS (SELECT 1 FROM entity_tags et WHERE et.entity_type = 'task' AND et.entity_id = t.id AND et.tag_id = ?)"
		m = m.Where(tagCond, req.TagId)
	}

	var list []api.TaskItem
	err = m.Page(req.Page, req.Size).Order("t.sort_order ASC, t.id DESC").Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务列表失败")
	}
	// 批量回填任务标签（一次 IN 查询，避免逐任务 N+1）
	if len(list) > 0 {
		ids := make([]int, 0, len(list))
		for _, it := range list {
			ids = append(ids, it.Id)
		}
		tagRows, terr := g.DB().Model("entity_tags et").Ctx(ctx).
			LeftJoin("tags g", "g.id = et.tag_id").
			Fields("et.entity_id, g.name").
			Where("et.entity_type", "task").
			WhereIn("et.entity_id", ids).
			All()
		if terr == nil {
			tagsByTask := make(map[int][]string)
			for _, tr := range tagRows {
				tagsByTask[tr["entity_id"].Int()] = append(tagsByTask[tr["entity_id"].Int()], tr["name"].String())
			}
			for i := range list {
				list[i].Tags = tagsByTask[list[i].Id]
			}
		}
	}
	res.List = list
	return res, nil
}

// ==================== 我的任务（跨项目聚合） ====================

// myTaskStatuses 解析状态过滤口径：缺省=活跃四态（含 blocked）；all=不过滤；其余按逗号列表
func myTaskStatuses(s string) []string {
	if s == "" {
		return consts.TaskActiveStatuses
	}
	if s == "all" {
		return nil
	}
	return strings.Split(s, ",")
}

// MyTaskList 当前用户被指派的任务跨项目聚合：管理员看全部项目，
// 普通用户仅所在项目（与 CanAccessProject 同口径）
func (s *sProject) MyTaskList(ctx context.Context, req *api.MyTaskListReq) (res *api.MyTaskListRes, err error) {
	res = &api.MyTaskListRes{}
	uid := perm.UserId(ctx)

	base := func() *gdb.Model {
		m := g.DB().Model("tasks t").Ctx(ctx).
			LeftJoin("projects p", "p.id = t.project_id").
			LeftJoin("sys_users au", "t.assignee_id = au.id").
			LeftJoin("sys_users cu", "t.creator_id = cu.id").
			Where("t.assignee_id", uid)
		if !perm.IsAdmin(ctx, uid) {
			m = m.Where("EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = t.project_id AND pm.user_id = ?)", uid)
		}
		if statuses := myTaskStatuses(req.Status); len(statuses) > 0 {
			m = m.Where("t.status IN (?)", statuses)
		}
		if req.ProjectId > 0 {
			m = m.Where("t.project_id", req.ProjectId)
		}
		if req.Keyword != "" {
			kw := "%" + escape.Like(req.Keyword) + "%"
			m = m.Where("(t.title LIKE ? ESCAPE '\\' OR t.description LIKE ? ESCAPE '\\')", kw, kw)
		}
		return m
	}

	if res.Total, err = base().Count(); err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务数量失败")
	}
	var list []api.MyTaskItem
	err = base().
		Fields("t.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name, COALESCE(p.name, '') as project_name").
		// 状态推进顺序排前，同态按优先级与更新时间倒序
		Order(fmt.Sprintf("CASE t.status WHEN '%s' THEN 0 WHEN '%s' THEN 1 WHEN '%s' THEN 2 ELSE 3 END, t.priority DESC, t.updated_at DESC", consts.TaskStatusOpen, consts.TaskStatusInProgress, consts.TaskStatusReview)).
		Page(req.Page, req.Size).
		Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务失败")
	}
	res.List = list
	return res, nil
}
