// Package agent 外部 Agent 接入协议实现（dev-docs/agent-protocol.md）。
// Agent 仅作身份标识（sys_users type=ai + bc_ key），与项目多对多解耦：
// 注册（公开纯身份）→ 项目接入码 join → 工作会话（agent+project 键）
package agent

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/agent"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/dbinit"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ---- key 工具（与 aiuser 包同实现；两侧独立演进故不共享） ----

func generateApiKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "bc_" + hex.EncodeToString(b)
}

func generateJoinCode() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "bcg_" + hex.EncodeToString(b)
}

func hashApiKey(apiKey, salt string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	mac.Write([]byte(apiKey))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ==================== 注册（公开：纯身份，无权限） ====================

// regLimit 注册端点 IP 限流：公开无认证端点必须防刷（每 IP 每小时 5 次）。
// 单机内存实现与部署形态匹配；多实例部署时需换共享存储
var regLimit = struct {
	mu sync.Mutex
	// per-IP 滑动窗口；global 为不分来源的总量窗口——gf 的 GetClientIp
	// 无条件信任 XFF，伪造随机 IP 即绕过每 IP 限制，全局配额封住刷号总量
	seen   map[string][]time.Time
	global []time.Time
}{seen: map[string][]time.Time{}}

const (
	regLimitWindow     = time.Hour
	regLimitPerIP      = 5
	regLimitGlobalHour = 50
	// 条目 TTL 与容量上限：防伪造 XFF 的随机 IP 把 map 撑到无限大
	// （内存 DoS，同 aiKeyGuard 的防护口径）
	regLimitEntryTTL   = 2 * time.Hour
	regLimitMaxEntries = 10000
)

func registerRateLimited(ip string) bool {
	regLimit.mu.Lock()
	defer regLimit.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-regLimitWindow)
	// TTL 清理：窗口外且早于 TTL 的条目直接删除（原实现只截断 slice，
	// map key 永不清理）
	for k, ts := range regLimit.seen {
		if len(ts) == 0 || ts[len(ts)-1].Before(now.Add(-regLimitEntryTTL)) {
			delete(regLimit.seen, k)
		}
	}
	// 容量兜底：仍超量时按"最后一次记录时间"淘汰最旧的一半
	if len(regLimit.seen) >= regLimitMaxEntries {
		type kv struct {
			k string
			t time.Time
		}
		all := make([]kv, 0, len(regLimit.seen))
		for k, ts := range regLimit.seen {
			all = append(all, kv{k, ts[len(ts)-1]})
		}
		sort.Slice(all, func(i, j int) bool { return all[i].t.Before(all[j].t) })
		for _, e := range all[:len(all)-regLimitMaxEntries/2] {
			delete(regLimit.seen, e.k)
		}
	}
	// 全局配额（不分来源）：正常使用远够不到，换 IP 刷号被封总量
	keptGlobal := regLimit.global[:0]
	for _, t := range regLimit.global {
		if t.After(cutoff) {
			keptGlobal = append(keptGlobal, t)
		}
	}
	regLimit.global = keptGlobal
	if len(regLimit.global) >= regLimitGlobalHour {
		return false
	}
	// per-IP 滑动窗口
	kept := regLimit.seen[ip][:0]
	for _, t := range regLimit.seen[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= regLimitPerIP {
		regLimit.seen[ip] = kept
		return false
	}
	kept = append(kept, now)
	regLimit.seen[ip] = kept
	regLimit.global = append(regLimit.global, now)
	return true
}

func Register(ctx context.Context, req *api.RegisterReq) (*api.RegisterRes, error) {
	// trim 前置：v 校验发生在原始串上，"  a  " 会被放行后 trim 成 1 字符
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 2 || len(req.Name) > 64 {
		return nil, fmt.Errorf("名称长度须为 2-64")
	}
	ip := "unknown"
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
	}
	if !registerRateLimited(ip) {
		return nil, fmt.Errorf("注册过于频繁，请稍后再试")
	}
	// 匿名审计：公开端点 uid=0 会被审计中间件跳过，这里显式留痕
	g.Log().Infof(ctx, "agent register: name=%s ip=%s", req.Name, ip)
	name := req.Name
	apiKey := generateApiKey()
	salt := generateSalt()

	// username 全局唯一（sys_users UNIQUE 兜底 + 友好提示）
	cnt, err := g.DB().Model("sys_users").Ctx(ctx).Where("username", name).Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "注册失败")
	}
	if cnt > 0 {
		return nil, fmt.Errorf("名称已存在：%s", name)
	}
	result, err := g.DB().Model("sys_users").Ctx(ctx).Insert(g.Map{
		"username":       name,
		"password":       "",
		"real_name":      name,
		"type":           "ai",
		"capabilities":   req.Capabilities,
		"api_key":        hashApiKey(apiKey, salt),
		"api_key_salt":   salt,
		"owner_human_id": 0, // 自助注册无属主
		"status":         1,
	})
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "注册失败")
	}
	lastId, _ := result.LastInsertId()
	return &api.RegisterRes{AgentId: int(lastId), ApiKey: apiKey}, nil
}

