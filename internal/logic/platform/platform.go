package platform

import (
	"context"
	"fmt"

	liberr "github.com/cicbyte/byte-code/library/liberr"
	api "github.com/cicbyte/byte-code/api/v1/platform"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/internal/consts"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
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
		return 0, liberr.WrapDb(ctx, err, "创建标签失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sPlatform) UpdateTag(ctx context.Context, req *api.TagUpdateReq) (err error) {
	_, err = g.DB().Model("tags").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"name":  req.Name,
		"color": req.Color,
	}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新标签失败")
	}
	return nil
}

func (s *sPlatform) DeleteTag(ctx context.Context, id int) (err error) {
	// 标签及其全部实体关联在同一事务内删除
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Delete("entity_tags", "tag_id", id); err != nil {
			return err
		}
		_, err := tx.Delete("tags", "id", id)
		return err
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除标签失败")
	}
	return nil
}

func (s *sPlatform) ListTags(ctx context.Context) (res *api.TagListRes, err error) {
	res = &api.TagListRes{}
	err = g.DB().Model("tags").Ctx(ctx).Order("id ASC").Limit(500).Scan(&res.List)
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

	// 指定项目须是成员/管理员；未指定时非管理员仅见自己所在项目的动态
	uid := perm.UserId(ctx)
	if req.ProjectId > 0 {
		if !perm.CanAccessProject(ctx, uid, req.ProjectId) {
			return nil, fmt.Errorf("无权限查看该项目动态")
		}
	} else if !perm.IsAdmin(ctx, uid) {
		m = m.Where(
			"EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = activities.project_id AND pm.user_id = ?)", uid)
	}

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
	if req.Unread == 2 {
		m = m.Where("is_read", 1)
	}
	if req.Type != "" {
		m = m.Where("type", req.Type)
	}

	res.Total, err = m.Count()
	if err != nil {
		return
	}

	err = m.Page(req.Page, req.Size).Order("id DESC").Scan(&res.List)
	return
}

