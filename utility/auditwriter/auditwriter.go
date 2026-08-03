package auditwriter

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Entry 一条审计日志（字段与 audit_logs 表对齐）
type Entry struct {
	ActorID    int
	ActorType  string
	Action     string
	TargetType string
	TargetID   int
	TargetName string
	IpAddress  string
	UserAgent  string
	ProjectId  int
}

const (
	bufferSize    = 1024            // 缓冲上限；审计尽力而为，满时丢弃并告警，绝不阻塞业务请求
	batchSize     = 100              // 单批写入上限
	flushInterval = 3 * time.Second // 批量落盘间隔
)

var ch = make(chan Entry, bufferSize)

// Record 非阻塞提交一条审计日志
func Record(e Entry) {
	select {
	case ch <- e:
	default:
		g.Log().Warning(context.Background(), "auditwriter: buffer full, audit entry dropped")
	}
}

// Start 启动异步落盘协程：积攒到 batchSize 或每 flushInterval 批量写入。
// SQLite 单写者模型下避免每个写请求同步叠加一条审计写事务。
func Start() {
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		var buf []Entry
		for {
			select {
			case e := <-ch:
				buf = append(buf, e)
				if len(buf) >= batchSize {
					flush(ctx, &buf)
				}
			case <-ticker.C:
				if len(buf) > 0 {
					flush(ctx, &buf)
				}
			}
		}
	}()
}

func flush(ctx context.Context, buf *[]Entry) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rows := make([]g.Map, 0, len(*buf))
	for _, e := range *buf {
		rows = append(rows, g.Map{
			"actor_id":    e.ActorID,
			"actor_type":  e.ActorType,
			"action":      e.Action,
			"target_type": e.TargetType,
			"target_id":   e.TargetID,
			"target_name": e.TargetName,
			"ip_address":  e.IpAddress,
			"user_agent":  e.UserAgent,
			"project_id":  e.ProjectId,
			"created_at":  now,
		})
	}
	if _, err := g.DB().Model("audit_logs").Ctx(ctx).Insert(rows); err != nil {
		g.Log().Warningf(ctx, "auditwriter: flush failed: %v", err)
	}
	*buf = (*buf)[:0]
}
