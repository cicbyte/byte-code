package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/auth"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
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

func (c *authController) ForgotPassword(ctx context.Context, req *api.ForgotPasswordReq) (res *api.ForgotPasswordRes, err error) {
	r := g.RequestFromCtx(ctx)
	// 重置链接的前端基址：优先 Origin（跨域浏览器必有），退化到请求 Host
	origin := r.Header.Get("Origin")
	if origin == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		origin = scheme + "://" + r.Host
	}
	if err = service.Auth().ForgotPassword(ctx, req.Account, origin); err != nil {
		return nil, err
	}
	return &api.ForgotPasswordRes{}, nil
}

func (c *authController) ResetPassword(ctx context.Context, req *api.ResetPasswordReq) (res *api.ResetPasswordRes, err error) {
	if err = service.Auth().ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return nil, err
	}
	return &api.ResetPasswordRes{}, nil
}

func (c *authController) Logout(ctx context.Context, req *api.LogoutReq) (res *api.LogoutRes, err error) {
	err = service.Auth().Logout(ctx)
	res = new(api.LogoutRes)
	return
}
