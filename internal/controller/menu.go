package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/menu"
	service "github.com/cicbyte/byte-code/internal/service"
)

var Menu = menuController{}

type menuController struct {
	BaseController
}

func (c *menuController) Menus(ctx context.Context, req *api.MenusReq) (res *api.MenusRes, err error) {
	list, err := service.Menu().Menus(ctx)
	if err != nil {
		return
	}
	res = &api.MenusRes{List: list}
	return
}

func (c *menuController) MenuList(ctx context.Context, req *api.MenuListReq) (res *api.MenuListRes, err error) {
	list, err := service.Menu().MenuList(ctx)
	if err != nil {
		return
	}
	res = &api.MenuListRes{List: list}
	return
}
