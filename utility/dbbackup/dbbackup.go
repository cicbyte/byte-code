package dbbackup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/cicbyte/byte-code/utility/dbinit"
)

const (
	backupDir      = "resource/backup" // 备份目录（.gitignore 忽略）
	retainBackups  = 7                 // 保留最近份数
)

// Run 用 VACUUM INTO 做在线备份（WAL 模式下直接拷贝 db 文件不可靠），
// 并按文件名时间戳轮转，仅保留最近 retainBackups 份。
// 磁盘文档（resource/projects/*/docs）是记忆/文档中枢的真相源——DB 丢了可重扫，
// 磁盘丢了文档就没了，故同轮打包备份
func Run(ctx context.Context) {
	// MySQL 无文件级备份：公网部署通常有外部方案（云 RDS 快照/mysqldump 定时），
	// 这里跳过 DB 备份但保留文档 vault 备份
	if dbinit.Dialect() == "mysql" {
		g.Log().Info(ctx, "dbbackup: MySQL 模式跳过内置 DB 备份（请用 mysqldump/云快照）；文档 vault 照常备份")
		backupDocs(ctx)
		rotate(ctx)
		return
	}
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
	backupDocs(ctx)
	rotate(ctx)
}

// backupDocs 打包全部项目文档目录（tar.gz，时间戳命名）。
// 排除 .history（快照体积大且可从主文件重建意义有限；如需完整历史可调开关）
func backupDocs(ctx context.Context) {
	src := filepath.Join("resource", "projects")
	if _, err := os.Stat(src); err != nil {
		return // 尚无任何项目数据
	}
	name := fmt.Sprintf("docs-%s.tar.gz", time.Now().Format("20060102-150405"))
	target := filepath.Join(backupDir, name)
	// 同名兜底：os.Create 遇已存在文件会静默截断旧档（db 分支同款防护）
	if _, err := os.Stat(target); err == nil {
		target = filepath.Join(backupDir,
			fmt.Sprintf("docs-%s-%d.tar.gz", time.Now().Format("20060102-150405"), time.Now().UnixNano()%1000))
	}
	f, err := os.Create(target)
	if err != nil {
		g.Log().Warningf(ctx, "dbbackup: create docs archive failed: %v", err)
		return
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	err = filepath.Walk(src, func(p string, fi os.FileInfo, werr error) error {
		if werr != nil {
			return werr
		}
		rel, rerr := filepath.Rel(src, p)
		if rerr != nil {
			return rerr
		}
		relSlash := filepath.ToSlash(rel)
		if fi.IsDir() {
			// 跳过 .history 目录（含其子树）
			if fi.Name() == ".history" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(relSlash, "/.history/") {
			return nil
		}
		hdr, herr := tar.FileInfoHeader(fi, "")
		if herr != nil {
			return herr
		}
		hdr.Name = relSlash
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		data, derr := os.ReadFile(p)
		if derr != nil {
			return derr
		}
		_, err := tw.Write(data)
		return err
	})
	if err != nil {
		g.Log().Warningf(ctx, "dbbackup: walk docs failed: %v", err)
		_ = os.Remove(target)
		return
	}
	g.Log().Infof(ctx, "dbbackup: created %s", target)
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

	// 文档归档同轮转策略
	var archives []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "docs-") && strings.HasSuffix(e.Name(), ".tar.gz") {
			archives = append(archives, e.Name())
		}
	}
	sort.Strings(archives)
	for len(archives) > retainBackups {
		oldest := archives[0]
		if err := os.Remove(filepath.Join(backupDir, oldest)); err != nil {
			g.Log().Warningf(ctx, "dbbackup: remove %s failed: %v", oldest, err)
			return
		}
		g.Log().Infof(ctx, "dbbackup: rotated out %s", oldest)
		archives = archives[1:]
	}
}
