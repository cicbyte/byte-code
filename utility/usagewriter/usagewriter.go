// Package usagewriter 使用埋点事件异步落盘（模式对齐 auditwriter）。
// usage_events 是可过期统计数据（30 天清理），尽力而为：满时丢弃告警，
// 绝不阻塞业务请求；SQLite 单写者模型下批量写避免叠加写事务。
package usagewriter

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Event 一条使用事件（字段与 usage_events 表对齐）
type Event struct {
	ActorID    int
	ActorType  string // human / ai
	Client     string // cli / web
	Method     string
	Endpoint   string // 模板化路径 /v1/tasks/{id}
	StatusCode int
	DurationMs int
	ProjectId  int
	SessionId  string
	ErrorCode  int // 业务响应壳 code（0=成功）
	Params     string
}

const (
	bufferSize    = 2048
	batchSize     = 200
	flushInterval = 3 * time.Second
	drainTimeout  = 3 * time.Second
)

var (
	ch      = make(chan Event, bufferSize)
	stopCh  = make(chan struct{})
	stopped = make(chan struct{})
)

// Record 非阻塞提交一条使用事件
func Record(e Event) {
	select {
	case <-stopCh:
		return
	case ch <- e:
	default:
		g.Log().Warning(context.Background(), "usagewriter: buffer full, usage event dropped")
	}
}

// Start 启动异步落盘协程
func Start() {
	go func() {
		defer close(stopped)
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "usagewriter: consumer panic recovered: %v", r)
			}
		}()
		ctx := context.Background()
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		var buf []Event
		for {
			select {
			case e := <-ch:
				buf = append(buf, e)
				if len(buf) >= batchSize {
					safeFlush(ctx, &buf)
				}
			case <-ticker.C:
				if len(buf) > 0 {
					safeFlush(ctx, &buf)
				}
			case <-stopCh:
				for {
					select {
					case e := <-ch:
						buf = append(buf, e)
						if len(buf) >= batchSize {
							safeFlush(ctx, &buf)
						}
					default:
						if len(buf) > 0 {
							safeFlush(ctx, &buf)
						}
						return
					}
				}
			}
		}
	}()
}

// Stop 优雅停机排水
func Stop() {
	select {
	case <-stopCh:
		return
	default:
	}
	close(stopCh)
	select {
	case <-stopped:
	case <-time.After(drainTimeout):
		g.Log().Warning(context.Background(), "usagewriter: drain timeout, remaining events dropped")
	}
}

func safeFlush(ctx context.Context, buf *[]Event) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "usagewriter: flush panic recovered: %v", r)
			*buf = (*buf)[:0]
		}
	}()
	flush(ctx, buf)
}

func flush(ctx context.Context, buf *[]Event) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rows := make([]g.Map, 0, len(*buf))
	for _, e := range *buf {
		rows = append(rows, g.Map{
			"actor_id":    e.ActorID,
			"actor_type":  e.ActorType,
			"client":      e.Client,
			"method":      e.Method,
			"endpoint":    e.Endpoint,
			"status_code": e.StatusCode,
			"duration_ms": e.DurationMs,
			"project_id":  e.ProjectId,
			"session_id":  e.SessionId,
			"error_code":  e.ErrorCode,
			"params":      e.Params,
			"created_at":  now,
		})
	}
	if _, err := g.DB().Model("usage_events").Ctx(ctx).Insert(rows); err != nil {
		g.Log().Warningf(ctx, "usagewriter: flush failed: %v", err)
	}
	*buf = (*buf)[:0]
}
