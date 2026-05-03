package service

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/menu"
)

type IMenu interface {
	Menus(ctx context.Context) (res []api.MenuItem, err error)
	MenuList(ctx context.Context) (res []api.MenuListItem, err error)
}

var localMenu IMenu

func Menu() IMenu {
	if localMenu == nil {
		panic("implement not found for interface IMenu, forgot register?")
	}
	return localMenu
}

func RegisterMenu(i IMenu) {
	localMenu = i
}
