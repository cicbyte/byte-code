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
	"errors"
	"regexp"
	"strings"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	api "github.com/cicbyte/byte-code/api/v1/project"
	apiTest "github.com/cicbyte/byte-code/api/v1/test"
	platApi "github.com/cicbyte/byte-code/api/v1/platform"
	apiDocs "github.com/cicbyte/byte-code/api/v1/docs"
	_ "github.com/cicbyte/byte-code/internal/logic/docs"
	"github.com/cicbyte/byte-code/internal/logic/attachment"
	_ "github.com/cicbyte/byte-code/internal/logic/platform"
	_ "github.com/cicbyte/byte-code/internal/logic/test"
	"github.com/cicbyte/byte-code/internal/service"
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
		// 107：members 表手工加的早期 agent（无 bindings 行）——兼容路径的真实建模
		{"project_id": 501, "user_id": 107, "role": "member"},
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
	// 111：迁移 69 前的存量绑定——capabilities 为 NULL（非空串）。
	// v0.3.0 曾把 NULL 误判为无绑定行锁死存量 agent（v0.3.1 修正），此处锚定
	if _, err := db.Model("sys_users").Ctx(ctx).Data(g.Map{
		"id": 111, "username": "mx-ai-nullcaps", "password": "x", "type": "ai", "status": 1,
	}).Insert(); err != nil {
		panic("seed user 111: " + err.Error())
	}
	if _, err := db.Model("agent_project_bindings").Ctx(ctx).Data(g.Map{
		"agent_id": 111, "project_id": 501, "role": "member",
	}).Insert(); err != nil {
		panic("seed binding 111: " + err.Error())
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
		{111, "tasks_write", true},  // 存量 NULL capabilities（迁移69前）= 全能力
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

// ---------- P3 @全员（#429） ----------

func TestMentionAll(t *testing.T) {
	tid := newTask(t, 101)

	// 人类评论带 @全员：项目全体成员（103）与绑定 agent（108）各收一条
	if _, err := s.CreateComment(ctxAs(101), &api.CommentCreateReq{
		TaskId: tid, Content: "通知 @全员 今晚发版",
	}); err != nil {
		t.Fatal(err)
	}
	n103 := watchNotifCount(t, 103, tid, "评论提及了全员")
	n108 := watchNotifCount(t, 108, tid, "评论提及了全员")
	n101 := watchNotifCount(t, 101, tid, "评论提及了全员")
	if n103 != 1 || n108 != 1 {
		t.Errorf("@全员应通知成员103与agent108, got 103:%d 108:%d", n103, n108)
	}
	if n101 != 0 {
		t.Errorf("评论作者不应收到自己的@全员通知, got %d", n101)
	}

	// agent 评论带 @全员：整条拒绝（明确报错优于静默不广播）
	if _, err := s.CreateComment(ctxAs(108), &api.CommentCreateReq{
		TaskId: tid, Content: "agent 也 @全员 试试",
	}); err == nil {
		t.Error("agent 评论带 @全员 应被拒绝")
	}

	// 宽口径：@全员后跟汉字（非 ASCII 词字节）按边界处理——仍算提及（与
	// @用户名 的宁可多命中不漏报口径一致），103 再收一条累计 2
	if _, err := s.CreateComment(ctxAs(101), &api.CommentCreateReq{
		TaskId: tid, Content: "这个 @全员通知 功能不错",
	}); err != nil {
		t.Fatal(err)
	}
	if n := watchNotifCount(t, 103, tid, "评论提及了全员"); n != 2 {
		t.Errorf("第二次 @全员（后跟汉字）后 103 应累计 2 条, got %d", n)
	}

	// 清理：本测试产生的评论与通知
	g.DB().Model("comments").Ctx(ctxAs(101)).Where("task_id", tid).Delete()
	g.DB().Model("notifications").Ctx(ctxAs(101)).Where("source_type", "task").Where("source_id", tid).Delete()
	g.DB().Model("tasks").Ctx(ctxAs(101)).Where("id", tid).Delete()
}

// ---------- 我的任务统计（个人效率视图） ----------

func TestMyTaskStats(t *testing.T) {
	ctx := context.Background()
	db := g.DB()
	ago := func(days int, hourOffset int) string {
		return time.Now().AddDate(0, 0, -days).Add(time.Duration(hourOffset) * time.Hour).Format("2006-01-02 15:04:05")
	}
	// 独立项目：103 非成员（非管理员作用域应排除该项目任务），104 管理员不受限
	if _, err := db.Model("projects").Ctx(ctx).Data(g.Map{"id": 502, "name": "mx-proj2", "code": "MXP2", "status": 1, "creator_id": 101}).Insert(); err != nil {
		t.Fatal(err)
	}
	ins := func(assignee, project int, status, due, created, completed string) {
		t.Helper()
		if _, err := db.Model("tasks").Ctx(ctx).Data(g.Map{
			"project_id": project, "title": "mx-stat", "creator_id": 101,
			"assignee_id": assignee, "status": status, "due_date": due,
			"created_at": created, "completed_at": completed,
		}).Insert(); err != nil {
			t.Fatal(err)
		}
	}
	yday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tmrw := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	ins(103, 501, "open", yday, ago(3, 0), "")            // 逾期
	ins(103, 501, "open", tmrw, ago(3, 0), "")            // 未逾期
	ins(103, 501, "in_progress", "", ago(2, 0), "")
	ins(103, 501, "blocked", "", ago(2, 0), "")
	ins(103, 501, "done", "", ago(1, -1), ago(0, -1))     // 24h 周期，进 30 天趋势
	ins(103, 501, "done", "", ago(35, -2), ago(35, 0))    // 2h 周期，出趋势进 90 天均值
	ins(103, 501, "done", "", ago(102, 0), ago(100, 0))   // 双窗口外
	ins(103, 502, "open", "", ago(1, 0), "")              // 非成员项目：非管理员不计
	ins(104, 502, "review", "", ago(1, 0), "")            // 管理员跨项目计数

	res, err := s.MyTaskStats(ctxAs(103), &api.MyTaskStatsReq{})
	if err != nil {
		t.Fatal(err)
	}
	byStatus := map[string]int{}
	for _, sc := range res.StatusCounts {
		byStatus[sc.Status] = sc.Count
	}
	if byStatus["open"] != 2 || byStatus["in_progress"] != 1 || byStatus["blocked"] != 1 || byStatus["done"] != 3 {
		t.Errorf("状态计数不符: %+v", byStatus)
	}
	if res.ActiveTotal != 4 {
		t.Errorf("活跃合计应 4, got %d", res.ActiveTotal)
	}
	if res.Overdue != 1 {
		t.Errorf("逾期应 1, got %d", res.Overdue)
	}
	if len(res.Trend) != 30 {
		t.Fatalf("趋势应 30 天零填充, got %d", len(res.Trend))
	}
	sum := 0
	for _, p := range res.Trend {
		sum += p.Done
	}
	if res.Completed30d != 1 || sum != 1 {
		t.Errorf("近30天完成应 1（sum=%d completed30d=%d）", sum, res.Completed30d)
	}
	// 均值 = (24h + 2h) / 2 = 13
	if res.AvgLeadHours != 13 {
		t.Errorf("平均周期应 13h, got %v", res.AvgLeadHours)
	}
	if len(res.ByProject) != 1 || res.ByProject[0].ProjectId != 501 || res.ByProject[0].Active != 4 {
		t.Errorf("项目分布不符: %+v", res.ByProject)
	}

	// 管理员（104）：无成员资格限制，502 的 review 任务计入
	resAdm, err := s.MyTaskStats(ctxAs(104), &api.MyTaskStatsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if resAdm.ActiveTotal != 1 || len(resAdm.ByProject) != 1 || resAdm.ByProject[0].ProjectId != 502 {
		t.Errorf("管理员统计不符: active=%d byProject=%+v", resAdm.ActiveTotal, resAdm.ByProject)
	}

	// 清理
	db.Model("tasks").Ctx(ctx).Where("title", "mx-stat").Delete()
	db.Model("projects").Ctx(ctx).Where("id", 502).Delete()
}

// ---------- 认领门禁收紧（#443：未接入拒绝 + 会话项目约束） ----------

func ctxAsAgent(uid, sessionProject int) context.Context {
	ctx := ctxAs(uid)
	if sessionProject > 0 {
		ctx = context.WithValue(ctx, "sessionProjectId", sessionProject)
	}
	return ctx
}

func TestClaimGates(t *testing.T) {
	ctx := context.Background()
	db := g.DB()

	// 未接入 agent：110 无 bindings 行也无 members 行
	if _, err := db.Model("sys_users").Ctx(ctx).Data(g.Map{
		"id": 110, "username": "mx-ai-orphan", "password": "x", "type": "ai", "status": 1,
	}).Insert(); err != nil {
		t.Fatal(err)
	}
	tid := newTask(t, 101)

	// ① 未接入项目的 agent：claim / 评论 / 关注全部拒绝（此前是无绑定=全能力放行）
	if err := s.ClaimTask(ctxAs(110), &api.TaskClaimReq{Id: tid}); err == nil || !strings.Contains(err.Error(), "未接入") {
		t.Errorf("未接入 agent 认领应拒（含文案）: %v", err)
	}
	if _, err := s.CreateComment(ctxAs(110), &api.CommentCreateReq{TaskId: tid, Content: "orphan"}); err == nil {
		t.Error("未接入 agent 评论应拒")
	}
	if err := s.WatchTask(ctxAs(110), tid); err == nil {
		t.Error("未接入 agent 关注应拒")
	}

	// ② 兼容存量：手工加进 members 的 agent（无 bindings）保持全能力
	if _, err := db.Model("project_members").Ctx(ctx).Data(g.Map{
		"project_id": 501, "user_id": 110, "role": "member",
	}).Insert(); err != nil {
		t.Fatal(err)
	}
	if err := s.ClaimTask(ctxAs(110), &api.TaskClaimReq{Id: tid}); err != nil {
		t.Errorf("members 存量 agent 认领应放行: %v", err)
	}
	if err := s.ReleaseTask(ctxAs(110), &api.TaskReleaseReq{Id: tid}); err != nil {
		t.Fatalf("释放回池失败: %v", err)
	}

	// ③ 会话项目不符：109 绑定 501（空能力=全能力），会话指向 502 → 拒
	if err := s.ClaimTask(ctxAsAgent(109, 502), &api.TaskClaimReq{Id: tid}); err == nil || !strings.Contains(err.Error(), "其它项目") {
		t.Errorf("跨会话项目认领应拒（含文案）: %v", err)
	}
	// ④ 会话项目一致 → 放行
	if err := s.ClaimTask(ctxAsAgent(109, 501), &api.TaskClaimReq{Id: tid}); err != nil {
		t.Errorf("会话项目一致认领应放行: %v", err)
	}
	if err := s.ReleaseTask(ctxAs(109), &api.TaskReleaseReq{Id: tid}); err != nil {
		t.Fatalf("释放回池失败: %v", err)
	}
	// ⑤ 无会话（直连 API）→ 仅准入门禁约束，正常放行
	if err := s.ClaimTask(ctxAs(109), &api.TaskClaimReq{Id: tid}); err != nil {
		t.Errorf("无会话直连认领应放行: %v", err)
	}
	if err := s.ReleaseTask(ctxAs(109), &api.TaskReleaseReq{Id: tid}); err != nil {
		t.Fatalf("释放回池失败: %v", err)
	}
	// ⑥ 人类不受会话约束（即便 ctx 混入会话值也直通）
	if err := s.ClaimTask(ctxAsAgent(101, 502), &api.TaskClaimReq{Id: tid}); err != nil {
		t.Errorf("人类认领应不受会话约束: %v", err)
	}

	// 清理：任务、110 的成员行与账号（本测试自建）
	db.Model("tasks").Ctx(ctx).Where("id", tid).Delete()
	db.Model("project_members").Ctx(ctx).Where("user_id", 110).Delete()
	db.Model("sys_users").Ctx(ctx).Where("id", 110).Delete()
}

// ---------- 发件侧反馈视图（#447） ----------

func TestFeedbackSent(t *testing.T) {
	ctx := context.Background()
	db := g.DB()
	ins := func(createdBy int, status, reason string) int {
		r, err := db.Model("project_feedbacks").Ctx(ctx).Data(g.Map{
			"project_id": 501, "source_project_id": 501,
			"title": "mx-sent", "content": "x", "status": status,
			"dismiss_reason": reason, "created_by": createdBy,
		}).Insert()
		if err != nil {
			t.Fatal(err)
		}
		id, _ := r.LastInsertId()
		return int(id)
	}
	id1 := ins(103, "open", "")
	id2 := ins(103, "dismissed", "mx-reason")
	ins(101, "open", "") // 他人发出的，不应出现（断言隐含在长度检查里）

	// 缺省 open：只见自己未处理的
	res, err := s.ListSentFeedbacks(ctxAs(103), &api.FeedbackSentReq{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.List) != 1 || res.List[0].Id != id1 {
		t.Fatalf("open 过滤应只见自己的未处理项, got %+v", res.List)
	}
	item := res.List[0]
	if item.TargetProjectId != 501 || item.TargetProjectName != "mx-proj" || item.Status != "open" {
		t.Errorf("字段回填不符: %+v", item)
	}

	// all：自己的两条（他人 id3 不出现）
	resAll, err := s.ListSentFeedbacks(ctxAs(103), &api.FeedbackSentReq{Status: "all"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resAll.List) != 2 {
		t.Fatalf("all 应见自己两条, got %d", len(resAll.List))
	}
	var dismissed *api.FeedbackSentItem
	for i := range resAll.List {
		if resAll.List[i].Id == id2 {
			dismissed = &resAll.List[i]
		}
	}
	if dismissed == nil || dismissed.DismissReason != "mx-reason" {
		t.Errorf("dismissed 项应带回忽略理由: %+v", dismissed)
	}

	db.Model("project_feedbacks").Ctx(ctx).Where("title", "mx-sent").Delete()
}

// ---------- 全局记忆提案（人人可发 + 管理员审核） ----------

func gmpNotifCount(t *testing.T, uid int, title string) int {
	t.Helper()
	n, err := g.DB().Model("notifications").Where("user_id", uid).Where("title", title).Count()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestGlobalMemoryProposal(t *testing.T) {
	ctx := context.Background()
	db := g.DB()

	// 占用：同名全局记忆已存在（任意状态）→ 拒
	if _, err := db.Model("project_memories").Ctx(ctx).Data(g.Map{
		"project_id": 0, "key": "mx-taken", "value": "v", "status": "active", "updated_by": 104,
	}).Insert(); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Docs().GlobalMemoryPropose(ctxAs(108), &apiDocs.GlobalMemoryProposeReq{Key: "mx-taken", Value: "x"}); err == nil || !strings.Contains(err.Error(), "已存在") {
		t.Errorf("同名占用应拒: %v", err)
	}

	// 只读 agent 也可提案（提案是低风险入口，风险由审核兜）
	pid, err := service.Docs().GlobalMemoryPropose(ctxAs(108), &apiDocs.GlobalMemoryProposeReq{
		Key: "mx-prop", Value: "mx-value", Note: "跨项目通用", Ttl: "7d",
	})
	if err != nil {
		t.Fatal(err)
	}
	// 同 key 未决防重
	if _, err := service.Docs().GlobalMemoryPropose(ctxAs(109), &apiDocs.GlobalMemoryProposeReq{Key: "mx-prop", Value: "y"}); err == nil || !strings.Contains(err.Error(), "审核中") {
		t.Errorf("未决重复应拒: %v", err)
	}

	// 管理员列表（缺省 submitted）含提交者名
	lres, err := service.Docs().GlobalMemoryProposalList(ctxAs(104), &apiDocs.GlobalMemoryProposalListReq{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, it := range lres.List {
		if it.Id == pid {
			found = true
			if it.ProposedByName != "mx-ai-ro" || it.Note != "跨项目通用" {
				t.Errorf("提案字段回填不符: %+v", it)
			}
		}
	}
	if !found {
		t.Fatalf("submitted 列表应含新提案 #%d", pid)
	}

	// 拒绝：理由必填
	if err := service.Docs().GlobalMemoryProposalReview(ctxAs(104), &apiDocs.GlobalMemoryProposalReviewReq{Id: pid, Decision: "rejected"}); err == nil || !strings.Contains(err.Error(), "理由") {
		t.Errorf("无理由拒绝应报错: %v", err)
	}
	// 拒绝（带理由）→ 状态 rejected + 提交者收拒绝通知
	if err := service.Docs().GlobalMemoryProposalReview(ctxAs(104), &apiDocs.GlobalMemoryProposalReviewReq{Id: pid, Decision: "rejected", Reason: "mx-no"}); err != nil {
		t.Fatal(err)
	}
	if n := gmpNotifCount(t, 108, "全局记忆提案被拒绝"); n != 1 {
		t.Errorf("提交者应收 1 条拒绝通知, got %d", n)
	}

	// 拒绝后同 key 可重新提案 → 采纳 → 正式记忆落库 + 通知
	pid2, err := service.Docs().GlobalMemoryPropose(ctxAs(108), &apiDocs.GlobalMemoryProposeReq{Key: "mx-prop", Value: "mx-value-2"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Docs().GlobalMemoryProposalReview(ctxAs(104), &apiDocs.GlobalMemoryProposalReviewReq{Id: pid2, Decision: "approved"}); err != nil {
		t.Fatal(err)
	}
	m, _ := db.Model("project_memories").Ctx(ctx).Where("project_id", 0).Where("key", "mx-prop").One()
	if m.IsEmpty() || m["value"].String() != "mx-value-2" || m["status"].String() != "active" {
		t.Errorf("采纳后应落 active 正式记忆: %+v", m)
	}
	if n := gmpNotifCount(t, 108, "全局记忆提案已采纳"); n != 1 {
		t.Errorf("提交者应收 1 条采纳通知, got %d", n)
	}
	// agent 可读全局记忆（MemGet(0)——#443 曾误伤 projectId=0 的读路径）
	gm, gerr := service.Docs().MemGet(ctxAs(108), 0, "mx-prop")
	if gerr != nil || gm.Value != "mx-value-2" {
		t.Errorf("agent 全局记忆读取应放行: err=%v value=%v", gerr, gm)
	}
	// 重复审核 → 拒
	if err := service.Docs().GlobalMemoryProposalReview(ctxAs(104), &apiDocs.GlobalMemoryProposalReviewReq{Id: pid2, Decision: "approved"}); err == nil || !strings.Contains(err.Error(), "已处理过") {
		t.Errorf("重复审核应报错: %v", err)
	}
	// 采纳后同名再提案 → 占用拒绝
	if _, err := service.Docs().GlobalMemoryPropose(ctxAs(109), &apiDocs.GlobalMemoryProposeReq{Key: "mx-prop", Value: "z"}); err == nil || !strings.Contains(err.Error(), "已存在") {
		t.Errorf("采纳后同名提案应拒: %v", err)
	}

	// 清理
	db.Model("global_memory_proposals").Ctx(ctx).Where("key IN (?)", []string{"mx-prop", "mx-taken"}).Delete()
	db.Model("project_memories").Ctx(ctx).Where("project_id", 0).Where("key IN (?)", []string{"mx-prop", "mx-taken"}).Delete()
	db.Model("notifications").Ctx(ctx).Where("user_id", 108).Where("title LIKE ?", "全局记忆提案%").Delete()
}

// ---------- 项目移交（邀请制：搜索选择 + 通知接受/拒绝） ----------

func ptNotifCount(t *testing.T, uid int, title string) int {
	t.Helper()
	n, err := g.DB().Model("notifications").Where("user_id", uid).Where("title", title).Count()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// timeRegExp 严格时间形态：抓 gtime 布局字面量（"2006-01-02 15:04:05"）
// 与微秒尾巴两类写法回归
var timeRegExp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)

func ptMemberRole(t *testing.T, pid, uid int) string {
	t.Helper()
	v, err := g.DB().Model("project_members").Where("project_id", pid).Where("user_id", uid).Fields("role").Value()
	if err != nil || v == nil {
		return ""
	}
	return v.String()
}

func TestOwnerTransferInvite(t *testing.T) {
	ctx := context.Background()
	db := g.DB()

	// 非 owner 发起 → 拒
	if _, err := s.InviteOwnerTransfer(ctxAs(102), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 103}); err == nil || !strings.Contains(err.Error(), "仅项目负责人") {
		t.Errorf("非 owner 发起应拒: %v", err)
	}
	// owner 邀请 103（留在项目）
	tid, err := s.InviteOwnerTransfer(ctxAs(101), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 103})
	if err != nil {
		t.Fatal(err)
	}
	if n := ptNotifCount(t, 103, "收到项目移交邀请"); n != 1 {
		t.Fatalf("受邀者应收邀请通知, got %d", n)
	}
	// 同项目 pending 防重
	if _, err := s.InviteOwnerTransfer(ctxAs(101), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 102}); err == nil || !strings.Contains(err.Error(), "待响应") {
		t.Errorf("pending 防重应拒: %v", err)
	}
	// 非受邀人响应 → 拒
	if err := s.RespondOwnerTransfer(ctxAs(102), &api.OwnerTransferRespondReq{Id: tid, Action: "accept"}); err == nil || !strings.Contains(err.Error(), "受邀本人") {
		t.Errorf("非受邀人响应应拒: %v", err)
	}
	// 拒绝 → 回告发起方
	if err := s.RespondOwnerTransfer(ctxAs(103), &api.OwnerTransferRespondReq{Id: tid, Action: "decline"}); err != nil {
		t.Fatal(err)
	}
	if n := ptNotifCount(t, 101, "移交邀请被拒绝"); n != 1 {
		t.Errorf("发起方应收拒绝通知, got %d", n)
	}
	if ptMemberRole(t, 501, 101) != "owner" {
		t.Error("拒绝后 owner 不应变化")
	}

	// 重新邀请 → 接受（留在项目：原 owner 降普通成员，受邀者升 owner）
	tid2, err := s.InviteOwnerTransfer(ctxAs(101), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 103})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RespondOwnerTransfer(ctxAs(103), &api.OwnerTransferRespondReq{Id: tid2, Action: "accept"}); err != nil {
		t.Fatal(err)
	}
	if ptMemberRole(t, 501, 103) != "owner" || ptMemberRole(t, 501, 101) != "member" {
		t.Errorf("接受后角色应变 103=owner/101=member, got %s/%s", ptMemberRole(t, 501, 103), ptMemberRole(t, 501, 101))
	}
	if n := ptNotifCount(t, 101, "移交邀请已接受"); n != 1 {
		t.Errorf("发起方应收接受通知, got %d", n)
	}
	// 列表回填 + 决议自动标读（#486）：受邀人视角该邀请通知应已读且
	// transferStatus=accepted——前端据此不再渲染接受/拒绝
	nres, nerr := service.Platform().ListNotifications(ctxAs(103), &platApi.NotificationListReq{Size: 50})
	if nerr != nil {
		t.Fatal(nerr)
	}
	for _, it := range nres.List {
		if it.SourceType == "transfer" && it.SourceId == tid2 {
			if it.TransferStatus != "accepted" {
				t.Errorf("已决邀请应回填 transferStatus=accepted, got %q", it.TransferStatus)
			}
			if it.IsRead != 1 {
				t.Errorf("决议后邀请通知应自动标读, got isRead=%d", it.IsRead)
			}
		}
	}
	// resolved_at 必须是 19 字符定长（YYYY-MM-DD HH:MM:SS）——gtime 对象
	// 会被驱动带微秒写入（26 字符），MySQL VARCHAR(19) 列直接 Data too
	// long（生产 #484；SQLite TEXT 存得下所以靠断言抓写法回归）
	if rv, _ := db.Model("project_transfers").Where("id", tid2).Fields("resolved_at").Value(); rv != nil {
		if v := rv.String(); !timeRegExp.MatchString(v) {
			t.Errorf("resolved_at 应为 YYYY-MM-DD HH:MM:SS 实际时间, got %q", v)
		}
	}
	// 重复响应 → 拒
	if err := s.RespondOwnerTransfer(ctxAs(103), &api.OwnerTransferRespondReq{Id: tid2, Action: "accept"}); err == nil || !strings.Contains(err.Error(), "已处理过") {
		t.Errorf("重复响应应拒: %v", err)
	}

	// leave 变体：超管 104（IsProjectOwner 直通）邀请 102，接受后原 owner(103) 移出
	tid3, err := s.InviteOwnerTransfer(ctxAs(104), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 102, Leave: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RespondOwnerTransfer(ctxAs(102), &api.OwnerTransferRespondReq{Id: tid3, Action: "accept"}); err != nil {
		t.Fatal(err)
	}
	if ptMemberRole(t, 501, 102) != "owner" || ptMemberRole(t, 501, 103) != "" {
		t.Errorf("leave 变体角色不符: 102=%s 103=%q", ptMemberRole(t, 501, 102), ptMemberRole(t, 501, 103))
	}

	// 清理：恢复 seed 成员关系 + 移交记录 + 通知
	db.Model("project_members").Ctx(ctx).Where("project_id", 501).Delete()
	for _, m := range []map[string]interface{}{
		{"project_id": 501, "user_id": 101, "role": "owner"},
		{"project_id": 501, "user_id": 102, "role": "maintainer"},
		{"project_id": 501, "user_id": 103, "role": "member"},
		{"project_id": 501, "user_id": 107, "role": "member"},
	} {
		db.Model("project_members").Ctx(ctx).Data(m).Insert()
	}
	db.Model("project_transfers").Ctx(ctx).Where("project_id", 501).Delete()
	db.Model("notifications").Ctx(ctx).Where("source_type", "transfer").Where("source_id IN (?)", []int{tid, tid2, tid3}).Delete()
	db.Model("notifications").Ctx(ctx).Where("title IN (?)", []string{"收到项目移交邀请", "移交邀请被拒绝", "移交邀请已接受"}).Delete()
}

// ---------- 分组归属化（创建者私有 + 移交自动退出原 owner 分组） ----------

func pgmCount(t *testing.T, groupId, projectId int) int {
	t.Helper()
	n, err := g.DB().Model("project_group_members").
		Where("group_id", groupId).Where("project_id", projectId).Count()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestGroupOwnership(t *testing.T) {
	ctx := context.Background()
	db := g.DB()

	// 101 建组 A，104（超管）建组 B
	gA, err := s.CreateGroup(ctxAs(101), &api.GroupCreateReq{Name: "mx-grp-a", Description: "owner101"})
	if err != nil {
		t.Fatal(err)
	}
	gB, err := s.CreateGroup(ctxAs(104), &api.GroupCreateReq{Name: "mx-grp-b", Description: "admin"})
	if err != nil {
		t.Fatal(err)
	}

	// 列表隔离：101 只见 A；104 见 A+B 且带归属名
	l101, _ := s.ListGroups(ctxAs(101))
	if len(l101.List) != 1 || l101.List[0].Id != gA || l101.List[0].OwnerName != "mx-owner" {
		t.Errorf("101 应只见自己的分组且带归属名: %+v", l101.List)
	}
	l104, _ := s.ListGroups(ctxAs(104))
	if len(l104.List) < 2 {
		t.Errorf("超管应见全部分组, got %d", len(l104.List))
	}

	// 更新：非创建者拒（102 是 501 maintainer 也无权动别人的分组）
	if err := s.UpdateGroup(ctxAs(102), &api.GroupUpdateReq{Id: gA, Description: ptrStr("hijack")}); err == nil || !strings.Contains(err.Error(), "创建者") {
		t.Errorf("非创建者更新应拒: %v", err)
	}
	// 创建者更新过；超管直通改他组也过（监管视角）
	if err := s.UpdateGroup(ctxAs(101), &api.GroupUpdateReq{Id: gA, Description: ptrStr("ok")}); err != nil {
		t.Errorf("创建者更新应过: %v", err)
	}
	if err := s.UpdateGroup(ctxAs(104), &api.GroupUpdateReq{Id: gA, Description: ptrStr("admin-touch")}); err != nil {
		t.Errorf("超管直通应过: %v", err)
	}
	// agent 建组拒
	if _, err := s.CreateGroup(ctxAs(109), &api.GroupCreateReq{Name: "mx-grp-agent"}); err == nil {
		t.Error("agent 建分组应拒")
	}

	// 加项目：102（501 maintainer，非分组归属者）把 501 挂进 A → 拒
	if err := s.AddProjectToGroup(ctxAs(102), &api.GroupMemberAddReq{Id: gA, ProjectId: 501}); err == nil || !strings.Contains(err.Error(), "创建者") {
		t.Errorf("非归属者加项目应拒: %v", err)
	}
	// 101（owner+归属者）挂 501 进 A → 过；104 挂 501 进 B → 过
	if err := s.AddProjectToGroup(ctxAs(101), &api.GroupMemberAddReq{Id: gA, ProjectId: 501}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddProjectToGroup(ctxAs(104), &api.GroupMemberAddReq{Id: gB, ProjectId: 501}); err != nil {
		t.Fatal(err)
	}
	if pgmCount(t, gA, 501) != 1 || pgmCount(t, gB, 501) != 1 {
		t.Fatal("挂载前置失败")
	}

	// 移交（邀请制）：101 → 103 接受后，501 从 A（101 名下）退出、B（104 名下）保留
	tid, err := s.InviteOwnerTransfer(ctxAs(101), &api.OwnerTransferInviteReq{ProjectId: 501, UserId: 103})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RespondOwnerTransfer(ctxAs(103), &api.OwnerTransferRespondReq{Id: tid, Action: "accept"}); err != nil {
		t.Fatal(err)
	}
	if n := pgmCount(t, gA, 501); n != 0 {
		t.Errorf("移交后应退出原 owner 分组 A, got %d", n)
	}
	if n := pgmCount(t, gB, 501); n != 1 {
		t.Errorf("第三方分组 B 应保留, got %d", n)
	}

	// 清理：恢复 501 owner、删分组与移交记录、清通知
	db.Model("project_members").Ctx(ctx).Where("project_id", 501).Delete()
	for _, m := range []map[string]interface{}{
		{"project_id": 501, "user_id": 101, "role": "owner"},
		{"project_id": 501, "user_id": 102, "role": "maintainer"},
		{"project_id": 501, "user_id": 103, "role": "member"},
		{"project_id": 501, "user_id": 107, "role": "member"},
	} {
		db.Model("project_members").Ctx(ctx).Data(m).Insert()
	}
	db.Model("project_groups").Ctx(ctx).Where("id IN (?)", []int{gA, gB}).Delete()
	db.Model("project_group_members").Ctx(ctx).Where("group_id IN (?)", []int{gA, gB}).Delete()
	db.Model("project_transfers").Ctx(ctx).Where("project_id", 501).Delete()
	db.Model("notifications").Ctx(ctx).Where("title LIKE ?", "移交邀请%").Delete()
	db.Model("notifications").Ctx(ctx).Where("title = ?", "移交邀请已接受").Delete()
}

func ptrStr(v string) *string { return &v }

// ---------- 项目列表归属筛选（scope：管理视角/个人视角） ----------

func TestProjectScope(t *testing.T) {
	// 管理员 104：all=全部（≥seed 的 501）；mine/owner=0（104 非任何项目成员）
	lall, err := s.ListProjects(ctxAs(104), &api.ProjectListReq{Size: 100})
	if err != nil {
		t.Fatal(err)
	}
	lmine, _ := s.ListProjects(ctxAs(104), &api.ProjectListReq{Size: 100, Scope: "mine"})
	lowner, _ := s.ListProjects(ctxAs(104), &api.ProjectListReq{Size: 100, Scope: "owner"})
	if lall.Total < 1 {
		t.Fatalf("管理员 all 应见平台项目, got %d", lall.Total)
	}
	if lmine.Total != 0 || lowner.Total != 0 {
		t.Errorf("管理员(非成员) mine/owner 应为 0, got %d/%d", lmine.Total, lowner.Total)
	}

	// 101（501 的 owner，非管理员）：all 也被收窄为成员可见；mine/owner 均见 501
	for _, sc := range []string{"", "all", "mine"} {
		lr, _ := s.ListProjects(ctxAs(101), &api.ProjectListReq{Size: 100, Scope: sc})
		if lr.Total != 1 || lr.List[0].Id != 501 {
			t.Errorf("101 scope=%q 应仅见 501, got total=%d", sc, lr.Total)
		}
	}
	lo, _ := s.ListProjects(ctxAs(101), &api.ProjectListReq{Size: 100, Scope: "owner"})
	if lo.Total != 1 || lo.List[0].Id != 501 {
		t.Errorf("101 owner 应见 501（其为 owner）, got %d", lo.Total)
	}
	// 103（501 的 member）：owner 视角为 0
	lo3, _ := s.ListProjects(ctxAs(103), &api.ProjectListReq{Size: 100, Scope: "owner"})
	if lo3.Total != 0 {
		t.Errorf("103 owner 应为 0（member 非 owner）, got %d", lo3.Total)
	}
	// ownerName 回填：501 现任 owner 是 101
	if lo.List[0].OwnerName != "mx-owner" || lo.List[0].OwnerId != 101 {
		t.Errorf("owner 回填不符: %+v", lo.List[0])
	}
}

// ---------- 测试执行记录（#504 pytest P1） ----------

func TestTestRunReport(t *testing.T) {
	ctx := context.Background()
	db := g.DB()

	mkReq := func() *apiTest.TestRunReportReq {
		return &apiTest.TestRunReportReq{
			ProjectId: 501,
			Source:   "pytest", Branch: "main", GitSha: "abc1234", Env: "ci",
			StartedAt: "2026-09-18 10:00:00", FinishedAt: "2026-09-18 10:01:30",
			Cases: []apiTest.TestRunCaseReport{
				{ExternalKey: "tests/test_a.py::test_ok", Status: "pass", DurationMs: 120},
				{ExternalKey: "tests/test_a.py::test_bad", Status: "fail", DurationMs: 30, Message: strings.Repeat("x", 9000)},
				{ExternalKey: "tests/test_b.py::test_skip", Status: "skip"},
				{ExternalKey: "tests/test_b.py::test_err", Status: "error", Message: "fixture boom", TestCaseId: 999999},
			},
		}
	}

	// ① 只读能力 agent（108：tasks_read,docs_read）上报 → 能力拒绝
	if _, err := service.Test().ReportRun(ctxAs(108), mkReq()); err == nil || !strings.Contains(err.Error(), "上报测试执行") {
		t.Errorf("缺 test_execute 能力应拒（含文案）: %v", err)
	}
	// ② 会话项目不符（109 会话指向 502）→ 拒
	if _, err := service.Test().ReportRun(ctxAsAgent(109, 502), mkReq()); err == nil || !strings.Contains(err.Error(), "其它项目") {
		t.Errorf("跨会话项目上报应拒（含文案）: %v", err)
	}

	// ③ 全能力 agent（109 空能力集）上报 → 汇总服务端重算 + 时间推导 + 截断
	runId, err := service.Test().ReportRun(ctxAsAgent(109, 501), mkReq())
	if err != nil {
		t.Fatalf("全能力 agent 上报应放行: %v", err)
	}
	run, err := db.Model("test_runs").Ctx(ctx).Where("id", runId).One()
	if err != nil || run.IsEmpty() {
		t.Fatalf("查询 run 失败: %v", err)
	}
	for col, want := range map[string]int{"total": 4, "passed": 1, "failed": 1, "skipped": 1, "errors": 1, "duration_ms": 90000, "triggered_by": 109} {
		if run[col].Int() != want {
			t.Errorf("%s = %d, want %d", col, run[col].Int(), want)
		}
	}
	if run["started_at"].String() != "2026-09-18 10:00:00" || run["finished_at"].String() != "2026-09-18 10:01:30" {
		t.Errorf("起止时间应透传客户端值: %s ~ %s", run["started_at"], run["finished_at"])
	}
	det, err := service.Test().GetRun(ctx, runId)
	if err != nil || len(det.Cases) != 4 {
		t.Fatalf("详情应含 4 条用例: %v", err)
	}
	// message 截断：8000 字符 + 截断标记；无效映射（999999 不存在）降级为 0
	for _, c := range det.Cases {
		if c.Status == "fail" {
			if l := len(c.Message); l > 8200 || !strings.HasSuffix(c.Message, "[truncated]") {
				t.Errorf("失败信息应截断（len=%d）: %q...", l, c.Message[:40])
			}
		}
		if c.ExternalKey == "tests/test_b.py::test_err" && c.TestCaseId != 0 {
			t.Errorf("无效映射应降级为 0: %+v", c)
		}
	}

	// ④ 人类成员（103）手工批次上报 → 放行
	hreq := mkReq()
	hreq.Source = "manual"
	hreq.Branch, hreq.GitSha, hreq.Env = "", "", ""
	if _, err := service.Test().ReportRun(ctxAs(103), hreq); err != nil {
		t.Errorf("人类成员上报应放行: %v", err)
	}

	// ⑤ 列表过滤：status=fail 命中两批（pytest 批含 fail/error，manual 批同构造）
	total, list, err := service.Test().ListRuns(ctx, &apiTest.TestRunListReq{ProjectId: 501, Status: "fail"})
	if err != nil || total != 2 {
		t.Errorf("fail 过滤应命中 2 批: total=%d err=%v", total, err)
	}
	if len(list) > 0 && list[0].TriggeredByName == "" {
		t.Errorf("triggeredByName 应回填: %+v", list[0])
	}

	// ⑥ 删除：member 拒 / maintainer 放行并级联清 cases
	if err := service.Test().DeleteRun(ctxAs(103), runId); err == nil || !strings.Contains(err.Error(), "仅项目管理员") {
		t.Errorf("member 删除应拒（含文案）: %v", err)
	}
	if err := service.Test().DeleteRun(ctxAs(102), runId); err != nil {
		t.Errorf("maintainer 删除应放行: %v", err)
	}
	if cnt, _ := db.Model("test_run_cases").Ctx(ctx).Where("test_run_id", runId).Count(); cnt != 0 {
		t.Errorf("run 用例应级联删除，残留 %d", cnt)
	}

	// ⑦ externalKey 精确查找（--bcode-sync 幂等依据）
	if _, err := db.Model("test_cases").Ctx(ctx).Data(g.Map{
		"project_id": 501, "title": "synced", "priority": "P2", "status": "active",
		"creator_id": 101, "external_key": "tests/test_a.py::test_ok",
	}).Insert(); err != nil {
		t.Fatal(err)
	}
	tcTotal, tcList, err := service.Test().ListCases(ctx, &apiTest.TestCaseListReq{ProjectId: 501, ExternalKey: "tests/test_a.py::test_ok"})
	if err != nil || tcTotal != 1 || len(tcList) != 1 || tcList[0].Title != "synced" {
		t.Errorf("externalKey 查找应精确命中 1 条: total=%d err=%v", tcTotal, err)
	}
	// priority 往返必须是 P 前缀字符串：gdb 曾按 INTEGER 列型把 'P2' gconv 成 0
	// （迁移 78 重建列型修复；MySQL 侧对应 Error 1366）
	if tcList[0].Priority != "P2" {
		t.Errorf("priority 往返失真: got %q want \"P2\"", tcList[0].Priority)
	}

	// 清理（501 是共享种子项目）
	db.Model("test_runs").Ctx(ctx).Where("project_id", 501).Delete()
	db.Model("test_run_cases").Ctx(ctx).Where("test_run_id NOT IN (SELECT id FROM test_runs)").Delete()
	db.Model("test_cases").Ctx(ctx).Where("title", "synced").Delete()
}

// ---------- 失败闭环 / Flaky / 趋势（#506） ----------

func TestTestRunClosedLoop(t *testing.T) {
	ctx := context.Background()
	db := g.DB()
	mk := func(k, status string) apiTest.TestRunCaseReport {
		return apiTest.TestRunCaseReport{ExternalKey: k, Status: status, Message: "boom-trace"}
	}
	// 三次执行：unstable 键 pass/fail 交替（flaky），stable 键恒 pass，bad 键恒 fail
	report := func() *apiTest.TestRunReportReq {
		return &apiTest.TestRunReportReq{ProjectId: 501, Source: "pytest"}
	}
	r1 := report(); r1.Cases = []apiTest.TestRunCaseReport{mk("t::unstable", "pass"), mk("t::stable", "pass"), mk("t::bad", "fail")}
	r2 := report(); r2.Cases = []apiTest.TestRunCaseReport{mk("t::unstable", "fail"), mk("t::stable", "pass"), mk("t::bad", "fail")}
	r3 := report(); r3.Cases = []apiTest.TestRunCaseReport{mk("t::unstable", "pass"), mk("t::stable", "pass"), mk("t::bad", "fail")}
	for _, r := range []*apiTest.TestRunReportReq{r1, r2, r3} {
		if _, err := service.Test().ReportRun(ctxAs(109), r); err != nil {
			t.Fatalf("seed run 失败: %v", err)
		}
	}

	// ① Flaky 列表：unstable 在列（1P/1F 至少），stable/bad 不在
	fl, err := service.Test().ListFlaky(ctx, &apiTest.TestFlakyListReq{ProjectId: 501})
	if err != nil {
		t.Fatalf("ListFlaky: %v", err)
	}
	var uni *apiTest.TestFlakyItem
	for i := range fl.List {
		if fl.List[i].ExternalKey == "t::unstable" {
			uni = &fl.List[i]
		}
		if fl.List[i].ExternalKey == "t::stable" || fl.List[i].ExternalKey == "t::bad" {
			t.Errorf("非抖动用例不应进 Flaky: %+v", fl.List[i])
		}
	}
	if uni == nil {
		t.Fatalf("unstable 应在 Flaky 列表: %+v", fl.List)
	}
	if uni.PassCount < 1 || uni.FailCount < 1 {
		t.Errorf("flaky 计数不符: %+v", uni)
	}

	// ② 详情富化：unstable 行带 flaky 标；bad 行无
	_, runList, _ := service.Test().ListRuns(ctx, &apiTest.TestRunListReq{ProjectId: 501, Source: "pytest"})
	if len(runList) == 0 {
		t.Fatal("应能列出 seeded runs")
	}
	lastId := runList[0].Id // id 倒序，第一行为最新
	det, err := service.Test().GetRun(ctx, lastId)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	for _, c := range det.Cases {
		if c.ExternalKey == "t::unstable" && !c.Flaky {
			t.Errorf("unstable 行应带 flaky 标: %+v", c)
		}
		if c.ExternalKey == "t::bad" && c.Flaky {
			t.Errorf("恒败用例不应误标 flaky: %+v", c)
		}
	}

	// 找 bad 键的执行用例 id（转缺陷对象）
	var badCaseId int
	for _, c := range det.Cases {
		if c.ExternalKey == "t::bad" {
			badCaseId = c.Id
		}
	}
	if badCaseId == 0 {
		t.Fatal("找不到 bad 用例行")
	}

	// ③ 转缺陷门禁：只读 agent（108）拒绝（走 tasks_write）
	if _, err := service.Test().CaseToBug(ctxAs(108), &apiTest.TestRunCaseBugReq{Id: badCaseId}); err == nil || !strings.Contains(err.Error(), "写任务") {
		t.Errorf("只读 agent 转缺陷应拒（含文案）: %v", err)
	}
	// ④ 人类成员转缺陷：建 bug 任务 + 挂钩 + 广播绑定 agent
	bg, err := service.Test().CaseToBug(ctxAs(103), &apiTest.TestRunCaseBugReq{Id: badCaseId, Title: "修复 bad"})
	if err != nil {
		t.Fatalf("member 转缺陷应放行: %v", err)
	}
	if !bg.Created || bg.TaskId == 0 {
		t.Fatalf("应新建任务: %+v", bg)
	}
	taskRow, _ := db.Model("tasks").Ctx(ctx).Where("id", bg.TaskId).One()
	if taskRow.IsEmpty() || taskRow["type"].String() != "bug" || taskRow["project_id"].Int() != 501 {
		t.Fatalf("任务形状不符: %+v", taskRow)
	}
	if !strings.Contains(taskRow["description"].String(), "boom-trace") {
		t.Errorf("任务描述应含失败信息: %q", taskRow["description"].String())
	}
	if cnt, _ := db.Model("notifications").Ctx(ctx).Where("source_type", "task").Where("source_id", bg.TaskId).Count(); cnt == 0 {
		t.Errorf("绑定 agent 应收到可认领广播（#109/#108 已绑定 501）")
	}
	linkRow, _ := db.Model("test_run_cases").Ctx(ctx).Where("id", badCaseId).One()
	if linkRow["bug_task_id"].Int() != bg.TaskId {
		t.Errorf("bug_task_id 应回填: %+v", linkRow)
	}
	// 重复转 → 明确拒绝
	if _, err := service.Test().CaseToBug(ctxAs(103), &apiTest.TestRunCaseBugReq{Id: badCaseId}); err == nil || !strings.Contains(err.Error(), "已挂接") {
		t.Errorf("重复转缺陷应拒（含文案）: %v", err)
	}

	// ⑤ 挂接既有任务：unstable 失败行 + 刚建的任务跨项目校验用新任务
	var uniCaseId int
	for _, c := range det.Cases {
		if c.ExternalKey == "t::unstable" {
			uniCaseId = c.Id
		}
	}
	// 跨项目任务：项目 502 建一个任务再挂 → 拒
	db.Model("projects").Ctx(ctx).Data(g.Map{"id": 502, "name": "mx-proj2", "status": 1, "creator_id": 101}).Insert()
	foreign, _ := db.Model("tasks").Ctx(ctx).Data(g.Map{"project_id": 502, "title": "foreign", "status": "open", "creator_id": 101}).Insert()
	fid, _ := foreign.LastInsertId()
	if _, err := service.Test().CaseToBug(ctxAs(103), &apiTest.TestRunCaseBugReq{Id: uniCaseId, TaskId: int(fid)}); err == nil || !strings.Contains(err.Error(), "不属于该项目") {
		t.Errorf("跨项目挂接应拒（含文案）: %v", err)
	}
	own, _ := db.Model("tasks").Ctx(ctx).Data(g.Map{"project_id": 501, "title": "own-bug", "status": "open", "creator_id": 101, "type": "bug"}).Insert()
	oid, _ := own.LastInsertId()
	lg, err := service.Test().CaseToBug(ctxAs(103), &apiTest.TestRunCaseBugReq{Id: uniCaseId, TaskId: int(oid)})
	if err != nil || lg.Created || lg.TaskId != int(oid) {
		t.Errorf("挂接既有任务应放行且不新建: %+v err=%v", lg, err)
	}

	// ⑥ 趋势：runs 时间升序 + topFailed 含 bad（3 败）与 unstable（窗口内 1 败）
	tr, err := service.Test().ListTrends(ctx, &apiTest.TestTrendReq{ProjectId: 501})
	if err != nil {
		t.Fatalf("ListTrends: %v", err)
	}
	if len(tr.Runs) < 3 {
		t.Errorf("趋势应含 3+ runs: %d", len(tr.Runs))
	}
	topMap := map[string]apiTest.TestFailTop{}
	for _, tf := range tr.TopFailed {
		topMap[tf.ExternalKey] = tf
	}
	if topMap["t::bad"].FailCount != 3 || topMap["t::bad"].TotalCount != 3 {
		t.Errorf("bad 失败 Top 统计不符: %+v", topMap["t::bad"])
	}
	if topMap["t::unstable"].FailCount != 1 {
		t.Errorf("unstable 失败计数应 1: %+v", topMap["t::unstable"])
	}

	// 清理（501 共享 + 本测试自建 502）
	db.Exec(ctx, "DELETE FROM notifications WHERE source_type = 'task' AND source_id IN (SELECT id FROM tasks WHERE title IN ('修复 bad','own-bug'))")
	db.Model("test_runs").Ctx(ctx).Where("project_id", 501).Delete()
	db.Model("test_run_cases").Ctx(ctx).Where("test_run_id NOT IN (SELECT id FROM test_runs)").Delete()
	db.Model("tasks").Ctx(ctx).Where("project_id IN (501,502)").Where("creator_id", 101).Delete()
	db.Model("tasks").Ctx(ctx).Where("id", bg.TaskId).Delete()
	db.Model("projects").Ctx(ctx).Where("id", 502).Delete()
}

// ---------- 上报幂等键 + 执行用例附件（#527） ----------

func TestTestRunIdempotency(t *testing.T) {
	ctx := context.Background()
	db := g.DB()
	mk := func(key string) *apiTest.TestRunReportReq {
		return &apiTest.TestRunReportReq{
			ProjectId: 501, Source: "pytest", IdempotencyKey: key,
			Cases: []apiTest.TestRunCaseReport{{ExternalKey: "t::idem", Status: "pass"}},
		}
	}

	// ① 同键两次：第二次命中既有 run（duplicate=true 且不新建）
	r1 := mk("idem-key-1")
	id1, err := service.Test().ReportRun(ctxAs(109), r1)
	if err != nil {
		t.Fatalf("首次上报失败: %v", err)
	}
	_, err = service.Test().ReportRun(ctxAs(109), mk("idem-key-1"))
	var hit *service.IdempotentHitError
	if !errors.As(err, &hit) || hit.RunId != id1 {
		t.Fatalf("同键重报应命中既有 #%d: %v", id1, err)
	}
	if cnt, _ := db.Model("test_runs").Ctx(ctx).Where("idempotency_key", "idem-key-1").Count(); cnt != 1 {
		t.Errorf("同键只应有一行: %d", cnt)
	}
	// ② 异键正常新建；③ 空键每次新建（NULL 不参与唯一约束）
	id2, err := service.Test().ReportRun(ctxAs(109), mk("idem-key-2"))
	if err != nil || id2 == id1 {
		t.Errorf("异键应新建: %d %v", id2, err)
	}
	a, _ := service.Test().ReportRun(ctxAs(109), mk(""))
	b, _ := service.Test().ReportRun(ctxAs(109), mk(""))
	if a == b {
		t.Errorf("空键不应去重: %d == %d", a, b)
	}

	// ④ 附件实体门禁：test_run_case 两跳解析——成员可挂、无关用户拒
	trcId := 0
	if v, _ := db.Model("test_run_cases").Ctx(ctx).Where("test_run_id", id1).Fields("id").Value(); v != nil {
		trcId = v.Int()
	}
	if trcId == 0 {
		t.Fatal("找不到 run 用例行")
	}
	if !attachmentAccessibleForTest(ctxAs(103), "test_run_case", trcId) {
		t.Error("项目成员应可向执行用例挂附件")
	}
	if attachmentAccessibleForTest(ctxAs(106), "test_run_case", trcId) {
		t.Error("非项目成员应被拒（106 无任何项目关系）")
	}

	// ⑤ 上报压测防退化（间歇性失败的观察项锚定）：串行 20 次零业务错
	for i := 0; i < 20; i++ {
		if _, err := service.Test().ReportRun(ctxAs(109), mk("")); err != nil {
			t.Fatalf("压测第 %d 次上报失败: %v", i, err)
		}
	}

	db.Model("test_runs").Ctx(ctx).Where("project_id", 501).Delete()
	db.Model("test_run_cases").Ctx(ctx).Where("test_run_id NOT IN (SELECT id FROM test_runs)").Delete()
}

// 附件门禁直测：logic/attachment 已导出 AttachmentEntityAccessible（纯判断可直调）
func attachmentAccessibleForTest(ctx context.Context, entityType string, entityId int) bool {
	return attachment.AttachmentEntityAccessible(ctx, perm.UserId(ctx), entityType, entityId, "")
}
