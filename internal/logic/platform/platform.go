package platform

import (
	"context"
	"fmt"
	"strings"

	api "github.com/cicbyte/byte-code/api/v1/platform"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterPlatform(New())
}

func New() *sPlatform {
	return &sPlatform{}
}

type sPlatform struct{}

// ========== 标签 ==========

func (s *sPlatform) CreateTag(ctx context.Context, req *api.TagCreateReq) (id int, err error) {
	result, err := g.DB().Model("tags").Ctx(ctx).Insert(g.Map{
		"name":       req.Name,
		"color":      req.Color,
		"creator_id": ctx.Value("userId").(int),
	})
	if err != nil {
		return 0, err
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sPlatform) UpdateTag(ctx context.Context, req *api.TagUpdateReq) (err error) {
	_, err = g.DB().Model("tags").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"name":  req.Name,
		"color": req.Color,
	}).Update()
	return
}

func (s *sPlatform) DeleteTag(ctx context.Context, id int) (err error) {
	// 删除关联
	g.DB().Model("entity_tags").Ctx(ctx).Where("tag_id", id).Delete()
	_, err = g.DB().Model("tags").Ctx(ctx).Where("id", id).Delete()
	return
}

func (s *sPlatform) ListTags(ctx context.Context) (res *api.TagListRes, err error) {
	res = &api.TagListRes{}
	err = g.DB().Model("tags").Ctx(ctx).Order("id ASC").Scan(&res.List)
	return
}

func (s *sPlatform) AttachTag(ctx context.Context, req *api.TagAttachReq) (err error) {
	_, err = g.DB().Model("entity_tags").Ctx(ctx).Insert(g.Map{
		"tag_id":      req.Id,
		"entity_type": req.EntityType,
		"entity_id":   req.EntityId,
	})
	return
}

func (s *sPlatform) DetachTag(ctx context.Context, tagId int, entityType string, entityId int) (err error) {
	_, err = g.DB().Model("entity_tags").Ctx(ctx).
		Where("tag_id", tagId).
		Where("entity_type", entityType).
		Where("entity_id", entityId).
		Delete()
	return
}

func (s *sPlatform) GetTagEntities(ctx context.Context, tagId int) (res *api.TagEntitiesRes, err error) {
	res = &api.TagEntitiesRes{}

	var relations []struct {
		EntityType string
		EntityId   int
	}
	err = g.DB().Model("entity_tags").Ctx(ctx).Where("tag_id", tagId).Scan(&relations)
	if err != nil {
		return
	}

	for _, r := range relations {
		switch r.EntityType {
		case "task":
			res.Tasks = append(res.Tasks, r.EntityId)
		case "requirement":
			res.Requirements = append(res.Requirements, r.EntityId)
		case "test_case":
			res.TestCases = append(res.TestCases, r.EntityId)
		}
	}
	return
}

// ========== 活动�?==========

func (s *sPlatform) ListActivities(ctx context.Context, req *api.ActivityListReq) (res *api.ActivityListRes, err error) {
	res = &api.ActivityListRes{}
	m := g.DB().Model("activities").Ctx(ctx)

	if req.Module != "" {
		m = m.Where("target_type", req.Module)
	}
	if req.ActorId > 0 {
		m = m.Where("actor_id", req.ActorId)
	}
	if req.ProjectId > 0 {
		m = m.Where("project_id", req.ProjectId)
	}

	res.Total, err = m.Count()
	if err != nil {
		return
	}

	err = m.Page(req.Page, req.Size).Order("id DESC").Scan(&res.List)
	return
}

// ========== 通知 ==========

func (s *sPlatform) ListNotifications(ctx context.Context, req *api.NotificationListReq) (res *api.NotificationListRes, err error) {
	res = &api.NotificationListRes{}
	userId := ctx.Value("userId").(int)
	m := g.DB().Model("notifications").Ctx(ctx).Where("user_id", userId)

	if req.Unread == 1 {
		m = m.Where("is_read", 0)
	}

	res.Total, err = m.Count()
	if err != nil {
		return
	}

	err = m.Page(req.Page, req.Size).Order("id DESC").Scan(&res.List)
	return
}

func (s *sPlatform) ReadNotification(ctx context.Context, id int) (err error) {
	_, err = g.DB().Model("notifications").Ctx(ctx).
		Where("id", id).
		Data("is_read", 1).Update()
	return
}

func (s *sPlatform) ReadAllNotifications(ctx context.Context) (err error) {
	userId := ctx.Value("userId").(int)
	_, err = g.DB().Model("notifications").Ctx(ctx).
		Where("user_id", userId).
		Data("is_read", 1).Update()
	return
}

func (s *sPlatform) UnreadCount(ctx context.Context) (count int, err error) {
	userId := ctx.Value("userId").(int)
	count, err = g.DB().Model("notifications").Ctx(ctx).
		Where("user_id", userId).
		Where("is_read", 0).Count()
	return
}

// ========== 全局搜索 ==========

