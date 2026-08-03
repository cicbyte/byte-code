package dbclean

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	retainAuditDays      = 90 // 审计日志与活动流保留天数
	retainReadNotifyDays = 30 // 已读通知保留天数
)

// Run 清理只增不减的表：过期 token、超期审计日志/活动流、超期已读通知。
// 启动时执行一次，之后由每日定时任务调用；未读通知不清理（等用户读后再计龄）。
func Run(ctx context.Context) {
	// 时间列均由 Go 侧以本地时间字符串写入，用相同时区参数比较
	now := time.Now().Format("2006-01-02 15:04:05")
	auditBefore := time.Now().AddDate(0, 0, -retainAuditDays).Format("2006-01-02 15:04:05")
	notifyBefore := time.Now().AddDate(0, 0, -retainReadNotifyDays).Format("2006-01-02 15:04:05")

	jobs := []struct {
		name string
		sql  string
		args []interface{}
	}{
		{"expired_tokens", "DELETE FROM sys_tokens WHERE expired_at <= ?", []interface{}{now}},
		{"stale_audit_logs", "DELETE FROM audit_logs WHERE created_at < ?", []interface{}{auditBefore}},
		{"stale_activities", "DELETE FROM activities WHERE created_at < ?", []interface{}{auditBefore}},
		{"stale_read_notifications", "DELETE FROM notifications WHERE is_read = 1 AND created_at < ?", []interface{}{notifyBefore}},
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
}
