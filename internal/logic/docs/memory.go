package docs

import (
	"context"
	"strconv"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/docs"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

const staleDaysConfigKey = "memory_stale_days"

var (
	staleDaysCache     int
	staleDaysCacheAt   time.Time
)

// memoryStaleDays 腐化阈值（天）。sys_config 每请求都查一次太重，
// 进程内缓存 30s——阈值是运营配置，秒级生效无意义（并发竞态最坏多查一次，无害）
func memoryStaleDays(ctx context.Context) int {
	if time.Since(staleDaysCacheAt) < 30*time.Second {
		return staleDaysCache
	}
	v, err := g.DB().Model("sys_config").Ctx(ctx).Where("key", staleDaysConfigKey).Value("value")
	n := 30
	if err == nil && v != nil {
		if parsed, perr := strconv.Atoi(v.String()); perr == nil && parsed > 0 {
			n = parsed
		}
	}
	staleDaysCache, staleDaysCacheAt = n, time.Now()
	return n
}

// parseTtl 解析 30m/12h/7d 形式的有效期；空串返回 0（永不过期）
func parseTtl(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, nil
	}
	unit := s[len(s)-1]
	num, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || num <= 0 {
		return 0, gerror.New("ttl 格式错误，应为 30m/12h/7d 形式")
	}
	switch unit {
	case 'm':
		return time.Duration(num) * time.Minute, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	default:
		return 0, gerror.New("ttl 单位仅支持 m/h/d")
	}
}

// materializeOne 惰性物化单条：TTL 到期→expired、超阈值未验证→stale
// evalMemoryStatus 惰性状态判定（只算不写）：TTL 到期→expired、超阈值未验证→stale。
// 物化（落库）统一由每日 MemMaterialize 定时任务负责——读路径逐行回写在 SQLite
// 单写者模型下是与 Web/AI 写请求抢锁的纯放大，且与定时任务双轨重复
func evalMemoryStatus(status string, expiresAt, lastVerified *gtime.Time, now *gtime.Time, staleDays int) string {
	if status == "expired" {
		return "expired"
	}
	if expiresAt != nil && !expiresAt.IsZero() && now.After(expiresAt) {
		return "expired"
	}
	if status == "stale" {
		return "stale"
	}
	if lastVerified != nil && !lastVerified.IsZero() {
		if now.Sub(lastVerified) > time.Duration(staleDays)*24*time.Hour {
			return "stale"
		}
	}
	if status == "pending" {
		return "pending"
	}
	return status
}

func (s *sVault) MemList(ctx context.Context, projectId int64, prefix, include string) ([]api.MemoryItem, error) {
	includeStale := strings.Contains(include, "stale")
	includeExpired := strings.Contains(include, "expired")

	model := g.DB().Model("project_memories").Ctx(ctx).
		LeftJoin("sys_users", "sys_users.id = project_memories.updated_by").
		Where("project_id", projectId)
	if prefix != "" {
		model = model.WhereLike("key", prefix+"%")
	}
	// 状态过滤在 SQL 层完成：expired/stale 行不在 include 时根本不取——
	// 否则长期运行的终态行会占满 Limit 配额，把活跃记忆挤出结果集。
	// 必须收进同一个 IN 列表：gf 的 WhereOr 不给已有 AND 组加括号，
	// 单独 OR 追加会让状态条件逃逸 project_id 过滤（P0 跨项目泄漏，
	// 见 status-2026-09-04.md）
	statuses := g.Slice{"pending", "active"}
	if includeStale {
		statuses = append(statuses, "stale")
	}
	if includeExpired {
		statuses = append(statuses, "expired")
	}
	model = model.Where("project_memories.status IN (?)", statuses)
	rows, err := model.Fields("project_memories.*, sys_users.username").Order("key ASC").Limit(1000).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询记忆列表失败")
	}

	now := gtime.Now()
	staleDays := memoryStaleDays(ctx)
	var list []api.MemoryItem
	for _, r := range rows {
		status := evalMemoryStatus(r["status"].String(),
			r["expires_at"].GTime(), r["last_verified_at"].GTime(), now, staleDays)
		// 惰性判定的终态（状态列未及物化）遵循 include 开关
		if status == "stale" && !includeStale {
			continue
		}
		if status == "expired" && !includeExpired {
			continue
		}
		list = append(list, memoryRowToItem(r, status, now, staleDays))
	}
	return list, nil
}

func (s *sVault) MemGet(ctx context.Context, projectId int64, key string) (*api.MemoryGetRes, error) {
	r, err := g.DB().Model("project_memories").Ctx(ctx).
		LeftJoin("sys_users", "sys_users.id = project_memories.updated_by").
		Where("project_id", projectId).Where("key", key).
		Fields("project_memories.*, sys_users.username").One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询记忆失败")
	}
	if r.IsEmpty() {
		return nil, gerror.New("记忆不存在")
	}
	now := gtime.Now()
	staleDays := memoryStaleDays(ctx)
	status := evalMemoryStatus(r["status"].String(),
		r["expires_at"].GTime(), r["last_verified_at"].GTime(), now, staleDays)
	return &api.MemoryGetRes{MemoryItem: memoryRowToItem(r, status, now, staleDays)}, nil
}

