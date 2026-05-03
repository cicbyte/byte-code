package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/setting"
)

type ISetting interface {
	GetProfile(ctx context.Context) (res *api.GetProfileRes, err error)
	UpdateProfile(ctx context.Context, req *api.UpdateProfileReq) (err error)
	ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (err error)
	GetSystemConfig(ctx context.Context) (res *api.GetSystemConfigRes, err error)
	UpdateSystemConfig(ctx context.Context, req *api.UpdateSystemConfigReq) (err error)
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
