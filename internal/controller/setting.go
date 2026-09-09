package controller

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"context"

	api "github.com/cicbyte/byte-code/api/v1/setting"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
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

func (c *settingController) UpdateAvatar(ctx context.Context, req *api.UpdateAvatarReq) (res *api.UpdateAvatarRes, err error) {
	return service.Setting().UpdateAvatar(ctx, req)
}

// Avatar 头像文件直出（失败直接写 404，不走统一 JSON 包装）
func (c *settingController) Avatar(ctx context.Context, req *api.AvatarReq) (res *api.AvatarRes, err error) {
	if serr := service.Setting().ServeAvatar(ctx, req.Id, req.Name); serr != nil {
		g.RequestFromCtx(ctx).Response.WriteStatus(404, serr.Error())
	}
	return
}

func (c *settingController) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (res *api.ChangePasswordRes, err error) {
	err = service.Setting().ChangePassword(ctx, req)
	res = new(api.ChangePasswordRes)
	return
}

func (c *settingController) SmtpTest(ctx context.Context, req *api.SmtpTestReq) (*api.SmtpTestRes, error) {
	if err := service.Setting().SendTestMail(ctx, req.To); err != nil {
		return nil, gerror.New(err.Error())
	}
	return &api.SmtpTestRes{}, nil
}

func (c *settingController) GetSystemConfig(ctx context.Context, req *api.GetSystemConfigReq) (res *api.GetSystemConfigRes, err error) {
	return service.Setting().GetSystemConfig(ctx)
}

func (c *settingController) UpdateSystemConfig(ctx context.Context, req *api.UpdateSystemConfigReq) (res *api.UpdateSystemConfigRes, err error) {
	err = service.Setting().UpdateSystemConfig(ctx, req)
	res = new(api.UpdateSystemConfigRes)
	return
}

func (c *settingController) GetAiEngineConfig(ctx context.Context, req *api.AiEngineConfigReq) (res *api.AiEngineConfigRes, err error) {
	return service.Setting().GetAiEngineConfig(ctx)
}

func (c *settingController) UpdateAiEngineConfig(ctx context.Context, req *api.AiEngineConfigUpdateReq) (res *api.AiEngineConfigUpdateRes, err error) {
	err = service.Setting().UpdateAiEngineConfig(ctx, req)
	res = new(api.AiEngineConfigUpdateRes)
	return
}

func (c *settingController) ToggleAiEngine(ctx context.Context, req *api.AiEngineToggleReq) (res *api.AiEngineToggleRes, err error) {
	running, err := service.Setting().ToggleAiEngine(ctx, req.Action)
	if err != nil {
		return nil, err
	}
	return &api.AiEngineToggleRes{Running: running}, nil
}