// ==================== 项目接入码 ====================

// JoinCodeCreate 项目 owner/超管生成一次性接入码（24h）
func JoinCodeCreate(ctx context.Context, projectId int) (*api.JoinCodeCreateRes, error) {
	uid := perm.UserId(ctx)
	if !perm.IsProjectOwner(ctx, uid, projectId) {
		return nil, fmt.Errorf("仅项目管理员可生成接入码")
	}
	code := generateJoinCode()
	expires := gtime.Now().Add(24 * time.Hour)
	if _, err := g.DB().Model("agent_join_codes").Ctx(ctx).Insert(g.Map{
		"code":       code,
		"project_id": projectId,
		"role":       "member",
		"created_by": uid,
		"expires_at": expires,
	}); err != nil {
		return nil, liberr.WrapDb(ctx, err, "生成接入码失败")
	}
	return &api.JoinCodeCreateRes{Code: code, ExpiresAt: expires.Format("Y-m-d H:i:s")}, nil
}

// Join Agent（bc key 已认证）凭接入码加入项目；幂等（重复 join 同项目无害）
func Join(ctx context.Context, code string) (*api.JoinRes, error) {
	agentId := perm.UserId(ctx) // TokenAuth 的 bc_ 分支注入的是 sys_users.id

	row, err := g.DB().Model("agent_join_codes").Ctx(ctx).
		Where("code", strings.TrimSpace(code)).One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "验证接入码失败")
	}
	if row.IsEmpty() {
		return nil, fmt.Errorf("接入码不存在")
	}
	if !row["used_at"].GTime().IsZero() {
		return nil, fmt.Errorf("接入码已被使用")
	}
	if gtime.Now().After(row["expires_at"].GTime()) {
		return nil, fmt.Errorf("接入码已过期")
	}
	projectId := row["project_id"].Int()

	// 码一次性消费 + 准入落库同事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		res, e := tx.Ctx(ctx).Model("agent_join_codes").Where("id", row["id"].Int()).
			Where("used_at IS NULL"). // 并发双兑同码：仅一个成功
			Data(g.Map{"used_at": gtime.Now(), "used_by": agentId}).Update()
		if e != nil {
			return e
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("接入码已被使用")
		}
		// SQLite: ON CONFLICT DO NOTHING / MySQL: INSERT IGNORE（都依赖
		// agent_project_bindings 的 UNIQUE(agent_id, project_id)）
		bindingSQL := `INSERT INTO agent_project_bindings (agent_id, project_id, role) VALUES (?,?,?)
			 ON CONFLICT(agent_id, project_id) DO NOTHING`
		if dbinit.Dialect() == "mysql" {
			bindingSQL = `INSERT IGNORE INTO agent_project_bindings (agent_id, project_id, role) VALUES (?,?,?)`
		}
		if _, e := tx.Ctx(ctx).Exec(bindingSQL, agentId, projectId, row["role"].String()); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "加入项目失败")
	}
	prow, _ := g.DB().Model("projects").Ctx(ctx).Where("id", projectId).Fields("name, code").One()
	return &api.JoinRes{ProjectId: projectId, ProjectCode: prow["code"].String(), ProjectName: prow["name"].String()}, nil
}

// ==================== 工作会话 ====================

