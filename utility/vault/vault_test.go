package vault

import (
	"strings"
	"testing"
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
