package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/user"
)

type IUser interface {
	List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error)
	Create(ctx context.Context, req *api.CreateReq) (id int, err error)
	Update(ctx context.Context, req *api.UpdateReq) (err error)
	ResetPassword(ctx context.Context, req *api.ResetPasswordReq) (err error)
	Delete(ctx context.Context, id int) (err error)
}

var localUser IUser

func User() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser, forgot register?")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}
