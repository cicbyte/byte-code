package aiengine

// 记忆/文档中枢工具：vault 磁盘真相源（utility/vault）+ KV 记忆（project_memories）。
// 与 logic 层同样的语义，工具层直连 g.DB() 保持本包自洽（与 tasks/project 工具一致）。

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cicbyte/byte-code/utility/docs"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ctxAIUserId executor 注入的 AI 用户 id（mem_set 来源标识）
const ctxAIUserId = "aiUserId"

// ctxTaskProjectId executor 注入的任务所属项目 id（安全边界，见 bindProjectId）
const ctxTaskProjectId = "taskProjectId"

func aiUserId(ctx context.Context) int {
	if v, ok := ctx.Value(ctxAIUserId).(int); ok && v > 0 {
		return v
	}
	return 0
}

// bindProjectId 工具安全边界：projectId 以 executor 注入的任务归属为准，
// LLM 参数传入的 projectId 不一致时拒绝执行——防止任务描述诱导 AI 读写
// 其他项目的 vault 文档与记忆（跨项目数据外泄/投毒）
func bindProjectId(ctx context.Context, argsProjectId int) (int, error) {
	bound, ok := ctx.Value(ctxTaskProjectId).(int)
	if !ok || bound <= 0 {
		// 无任务上下文（如手工调试调用）时退回参数值
		if argsProjectId > 0 {
			return argsProjectId, nil
		}
		return 0, fmt.Errorf("缺少任务项目上下文")
	}
	if argsProjectId > 0 && argsProjectId != bound {
		return 0, fmt.Errorf("projectId 与任务所属项目不符（任务属项目 %d），禁止跨项目访问", bound)
	}
	return bound, nil
}

// ==================== 文档工具 ====================

// ---- 读取文档正文 ----

type ReadDocArgs struct {
	ProjectId int    `json:"projectId"`
	Path      string `json:"path"`
}

type ReadDocTool struct{}

func (t *ReadDocTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("read_doc", "读取项目文档正文（markdown 返回 frontmatter 元数据+正文；二进制只返回元数据）",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
			"path":      paramInfo("string", "vault 内相对路径（来自 search_docs/doc_linked 结果）"),
		}, nil,
	), nil
}

func (t *ReadDocTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p ReadDocArgs
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	abs, err := docs.SafeJoin(int64(p.ProjectId), p.Path)
	if err != nil {
		return "", fmt.Errorf("路径非法")
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %s", p.Path)
	}
	ext := strings.ToLower(filepath.Ext(p.Path))
	if !textExt(ext) {
		return marshalString(g.Map{"binary": true, "path": p.Path, "size": len(data),
			"hint": "二进制文件，内容不可读"}), nil
	}
	const maxBody = 8 * 1024
	fm, body, _ := docs.ParseFrontmatter(string(data))
	truncated := false
	if len(body) > maxBody {
		body = body[:maxBody]
		truncated = true
	}
	return marshalString(g.Map{
		"path": p.Path, "title": fm.Title, "space": fm.Space, "type": fm.Type,
		"status": fm.Status, "tags": fm.Tags, "linked": fm.Linked,
		"content": body, "truncated": truncated,
	}), nil
}

func textExt(ext string) bool {
	switch ext {
	case ".md", ".markdown", ".mdx", ".txt":
		return true
	}
	return false
}

// ---- 知识库开工注入 ----

type KbConventionsTool struct{}

func (t *KbConventionsTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("kb_get_conventions", "获取项目知识库已发布文档（规范/架构/决策/手册）与全局约定（conventions.* 记忆），开工前优先阅读",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
		}, nil,
	), nil
}

func (t *KbConventionsTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct{ ProjectId int }
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", p.ProjectId).
		Where("space", "knowledge").
		Where("status", "published").
		Order("updated_at DESC").Limit(5).All()
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	// 全局约定（project_id=0 的 conventions.* 记忆）：即使项目知识库为空也要带出，
	// 公司级规范不依赖单个项目的内容产出
	globalMemos := globalConventionMemos(ctx)
	if len(rows) == 0 {
		resp := g.Map{"items": []string{}}
		if len(globalMemos) > 0 {
			resp["globalConventions"] = globalMemos
			resp["hint"] = "知识库暂无已发布文档，已附全局约定"
		} else {
			resp["hint"] = "知识库暂无已发布文档"
		}
		return marshalString(resp), nil
	}
	const maxBody = 1500
	root := docs.RootPath(int64(p.ProjectId))
	items := make([]g.Map, 0, len(rows))
	for _, r := range rows {
		item := g.Map{
			"path": r["path"].String(), "title": r["title"].String(),
			"type": r["type"].String(), "tags": r["tags"].String(),
		}
		if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(r["path"].String()))); err == nil {
			_, body, _ := docs.ParseFrontmatter(string(data))
			if len(body) > maxBody {
				body = body[:maxBody] + "…（截断，read_doc 读全文）"
			}
			item["content"] = body
		}
		items = append(items, item)
	}
	resp := g.Map{"items": items}
	if len(globalMemos) > 0 {
		resp["globalConventions"] = globalMemos
	}
	return marshalString(resp), nil
}

