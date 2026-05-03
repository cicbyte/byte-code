package aiuser

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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

func (s *sAiUser) LoginByApiKey(ctx context.Context, apiKey string) (token string, err error) {
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
		return "", err
	}

	for _, u := range users {
		hashedInput := hashApiKey(apiKey, u.ApiKeySalt)
		if hashedInput == u.ApiKey {
			token, err = auth.GenerateToken(u.Id, u.Username)
			if err != nil {
				return "", err
			}
			return token, nil
		}
	}

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
