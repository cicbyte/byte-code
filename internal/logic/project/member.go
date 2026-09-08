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
	// owner 只能由项目创建流程产生：API 传 owner 会凭空造出第二个 owner，
	// 绕过 IsProjectOwner 的唯一性假设
	if req.Role != "" && req.Role != "member" {
		return fmt.Errorf("仅支持添加普通成员（role=member）")
	}
	// 用户必须真实存在（表无 FK 强制），且不允许把 AI 账号加为项目成员
	uType, uerr := g.DB().Model("sys_users").Ctx(ctx).Where("id", req.UserId).Fields("type").Value()
	if uerr != nil || uType == nil {
		return fmt.Errorf("用户不存在")
	}
	if uType.String() == "ai" {
		return fmt.Errorf("Agent 账号不能添加为项目成员")
	}
	// 重复添加转友好提示（UNIQUE(project_id,user_id) 裸错误对用户无意义）
	if cnt, _ := g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", req.ProjectId).Where("user_id", req.UserId).Count(); cnt > 0 {
		return fmt.Errorf("该用户已是项目成员")
	}
	_, err = g.DB().Model("project_members").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"user_id":    req.UserId,
		"role":       "member",
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
	// 成员两来源合并：project_members（人/早期手动加的 agent）+
	// agent_project_bindings（协议接入的 agent）。重叠去重以 members 行
	// 为准（老数据兼容）；agent 行的移除走 DeleteAgentProject 而非
	// removeMember——前端按 user_type 分流
	var list []api.MemberItem
	err = g.DB().Ctx(ctx).Raw(`
		SELECT * FROM (
			SELECT pm.id, pm.user_id, u.username, COALESCE(u.real_name, '') AS real_name,
			       pm.role, pm.created_at AS joined_at,
			       COALESCE(u.type, 'human') AS user_type, 0 AS via_binding
			FROM project_members pm LEFT JOIN sys_users u ON pm.user_id = u.id
			WHERE pm.project_id = ?
			UNION ALL
			SELECT b.id, b.agent_id, u.username, COALESCE(u.real_name, '') AS real_name,
			       b.role, b.joined_at, 'ai' AS user_type, 1 AS via_binding
			FROM agent_project_bindings b LEFT JOIN sys_users u ON b.agent_id = u.id
			WHERE b.project_id = ?
			  AND b.agent_id NOT IN (SELECT user_id FROM project_members WHERE project_id = ?)
			) t ORDER BY user_type ASC, id ASC`, projectId, projectId, projectId).
		Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询成员列表失败")
	}
	res.List = list
	return res, nil
}
