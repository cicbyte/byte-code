package dbclean

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	retainAuditDays      = 90 // 审计日志与活动流保留天数
	retainReadNotifyDays = 30 // 已读通知保留天数
	retainExpiredMemDays = 30 // expired 记忆删除前保留天数（终态，仅留观察窗口）
	retainAiLogDays      = 90 // AI 执行日志保留天数（detail 存完整 LLM 输出，增长最快）
	retainHistoryPerFile = 20 // vault 每文件 .history 快照保留份数
)

// Run 清理只增不减的数据：过期 token、超期审计日志/活动流、超期已读通知、
// 超期 expired 记忆、超期 AI 执行日志、超量 vault 快照。
// 启动时执行一次，之后由每日定时任务调用；未读通知不清理（等用户读后再计龄）。
func Run(ctx context.Context) {
	// 时间列均由 Go 侧以本地时间字符串写入，用相同时区参数比较
	now := time.Now().Format("2006-01-02 15:04:05")
	auditBefore := time.Now().AddDate(0, 0, -retainAuditDays).Format("2006-01-02 15:04:05")
	notifyBefore := time.Now().AddDate(0, 0, -retainReadNotifyDays).Format("2006-01-02 15:04:05")
	memBefore := time.Now().AddDate(0, 0, -retainExpiredMemDays).Format("2006-01-02 15:04:05")
	aiLogBefore := time.Now().AddDate(0, 0, -retainAiLogDays).Format("2006-01-02 15:04:05")

	jobs := []struct {
		name string
		sql  string
		args []interface{}
	}{
		{"expired_tokens", "DELETE FROM sys_tokens WHERE expired_at <= ?", []interface{}{now}},
		{"stale_audit_logs", "DELETE FROM audit_logs WHERE created_at < ?", []interface{}{auditBefore}},
		{"stale_activities", "DELETE FROM activities WHERE created_at < ?", []interface{}{auditBefore}},
		{"stale_read_notifications", "DELETE FROM notifications WHERE is_read = 1 AND created_at < ?", []interface{}{notifyBefore}},
		// expired 记忆是状态机终态（无任何删除路径），不清理会挤占 MemList 配额并无限膨胀
		{"expired_memories", "DELETE FROM project_memories WHERE status = 'expired' AND updated_at < ?", []interface{}{memBefore}},
		// AI 执行日志每任务至少 3 条、complete 存完整 LLM 输出，开启引擎后增长最快
		{"stale_ai_logs", "DELETE FROM ai_execution_logs WHERE created_at < ?", []interface{}{aiLogBefore}},
	}
	for _, j := range jobs {
		res, err := g.DB().Exec(ctx, j.sql, j.args...)
		if err != nil {
			g.Log().Warningf(ctx, "dbclean %s failed: %v", j.name, err)
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			g.Log().Infof(ctx, "dbclean %s: removed %d rows", j.name, n)
		}
	}
	cleanVaultHistory(ctx)
}

// cleanVaultHistory 每文件 .history 快照保留最近 retainHistoryPerFile 份：
// 高频编辑的文档每天可产生数百份快照，无保留策略磁盘会被静默耗尽
func cleanVaultHistory(ctx context.Context) {
	projectsRoot := filepath.Join("resource", "projects")
	entries, err := os.ReadDir(projectsRoot)
	if err != nil {
		return
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		histRoot := filepath.Join(projectsRoot, e.Name(), "vault", ".history")
		_ = filepath.Walk(histRoot, func(p string, fi os.FileInfo, err error) error {
			if err != nil || !fi.IsDir() {
				return nil //nolint
			}
			files, rerr := os.ReadDir(p)
			if rerr != nil || len(files) <= retainHistoryPerFile {
				return nil //nolint
			}
			// 目录内全是快照文件（文件名即时间戳），按名排序淘汰最旧
			sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
			for _, f := range files[:len(files)-retainHistoryPerFile] {
				if f.IsDir() {
					continue
				}
				if os.Remove(filepath.Join(p, f.Name())) == nil {
					removed++
				}
			}
			return nil //nolint
		})
	}
	if removed > 0 {
		g.Log().Infof(ctx, "dbclean vault_history: removed %d snapshots (keep %d per file)", removed, retainHistoryPerFile)
	}
}
