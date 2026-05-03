package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/menu"
)

func (c *ControllerMenu) Menus(ctx context.Context, req *menu.MenusReq) (res *menu.MenusRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
