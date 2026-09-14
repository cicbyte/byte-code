package project

// 权限矩阵回归测试：把 2026-09 权限三阶段（平台字典 / 项目三档 / Agent
// 能力集）与 P0 删除门禁修复的关键用例固化为可重复执行的断言。
//
// DB 走临时 SQLite + 全量迁移：resource/sql 复制进临时目录后跑
// dbinit.AutoMigrate（同时兼作迁移链的冒烟测试），避免与 dev 实例的
// resource/data/app.lock 争锁、也绝不触碰真实数据。
// 配置用空文件顶掉仓库 config.yaml（GF_GCFG_FILE），数据库配置由
// gdb.SetConfig 注入——同 internal/logic/aiengine/engine_test.go 的既有模式。

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/utility/dbinit"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

var s = &sProject{}

func ctxAs(uid int) context.Context {
	return context.WithValue(context.Background(), "userId", uid)
}

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "permmatrix-*")
	if err != nil {
		panic(err)
	}
	// resource/sql → 临时目录（迁移按 CWD 相对路径读取；go test 的 CWD 是包
	// 目录，须用本文件位置反推仓库根）
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
	if err := copyDir(filepath.Join(repoRoot, "resource", "sql"), filepath.Join(tmp, "resource", "sql")); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "resource", "data"), 0o755); err != nil {
		panic(err)
	}
	// 空配置顶掉仓库 config.yaml，防止 g.DB() 解析到 dev 库
	emptyCfg := filepath.Join(tmp, "empty.yaml")
	_ = os.WriteFile(emptyCfg, []byte{}, 0o644)
	_ = os.Setenv("GF_GCFG_FILE", emptyCfg)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		panic(err)
	}
	gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{
		gdb.ConfigNode{Type: "sqlite", Link: "sqlite::@file(" + filepath.Join(tmp, "t.db") + ")"},
	}})
	if err := dbinit.AutoMigrate(context.Background()); err != nil {
		panic("AutoMigrate: " + err.Error())
	}
	seed()

	code := m.Run()
	_ = g.DB().Close(context.Background())
	_ = os.Chdir(oldWd)
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

// seed 固定 id 段：用户 101-110 / 项目 501 / 敏捷实体 6xx-7xx，避免与种子数据碰撞
func seed() {
	db := g.DB()
	ctx := context.Background()
	must := func(err error, what string) {
		if err != nil {
			panic(what + ": " + err.Error())
		}
	}
	for _, u := range []map[string]interface{}{
		{"id": 101, "username": "mx-owner", "password": "x", "type": "human", "status": 1},
		{"id": 102, "username": "mx-mnt", "password": "x", "type": "human", "status": 1},
		{"id": 103, "username": "mx-mb", "password": "x", "type": "human", "status": 1},
		{"id": 104, "username": "mx-admin", "password": "x", "type": "human", "status": 1},
		{"id": 105, "username": "mx-r2", "password": "x", "type": "human", "status": 1},
		{"id": 106, "username": "mx-norole", "password": "x", "type": "human", "status": 1},
		{"id": 107, "username": "mx-ai-legacy", "password": "x", "type": "ai", "status": 1},
		{"id": 108, "username": "mx-ai-ro", "password": "x", "type": "ai", "status": 1},
		{"id": 109, "username": "mx-ai-full", "password": "x", "type": "ai", "status": 1},
	} {
		_, err := db.Model("sys_users").Ctx(ctx).Data(u).Insert()
		must(err, "seed user "+u["username"].(string))
	}
	// 平台角色：104=role1（超管）、105=role2（迁移68 预绑 tags+create）
	_, err := db.Exec(ctx, "INSERT INTO sys_user_roles (user_id, role_id) VALUES (104,1),(105,2)")
	must(err, "seed user roles")

	_, err = db.Model("projects").Ctx(ctx).Data(g.Map{"id": 501, "name": "mx-proj", "status": 1, "creator_id": 101}).Insert()
	must(err, "seed project")
	for _, m := range []map[string]interface{}{
		{"project_id": 501, "user_id": 101, "role": "owner"},
		{"project_id": 501, "user_id": 102, "role": "maintainer"},
		{"project_id": 501, "user_id": 103, "role": "member"},
	} {
		_, err := db.Model("project_members").Ctx(ctx).Data(m).Insert()
		must(err, "seed member")
	}
	for _, b := range []map[string]interface{}{
		{"agent_id": 108, "project_id": 501, "role": "member", "capabilities": "tasks_read,docs_read"},
		{"agent_id": 109, "project_id": 501, "role": "member", "capabilities": ""},
	} {
		_, err := db.Model("agent_project_bindings").Ctx(ctx).Data(b).Insert()
		must(err, "seed binding")
	}
}

