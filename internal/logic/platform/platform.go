package platform

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	liberr "github.com/cicbyte/byte-code/library/liberr"
	api "github.com/cicbyte/byte-code/api/v1/platform"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/internal/consts"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
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
	// 标签定义属全局共享数据，按权限字典放行（原超管组硬门槛改为字典驱动）
	if err := perm.RequireMenuPerm(ctx, "platform_tags"); err != nil {
		return 0, err
	}
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
	if err := perm.RequireMenuPerm(ctx, "platform_tags"); err != nil {
		return err
	}
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
	if err := perm.RequireMenuPerm(ctx, "platform_tags"); err != nil {
		return err
	}
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

// enrichTransferStatus transfer 类通知回填邀请当前状态：前端据此只对
// pending 显示接受/拒绝（已决显示结果标签），避免对已处理邀请操作报
// 「邀请已处理过」。SSE 实时推送不带此字段，前端按 pending 兜底
func enrichTransferStatus(ctx context.Context, list []api.NotificationItem) {
	ids := make([]int, 0, 4)
	for i := range list {
		if list[i].SourceType == "transfer" && list[i].SourceId > 0 {
			ids = append(ids, list[i].SourceId)
		}
	}
	if len(ids) == 0 {
		return
	}
	rows, err := g.DB().Model("project_transfers").Ctx(ctx).
		Where("id IN (?)", ids).Fields("id, status").All()
	if err != nil {
		return // 回填尽力而为，失败不阻断列表
	}
	st := make(map[int]string, len(rows))
	for _, r := range rows {
		st[r["id"].Int()] = r["status"].String()
	}
	for i := range list {
		if list[i].SourceType == "transfer" {
			list[i].TransferStatus = st[list[i].SourceId]
		}
	}
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
	if err != nil {
		return
	}
	enrichTransferStatus(ctx, res.List)
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
		// MySQL 只计「实际变更行」：已是已读态的 1→1 更新返回 0（SQLite
		// 计匹配行无此差异）——需存在性复核区分「已读（幂等成功）」与
		// 「真不存在」。生产实测：移交决议后端已标读，前端补一次标读
		// 在 MySQL 上误报「通知不存在」即此因（#496）
		cnt, cerr := g.DB().Model("notifications").Ctx(ctx).
			Where("id", id).
			Where("user_id", userId).
			Count()
		if cerr != nil {
			return fmt.Errorf("标记已读失败")
		}
		if cnt == 0 {
			return fmt.Errorf("通知不存在")
		}
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
		"task":        {"tasks", "title LIKE ? ESCAPE '|' OR description LIKE ? ESCAPE '|'"},
		"requirement": {"requirements", "title LIKE ? ESCAPE '|' OR description LIKE ? ESCAPE '|'"},
		"doc":         {"project_document_index", "title LIKE ? ESCAPE '|' OR tags LIKE ? ESCAPE '|'"},
		"test_case":   {"test_cases", "title LIKE ? ESCAPE '|' OR steps LIKE ? ESCAPE '|'"},
		// 记忆中枢定位下记忆必须可搜：key 与 value 全文命中
		"memory":      {"project_memories", "`key` LIKE ? ESCAPE '|' OR value LIKE ? ESCAPE '|'"},
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

	// 最近任务（project_id 供前端跳转任务详情）
	recentM := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, project_id, title, status, updated_at").
		Order("updated_at DESC").Limit(5)
	if memberOnly {
		recentM = recentM.Where(taskScope, uid)
	}
	err = recentM.Scan(&res.RecentTasks)

	return
}

// ========== 审计日志 ==========

// ExportAuditLogs 审计日志 CSV 导出：复用列表筛选口径，上限 10000 行。
// 走临时文件 + ServeFileDownload 直出（绕开 JSON 响应封装），BOM 保证
// Excel 直接打开不乱码
func (s *sPlatform) ExportAuditLogs(ctx context.Context, r *ghttp.Request, req *api.AuditLogExportReq) error {
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
	var rows []struct {
		Id         int
		ActorId    int
		ActorType  string
		Action     string
		TargetType string
		TargetId   int
		TargetName string
		IpAddress  string
		UserAgent  string
		CreatedAt  string
	}
	if err := m.Order("id DESC").Limit(10000).Scan(&rows); err != nil {
		return liberr.WrapDb(ctx, err, "查询审计日志失败")
	}
	var b strings.Builder
	b.WriteString("\ufeff" + "id,时间,操作者ID,操作者类型,动作,目标类型,目标ID,目标明细,IP,User-Agent\n")
	for _, x := range rows {
		b.WriteString(fmt.Sprintf("%d,%s,%d,%s,%s,%s,%d,%s,%s,%s\n",
			x.Id, x.CreatedAt, x.ActorId, x.ActorType, x.Action, x.TargetType,
			x.TargetId, csvField(x.TargetName), csvField(x.IpAddress), csvField(x.UserAgent)))
	}
	tmp, err := os.CreateTemp("", "audit-*.csv")
	if err != nil {
		return fmt.Errorf("创建临时文件失败")
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		return fmt.Errorf("写入临时文件失败")
	}
	tmp.Close()
	r.Response.ServeFileDownload(tmp.Name(), fmt.Sprintf("audit-logs-%s.csv", time.Now().Format("20060102-150405")))
	return nil
}

