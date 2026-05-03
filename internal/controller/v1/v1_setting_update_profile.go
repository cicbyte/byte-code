package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/setting"
)

func (c *ControllerSetting) UpdateProfile(ctx context.Context, req *setting.UpdateProfileReq) (res *setting.UpdateProfileRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