func (s *sVault) MemSet(ctx context.Context, projectId int64, key string, req *api.MemorySetReq) error {
	key = strings.TrimSpace(strings.Trim(key, "/"))
	if key == "" || strings.Contains(key, "/") {
		return gerror.New("key 不能为空且不含 /（点分层级：build.cmd / conventions.naming）")
	}
	if len(req.Value) > 64*1024 {
		return gerror.New("单值上限 64KB")
	}
	ttl, err := parseTtl(req.Ttl)
	if err != nil {
		return err
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	uid := int(perm.UserId(ctx))
	now := gtime.Now()
	var expiresAt interface{}
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}
	// pending 语义是"未确认推测"：写 NULL 验证时间，避免与 active 同天起算
	// 腐化阈值（否则 pending 永不比 active 更快变 stale，状态失去区分度）
	var lastVerified interface{}
	if status == "active" {
		lastVerified = now
	}
	// 单语句原子 upsert：count-then-insert 在并发写同 key 时会撞
	// UNIQUE(project_id, key)——报"索引同步失败"完全误导排障方向
	_, err = g.DB().Exec(ctx,
		`INSERT INTO project_memories
			(project_id, key, value, status, expires_at, last_verified_at, verified_by, updated_by, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(project_id, key) DO UPDATE SET
			value = excluded.value, status = excluded.status, expires_at = excluded.expires_at,
			last_verified_at = excluded.last_verified_at, verified_by = excluded.verified_by,
			updated_by = excluded.updated_by, updated_at = excluded.updated_at`,
		projectId, key, req.Value, status, expiresAt, lastVerified, uid, uid, now, now)
	if err != nil {
		return liberr.WrapDb(ctx, err, "写入记忆失败")
	}
	return nil
}

func (s *sVault) MemVerify(ctx context.Context, projectId int64, key string, userId int64) error {
	return s.memVerifyDo(ctx, projectId, key, userId, gtime.Now())
}

// memVerifyDo 验证保鲜：置回 active 并刷新验证时间；已过期的 TTL 一并清除
// （验证意味着内容确认有效，留着过去的 expires_at 会让读路径的惰性判定
// 立刻又判回 expired，落库状态与逻辑状态打架）
func (s *sVault) memVerifyDo(ctx context.Context, projectId int64, key string, userId int64, now *gtime.Time) error {
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).
		Data(g.Map{
			"status":           "active",
			"last_verified_at": now,
			"verified_by":      userId,
			"updated_at":       now,
		}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "验证记忆失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return gerror.New("记忆不存在")
	}
	_, err = g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).
		Where("expires_at IS NOT NULL").Where("expires_at <= ?", now).
		Data("expires_at", nil).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "验证记忆失败")
	}
	return nil
}

func (s *sVault) MemExpire(ctx context.Context, projectId int64, key string) error {
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).
		Data("status", "expired").Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "废弃记忆失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return gerror.New("记忆不存在")
	}
	return nil
}

func (s *sVault) MemDelete(ctx context.Context, projectId int64, key string) error {
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除记忆失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return gerror.New("记忆不存在")
	}
	return nil
}

// MemMaterialize 批量物化腐化状态（定时任务调用；读取路径另有惰性物化）
func (s *sVault) MemMaterialize(ctx context.Context) error {
	now := gtime.Now()
	staleDays := memoryStaleDays(ctx)

	// TTL 到期 → expired
	if _, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("status IN (?)", g.Slice{"pending", "active"}).
		Where("expires_at IS NOT NULL").
		Where("expires_at <= ?", now).
		Data("status", "expired").Update(); err != nil {
		return liberr.WrapDb(ctx, err, "记忆状态物化失败")
	}
	// 超阈值未验证 → stale
	threshold := now.Add(-time.Duration(staleDays) * 24 * time.Hour)
	if _, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("status IN (?)", g.Slice{"pending", "active"}).
		Where("last_verified_at IS NOT NULL").
		Where("last_verified_at < ?", threshold).
		Data("status", "stale").Update(); err != nil {
		return liberr.WrapDb(ctx, err, "记忆状态物化失败")
	}
	return nil
}

func memoryRowToItem(r gdb.Record, status string, now *gtime.Time, staleDays int) api.MemoryItem {
	lv := r["last_verified_at"].GTime()
	days := 0
	if lv != nil && !lv.IsZero() {
		days = int(now.Sub(lv).Hours() / 24)
	}
	return api.MemoryItem{
		Key:            r["key"].String(),
		Value:          r["value"].String(),
		Status:         status,
		StaleDays:      days,
		LastVerifiedAt: r["last_verified_at"].String(),
		ExpiresAt:      r["expires_at"].String(),
		Source:         r["username"].String(),
		UpdatedAt:      r["updated_at"].String(),
	}
}
