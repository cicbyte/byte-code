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
	"strings"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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