// SessionCreate 建立/复用会话（键=agent+project）并返回开工包。
// 项目标识：code 优先（跨环境稳定），数字 id 兼容
func SessionCreate(ctx context.Context, req *api.SessionCreateReq) (*api.SessionCreateRes, error) {
	agentId := perm.UserId(ctx)
	// 解析项目：code → id
	if req.ProjectCode != "" {
		v, err := g.DB().Model("projects").Ctx(ctx).Where("code", req.ProjectCode).Fields("id").Value()
		if err != nil || v == nil {
			return nil, fmt.Errorf("项目 code 不存在：%s", req.ProjectCode)
		}
		req.ProjectId = v.Int()
	}
	if req.ProjectId <= 0 {
		return nil, fmt.Errorf("projectId 或 projectCode 必须提供一个")
	}

	// 准入校验（不经会话实时判定）
	bound, err := g.DB().Model("agent_project_bindings").Ctx(ctx).
		Where("agent_id", agentId).Where("project_id", req.ProjectId).Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询项目准入失败")
	}
	if bound == 0 {
		return nil, fmt.Errorf("未加入该项目，请先凭接入码加入")
	}

	// 会话键=(agent,project)：存在则续期复用，否则签发
	sid := ""
	existing, err := g.DB().Model("agent_sessions").Ctx(ctx).
		Where("agent_id", agentId).Where("project_id", req.ProjectId).One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询会话失败")
	}
	expires := gtime.Now().Add(24 * time.Hour)
	if !existing.IsEmpty() {
		sid = existing["session_id"].String()
		_, err = g.DB().Model("agent_sessions").Ctx(ctx).
			Where("agent_id", agentId).Where("project_id", req.ProjectId).
			Data("expires_at", expires).Update()
	} else {
		sid = "bcsh_" + generateSalt()
		_, err = g.DB().Model("agent_sessions").Ctx(ctx).Insert(g.Map{
			"session_id": sid,
			"agent_id":   agentId,
			"project_id": req.ProjectId,
			"expires_at": expires,
		})
	}
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "建立会话失败")
	}

	res, err := buildContextPack(ctx, agentId, req.ProjectId)
	if err != nil {
		return nil, err
	}
	res.SessionId = sid
	return res, nil
}