// csvField RFC4180：含逗号/引号/换行的字段加引号并转义内部引号
func csvField(v string) string {
	if strings.ContainsAny(v, ",\"\r\n") {
		return "\"" + strings.ReplaceAll(v, "\"", "\"\"") + "\""
	}
	return v
}

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

// ========== 使用分析（usage_events 聚合） ==========

// weightedDurationRow 加权分组行：SQLite/MySQL 都没有 percentile 函数，
// 按 (endpoint, method, duration) 分组把行数压到延迟值域内，Go 内用权重算分位
type weightedDurationRow struct {
	Endpoint   string
	Method     string
	Client     string
	DurationMs int
	Cnt        int
	ErrCnt     int
	LastAt     string
}

func (s *sPlatform) UsageOverview(ctx context.Context, req *api.UsageOverviewReq) (res *api.UsageOverviewRes, err error) {
	days := req.Days
	if days <= 0 || days > 30 {
		days = 7
	}
	before := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")

	res = &api.UsageOverviewRes{
		WindowDays: days,
		Endpoints:  []api.UsageEndpointStat{},
		Errors:     []api.UsageErrorItem{},
	}

	// 命令热度：按端点×方法×客户端×延迟值分组（加权），行数=延迟值域而非明细数
	rows := []weightedDurationRow{}
	err = g.DB().Model("usage_events").Ctx(ctx).
		Fields("endpoint, method, client, duration_ms, COUNT(*) AS cnt, MAX(created_at) AS last_at").
		Where("created_at >= ?", before).
		Where("endpoint != ''").
		Group("endpoint, method, client, duration_ms").
		Scan(&rows)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "使用统计查询失败")
	}

	type agg struct {
		st          api.UsageEndpointStat
		durations   []int // 按出现次数展开（加权）；延迟毫秒通常 <10^4，总量可控
		errSum      int
		durSum      int
		lastAtMax   string
	}
	byKey := map[string]*agg{}
	for _, r := range rows {
		key := r.Method + " " + r.Endpoint
		a, ok := byKey[key]
		if !ok {
			a = &agg{st: api.UsageEndpointStat{Endpoint: r.Endpoint, Method: r.Method, LastUsed: r.LastAt}}
			byKey[key] = a
		}
		if r.Client == "cli" {
			a.st.CliCount += r.Cnt
			res.CliCalls += r.Cnt
		} else {
			a.st.WebCount += r.Cnt
			res.WebCalls += r.Cnt
		}
		a.st.Total += r.Cnt
		a.st.ErrCount += r.ErrCnt
		a.errSum += r.ErrCnt
		a.durSum += r.DurationMs * r.Cnt
		for i := 0; i < r.Cnt; i++ {
			a.durations = append(a.durations, r.DurationMs)
		}
		if r.LastAt > a.lastAtMax {
			a.lastAtMax = r.LastAt
			a.st.LastUsed = r.LastAt
		}
	}
	res.TotalCalls = res.CliCalls + res.WebCalls

	for _, a := range byKey {
		st := a.st
		st.LastUsed = a.lastAtMax
		if st.Total > 0 {
			st.ErrorRate = float64(a.st.ErrCount) / float64(st.Total)
			st.AvgMs = a.durSum / st.Total
		}
		st.P50Ms = percentile(a.durations, 50)
		st.P95Ms = percentile(a.durations, 95)
		res.Endpoints = append(res.Endpoints, st)
	}
	// 热度倒序：调用最多的排前面
	sort.Slice(res.Endpoints, func(i, j int) bool { return res.Endpoints[i].Total > res.Endpoints[j].Total })

	// 错误 TopN：端点×错误码分组（错误行必须带错误才进结果）
	errRows := []struct {
		Endpoint   string
		Method     string
		ErrorCode  int
		StatusCode int
		Cnt        int
		LastAt     string
	}{}
	err = g.DB().Model("usage_events").Ctx(ctx).
		Fields("endpoint, method, error_code, status_code, COUNT(*) AS cnt, MAX(created_at) AS last_at").
		Where("created_at >= ?", before).
		Where("endpoint != ''").
		Where("(error_code != 0 OR status_code >= 400)").
		Group("endpoint, method, error_code, status_code").
		Order("cnt DESC").Limit(20).
		Scan(&errRows)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "错误统计查询失败")
	}
	for _, r := range errRows {
		res.Errors = append(res.Errors, api.UsageErrorItem{
			Endpoint: r.Endpoint, Method: r.Method,
			ErrorCode: r.ErrorCode, StatusCode: r.StatusCode,
			Count: r.Cnt, LastSeen: r.LastAt,
		})
	}
	return
}

// percentile 就地展开数组的分位值（p∈(0,100]）；空数组返回 0
func percentile(arr []int, p int) int {
	if len(arr) == 0 {
		return 0
	}
	sort.Ints(arr)
	idx := len(arr) * p / 100
	if idx >= len(arr) {
		idx = len(arr) - 1
	}
	return arr[idx]
}