func (s *sPlatform) ReadNotification(ctx context.Context, id int) (err error) {
	// 只能操作自己的通知：无 user_id 条件会允许标记他人通知已读
	userId := perm.UserId(ctx)
	result, err := g.DB().Model("notifications").Ctx(ctx).
		Where("id", id).
		Where("user_id", userId).
		Data("is_read", 1).Update()
	if err != nil {
		return fmt.Errorf("标记已读失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("通知不存在")
	}
	return nil
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
	// 作用域与 DashboardStats 同口径：非管理员只能命中自己有访问权的项目
	// （human=成员项目，agent=绑定项目；sys_users 的 id 全局唯一，两种
	// EXISTS 不会跨类型误配）。无过滤时任意已认证账号——含零权限的自助
	// 注册 agent——可全库检索所有项目的任务/需求/记忆，属跨项目泄露
	uid := perm.UserId(ctx)
	memberOnly := uid > 0 && !perm.IsAdmin(ctx, uid)
	if uid <= 0 {
		return res, nil
	}
	keyword := "%" + escape.Like(req.Q) + "%"

	// 各模块统一 LIKE 检索，返回所属项目便于前端跳转；
	// docs 的 FTS5 分支已移除：docs_fts 表从未创建，原实现恒降级到 LIKE
	searchMap := map[string]struct {
		table string
		where string
	}{
		"task":        {"tasks", "title LIKE ? ESCAPE '\\' OR description LIKE ? ESCAPE '\\'"},
		"requirement": {"requirements", "title LIKE ? ESCAPE '\\' OR description LIKE ? ESCAPE '\\'"},
		"doc":         {"project_document_index", "title LIKE ? ESCAPE '\\' OR tags LIKE ? ESCAPE '\\'"},
		"test_case":   {"test_cases", "title LIKE ? ESCAPE '\\' OR steps LIKE ? ESCAPE '\\'"},
		// 记忆中枢定位下记忆必须可搜：key 与 value 全文命中
		"memory":      {"project_memories", "`key` LIKE ? ESCAPE '\\' OR value LIKE ? ESCAPE '\\'"},
	}
	modules := []string{"task", "requirement", "doc", "test_case", "memory"}
	if req.Module != "" {
		modules = []string{req.Module}
	}

	for _, mod := range modules {
		ms, ok := searchMap[mod]
		if !ok {
			continue
		}
		// 原条件含 OR（title OR description），拼 AND 前必须整体加括号：
		// AND 优先级高于 OR，裸拼会让 title 命中的行绕过访问范围
		where, args := "("+ms.where+")", []interface{}{keyword, keyword}
		if memberOnly {
			// 全局记忆（project_id=0）按既有语义全员可读（GET /global-memories
			// 同口径），仅项目级数据收进访问范围
			globalOk := ""
			if ms.table == "project_memories" {
				globalOk = ms.table + ".project_id = 0 OR "
			}
			where += fmt.Sprintf(
				" AND ("+globalOk+"EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = %s.project_id AND pm.user_id = ?)"+
					" OR EXISTS (SELECT 1 FROM agent_project_bindings ab WHERE ab.project_id = %s.project_id AND ab.agent_id = ?))",
				ms.table, ms.table)
			args = append(args, uid, uid)
		}
		// Total 汇总各模块的总命中数（此前误把当页条数当总数）
		total, err := g.DB().Model(ms.table).Ctx(ctx).Where(where, args...).Count()
		if err != nil {
			return nil, liberr.WrapDb(ctx, err, "搜索失败")
		}
		res.Total += total

		var items []struct {
			Id        int
			ProjectId int
			Title     string
		}
		// vault 索引表（project_document_index）主键是 (project_id, path) 复合键，
		// 无自增 id：按表选择字段与排序
		fields, order := "id, project_id, title", "id DESC"
		if ms.table == "project_document_index" {
			fields, order = "path, project_id, title", "updated_at DESC"
		}
		// memory 的展示标题是 key 列
		if ms.table == "project_memories" {
			fields = "id, project_id, [key] AS title"
		}
		err = g.DB().Model(ms.table).Ctx(ctx).
			Fields(fields).
			Where(where, args...).
			Page(req.Page, req.Size).
			Order(order).
			Scan(&items)
		if err != nil {
			return nil, liberr.WrapDb(ctx, err, "搜索失败")
		}
		for _, item := range items {
			res.List = append(res.List, api.SearchResult{
				Module:    mod,
				Id:        item.Id,
				ProjectId: item.ProjectId,
				Title:     item.Title,
			})
		}
	}

	return
}

func (s *sPlatform) DashboardStats(ctx context.Context) (res *api.DashboardStatsRes, err error) {
	res = &api.DashboardStatsRes{
		AiStats:     []api.AiStatItem{},
		RecentTasks: []api.RecentTaskItem{},
	}

	// 非管理员统计范围限定在自己所在的项目，避免跨项目数据（含任务标题）泄露
	uid := perm.UserId(ctx)
	memberOnly := uid > 0 && !perm.IsAdmin(ctx, uid)
	taskScope := "EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = tasks.project_id AND pm.user_id = ?)"
	reqScope := "EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = requirements.project_id AND pm.user_id = ?)"
	tpcScope := "EXISTS (SELECT 1 FROM test_plans tp JOIN project_members pm ON pm.project_id = tp.project_id WHERE tp.id = test_plan_cases.test_plan_id AND pm.user_id = ?)"

	// 需求总数
	reqM := g.DB().Model("requirements").Ctx(ctx)
	if memberOnly {
		reqM = reqM.Where(reqScope, uid)
	}
	res.TotalRequirements, err = reqM.Count()

	// 任务统计
	scopedTasks := func(status string) *gdb.Model {
		m := g.DB().Model("tasks").Ctx(ctx)
		if status != "" {
			m = m.Where("status", status)
		}
		if memberOnly {
			m = m.Where(taskScope, uid)
		}
		return m
	}
	res.TotalTasks, err = scopedTasks("").Count()
	res.InProgressTasks, err = scopedTasks("in_progress").Count()
	res.ReviewTasks, err = scopedTasks("review").Count()

	// 测试通过率
	tpcScoped := func() *gdb.Model {
		m := g.DB().Model("test_plan_cases").Ctx(ctx)
		if memberOnly {
			m = m.Where(tpcScope, uid)
		}
		return m
	}
	var passCount, totalCount int
	totalCount, err = tpcScoped().Where("status != ?", consts.TestPlanCasePending).Count()
	passCount, err = tpcScoped().Where("status", consts.TestPlanCasePass).Count()
	if totalCount > 0 {
		res.TestPassRate = float64(passCount) / float64(totalCount) * 100
	}

	// AI 统计
	type aiStat struct {
		AiName    string
		TaskCount int
	}
	var aiStats []aiStat
	aiSql := `
		SELECT u.real_name as ai_name, COUNT(t.id) as task_count
		FROM tasks t
		JOIN sys_users u ON t.assignee_id = u.id AND u.type = 'ai'`
	if memberOnly {
		aiSql += ` WHERE EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = t.project_id AND pm.user_id = ?)`
	}
	aiSql += ` GROUP BY t.assignee_id ORDER BY task_count DESC LIMIT 10`
	if memberOnly {
		g.DB().Ctx(ctx).Raw(aiSql, uid).Scan(&aiStats)
	} else {
		g.DB().Ctx(ctx).Raw(aiSql).Scan(&aiStats)
	}
	for _, s := range aiStats {
		res.AiStats = append(res.AiStats, api.AiStatItem{AiName: s.AiName, TaskCount: s.TaskCount})
	}

	// 最近任务
	recentM := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, title, status, updated_at").
		Order("updated_at DESC").Limit(5)
	if memberOnly {
		recentM = recentM.Where(taskScope, uid)
	}
	err = recentM.Scan(&res.RecentTasks)

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
