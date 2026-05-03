package v1

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/cicbyte/byte-code/api/v1/categories"
)

func (c *ControllerCategories) CategoriesList(ctx context.Context, req *categories.CategoriesListReq) (res *categories.CategoriesListRes, err error) {
	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
