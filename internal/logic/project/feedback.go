package project

// 跨项目反馈实现：投递（A→B，需关联）→ B 的准入 agent 阅读 →
// convert（建任务，血缘回填）或 dismiss（理由回告）。反馈不进任务列表。

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
)

func (s *sProject) CreateFeedback(ctx context.Context, req *api.FeedbackCreateReq) (id int, err error) {
	uid := perm.UserId(ctx)
	// 投递门槛（bcode-cli 反馈 B1 后放宽）：反馈通道的价值正是给无准入方的
	// 唯一口子——已有目标准入的人直接建任务即可，无需反馈。故不再要求发起人
	// 可访问目标项目，防任意投递由唯一条件承担：目标项目必须已关联来源项目
	// （B 的 owner 有意识地开放了反馈面）。
	// 来源项目：来源任务所属项目；无来源任务时取发起人的成员项目——
	// 人类看 project_members，agent 只有 bindings 无成员行（bcode-cli 实测
	// 盲点：agent 不带 --task 投递恒报"无法确定来源项目"），两侧都要兜
	sourceProject := 0
	if req.SourceTaskId > 0 {
		sourceProject = perm.EntityProjectId(ctx, "tasks", req.SourceTaskId)
	}
	if sourceProject == 0 {
		if v, _ := g.DB().Model("project_members").Ctx(ctx).Where("user_id", uid).Order("id ASC").Value("project_id"); v != nil {
			sourceProject = v.Int()
		}
	}
	if sourceProject == 0 {
		if v, _ := g.DB().Model("agent_project_bindings").Ctx(ctx).Where("agent_id", uid).Order("id ASC").Value("project_id"); v != nil {
			sourceProject = v.Int()
		}
	}
	if sourceProject == 0 || sourceProject == req.ProjectId {
		return 0, fmt.Errorf("无法确定来源项目（跨项目反馈需要来源）")
	}
	// 目标项目须已关联来源项目（投递面由 B 的 owner 治理）
	if cnt, _ := g.DB().Model("project_relations").Ctx(ctx).
		Where("project_id", req.ProjectId).
		Where("related_project_id", sourceProject).Count(); cnt == 0 {
		return 0, fmt.Errorf("目标项目未关联来源项目，不能投递（请对方项目管理员先建立关联）")
	}
	result, err := g.DB().Model("project_feedbacks").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId, "source_project_id": sourceProject,
		"source_task_id": req.SourceTaskId, "title": req.Title, "content": req.Content,
		"status": "open", "created_by": uid,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "投递反馈失败")
	}
	lastId, _ := result.LastInsertId()

	// 通知目标项目的准入 agents 与人类成员：像新任务广播一样的事件驱动
	rows, _ := g.DB().Model("project_members pm").Ctx(ctx).
		Fields("pm.user_id").Where("pm.project_id", req.ProjectId).All()
	for _, r := range rows {
		notify.Send(ctx, r["user_id"].Int(), "收到跨项目反馈",
			fmt.Sprintf("「%s」请分析是否建立任务", req.Title), "info", "project", req.ProjectId)
	}
	agentRows, _ := g.DB().Model("agent_project_bindings b").Ctx(ctx).
		Fields("b.agent_id").Where("b.project_id", req.ProjectId).All()
	for _, r := range agentRows {
		notify.Send(ctx, r["agent_id"].Int(), "收到跨项目反馈",
			fmt.Sprintf("「%s」请阅读分析（feedback %d），决定是否建任务", req.Title, int(lastId)),
			"info", "project", req.ProjectId)
	}
	return int(lastId), nil
}

func (s *sProject) ListFeedbacks(ctx context.Context, req *api.FeedbackListReq) (res *api.FeedbackListRes, err error) {
	res = &api.FeedbackListRes{List: []api.FeedbackItem{}}
	m := g.DB().Model("project_feedbacks f").Ctx(ctx).
		LeftJoin("projects p", "p.id = f.source_project_id").
		Where("f.project_id", req.ProjectId)
	if req.Status != "all" {
		m = m.Where("f.status", req.Status)
	}
	total, err := m.Clone().Count()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询反馈失败")
	}
	res.Total = total
	rows, err := m.Fields("f.id, f.title, f.content, f.status, f.source_project_id, p.name as source_project_name, f.source_task_id, f.converted_task_id, f.dismiss_reason, f.handled_by, f.created_at").
		Page(req.Page, req.Size).Order("f.id DESC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询反馈失败")
	}
	for _, r := range rows {
		res.List = append(res.List, api.FeedbackItem{
			Id: r["id"].Int(), Title: r["title"].String(), Content: r["content"].String(),
			Status: r["status"].String(), SourceProjectId: r["source_project_id"].Int(),
			SourceProjectName: r["source_project_name"].String(), SourceTaskId: r["source_task_id"].Int(),
			ConvertedTaskId: r["converted_task_id"].Int(), DismissReason: r["dismiss_reason"].String(),
			HandledBy: r["handled_by"].Int(), CreatedAt: r["created_at"].String(),
		})
	}
	return res, nil
}