func newTask(t *testing.T, creator int) int {
	t.Helper()
	res, err := g.DB().Model("tasks").Ctx(ctxAs(creator)).Data(g.Map{
		"project_id": 501, "title": "mx-task", "creator_id": creator, "status": "open",
	}).Insert()
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

// ---------- P1 平台字典 ----------

func TestRequireMenuPermMatrix(t *testing.T) {
	cases := []struct {
		uid  int
		key  string
		want bool // true=放行
	}{
		{106, "platform_tags", false}, {106, "platform_groups", false}, {106, "project_create", false},
		{105, "platform_tags", true}, {105, "project_create", true}, {105, "platform_groups", false},
		{104, "platform_tags", true}, {104, "platform_groups", true}, {104, "project_create", true},
	}
	for _, c := range cases {
		err := perm.RequireMenuPerm(ctxAs(c.uid), c.key)
		if (err == nil) != c.want {
			t.Errorf("uid=%d key=%s: 放行=%v, 期望 %v (%v)", c.uid, c.key, err == nil, c.want, err)
		}
	}
}

// ---------- P2 项目三档 ----------

func TestProjectTiers(t *testing.T) {
	mnt := []struct {
		uid int
		want bool
	}{{101, true}, {102, true}, {103, false}, {106, false}, {104, true}}
	for _, c := range mnt {
		if got := perm.IsProjectMaintainer(ctxAs(c.uid), c.uid, 501); got != c.want {
			t.Errorf("IsProjectMaintainer(%d)=%v want %v", c.uid, got, c.want)
		}
	}
	own := []struct {
		uid int
		want bool
	}{{101, true}, {102, false}, {104, true}}
	for _, c := range own {
		if got := perm.IsProjectOwner(ctxAs(c.uid), c.uid, 501); got != c.want {
			t.Errorf("IsProjectOwner(%d)=%v want %v", c.uid, got, c.want)
		}
	}
}

func TestMemberRemoveTiers(t *testing.T) {
	// 供移除用的一次性成员
	if _, err := g.DB().Model("project_members").Ctx(ctxAs(102)).
		Data(g.Map{"project_id": 501, "user_id": 110, "role": "member"}).Insert(); err != nil {
		// 已存在（重跑）则忽略
		_ = err
	}
	if err := s.RemoveMember(ctxAs(103), 501, 110); err == nil {
		t.Error("member 移除他人应被拒")
	}
	if err := s.RemoveMember(ctxAs(102), 501, 102); err == nil {
		t.Error("maintainer 移除自己（maintainer）应被拒")
	}
	if err := s.RemoveMember(ctxAs(102), 501, 101); err == nil {
		t.Error("maintainer 移除 owner 应被拒")
	}
	if err := s.RemoveMember(ctxAs(101), 501, 110); err != nil {
		t.Errorf("owner 移除 member 应放行: %v", err)
	}
}

// ---------- P3 Agent 能力集 ----------

func TestAgentRequire(t *testing.T) {
	cases := []struct {
		uid  int
		cap  string
		want bool
	}{
		{103, "tasks_write", true},  // 人类直通
		{107, "tasks_write", true},  // 无绑定行（members 表手工 agent）= 全能力兼容
		{109, "tasks_write", true},  // 空能力集 = 全能力
		{108, "tasks_read", true},   // 授予项
		{108, "tasks_write", false}, // 未授予
		{108, "docs_write", false},
		{108, "memory_write", false},
	}
	for _, c := range cases {
		err := perm.AgentRequire(ctxAs(c.uid), 501, c.cap)
		if (err == nil) != c.want {
			t.Errorf("AgentRequire(uid=%d, %s) 放行=%v want %v (%v)", c.uid, c.cap, err == nil, c.want, err)
		}
	}
}

// ---------- P0 删除门禁（2026-09-16 审查实测越权的回归锚点） ----------

func TestDeleteTaskGates(t *testing.T) {
	if err := s.DeleteTask(ctxAs(103), newTask(t, 101)); err == nil {
		t.Error("member 删除他人任务应被拒")
	}
	if err := s.DeleteTask(ctxAs(103), newTask(t, 103)); err != nil {
		t.Errorf("member 删除自建任务应放行: %v", err)
	}
	if err := s.DeleteTask(ctxAs(102), newTask(t, 101)); err != nil {
		t.Errorf("maintainer 删除任务应放行: %v", err)
	}
	if err := s.DeleteTask(ctxAs(108), newTask(t, 101)); err == nil {
		t.Error("只读 agent（tasks_read）删除任务应被拒")
	}
	if err := s.DeleteTask(ctxAs(109), newTask(t, 101)); err == nil {
		t.Error("全能力 agent 删除他人任务应被拒（非 creator 非 maintainer）")
	}
	if err := s.DeleteTask(ctxAs(109), newTask(t, 109)); err != nil {
		t.Errorf("全能力 agent 删除自建任务应放行: %v", err)
	}
}

func newAgile(t *testing.T, table, title string) int {
	t.Helper()
	// status 走各表默认值（枚举 CHECK 约束不同，统一传值会触约束）；
	// sprints 的 start/end_date NOT NULL 无默认
	data := g.Map{"project_id": 501, "title": title, "name": title}
	if table == "sprints" {
		data["start_date"] = "2026-09-01"
		data["end_date"] = "2026-09-30"
	}
	res, err := g.DB().Model(table).Ctx(context.Background()).Data(data).Insert()
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return int(id)
}

func TestMaintainerTaskBypass(t *testing.T) {
	// assignee 之外，maintainer 也可完成/阻塞/解除（P2 #419 补齐口径）
	tid := newTask(t, 101)
	if err := s.CompleteTask(ctxAs(102), &api.TaskCompleteReq{Id: tid}); err != nil {
		t.Errorf("maintainer 完成他人任务应放行: %v", err)
	}
	tid2 := newTask(t, 101)
	if _, err := g.DB().Model("tasks").Ctx(ctxAs(102)).Where("id", tid2).
		Data(g.Map{"assignee_id": 102, "status": "in_progress"}).Update(); err != nil {
		t.Fatal(err)
	}
	if err := s.BlockTask(ctxAs(102), &api.TaskBlockReq{Id: tid2, Reason: "x"}); err != nil {
		t.Errorf("maintainer 上报阻塞应放行: %v", err)
	}
	if err := s.UnblockTask(ctxAs(102), &api.TaskUnblockReq{Id: tid2}); err != nil {
		t.Errorf("maintainer 解除阻塞应放行: %v", err)
	}
}

func TestDeleteAgileGates(t *testing.T) {
	if err := s.DeleteRequirement(ctxAs(103), newAgile(t, "requirements", "r")); err == nil {
		t.Error("member 删除需求应被拒")
	}
	if err := s.DeleteRequirement(ctxAs(102), newAgile(t, "requirements", "r2")); err != nil {
		t.Errorf("maintainer 删除需求应放行: %v", err)
	}
	if err := s.DeleteMilestone(ctxAs(103), newAgile(t, "milestones", "m")); err == nil {
		t.Error("member 删除里程碑应被拒")
	}
	if err := s.DeleteSprint(ctxAs(103), newAgile(t, "sprints", "s")); err == nil {
		t.Error("member 删除迭代应被拒")
	}
	if err := s.DeleteSprint(ctxAs(102), newAgile(t, "sprints", "s2")); err != nil {
		t.Errorf("maintainer 删除迭代应放行: %v", err)
	}
}

// ---------- P2 watcher 订阅（#424） ----------

// watchNotifCount 某用户在某任务上指定标题的通知条数
func watchNotifCount(t *testing.T, uid, taskId int, title string) int {
	t.Helper()
	n, err := g.DB().Model("notifications").Ctx(ctxAs(uid)).
		Where("user_id", uid).Where("source_type", "task").
		Where("source_id", taskId).Where("title", title).Count()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTaskWatchers(t *testing.T) {
	tid := newTask(t, 101)

	// 关注与幂等：重复 watch 不报错不重复行
	if err := s.WatchTask(ctxAs(103), tid); err != nil {
		t.Fatalf("member 关注任务应放行: %v", err)
	}
	if err := s.WatchTask(ctxAs(103), tid); err != nil {
		t.Errorf("重复关注应幂等: %v", err)
	}
	// 只读 agent（tasks_read）可关注：订阅是读语义
	if err := s.WatchTask(ctxAs(108), tid); err != nil {
		t.Errorf("只读 agent 关注任务应放行: %v", err)
	}

	// 详情回填：watching/watcherCount/watchers
	det, err := s.GetTask(ctxAs(103), tid)
	if err != nil {
		t.Fatal(err)
	}
	if !det.Watching || det.WatcherCount != 2 || len(det.Watchers) != 2 {
		t.Errorf("详情回填错误: watching=%v count=%d names=%v", det.Watching, det.WatcherCount, det.Watchers)
	}

	// 评论扇出：creator(101) 评论 → 两位 watcher 各一条「任务新评论」，作者自己不收
	if _, err := s.CreateComment(ctxAs(101), &api.CommentCreateReq{TaskId: tid, Content: "watcher 扇出验证"}); err != nil {
		t.Fatal(err)
	}
	if n := watchNotifCount(t, 103, tid, "任务新评论"); n != 1 {
		t.Errorf("watcher103 应收 1 条新评论通知, got %d", n)
	}
	if n := watchNotifCount(t, 108, tid, "任务新评论"); n != 1 {
		t.Errorf("watcher108 应收 1 条新评论通知, got %d", n)
	}
	if n := watchNotifCount(t, 101, tid, "任务新评论"); n != 0 {
		t.Errorf("评论作者不应收到自己的评论通知, got %d", n)
	}

	// 完成扇出：认领→完成 → watcher 收「关注的任务待审核」
	if err := s.ClaimTask(ctxAs(103), &api.TaskClaimReq{Id: tid}); err != nil {
		t.Fatal(err)
	}
	if n := watchNotifCount(t, 108, tid, "关注的任务被认领"); n != 1 {
		t.Errorf("watcher108 应收 1 条被认领通知, got %d", n)
	}
	if err := s.CompleteTask(ctxAs(103), &api.TaskCompleteReq{Id: tid}); err != nil {
		t.Fatal(err)
	}
	if n := watchNotifCount(t, 108, tid, "关注的任务待审核"); n != 1 {
		t.Errorf("watcher108 应收 1 条待审核通知, got %d", n)
	}

	// 取关后不再收
	if err := s.UnwatchTask(ctxAs(108), tid); err != nil {
		t.Fatalf("取关应放行: %v", err)
	}
	if _, err := s.CreateComment(ctxAs(101), &api.CommentCreateReq{TaskId: tid, Content: "取关后验证"}); err != nil {
		t.Fatal(err)
	}
	if n := watchNotifCount(t, 108, tid, "任务新评论"); n != 1 {
		t.Errorf("取关后不应再收新评论通知, got %d", n)
	}
	if n := watchNotifCount(t, 103, tid, "任务新评论"); n != 2 {
		t.Errorf("仍关注的 watcher 应继续收通知, got %d", n)
	}

	// 删除任务级联清理关注行
	if err := s.DeleteTask(ctxAs(101), tid); err != nil {
		t.Fatalf("owner 删除任务应放行: %v", err)
	}
	if n, _ := g.DB().Model("task_watchers").Ctx(ctxAs(101)).Where("task_id", tid).Count(); n != 0 {
		t.Errorf("删除任务应级联清理 watcher, 残留 %d 行", n)
	}
}
