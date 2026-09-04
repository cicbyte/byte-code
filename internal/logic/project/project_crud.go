package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 项目 CRUD 与列表

func (s *sProject) CreateProject(ctx context.Context, req *api.ProjectCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	// 同名项目查重（projects.name 不设 DB 唯一约束，由业务层保证）
	if cnt, _ := g.DB().Model("projects").Ctx(ctx).Where("name", req.Name).Count(); cnt > 0 {
		return 0, fmt.Errorf("项目名称已存在")
	}

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
			return liberr.WrapDb(ctx, err, "创建项目失败")
		}
		lastId, _ := result.LastInsertId()
		id = int(lastId)

		// 自动将创建者添加为 owner
		_, err = tx.Insert("project_members", g.Map{
			"project_id": id,
			"user_id":    uid,
			"role":       consts.MemberRoleOwner,
		})
		if err != nil {
			return liberr.WrapDb(ctx, err, "添加项目成员失败")
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
	if req.Name != nil {
		data["name"] = *req.Name
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.Status != nil {
		data["status"] = *req.Status
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("projects").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新项目失败")
	}
	return nil
}

func (s *sProject) DeleteProject(ctx context.Context, id int) (err error) {
	// 删除项目属管理级操作，仅 owner 或超管可执行
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, id) {
		return fmt.Errorf("仅项目管理员可删除项目")
	}
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先清理依赖项目内实体的关联数据，再删实体与项目本身；
		// foreign_keys 当前关闭，孤儿数据必须在这里显式清干净
		cascadeStmts := []struct {
			sql  string
			args []interface{}
		}{
			{`DELETE FROM comments WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)`, []interface{}{id}},
			{`DELETE FROM ai_execution_logs WHERE task_id IN (SELECT id FROM tasks WHERE project_id = ?)`, []interface{}{id}},
			{`DELETE FROM test_plan_cases WHERE test_plan_id IN (SELECT id FROM test_plans WHERE project_id = ?)`, []interface{}{id}},
			{`DELETE FROM db_columns WHERE table_id IN (SELECT id FROM db_tables WHERE project_id = ?)`, []interface{}{id}},
			{`DELETE FROM entity_tags WHERE
				(entity_type = 'task' AND entity_id IN (SELECT id FROM tasks WHERE project_id = ?)) OR
				(entity_type = 'requirement' AND entity_id IN (SELECT id FROM requirements WHERE project_id = ?)) OR
				(entity_type = 'test_case' AND entity_id IN (SELECT id FROM test_cases WHERE project_id = ?))`, []interface{}{id, id, id}},
			{`DELETE FROM attachments WHERE
				(entity_type = 'project' AND entity_id = ?) OR
				(entity_type = 'task' AND entity_id IN (SELECT id FROM tasks WHERE project_id = ?)) OR
				(entity_type = 'requirement' AND entity_id IN (SELECT id FROM requirements WHERE project_id = ?)) OR
				(entity_type = 'doc' AND entity_id IN (SELECT id FROM docs WHERE project_id = ?)) OR
				(entity_type = 'test_case' AND entity_id IN (SELECT id FROM test_cases WHERE project_id = ?))`, []interface{}{id, id, id, id, id}},
		}
		for _, stmt := range cascadeStmts {
			if _, err := tx.Exec(stmt.sql, stmt.args...); err != nil {
				return err
			}
		}

		// 项目级实体（activities 无外键但同属项目数据，一并清理避免孤儿）
		for _, table := range []string{
			"tasks", "sprints", "requirements", "milestones",
			"test_cases", "test_plans", "project_databases", "db_tables",
			"schema_versions", "project_members", "activities",
		} {
			if _, err := tx.Delete(table, "project_id", id); err != nil {
				return err
			}
		}
		// 项目本身
		if _, err := tx.Delete("projects", "id", id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除项目失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询项目失败")
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
		kw := "%" + escape.Like(req.Keyword) + "%"
		countM = countM.Where("(p.name LIKE ? ESCAPE '\\' OR p.description LIKE ? ESCAPE '\\')", kw, kw)
	}

	total, err := countM.Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询项目数量失败")
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
		kw := "%" + escape.Like(req.Keyword) + "%"
		m = m.Where("(p.name LIKE ? ESCAPE '\\' OR p.description LIKE ? ESCAPE '\\')", kw, kw)
	}

	var list []api.ProjectItem
	err = m.Page(req.Page, req.Size).Order("p.id DESC").Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询项目列表失败")
	}
	res.List = list
	return res, nil
}
