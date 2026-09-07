package project

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Sprint CRUD 与任务管理

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
		return 0, liberr.WrapDb(ctx, err, "创建Sprint失败")
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
	if req.Name != nil {
		data["name"] = *req.Name
	}
	if req.Goal != nil {
		data["goal"] = *req.Goal
	}
	if req.StartDate != nil {
		data["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		data["end_date"] = *req.EndDate
	}
	if req.Status != nil {
		// 与 sprints 表 CHECK(planning/active/completed) 对齐：非法值给业务错误而非裸 CHECK 报错
		switch *req.Status {
		case "planning", "active", "completed":
		default:
			return fmt.Errorf("状态必须是 planning/active/completed")
		}
		data["status"] = *req.Status
	}
	if req.StartDate != nil && *req.StartDate != "" {
		if _, e := time.Parse("2006-01-02", (*req.StartDate)[:min(10, len(*req.StartDate))]); e != nil {
			return fmt.Errorf("开始日期格式应为 Y-m-d")
		}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		if _, e := time.Parse("2006-01-02", (*req.EndDate)[:min(10, len(*req.EndDate))]); e != nil {
			return fmt.Errorf("结束日期格式应为 Y-m-d")
		}
	}
	if len(data) == 0 {
		return nil
	}
	res, err := g.DB().Model("sprints").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新Sprint失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("Sprint 不存在")
	}
	return nil
}

func (s *sProject) DeleteSprint(ctx context.Context, id int) (err error) {
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 将 Sprint 下的任务的 sprint_id 置 0（解除绑定，任务保留）
		if _, err := tx.Model("tasks").Where("sprint_id", id).Data(g.Map{"sprint_id": 0}).Update(); err != nil {
			return err
		}
		_, err := tx.Delete("sprints", "id", id)
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除Sprint失败")
	}
	return nil
}

func (s *sProject) GetSprint(ctx context.Context, id int) (res *api.SprintDetailRes, err error) {
	var item api.SprintItem
	err = g.DB().Model("sprints").Ctx(ctx).
		Where("id", id).
		Scan(&item)
	if err != nil && !isNoRows(err) {
		return nil, liberr.WrapDb(ctx, err, "查询Sprint失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询Sprint列表失败")
	}
	res.List = list
	return res, nil
}

// ==================== Sprint 任务管理 ====================

// assertSprintTaskSameProject 校验任务与 Sprint 属同一项目：跨项目绑定会把
// 别项目的任务拉进本冲刺（越权写 + 燃尽图按 sprint_id 全表计数被污染）。
// CreateTask/UpdateTask/Import 的 SprintId 写入同口径走此校验
func assertSprintTaskSameProject(ctx context.Context, sprintId, taskId int) error {
	if sprintId <= 0 || taskId <= 0 {
		return nil
	}
	taskProject := perm.EntityProjectId(ctx, "tasks", taskId)
	if taskProject == 0 {
		return fmt.Errorf("任务不存在")
	}
	sp, err := g.DB().Model("sprints").Ctx(ctx).Where("id", sprintId).Value("project_id")
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询Sprint失败")
	}
	if sp == nil || sp.Int() == 0 {
		return fmt.Errorf("Sprint 不存在")
	}
	if taskProject != sp.Int() {
		return fmt.Errorf("任务与 Sprint 不属于同一项目，禁止跨项目绑定")
	}
	return nil
}

func (s *sProject) AddTaskToSprint(ctx context.Context, sprintId, taskId int) (err error) {
	if err := assertSprintTaskSameProject(ctx, sprintId, taskId); err != nil {
		return err
	}
	result, err := g.DB().Model("tasks").Ctx(ctx).
		Where("id", taskId).
		Data(g.Map{"sprint_id": sprintId}).
		Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "添加任务到Sprint失败")
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return fmt.Errorf("任务不存在")
	}
	return nil
}

func (s *sProject) RemoveTaskFromSprint(ctx context.Context, sprintId, taskId int) (err error) {
	if err := assertSprintTaskSameProject(ctx, sprintId, taskId); err != nil {
		return err
	}
	_, err = g.DB().Model("tasks").Ctx(ctx).
		Where("id", taskId).
		Where("sprint_id", sprintId).
		Data(g.Map{"sprint_id": 0}).
		Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "从Sprint移除任务失败")
	}
	return nil
}
