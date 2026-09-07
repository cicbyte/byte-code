package menu

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MenusReq 获取当前用户动态菜单（前端路由生成用）
type MenusReq struct {
	g.Meta `path:"/menus" method:"get" tags:"菜单" summary:"获取用户菜单"`
}

type MenusRes struct {
	List []MenuItem `json:"list"`
}

type MenuItem struct {
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Component string     `json:"component"`
	Redirect  string     `json:"redirect,omitempty"`
	Meta      MenuMeta   `json:"meta"`
	Children  []MenuItem `json:"children,omitempty"`
}

type MenuMeta struct {
	Icon        string `json:"icon,omitempty"`
	Title       string `json:"title"`
	Hidden      bool   `json:"hidden,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// MenuListReq 获取全部菜单管理列表
type MenuListReq struct {
	g.Meta `path:"/menu/list" method:"get" tags:"菜单" summary:"获取菜单列表"`
}

type MenuListRes struct {
	List []MenuListItem `json:"list"`
}

type MenuListItem struct {
	Id       int            `json:"id"`
	Label    string         `json:"label"`
	Key      string         `json:"key"`
	Type     int            `json:"type"`
	Subtitle string         `json:"subtitle"`
	OpenType int            `json:"openType"`
	Auth     string         `json:"auth"`
	Path     string         `json:"path"`
	Children []MenuListItem `json:"children,omitempty"`
}
