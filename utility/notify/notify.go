package notify

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

func Send(ctx context.Context, userId int, title, content, notifyType, sourceType string, sourceId int) {
	_, err := g.DB().Model("notifications").Ctx(ctx).Insert(g.Map{
		"user_id":     userId,
		"title":       title,
		"content":     content,
		"type":        notifyType,
		"is_read":     0,
		"source_type": sourceType,
		"source_id":   sourceId,
		"created_at":  time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		g.Log().Warningf(ctx, "Failed to send notification: %v", err)
	}
	// 实时推送：在线的 SSE 订阅者立即收到（离线者下次拉列表可见，DB 仍是真相源）
	Publish(userId, map[string]interface{}{
		"title": title, "content": content, "type": notifyType,
		"sourceType": sourceType, "sourceId": sourceId,
	})
}

// ==================== SSE 订阅 hub ====================
// 每用户一组合有缓冲 channel；Send 落库后 Publish 非阻塞广播。
// 通知的真相源是 DB 表——hub 只做"在线加速"，连接断开/重启不丢通知

type subscriber struct {
	ch chan []byte
}

var (
	hubMu   sync.RWMutex
	hubSubs = make(map[int]map[*subscriber]struct{})
)

// Subscribe 订阅某用户的实时通知；返回事件流 channel 与取消函数
func Subscribe(userId int) (<-chan []byte, context.CancelFunc) {
	sub := &subscriber{ch: make(chan []byte, 8)}
	hubMu.Lock()
	if hubSubs[userId] == nil {
		hubSubs[userId] = make(map[*subscriber]struct{})
	}
	hubSubs[userId][sub] = struct{}{}
	hubMu.Unlock()
	cleanup := func() {
		hubMu.Lock()
		delete(hubSubs[userId], sub)
		if len(hubSubs[userId]) == 0 {
			delete(hubSubs, userId)
		}
		hubMu.Unlock()
	}
	return sub.ch, cleanup
}

// Publish 向某用户的全部在线订阅广播（满缓冲则丢弃——客户端还有轮询兜底与 DB）
func Publish(userId int, payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	hubMu.RLock()
	subs := hubSubs[userId]
	hubMu.RUnlock()
	for sub := range subs {
		select {
		case sub.ch <- data:
		default: // 缓冲满：慢消费者不阻塞发送方
		}
	}
}
