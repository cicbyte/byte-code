package setting

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetProfileReq 获取个人信息
type GetProfileReq struct {
	g.Meta `path:"/account/profile" method:"get" tags:"个人设置" summary:"获取个人信息"`
}

type GetProfileRes struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Avatar   string `json:"avatar"`
}

// UpdateProfileReq 更新个人信息
type UpdateProfileReq struct {
	g.Meta   `path:"/account/profile" method:"put" tags:"个人设置" summary:"更新个人信息"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" v:"email#邮箱格式不正确"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type UpdateProfileRes struct {
	g.Meta `mime:"application/json"`
}

// ChangePasswordReq 修改密码
type ChangePasswordReq struct {
	g.Meta      `path:"/account/password" method:"put" tags:"个人设置" summary:"修改密码"`
	OldPassword string `json:"oldPassword" v:"required#请输入旧密码"`
	NewPassword string `json:"newPassword" v:"required|length:6,20#请输入新密码|密码长度为6-20位"`
}

type ChangePasswordRes struct {
	g.Meta `mime:"application/json"`
}

// GetSystemConfigReq 获取系统配置
type GetSystemConfigReq struct {
	g.Meta `path:"/system/config" method:"get" tags:"系统设置" summary:"获取系统配置"`
}

type GetSystemConfigRes struct {
	SiteName     string `json:"siteName"`
	SiteIcp      string `json:"siteIcp"`
	SitePhone    string `json:"sitePhone"`
	SiteAddress  string `json:"siteAddress"`
	LoginCaptcha int    `json:"loginCaptcha"`
	SiteOpen     bool   `json:"siteOpen"`
	SiteCloseText string `json:"siteCloseText"`
	SmtpHost     string `json:"smtpHost"`
	SmtpPort     string `json:"smtpPort"`
	SmtpUser     string `json:"smtpUser"`
	SmtpFrom     string `json:"smtpFrom"`
}

// UpdateSystemConfigReq 更新系统配置
type UpdateSystemConfigReq struct {
	g.Meta        `path:"/system/config" method:"put" tags:"系统设置" summary:"更新系统配置"`
	SiteName      string `json:"siteName"`
	SiteIcp       string `json:"siteIcp"`
	SitePhone     string `json:"sitePhone"`
	SiteAddress   string `json:"siteAddress"`
	LoginCaptcha  int    `json:"loginCaptcha"`
	SiteOpen      bool   `json:"siteOpen"`
	SiteCloseText string `json:"siteCloseText"`
	SmtpHost      string `json:"smtpHost"`
	SmtpPort      string `json:"smtpPort"`
	SmtpUser      string `json:"smtpUser"`
	SmtpPass      string `json:"smtpPass"`
	SmtpFrom      string `json:"smtpFrom"`
}

type UpdateSystemConfigRes struct {
	g.Meta `mime:"application/json"`
}
