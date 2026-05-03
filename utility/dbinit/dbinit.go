package dbinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
)

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
	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		if executedMap[filename] {
			continue
		}
		content := gfile.GetContents(filepath.Join(sqlDir, filename))
		if content == "" {
			continue
		}

		g.Log().Infof(ctx, "Executing migration: %s", filename)
		_, err = db.Exec(ctx, content)
		if err != nil {
			// SQLite ALTER TABLE 不支持 IF NOT EXISTS 的某些情况，忽略重复列错误
			if strings.Contains(err.Error(), "duplicate column name") {
				g.Log().Warningf(ctx, "Migration %s: column already exists, skipping", filename)
			} else {
				return fmt.Errorf("execute migration %s: %w", filename, err)
			}
		}

		_, err = db.Exec(ctx, "INSERT INTO _migrations (filename) VALUES (?)", filename)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", filename, err)
		}
		g.Log().Infof(ctx, "Migration completed: %s", filename)
	}

	return nil
}
