package setting

import (
	"context"
	"fmt"
	"strconv"

	api "github.com/cicbyte/byte-code/api/v1/setting"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	service.RegisterSetting(New())
}

func New() *sSetting {
	return &sSetting{}
}

type sSetting struct{}

func (s *sSetting) GetProfile(ctx context.Context) (res *api.GetProfileRes, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return nil, fmt.Errorf("未获取到用户信息")
	}
	record, err := g.DB().Model("sys_users").Where("id", userId).One()
	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败")
	}
	if record == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return &api.GetProfileRes{
		Nickname: record["real_name"].String(),
		Email:    record["email"].String(),
		Phone:    record["phone"].String(),
		Address:  record["address"].String(),
		Avatar:   record["avatar"].String(),
	}, nil
}

func (s *sSetting) UpdateProfile(ctx context.Context, req *api.UpdateProfileReq) (err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return fmt.Errorf("未获取到用户信息")
	}
	_, err = g.DB().Model("sys_users").Where("id", userId).Data(g.Map{
		"real_name": req.Nickname,
		"email":     req.Email,
		"phone":     req.Phone,
		"address":   req.Address,
	}).Update()
	if err != nil {
		return fmt.Errorf("更新失败")
	}
	return nil
}

func (s *sSetting) ChangePassword(ctx context.Context, req *api.ChangePasswordReq) (err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return fmt.Errorf("未获取到用户信息")
	}
	record, err := g.DB().Model("sys_users").Where("id", userId).One()
	if err != nil || record == nil {
		return fmt.Errorf("用户不存在")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(record["password"].String()), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("旧密码不正确")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}
	_, err = g.DB().Model("sys_users").Where("id", userId).Data(g.Map{
		"password":             string(hash),
		"must_change_password": 0,
	}).Update()
	if err != nil {
		return fmt.Errorf("密码更新失败")
	}
	return nil
}

func (s *sSetting) GetSystemConfig(ctx context.Context) (res *api.GetSystemConfigRes, err error) {
	type cfgRow struct {
		Key   string
		Value string
	}
	var rows []cfgRow
	err = g.DB().Model("sys_config").Scan(&rows)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败")
	}
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.Key] = r.Value
	}
	loginCaptcha, _ := strconv.Atoi(m["login_captcha"])
	siteOpen := m["site_open"] != "0"
	return &api.GetSystemConfigRes{
		SiteName:      m["site_name"],
		SiteIcp:       m["site_icp"],
		SitePhone:     m["site_phone"],
		SiteAddress:   m["site_address"],
		LoginCaptcha:  loginCaptcha,
		SiteOpen:      siteOpen,
		SiteCloseText: m["site_close_text"],
		SmtpHost:      m["smtp_host"],
		SmtpPort:      m["smtp_port"],
		SmtpUser:      m["smtp_user"],
		SmtpFrom:      m["smtp_from"],
	}, nil
}

func (s *sSetting) UpdateSystemConfig(ctx context.Context, req *api.UpdateSystemConfigReq) (err error) {
	loginCaptcha := "0"
	if req.LoginCaptcha == 1 {
		loginCaptcha = "1"
	}
	siteOpen := "0"
	if req.SiteOpen {
		siteOpen = "1"
	}
	configs := map[string]string{
		"site_name":       req.SiteName,
		"site_icp":        req.SiteIcp,
		"site_phone":      req.SitePhone,
		"site_address":    req.SiteAddress,
		"login_captcha":   loginCaptcha,
		"site_open":       siteOpen,
		"site_close_text": req.SiteCloseText,
		"smtp_host":       req.SmtpHost,
		"smtp_port":       req.SmtpPort,
		"smtp_user":       req.SmtpUser,
		"smtp_from":       req.SmtpFrom,
	}
	if req.SmtpPass != "" {
		configs["smtp_pass"] = req.SmtpPass
	}
	for k, v := range configs {
		_, err = g.DB().Model("sys_config").Data(g.Map{"value": v}).Where("key", k).Update()
		if err != nil {
			return fmt.Errorf("更新配置失败")
		}
	}
	return nil
}
