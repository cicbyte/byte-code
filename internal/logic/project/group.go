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

// ==================== 项目分组（多对多）：同 group 自动互为关联 ====================

func (s *sProject) CreateGroup(ctx context.Context, req *api.GroupCreateReq) (id int, err error) {
	uid := perm.UserId(ctx)
	if cnt, _ := g.DB().Model("project_groups").Ctx(ctx).Where("name", req.Name).Count(); cnt > 0 {
		return 0, fmt.Errorf("分组名已存在")
	}
	result, err := g.DB().Model("project_groups").Ctx(ctx).Insert(g.Map{
		"name": req.Name, "description": req.Description,
		"created_by": uid, "created_at": time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建分组失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) ListGroups(ctx context.Context) (res *api.GroupListRes, err error) {
	res = &api.GroupListRes{List: []api.GroupListItem{}}
	var groups []api.GroupListItem
	err = g.DB().Model("project_groups g").Ctx(ctx).
		Order("g.id ASC").
		Scan(&groups)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询分组失败")
	}
	// 批量回填成员项目
	if len(groups) > 0 {
		var members []struct {
			GroupId int
			Id      int
			Code    string
			Name    string
		}
		g.DB().Model("project_group_members pgm").Ctx(ctx).
			InnerJoin("projects p", "p.id = pgm.project_id").
			Fields("pgm.group_id, p.id, p.code, p.name").
			Order("pgm.group_id ASC, p.id ASC").
			Scan(&members)
		memberMap := make(map[int][]api.GroupProjectBrief)
		for _, m := range members {
			memberMap[m.GroupId] = append(memberMap[m.GroupId], api.GroupProjectBrief{Id: m.Id, Code: m.Code, Name: m.Name})
		}
		for i := range groups {
			if projects, ok := memberMap[groups[i].Id]; ok {
				groups[i].Projects = projects
			} else {
				groups[i].Projects = []api.GroupProjectBrief{}
			}
		}
	}
	res.List = groups
	return res, nil
}

func (s *sProject) UpdateGroup(ctx context.Context, req *api.GroupUpdateReq) (err error) {
	data := g.Map{}
	if req.Name != nil {
		data["name"] = *req.Name
	}
	if req.Description != nil {
		data["description"] = *req.Description
	}
	if len(data) == 0 {
		return nil
	}
	_, err = g.DB().Model("project_groups").Ctx(ctx).Where("id", req.Id).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新分组失败")
	}
	return nil
}

func (s *sProject) DeleteGroup(ctx context.Context, id int) (err error) {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec("DELETE FROM project_group_members WHERE group_id = ?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM project_groups WHERE id = ?", id); err != nil {
			return err
		}
		return nil
	})
}

func (s *sProject) AddProjectToGroup(ctx context.Context, req *api.GroupMemberAddReq) (err error) {
	// 校验项目存在
	if cnt, _ := g.DB().Model("projects").Ctx(ctx).Where("id", req.ProjectId).Count(); cnt == 0 {
		return fmt.Errorf("项目不存在")
	}
	// 幂等：已存在则跳过
	if cnt, _ := g.DB().Model("project_group_members").Ctx(ctx).
		Where("group_id", req.Id).Where("project_id", req.ProjectId).Count(); cnt > 0 {
		return nil
	}
	_, err = g.DB().Model("project_group_members").Ctx(ctx).Insert(g.Map{
		"group_id": req.Id, "project_id": req.ProjectId,
		"added_by": perm.UserId(ctx), "created_at": time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "添加到分组失败")
	}
	return nil
}

func (s *sProject) RemoveProjectFromGroup(ctx context.Context, groupId, projectId int) (err error) {
	_, err = g.DB().Model("project_group_members").Ctx(ctx).
		Where("group_id", groupId).Where("project_id", projectId).Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "从分组移除失败")
	}
	return nil
}

// ShareGroup 判断两个项目是否有公共分组（反馈/引用的隐式关联依据）
func ShareGroup(ctx context.Context, projectA, projectB int) bool {
	if projectA == projectB || projectA <= 0 || projectB <= 0 {
		return false
	}
	cnt, err := g.DB().Model("project_group_members a").Ctx(ctx).
		InnerJoin("project_group_members b", "a.group_id = b.group_id").
		Where("a.project_id", projectA).
		Where("b.project_id", projectB).
		Count()
	return err == nil && cnt > 0
}
