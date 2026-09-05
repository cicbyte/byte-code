package aiuser

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/aiuser"
	"github.com/cicbyte/byte-code/internal/logic/auth"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterAiUser(New())
}

func New() *sAiUser {
	return &sAiUser{}
}

type sAiUser struct{}

func (s *sAiUser) Create(ctx context.Context, req *api.AiUserCreateReq) (id int, apiKey string, err error) {
	apiKey = generateApiKey()
	salt := generateSalt()
	hashedKey := hashApiKey(apiKey, salt)

	result, err := g.DB().Model("sys_users").Ctx(ctx).Insert(g.Map{
		"username":       req.Username,
		"password":       "",
		"real_name":      req.RealName,
		"type":           "ai",
		"capabilities":   req.Capabilities,
		"api_key":        hashedKey,
		"api_key_salt":   salt,
		"owner_human_id": ctx.Value("userId").(int),
		"status":         1,
	})
	if err != nil {
		return 0, "", err
	}

	lastId, _ := result.LastInsertId()
	return int(lastId), apiKey, nil
}

func (s *sAiUser) Update(ctx context.Context, req *api.AiUserUpdateReq) (err error) {
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, e := tx.Ctx(ctx).Model("sys_users").
			Where("id", req.Id).
			Where("type", "ai").
			Data(g.Map{
				"real_name":    req.RealName,
				"capabilities": req.Capabilities,
				"status":       req.Status,
			}).Update(); e != nil {
			return e
		}
		// 禁用即时生效：清准入/会话并踢掉已签发 token——否则旧 JWT 在
		// 有效期内仍可通过校验（ValidateToken 只查 token 表不查账号状态）
		if req.Status == 0 {
			return PurgeAgentAccess(tx, req.Id)
		}
		return nil
	})
	return
}

// PurgeAgentAccess 撤销 agent 的全部访问通道：项目准入、工作会话、已签发 token
func PurgeAgentAccess(tx gdb.TX, agentId int) error {
	if _, err := tx.Exec("DELETE FROM agent_project_bindings WHERE agent_id = ?", agentId); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM agent_sessions WHERE agent_id = ?", agentId); err != nil {
		return err
	}
	_, err := tx.Exec("DELETE FROM sys_tokens WHERE user_id = ?", agentId)
	return err
}

func (s *sAiUser) Delete(ctx context.Context, id int) (err error) {
	// 删号同事务清准入/会话/token：binding 残留会让 IsAgentBound 持续为真
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := PurgeAgentAccess(tx, id); err != nil {
			return err
		}
		_, err := tx.Ctx(ctx).Model("sys_users").
			Where("id", id).
			Where("type", "ai").
			Delete()
		return err
	})
}

func (s *sAiUser) List(ctx context.Context, req *api.AiUserListReq) (res *api.AiUserListRes, err error) {
	res = &api.AiUserListRes{}
	m := g.DB().Model("sys_users").Ctx(ctx).Where("type", "ai")

	total, err := m.Count()
	if err != nil {
		return
	}
	res.Total = total

	var list []api.AiUserItem
	err = m.Page(req.Page, req.Size).Order("id DESC").Scan(&list)
	if err != nil {
		return
	}
	// 已接入项目名回填（一次聚合查询，避免 N+1）
	if len(list) > 0 {
		ids := make([]int, 0, len(list))
		for _, u := range list {
			ids = append(ids, u.Id)
		}
		rows, berr := g.DB().Model("agent_project_bindings b").Ctx(ctx).
			Fields("b.agent_id, p.name").
			LeftJoin("projects p", "p.id = b.project_id").
			WhereIn("b.agent_id", ids).All()
		if berr == nil {
			pm := map[int][]string{}
			for _, r := range rows {
				pm[r["agent_id"].Int()] = append(pm[r["agent_id"].Int()], r["name"].String())
			}
			for i := range list {
				list[i].Projects = pm[list[i].Id]
			}
		}
	}
	res.List = list
	return
}

func (s *sAiUser) ResetKey(ctx context.Context, id int) (apiKey string, err error) {
	apiKey = generateApiKey()
	salt := generateSalt()
	hashedKey := hashApiKey(apiKey, salt)

	_, err = g.DB().Model("sys_users").Ctx(ctx).
		Where("id", id).
		Where("type", "ai").
		Data(g.Map{
			"api_key":      hashedKey,
			"api_key_salt": salt,
		}).Update()
	return
}

// ==================== AI Key 登录防爆破 ====================

const (
	aiKeyMaxFailures  = 5
	aiKeyLockDuration = 15 * time.Minute
	// aiKeyFailTTL 失败记录保留时长：过期即失效，防止伪造 XFF 的随机 IP 无限堆积
	aiKeyFailTTL = 30 * time.Minute
	// aiKeyFailMaxEntries 失败表容量上限：超出时清理最旧记录，封死内存膨胀面
	aiKeyFailMaxEntries = 10000
)

var aiKeyGuard = struct {
	sync.Mutex
	failures map[string]*aiKeyFailRecord
}{failures: make(map[string]*aiKeyFailRecord)}

type aiKeyFailRecord struct {
	count       int
	lockedUntil time.Time
	lastFailAt  time.Time
}