// globalConventionMemos 全局记忆中 conventions. 前缀的约定（project_id=0，跨项目通用，
// 仅管理员经 Web 端维护；AI 侧只读注入）
func globalConventionMemos(ctx context.Context) []g.Map {
	rows, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", 0).
		Where("status IN (?)", g.Slice{"pending", "active"}).
		WhereLike("key", "conventions.%").
		Order("key ASC").Limit(10).All()
	if err != nil {
		return nil
	}
	memos := make([]g.Map, 0, len(rows))
	now := gtime.Now()
	sd := memoryStaleDays(ctx)
	for _, r := range rows {
		// 物化是每日定时任务，TTL 刚到期的行 status 列还停在 active——
		// 与 mem_list 同口径惰性判定，过期条目跳过不注入
		st, _ := memStatusHint(r["status"].String(), r["expires_at"].GTime(), r["last_verified_at"].GTime(), now, sd)
		if st == "expired" {
			continue
		}
		v := r["value"].String()
		if len(v) > 600 {
			v = v[:600] + "…"
		}
		memos = append(memos, g.Map{"key": r["key"].String(), "value": v})
	}
	return memos
}

// ---- 按任务反查关联文档 ----

type DocLinkedTool struct{}

func (t *DocLinkedTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("doc_linked", "反查与任务/需求/测试用例关联的文档（frontmatter linked），处理任务前先读关联文档",
		map[string]*ParameterInfo{
			"projectId":  paramInfo("integer", "项目ID"),
			"targetType": paramInfo("string", "关联类型：task/req/tc"),
			"targetId":   paramInfo("integer", "关联对象ID"),
		}, nil,
	), nil
}

func (t *DocLinkedTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct {
		ProjectId  int    `json:"projectId"`
		TargetType string `json:"targetType"`
		TargetId   int    `json:"targetId"`
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	switch p.TargetType {
	case "task", "req", "tc":
	default:
		return "", fmt.Errorf("targetType 必须是 task/req/tc")
	}
	target := fmt.Sprintf("%s:%d", p.TargetType, p.TargetId)
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", p.ProjectId).
		Where("','||linked||',' LIKE ?", "%,"+target+",%").
		Limit(5).All()
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	if len(rows) == 0 {
		return marshalString(g.Map{"target": target, "items": []string{}}), nil
	}
	const maxBody = 1500
	root := docs.RootPath(int64(p.ProjectId))
	items := make([]g.Map, 0, len(rows))
	for _, r := range rows {
		item := g.Map{"path": r["path"].String(), "title": r["title"].String()}
		if data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(r["path"].String()))); err == nil {
			_, body, _ := docs.ParseFrontmatter(string(data))
			if len(body) > maxBody {
				body = body[:maxBody] + "…（截断，read_doc 读全文）"
			}
			item["content"] = body
		}
		items = append(items, item)
	}
	return marshalString(g.Map{"target": target, "items": items}), nil
}

// ==================== 记忆工具 ====================

var (
	staleDaysMu      sync.Mutex
	staleDaysCache   int
	staleDaysCacheAt time.Time
)

// memoryStaleDays 腐化阈值（天）。sys_config 每请求都查一次太重，
// 进程内缓存 30s——阈值是运营配置，秒级生效无意义（并发竞态最坏多查一次，无害）
func memoryStaleDays(ctx context.Context) int {
	staleDaysMu.Lock()
	defer staleDaysMu.Unlock()
	if time.Since(staleDaysCacheAt) < 30*time.Second {
		return staleDaysCache
	}
	v, err := g.DB().Model("sys_config").Ctx(ctx).Where("key", "memory_stale_days").Value("value")
	n := 30
	if err == nil && v != nil {
		if parsed, perr := strconv.Atoi(v.String()); perr == nil && parsed > 0 {
			n = parsed
		}
	}
	staleDaysCache, staleDaysCacheAt = n, time.Now()
	return n
}

