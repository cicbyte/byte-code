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
	drainTimeout  = 3 * time.Second // 停机排水上限：超时剩余丢弃（不无限阻塞退出）
)

var (
	ch      = make(chan Entry, bufferSize)
	stopCh  = make(chan struct{})
	stopped = make(chan struct{})
)

// Record 非阻塞提交一条审计日志
func Record(e Entry) {
	select {
	case <-stopCh:
		return // 已停机：不再接收（进程退出路径）
	case ch <- e:
	default:
		g.Log().Warning(context.Background(), "auditwriter: buffer full, audit entry dropped")
	}
}

// Start 启动异步落盘协程：积攒到 batchSize 或每 flushInterval 批量写入。
// SQLite 单写者模型下避免每个写请求同步叠加一条审计写事务。
// flush 外层 recover：单次 panic 不至于杀死消费协程（死后 channel 塞满，
// 全部后续审计永久丢弃只剩告警）
func Start() {
	go func() {
		defer close(stopped)
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "auditwriter: consumer panic recovered: %v", r)
			}
		}()
		ctx := context.Background()
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		var buf []Entry
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
				// 停机排水：把缓冲里剩余的尽量写完
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

// Stop 优雅停机：停止接收新条目，排空缓冲后返回（超时放弃剩余）。
// 由服务关停钩子调用，避免进程退出丢最后 3 秒窗口内的审计
func Stop() {
	select {
	case <-stopCh:
		return // 已停
	default:
	}
	close(stopCh)
	select {
	case <-stopped:
	case <-time.After(drainTimeout):
		g.Log().Warning(context.Background(), "auditwriter: drain timeout, remaining entries dropped")
	}
}

func safeFlush(ctx context.Context, buf *[]Entry) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "auditwriter: flush panic recovered: %v", r)
			*buf = (*buf)[:0] // panic 可能发生在写库中段，防脏缓冲复用
		}
	}()
	flush(ctx, buf)
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
