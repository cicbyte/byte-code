package project

// 项目移交（邀请制）：owner 用户名搜索选择目标 → 目标收通知点接受/拒绝。
// 相比旧的单步转交（TransferOwner），目标不必预先是项目成员（接受时自动
// 入项），且交接需对方显式同意——避免"被移交"的意外。旧端点保留作
// 管理员直通路径。

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// InviteOwnerTransfer 发起移交邀请（仅 owner/超管）：目标须为启用的人类
// 账号且非本人；同项目同时只允许一条 pending（重复发起先撤后立）。
func (s *sProject) InviteOwnerTransfer(ctx context.Context, req *api.OwnerTransferInviteReq) (id int, err error) {
	uid := perm.UserId(ctx)
	if !perm.IsProjectOwner(ctx, uid, req.ProjectId) {
		return 0, fmt.Errorf("仅项目负责人可发起移交")
	}
	if req.UserId == uid {
		return 0, fmt.Errorf("不能移交给自己")
	}
	t, terr := g.DB().Model("sys_users").Ctx(ctx).
		Where("id", req.UserId).Fields("type, status, COALESCE(NULLIF(real_name, ''), username) AS name").One()
	if terr != nil || t.IsEmpty() {
		return 0, fmt.Errorf("目标用户不存在")
	}
	if t["type"].String() == "ai" {
		return 0, fmt.Errorf("不能移交给 Agent 账号")
	}
	if t["status"].Int() != 1 {
		return 0, fmt.Errorf("目标用户已被禁用")
	}
	if cnt, _ := g.DB().Model("project_transfers").Ctx(ctx).
		Where("project_id", req.ProjectId).Where("status", "pending").Count(); cnt > 0 {
		return 0, fmt.Errorf("该项目已有待响应的移交邀请（对方处理或过期前不可重复发起）")
	}

	r, ierr := g.DB().Model("project_transfers").Ctx(ctx).Data(g.Map{
		"project_id": req.ProjectId, "from_user_id": uid, "to_user_id": req.UserId,
		"leave_project": boolToInt(req.Leave), "status": "pending",
	}).Insert()
	if ierr != nil {
		return 0, liberr.WrapDb(ctx, ierr, "写入移交邀请失败")
	}
	lastId, _ := r.LastInsertId()

	pname, _ := g.DB().Model("projects").Ctx(ctx).Where("id", req.ProjectId).Fields("name").Value()
	projectName := ""
	if pname != nil {
		projectName = pname.String()
	}
	after := "原负责人将留在项目降为普通成员"
	if req.Leave {
		after = "原负责人将退出项目"
	}
	notify.Send(ctx, req.UserId, "收到项目移交邀请",
		fmt.Sprintf("「%s」项目负责人想把项目移交给你（%s）。请在通知中心接受或拒绝。", projectName, after),
		"info", "transfer", int(lastId))
	s.recordActivity(ctx, uid, "project.transfer_invited", "project", req.ProjectId,
		fmt.Sprintf("#%d", req.UserId), req.ProjectId, "发起移交邀请")
	return int(lastId), nil
}

