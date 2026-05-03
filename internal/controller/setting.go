package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/setting"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Setting = settingController{}

type settingController struct {
	BaseController
}

func (c *settingController) GetProfile(ctx context.Context, req *api.GetProfileReq) (res *api.GetProfileRes, err error) {
	return service.Setting().GetProfile(ctx)
}

func (c *settingController) UpdateProfile(ctx context.Context, req *api.UpdateProfileReq) (res *api.UpdateProfileRes, err error) {
	err = service.Setting().UpdateProfile(ctx, req)
	res = new(api.UpdateProfileRes)
	return
}

func (c *settingController) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (res *api.ChangePasswordRes, err error) {
	err = service.Setting().ChangePassword(ctx, req)
	res = new(api.ChangePasswordRes)
	return
}

func (c *settingController) GetSystemConfig(ctx context.Context, req *api.GetSystemConfigReq) (res *api.GetSystemConfigRes, err error) {
	return service.Setting().GetSystemConfig(ctx)
}

func (c *settingController) UpdateSystemConfig(ctx context.Context, req *api.UpdateSystemConfigReq) (res *api.UpdateSystemConfigRes, err error) {
	err = service.Setting().UpdateSystemConfig(ctx, req)
	res = new(api.UpdateSystemConfigRes)
	return
}
