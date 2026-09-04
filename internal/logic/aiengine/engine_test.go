package aiengine

// 引擎启停状态机测试：P0-4 的 bug（sync.Once 导致停后无法重启）正是
// "启停启"断言能当场抓住的。DB 走临时 SQLite（启停持久化写 sys_config）

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func setupEngineTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "none.yaml")
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
	gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{
		gdb.ConfigNode{Type: "sqlite", Link: "sqlite::@file(" + filepath.Join(dir, "t.db") + ")"},
	}})
	_, err := g.DB().Exec(context.Background(), `CREATE TABLE sys_config (key TEXT PRIMARY KEY, value TEXT DEFAULT '')`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEngineToggleCycles(t *testing.T) {
	setupEngineTest(t)
	ctx := context.Background()

	// 启停启两轮：旧 sync.Once 实现在第二次 start 后 cron 不再注册（状态却显示运行）
	StartEngine(ctx)
	if !IsRunning() {
		t.Fatal("start 后应运行")
	}
	StopEngine(ctx)
	if IsRunning() {
		t.Fatal("stop 后应停止")
	}
	StartEngine(ctx)
	if !IsRunning() {
		t.Fatal("第二次 start 后应运行（P0-4 回归断言）")
	}
	StopEngine(ctx)
	StartEngine(ctx)
	StopEngine(ctx)

	// 幂等：重复 start/stop 不 panic 不变状态
	StartEngine(ctx)
	StartEngine(ctx)
	if !IsRunning() {
		t.Fatal("重复 start 应保持运行")
	}
	StopEngine(ctx)
	StopEngine(ctx)
	if IsRunning() {
		t.Fatal("重复 stop 应保持停止")
	}

	// 持久化：最后一次 stop 落库为 0
	v, _ := g.DB().Model("sys_config").Ctx(ctx).Where("key", enabledConfigKey).Value("value")
	if v.String() != "0" {
		t.Fatalf("stop 后 enabled 应为 0，实际 %q", v.String())
	}
}
