package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// 项目成员管理

func (s *sProject) AddMember(ctx context.Context, req *api.MemberAddReq) (err error) {
	// 成员管理属管理级操作，owner/maintainer 可执行（超管直通）
	if uid := perm.UserId(ctx); !perm.IsProjectMaintainer(ctx, uid, req.ProjectId) {
		return fmt.Errorf("仅项目管理员可管理成员")
	}
	// owner 只能由项目创建流程/转交产生；maintainer 档仅 owner 可授予
	// （maintainer 不可拉平级：与移除侧「仅 owner 可动 maintainer」对称）
	role := req.Role
	if role == "" {
		role = consts.MemberRoleMember
	}
	if role != consts.MemberRoleMember && role != consts.MemberRoleMaintainer {
		return fmt.Errorf("不支持的角色（可选 member/maintainer）")
	}
	if role == consts.MemberRoleMaintainer && !perm.IsProjectOwner(ctx, perm.UserId(ctx), req.ProjectId) {
		return fmt.Errorf("仅项目负责人可添加维护者")
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
		"role":       role,
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "添加成员失败")
	}

	// 发送通知
	notify.Send(ctx, req.UserId, "加入项目", fmt.Sprintf("您已被添加到项目中"), "info", "project", req.ProjectId)
	return nil
}

func (s *sProject) RemoveMember(ctx context.Context, projectId, userId int) (err error) {
	// 成员管理属管理级操作，owner/maintainer 可执行（超管直通）
	if uid := perm.UserId(ctx); !perm.IsProjectMaintainer(ctx, uid, projectId) {
		return fmt.Errorf("仅项目管理员可管理成员")
	}
	// 目标按档位保护：owner 不可被移除（避免无主项目）；maintainer 仅
	// owner/超管可移除（maintainer 不能动平级，PRD §4.1）
	if v, _ := g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).Where("user_id", userId).Fields("role").Value(); v != nil {
		switch v.String() {
		case consts.MemberRoleOwner:
			if !perm.IsAdmin(ctx, perm.UserId(ctx)) {
				if userId == perm.UserId(ctx) {
					// owner 自己移自己 = 退出，暂不允许（避免无主项目）
					return fmt.Errorf("项目管理员不能移除自己")
				}
				return fmt.Errorf("不能移除项目负责人")
			}
		case consts.MemberRoleMaintainer:
			if !perm.IsProjectOwner(ctx, perm.UserId(ctx), projectId) {
				return fmt.Errorf("仅项目负责人可移除维护者")
			}
		}
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

func (s *sProject) LeaveMember(ctx context.Context, projectId int) (err error) {
	uid := perm.UserId(ctx)
	// 仅人类成员自助退出；agent 的准入移除走管理侧（RemoveAgentProject）
	if v, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", uid).Fields("type").Value(); v != nil && v.String() == "ai" {
		return fmt.Errorf("Agent 准入请由项目管理员移除")
	}
	row, err := g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).Where("user_id", uid).Fields("role").Value()
	if err != nil || row == nil {
		return fmt.Errorf("您不是本项目成员")
	}
	// owner 退出会产生无主项目：必须先转交（单人 owner 不变量与转交/移除侧一致）
	if row.String() == consts.MemberRoleOwner {
		return fmt.Errorf("项目负责人不能直接退出，请先转交负责人")
	}
	if _, err = g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).Where("user_id", uid).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "退出项目失败")
	}
	// 留痕 + 通知 owner（退出对项目而言是被动事件）
	s.recordActivity(ctx, uid, "project.member_left", "project", projectId,
		fmt.Sprintf("#%d", uid), projectId, "退出了项目")
	if ov, _ := g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", projectId).Where("role", consts.MemberRoleOwner).Fields("user_id").Value(); ov != nil {
		un, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", uid).Fields("username").Value()
		name := ""
		if un != nil {
			name = un.String()
		}
		notify.Send(ctx, ov.Int(), "成员退出项目", fmt.Sprintf("成员「%s」已退出项目", name), "warning", "project", projectId)
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
			       COALESCE(u.type, 'human') AS user_type, 0 AS via_binding,
			       '' AS capabilities
			FROM project_members pm LEFT JOIN sys_users u ON pm.user_id = u.id
			WHERE pm.project_id = ?
			UNION ALL
			SELECT b.id, b.agent_id, u.username, COALESCE(u.real_name, '') AS real_name,
			       b.role, b.joined_at, 'ai' AS user_type, 1 AS via_binding,
			       COALESCE(b.capabilities, '') AS capabilities
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

