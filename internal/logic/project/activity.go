package project

import (
	"context"
	"fmt"
	"strings"

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

	// 同主题聚合：同 actor+action+target 1 小时内的重复动态不新增，
	// 而是累计到已有条目（detail 带 ×N 次数）——批量拆解/阶段推进等
	// 高频操作不再逐条刷屏动态流
	if row, err := g.DB().Model("activities").Ctx(ctx).
		Where("actor_id", actorId).
		Where("action", action).
		Where("target_type", targetType).
		Where("target_id", targetId).
		Where("created_at > datetime('now', 'localtime', '-1 hour')").
		Order("id DESC").One(); err == nil && !row.IsEmpty() {
		cnt := row["detail"].String()
		n := 1
		if strings.HasPrefix(cnt, "×") {
			fmt.Sscanf(cnt, "×%d", &n)
			cnt = ""
		}
		newDetail := detail
		if newDetail == "" {
			newDetail = cnt
		}
		g.DB().Model("activities").Ctx(ctx).Where("id", row["id"].Int()).
			Data(g.Map{"detail": fmt.Sprintf("×%d %s", n+1, newDetail), "target_name": targetName}).Update()
		return
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
