package vault

import (
	"context"
	"strconv"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/vault"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// staleDaysConfigKey 腐化阈值（天）：last_verified_at 距今超过该值视为 stale
const staleDaysConfigKey = "memory_stale_days"

func memoryStaleDays(ctx context.Context) int {
	v, err := g.DB().Model("sys_config").Ctx(ctx).Where("key", staleDaysConfigKey).Value("value")
	if err != nil || v == nil {
		return 30
	}
	if n, err := strconv.Atoi(v.String()); err == nil && n > 0 {
		return n
	}
	return 30
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
func materializeOne(ctx context.Context, id int64, status string, expiresAt, lastVerified *gtime.Time, now *gtime.Time, staleDays int) string {
	if status == "pending" || status == "active" {
		if expiresAt != nil && !expiresAt.IsZero() && now.After(expiresAt) {
			_, _ = g.DB().Model("project_memories").Ctx(ctx).Where("id", id).
				Data("status", "expired").Update()
			return "expired"
		}
		base := lastVerified
		if base == nil || base.IsZero() {
			base = now // 从未验证的记录以当下计，不立即腐化
		} else if now.Sub(base) > time.Duration(staleDays)*24*time.Hour {
			_, _ = g.DB().Model("project_memories").Ctx(ctx).Where("id", id).
				Data("status", "stale").Update()
			return "stale"
		}
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
	// 默认 pending+active；stale/expired 仅在显式 include 时返回（读取路径惰性物化）
	model = model.Where("project_memories.status IN (?)", g.Slice{"pending", "active", "stale", "expired"})
	rows, err := model.Fields("project_memories.*, sys_users.username").Order("key ASC").Limit(1000).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}

	now := gtime.Now()
	staleDays := memoryStaleDays(ctx)
	var list []api.MemoryItem
	for _, r := range rows {
		status := materializeOne(ctx, r["id"].Int64(), r["status"].String(),
			r["expires_at"].GTime(), r["last_verified_at"].GTime(), now, staleDays)
		switch status {
		case "stale":
			if !includeStale {
				continue
			}
		case "expired":
			if !includeExpired {
				continue
			}
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
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	if r.IsEmpty() {
		return nil, gerror.New("记忆不存在")
	}
	now := gtime.Now()
	staleDays := memoryStaleDays(ctx)
	status := materializeOne(ctx, r["id"].Int64(), r["status"].String(),
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
	data := g.Map{
		"value":             req.Value,
		"status":            status,
		"expires_at":        nil,
		"last_verified_at":  now,
		"verified_by":       uid,
		"updated_by":        uid,
		"updated_at":        now,
	}
	if ttl > 0 {
		data["expires_at"] = now.Add(ttl)
	}
	cnt, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).Count()
	if err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	if cnt > 0 {
		if _, err := g.DB().Model("project_memories").Ctx(ctx).
			Where("project_id", projectId).Where("key", key).Data(data).Update(); err != nil {
			return liberr.WrapDb(ctx, err, "索引同步失败")
		}
		return nil
	}
	data["project_id"] = projectId
	data["key"] = key
	data["created_at"] = now
	if _, err := g.DB().Model("project_memories").Ctx(ctx).Data(data).Insert(); err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return nil
}

func (s *sVault) MemVerify(ctx context.Context, projectId int64, key string, userId int64) error {
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).
		Data(g.Map{
			"status":           "active",
			"last_verified_at": gtime.Now(),
			"verified_by":      userId,
			"updated_at":       gtime.Now(),
		}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return gerror.New("记忆不存在")
	}
	return nil
}

func (s *sVault) MemExpire(ctx context.Context, projectId int64, key string) error {
	res, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("project_id", projectId).Where("key", key).
		Data("status", "expired").Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
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
		return liberr.WrapDb(ctx, err, "索引同步失败")
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
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	// 超阈值未验证 → stale
	threshold := now.Add(-time.Duration(staleDays) * 24 * time.Hour)
	if _, err := g.DB().Model("project_memories").Ctx(ctx).
		Where("status IN (?)", g.Slice{"pending", "active"}).
		Where("last_verified_at IS NOT NULL").
		Where("last_verified_at < ?", threshold).
		Data("status", "stale").Update(); err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
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
