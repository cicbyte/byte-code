package menu

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/menu"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterMenu(New())
}

func New() *sMenu {
	return &sMenu{}
}

type sMenu struct{}

type dbMenu struct {
	Id        int
	ParentId  int
	Name      string
	Path      string
	Component string
	Redirect  string
	Title     string
	Icon      string
	Sort      int
	Status    int
	Hidden    int
	Type      int
	Auth      string
}

func (s *sMenu) Menus(ctx context.Context) (res []api.MenuItem, err error) {
	var menus []dbMenu
	err = g.DB().Model("sys_menus").Where("status", 1).Order("sort desc, id asc").Scan(&menus)
	if err != nil {
		return nil, err
	}
	res = buildMenuTree(menus, 0)
	return res, nil
}

func (s *sMenu) MenuList(ctx context.Context) (res []api.MenuListItem, err error) {
	var menus []dbMenu
	err = g.DB().Model("sys_menus").Where("status", 1).Order("sort desc, id asc").Scan(&menus)
	if err != nil {
		return nil, err
	}
	res = buildMenuListTree(menus, 0)
	return res, nil
}

func buildMenuTree(menus []dbMenu, parentId int) []api.MenuItem {
	var items []api.MenuItem
	for _, m := range menus {
		if m.ParentId != parentId {
			continue
		}
		item := api.MenuItem{
			Path:      m.Path,
			Name:      m.Name,
			Component: m.Component,
			Redirect:  m.Redirect,
			Meta: api.MenuMeta{
				Title:  m.Title,
				Icon:   m.Icon,
				Hidden: m.Hidden == 1,
			},
		}
		children := buildMenuTree(menus, m.Id)
		if len(children) > 0 {
			item.Children = children
		}
		items = append(items, item)
	}
	return items
}

func buildMenuListTree(menus []dbMenu, parentId int) []api.MenuListItem {
	var items []api.MenuListItem
	for _, m := range menus {
		if m.ParentId != parentId {
			continue
		}
		item := api.MenuListItem{
			Id:       m.Id,
			Label:    m.Title,
			Key:      m.Name,
			Type:     m.Type,
			Subtitle: m.Name,
			OpenType: 1,
			Auth:     m.Auth,
			Path:     m.Path,
		}
		children := buildMenuListTree(menus, m.Id)
		if len(children) > 0 {
			item.Children = children
		}
		items = append(items, item)
	}
	return items
}
