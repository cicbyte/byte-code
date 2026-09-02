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
	// 首次改密（账号带 must_change_password 标记）免验旧密码，由 logic 层判断
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword" v:"required|length:8,20#请输入新密码|密码长度为8-20位"`
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


// ==================== AI 引擎管理（仅超管） ====================

type AiEngineConfigReq struct {
	g.Meta  `path:"/admin/ai-engine/config" method:"get" tags:"AI引擎" summary:"获取AI引擎配置"`
}

type AiEngineConfigRes struct {
	g.Meta   `mime:"application/json"`
	BaseURL  string `json:"baseUrl"`
	Model    string `json:"model"`
	HasApiKey bool  `json:"hasApiKey"`
	Running  bool   `json:"running"`
}

type AiEngineConfigUpdateReq struct {
	g.Meta  `path:"/admin/ai-engine/config" method:"put" tags:"AI引擎" summary:"更新AI引擎配置"`
	BaseURL string `json:"baseUrl" v:"required#模型服务地址不能为空"`
	ApiKey  string `json:"apiKey"`
	Model   string `json:"model" v:"required#模型名不能为空"`
}

type AiEngineConfigUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type AiEngineToggleReq struct {
	g.Meta  `path:"/admin/ai-engine/toggle" method:"post" tags:"AI引擎" summary:"启动/停止AI引擎"`
	Action  string `json:"action" v:"required|in:start,stop#操作不能为空|必须是start或stop"`
}

type AiEngineToggleRes struct {
	g.Meta  `mime:"application/json"`
	Running bool `json:"running"`
}