// TransferOwner 转移项目负责人：owner 终身制的唯一出口。
// 约束：仅当前 owner 或超管可执行（IsProjectOwner 同口径）；目标必须是已存在的
// 项目人类成员（agent 不能担任负责人）；事务内换主——旧 owner 降为 member、
// 新 owner 升为 owner，维持 project_members 中 owner 唯一的不变量
func (s *sProject) TransferOwner(ctx context.Context, req *api.OwnerTransferReq) (err error) {
	uid := perm.UserId(ctx)
	if !perm.IsProjectOwner(ctx, uid, req.ProjectId) {
		return fmt.Errorf("仅项目负责人可转移负责人")
	}
	if req.UserId == uid {
		return fmt.Errorf("不能转移给自己")
	}

	// 目标校验：必须是 project_members 里的真人（绑定 agent 不在 members 表，
	// 此处一并挡掉早期手动加入的 ai 账号）
	var target struct {
		Role     string `json:"role"`
		UserType string `json:"user_type"`
		Name     string `json:"name"`
	}
	terr := g.DB().Model("project_members pm").Ctx(ctx).
		Fields("pm.role, COALESCE(u.type,'human') AS user_type, COALESCE(u.real_name, u.username) AS name").
		LeftJoin("sys_users u", "u.id = pm.user_id").
		Where("pm.project_id", req.ProjectId).
		Where("pm.user_id", req.UserId).
		Scan(&target)
	if terr != nil || target.Role == "" {
		return fmt.Errorf("目标用户不是项目成员")
	}
	if target.UserType == "ai" {
		return fmt.Errorf("Agent 不能担任项目负责人")
	}
	if target.Role == "owner" {
		return fmt.Errorf("目标已是项目负责人")
	}

	// 原负责人（leave 语义作用于交接方；超管代办时同为移出原负责人）
	oldOwnerV, _ := g.DB().Model("project_members").Ctx(ctx).
		Where("project_id", req.ProjectId).
		Where("role", "owner").Fields("user_id").Value()
	oldOwner := 0
	if oldOwnerV != nil {
		oldOwner = oldOwnerV.Int()
	}

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先处理旧 owner 再升新：owner 唯一性没有 DB 约束（role 无 UNIQUE），靠顺序保证。
		// leave=原负责人退出（成员行删除，隔离交接）；缺省=降为普通成员
		if req.Leave {
			if _, e := tx.Model("project_members").Ctx(ctx).
				Where("project_id", req.ProjectId).
				Where("role", "owner").
				Delete(); e != nil {
				return e
			}
		} else if _, e := tx.Model("project_members").Ctx(ctx).
			Where("project_id", req.ProjectId).
			Where("role", "owner").
			Data(g.Map{"role": "member"}).Update(); e != nil {
			return e
		}
		if _, e := tx.Model("project_members").Ctx(ctx).
			Where("project_id", req.ProjectId).
			Where("user_id", req.UserId).
			Data(g.Map{"role": "owner"}).Update(); e != nil {
			return e
		}
		// 分组私有化（#480）：项目自动退出原负责人名下分组（事务内）
		_, derr := detachOwnerGroups(tx.Exec, req.ProjectId, oldOwner)
		return derr
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "转移负责人失败")
	}

	name := target.Name
	mode := "留在项目"
	if req.Leave {
		mode = "退出项目"
	}
	s.recordActivity(ctx, uid, "project.owner_transfer", "project", req.ProjectId, name, req.ProjectId, fmt.Sprintf("负责人转移给 %s（%s）", name, mode))
	notify.Send(ctx, req.UserId, "成为项目负责人", "项目负责人已转移给您，您现在可以管理成员与项目设置", "info", "project", req.ProjectId)
	// 被移出方非操作者本人时补一条通知（自己操作的转交弹窗已明示，不发噪音）
	if req.Leave && oldOwner > 0 && oldOwner != uid {
		notify.Send(ctx, oldOwner, "已退出项目", "项目负责人已转移，您已退出该项目", "warning", "project", req.ProjectId)
	}
	return nil
}
