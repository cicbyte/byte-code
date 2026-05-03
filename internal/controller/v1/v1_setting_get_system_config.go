package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/setting"
)

func (c *ControllerSetting) GetSystemConfig(ctx context.Context, req *setting.GetSystemConfigReq) (res *setting.GetSystemConfigRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
