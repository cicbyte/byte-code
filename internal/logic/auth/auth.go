package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/auth"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/cicbyte/byte-code/utility/auditwriter"
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
	loginMaxFailures   = 5                // 同一用户名连续失败阈值
	loginIPMaxFailures = 30               // 同一来源 IP 连续失败阈值（防密码喷洒：每用户名只试一两次永不触用户名锁）
	loginLockDuration  = 15 * time.Minute // 达到阈值后的锁定时长
)

type loginFailRecord struct {
	count       int
	lastFail    time.Time
	lockedUntil time.Time
}

// 进程内失败计数：用户名维度（无论账号是否存在，防止借此枚举探测）+
// IP 维度（对齐 bc_ key 侧 aiKeyGuard 的双维度口径，阻断跨用户名喷洒）。
// GetClientIp 信任 XFF，伪造可绕 IP 锁但触不破用户名锁，且惰性清扫限制内存膨胀
var loginGuard = struct {
	sync.Mutex
	failures   map[string]*loginFailRecord
	ipFailures map[string]*loginFailRecord
}{failures: make(map[string]*loginFailRecord), ipFailures: make(map[string]*loginFailRecord)}

// loginSweepStale 惰性清扫：超过容量上限时删除 1 小时无新失败的记录
func loginSweepStale(m map[string]*loginFailRecord) {
	if len(m) < 4096 {
		return
	}
	cutoff := time.Now().Add(-time.Hour)
	for k, rec := range m {
		if rec.lastFail.Before(cutoff) {
			delete(m, k)
		}
	}
}

// loginDummyHash 用户不存在时也执行一次同代价的 bcrypt 比较，消除通过响应时间差枚举用户名
var loginDummyHash, _ = bcrypt.GenerateFromPassword([]byte("byte-code-dummy-password"), 10)

func loginLockCheck(username, clientIP string) error {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	now := time.Now()
	for _, rec := range []*loginFailRecord{loginGuard.failures[username], loginGuard.ipFailures[clientIP]} {
		if rec != nil && !rec.lockedUntil.IsZero() && now.Before(rec.lockedUntil) {
			minutes := int(time.Until(rec.lockedUntil).Minutes()) + 1
			return fmt.Errorf("失败次数过多，请约%d分钟后再试", minutes)
		}
	}
	return nil
}

func loginRecordFailure(username, clientIP string) {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	bump := func(m map[string]*loginFailRecord, key string, max int) {
		rec := m[key]
		if rec == nil {
			rec = &loginFailRecord{}
			m[key] = rec
		}
		rec.count++
		rec.lastFail = time.Now()
		if rec.count >= max {
			rec.lockedUntil = time.Now().Add(loginLockDuration)
			rec.count = 0
		}
	}
	bump(loginGuard.failures, username, loginMaxFailures)
	if clientIP != "" {
		bump(loginGuard.ipFailures, clientIP, loginIPMaxFailures)
	}
	loginSweepStale(loginGuard.failures)
	loginSweepStale(loginGuard.ipFailures)
}

func loginResetFailures(username string) {
	loginGuard.Lock()
	defer loginGuard.Unlock()
	delete(loginGuard.failures, username)
}

// loginAudit 登录事件审计（非阻塞尽力而为；公开组不走审计中间件，这里单独记）
func loginAudit(ctx context.Context, uid int, username, ip, action string) {
	ua := ""
	if r := g.RequestFromCtx(ctx); r != nil {
		ua = r.UserAgent()
	}
	auditwriter.Record(auditwriter.Entry{
		ActorID: uid, ActorType: "human", Action: action,
		TargetType: "auth", TargetName: username,
		IpAddress: ip, UserAgent: ua,
	})
}

func (s *sAuth) Login(ctx context.Context, req *api.LoginReq) (res *api.LoginRes, err error) {
	// 登录事件入审计（成功/失败均记，含来源 IP——防爆破事后可追溯）
	clientIP := ""
	if r := g.RequestFromCtx(ctx); r != nil {
		clientIP = r.GetClientIp()
	}
	// 连续失败锁定期内直接拒绝（用户名 + 来源 IP 双维度）
	if err = loginLockCheck(req.Username, clientIP); err != nil {
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
		loginRecordFailure(req.Username, clientIP)
		loginAudit(ctx, 0, req.Username, clientIP, "login_failed")
		return nil, fmt.Errorf("用户名或密码错误")
	}
	// 先验证密码再判断禁用状态，否则无需密码即可确认某用户名存在
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		loginRecordFailure(req.Username, clientIP)
		loginAudit(ctx, 0, req.Username, clientIP, "login_failed")
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
	loginAudit(ctx, user.Id, req.Username, clientIP, "login")
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
		// 权限字典（sys_menus.name）现行 10 项的显示名
		"system_menu":           "菜单管理",
		"system_role":           "角色管理",
		"ai_users":              "Agent 账号",
		"setting_system":        "系统设置",
		"setting_global_memory": "全局记忆",
		"platform_audit":        "审计日志",
		"platform_usage":        "使用分析",
		"platform_tags":         "标签管理",
		"platform_groups":       "项目分组",
		"project_create":        "创建项目",
	}
	if label, ok := labels[key]; ok {
		return label
	}
	return key
}