// buildContextPack 开工包：项目 + 全局/项目约定 + 我的任务 + 我提交的待审
func buildContextPack(ctx context.Context, agentId, projectId int) (*api.SessionCreateRes, error) {
	prow, perr := g.DB().Model("projects").Ctx(ctx).Where("id", projectId).Fields("name, code").One()
	if perr != nil || prow.IsEmpty() {
		return nil, fmt.Errorf("查询项目失败")
	}
	pname := prow["name"].String()
	pcode := prow["code"].String()
	res := &api.SessionCreateRes{
		Project: api.ProjectBrief{Id: projectId, Code: pcode, Name: pname},
	}

	// 约定：全局 conventions.* + 项目全部记忆（与 kb_get_conventions 同口径，
	// 值截断 600 字；过期/腐化跳过）
	rows, err := g.DB().Model("project_memories").Ctx(ctx).
		Fields("project_id, key, value, status, expires_at, last_verified_at").
		Where("(project_id = 0 AND `key` LIKE 'conventions.%') OR project_id = ?", projectId).
		Where("status IN (?)", g.Slice{"pending", "active"}).
		Order("project_id ASC, `key` ASC").Limit(100).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询记忆失败")
	}
	for _, r := range rows {
		v := r["value"].String()
		if len(v) > 600 {
			v = v[:600] + "…"
		}
		scope := "project"
		if r["project_id"].Int64() == 0 {
			scope = "global"
		}
		res.Conventions = append(res.Conventions, api.ConventionItem{
			Key: r["key"].String(), Value: v, Scope: scope,
		})
	}

	// 我的任务（assignee=agent，未完成三态）
	myRows, err := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, title, status, type, priority, due_date, updated_at").
		Where("assignee_id", agentId).
		Where("project_id", projectId).
		Where("status IN (?)", g.Slice{"open", "in_progress", "review"}).
		Order("priority DESC, updated_at DESC").Limit(50).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务失败")
	}
	for _, r := range myRows {
		res.MyTasks = append(res.MyTasks, rowToTaskBrief(r))
	}

	// 我提交的待审（claim 的任务完成进入 review）
	rvRows, err := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, title, status, type, priority, due_date, updated_at").
		Where("project_id", projectId).
		Where("assignee_id", agentId).
		Where("status", "review").
		Limit(50).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询待审失败")
	}
	for _, r := range rvRows {
		res.PendingReviews = append(res.PendingReviews, rowToTaskBrief(r))
	}
	fillTaskBriefTags(ctx, res.MyTasks)
	fillTaskBriefTags(ctx, res.PendingReviews)

	// 待分析的跨项目反馈：对方项目投递的线索，由本 agent 阅读后决定
	// 是否建任务（POST /v1/projects/{pid}/feedbacks/{id}/convert|dismiss）
	res.PendingFeedbacks = []api.FeedbackBrief{}
	fbRows, ferr := g.DB().Model("project_feedbacks f").Ctx(ctx).
		LeftJoin("projects p", "p.id = f.source_project_id").
		Fields("f.id, f.title, f.content, p.name AS source_project_name, f.source_task_id").
		Where("f.project_id", projectId).Where("f.status", "open").
		Order("f.id ASC").Limit(20).All()
	if ferr == nil {
		for _, r := range fbRows {
			res.PendingFeedbacks = append(res.PendingFeedbacks, api.FeedbackBrief{
				Id: r["id"].Int(), Title: r["title"].String(), Content: r["content"].String(),
				SourceProjectName: r["source_project_name"].String(), SourceTaskId: r["source_task_id"].Int(),
			})
		}
	}
	// 分配给本 agent 的进行中专题（长任务工作流入口）
	res.ActiveTopics = []api.TopicBrief{}
	tpRows, terr := g.DB().Model("topics t").Ctx(ctx).
		Where("t.project_id", projectId).Where("t.status", "active").
		Where("t.assignee_id", agentId).Order("t.id ASC").Limit(10).All()
	if terr == nil {
		for _, r := range tpRows {
			brief := api.TopicBrief{
				Id: r["id"].Int(), Title: r["title"].String(),
				Goal: r["goal"].String(), DocPath: r["doc_path"].String(),
			}
			if ph, _ := g.DB().Model("topic_phases").Ctx(ctx).Where("topic_id", brief.Id).Fields("COUNT(*) AS c, SUM(CASE WHEN status='done' THEN 1 ELSE 0 END) AS d").One(); !ph.IsEmpty() {
				brief.PhaseTotal = ph["c"].Int()
				brief.PhaseDone = ph["d"].Int()
			}
			if hv, _ := g.DB().Model("ai_execution_logs").Ctx(ctx).Where("topic_id", brief.Id).Where("action", "handoff").Order("id DESC").Fields("detail").Value(); hv != nil {
				brief.LastHandoff = hv.String()
			}
			res.ActiveTopics = append(res.ActiveTopics, brief)
		}
	}
	// 高频 QA（按命中数前 5）：新会话最可能用到的问题先给
	res.TopQas = []api.QaBrief{}
	if qaRows, qerr := g.DB().Model("project_qas").Ctx(ctx).
		Where("project_id", projectId).Where("status", "active").
		Order("hits DESC, id DESC").Limit(5).All(); qerr == nil {
		for _, qr := range qaRows {
			ans := qr["answer"].String()
			if len(ans) > 400 {
				ans = ans[:400]
			}
			res.TopQas = append(res.TopQas, api.QaBrief{
				Id: qr["id"].Int(), Question: qr["question"].String(), Answer: ans, Hits: qr["hits"].Int(),
			})
		}
	}
	return res, nil
}

func rowToTaskBrief(r gdb.Record) api.TaskBrief {
	return api.TaskBrief{
		Id: r["id"].Int(), Title: r["title"].String(), Status: r["status"].String(),
		Type: r["type"].String(), Priority: r["priority"].Int(),
		DueDate: r["due_date"].String(), UpdatedAt: r["updated_at"].String(),
	}
}

// fillTaskBriefTags 批量回填任务标签（单条 N+1 在 50 条列表下不可接受）
func fillTaskBriefTags(ctx context.Context, briefs []api.TaskBrief) {
	ids := make([]int, 0, len(briefs))
	for _, b := range briefs {
		ids = append(ids, b.Id)
	}
	if len(ids) == 0 {
		return
	}
	rows, err := g.DB().Model("entity_tags et").Ctx(ctx).
		InnerJoin("tags t", "t.id = et.tag_id").
		Fields("et.entity_id, t.name").
		Where("et.entity_type", "task").
		Where("et.entity_id IN (?)", ids).
		Order("et.entity_id ASC, t.name ASC").All()
	if err != nil {
		return
	}
	m := make(map[int][]string)
	for _, r := range rows {
		id := r["entity_id"].Int()
		m[id] = append(m[id], r["name"].String())
	}
	for i := range briefs {
		briefs[i].Tags = m[briefs[i].Id]
	}
}

