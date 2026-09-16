package docs

// 全局记忆提案队列：人人可发 + 管理员审核。
// 独立表而非复用 project_memories.pending——pending 语义是「未确认推测」
// （对读者可见、腐化更快），不是审核队列；未审内容绝不能进读者视野。
// 批准时由审核端走既有 MemSet(project_id=0, active) 落正式记忆，
// 消费侧（开工包/CLI memory --global）零改动。

import (
	"context"
	"fmt"
	"strings"

	api "github.com/cicbyte/byte-code/api/v1/docs"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// GlobalMemoryPropose 提交全局记忆提案（任何认证用户含 agent）。
// 同名占用即拒：已有全局记忆（任意状态）提示走管理员更新；已有同 key
// 未决提案也拒（防队列堆积重复）。
func (s *sVault) GlobalMemoryPropose(ctx context.Context, req *api.GlobalMemoryProposeReq) (id int, err error) {
	key := strings.TrimSpace(strings.Trim(req.Key, "/"))
	if key == "" || strings.Contains(key, "/") {
		return 0, gerror.New("key 不能为空且不含 /（点分层级：build.cmd / conventions.naming）")
	}
	ttl := strings.TrimSpace(req.Ttl)
	if ttl != "" {
		if _, err = parseTtl(ttl); err != nil {
			return 0, err
		}
	}
	exist, eerr := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", 0).Where("key", key).Count()
	if eerr != nil {
		return 0, liberr.WrapDb(ctx, eerr, "查询全局记忆失败")
	}
	if exist > 0 {
		return 0, gerror.Newf("同名全局记忆已存在（%s）：更新请走管理员（平台设置-记忆管理）或先废弃原条目", key)
	}
	pendingCnt, perr := g.DB().Model("global_memory_proposals").Ctx(ctx).
		Where("key", key).Where("status", "submitted").Count()
	if perr != nil {
		return 0, liberr.WrapDb(ctx, perr, "查询待审提案失败")
	}
	if pendingCnt > 0 {
		return 0, gerror.Newf("同名提案已在审核中（%s），请等待管理员处理", key)
	}

	uid := perm.UserId(ctx)
	r, ierr := g.DB().Model("global_memory_proposals").Ctx(ctx).Data(g.Map{
		"key": key, "value": req.Value, "ttl": ttl,
		"note": req.Note, "status": "submitted", "proposed_by": uid,
	}).Insert()
	if ierr != nil {
		return 0, liberr.WrapDb(ctx, ierr, "写入提案失败")
	}
	lastId, _ := r.LastInsertId()

	// 通知平台超管（role_id=1 且启用）——提案不通知就会沉底
	who := "用户"
	if un, uerr := g.DB().Model("sys_users").Ctx(ctx).Where("id", uid).Fields("username").Value(); uerr == nil && un != nil {
		who = un.String()
	}
	admins, aerr := g.DB().Model("sys_user_roles ur").Ctx(ctx).
		InnerJoin("sys_users u", "u.id = ur.user_id").
		Where("ur.role_id", 1).Where("u.status", 1).
		Fields("u.id").All()
	if aerr == nil {
		content := fmt.Sprintf("%s 提交全局记忆提案「%s」（平台设置-记忆管理审核）", who, key)
		if req.Note != "" {
			content += "：" + req.Note
		}
		for _, a := range admins {
			notify.Send(ctx, a["id"].Int(), "收到全局记忆提案", content, "info", "system", 0)
		}
	}
	return int(lastId), nil
}

// GlobalMemoryProposalList 提案列表（仅管理员——路由层 AdminAuth 把关）
func (s *sVault) GlobalMemoryProposalList(ctx context.Context, req *api.GlobalMemoryProposalListReq) (res *api.GlobalMemoryProposalListRes, err error) {
	res = &api.GlobalMemoryProposalListRes{List: []api.GlobalMemoryProposalItem{}}
	m := g.DB().Model("global_memory_proposals p").Ctx(ctx).
		LeftJoin("sys_users u", "u.id = p.proposed_by")
	if req.Status != "all" {
		st := req.Status
		if st == "" {
			st = "submitted" // d:"submitted" 只在 HTTP 绑定生效，直调兜底同语义
		}
		m = m.Where("p.status", st)
	}
	rows, qerr := m.Fields("p.*, u.username AS proposer").Order("p.id DESC").Limit(200).All()
	if qerr != nil {
		return nil, liberr.WrapDb(ctx, qerr, "查询提案列表失败")
	}
	for _, r := range rows {
		res.List = append(res.List, proposalRowToItem(r))
	}
	return res, nil
}

// GlobalMemoryProposalReview 审核提案（仅管理员）：批准 = 走既有 MemSet
// 落 active 正式记忆（读者立即可见）；拒绝 = 必给理由。两种决定都回告提交者。
func (s *sVault) GlobalMemoryProposalReview(ctx context.Context, req *api.GlobalMemoryProposalReviewReq) error {
	row, ferr := g.DB().Model("global_memory_proposals").Ctx(ctx).Where("id", req.Id).One()
	if ferr != nil {
		return liberr.WrapDb(ctx, ferr, "查询提案失败")
	}
	if row.IsEmpty() {
		return gerror.New("提案不存在")
	}
	if row["status"].String() != "submitted" {
		return gerror.Newf("提案已处理过（%s），不可重复审核", row["status"].String())
	}
	uid := perm.UserId(ctx)
	key := row["key"].String()
	proposer := row["proposed_by"].Int()
	now := gtime.Now()

	if req.Decision == "approved" {
		// 同名防御（提交后可能有人工建了同名记忆）
		exist, eerr := g.DB().Model("project_memories").Ctx(ctx).
			Where("project_id", 0).Where("key", key).Count()
		if eerr != nil {
			return liberr.WrapDb(ctx, eerr, "查询全局记忆失败")
		}
		if exist > 0 {
			return gerror.Newf("同名全局记忆已存在（%s），请先处理原条目再审核", key)
		}
		// 审核人是人类管理员，MemSet 的 AgentRequire 人类直通
		if err := s.MemSet(ctx, 0, key, &api.MemorySetReq{
			Value: row["value"].String(), Ttl: row["ttl"].String(), Status: "active",
		}); err != nil {
			return err
		}
		if _, uerr := g.DB().Model("global_memory_proposals").Ctx(ctx).Where("id", req.Id).Data(g.Map{
			"status": "approved", "reviewed_by": uid, "reviewed_at": now,
		}).Update(); uerr != nil {
			return liberr.WrapDb(ctx, uerr, "更新提案状态失败")
		}
		notify.Send(ctx, proposer, "全局记忆提案已采纳",
			fmt.Sprintf("你的提案「%s」已被采纳为全局记忆，即刻对所有项目的 agent 生效", key), "success", "system", 0)
		return nil
	}

	// 拒绝：理由必填（回告提交者，镜像 feedback dismiss 的口径）
	if strings.TrimSpace(req.Reason) == "" {
		return gerror.New("拒绝必须给理由（将回告提交者）")
	}
	if _, uerr := g.DB().Model("global_memory_proposals").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"status": "rejected", "reviewed_by": uid, "reviewed_at": now, "review_reason": req.Reason,
	}).Update(); uerr != nil {
		return liberr.WrapDb(ctx, uerr, "更新提案状态失败")
	}
	notify.Send(ctx, proposer, "全局记忆提案被拒绝",
		fmt.Sprintf("你的提案「%s」被拒绝：%s", key, req.Reason), "warning", "system", 0)
	return nil
}

func proposalRowToItem(r gdb.Record) api.GlobalMemoryProposalItem {
	return api.GlobalMemoryProposalItem{
		Id: r["id"].Int(), Key: r["key"].String(), Value: r["value"].String(),
		Ttl: r["ttl"].String(), Note: r["note"].String(), Status: r["status"].String(),
		ProposedBy: r["proposed_by"].Int(), ProposedByName: r["proposer"].String(),
		ReviewedBy: r["reviewed_by"].Int(), ReviewReason: r["review_reason"].String(),
		CreatedAt: r["created_at"].String(), ReviewedAt: r["reviewed_at"].String(),
	}
}
