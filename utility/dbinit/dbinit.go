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

// acquireInstanceLock 单实例启动锁：双实例同库并发迁移时，后启动进程会从 01
// 重放并在 _migrations UNIQUE 冲突上整体崩溃（事务保证了数据不坏，但表现为
// “第二个实例神秘退出”）。锁文件生命周期与进程一致（句柄持有，进程退出自动释放）
func acquireInstanceLock(ctx context.Context) (*os.File, error) {
	lockPath := "resource/data/app.lock"
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
	if err != nil {
		if os.IsExist(err) {
			// 尝试探测陈旧锁：Windows 上进程死亡后锁文件残留但无法以排他方式重开。
			// 简化处理：尝试打开写入探测——被占用说明进程活着，否则视为陈旧删除重试
			probe, perr := os.OpenFile(lockPath, os.O_RDWR, 0o644)
			if perr != nil {
				return nil, fmt.Errorf("另一实例正在运行（%s）；如确认无实例请删除该文件后重启", lockPath)
			}
			probe.Close()
			// 能打开：文件未被占用，视为陈旧锁（上次进程未清理）
			_ = os.Remove(lockPath)
			f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
			if err != nil {
				return nil, fmt.Errorf("获取启动锁失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("获取启动锁失败: %w", err)
		}
	}
	return f, nil
}

func AutoMigrate(ctx context.Context) error {
	lock, err := acquireInstanceLock(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = lock.Close()
		_ = os.Remove("resource/data/app.lock")
	}()
	db := g.DB()

	// 创建 migrations 跟踪表
	_, err2 := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS _migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filename TEXT NOT NULL UNIQUE,
		executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err2 != nil {
		return fmt.Errorf("create _migrations table: %w", err2)
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

	// 迁移号冲突检测：同号多文件（并行开发常见）静默按字典序执行是流程性风险
	nums := map[int]string{}
	for _, name := range sqlFiles {
		n := migrationNumber(name)
		if prev, dup := nums[n]; dup && n != math.MaxInt32 {
			return fmt.Errorf("迁移号冲突: %d 同时被 %s 与 %s 使用", n, prev, name)
		}
		nums[n] = name
	}

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
