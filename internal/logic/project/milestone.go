package project

import (
	"context"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/frame/g"
)

// 里程碑 CRUD

func (s *sProject) CreateMilestone(ctx context.Context, req *api.MilestoneCreateReq) (id int, err error) {
	result, err := g.DB().Model("milestones").Ctx(ctx).Insert(g.Map{
		"project_id":  req.ProjectId,
		"name":        req.Name,
		"description": req.Description,
		"target_date": req.TargetDate,
		"status":      "planning",
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建里程碑失败")
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
		return nil, liberr.WrapDb(ctx, err, "查询里程碑列表失败")
	}
	res.List = list
	return res, nil
}
func (s *sProject) UpdateMilestone(ctx context.Context, req *api.MilestoneUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != nil {
		data["name"] = *req.Name
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if req.TargetDate != nil {
		data["target_date"] = *req.TargetDate
	}
	if req.Status != nil {
		data["status"] = *req.Status
	}
	if len(data) == 0 {
		return nil
	}
	data["updated_at"] = time.Now().Format("2006-01-02 15:04:05")
	_, err = g.DB().Model("milestones").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新里程碑失败")
	}
	return nil
}

func (s *sProject) DeleteMilestone(ctx context.Context, id int) (err error) {
	// 解除关联的需求（milestone_id 置 0，需求保留）
	if _, err := g.DB().Model("requirements").Ctx(ctx).
		Where("milestone_id", id).Data("milestone_id", 0).Update(); err != nil {
		return liberr.WrapDb(ctx, err, "解除需求关联失败")
	}
	_, err = g.DB().Model("milestones").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除里程碑失败")
	}
	return nil
}
