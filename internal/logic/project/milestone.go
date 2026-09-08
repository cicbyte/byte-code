package project

import (
	"context"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/database/gdb"
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
	// 进度聚合：单查询 GROUP BY 回填（此前里程碑纯 CRUD 孤岛，需求关联字段
	// 早已存在但零呈现——审计挂账项）
	type aggRow struct {
		MilestoneId int
		Total       int
		Done        int
	}
	var aggs []aggRow
	if aerr := g.DB().Model("requirements r").Ctx(ctx).
		Where("r.project_id", projectId).
		Where("r.milestone_id > 0").
		Fields("r.milestone_id, COUNT(*) AS total, SUM(CASE WHEN r.status IN ('implemented','confirmed') THEN 1 ELSE 0 END) AS done").
		Group("r.milestone_id").
		Scan(&aggs); aerr == nil {
		aggMap := make(map[int]*aggRow, len(aggs))
		for i := range aggs {
			aggMap[aggs[i].MilestoneId] = &aggs[i]
		}
		for i := range list {
			if a := aggMap[list[i].Id]; a != nil {
				list[i].ReqTotal = a.Total
				list[i].ReqDone = a.Done
			}
		}
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
	// 解除关联的需求（milestone_id 置 0，需求保留）与删除同事务——
	// 分开执行时第二步失败会让关联信息已丢而里程碑还在
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Model("requirements").
			Where("milestone_id", id).Data("milestone_id", 0).Update(); err != nil {
			return err
		}
		_, err := tx.Ctx(ctx).Model("milestones").Where("id", id).Delete()
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除里程碑失败")
	}
	return nil
}
