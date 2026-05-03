package dashboard

import (
	"github.com/gogf/gf/v2/frame/g"
)

type ConsoleReq struct {
	g.Meta `path:"/dashboard/console" method:"get" tags:"Dashboard" summary:"仪表盘数据"`
}

type ConsoleRes struct {
	UserCount  int64 `json:"userCount"`
	RoleCount  int64 `json:"roleCount"`
	MenuCount  int64 `json:"menuCount"`
	OnlineUser int64 `json:"onlineUser"`
}
