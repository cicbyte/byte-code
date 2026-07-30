package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/auth"
)

type IAuth interface {
	Login(ctx context.Context, req *api.LoginReq) (res *api.LoginRes, err error)
	AdminInfo(ctx context.Context) (res *api.AdminInfoRes, err error)
	Logout(ctx context.Context) (err error)
	ValidateToken(ctx context.Context, tokenStr string) (userId int, err error)
}

var localAuth IAuth

func Auth() IAuth {
	if localAuth == nil {
		panic("implement not found for interface IAuth, forgot register?")
	}
	return localAuth
}

func RegisterAuth(i IAuth) {
	localAuth = i
}
