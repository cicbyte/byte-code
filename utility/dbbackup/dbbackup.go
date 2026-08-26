package dbbackup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	backupDir      = "resource/backup" // 备份目录（.gitignore 忽略）
	retainBackups  = 7                 // 保留最近份数
)

// Run 用 VACUUM INTO 做在线备份（WAL 模式下直接拷贝 db 文件不可靠），
// 并按文件名时间戳轮转，仅保留最近 retainBackups 份
func Run(ctx context.Context) {
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		g.Log().Warningf(ctx, "dbbackup: create dir failed: %v", err)
		return
	}
	name := fmt.Sprintf("app-%s.db", time.Now().Format("20060102-150405"))
	target := filepath.ToSlash(filepath.Join(backupDir, name))

	// VACUUM INTO 目标文件不能已存在
	if _, err := os.Stat(target); err == nil {
		target = filepath.ToSlash(filepath.Join(backupDir,
			fmt.Sprintf("app-%s-%d.db", time.Now().Format("20060102-150405"), time.Now().UnixNano()%1000)))
	}
	if _, err := g.DB().Exec(ctx, "VACUUM INTO ?", target); err != nil {
		g.Log().Warningf(ctx, "dbbackup: vacuum into failed: %v", err)
		return
	}
	g.Log().Infof(ctx, "dbbackup: created %s", target)
	rotate(ctx)
}

func rotate(ctx context.Context) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		g.Log().Warningf(ctx, "dbbackup: read dir failed: %v", err)
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "app-") && strings.HasSuffix(e.Name(), ".db") {
			names = append(names, e.Name())
		}
	}
	// 文件名含时间戳，字典序即时间序
	sort.Strings(names)
	for len(names) > retainBackups {
		oldest := names[0]
		if err := os.Remove(filepath.Join(backupDir, oldest)); err != nil {
			g.Log().Warningf(ctx, "dbbackup: remove %s failed: %v", oldest, err)
			return
		}
		g.Log().Infof(ctx, "dbbackup: rotated out %s", oldest)
		names = names[1:]
	}
}
