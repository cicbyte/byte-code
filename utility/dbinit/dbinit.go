package dbinit

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

// migrationPrefixRe 提取文件名前缀数字；
// 必须按数字排序执行：字典序会把 "100_x.sql" 排在 "40_x.sql" 之前导致顺序错乱
var migrationPrefixRe = regexp.MustCompile(`^(\d+)_`)

func AutoMigrate(ctx context.Context) error {
	db := g.DB()

	// 创建 migrations 跟踪表
	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS _migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL UNIQUE,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create _migrations table: %w", err)
	}

	// 获取已执行的文件列表
	var executed []struct{ Filename string }
	err = db.Model("_migrations").Ctx(ctx).Scan(&executed)
	if err != nil {
		return err
	}
	executedMap := make(map[string]bool)
	for _, e := range executed {
		executedMap[e.Filename] = true
	}

	// 读取 SQL 目录
	sqlDir := "resource/sql/sqlite"
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		return fmt.Errorf("read sql dir: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Slice(sqlFiles, func(i, j int) bool {
		ni, nj := migrationNumber(sqlFiles[i]), migrationNumber(sqlFiles[j])
		if ni != nj {
			return ni < nj
		}
		return sqlFiles[i] < sqlFiles[j]
	})

	for _, filename := range sqlFiles {
		if executedMap[filename] {
			continue
		}
		content := gfile.GetContents(filepath.Join(sqlDir, filename))
		if content == "" {
			continue
		}

		g.Log().Infof(ctx, "Executing migration: %s", filename)
		// 单文件整体事务：SQL 与 _migrations 记录同生共死，中途失败整体回滚
		// 且不记录执行，下次启动可完整重试，不会留下半成品 schema 导致启动死循环。
		// 注意：迁移文件内不要再写 BEGIN/COMMIT 或 PRAGMA（PRAGMA 在事务内是空操作）
		err = db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			if _, err := tx.Exec(content); err != nil {
				return err
			}
			if _, err := tx.Exec("INSERT INTO _migrations (filename) VALUES (?)", filename); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("execute migration %s: %w", filename, err)
		}
		g.Log().Infof(ctx, "Migration completed: %s", filename)
	}

	return nil
}

// migrationNumber 取文件名前缀数字；无数字前缀的文件排在最后执行
func migrationNumber(name string) int {
	if m := migrationPrefixRe.FindStringSubmatch(name); m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return math.MaxInt32
}
