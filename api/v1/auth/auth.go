package auth

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginReq 登录请求
type LoginReq struct {
	g.Meta   `path:"/login" method:"post" tags:"认证" summary:"用户登录"`
	Username string `json:"username" v:"required#用户名不能为空"`
	Password string `json:"password" v:"required#密码不能为空"`
}

type LoginRes struct {
	Token string `json:"token"`
	// MustChangePassword 为 true 时账号仍在使用初始默认密码，前端应引导跳转改密页
	MustChangePassword bool `json:"mustChangePassword"`
}

// AdminInfoReq 获取用户信息
type AdminInfoReq struct {
	g.Meta `path:"/admin_info" method:"get" tags:"认证" summary:"获取用户信息"`
}

type AdminInfoRes struct {
	UserId      string              `json:"userId"`
	Username    string              `json:"username"`
	RealName    string              `json:"realName"`
	Avatar      string              `json:"avatar"`
	Desc        string              `json:"desc"`
	Permissions []PermissionItem    `json:"permissions"`
}

type PermissionItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// LogoutReq 登出
type ForgotPasswordReq struct {
	g.Meta `path:"/auth/forgot-password" method:"post" tags:"认证" summary:"发起密码重置（防枚举：恒定成功）"`
	// 用户名或绑定邮箱
	Account string `json:"account" v:"required#请输入用户名或邮箱"`
}

type ForgotPasswordRes struct {
	g.Meta `mime:"application/json"`
}

type ResetPasswordReq struct {
	g.Meta      `path:"/auth/reset-password" method:"post" tags:"认证" summary:"凭重置令牌设置新密码（单次/30min）"`
	Token       string `json:"token" v:"required#重置令牌不能为空"`
	NewPassword string `json:"newPassword" v:"required|length:8,20#新密码不能为空|新密码长度8-20位"`
}

type ResetPasswordRes struct {
	g.Meta `mime:"application/json"`
}

type LogoutReq struct {
	g.Meta `path:"/login/logout" method:"post" tags:"认证" summary:"用户登出"`
}

type LogoutRes struct {
	g.Meta `mime:"application/json"`
}
