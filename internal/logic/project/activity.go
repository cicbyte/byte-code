package project

import (
	"time"
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
	// 高频操作不再逐条刷屏动态流。
	// 阈值在应用层算好传参：datetime('now','localtime') 是 SQLite 方言，
	// MySQL 无此函数恒报错 → 聚合静默失效退化为逐条刷屏（dashboard/auth 同因已修）
	oneHourAgo := time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05")
	if row, err := g.DB().Model("activities").Ctx(ctx).
		Where("actor_id", actorId).
		Where("action", action).
		Where("target_type", targetType).
		Where("target_id", targetId).
		Where("created_at > ?", oneHourAgo).
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
		if _, ue := g.DB().Model("activities").Ctx(ctx).Where("id", row["id"].Int()).
			Data(g.Map{"detail": fmt.Sprintf("×%d %s", n+1, newDetail), "target_name": targetName}).Update(); ue != nil {
			g.Log().Warningf(ctx, "activity aggregate update failed: %v", ue)
		}
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
