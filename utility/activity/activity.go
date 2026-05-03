package activity

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ActivityInput struct {
	ActorID    int
	ActorType  string // human, ai, system
	ActorName  string
	Action     string // task.created, task.status_changed, etc.
	TargetType string // task, doc, requirement, etc.
	TargetID   int
	TargetName string
	ProjectID  int
	Detail     string
}

func Record(ctx context.Context, input ActivityInput) {
	_, err := g.DB().Model("activities").Ctx(ctx).Insert(g.Map{
		"actor_id":    input.ActorID,
		"actor_type":  input.ActorType,
		"actor_name":  input.ActorName,
		"action":      input.Action,
		"target_type": input.TargetType,
		"target_id":   input.TargetID,
		"target_name": input.TargetName,
		"project_id":  input.ProjectID,
		"detail":      input.Detail,
		"created_at":  time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		g.Log().Warningf(ctx, "Failed to record activity: %v", err)
	}
}
