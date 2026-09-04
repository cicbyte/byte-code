package docs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func TestParseFrontmatterFull(t *testing.T) {
	content := "---\ntitle: 登录模块重构设计\nspace: knowledge\ntype: decision\ntags: [auth, security]\nstatus: draft\nlinked: [task:123, req:45]\n---\n\n正文开始\n第二行\n"
	fm, body, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm.Title != "登录模块重构设计" || fm.Space != "knowledge" || fm.Type != "decision" || fm.Status != "draft" {
		t.Fatalf("frontmatter mismatch: %+v", fm)
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "auth" || fm.Tags[1] != "security" {
		t.Fatalf("tags mismatch: %v", fm.Tags)
	}
	if len(fm.Linked) != 2 || fm.Linked[0] != "task:123" || fm.Linked[1] != "req:45" {
		t.Fatalf("linked mismatch: %v", fm.Linked)
	}
	// 关闭符 --- 行尾换行被剥掉；正文前作者留的空行保留
	if body != "\n正文开始\n第二行\n" {
		t.Fatalf("body mismatch: %q", body)
	}
}

func TestParseFrontmatterBare(t *testing.T) {
	fm, body, err := ParseFrontmatter("没有 frontmatter 的裸文件\n")
	if err != nil || fm.Title != "" || body != "没有 frontmatter 的裸文件\n" {
		t.Fatalf("bare file mismatch: fm=%+v body=%q err=%v", fm, body, err)
	}
}

func TestParseFrontmatterBroken(t *testing.T) {
	// frontmatter 语法损坏：按裸文件处理，不报错
	fm, body, err := ParseFrontmatter("---\ntitle: [broken\n---\nbody")
	if err != nil || fm.Title != "" || !strings.Contains(body, "body") {
		t.Fatalf("broken fm should degrade to bare: %+v %q %v", fm, body, err)
	}
}

func TestParseFrontmatterScalarTags(t *testing.T) {
	// 手写逗号标量（tags: go,api / linked: task:3）：曾因 []string 严格解码
	// 导致整个 frontmatter 静默丢失、索引只剩默认值
	content := "---\ntitle: 手写文档\nspace: knowledge\ntype: convention\nstatus: published\ntags: go,api\nlinked: task:3,req:9\n---\n\n正文\n"
	fm, _, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm.Title != "手写文档" || fm.Space != "knowledge" || fm.Type != "convention" || fm.Status != "published" {
		t.Fatalf("scalar frontmatter mismatch: %+v", fm)
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "go" || fm.Tags[1] != "api" {
		t.Fatalf("scalar tags mismatch: %v", fm.Tags)
	}
	if len(fm.Linked) != 2 || fm.Linked[0] != "task:3" || fm.Linked[1] != "req:9" {
		t.Fatalf("scalar linked mismatch: %v", fm.Linked)
	}
}

func TestFrontmatterRoundTrip(t *testing.T) {
	fm := Frontmatter{
		Title:  `标题: 含"引号"与\反斜杠`,
		Space:  "work",
		Type:   "design",
		Status: "draft",
		Tags:   []string{"a b", "c,d", "普通"},
		Linked: []string{"task:1"},
	}
	out := RenderFrontmatter(fm, "内容")
	got, body, err := ParseFrontmatter(out)
	if err != nil {
		t.Fatalf("parse rendered: %v\n%s", err, out)
	}
	if got.Title != fm.Title || got.Space != fm.Space || got.Type != fm.Type || got.Status != fm.Status {
		t.Fatalf("round-trip scalar mismatch: %+v", got)
	}
	if strings.Join(got.Tags, "|") != strings.Join(fm.Tags, "|") {
		t.Fatalf("tags mismatch: %v", got.Tags)
	}
	if strings.Join(got.Linked, "|") != strings.Join(fm.Linked, "|") {
		t.Fatalf("linked mismatch: %v", got.Linked)
	}
	if body != "内容\n" {
		t.Fatalf("body mismatch: %q", body)
	}
}

func TestSafeJoin(t *testing.T) {
	cases := []struct {
		rel  string
		ok   bool
		want string
	}{
		{"设计/login.md", true, "设计/login.md"},
		{"./设计/a.md", true, "设计/a.md"},
		{"a/../../etc/passwd", false, ""},
		{"../escape.md", false, ""},
		{".history/x.md", false, ""},
		{"/abs/path.md", true, "abs/path.md"},
	}
	for _, c := range cases {
		got, err := SafeJoin(1, c.rel)
		if c.ok && err != nil {
			t.Fatalf("SafeJoin(%q) unexpected err: %v", c.rel, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("SafeJoin(%q) should reject", c.rel)
		}
		if c.ok {
			gotSlash := strings.ReplaceAll(got, "\\", "/")
			if !strings.HasSuffix(gotSlash, c.want) {
				t.Fatalf("SafeJoin(%q) = %q, want suffix %q", c.rel, gotSlash, c.want)
			}
		}
	}
}

func TestSanitizeName(t *testing.T) {
	cases := map[string]string{
		`正常名称`:     "正常名称",
		`a/b\c:d*e?f"g<h>i|j`: "a_b_c_d_e_f_g_h_i_j",
		`  空白  `:    "空白",
		`...`:       "untitled",
		"":          "untitled",
	}
	for in, want := range cases {
		if got := sanitizeName(in); got != want {
			t.Fatalf("sanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSnapshotNoClobber(t *testing.T) {
	// P0-6：同一秒（甚至同一毫秒）内两次快照必须产生两份独立文件，
	// 旧实现秒级时间戳同名互覆会静默丢历史
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	rel := "知识库/doc.md"
	abs := filepath.Join(RootPath(1), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 连续两次快照（Windows 计时精度内大概率同毫秒）
	if err := Snapshot(1, rel); err != nil {
		t.Fatal(err)
	}
	if err := Snapshot(1, rel); err != nil {
		t.Fatal(err)
	}
	histDir := filepath.Join(RootPath(1), HistoryDir, filepath.FromSlash(rel))
	entries, err := os.ReadDir(histDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("两次快照应有两份文件，实际 %d 份（互覆）", len(entries))
	}
}

func TestSafeJoinHistoryBypass(t *testing.T) {
	// S-1：Windows 命名空间变体不得访问 .history（大小写/尾点/尾空格/8.3 短名）
	deny := []string{
		".history/x.md", ".History/x.md", ".HISTORY/x.md",
		".history./x.md", ".history /x.md", ".history../x.md",
		"HISTOR~1/x.md", "histor~1/x.md", ".history",
	}
	for _, p := range deny {
		if _, err := SafeJoin(1, p); err == nil {
			t.Fatalf("SafeJoin(%q) 应拒绝（Windows 命名空间变体可解析进快照目录）", p)
		}
	}
	// 正常路径不受影响
	allow := []string{"知识库/doc.md", ".hidden.md", "设计 v2/x.md", "hist/doc.md"}
	for _, p := range allow {
		if _, err := SafeJoin(1, p); err != nil {
			t.Fatalf("SafeJoin(%q) 不应拒绝: %v", p, err)
		}
	}
}

func TestUpsertAndDeletePath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/none.yaml"
	os.WriteFile(cfgPath, []byte{}, 0o644)
	t.Setenv("GF_GCFG_FILE", cfgPath)
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = g.DB().Close(context.Background())
		_ = os.Chdir(oldWd)
	})

	dbPath := dir + "/t.db"
	gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{
		gdb.ConfigNode{Type: "sqlite", Link: "sqlite::@file(" + dbPath + ")"},
	}})
	_, _ = g.DB().Exec(context.Background(), `CREATE TABLE project_document_index (
		project_id INTEGER NOT NULL, path TEXT NOT NULL, space TEXT DEFAULT '', title TEXT DEFAULT '',
		type TEXT DEFAULT '', status TEXT DEFAULT 'published', tags TEXT DEFAULT '', linked TEXT DEFAULT '',
		ext TEXT DEFAULT '', size INTEGER DEFAULT 0, checksum TEXT DEFAULT '',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(project_id, path))`)

	rel := "知识库/upsert.md"
	abs, _ := SafeJoin(1, rel)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	os.WriteFile(abs, []byte("---\ntitle: 增量测试\nspace: knowledge\n---\n正文"), 0o644)

	// UpsertPath：新建行
	if err := UpsertPath(context.Background(), 1, rel); err != nil {
		t.Fatalf("UpsertPath: %v", err)
	}
	v, _ := g.DB().Model("project_document_index").Ctx(context.Background()).
		Where("project_id", 1).Where("path", rel).Value("title")
	if v.String() != "增量测试" {
		t.Fatalf("title = %q", v.String())
	}

	// 内容变更后再 Upsert：checksum 更新（非重复插入）
	os.WriteFile(abs, []byte("---\ntitle: 增量测试2\nspace: knowledge\n---\n正文v2"), 0o644)
	if err := UpsertPath(context.Background(), 1, rel); err != nil {
		t.Fatal(err)
	}
	cnt, _ := g.DB().Model("project_document_index").Ctx(context.Background()).
		Where("project_id", 1).Where("path", rel).Count()
	if cnt != 1 {
		t.Fatalf("重复 upsert 应保持单行，实际 %d", cnt)
	}

	// 子路径批量删除
	for _, p := range []string{"设计/sub/x.md", "设计/sub/y.md", "设计/other.md"} {
		a, _ := SafeJoin(1, p)
		os.MkdirAll(filepath.Dir(a), 0o755)
		os.WriteFile(a, []byte("x"), 0o644)
		UpsertPath(context.Background(), 1, p)
	}
	if err := DeletePath(context.Background(), 1, "设计/sub"); err != nil {
		t.Fatal(err)
	}
	rows, _ := g.DB().Model("project_document_index").Ctx(context.Background()).
		Where("project_id", 1).Where("path LIKE ?", "设计/%").All()
	paths := []string{}
	for _, r := range rows {
		paths = append(paths, r["path"].String())
	}
	if len(paths) != 1 || paths[0] != "设计/other.md" {
		t.Fatalf("子树删除应只留 other.md，实际 %v", paths)
	}
}
