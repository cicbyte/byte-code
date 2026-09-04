package project

import (
	"context"
	"fmt"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/consts"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Sprint 燃尽图

func (s *sProject) GetBurndown(ctx context.Context, sprintId int) (res *api.BurndownRes, err error) {
	res = &api.BurndownRes{}

	// 获取 Sprint 信息
	var sprint api.SprintItem
	err = g.DB().Model("sprints").Ctx(ctx).Where("id", sprintId).Scan(&sprint)
	if err != nil || sprint.Id == 0 {
		return nil, fmt.Errorf("Sprint不存在")
	}

	// Sprint 下的总任务数
	totalCount, err := g.DB().Model("tasks").Ctx(ctx).
		Where("sprint_id", sprintId).
		Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询任务数量失败")
	}

	// 按天统计完成的任务数
	type dayCount struct {
		Date      string
		Completed int
	}
	var completedByDay []dayCount
	err = g.DB().Model("tasks").Ctx(ctx).
		Fields("DATE(COALESCE(NULLIF(completed_at, ''), updated_at)) as date, COUNT(*) as completed").
		Where("sprint_id", sprintId).
		WhereIn("status", consts.TaskTerminalStatuses).
		Group(`DATE(COALESCE(NULLIF(completed_at, ''), updated_at))`).
		Order("date ASC").
		Scan(&completedByDay)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "统计燃尽图数据失败")
	}

	// 构建按天的数据
	completedMap := make(map[string]int)
	for _, dc := range completedByDay {
		completedMap[dc.Date] = dc.Completed
	}

	startDate, err := gtime.StrToTime(sprint.StartDate)
	if err != nil {
		startDate = gtime.Now()
	}
	endDate, err := gtime.StrToTime(sprint.EndDate)
	if err != nil {
		endDate = gtime.Now()
	}

	items := make([]api.BurndownItem, 0)
	cumulativeCompleted := 0
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("Y-m-d")
		if c, ok := completedMap[dateStr]; ok {
			cumulativeCompleted += c
		}
		items = append(items, api.BurndownItem{
			Date:      dateStr,
			Remaining: totalCount - cumulativeCompleted,
			Completed: cumulativeCompleted,
		})
	}
	res.Items = items
	return res, nil
}
