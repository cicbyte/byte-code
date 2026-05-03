package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func (s *sMiddleware) MiddlewareAuditLog(r *ghttp.Request) {
	r.Middleware.Next()

	method := r.Method
	if method != "POST" && method != "PUT" && method != "DELETE" {
		return
	}

	ctx := r.Context()
	userId, _ := ctx.Value("userId").(int)
	if userId == 0 {
		return
	}

	path := r.URL.Path
	action := "create"
	if method == "PUT" {
		action = "update"
	} else if method == "DELETE" {
		action = "delete"
	}

	target := parseTarget(path)

	actorType := "human"
	actor, _ := ctx.Value("actorType").(string)
	if actor == "ai" {
		actorType = "ai"
	}

	projectId, _ := ctx.Value("projectId").(int)

	_, err := g.DB().Model("audit_logs").Ctx(ctx).Insert(g.Map{
		"actor_id":    userId,
		"actor_type":  actorType,
		"action":      action,
		"target_type": target.entityType,
		"target_id":   target.entityId,
		"target_name": "",
		"ip_address":  r.GetClientIp(),
		"user_agent":  r.UserAgent(),
		"project_id":  projectId,
		"created_at":  time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		g.Log().Warningf(ctx, "Failed to write audit log: %v", err)
	}
}

type targetInfo struct {
	entityType string
	entityId   int
}

func parseTarget(path string) targetInfo {
	// /api/v1/projects/123/tasks/456
	// /api/v1/tasks/456
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	info := targetInfo{}

	for i := 0; i < len(parts)-1; i++ {
		// 如果下一部分是数字，当前部分是实体类型
		if isNumeric(parts[i+1]) && i%2 == 0 {
			info.entityType = parts[i]
			// 取最后一个匹配
		}
	}

	// 取最后一个数字作为 entityId
	for i := len(parts) - 1; i >= 0; i-- {
		if isNumeric(parts[i]) {
			info.entityId, _ = strconv.Atoi(parts[i])
			break
		}
	}

	return info
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
