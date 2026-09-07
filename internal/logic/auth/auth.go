package auth

import (
	"context"
	"fmt"
	"strconv"
	"sync"
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

// ==================== 登录防爆破 ====================

const (
	loginMaxFailures  = 5                // 同一用户名连续失败阈值
	loginLockDuration = 15 * time.Minute // 达到阈值后的锁定时长
)

type loginFailRecord struct {
	count       int
	lockedUntil time.Time
}

// 进程内失败计数：锁定按输入的用户名计（无论账号是否存在，防止借此枚举探测）
var loginGuard = struct {
	sync.Mutex
	failures map[string]*loginFailRecord
}{failures: make(map[string]*loginFailRecord)}

// loginDummyHash 用户不存在时也执行一次同代价的 bcrypt 比较，消除通过响应时间差枚举用户名
var loginDummyHash, _ = bcrypt.GenerateFromPassword([]byte("byte-code-dummy-password"), 10)

func loginLockCheck(username string) error {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	rec := loginGuard.failures[username]
	if rec != nil && !rec.lockedUntil.IsZero() && time.Now().Before(rec.lockedUntil) {
		minutes := int(time.Until(rec.lockedUntil).Minutes()) + 1
		return fmt.Errorf("失败次数过多，请约%d分钟后再试", minutes)
	}
	return nil
}

func loginRecordFailure(username string) {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	rec := loginGuard.failures[username]
	if rec == nil {
		rec = &loginFailRecord{}
		loginGuard.failures[username] = rec
	}
	rec.count++
	if rec.count >= loginMaxFailures {
		rec.lockedUntil = time.Now().Add(loginLockDuration)
		rec.count = 0
	}
}

func loginResetFailures(username string) {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	delete(loginGuard.failures, username)
}

func (s *sAuth) Login(ctx context.Context, req *api.LoginReq) (res *api.LoginRes, err error) {
	// 连续失败锁定期内直接拒绝
	if err = loginLockCheck(req.Username); err != nil {
		return nil, err
	}
	var user struct {
		Id                 int
		Username           string
		Password           string
		Status             int
		MustChangePassword int
	}
	err = g.DB().Model("sys_users").Where("username", req.Username).Scan(&user)
	// 注意：GoFrame 对 struct 目标查不到行会返回 sql.ErrNoRows，须与真实查询错误
	// 一律按"用户名或密码错误"处理，避免借此区分用户名是否存在
	if err != nil || user.Id == 0 {
		_ = bcrypt.CompareHashAndPassword(loginDummyHash, []byte(req.Password))
		loginRecordFailure(req.Username)
		return nil, fmt.Errorf("用户名或密码错误")
	}
	// 先验证密码再判断禁用状态，否则无需密码即可确认某用户名存在
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		loginRecordFailure(req.Username)
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if user.Status != 1 {
		return nil, fmt.Errorf("用户已被禁用")
	}
	token, err := GenerateToken(user.Id, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成token失败")
	}
	_, err = g.DB().Model("sys_tokens").Insert(g.Map{
		"user_id":    user.Id,
		"token":      token,
		"expired_at": time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return nil, fmt.Errorf("保存token失败")
	}
	loginResetFailures(req.Username)
	return &api.LoginRes{
		Token:              token,
		MustChangePassword: user.MustChangePassword == 1,
	}, nil
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
	// 兜底集只属于「完全没有角色绑定」的账号；绑定了角色但菜单为空 = 真实空权限
	// （角色权限页清空菜单应生效，而不是悄悄回落到基础集）
	if err != nil {
		rows = nil
	}
	if len(rows) == 0 {
		roleCnt, _ := g.DB().Model("sys_user_roles").Where("user_id", userId).Count()
		if roleCnt == 0 {
			rows = []nameRow{
				{"dashboard_console"}, {"dashboard_monitor"}, {"dashboard_workplace"},
				{"basic_list"}, {"basic_list_delete"},
			}
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
	// expired_at 由登录时以本地时间字符串写入，须用同时区参数比较
	// （SQLite 无 now() 且 datetime('now') 为 UTC，会有时区偏差）；
	// 除存在性外同时校验过期时间，使主动注销/改密踢出即时生效
	count, err := g.DB().Model("sys_tokens").
		Where("token", tokenStr).
		Where("expired_at > ?", time.Now().Format("2006-01-02 15:04:05")).
		Count()
	if err != nil || count == 0 {
		return 0, fmt.Errorf("token已失效")
	}
	// 账号级兜底：被删除/禁用的账号，残留 token 一律失效
	uv, uerr := g.DB().Model("sys_users").Where("id", claims.UserId).Fields("status").Value()
	if uerr != nil || uv == nil || uv.Int() != 1 {
		return 0, fmt.Errorf("账号不可用")
	}
	return claims.UserId, nil
}
