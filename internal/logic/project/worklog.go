package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/cicbyte/byte-code/utility/activity"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

// 工作日志：项目演化叙事（「做了什么、结果如何」），日期分组时间轴
// 呈现。权限：读=项目成员+绑定 agent；撰写 agent 走 worklog 能力位；
// 编辑/删除 = 作者或 maintainer。草稿生成是只读汇总，不落库。

// worklogGate 写操作门禁：agent 需 worklog 能力位（人类直通）
func worklogGate(ctx context.Context, projectId int) error {
	return perm.AgentTaskGate(ctx, projectId, "worklog")
}

// worklogEditGate 编辑/删除：作者或 maintainer（与讨论区同构）
func worklogEditGate(ctx context.Context, row gdb.Record) error {
	uid := ctx.Value("userId")
	userId := 0
	if uid != nil {
		userId = uid.(int)
	}
	if row["author_id"].Int() == userId {
		return nil
	}
	if perm.IsProjectMaintainer(ctx, userId, row["project_id"].Int()) {
		return nil
	}
	return fmt.Errorf("无权限：仅作者或项目 maintainer 可操作")
}

func (s *sProject) CreateWorklog(ctx context.Context, req *api.WorklogCreateReq) (id int, err error) {
	if err := worklogGate(ctx, req.ProjectId); err != nil {
		return 0, err
	}
	userId := ctx.Value("userId").(int)
	source := req.Source
	if source == "" {
		source = "manual"
	}
	result, err := g.DB().Model("worklogs").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"author_id":  userId,
		"content":    req.Content,
		"source":     source,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "记录工作日志失败")
	}
	lastId, _ := result.LastInsertId()
	activity.Record(ctx, activity.ActivityInput{
		ActorID: userId, ActorType: actorType(ctx, userId), ActorName: actorName(ctx, userId),
		Action: "worklog.created", TargetType: "worklog", TargetID: int(lastId),
		TargetName: firstLine(req.Content, 60), ProjectID: req.ProjectId,
	})
	return int(lastId), nil
}

func (s *sProject) ListWorklogs(ctx context.Context, req *api.WorklogListReq) (total int, list []api.WorklogItem, err error) {
	m := g.DB().Model("worklogs").Ctx(ctx).Where("project_id", req.ProjectId)
	total, err = m.Count()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "统计工作日志失败")
	}
	pageNum, pageSize := req.PageNum, req.PageSize
	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	rows, err := m.Fields("*").Order("id DESC").Page(pageNum, pageSize).All()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询工作日志失败")
	}
	list = []api.WorklogItem{}
	for _, r := range rows {
		list = append(list, api.WorklogItem{
			Id: r["id"].Int(), ProjectId: r["project_id"].Int(),
			AuthorId: r["author_id"].Int(),
			AuthorName: actorName(ctx, r["author_id"].Int()),
			AuthorType: actorType(ctx, r["author_id"].Int()),
			Content:   r["content"].String(), Source: r["source"].String(),
			CreatedAt: r["created_at"].String(), UpdatedAt: r["updated_at"].String(),
		})
	}
	return total, list, nil
}

func (s *sProject) UpdateWorklog(ctx context.Context, req *api.WorklogUpdateReq) (err error) {
	row, err := g.DB().Model("worklogs").Ctx(ctx).Where("id", req.Id).One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询工作日志失败")
	}
	if row.IsEmpty() {
		return fmt.Errorf("日志不存在")
	}
	if err := worklogEditGate(ctx, row); err != nil {
		return err
	}
	if _, err := g.DB().Model("worklogs").Ctx(ctx).Where("id", req.Id).
		Data(g.Map{"content": req.Content, "updated_at": time.Now().Format("2006-01-02 15:04:05")}).
		Update(); err != nil {
		return liberr.WrapDb(ctx, err, "更新工作日志失败")
	}
	return nil
}

func (s *sProject) DeleteWorklog(ctx context.Context, id int) (err error) {
	row, err := g.DB().Model("worklogs").Ctx(ctx).Where("id", id).One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询工作日志失败")
	}
	if row.IsEmpty() {
		return fmt.Errorf("日志不存在")
	}
	if err := worklogEditGate(ctx, row); err != nil {
		return err
	}
	if _, err := g.DB().Model("worklogs").Ctx(ctx).Where("id", id).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "删除工作日志失败")
	}
	return nil
}

