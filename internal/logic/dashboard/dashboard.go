package dashboard

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/dashboard"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterDashboard(New())
}

func New() *sDashboard {
	return &sDashboard{}
}

type sDashboard struct{}

func (s *sDashboard) Console(ctx context.Context) (res *api.ConsoleRes, err error) {
	userCount, _ := g.DB().Model("sys_users").Count()
	roleCount, _ := g.DB().Model("sys_roles").Count()
	menuCount, _ := g.DB().Model("sys_menus").Count()
	onlineUser, _ := g.DB().Model("sys_tokens").Where("expired_at > now()").Count()

	return &api.ConsoleRes{
		UserCount:  int64(userCount),
		RoleCount:  int64(roleCount),
		MenuCount:  int64(menuCount),
		OnlineUser: int64(onlineUser),
	}, nil
}