// RemoveAgentProject owner 移除 agent 的项目准入：binding 与该项目的会话一并清除
// （会话只做路由不做权限，但清除可让免参端点立即 403 而非等到过期）
func RemoveAgentProject(ctx context.Context, projectId, agentId, operator int) error {
	if !perm.IsProjectOwner(ctx, operator, projectId) {
		return fmt.Errorf("仅项目管理员可移除 Agent 准入")
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Exec(
			"DELETE FROM agent_project_bindings WHERE agent_id = ? AND project_id = ?", agentId, projectId); err != nil {
			return err
		}
		_, err := tx.Exec(
			"DELETE FROM agent_sessions WHERE agent_id = ? AND project_id = ?", agentId, projectId)
		return err
	})
}

// ==================== 会话解析（免参端点用） ====================

// ResolveSession 由 X-Session 解析 (agentId, projectId)；会话属于谁就只认谁
func ResolveSession(ctx context.Context, sessionId string, claimedAgent int) (int, error) {
	if sessionId == "" {
		return 0, fmt.Errorf("缺少 X-Session 会话头")
	}
	row, err := g.DB().Model("agent_sessions").Ctx(ctx).
		Where("session_id", sessionId).One()
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "查询会话失败")
	}
	if row.IsEmpty() {
		return 0, fmt.Errorf("会话不存在")
	}
	if row["agent_id"].Int() != claimedAgent {
		return 0, fmt.Errorf("会话不属于当前 Agent")
	}
	if gtime.Now().After(row["expires_at"].GTime()) {
		return 0, fmt.Errorf("会话已过期，请重建")
	}
	// 复验项目准入：会话只做路由不做权限，但 binding 被移除后免参端点
	// 应立即失效（不等 24h 过期）——权限主体端点由 CanAccessProject 实时挡
	projectId := row["project_id"].Int()
	if cnt, _ := g.DB().Model("agent_project_bindings").Ctx(ctx).
		Where("agent_id", claimedAgent).Where("project_id", projectId).Count(); cnt == 0 {
		return 0, fmt.Errorf("项目准入已被移除")
	}
	return projectId, nil
}

// ==================== 免参任务列表 ====================

// AgentProjects bc key 认证下的已接入项目列表（免参，跨项目全量）
func AgentProjects(ctx context.Context, agentId int) (res *api.AgentProjectsRes, err error) {
	res = &api.AgentProjectsRes{List: []api.ProjectBrief{}}
	rows, qerr := g.DB().Model("agent_project_bindings b").Ctx(ctx).
		InnerJoin("projects p", "p.id = b.project_id").
		Fields("p.id, p.code, p.name").
		Where("b.agent_id", agentId).
		Order("p.id ASC").All()
	if qerr != nil {
		return nil, liberr.WrapDb(ctx, qerr, "查询项目列表失败")
	}
	for _, r := range rows {
		res.List = append(res.List, api.ProjectBrief{
			Id: r["id"].Int(), Code: r["code"].String(), Name: r["name"].String(),
		})
	}
	return res, nil
}

func AgentTasks(ctx context.Context, agentId int, session, status, keyword string) (*api.AgentTasksRes, error) {
	projectId, err := ResolveSession(ctx, session, agentId)
	if err != nil {
		return nil, err
	}
	res := &api.AgentTasksRes{}
	statuses := g.Slice{"open", "in_progress", "review"}
	if status == "all" {
		statuses = nil
	} else if status != "" {
		statuses = g.Slice{status}
	}
	// Count 与数据查询分开构建：带 Fields 的 model 直接 Count 会生成
	// COUNT(多列) 非法 SQL
	base := func() *gdb.Model {
		m := g.DB().Model("tasks").Ctx(ctx).Where("project_id", projectId)
		if statuses != nil {
			m = m.Where("status IN (?)", statuses)
		}
		if keyword != "" {
			m = m.WhereLike("title", "%"+keyword+"%")
		}
		return m
	}
	if res.Total, err = base().Count(); err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务失败")
	}
	rows, err := base().
		Fields("id, title, status, type, priority, due_date, updated_at").
		Order("priority DESC, updated_at DESC").Limit(200).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务失败")
	}
	for _, r := range rows {
		res.List = append(res.List, rowToTaskBrief(r))
	}
	fillTaskBriefTags(ctx, res.List)
	return res, nil
}