// ForgotPassword 发起密码重置。防枚举：账号不存在 / 无邮箱 / 未配 SMTP /
// 限频窗口内均静默返回成功（响应不可区分）。令牌 32 字节随机，库中只存
// SHA-256 摘要，30 分钟单次有效；同账号 60s 限频。
func (s *sAuth) ForgotPassword(ctx context.Context, account, origin string) (err error) {
	row, qerr := g.DB().Model("sys_users").Ctx(ctx).
		Where("(username = ? OR email = ?) AND type = 'human' AND status = 1", account, account).
		Fields("id, username, email").One()
	if qerr != nil || row.IsEmpty() || row["email"].String() == "" {
		return nil
	}
	uid := row["id"].Int()
	// 限频：同账号 60s 一封（静默——不改变响应形态）
	if last, _ := g.DB().Model("password_resets").Ctx(ctx).
		Where("user_id", uid).Order("id DESC").Fields("created_at").Value(); last != nil {
		if t, e := time.ParseInLocation("2006-01-02 15:04:05", last.String(), time.Local); e == nil && time.Since(t) < time.Minute {
			return nil
		}
	}
	raw := make([]byte, 32)
	if _, rerr := rand.Read(raw); rerr != nil {
		return nil
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	now := time.Now()
	if _, ierr := g.DB().Model("password_resets").Ctx(ctx).Insert(g.Map{
		"user_id":    uid,
		"token_hash": hex.EncodeToString(sum[:]),
		"expires_at": now.Add(30 * time.Minute).Format("2006-01-02 15:04:05"),
		"created_at": now.Format("2006-01-02 15:04:05"),
	}); ierr != nil {
		return nil
	}
	link := strings.TrimRight(origin, "/") + "/reset-password?token=" + token
	body := fmt.Sprintf("你（或他人）发起了 ByteCode 账号「%s」的密码重置。\n\n重置链接（30 分钟内有效，仅可使用一次）：\n%s\n\n如非本人操作，忽略本邮件即可，你的密码不会被更改。",
		row["username"].String(), link)
	// 发信失败也返回成功：防枚举要求存在/不存在账号响应不可区分
	// （失败只记服务端日志，管理员应保证 SMTP 已配置）
	if merr := service.Setting().SendMail(ctx, row["email"].String(), "ByteCode 密码重置", body); merr != nil {
		g.Log().Warningf(ctx, "forgot-password mail send failed (user=%d): %v", uid, merr)
	}
	return nil
}

// ResetPassword 凭令牌重置：校验 单次/未过期 → 改密 → 该用户全部重置令牌
// 置 used + 踢掉全部登录 token（改密即踢下线，与后台重置密码同口径）
func (s *sAuth) ResetPassword(ctx context.Context, token, newPassword string) (err error) {
	sum := sha256.Sum256([]byte(token))
	row, qerr := g.DB().Model("password_resets").Ctx(ctx).
		Where("token_hash", hex.EncodeToString(sum[:])).One()
	if qerr != nil || row.IsEmpty() {
		return fmt.Errorf("重置链接无效")
	}
	if row["used"].Int() == 1 {
		return fmt.Errorf("重置链接已使用，请重新发起")
	}
	exp, e := time.ParseInLocation("2006-01-02 15:04:05", row["expires_at"].String(), time.Local)
	if e != nil || time.Now().After(exp) {
		return fmt.Errorf("重置链接已过期，请重新发起")
	}
	uid := row["user_id"].Int()
	hash, herr := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if herr != nil {
		return fmt.Errorf("密码加密失败")
	}
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, uerr := tx.Exec("UPDATE sys_users SET password = ?, updated_at = ? WHERE id = ?",
			string(hash), time.Now().Format("2006-01-02 15:04:05"), uid); uerr != nil {
			return uerr
		}
		if _, uerr := tx.Exec("UPDATE password_resets SET used = 1 WHERE user_id = ?", uid); uerr != nil {
			return uerr
		}
		_, derr := tx.Exec("DELETE FROM sys_tokens WHERE user_id = ?", uid)
		return derr
	})
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