func (s *sPlatform) Search(ctx context.Context, req *api.SearchReq) (res *api.SearchRes, err error) {
	res = &api.SearchRes{}
	keyword := "%" + req.Q + "%"

	modules := []string{"task", "requirement", "doc", "test_case"}
	if req.Module != "" {
		modules = []string{req.Module}
	}

	for _, mod := range modules {
		var results []api.SearchResult

		switch mod {
		case "task":
			var items []struct {
				Id    int
				Title string
			}
			err = g.DB().Model("tasks").Ctx(ctx).
				Where("title LIKE ? OR description LIKE ?", keyword, keyword).
				Page(req.Page, req.Size).
				Scan(&items)
			if err == nil {
				for _, item := range items {
					results = append(results, api.SearchResult{Module: "task", Id: item.Id, Title: item.Title})
				}
			}

		case "requirement":
			var items []struct {
				Id    int
				Title string
			}
			err = g.DB().Model("requirements").Ctx(ctx).
				Where("title LIKE ? OR description LIKE ?", keyword, keyword).
				Page(req.Page, req.Size).
				Scan(&items)
			if err == nil {
				for _, item := range items {
					results = append(results, api.SearchResult{Module: "requirement", Id: item.Id, Title: item.Title})
				}
			}

		case "doc":
			var items []struct {
				Id    int
				Title string
			}
			// FTS5 全文搜索
			ftsQuery := fmt.Sprintf(
				"SELECT rowid as id, title FROM docs_fts WHERE docs_fts MATCH ? LIMIT %d",
				req.Size,
			)
			ftsErr := g.DB().Ctx(ctx).Raw(ftsQuery, req.Q).Scan(&items)
			if ftsErr != nil || len(items) == 0 {
				// FTS5 失败或无结果，降级到 LIKE
				g.DB().Model("docs").Ctx(ctx).
					Where("title LIKE ? OR content LIKE ?", keyword, keyword).
					Page(req.Page, req.Size).
					Scan(&items)
			}
			for _, item := range items {
				results = append(results, api.SearchResult{Module: "doc", Id: item.Id, Title: item.Title})
			}

		case "test_case":
			var items []struct {
				Id    int
				Title string
			}
			err = g.DB().Model("test_cases").Ctx(ctx).
				Where("title LIKE ? OR steps LIKE ?", keyword, keyword).
				Page(req.Page, req.Size).
				Scan(&items)
			if err == nil {
				for _, item := range items {
					results = append(results, api.SearchResult{Module: "test_case", Id: item.Id, Title: item.Title})
				}
			}
		}

		res.List = append(res.List, results...)
		res.Total += len(results)
	}

	return
}

func escapeFts(q string) string {
	// FTS5 特殊字符转义
	replacer := strings.NewReplacer(
		`"`, `""`,
		`'`, `''`,
	)
	return replacer.Replace(q)
}

// ========== 仪表�?==========

func (s *sPlatform) DashboardStats(ctx context.Context) (res *api.DashboardStatsRes, err error) {
	res = &api.DashboardStatsRes{}

	// 需求总数
	res.TotalRequirements, _ = g.DB().Model("requirements").Ctx(ctx).Count()

	// 任务统计
	res.TotalTasks, _ = g.DB().Model("tasks").Ctx(ctx).Count()
	res.InProgressTasks, _ = g.DB().Model("tasks").Ctx(ctx).Where("status", "in_progress").Count()
	res.ReviewTasks, _ = g.DB().Model("tasks").Ctx(ctx).Where("status", "review").Count()

	// 测试通过�?
	var passCount, totalCount int
	totalCount, _ = g.DB().Model("test_plan_cases").Ctx(ctx).Where("status != ?", "pending").Count()
	passCount, _ = g.DB().Model("test_plan_cases").Ctx(ctx).Where("status", "pass").Count()
	if totalCount > 0 {
		res.TestPassRate = float64(passCount) / float64(totalCount) * 100
	}

	// AI 统计
	type aiStat struct {
		AiName    string
		TaskCount int
	}
	var aiStats []aiStat
	g.DB().Ctx(ctx).Raw(`
		SELECT u.real_name as ai_name, COUNT(t.id) as task_count
		FROM tasks t
		JOIN sys_users u ON t.assignee_id = u.id AND u.type = 'ai'
		GROUP BY t.assignee_id
		ORDER BY task_count DESC
		LIMIT 10
	`).Scan(&aiStats)
	for _, s := range aiStats {
		res.AiStats = append(res.AiStats, api.AiStatItem{AiName: s.AiName, TaskCount: s.TaskCount})
	}

	// 最近任�?
	err = g.DB().Model("tasks").Ctx(ctx).
		Fields("id, title, status, updated_at").
		Order("updated_at DESC").Limit(5).
		Scan(&res.RecentTasks)

	return
}

// ========== 审计日志 ==========

func (s *sPlatform) ListAuditLogs(ctx context.Context, req *api.AuditLogListReq) (res *api.AuditLogListRes, err error) {
	res = &api.AuditLogListRes{}
	m := g.DB().Model("audit_logs").Ctx(ctx)

	if req.TargetType != "" {
		m = m.Where("target_type", req.TargetType)
	}
	if req.ActorId > 0 {
		m = m.Where("actor_id", req.ActorId)
	}
	if req.Action != "" {
		m = m.Where("action", req.Action)
	}
	if req.ProjectId > 0 {
		m = m.Where("project_id", req.ProjectId)
	}

	res.Total, err = m.Count()
	if err != nil {
		return
	}

	err = m.Page(req.Page, req.Size).Order("id DESC").Scan(&res.List)
	return
}
