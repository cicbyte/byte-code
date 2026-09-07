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

	dialect := Dialect()

	// 创建 migrations 跟踪表（按方言取合法 DDL）
	trackDDL := `CREATE TABLE IF NOT EXISTS _migrations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	filename TEXT NOT NULL UNIQUE,
	executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`
	if dialect == "mysql" {
		trackDDL = "CREATE TABLE IF NOT EXISTS `_migrations` (\n" +
			"\t`id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,\n" +
			"\t`filename` VARCHAR(255) NOT NULL UNIQUE,\n" +
			"\t`executed_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP\n" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"
	}
	if _, err2 := db.Exec(ctx, trackDDL); err2 != nil {
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

	// 迁移目录按方言区分：sqlite / mysql 双轨维护
	sqlDir := filepath.Join("resource", "sql", dialect)
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
		// SQLite 驱动接受多语句整发；MySQL 驱动默认拒绝，切分后逐条执行
		// （迁移禁用存储过程/触发器，无需处理 DELIMITER）
		var stmts []string
		if dialect == "mysql" {
			stmts = splitSQLStatements(content)
		} else {
			stmts = []string{content}
		}
		err = db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			for _, st := range stmts {
				if _, err := tx.Exec(st); err != nil {
					return err
				}
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

// splitSQLStatements 按分号切分 SQL 文件为独立语句（MySQL 驱动需要逐条执行）。
// 正确跳过：单/双/反引号字符串、-- 行注释、/* */ 块注释；切分后去注释与空白。
func splitSQLStatements(content string) []string {
	var (
		out []string
		buf strings.Builder
		i   int
	)
	flush := func() {
		s := strings.TrimSpace(buf.String())
		if s != "" {
			out = append(out, s)
		}
		buf.Reset()
	}
	for i < len(content) {
		c := content[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			// 引号串原样吞（含 '' 转义）
			quote := c
			buf.WriteByte(c)
			i++
			for i < len(content) {
				buf.WriteByte(content[i])
				if content[i] == quote {
					if i+1 < len(content) && content[i+1] == quote {
						buf.WriteByte(content[i+1])
						i++
					} else {
						i++
						break
					}
				}
				i++
			}
		case c == '-' && i+1 < len(content) && content[i+1] == '-':
			// 行注释丢弃
			for i < len(content) && content[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(content) && content[i+1] == '*':
			// 块注释丢弃
			i += 2
			for i+1 < len(content) && !(content[i] == '*' && content[i+1] == '/') {
				i++
			}
			i += 2
		case c == ';':
			flush()
			i++
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return out
}

// Dialect 取当前配置的数据库方言（"sqlite" | "mysql"）。
// GoFrame 的 link 形如 "mysql:user:pass@tcp(...)/db" / "sqlite::@file(path)"，
// 以前缀区分；无 link 配置时兜底 sqlite
func Dialect() string {
	var link string
	cfg := g.DB().GetConfig()
	if cfg != nil && cfg.Link != "" {
		link = cfg.Link
	} else if cfg != nil && cfg.Type != "" {
		link = cfg.Type + ":"
	}
	link = strings.ToLower(strings.TrimSpace(link))
	if strings.HasPrefix(link, "mysql") {
		return "mysql"
	}
	return "sqlite"
}