// pruneExpiredLocked 清理过期与超量记录（调用方需持锁）。伪造 XFF 可产生随机 key，
// 无 TTL 与容量上限的 map 会无限膨胀（内存 DoS）
func pruneExpiredLocked(now time.Time) {
	for k, rec := range aiKeyGuard.failures {
		if rec.lockedUntil.Before(now) && now.Sub(rec.lastFailAt) > aiKeyFailTTL {
			delete(aiKeyGuard.failures, k)
		}
	}
	if len(aiKeyGuard.failures) <= aiKeyFailMaxEntries {
		return
	}
	// 仍超量：按最后失败时间淘汰最旧的一半
	type kv struct {
		k string
		t time.Time
	}
	all := make([]kv, 0, len(aiKeyGuard.failures))
	for k, rec := range aiKeyGuard.failures {
		all = append(all, kv{k, rec.lastFailAt})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].t.Before(all[j].t) })
	for _, e := range all[:len(all)-aiKeyFailMaxEntries/2] {
		delete(aiKeyGuard.failures, e.k)
	}
}

func aiKeyLockCheck(lockKey string) error {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	now := time.Now()
	pruneExpiredLocked(now)
	rec := aiKeyGuard.failures[lockKey]
	if rec != nil && !rec.lockedUntil.IsZero() && now.Before(rec.lockedUntil) {
		minutes := int(time.Until(rec.lockedUntil).Minutes()) + 1
		return fmt.Errorf("失败次数过多，请约%d分钟后再试", minutes)
	}
	return nil
}

func aiKeyRecordFailure(lockKey string) {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	now := time.Now()
	pruneExpiredLocked(now)
	rec := aiKeyGuard.failures[lockKey]
	if rec == nil {
		rec = &aiKeyFailRecord{}
		aiKeyGuard.failures[lockKey] = rec
	}
	rec.count++
	rec.lastFailAt = now
	if rec.count >= aiKeyMaxFailures {
		rec.lockedUntil = now.Add(aiKeyLockDuration)
		rec.count = 0
	}
}

func aiKeyResetFailures(lockKey string) {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	delete(aiKeyGuard.failures, lockKey)
}

func (s *sAiUser) LoginByApiKey(ctx context.Context, apiKey string) (token string, err error) {
	userId, username, err := s.verifyApiKeyHash(ctx, apiKey)
	if err != nil {
		return "", err
	}
	token, err = auth.GenerateToken(userId, username)
	if err != nil {
		return "", err
	}
	// 与账密登录一致：token 入库，否则 ValidateToken 校验不过（AI token 全部失效）
	_, err = g.DB().Model("sys_tokens").Ctx(ctx).Insert(g.Map{
		"user_id":    userId,
		"token":      token,
		"expired_at": time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return "", fmt.Errorf("保存token失败")
	}
	return token, nil
}

// VerifyApiKey API Key 直认证（不签发 JWT）：供 TokenAuth 中间件
// 的 bc_ 前缀分支使用，外部 agent 免登录流程直调 API
func (s *sAiUser) VerifyApiKey(ctx context.Context, apiKey string) (userId int64, err error) {
	id, _, err := s.verifyApiKeyHash(ctx, apiKey)
	return int64(id), err
}

// verifyApiKeyHash 遍历 AI 用户做哈希比对（key 带 per-user salt 的 HMAC，无法索引查询），
// 含按来源 IP 的失败锁定防爆破
func (s *sAiUser) verifyApiKeyHash(ctx context.Context, apiKey string) (userId int, username string, err error) {
	ip := "unknown"
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
	}
	// 锁定维度 = IP + key 指纹：gf 的 GetClientIp 无条件信任 XFF（可伪造轮换），
	// 仅按 IP 锁定时攻击者换个头即绕过；叠加 key 内容哈希后，对同一把 key 的
	// 连续爆破无论 IP 怎么变都会累积失败计数
	// 双维度独立计数：IP 桶挡单 IP 撒网式尝试；key 指纹桶挡轮换 XFF 定向爆破同一把
	// key——组合键（ip|keyFp）换 IP 即换桶起不到防绕过作用
	keyFp := fmt.Sprintf("%x", sha256.Sum256([]byte("bc-key-fp:"+apiKey)))[:16]
	if err = aiKeyLockCheck(ip); err != nil {
		return 0, "", err
	}
	if err = aiKeyLockCheck(keyFp); err != nil {
		return 0, "", err
	}

	var users []struct {
		Id         int
		Username   string
		ApiKey     string
		ApiKeySalt string
		Status     int
	}
	err = g.DB().Model("sys_users").Ctx(ctx).
		Where("type", "ai").
		Where("status", 1).
		Scan(&users)
	if err != nil {
		return 0, "", fmt.Errorf("查询AI用户失败")
	}

	for _, u := range users {
		hashedInput := hashApiKey(apiKey, u.ApiKeySalt)
		// 常数时间比较，避免逐字节比较的时序侧信道
		if hmac.Equal([]byte(hashedInput), []byte(u.ApiKey)) {
			aiKeyResetFailures(ip)
			aiKeyResetFailures(keyFp)
			return u.Id, u.Username, nil
		}
	}

	aiKeyRecordFailure(ip)
	aiKeyRecordFailure(keyFp)
	return 0, "", fmt.Errorf("invalid API Key")
}

func generateApiKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "bc_" + hex.EncodeToString(b)
}

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashApiKey(apiKey, salt string) string {
	mac := hmac.New(sha256.New, []byte(salt))
	mac.Write([]byte(apiKey))
	return hex.EncodeToString(mac.Sum(nil))
}
