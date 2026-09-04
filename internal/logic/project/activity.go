package project

import (
	"context"

	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/gogf/gf/v2/frame/g"
)

// 活动记录辅助

func (s *sProject) recordActivity(ctx context.Context, actorId int, action, targetType string, targetId int, targetName string, projectId int, detail string) {
	actorName := ""
	if actorId > 0 {
		var user struct {
			Username string
			RealName string
		}
		err := g.DB().Model("sys_users").Ctx(ctx).Where("id", actorId).Scan(&user)
		if err == nil && user.RealName != "" {
			actorName = user.RealName
		} else if err == nil {
			actorName = user.Username
		}
	}

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    actorId,
		ActorType:  "human",
		ActorName:  actorName,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetId,
		TargetName: targetName,
		ProjectID:  projectId,
		Detail:     detail,
	})
}
