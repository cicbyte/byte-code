package dbinit

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 时区统一后处理（SQLite 专属） ====================
//
// 统一口径：全库时间列一律本地时间字符串（YYYY-MM-DD HH:MM:SS）。
// 触发器改写走迁移 64；以下两件事无法在迁移框架内做（单文件事务里
// PRAGMA writable_schema 是空操作），故在迁移完成后以 autocommit 执行：
//
//  1. DEFAULT 改写：把建表 SQL 里的 CURRENT_TIMESTAMP 默认值替换为
//     datetime('now','localtime')（PRAGMA writable_schema 直改 sqlite_master，
//     官方支持的 schema 修改通道，零表重建）。幂等自愈：仍有残留才执行。
//  2. 存量修正：tasks.updated_at / completed_at 里「无 +hh:mm 后缀」的
//     纯 UTC 值（触发器/默认值产出）+8h 转本地——切换瞬间若不修，租约
//     扫描（本地阈值 vs UTC 值差 8h）会把刚认领的任务全体误释放。
//     仅此两列（时间比较的真实消费方）；其余表历史值偏移明确接受
//     （仅显示晚 8h，无逻辑消费）。幂等标记 sys_config tz_unified_local=1。

func sqlitePostTimezone(ctx context.Context) error {
	// 幂等闸门：标记已置则跳过（writable_schema 扫描便宜，但存量修正必须只跑一次）
	if v, _ := g.DB().Model("sys_config").Ctx(ctx).Where("`key`", "tz_unified_local").Value("value"); v != nil && v.String() == "1" {
		return nil
	}

	// ---- 1. DEFAULT 改写（幂等：无残留则不动）----
	rows, err := g.DB().GetAll(ctx,
		"SELECT name, sql FROM sqlite_master WHERE type='table' AND sql LIKE '%CURRENT_TIMESTAMP%' AND name NOT LIKE 'sqlite_%' AND name != '_migrations'")
	if err != nil {
		return fmt.Errorf("timezone: scan sqlite_master failed: %w", err)
	}
	if len(rows) > 0 {
		if _, err := g.DB().Exec(ctx, "PRAGMA writable_schema=ON"); err != nil {
			return fmt.Errorf("timezone: writable_schema on failed: %w", err)
		}
		for _, r := range rows {
			newSQL := strings.ReplaceAll(r["sql"].String(), "DEFAULT CURRENT_TIMESTAMP", "DEFAULT (datetime('now','localtime'))")
			if _, err := g.DB().Exec(ctx,
				"UPDATE sqlite_master SET sql = ? WHERE type='table' AND name = ?", newSQL, r["name"].String()); err != nil {
				_, _ = g.DB().Exec(ctx, "PRAGMA writable_schema=OFF")
				return fmt.Errorf("timezone: rewrite %s failed: %w", r["name"].String(), err)
			}
		}
		if _, err := g.DB().Exec(ctx, "PRAGMA writable_schema=OFF"); err != nil {
			return fmt.Errorf("timezone: writable_schema off failed: %w", err)
		}
		// schema 改写后官方建议 VACUUM 使其稳固（库大时可耗时，接受一次性成本）
		if _, err := g.DB().Exec(ctx, "VACUUM"); err != nil {
			return fmt.Errorf("timezone: vacuum failed: %w", err)
		}
		g.Log().Infof(ctx, "timezone: rewrote CURRENT_TIMESTAMP defaults on %d tables", len(rows))
	}

	// ---- 2. 存量修正（仅租约消费列；条件「无 + 后缀」区分 app 写入的本地值）----
	for _, col := range []string{"updated_at", "completed_at"} {
		res, err := g.DB().Exec(ctx, fmt.Sprintf(
			"UPDATE tasks SET %s = datetime(%s, 'localtime') WHERE %s != '' AND %s NOT LIKE '%%+%%'", col, col, col, col))
		if err != nil {
			return fmt.Errorf("timezone: fix tasks.%s failed: %w", col, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			g.Log().Infof(ctx, "timezone: shifted %d rows in tasks.%s (UTC → local)", n, col)
		}
	}

	// ---- 3. 幂等标记（复用 dbutil 的方言 upsert 语义，此处直写避免包依赖环）----
	if _, err := g.DB().Exec(ctx,
		"INSERT INTO sys_config (`key`, `value`) VALUES ('tz_unified_local', '1') ON CONFLICT(`key`) DO UPDATE SET `value`='1'"); err != nil {
		return fmt.Errorf("timezone: mark failed: %w", err)
	}
	g.Log().Info(ctx, "timezone: unified to local-time convention")
	return nil
}
