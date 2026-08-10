package aiuser

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/aiuser"
	"github.com/cicbyte/byte-code/internal/logic/auth"
	service "github.com/cicbyte/byte-code/internal/service"
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
		"username":      req.Username,
		"password":      "",
		"real_name":     req.RealName,
		"type":          "ai",
		"capabilities":  req.Capabilities,
		"api_key":       hashedKey,
		"api_key_salt":  salt,
		"owner_human_id": ctx.Value("userId").(int),
		"status":        1,
	})
	if err != nil {
		return 0, "", err
	}

	lastId, _ := result.LastInsertId()
	return int(lastId), apiKey, nil
}

func (s *sAiUser) Update(ctx context.Context, req *api.AiUserUpdateReq) (err error) {
	_, err = g.DB().Model("sys_users").Ctx(ctx).
		Where("id", req.Id).
		Where("type", "ai").
		Data(g.Map{
			"real_name":     req.RealName,
			"capabilities":  req.Capabilities,
			"status":        req.Status,
		}).Update()
	return
}

func (s *sAiUser) Delete(ctx context.Context, id int) (err error) {
	_, err = g.DB().Model("sys_users").Ctx(ctx).
		Where("id", id).
		Where("type", "ai").
		Delete()
	return
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
)

var aiKeyGuard = struct {
	sync.Mutex
	failures map[string]*aiKeyFailRecord
}{failures: make(map[string]*aiKeyFailRecord)}

type aiKeyFailRecord struct {
	count       int
	lockedUntil time.Time
}

func aiKeyLockCheck(ip string) error {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	rec := aiKeyGuard.failures[ip]
	if rec != nil && !rec.lockedUntil.IsZero() && time.Now().Before(rec.lockedUntil) {
		minutes := int(time.Until(rec.lockedUntil).Minutes()) + 1
		return fmt.Errorf("失败次数过多，请约%d分钟后再试", minutes)
	}
	return nil
}

func aiKeyRecordFailure(ip string) {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	rec := aiKeyGuard.failures[ip]
	if rec == nil {
		rec = &aiKeyFailRecord{}
		aiKeyGuard.failures[ip] = rec
	}
	rec.count++
	if rec.count >= aiKeyMaxFailures {
		rec.lockedUntil = time.Now().Add(aiKeyLockDuration)
		rec.count = 0
	}
}

func aiKeyResetFailures(ip string) {
	aiKeyGuard.Lock()
	defer aiKeyGuard.Unlock()
	delete(aiKeyGuard.failures, ip)
}

func (s *sAiUser) LoginByApiKey(ctx context.Context, apiKey string) (token string, err error) {
	// 按来源 IP 限制失败尝试，防止在线爆破 API Key
	ip := "unknown"
	if r := g.RequestFromCtx(ctx); r != nil {
		ip = r.GetClientIp()
	}
	if err = aiKeyLockCheck(ip); err != nil {
		return "", err
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
		return "", fmt.Errorf("查询AI用户失败")
	}

	for _, u := range users {
		hashedInput := hashApiKey(apiKey, u.ApiKeySalt)
		// 常数时间比较，避免逐字节比较的时序侧信道
		if hmac.Equal([]byte(hashedInput), []byte(u.ApiKey)) {
			token, err = auth.GenerateToken(u.Id, u.Username)
			if err != nil {
				return "", err
			}
			// 与账密登录一致：token 入库，否则 ValidateToken 校验不过（AI token 全部失效）
			_, err = g.DB().Model("sys_tokens").Ctx(ctx).Insert(g.Map{
				"user_id":    u.Id,
				"token":      token,
				"expired_at": time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
			})
			if err != nil {
				return "", fmt.Errorf("保存token失败")
			}
			aiKeyResetFailures(ip)
			return token, nil
		}
	}

	aiKeyRecordFailure(ip)
	return "", fmt.Errorf("invalid API Key")
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
