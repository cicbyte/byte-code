package role

import (
	commonApi "github.com/cicbyte/byte-code/api/v1/common"
	"github.com/gogf/gf/v2/frame/g"
)

// ListReq 角色列表
type ListReq struct {
	g.Meta   `path:"/role/list" method:"get" tags:"角色" summary:"角色列表"`
	commonApi.PageReq
	Name string `json:"name"`
}

type ListRes struct {
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
	PageCount int        `json:"pageCount"`
	List      []RoleItem `json:"list"`
}

type RoleItem struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Explain     string   `json:"explain"`
	IsDefault   bool     `json:"isDefault"`
	MenuIds    []int    `json:"menuIds"`
	MenuKeys   []string `json:"menuKeys"`
	CreateDate string   `json:"createDate"`
	Status      string   `json:"status"`
}


// ==================== 角色 CRUD（仅超管） ====================

type CreateReq struct {
	g.Meta  `path:"/role" method:"post" tags:"角色" summary:"创建角色"`
	Name    string   `json:"name" v:"required|length:2,30#角色名不能为空|角色名长度2-30位"`
	Explain string   `json:"explain"`
	MenuIds []int    `json:"menuIds"`
}

type CreateRes struct {
	Id int `json:"id"`
}

type UpdateReq struct {
	g.Meta  `path:"/role/{id}" method:"put" tags:"角色" summary:"更新角色"`
	Id      int      `json:"id" v:"required" in:"path"`
	Name    *string  `json:"name"`
	Explain *string  `json:"explain"`
	Status  *string  `json:"status" v:"in:normal,disabled#状态必须是normal/disabled"`
	MenuIds []int    `json:"menuIds"`
}

type UpdateRes struct {
	g.Meta `mime:"application/json"`
}

type DeleteReq struct {
	g.Meta `path:"/role/{id}" method:"delete" tags:"角色" summary:"删除角色"`
	Id     int `json:"id" v:"required" in:"path"`
}

type DeleteRes struct {
	g.Meta `mime:"application/json"`
}

type UpdateMenusReq struct {
	g.Meta  `path:"/role/{id}/menus" method:"put" tags:"角色" summary:"分配菜单权限"`
	Id      int   `json:"id" v:"required" in:"path"`
	MenuIds []int `json:"menuIds"`
}

type UpdateMenusRes struct {
	g.Meta `mime:"application/json"`
}