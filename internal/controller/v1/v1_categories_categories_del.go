package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/categories"
)

func (c *ControllerCategories) CategoriesDel(ctx context.Context, req *categories.CategoriesDelReq) (res *categories.CategoriesDelRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
