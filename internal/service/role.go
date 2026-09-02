package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/role"
)

type IRole interface {
	List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error)
	Create(ctx context.Context, req *api.CreateReq) (id int, err error)
	Update(ctx context.Context, req *api.UpdateReq) (err error)
	Delete(ctx context.Context, id int) (err error)
	UpdateMenus(ctx context.Context, id int, menuIds []int) (err error)
}

var localRole IRole

func Role() IRole {
	if localRole == nil {
		panic("implement not found for interface IRole, forgot register?")
	}
	return localRole
}

func RegisterRole(i IRole) {
	localRole = i
}
