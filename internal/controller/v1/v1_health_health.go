package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/health"
)

func (c *ControllerHealth) Health(ctx context.Context, req *health.HealthReq) (res *health.HealthRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
