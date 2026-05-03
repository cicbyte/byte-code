package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/role"
)

type IRole interface {
	List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error)
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
