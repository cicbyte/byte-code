package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/auth"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Auth = authController{}

type authController struct {
	BaseController
}

func (c *authController) Login(ctx context.Context, req *api.LoginReq) (res *api.LoginRes, err error) {
	return service.Auth().Login(ctx, req)
}

func (c *authController) AdminInfo(ctx context.Context, req *api.AdminInfoReq) (res *api.AdminInfoRes, err error) {
	return service.Auth().AdminInfo(ctx)
}

func (c *authController) Logout(ctx context.Context, req *api.LogoutReq) (res *api.LogoutRes, err error) {
	err = service.Auth().Logout(ctx)
	res = new(api.LogoutRes)
	return
}