// memStatusHint 惰性判定（不物化，物化由定时任务负责）：
// 返回当前可用状态与人类可读提示
func memStatusHint(status string, expiresAt, lastVerified *gtime.Time, now *gtime.Time, staleDays int) (string, string) {
	if status == "expired" {
		return "expired", "已过期（确定失效，勿依赖）"
	}
	if expiresAt != nil && !expiresAt.IsZero() && now.After(expiresAt) {
		return "expired", "TTL 已到期（勿依赖，可 mem_set 覆盖更新）"
	}
	if status == "stale" {
		return "stale", "已腐化（超阈值未验证，谨慎依赖）"
	}
	if lastVerified != nil && !lastVerified.IsZero() {
		if days := int(now.Sub(lastVerified).Hours() / 24); days >= staleDays {
			return "stale", fmt.Sprintf("已 %d 天未验证（阈值 %d 天），疑似过时，确认后请 mem_verify", days, staleDays)
		}
	}
	if status == "pending" {
		return "pending", "待验证（推测值，未经确认）"
	}
	return status, ""
}

// ---- 记忆列表 ----

type MemListTool struct{}

func (t *MemListTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("mem_list", "列出全局与项目 KV 记忆（点分层级 key，如 conventions. / build.；scope=global 为跨项目通用约定，优先遵守），开工前先扫描记忆",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
		},
		map[string]*ParameterInfo{
			"prefix": paramInfo("string", "key 前缀过滤，如 conventions."),
		},
	), nil
}

func (t *MemListTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct {
		ProjectId int    `json:"projectId"`
		Prefix    string `json:"prefix,omitempty"`
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	// project_id IN (0, pid)：全局记忆（0）与项目记忆合并返回，全局在前
	m := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id IN (?)", g.Slice{0, p.ProjectId}).
		Where("status IN (?)", g.Slice{"pending", "active"})
	if p.Prefix != "" {
		m = m.WhereLike("key", p.Prefix+"%")
	}
	rows, err := m.Order("project_id ASC, key ASC").Limit(50).All()
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	if len(rows) == 0 {
		return marshalString(g.Map{"list": []string{}, "hint": "暂无可用记忆"}), nil
	}
	now := gtime.Now()
	sd := memoryStaleDays(ctx)
	type item struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Status   string `json:"status"`
		Scope    string `json:"scope" dc:"global=全局约定（跨项目通用），project=项目记忆"`
		Hint     string `json:"hint,omitempty"`
		Verified string `json:"lastVerifiedAt"`
	}
	list := make([]item, 0, len(rows))
	for _, r := range rows {
		st, hint := memStatusHint(r["status"].String(), r["expires_at"].GTime(), r["last_verified_at"].GTime(), now, sd)
		v := r["value"].String()
		if len(v) > 600 {
			v = v[:600] + "…"
		}
		scope := "project"
		if r["project_id"].Int64() == 0 {
			scope = "global"
		}
		list = append(list, item{Key: r["key"].String(), Value: v, Status: st, Scope: scope, Hint: hint, Verified: r["last_verified_at"].String()})
	}
	return marshalString(g.Map{"list": list}), nil
}

// ---- 记忆读取 ----

type MemGetTool struct{}

func (t *MemGetTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("mem_get", "读取单条记忆（项目作用域优先，未命中回落全局约定；含可用性提示：过期/腐化/待验证）",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
			"key":       paramInfo("string", "记忆 key，点分层级如 build.cmd"),
		}, nil,
	), nil
}

func (t *MemGetTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct {
		ProjectId int    `json:"projectId"`
		Key       string `json:"key"`
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	r, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", p.ProjectId).Where("key", p.Key).One()
	if err != nil {
		return "", fmt.Errorf("查询失败")
	}
	scope := "project"
	if r.IsEmpty() {
		// 项目作用域未命中回落全局（project_id=0）：同名 key 项目记忆可覆盖全局约定
		r, err = g.DB().Model("project_memories").Ctx(ctx).
			Where("project_id", 0).Where("key", p.Key).One()
		if err != nil {
			return "", fmt.Errorf("查询失败")
		}
		if r.IsEmpty() {
			return "", fmt.Errorf("记忆不存在: %s", p.Key)
		}
		scope = "global"
	}
	st, hint := memStatusHint(r["status"].String(), r["expires_at"].GTime(), r["last_verified_at"].GTime(), gtime.Now(), memoryStaleDays(ctx))
	return marshalString(g.Map{
		"key": p.Key, "value": r["value"].String(), "status": st, "hint": hint, "scope": scope,
		"lastVerifiedAt": r["last_verified_at"].String(),
	}), nil
}

// ---- 记忆写入 ----

type MemSetTool struct{}

func (t *MemSetTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("mem_set", "写入/更新项目记忆（学到的约定、偏好、结论等；upsert）",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
			"key":       paramInfo("string", "记忆 key，点分层级如 conventions.naming"),
			"value":     paramInfo("string", "记忆内容（纯文本，上限 64KB）"),
		},
		map[string]*ParameterInfo{
			"ttl":    paramInfo("string", "有效期 30m/12h/7d，缺省永不过期"),
			"status": paramInfo("string", "pending（推测未确认）或 active（默认，已验证）"),
		},
	), nil
}

