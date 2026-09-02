package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/aiuser"
)

type IAiUser interface {
	Create(ctx context.Context, req *api.AiUserCreateReq) (id int, apiKey string, err error)
	Update(ctx context.Context, req *api.AiUserUpdateReq) (err error)
	Delete(ctx context.Context, id int) (err error)
	List(ctx context.Context, req *api.AiUserListReq) (res *api.AiUserListRes, err error)
	ResetKey(ctx context.Context, id int) (apiKey string, err error)
	LoginByApiKey(ctx context.Context, apiKey string) (token string, err error)
	// VerifyApiKey API Key 直认证（不签发 JWT），供 TokenAuth 的 bc_ 前缀分支
	VerifyApiKey(ctx context.Context, apiKey string) (userId int64, err error)
}

var localAiUser IAiUser

func AiUser() IAiUser {
	if localAiUser == nil {
		panic("implement not found for interface IAiUser")
	}
	return localAiUser
}

func RegisterAiUser(i IAiUser) {
	localAiUser = i
}