// ConvertFeedback 反馈转任务：建 B 项目任务（默认 quiet 不广播——由处理者
// 自主决策的产物，处理者自己认领或指派），血缘回填并回告发起方
func (s *sProject) ConvertFeedback(ctx context.Context, req *api.FeedbackConvertReq) (taskId int, err error) {
	uid := perm.UserId(ctx)
	fb, err := g.DB().Model("project_feedbacks").Ctx(ctx).Where("id", req.Id).Where("project_id", req.ProjectId).One()
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "查询反馈失败")
	}
	if fb.IsEmpty() {
		return 0, fmt.Errorf("反馈不存在")
	}
	if fb["status"].String() != "open" {
		return 0, fmt.Errorf("反馈已处理（%s）", fb["status"].String())
	}
	title := req.Title
	if title == "" {
		title = fb["title"].String()
	}
	// Type 必须显式给值：gf 的 d: 默认值只在 HTTP 绑定生效，代码内构造
	// 的 req 是零值（空串违反 tasks.type 的 CHECK 约束）
	taskId, err = s.CreateTask(ctx, &api.TaskCreateReq{
		ProjectId: req.ProjectId, Title: title, Type: "chore",
		Description: fmt.Sprintf("来自跨项目反馈 #%d（项目 #%d）：\n\n%s", req.Id, fb["source_project_id"].Int(), fb["content"].String()),
		Quiet:       true,
	})
	if err != nil {
		return 0, err
	}
	// 状态流转 + 血缘：新任务反向引用来源任务（跨项目引用链）
	if fb["source_task_id"].Int() > 0 {
		refs := fmt.Sprintf(`[{"type":"task","projectId":%d,"id":%d,"title":""}]`,
			fb["source_project_id"].Int(), fb["source_task_id"].Int())
		if _, uerr := g.DB().Model("tasks").Ctx(ctx).Where("id", taskId).Data(g.Map{"related_refs": refs}).Update(); uerr != nil {
			g.Log().Warningf(ctx, "feedback ref link failed: %v", uerr)
		}
	}
	now := gtimeNow()
	if _, err = g.DB().Model("project_feedbacks").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"status": "converted", "converted_task_id": taskId, "handled_by": uid, "handled_at": now,
	}).Update(); err != nil {
		return 0, liberr.WrapDb(ctx, err, "反馈状态更新失败")
	}
	notify.Send(ctx, fb["created_by"].Int(), "反馈已转化",
		fmt.Sprintf("你的反馈「%s」已转化为任务 #%d", fb["title"].String(), taskId), "success", "task", taskId)
	s.recordActivity(ctx, uid, "feedback.converted", "task", taskId, title, req.ProjectId, fmt.Sprintf("由反馈 #%d 转化", req.Id))
	return taskId, nil
}

func (s *sProject) DismissFeedback(ctx context.Context, req *api.FeedbackDismissReq) (err error) {
	uid := perm.UserId(ctx)
	fb, err := g.DB().Model("project_feedbacks").Ctx(ctx).Where("id", req.Id).Where("project_id", req.ProjectId).One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询反馈失败")
	}
	if fb.IsEmpty() {
		return fmt.Errorf("反馈不存在")
	}
	if fb["status"].String() != "open" {
		return fmt.Errorf("反馈已处理")
	}
	if _, err = g.DB().Model("project_feedbacks").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"status": "dismissed", "dismiss_reason": req.Reason, "handled_by": uid, "handled_at": gtimeNow(),
	}).Update(); err != nil {
		return liberr.WrapDb(ctx, err, "反馈状态更新失败")
	}
	notify.Send(ctx, fb["created_by"].Int(), "反馈被忽略",
		fmt.Sprintf("你的反馈「%s」未被采纳：%s", fb["title"].String(), req.Reason), "warning", "project", req.ProjectId)
	s.recordActivity(ctx, uid, "feedback.dismissed", "feedback", req.Id, fb["title"].String(), req.ProjectId, req.Reason)
	return nil
}

func gtimeNow() string { return time.Now().Format("2006-01-02 15:04:05") }
