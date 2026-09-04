package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 需求 CRUD 与树

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
		return 0, liberr.WrapDb(ctx, err, "创建需求失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (err error) {
	data := g.Map{}
	if req.Title != nil {
		data["title"] = *req.Title
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.Status != nil {
		data["status"] = *req.Status
	}
	if req.Priority != nil {
		data["priority"] = *req.Priority
	}
	if req.AssigneeId != nil {
		data["assignee_id"] = *req.AssigneeId
	}
	if req.MilestoneId != nil {
		data["milestone_id"] = *req.MilestoneId
	}
	if req.AcceptanceCriteria != nil {
		data["acceptance_criteria"] = *req.AcceptanceCriteria
	}
	if req.SortOrder != nil {
		data["sort_order"] = *req.SortOrder
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("requirements").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新需求失败")
	}
	return nil
}

func (s *sProject) DeleteRequirement(ctx context.Context, id int) (err error) {
	// 递归 CTE 收集整棵子树（含自身），一并清理关联数据
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		subtree := `WITH RECURSIVE sub(id) AS (
			SELECT id FROM requirements WHERE id = ?
			UNION ALL
			SELECT r.id FROM requirements r JOIN sub ON r.parent_id = sub.id
		) SELECT id FROM sub`
		if _, err := tx.Exec("UPDATE tasks SET requirement_id = 0 WHERE requirement_id IN ("+subtree+")", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM entity_tags WHERE entity_type = 'requirement' AND entity_id IN ("+subtree+")", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM attachments WHERE entity_type = 'requirement' AND entity_id IN ("+subtree+")", id); err != nil {
			return err
		}
		_, err := tx.Exec("DELETE FROM requirements WHERE id IN ("+subtree+")", id)
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除需求失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询需求失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询需求数量失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询需求列表失败")
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
