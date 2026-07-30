package dashboard

import (
	"context"
	"fmt"
	"time"

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
	userCount, err := g.DB().Model("sys_users").Count()
	if err != nil {
		return nil, fmt.Errorf("获取用户数失败")
	}
	roleCount, err := g.DB().Model("sys_roles").Count()
	if err != nil {
		return nil, fmt.Errorf("获取角色数失败")
	}
	menuCount, err := g.DB().Model("sys_menus").Count()
	if err != nil {
		return nil, fmt.Errorf("获取菜单数失败")
	}
	// expired_at 由登录逻辑以本地时间字符串写入，比较须用相同时区的参数：
	// SQLite 没有 now() 函数（原查询必然报错使在线人数恒为 0），
	// 而 datetime('now') 是 UTC，与本地时间写入的值比较会有时区偏差
	onlineUser, err := g.DB().Model("sys_tokens").
		Where("expired_at > ?", time.Now().Format("2006-01-02 15:04:05")).
		Count()
	if err != nil {
		return nil, fmt.Errorf("获取在线用户数失败")
	}

	return &api.ConsoleRes{
		UserCount:  int64(userCount),
		RoleCount:  int64(roleCount),
		MenuCount:  int64(menuCount),
		OnlineUser: int64(onlineUser),
	}, nil
}
