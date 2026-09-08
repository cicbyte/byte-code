package setting

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/setting"
	aiengine "github.com/cicbyte/byte-code/internal/logic/aiengine"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/dbutil"
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

// ==================== 头像 ====================

// 头像落盘目录（与 jwt.secret 同属 resource/data，属运行期数据不进 git）
const avatarDir = "resource/data/avatars"

// avatarExts 允许的图片扩展名 -> Content-Type（白名单同时防目录穿越：文件名由此拼出）
var avatarExts = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// avatarNameRe 合法头像文件名：{userId}_{unix秒}.{ext}——读取端据此校验，
// 拒绝任何路径片段，天然免疫 ../ 穿越
var avatarNameRe = regexp.MustCompile(`^[1-9]\d*_\d+\.(png|jpg|jpeg|gif|webp)$`)

const avatarMaxSize = 2 << 20 // 2MB

func (s *sSetting) UpdateAvatar(ctx context.Context, req *api.UpdateAvatarReq) (res *api.UpdateAvatarRes, err error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return nil, fmt.Errorf("未获取到用户信息")
	}
	uid, _ := strconv.Atoi(fmt.Sprintf("%v", userId))
	if uid <= 0 {
		return nil, fmt.Errorf("未获取到用户信息")
	}
	f := req.File
	if f == nil {
		return nil, fmt.Errorf("请选择头像图片")
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if _, ok := avatarExts[ext]; !ok {
		return nil, fmt.Errorf("仅支持 png/jpg/jpeg/gif/webp 格式")
	}
	if f.Size > avatarMaxSize {
		return nil, fmt.Errorf("头像不能超过 2MB")
	}

	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建头像目录失败")
	}
	name := fmt.Sprintf("%d_%d%s", uid, time.Now().UnixNano(), ext)
	dst := filepath.Join(avatarDir, name)
	// 不用 UploadFile.Save：gf 会把不存在的目标路径当目录建出来（实测生成
	// "1_x.png/me.png"），手动读流落盘路径完全可控
	src, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败")
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("读取上传文件失败")
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, fmt.Errorf("保存头像失败")
	}
	// 同用户旧头像文件清理（前缀精确到 "{uid}_"，不误删他人）
	if entries, err := os.ReadDir(avatarDir); err == nil {
		prefix := fmt.Sprintf("%d_", uid)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) && e.Name() != name {
				_ = os.Remove(filepath.Join(avatarDir, e.Name()))
			}
		}
	}
	url := fmt.Sprintf("/api/account/avatar/%d/%s", uid, name)
	if _, err := g.DB().Model("sys_users").Where("id", uid).Data("avatar", url).Update(); err != nil {
		return nil, fmt.Errorf("更新头像失败")
	}
	return &api.UpdateAvatarRes{Avatar: url}, nil
}

// ServeAvatar 头像文件直出。URL 含时间戳版本号（同 URL 内容不变），
// 可放心长缓存；换头像生成新 URL，浏览器自然破缓存
func (s *sSetting) ServeAvatar(ctx context.Context, userId int, name string) (err error) {
	if userId <= 0 || !avatarNameRe.MatchString(name) {
		return fmt.Errorf("头像不存在")
	}
	// 文件名中的 uid 段必须与路径 id 一致，防止读他人目录之外的拼造名
	if !strings.HasPrefix(name, fmt.Sprintf("%d_", userId)) {
		return fmt.Errorf("头像不存在")
	}
	ext := strings.ToLower(filepath.Ext(name))
	data, err := os.ReadFile(filepath.Join(avatarDir, name))
	if err != nil {
		return fmt.Errorf("头像不存在")
	}
	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Type", avatarExts[ext])
	r.Response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	r.Response.Write(data)
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
	// 首次改密（登录即证实身份）免验旧密码；常规改密必须验旧密码
	if record["must_change_password"].Int() != 1 {
		if req.OldPassword == "" {
			return fmt.Errorf("请输入旧密码")
		}
		if err = bcrypt.CompareHashAndPassword([]byte(record["password"].String()), []byte(req.OldPassword)); err != nil {
			return fmt.Errorf("旧密码不正确")
		}
	}
	// 密码复杂度：须同时包含字母与数字（长度由 API 校验为 8-20 位）
	hasLetter, hasDigit := false, false
	for _, c := range req.NewPassword {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("新密码必须同时包含字母和数字")
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
	// 改密成功后踢掉该用户全部已发 token（含当前会话），
	// 防止密码泄露后改密而攻击者的旧会话仍继续有效
	if _, err = g.DB().Model("sys_tokens").Where("user_id", userId).Delete(); err != nil {
		return fmt.Errorf("密码已更新，但注销旧登录态失败，请重新登录")
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
		if err := dbutil.UpsertConfig(ctx, k, v); err != nil {
			return fmt.Errorf("更新配置失败")
		}
	}
	return nil
}

// ==================== AI 引擎管理 ====================

func (s *sSetting) GetAiEngineConfig(ctx context.Context) (res *api.AiEngineConfigRes, err error) {
	var cfgs []struct{ Key, Value string }
	err = g.DB().Model("sys_config").Ctx(ctx).
		WhereIn("key", g.Slice{"ai_engine_base_url", "ai_engine_api_key", "ai_engine_model"}).Scan(&cfgs)
	if err != nil {
		return nil, fmt.Errorf("查询配置失败")
	}
	m := make(map[string]string)
	for _, c := range cfgs {
		m[c.Key] = c.Value
	}
	return &api.AiEngineConfigRes{
		BaseURL:   m["ai_engine_base_url"],
		Model:     m["ai_engine_model"],
		HasApiKey: m["ai_engine_api_key"] != "",
		Running:   aiengine.IsRunning(),
	}, nil
}

func (s *sSetting) UpdateAiEngineConfig(ctx context.Context, req *api.AiEngineConfigUpdateReq) (err error) {
	items := map[string]string{
		"ai_engine_base_url": req.BaseURL,
		"ai_engine_model":    req.Model,
	}
	// ApiKey 为空表示保持原值
	if req.ApiKey != "" {
		items["ai_engine_api_key"] = req.ApiKey
	}
	for key, value := range items {
		if err := dbutil.UpsertConfig(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *sSetting) ToggleAiEngine(ctx context.Context, action string) (running bool, err error) {
	if action == "start" {
		// 检查配置
		cfg, err := s.GetAiEngineConfig(ctx)
		if err != nil {
			return false, fmt.Errorf("获取配置失败")
		}
		if cfg.BaseURL == "" || cfg.Model == "" {
			return false, fmt.Errorf("请先配置模型服务地址和模型名")
		}
		if !cfg.HasApiKey {
			// OpenAI-compatible 兼容 Ollama 无 Key 场景，不强制
		}
		aiengine.StartEngine(ctx)
		return true, nil
	}
	aiengine.StopEngine(ctx)
	return false, nil
}
