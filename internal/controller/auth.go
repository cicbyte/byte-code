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
	token, err := service.Auth().Login(ctx, req)
	if err != nil {
		return nil, err
	}
	return &api.LoginRes{Token: token}, nil
}

func (c *authController) AdminInfo(ctx context.Context, req *api.AdminInfoReq) (res *api.AdminInfoRes, err error) {
	return service.Auth().AdminInfo(ctx)
}

func (c *authController) Logout(ctx context.Context, req *api.LogoutReq) (res *api.LogoutRes, err error) {
	err = service.Auth().Logout(ctx)
	res = new(api.LogoutRes)
	return
}
