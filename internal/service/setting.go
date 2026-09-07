package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/setting"
)

type ISetting interface {
	GetProfile(ctx context.Context) (res *api.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *api.UpdateProfileReq) (err error)
	UpdateAvatar(ctx context.Context, req *api.UpdateAvatarReq) (res *api.UpdateAvatarRes, err error)
	ServeAvatar(ctx context.Context, userId int, name string) (err error)
	ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (err error)
	GetSystemConfig(ctx context.Context) (res *api.GetSystemConfigRes, err error)
	UpdateSystemConfig(ctx context.Context, req *api.UpdateSystemConfigReq) (err error)
	GetAiEngineConfig(ctx context.Context) (res *api.AiEngineConfigRes, err error)
	UpdateAiEngineConfig(ctx context.Context, req *api.AiEngineConfigUpdateReq) (err error)
	ToggleAiEngine(ctx context.Context, action string) (running bool, err error)
}

var localSetting ISetting

func Setting() ISetting {
	if localSetting == nil {
		panic("implement not found for interface ISetting, forgot register?")
	}
	return localSetting
}

func RegisterSetting(i ISetting) {
	localSetting = i
}