func (t *MemSetTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct {
		ProjectId int    `json:"projectId"`
		Key       string `json:"key"`
		Value     string `json:"value"`
		Ttl       string `json:"ttl,omitempty"`
		Status    string `json:"status,omitempty"`
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	if strings.ContainsAny(p.Key, "/ ") || p.Key == "" {
		return "", fmt.Errorf("key 非法（点分层级，不含空格和斜杠）")
	}
	if len(p.Value) > 64*1024 {
		return "", fmt.Errorf("value 超过 64KB 上限")
	}
	if p.Status == "" {
		p.Status = "active"
	}
	if p.Status != "active" && p.Status != "pending" {
		return "", fmt.Errorf("status 仅支持 active/pending")
	}
	ttl, err := parseTtlSimple(p.Ttl)
	if err != nil {
		return "", err
	}
	uid := aiUserId(ctx)
	now := gtime.Now()
	// pending 语义是未确认推测：写 NULL 验证时间，与 docs 包 MemSet 同口径
	var lastVerified interface{}
	if p.Status == "active" {
		lastVerified = now
	}
	// 单语句原子 upsert（与 docs 包 MemSet 同口径）：count-then-insert 并发撞 UNIQUE
	var expiresAt interface{}
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}
	if _, err := g.DB().Exec(ctx,
		`INSERT INTO project_memories
			(project_id, key, value, status, expires_at, last_verified_at, verified_by, updated_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(project_id, key) DO UPDATE SET
			value = excluded.value, status = excluded.status, expires_at = excluded.expires_at,
			last_verified_at = excluded.last_verified_at, verified_by = excluded.verified_by,
			updated_by = excluded.updated_by, updated_at = excluded.updated_at`,
		p.ProjectId, p.Key, p.Value, p.Status, expiresAt, lastVerified, uid, uid, now, now); err != nil {
		return "", fmt.Errorf("写入失败")
	}
	return marshalString(g.Map{"key": p.Key, "status": "saved"}), nil
}

// ---- 记忆验证保鲜 ----

type MemVerifyTool struct{}

func (t *MemVerifyTool) Info(ctx context.Context) (*ToolInfo, error) {
	return toolInfo("mem_verify", "确认记忆仍正确时刷新验证时间（保鲜，防止腐化）",
		map[string]*ParameterInfo{
			"projectId": paramInfo("integer", "项目ID"),
			"key":       paramInfo("string", "记忆 key"),
		}, nil,
	), nil
}

func (t *MemVerifyTool) InvokableRun(ctx context.Context, argsInJSON string, opts ...ToolOption) (string, error) {
	var p struct {
		ProjectId int    `json:"projectId"`
		Key       string `json:"key"`
	}
	if err := json.Unmarshal([]byte(argsInJSON), &p); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}
	if pid, err := bindProjectId(ctx, p.ProjectId); err != nil {
		return "", err
	} else {
		p.ProjectId = pid
	}
	// 与 docs 包 memVerifyDo 同口径：置 active + 刷新验证时间后，
	// 同步清除已过期的 TTL——否则次日物化会把 status 翻回 expired，验证被静默撤销
	now := gtime.Now()
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", p.ProjectId).Where("key", p.Key).
		Data(g.Map{"status": "active", "last_verified_at": now,
			"verified_by": aiUserId(ctx), "updated_at": now}).Update()
	if err != nil {
		return "", fmt.Errorf("更新失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", fmt.Errorf("记忆不存在: %s", p.Key)
	}
	if _, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", p.ProjectId).Where("key", p.Key).
		Where("expires_at IS NOT NULL").Where("expires_at <= ?", now).
		Data("expires_at", nil).Update(); err != nil {
		return "", fmt.Errorf("更新失败")
	}
	return marshalString(g.Map{"key": p.Key, "status": "verified"}), nil
}

// ==================== 内部工具 ====================

func marshalString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// parseTtlSimple 30m/12h/7d；空串永不过期
func parseTtlSimple(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, nil
	}
	if len(s) < 2 {
		return 0, fmt.Errorf("ttl 格式应为 30m/12h/7d")
	}
	unit := s[len(s)-1]
	num, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("ttl 格式应为 30m/12h/7d")
	}
	switch unit {
	case 'm':
		return time.Duration(num) * time.Minute, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("ttl 单位仅支持 m/h/d")
}
