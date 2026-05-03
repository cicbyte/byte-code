package product

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/product"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterProduct(New())
}

func New() *sProduct {
	return &sProduct{}
}

type sProduct struct{}

// ==================== 产品 ====================

func (s *sProduct) CreateProduct(ctx context.Context, req *api.ProductCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	result, err := g.DB().Model("products").Ctx(ctx).Insert(g.Map{
		"name":        req.Name,
		"description": req.Description,
		"owner_id":    uid,
		"status":      "active",
	})
	if err != nil {
		return 0, fmt.Errorf("创建产品失败: %v", err)
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProduct) UpdateProduct(ctx context.Context, req *api.ProductUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != "" {
		data["name"] = req.Name
	}
	if req.Description != "" {
		data["description"] = req.Description
	}
	if req.Status != "" {
		data["status"] = req.Status
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("products").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return fmt.Errorf("更新产品失败: %v", err)
	}
	return nil
}

func (s *sProduct) DeleteProduct(ctx context.Context, id int) (err error) {
	_, err = g.DB().Model("products").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return fmt.Errorf("删除产品失败: %v", err)
	}
	return nil
}

func (s *sProduct) GetProduct(ctx context.Context, id int) (res *api.ProductDetailRes, err error) {
	var item api.ProductItem
	err = g.DB().Model("products p").Ctx(ctx).
		LeftJoin("sys_users u", "p.owner_id = u.id").
		Fields("p.id, p.name, p.description, p.owner_id, p.status, p.created_at, p.updated_at, COALESCE(u.real_name, u.username) as owner_name").
		Where("p.id", id).
		Scan(&item)
	if err != nil {
		return nil, fmt.Errorf("查询产品失败: %v", err)
	}
	if item.Id == 0 {
		return nil, fmt.Errorf("产品不存在")
	}
	return &api.ProductDetailRes{ProductItem: item}, nil
}

func (s *sProduct) ListProducts(ctx context.Context, req *api.ProductListReq) (res *api.ProductListRes, err error) {
	res = &api.ProductListRes{
		Page: req.Page,
		Size: req.Size,
	}
	m := g.DB().Model("products p").Ctx(ctx).
		LeftJoin("sys_users u", "p.owner_id = u.id").
		Fields("p.id, p.name, p.description, p.owner_id, p.status, p.created_at, p.updated_at, COALESCE(u.real_name, u.username) as owner_name")

	if req.Status != "" {
		m = m.Where("p.status", req.Status)
	}

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("查询产品数量失败: %v", err)
	}
	res.Total = total

	var list []api.ProductItem
	err = m.Page(req.Page, req.Size).Order("p.id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询产品列表失败: %v", err)
	}
	res.List = list
	return res, nil
}

// ==================== 需求 ====================

func (s *sProduct) CreateRequirement(ctx context.Context, req *api.RequirementCreateReq) (id int, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未获取到用户信息")
	}
	uid := userId.(int)

	result, err := g.DB().Model("requirements").Ctx(ctx).Insert(g.Map{
		"product_id":          req.ProductId,
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

func (s *sProduct) UpdateRequirement(ctx context.Context, req *api.RequirementUpdateReq) (err error) {
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

func (s *sProduct) DeleteRequirement(ctx context.Context, id int) (err error) {
	// 先删除子需求
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

func (s *sProduct) GetRequirement(ctx context.Context, id int) (res *api.RequirementDetailRes, err error) {
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

func (s *sProduct) ListRequirements(ctx context.Context, req *api.RequirementListReq) (res *api.RequirementListRes, err error) {
	res = &api.RequirementListRes{}

	m := g.DB().Model("requirements r").Ctx(ctx).
		LeftJoin("sys_users au", "r.assignee_id = au.id").
		LeftJoin("sys_users cu", "r.creator_id = cu.id").
		Fields("r.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		Where("r.product_id", req.ProductId)

	if req.Type != "" {
		m = m.Where("r.type", req.Type)
	}
	if req.Status != "" {
		m = m.Where("r.status", req.Status)
	}
	if req.ParentId > 0 {
		m = m.Where("r.parent_id", req.ParentId)
	}

	total, err := m.Count()
	if err != nil {
		return nil, fmt.Errorf("查询需求数量失败: %v", err)
	}
	res.Total = total

	var list []api.RequirementItem
	err = m.Page(req.Page, req.Size).Order("r.sort_order ASC, r.id DESC").Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询需求列表失败: %v", err)
	}

	// 构建树形结构
	if req.ParentId == 0 {
		res.List = s.buildRequirementTree(list)
	} else {
		res.List = list
	}

	return res, nil
}

// buildRequirementTree 将扁平列表转换为树形结构
func (s *sProduct) buildRequirementTree(list []api.RequirementItem) []api.RequirementItem {
	if len(list) == 0 {
		return list
	}
	// 收集所有顶级需求的 ID
	parentIds := make([]int, 0)
	for _, item := range list {
		parentIds = append(parentIds, item.Id)
	}

	// 查询子需求
	var children []api.RequirementItem
	err := g.DB().Model("requirements r").
		LeftJoin("sys_users au", "r.assignee_id = au.id").
		LeftJoin("sys_users cu", "r.creator_id = cu.id").
		Fields("r.*, COALESCE(au.real_name, au.username) as assignee_name, COALESCE(cu.real_name, cu.username) as creator_name").
		WhereIn("r.parent_id", parentIds).
		Order("r.sort_order ASC, r.id ASC").
		Scan(&children)
	if err != nil {
		// 子查询失败时不影响主列表
		return list
	}

	// 建立子需求的映射
	childrenMap := make(map[int][]api.RequirementItem)
	for _, child := range children {
		childrenMap[child.ParentId] = append(childrenMap[child.ParentId], child)
	}

	// 组装树形结构
	for i := range list {
		if childs, ok := childrenMap[list[i].Id]; ok {
			list[i].Children = childs
		}
	}

	return list
}

// ==================== 里程碑 ====================

func (s *sProduct) CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (id int, err error) {
	result, err := g.DB().Model("milestones").Ctx(ctx).Insert(g.Map{
		"product_id":  req.ProductId,
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

func (s *sProduct) ListMilestones(ctx context.Context, productId int) (res *api.MilestoneListRes, err error) {
	res = &api.MilestoneListRes{}
	var list []api.MilestoneItem
	err = g.DB().Model("milestones").Ctx(ctx).
		Where("product_id", productId).
		Order("id DESC").
		Scan(&list)
	if err != nil {
		return nil, fmt.Errorf("查询里程碑列表失败: %v", err)
	}
	res.List = list
	return res, nil
}
