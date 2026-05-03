package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/dashboard"
)

func (c *ControllerDashboard) Console(ctx context.Context, req *dashboard.ConsoleReq) (res *dashboard.ConsoleRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
