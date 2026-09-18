package project

import (
	"context"
	"database/sql"
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
	// 分组私有化（#480）：任何人类用户可建自己的分组（归属即边界），
	// agent 不参与组织结构。此前挂 platform_groups 菜单权限——私有化后
	// 普通项目 owner 建不了自己的分组，与归属语义矛盾，故移除
	if isAgent, aerr := actorIsAgent(ctx); aerr != nil {
		return 0, aerr
	} else if isAgent {
		return 0, fmt.Errorf("分组为人类用户的组织结构，Agent 不可操作")
	}
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
	// 分组归属创建者（数据隔离）：普通用户只见自己的分组；超管全量
	// 可见（监管视角，靠 ownerName 区分归属）
	uid := perm.UserId(ctx)
	m := g.DB().Model("project_groups g").Ctx(ctx).
		LeftJoin("sys_users u", "u.id = g.created_by").
		Fields("g.id, g.name, g.description, g.created_by, g.created_at, COALESCE(NULLIF(u.real_name, ''), u.username) AS owner_name")
	if !perm.IsAdmin(ctx, uid) {
		m = m.Where("g.created_by", uid)
	}
	rows, err := m.Order("g.id ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询分组失败")
	}
	// 手工组装：gconv 的 json tag 匹配对蛇形别名不生效（owner_name 落不进
	// OwnerName），代码库其余 JOIN 查询同为手工行的模式
	groups := make([]api.GroupListItem, 0, len(rows))
	for _, r := range rows {
		groups = append(groups, api.GroupListItem{
			Id: r["id"].Int(), Name: r["name"].String(), Description: r["description"].String(),
			CreatedBy: r["created_by"].Int(), OwnerName: r["owner_name"].String(), CreatedAt: r["created_at"].String(),
			Projects: []api.GroupProjectBrief{},
		})
	}
	// 批量回填成员项目（限定本页分组，避免全表捞）
	if len(groups) > 0 {
		ids := make([]int, 0, len(groups))
		for _, gr := range groups {
			ids = append(ids, gr.Id)
		}
		var members []struct {
			GroupId int
			Id      int
			Code    string
			Name    string
		}
		g.DB().Model("project_group_members pgm").Ctx(ctx).
			InnerJoin("projects p", "p.id = pgm.project_id").
			Fields("pgm.group_id, p.id, p.code, p.name").
			Where("pgm.group_id IN (?)", ids).
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
	if err := groupOwnedByCreator(ctx, req.Id); err != nil {
		return err
	}
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
	if err := groupOwnedByCreator(ctx, id); err != nil {
		return err
	}
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
	// 把项目挂进分组：双重门槛——项目侧 owner/maintainer（超管直通）+
	// 分组归属者（分组私有化后只能组织自己的分组）
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), req.ProjectId) {
		return fmt.Errorf("仅项目管理员可调整分组")
	}
	if err := groupOwnedByCreator(ctx, req.Id); err != nil {
		return err
	}
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
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), projectId) {
		return fmt.Errorf("仅项目管理员可调整分组")
	}
	if err := groupOwnedByCreator(ctx, groupId); err != nil {
		return err
	}
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

// groupOwnedByCreator 分组归属校验（私有化）：仅创建者可操作，超管直通。
// 不存在的分组按越权报出（更新/删除原实现对不存在 id 静默成功，一并修正）
func groupOwnedByCreator(ctx context.Context, groupId int) error {
	uid := perm.UserId(ctx)
	row, err := g.DB().Model("project_groups").Ctx(ctx).
		Where("id", groupId).Fields("created_by").One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询分组失败")
	}
	if row.IsEmpty() {
		return fmt.Errorf("分组不存在")
	}
	if row["created_by"].Int() != uid && !perm.IsAdmin(ctx, uid) {
		return fmt.Errorf("仅分组创建者可操作该分组")
	}
	return nil
}

// detachOwnerGroupsSQL 项目移出指定用户名下全部分组（移交自动退出，#480）。
// 只动该用户创建的分组——第三方（如超管）建的跨项目分组不随人事变动
const detachOwnerGroupsSQL = `DELETE FROM project_group_members WHERE project_id = ? AND group_id IN (SELECT id FROM project_groups WHERE created_by = ?)`

// detachOwnerGroups 执行上述移出并返回退出的分组数（0=无变化）。
// exec 闭包抹平 g.DB()（带 ctx）与 gdb.TX（不带）的 Exec 签名差异，
// 供移交事务内调用；ownerId<=0（无时任 owner）视为无变化
func detachOwnerGroups(exec func(sql string, args ...interface{}) (sql.Result, error), projectId, ownerId int) (int64, error) {
	if ownerId <= 0 || projectId <= 0 {
		return 0, nil
	}
	res, err := exec(detachOwnerGroupsSQL, projectId, ownerId)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
