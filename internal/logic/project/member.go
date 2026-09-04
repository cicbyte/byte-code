package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

// 项目成员管理

func (s *sProject) AddMember(ctx context.Context, req *api.MemberAddReq) (err error) {
	// 成员管理属管理级操作，仅 owner 或超管可执行
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, req.ProjectId) {
		return fmt.Errorf("仅项目管理员可管理成员")
	}
	_, err = g.DB().Model("project_members").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"user_id":    req.UserId,
		"role":       req.Role,
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "添加成员失败")
	}

	// 发送通知
	notify.Send(ctx, req.UserId, "加入项目", fmt.Sprintf("您已被添加到项目中"), "info", "project", req.ProjectId)
	return nil
}

func (s *sProject) RemoveMember(ctx context.Context, projectId, userId int) (err error) {
	// 成员管理属管理级操作，仅 owner 或超管可执行
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, projectId) {
		return fmt.Errorf("仅项目管理员可管理成员")
	}
	// 不能移除项目 owner（避免最后一个 owner 被移走后项目无主）
	if perm.IsProjectOwner(ctx, userId, projectId) && !perm.IsAdmin(ctx, perm.UserId(ctx)) {
		if userId == perm.UserId(ctx) {
			// owner 自己移自己 = 退出，暂不允许（避免无主项目）
			return fmt.Errorf("项目管理员不能移除自己")
		}
		return fmt.Errorf("不能移除项目管理员")
	}
	_, err = g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).
		Where("user_id", userId).
		Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "移除成员失败")
	}
	return nil
}

func (s *sProject) ListMembers(ctx context.Context, projectId int) (res *api.MemberListRes, err error) {
	res = &api.MemberListRes{}
	var list []api.MemberItem
	err = g.DB().Model("project_members pm").Ctx(ctx).
		LeftJoin("sys_users u", "pm.user_id = u.id").
		Fields("pm.id, pm.user_id, u.username, COALESCE(u.real_name, '') as real_name, pm.role, pm.created_at as joined_at, COALESCE(u.type, 'human') as user_type").
		Where("pm.project_id", projectId).
		Order("pm.id ASC").
		Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询成员列表失败")
	}
	res.List = list
	return res, nil
}
