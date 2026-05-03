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
	MenuKeys    []string `json:"menu_keys"`
	CreateDate  string   `json:"create_date"`
	Status      string   `json:"status"`
}