// RespondOwnerTransfer 目标本人响应：接受 = 事务内转交（目标非成员则自动
// 入项为 owner，原 owner 按 leave 降级或移出）；拒绝 = 仅回告发起方。
func (s *sProject) RespondOwnerTransfer(ctx context.Context, req *api.OwnerTransferRespondReq) error {
	uid := perm.UserId(ctx)
	row, ferr := g.DB().Model("project_transfers").Ctx(ctx).Where("id", req.Id).One()
	if ferr != nil {
		return liberr.WrapDb(ctx, ferr, "查询移交邀请失败")
	}
	if row.IsEmpty() {
		return gerror.New("移交邀请不存在")
	}
	if row["to_user_id"].Int() != uid {
		return gerror.New("只有受邀本人可响应移交")
	}
	if row["status"].String() != "pending" {
		return gerror.Newf("邀请已处理过（%s）", row["status"].String())
	}
	pid := row["project_id"].Int()
	from := row["from_user_id"].Int()
	leave := row["leave_project"].Int() == 1
	// 定长字符串而非 gtime 对象：gdb 驱动把 gtime.Time 绑定为带微秒
	// （26 字符），MySQL VARCHAR(19) 列报 Data too long（生产实测）。
	// 用 stdlib time 而非 gtime.Format——gtime 的格式 token 是自家的
	// Y-m-d H:i:s，传 Go 布局串会原样返回字面量（MySQL 复验抓出）
	now := time.Now().Format("2006-01-02 15:04:05")

	if req.Action == "decline" {
		if _, uerr := g.DB().Model("project_transfers").Ctx(ctx).Where("id", req.Id).Data(g.Map{
			"status": "declined", "resolved_at": now, "resolved_by": uid,
		}).Update(); uerr != nil {
			return liberr.WrapDb(ctx, uerr, "更新邀请状态失败")
		}
		notify.Send(ctx, from, "移交邀请被拒绝",
			"对方拒绝了你的项目移交邀请（可在成员页重新发起或另选目标）", "warning", "transfer", req.Id)
		s.recordActivity(ctx, uid, "project.transfer_declined", "project", pid,
			fmt.Sprintf("#%d", from), pid, "拒绝移交邀请")
		markInviteRead(ctx, uid, req.Id)
		return nil
	}

	// 接受：角色变更整体事务（单 owner 不变量：目标升 owner、原 owner 降级/移出、
	// 邀请落定三者同生共死）
	var exited int64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先取时任 owner（升新后按角色找人会命中两个 owner）
		curOwnerV, _ := tx.Model("project_members").Ctx(ctx).
			Where("project_id", pid).Where("role", "owner").
			Where("user_id != ?", uid).Fields("user_id").Value()
		curOwner := 0
		if curOwnerV != nil {
			curOwner = curOwnerV.Int()
		}
		// 目标非成员则自动入项（邀请制的意义：不必先加成员）
		cnt, cerr := tx.Model("project_members").Ctx(ctx).
			Where("project_id", pid).Where("user_id", uid).Count()
		if cerr != nil {
			return cerr
		}
		if cnt == 0 {
			if _, ierr := tx.Model("project_members").Ctx(ctx).Data(g.Map{
				"project_id": pid, "user_id": uid, "role": "owner",
			}).Insert(); ierr != nil {
				return ierr
			}
		} else if _, uerr := tx.Model("project_members").Ctx(ctx).
			Where("project_id", pid).Where("user_id", uid).
			Data("role", "owner").Update(); uerr != nil {
			return uerr
		}
		// 原负责人按时任 owner 角色降级/移出（而非邀请人——超管代发场景
		// 邀请人未必是 owner，按人操作会留下双 owner 破坏单 owner 不变量）
		if leave {
			if _, derr := tx.Model("project_members").Ctx(ctx).
				Where("project_id", pid).Where("role", "owner").
				Where("user_id != ?", uid).Delete(); derr != nil {
				return derr
			}
		} else if _, uerr := tx.Model("project_members").Ctx(ctx).
			Where("project_id", pid).Where("role", "owner").
			Where("user_id != ?", uid).
			Data("role", "member").Update(); uerr != nil {
			return uerr
		}
		_, rerr := tx.Model("project_transfers").Ctx(ctx).Where("id", req.Id).Data(g.Map{
			"status": "accepted", "resolved_at": now, "resolved_by": uid,
		}).Update()
		if rerr != nil {
			return rerr
		}
		// 分组私有化（#480）：项目自动退出原负责人名下分组（事务内，
		// 与角色变更同生共死）；第三方分组不动
		var derr error
		exited, derr = detachOwnerGroups(ctx, tx, pid, curOwner)
		return derr
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "移交失败")
	}
	after := "你已成为普通成员"
	if leave {
		after = "你已退出项目"
	}
	groupNote := ""
	if exited > 0 {
		groupNote = fmt.Sprintf("；项目已自动移出你名下 %d 个分组", exited)
	}
	notify.Send(ctx, from, "移交邀请已接受",
		fmt.Sprintf("项目移交已完成（%s）%s。", after, groupNote), "success", "transfer", req.Id)
	s.recordActivity(ctx, uid, "project.transferred", "project", pid,
		fmt.Sprintf("#%d", from), pid, "移交完成，负责人变更")
	markInviteRead(ctx, uid, req.Id)
	return nil
}

// markInviteRead 决议后把受邀人的邀请通知标读：已读 + 列表回填的
// transferStatus 双信号，让前端不再对已决邀请显示操作
func markInviteRead(ctx context.Context, uid, transferId int) {
	g.DB().Model("notifications").Ctx(ctx).
		Where("user_id", uid).
		Where("source_type", "transfer").
		Where("source_id", transferId).
		Data("is_read", 1).Update()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
