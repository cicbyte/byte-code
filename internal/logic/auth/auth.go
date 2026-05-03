package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/auth"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	service.RegisterAuth(New())
}

func New() *sAuth {
	return &sAuth{}
}

type sAuth struct{}

func (s *sAuth) Login(ctx context.Context, req *api.LoginReq) (token string, err error) {
	var user struct {
		Id       int
		Username string
		Password string
		Status   int
	}
	err = g.DB().Model("sys_users").Where("username", req.Username).Scan(&user)
	if err != nil {
		return "", fmt.Errorf("查询用户失败")
	}
	if user.Id == 0 {
		return "", fmt.Errorf("用户名或密码错误")
	}
	if user.Status != 1 {
		return "", fmt.Errorf("用户已被禁用")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("用户名或密码错误")
	}
	token, err = GenerateToken(user.Id, user.Username)
	if err != nil {
		return "", fmt.Errorf("生成token失败")
	}
	_, err = g.DB().Model("sys_tokens").Insert(g.Map{
		"user_id":    user.Id,
		"token":      token,
		"expired_at": time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return "", fmt.Errorf("保存token失败")
	}
	return token, nil
}

func (s *sAuth) AdminInfo(ctx context.Context) (res *api.AdminInfoRes, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return nil, fmt.Errorf("未获取到用户信息")
	}
	var user struct {
		Id       int
		Username string
		RealName string
		Avatar   string
		Desc     string
	}
	err = g.DB().Model("sys_users").Where("id", userId).Scan(&user)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败")
	}
	if user.Id == 0 {
		return nil, fmt.Errorf("用户不存在")
	}

	permissions := s.getUserPermissions(ctx, user.Id)

	res = &api.AdminInfoRes{
		UserId:      strconv.Itoa(user.Id),
		Username:    user.Username,
		RealName:    user.RealName,
		Avatar:      user.Avatar,
		Desc:        user.Desc,
		Permissions: permissions,
	}
	return res, nil
}

func (s *sAuth) getUserPermissions(ctx context.Context, userId int) []api.PermissionItem {
	type nameRow struct {
		Name string
	}
	var rows []nameRow
	err := g.DB().Model("sys_menus m").
		InnerJoin("sys_role_menus rm", "m.id = rm.menu_id").
		InnerJoin("sys_user_roles ur", "rm.role_id = ur.role_id").
		Where("ur.user_id", userId).
		Where("m.status", 1).
		Fields("DISTINCT m.name").
		Scan(&rows)
	if err != nil || len(rows) == 0 {
		rows = []nameRow{
			{"dashboard_console"}, {"dashboard_monitor"}, {"dashboard_workplace"},
			{"basic_list"}, {"basic_list_delete"},
		}
	}
	permissions := make([]api.PermissionItem, 0, len(rows))
	for _, row := range rows {
		label := s.permissionLabel(row.Name)
		permissions = append(permissions, api.PermissionItem{
			Label: label,
			Value: row.Name,
		})
	}
	return permissions
}

func (s *sAuth) permissionLabel(key string) string {
	labels := map[string]string{
		"dashboard_console":   "仪表盘",
		"dashboard_monitor":   "监控页",
		"dashboard_workplace": "工作台",
		"basic_list":          "基础列表",
		"basic_list_delete":   "基础列表删除",
	}
	if label, ok := labels[key]; ok {
		return label
	}
	return key
}

func (s *sAuth) Logout(ctx context.Context) (err error) {
	token := ctx.Value("token")
	if token != nil {
		_, _ = g.DB().Model("sys_tokens").Where("token", token).Delete()
	}
	return nil
}

func (s *sAuth) ValidateToken(ctx context.Context, tokenStr string) (userId int, err error) {
	claims, err := ParseToken(tokenStr)
	if err != nil {
		return 0, fmt.Errorf("token无效")
	}
	count, err := g.DB().Model("sys_tokens").Where("token", tokenStr).Count()
	if err != nil || count == 0 {
		return 0, fmt.Errorf("token已失效")
	}
	return claims.UserId, nil
}
