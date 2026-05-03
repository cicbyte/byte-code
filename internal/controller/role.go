package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/role"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Role = roleController{}

type roleController struct {
	BaseController
}

func (c *roleController) List(ctx context.Context, req *api.ListReq) (res *api.ListRes, err error) {
	return service.Role().List(ctx, req)
}
