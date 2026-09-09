package setting

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ==================== SMTP 发送 ====================
// 审计三代挂账项：配置页存在但全库零发信调用。此文件补上发送能力，
// 供测试端点与未来通知外发消费。

func loadSmtpConfig(ctx context.Context) (*smtpParams, error) {
	keys := []string{"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from"}
	vals := make(map[string]string)
	for _, k := range keys {
		v, _ := g.DB().Model("sys_config").Ctx(ctx).Where("`key`", k).Value("value")
		vals[k] = v.String()
	}
	if vals["smtp_host"] == "" || vals["smtp_port"] == "" {
		return nil, fmt.Errorf("SMTP 未配置（请在系统设置 → 邮件设置中填写服务器信息）")
	}
	return &smtpParams{
		Host: vals["smtp_host"],
		Port: vals["smtp_port"],
		User: vals["smtp_user"],
		Pass: vals["smtp_pass"],
		From: vals["smtp_from"],
	}, nil
}

type smtpParams struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// SendTestMail 发送测试邮件（验证 SMTP 配置可用性）
func (s *sSetting) SendTestMail(ctx context.Context, to string) error {
	if to == "" || !strings.Contains(to, "@") {
		return fmt.Errorf("请输入有效的收件邮箱")
	}
	cfg, err := loadSmtpConfig(ctx)
	if err != nil {
		return err
	}

	port, _ := strconv.Atoi(cfg.Port)
	addr := fmt.Sprintf("%s:%d", cfg.Host, port)

	from := cfg.From
	if from == "" {
		from = cfg.User // 发件人未配置时退化为用户名
	}

	subject := "ByteCode 测试邮件"
	body := fmt.Sprintf("这是一封来自 ByteCode 平台的测试邮件。\n\n如果你收到了这封邮件，说明 SMTP 配置正确。\n\n服务器: %s\n端口: %s\n时间: %s\n",
		cfg.Host, cfg.Port, time.Now().Format("2006-01-02 15:04:05"))

	msg := buildMime(from, to, subject, body)

	var auth smtp.Auth
	if cfg.User != "" && cfg.Pass != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	// 465 端口走隐式 SSL；587/25 走明文/STARTTLS
	if port == 465 {
		return sendSSL(addr, cfg.Host, auth, from, []string{to}, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// sendSSL 隐式 SSL 连接（465 端口，net/smtp 标准库不直接支持需手动拨号）
func sendSSL(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("SSL 连接失败: %v", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %v", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("认证失败（检查用户名/密码）: %v", err)
		}
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("设置收件人 %s 失败: %v", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("打开数据流失败: %v", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("关闭数据流失败: %v", err)
	}
	return client.Quit()
}

// buildMime 构造 UTF-8 纯文本 MIME 邮件
func buildMime(from, to, subject, body string) []byte {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", mimeEncode(subject)),
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset=UTF-8`,
		"Content-Transfer-Encoding: base64",
		"",
	}
	// base64 编码正文（76 字符换行，MIME 规范）
	enc := base64.StdEncoding.EncodeToString([]byte(body))
	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(h)
		sb.WriteString("\r\n")
	}
	for i := 0; i < len(enc); i += 76 {
		end := i + 76
		if end > len(enc) {
			end = len(enc)
		}
		sb.WriteString(enc[i:end])
		sb.WriteString("\r\n")
	}
	return []byte(sb.String())
}

// mimeEncode 对非 ASCII 主题做 RFC 2047 B 编码
func mimeEncode(s string) string {
	for _, r := range s {
		if r > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
		}
	}
	return s
}
