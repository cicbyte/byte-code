package notify

import (
	"context"
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
}
