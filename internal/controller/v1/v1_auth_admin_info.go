package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/auth"
)

func (c *ControllerAuth) AdminInfo(ctx context.Context, req *auth.AdminInfoReq) (res *auth.AdminInfoRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
