package aiengine

// eino 记忆/文档工具集成测试：临时工作目录（vault 相对路径根）+ 绝对路径 SQLite，
// 全程不触碰项目真实 resource/data 数据库

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cicbyte/byte-code/utility/docs"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func setupToolsTest(t *testing.T) context.Context {
	t.Helper()
	dir := t.TempDir()
	// 空占位配置文件：阻断读取项目真实 config；DB 配置由下方 gdb.SetConfig 提供
	cfgPath := filepath.Join(dir, "none.yaml")
	os.WriteFile(cfgPath, []byte{}, 0o644)
	t.Setenv("GF_GCFG_FILE", cfgPath)
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Windows 下连接池握着 db 文件会导致 TempDir 清理失败，先关连接
		_ = g.DB().Close(context.Background())
		_ = os.Chdir(oldWd)
	})

	dbPath := filepath.Join(dir, "test.db")
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{Type: "sqlite", Link: "sqlite::@file(" + dbPath + ")"},
		},
	})

	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS project_memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL,
			key TEXT NOT NULL, value TEXT DEFAULT '', status TEXT DEFAULT 'active',
			expires_at TIMESTAMP NULL, last_verified_at TIMESTAMP NULL,
			verified_by INTEGER DEFAULT 0, updated_by INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, key))`,
		`CREATE TABLE IF NOT EXISTS project_document_index (
			project_id INTEGER NOT NULL, path TEXT NOT NULL,
			space TEXT DEFAULT '', title TEXT DEFAULT '', type TEXT DEFAULT '',
			status TEXT DEFAULT 'published', tags TEXT DEFAULT '', linked TEXT DEFAULT '',
			ext TEXT DEFAULT '', size INTEGER DEFAULT 0, checksum TEXT DEFAULT '',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(project_id, path))`,
		`CREATE TABLE IF NOT EXISTS sys_config (key TEXT PRIMARY KEY, value TEXT DEFAULT '')`,
	} {
		if _, err := g.DB().Exec(context.Background(), ddl); err != nil {
			t.Fatal(err)
		}
	}

	// vault fixture：知识库已发布一篇、工作区一篇裸文件
	vroot := filepath.Join(dir, "resource", "projects", "1", "vault")
	kbDir := filepath.Join(vroot, docs.KnowledgeDir)
	if err := os.MkdirAll(kbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(kbDir, "coding.md"),
		[]byte("---\ntitle: 编码规范\nspace: knowledge\nstatus: published\n---\nGo 文件小写命名。"), 0o644)
	os.WriteFile(filepath.Join(vroot, "login.md"), []byte("# 登录设计\n旧方案。"), 0o644)

	// 索引 fixture
	g.DB().Exec(context.Background(),
		`INSERT INTO project_document_index (project_id, path, space, title, type, status, tags, linked, ext, size)
		VALUES (1, ?, 'knowledge', '编码规范', 'convention', 'published', 'go', '', '.md', 10),
		       (1, ?, 'work', '登录设计', 'note', 'published', '', 'task:9', '.md', 8)`,
		docs.KnowledgeDir+"/coding.md", "login.md")
	g.DB().Exec(context.Background(), `INSERT INTO sys_config (key, value) VALUES ('memory_stale_days', '30')`)

	return context.WithValue(context.Background(), ctxAIUserId, 42)
}

func runTool(t *testing.T, ctx context.Context, tool interface {
	InvokableRun(context.Context, string, ...ToolOption) (string, error)
}, args string) map[string]interface{} {
	t.Helper()
	out, err := tool.InvokableRun(ctx, args)
	if err != nil {
		t.Fatalf("tool run: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("json: %v\nraw: %s", err, out)
	}
	return m
}

func TestMemToolsRoundTrip(t *testing.T) {
	ctx := setupToolsTest(t)

	set := runTool(t, ctx, &MemSetTool{}, `{"projectId":1,"key":"build.cmd","value":"go build ./...","ttl":"1h"}`)
	if set["status"] != "saved" {
		t.Fatalf("mem_set: %v", set)
	}

	got := runTool(t, ctx, &MemGetTool{}, `{"projectId":1,"key":"build.cmd"}`)
	if got["value"] != "go build ./..." || got["status"] != "active" {
		t.Fatalf("mem_get: %v", got)
	}

	// aiUserId 注入来源标识
	if v, _ := g.DB().Model("project_memories").Where("key", "build.cmd").Value("updated_by"); v.Int() != 42 {
		t.Fatalf("updated_by = %v, want 42", v)
	}

	verified := runTool(t, ctx, &MemVerifyTool{}, `{"projectId":1,"key":"build.cmd"}`)
	if verified["status"] != "verified" {
		t.Fatalf("mem_verify: %v", verified)
	}

	list := runTool(t, ctx, &MemListTool{}, `{"projectId":1,"prefix":"build."}`)
	l := list["list"].([]interface{})
	if len(l) != 1 || l[0].(map[string]interface{})["key"] != "build.cmd" {
		t.Fatalf("mem_list: %v", list)
	}

	// 腐化提示：把 last_verified_at 拨回 40 天前
	g.DB().Model("project_memories").Where("key", "build.cmd").
		Data("last_verified_at", gtime.Now().AddDate(0, 0, -40)).Update()
	stale := runTool(t, ctx, &MemGetTool{}, `{"projectId":1,"key":"build.cmd"}`)
	if stale["status"] != "stale" || stale["hint"] == nil {
		t.Fatalf("stale hint: %v", stale)
	}

	// 非法输入
	if _, err := (&MemSetTool{}).InvokableRun(ctx, `{"projectId":1,"key":"bad key","value":"x"}`); err == nil {
		t.Fatal("key 含空格应拒绝")
	}
	if _, err := (&MemSetTool{}).InvokableRun(ctx, `{"projectId":1,"key":"k","value":"x","ttl":"3x"}`); err == nil {
		t.Fatal("非法 ttl 应拒绝")
	}
}

func TestDocTools(t *testing.T) {
	ctx := setupToolsTest(t)

	// read_doc：文本 + frontmatter
	rd := runTool(t, ctx, &ReadDocTool{}, `{"projectId":1,"path":"知识库/coding.md"}`)
	if rd["title"] != "编码规范" || rd["content"] != "Go 文件小写命名。" {
		t.Fatalf("read_doc: %v", rd)
	}

	// kb_get_conventions：只有 published 的知识库文档
	kb := runTool(t, ctx, &KbConventionsTool{}, `{"projectId":1}`)
	items := kb["items"].([]interface{})
	if len(items) != 1 || items[0].(map[string]interface{})["path"] != "知识库/coding.md" {
		t.Fatalf("kb_get_conventions: %v", kb)
	}

	// doc_linked：task:9 命中 login.md
	dl := runTool(t, ctx, &DocLinkedTool{}, `{"projectId":1,"targetType":"task","targetId":9}`)
	items = dl["items"].([]interface{})
	if len(items) != 1 || items[0].(map[string]interface{})["path"] != "login.md" {
		t.Fatalf("doc_linked: %v", dl)
	}

	// search_docs：标题命中（返回 path 数组）
	out, err := (&SearchDocsTool{}).InvokableRun(ctx, `{"projectId":1,"keyword":"编码"}`)
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &rows); err != nil || len(rows) == 0 {
		t.Fatalf("search_docs: %v\nraw: %s", err, out)
	}
	if rows[0]["path"] != "知识库/coding.md" {
		t.Fatalf("search_docs 首条应为知识库文档: %v", rows[0])
	}

	// read_doc 路径穿越拒绝
	if _, err := (&ReadDocTool{}).InvokableRun(ctx, `{"projectId":1,"path":"../../../etc/passwd"}`); err == nil {
		t.Fatal("路径穿越应拒绝")
	}
}

func TestGetToolsRegistration(t *testing.T) {
	ctx := context.Background()
	tools := GetTools()
	if len(tools) != 10 {
		t.Fatalf("工具数 = %d, want 10", len(tools))
	}
	names := map[string]bool{}
	for _, tl := range tools {
		info, err := tl.Info(ctx)
		if err != nil || info == nil || info.Name == "" {
			t.Fatalf("工具 Info 失败: %v", err)
		}
		if info.ParamsOneOf == nil {
			t.Fatalf("工具 %s 缺参数定义", info.Name)
		}
		names[info.Name] = true
	}
	for _, want := range []string{"list_tasks", "get_project", "search_docs", "read_doc",
		"kb_get_conventions", "doc_linked", "mem_list", "mem_get", "mem_set", "mem_verify"} {
		if !names[want] {
			t.Fatalf("缺少工具: %s", want)
		}
	}
}

func TestBindProjectIdSecurityBoundary(t *testing.T) {
	// P0-1 安全边界：任务上下文注入的项目 id 是唯一可信来源
	ctx := context.WithValue(context.Background(), ctxTaskProjectId, 1)

	// 一致 → 放行并归一
	pid, err := bindProjectId(ctx, 1)
	if err != nil || pid != 1 {
		t.Fatalf("一致 projectId 应放行: %d %v", pid, err)
	}
	// 缺省（LLM 未传）→ 用任务归属补齐
	pid, err = bindProjectId(ctx, 0)
	if err != nil || pid != 1 {
		t.Fatalf("缺省 projectId 应取任务归属: %d %v", pid, err)
	}
	// 不一致（诱导跨项目）→ 拒绝
	if _, err := bindProjectId(ctx, 2); err == nil {
		t.Fatal("projectId=2 与任务归属 1 不符应拒绝")
	}
	// 工具层端到端：ctx 绑定项目 1，LLM 传 projectId=2 → read_doc 拒绝且不触盘
	bound := runToolErr(t, &ReadDocTool{}, ctx, `{"projectId":2,"path":"知识库/coding.md"}`)
	if bound == nil {
		t.Fatal("跨项目 read_doc 应报错")
	}
	// 无任务上下文（手工调试）→ 参数值放行（回归保护）
	pid, err = bindProjectId(context.Background(), 3)
	if err != nil || pid != 3 {
		t.Fatalf("无任务上下文应退回参数值: %d %v", pid, err)
	}
}

func runToolErr(t *testing.T, tool interface {
	InvokableRun(context.Context, string, ...ToolOption) (string, error)
}, ctx context.Context, args string) error {
	t.Helper()
	_, err := tool.InvokableRun(ctx, args)
	return err
}
