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
type LogoutReq struct {
	g.Meta `path:"/login/logout" method:"post" tags:"认证" summary:"用户登出"`
}

type LogoutRes struct {
	g.Meta `mime:"application/json"`
}
