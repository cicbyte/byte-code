// Package vault 项目文档中枢的磁盘真相源层。
// 目录=磁盘目录、文件=磁盘文件、元数据=frontmatter；SQLite 索引只是扫描缓存。
package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gyaml"
	"github.com/gogf/gf/v2/os/gfile"
)

// HistoryDir vault 内版本快照目录名（服务写文件前自动拷贝旧版到此）
const HistoryDir = ".history"

// KnowledgeDir 顶层知识库目录名（space 缺省约定的 knowledge 空间）
const KnowledgeDir = "知识库"

// Frontmatter 文档元数据（自包含在 md 文件头部）
type Frontmatter struct {
	Title  string   `json:"title"`
	Space  string   `json:"space"`  // knowledge | work
	Type   string   `json:"type"`   // knowledge: convention/architecture/decision/runbook; work: note/api/design
	Status string   `json:"status"` // 仅 knowledge: draft | published
	Tags   []string `json:"tags"`
	Linked []string `json:"linked"` // task:123 / req:45 / tc:7
}

// RootPath 项目 vault 根目录
func RootPath(projectId int64) string {
	return filepath.Join("resource", "projects", fmt.Sprintf("%d", projectId), "vault")
}

// SafeJoin 校验 rel 不越出 vault 后拼接绝对路径；rel 以 / 或 \ 开头、含 .. 均拒绝
func SafeJoin(projectId int64, rel string) (string, error) {
	rel = strings.TrimPrefix(strings.ReplaceAll(rel, "\\", "/"), "/")
	if rel == "" || rel == "." {
		return RootPath(projectId), nil
	}
	clean := filepath.ToSlash(filepath.Clean(rel))
	if strings.HasPrefix(clean, "../") || clean == ".." || filepath.IsAbs(rel) || strings.Contains(rel, "\x00") {
		return "", fmt.Errorf("invalid vault path: %s", rel)
	}
	if strings.HasPrefix(clean, HistoryDir+"/") || clean == HistoryDir {
		return "", fmt.Errorf("access to %s is not allowed", HistoryDir)
	}
	return filepath.Join(RootPath(projectId), filepath.FromSlash(clean)), nil
}

// Checksum 内容摘要（sha256 前 16 hex 字符），用于增量扫描比对
func Checksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// ParseFrontmatter 解析 md 内容头部的 YAML frontmatter；裸文件（无 --- 块）合法，返回零值
func ParseFrontmatter(content string) (fm Frontmatter, body string, err error) {
	s := strings.TrimPrefix(content, "\ufeff")
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return Frontmatter{}, content, nil
	}
	rest := s[4:]
	if strings.HasPrefix(s, "---\r\n") {
		rest = s[5:]
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Frontmatter{}, content, nil
	}
	head := rest[:end]
	body = rest[end+4:]
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimPrefix(body, "\r\n")

	var raw struct {
		Title  string   `yaml:"title" json:"title"`
		Space  string   `yaml:"space" json:"space"`
		Type   string   `yaml:"type" json:"type"`
		Status string   `yaml:"status" json:"status"`
		Tags   []string `yaml:"tags" json:"tags"`
		Linked []string `yaml:"linked" json:"linked"`
	}
	if err = gyaml.DecodeTo([]byte(head), &raw); err != nil {
		// frontmatter 损坏不致命：按裸文件处理，避免单个坏文件拖垮扫描
		return Frontmatter{}, content, nil
	}
	fm = Frontmatter{
		Title:  strings.TrimSpace(raw.Title),
		Space:  strings.TrimSpace(raw.Space),
		Type:   strings.TrimSpace(raw.Type),
		Status: strings.TrimSpace(raw.Status),
		Tags:   raw.Tags,
		Linked: raw.Linked,
	}
	for i := range fm.Tags {
		fm.Tags[i] = strings.TrimSpace(fm.Tags[i])
	}
	for i := range fm.Linked {
		fm.Linked[i] = strings.TrimSpace(fm.Linked[i])
	}
	return fm, body, nil
}

var yamlQuoteNeeds = regexp.MustCompile(`[:#\[\]{}&*!|>'"%@,]`)

// yamlScalar 标量安全序列化：含 YAML 特殊字符或空格时加双引号转义
func yamlScalar(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, "\n\r\t") || yamlQuoteNeeds.MatchString(s) || strings.TrimSpace(s) != s {
		return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
	}
	return s
}

// RenderFrontmatter 生成带 frontmatter 的完整文件内容（字段有序，便于人读与 diff）
func RenderFrontmatter(fm Frontmatter, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("title: " + yamlScalar(fm.Title) + "\n")
	if fm.Space != "" {
		b.WriteString("space: " + yamlScalar(fm.Space) + "\n")
	}
	if fm.Type != "" {
		b.WriteString("type: " + yamlScalar(fm.Type) + "\n")
	}
	if len(fm.Tags) > 0 {
		quoted := make([]string, len(fm.Tags))
		for i, t := range fm.Tags {
			quoted[i] = yamlScalar(t)
		}
		b.WriteString("tags: [" + strings.Join(quoted, ", ") + "]\n")
	}
	if len(fm.Linked) > 0 {
		quoted := make([]string, len(fm.Linked))
		for i, l := range fm.Linked {
			quoted[i] = yamlScalar(l)
		}
		b.WriteString("linked: [" + strings.Join(quoted, ", ") + "]\n")
	}
	if fm.Status != "" {
		b.WriteString("status: " + yamlScalar(fm.Status) + "\n")
	}
	b.WriteString("---\n")
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	b.WriteString(body)
	return b.String()
}

// Snapshot 写前版本快照：旧版拷贝至 .history/{相对路径}/{yyyyMMdd-HHmmss}{扩展名}
// 文件不存在时为静默 no-op
func Snapshot(projectId int64, rel string) error {
	abs, err := SafeJoin(projectId, rel)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("cannot snapshot directory: %s", rel)
	}
	relSlash := filepath.ToSlash(rel)
	histDir := filepath.Join(RootPath(projectId), HistoryDir, filepath.FromSlash(relSlash))
	if err := os.MkdirAll(histDir, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(histDir, time.Now().Format("20060102-150405")+filepath.Ext(rel))
	return gfile.CopyFile(abs, dst)
}
