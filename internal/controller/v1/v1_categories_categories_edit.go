package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/categories"
)

func (c *ControllerCategories) CategoriesEdit(ctx context.Context, req *categories.CategoriesEditReq) (res *categories.CategoriesEditRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