// WorklogDraft 汇总时间范围内已完成的任务（标题+执行者+产物）与发布
// （版本+说明）为草稿文本——不落库，返回给前端进编辑器人工润色。
func (s *sProject) WorklogDraft(ctx context.Context, req *api.WorklogDraftReq) (res *api.WorklogDraftRes, err error) {
	from, err := time.ParseInLocation("2006-01-02", req.From, time.Local)
	if err != nil {
		return nil, fmt.Errorf("起始日期格式应为 YYYY-MM-DD")
	}
	to, err := time.ParseInLocation("2006-01-02", req.To, time.Local)
	if err != nil {
		return nil, fmt.Errorf("结束日期格式应为 YYYY-MM-DD")
	}
	if to.Before(from) {
		return nil, fmt.Errorf("结束日期不能早于起始日期")
	}
	// 范围上限 92 天：草稿是给人润色的近期总结，不是年度报表
	if to.Sub(from) > 92*24*time.Hour {
		return nil, fmt.Errorf("时间范围不能超过 92 天")
	}
	dayEnd := to.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	dayStart := from.Format("2006-01-02 15:04:05")

	res = &api.WorklogDraftRes{Content: ""}

	// 已完成任务：completed_at 落在范围内即算（complete 时写入、后续
	// 编辑不重写，见 task_ops.go）；review（待审）也算已完成工作，
	// 驳回回 in_progress 的自然被状态过滤排除
	tasks, err := g.DB().Model("tasks").Ctx(ctx).
		Fields("id, title, assignee_id, artifacts, completed_at, status").
		Where("project_id", req.ProjectId).
		Where("completed_at >= ?", dayStart).
		Where("completed_at < ?", dayEnd).
		Where("status IN ('review','done','closed')").
		Order("completed_at ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "汇总任务失败")
	}

	// 期内发布：版本节点是演化路径的锚点
	releases, err := g.DB().Model("project_releases").Ctx(ctx).
		Fields("version, title, notes, created_at").
		Where("project_id", req.ProjectId).
		Where("created_at >= ?", dayStart).
		Where("created_at < ?", dayEnd).
		Order("created_at ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "汇总发布失败")
	}

	var b strings.Builder
	res.TaskCount = len(tasks)
	res.ReleaseCount = len(releases)
	if len(tasks) == 0 && len(releases) == 0 {
		b.WriteString("该时间段内没有已完成的任务或发布。")
		res.Content = b.String()
		return res, nil
	}

	if len(releases) > 0 {
		fmt.Fprintf(&b, "### 发布（%d）\n", len(releases))
		for _, r := range releases {
			line := fmt.Sprintf("- **v%s** %s", r["version"].String(), r["title"].String())
			if notes := strings.TrimSpace(r["notes"].String()); notes != "" {
				line += "：" + oneLine(notes, 200)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
	if len(tasks) > 0 {
		fmt.Fprintf(&b, "### 已完成任务（%d）\n", len(tasks))
		for i, t := range tasks {
			who := actorName(ctx, t["assignee_id"].Int())
			at := strings.TrimSpace(t["completed_at"].String())
			if len(at) >= 16 {
				at = at[11:16]
			}
			whoPart := ""
			if who != "" {
				whoPart = " @" + who
			}
			fmt.Fprintf(&b, "%d. 「%s」%s（%s 完成）\n", i+1, t["title"].String(), whoPart, at)
			if art := strings.TrimSpace(t["artifacts"].String()); art != "" {
				fmt.Fprintf(&b, "   - 产物：%s\n", oneLine(art, 300))
			}
		}
	}
	res.Content = strings.TrimRight(b.String(), "\n")
	return res, nil
}

// firstLine 取首行（活动流 target_name 用，日志正文常多行）
func firstLine(s string, max int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}

// oneLine 压平多行为单行并截断（artifacts/notes 可能带换行与长路径）
func oneLine(s string, max int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}
