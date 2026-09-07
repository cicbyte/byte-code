package project

// 专题（long-task）实现：创建（关联 PRD 文档）→ 阶段拆解（整体替换导入）→
// agent 循环推进（toggle/log/handoff）→ 人终验收。阶段可转日常任务。
// 持续运行的驱动在 agent 侧（拉模式），平台不调度。

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
)

func (s *sProject) CreateTopic(ctx context.Context, req *api.TopicCreateReq) (id int, err error) {
	uid := perm.UserId(ctx)
	// PRD 文档关联校验：路径须存在于本项目文档索引
	if req.DocPath != "" {
		if cnt, _ := g.DB().Model("project_document_index").Ctx(ctx).
			Where("project_id", req.ProjectId).Where("path", req.DocPath).Count(); cnt == 0 {
			return 0, fmt.Errorf("关联文档不存在：%s（先在知识库创建 PRD）", req.DocPath)
		}
	}
	// 执行者须为 ai 账号（专题是 agent 的持续工程；人创建人验收）
	if req.AssigneeId > 0 {
		if t, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", req.AssigneeId).Fields("type").Value(); t == nil || t.String() != "ai" {
			return 0, fmt.Errorf("执行者必须是 Agent 账号")
		}
	}
	result, err := g.DB().Model("topics").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId, "title": req.Title,
		"goal": req.Goal, "acceptance": req.Acceptance, "doc_path": req.DocPath,
		"assignee_id": req.AssigneeId, "status": "active", "created_by": uid,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建专题失败")
	}
	lastId, _ := result.LastInsertId()
	s.recordActivity(ctx, uid, "topic.created", "topic", int(lastId), req.Title, req.ProjectId, "")
	if req.AssigneeId > 0 {
		notify.Send(ctx, req.AssigneeId, "分到专题",
			fmt.Sprintf("「%s」已分配给你，bcode topic work 可开始推进", req.Title), "info", "project", req.ProjectId)
	}
	return int(lastId), nil
}

func (s *sProject) topicRow(ctx context.Context, projectId, id int) (gdb.Record, error) {
	row, err := g.DB().Model("topics").Ctx(ctx).Where("id", id).Where("project_id", projectId).One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询专题失败")
	}
	if row.IsEmpty() {
		return nil, fmt.Errorf("专题不存在")
	}
	return row, nil
}

func (s *sProject) fillTopic(ctx context.Context, m gdb.Record) api.TopicItem {
	item := api.TopicItem{
		Id: m["id"].Int(), Title: m["title"].String(),
		Goal: m["goal"].String(), Acceptance: m["acceptance"].String(),
		DocPath:    m["doc_path"].String(),
		AssigneeId: m["assignee_id"].Int(), Status: m["status"].String(),
		CreatedAt: m["created_at"].String(), CompletedAt: m["completed_at"].String(),
		Phases: []api.TopicPhaseBrief{},
	}
	if v, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", item.AssigneeId).Fields("real_name, username").One(); !v.IsEmpty() {
		item.AssigneeName = v["real_name"].String()
		if item.AssigneeName == "" {
			item.AssigneeName = v["username"].String()
		}
	}
	phases, _ := g.DB().Model("topic_phases").Ctx(ctx).
		Where("topic_id", item.Id).Order("sort_order ASC, id ASC").All()
	// 转出任务的实况一次取齐（专题内直读）
	taskIds := g.Slice{}
	for _, p := range phases {
		if p["task_id"].Int() > 0 {
			taskIds = append(taskIds, p["task_id"].Int())
		}
	}
	taskMap := map[int]string{}
	taskStatus := map[int]string{}
	if len(taskIds) > 0 {
		if tRows, _ := g.DB().Model("tasks").Ctx(ctx).WhereIn("id", taskIds).
			Fields("id, title, status, assignee_id").All(); tRows != nil {
			assigneeIds := g.Slice{}
			for _, tr := range tRows {
				if tr["assignee_id"].Int() > 0 {
					assigneeIds = append(assigneeIds, tr["assignee_id"].Int())
				}
			}
			nameMap := map[int]string{}
			if len(assigneeIds) > 0 {
				if uRows, _ := g.DB().Model("sys_users").Ctx(ctx).WhereIn("id", assigneeIds).
					Fields("id, username, real_name").All(); uRows != nil {
					for _, ur := range uRows {
						n := ur["real_name"].String()
						if n == "" {
							n = ur["username"].String()
						}
						nameMap[ur["id"].Int()] = n
					}
				}
			}
			for _, tr := range tRows {
				tid := tr["id"].Int()
				taskMap[tid] = tr["title"].String()
				taskStatus[tid] = tr["status"].String() + "|" + nameMap[tr["assignee_id"].Int()]
			}
		}
	}
	for _, p := range phases {
		brief := api.TopicPhaseBrief{
			Id: p["id"].Int(), Title: p["title"].String(), Detail: p["detail"].String(),
			Status: p["status"].String(), TaskId: p["task_id"].Int(), SortOrder: p["sort_order"].Int(),
			AssigneeId: p["assignee_id"].Int(), Artifacts: p["artifacts"].String(),
		}
		if brief.AssigneeId > 0 {
			if ur, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", brief.AssigneeId).Fields("username, real_name").One(); !ur.IsEmpty() {
				brief.AssigneeName = ur["real_name"].String()
				if brief.AssigneeName == "" {
					brief.AssigneeName = ur["username"].String()
				}
			}
		}
		if brief.TaskId > 0 {
			brief.TaskTitle = taskMap[brief.TaskId]
			if parts := strings.SplitN(taskStatus[brief.TaskId], "|", 2); len(parts) == 2 {
				brief.TaskStatus = parts[0]
				brief.TaskAssignee = parts[1]
			}
		}
		item.Phases = append(item.Phases, brief)
		if p["status"].String() == "done" {
			item.PhaseDone++
		}
	}
	item.PhaseTotal = len(item.Phases)
	// 最近一次 handoff（下个会话恢复点）
	if v, _ := g.DB().Model("ai_execution_logs").Ctx(ctx).
		Where("topic_id", item.Id).Where("action", "handoff").
		Order("id DESC").Fields("detail").Value(); v != nil {
		item.LastHandoff = v.String()
	}
	return item
}

func (s *sProject) ListTopics(ctx context.Context, req *api.TopicListReq) (res *api.TopicListRes, err error) {
	res = &api.TopicListRes{List: []api.TopicItem{}}
	m := g.DB().Model("topics").Ctx(ctx).Where("project_id", req.ProjectId)
	if req.Status != "all" {
		m = m.Where("status", req.Status)
	}
	rows, err := m.Order("id DESC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询专题失败")
	}
	for _, r := range rows {
		res.List = append(res.List, s.fillTopic(ctx, r))
	}
	return res, nil
}

func (s *sProject) GetTopic(ctx context.Context, projectId, id int) (res *api.TopicDetailRes, err error) {
	m, err := s.topicRow(ctx, projectId, id)
	if err != nil {
		return nil, err
	}
	res = &api.TopicDetailRes{}
	res.TopicItem = s.fillTopic(ctx, m)
	return res, nil
}

// UpsertTopicPhases 批量写入阶段（整体替换）：agent 读 PRD 拆解后导入的入口。
// 已转出任务的阶段保留（task_id 不丢）——拆了单的不能被清单重建冲掉。
func (s *sProject) UpsertTopicPhases(ctx context.Context, req *api.TopicPhaseUpsertReq) (err error) {
	if _, err = s.topicRow(ctx, req.ProjectId, req.Id); err != nil {
		return err
	}
	kept, _ := g.DB().Model("topic_phases").Ctx(ctx).
		Where("topic_id", req.Id).Where("task_id > 0").Order("sort_order ASC, id ASC").All()
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, e := tx.Ctx(ctx).Exec("DELETE FROM topic_phases WHERE topic_id = ? AND task_id = 0", req.Id); e != nil {
			return e
		}
		order := len(kept)
		for _, p := range req.Phases {
			order++
			if _, e := tx.Ctx(ctx).Model("topic_phases").Insert(g.Map{
				"topic_id": req.Id, "title": p.Title, "detail": p.Detail, "sort_order": order,
			}); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "写入阶段失败")
	}
	return nil
}

// AppendTopicPhase 手动追加单个阶段（排到清单末尾）
func (s *sProject) AppendTopicPhase(ctx context.Context, req *api.TopicPhaseAddReq) (id int, err error) {
	if _, err = s.topicRow(ctx, req.ProjectId, req.Id); err != nil {
		return 0, err
	}
	maxOrder := 0
	if v, _ := g.DB().Model("topic_phases").Ctx(ctx).Where("topic_id", req.Id).Fields("MAX(sort_order)").Value(); v != nil {
		maxOrder = v.Int()
	}
	result, err := g.DB().Model("topic_phases").Ctx(ctx).Insert(g.Map{
		"topic_id": req.Id, "title": req.Title, "detail": req.Detail, "sort_order": maxOrder + 1,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "添加阶段失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

// UpdateTopicPhase 编辑阶段字段（指针语义，nil 不更新）
func (s *sProject) UpdateTopicPhase(ctx context.Context, req *api.TopicPhaseUpdateReq) (err error) {
	if _, err = s.topicRow(ctx, req.ProjectId, req.TopicId); err != nil {
		return err
	}
	data := g.Map{}
	if req.Title != nil && *req.Title != "" {
		data["title"] = *req.Title
	}
	if req.Detail != nil {
		data["detail"] = *req.Detail
	}
	if req.AssigneeId != nil {
		data["assignee_id"] = *req.AssigneeId
	}
	if req.Artifacts != nil {
		data["artifacts"] = *req.Artifacts
	}
	if len(data) == 0 {
		return nil
	}
	result, err := g.DB().Model("topic_phases").Ctx(ctx).
		Where("id", req.PhaseId).Where("topic_id", req.TopicId).Data(data).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新阶段失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("阶段不存在")
	}
	return nil
}

// DeleteTopicPhase 删除阶段（已转出任务的历史阶段保血缘不可删）
func (s *sProject) DeleteTopicPhase(ctx context.Context, projectId, topicId, phaseId int) (err error) {
	if _, err = s.topicRow(ctx, projectId, topicId); err != nil {
		return err
	}
	ph, _ := g.DB().Model("topic_phases").Ctx(ctx).Where("id", phaseId).Where("topic_id", topicId).One()
	if ph.IsEmpty() {
		return fmt.Errorf("阶段不存在")
	}
	if ph["task_id"].Int() > 0 {
		return fmt.Errorf("已转出任务的阶段不能删除（保留血缘）")
	}
	if _, err = g.DB().Model("topic_phases").Ctx(ctx).Where("id", phaseId).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "删除阶段失败")
	}
	return nil
}

// ToggleTopicPhase 阶段推进（agent/成员打勾即进展）
func (s *sProject) ToggleTopicPhase(ctx context.Context, req *api.TopicPhaseToggleReq) (err error) {
	if _, err = s.topicRow(ctx, req.ProjectId, req.TopicId); err != nil {
		return err
	}
	result, err := g.DB().Model("topic_phases").Ctx(ctx).
		Where("id", req.PhaseId).Where("topic_id", req.TopicId).
		Data(g.Map{"status": req.Status}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "更新阶段失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("阶段不存在")
	}
	return nil
}

// ConvertTopicPhase 阶段转日常任务：落任务池可被认领；阶段标记 done 并回填
// task_id（血缘）；任务 related_refs 反向指向——不，任务不知道 topic 实体，
// 血缘用描述注明即可（一期不扩展 refs 类型）
func (s *sProject) ConvertTopicPhase(ctx context.Context, req *api.TopicPhaseConvertReq) (taskId int, err error) {
	ph, err := g.DB().Model("topic_phases").Ctx(ctx).
		Where("id", req.PhaseId).Where("topic_id", req.TopicId).One()
	if err != nil || ph.IsEmpty() {
		return 0, fmt.Errorf("阶段不存在")
	}
	tp, err := s.topicRow(ctx, req.ProjectId, req.TopicId)
	if err != nil {
		return 0, err
	}
	taskId, err = s.CreateTask(ctx, &api.TaskCreateReq{
		ProjectId: req.ProjectId, Title: ph["title"].String(), Type: "feature",
		Description: fmt.Sprintf("来自专题「%s」阶段：%s\n%s", tp["title"], ph["title"].String(), ph["detail"].String()),
	})
	if err != nil {
		return 0, err
	}
	if _, err = g.DB().Model("topic_phases").Ctx(ctx).Where("id", req.PhaseId).
		Data(g.Map{"task_id": taskId}).Update(); err != nil {
		return 0, liberr.WrapDb(ctx, err, "阶段回填失败")
	}
	return taskId, nil
}

// LogTopic 专题留痕：progress（执行进展）或 handoff（交接摘要，
// 下个会话的恢复点）。agent 侧免 task 维度，走 topic_id。
func (s *sProject) LogTopic(ctx context.Context, req *api.TopicLogReq) (err error) {
	tp, err := s.topicRow(ctx, req.ProjectId, req.Id)
	if err != nil {
		return err
	}
	uid := perm.UserId(ctx)
	if _, err = g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"task_id": 0, "topic_id": req.Id, "ai_user_id": uid,
		"action": req.Action, "detail": req.Detail,
		"status": map[string]string{"progress": "success", "handoff": "success"}[req.Action],
	}); err != nil {
		return liberr.WrapDb(ctx, err, "留痕失败")
	}
	if req.Action == "handoff" {
		// 交接即里程碑信号：通知创建者（人可周期性只看 handoff 掌握进度）
		notify.Send(ctx, tp["created_by"].Int(), "专题交接摘要",
			fmt.Sprintf("「%s」：%s", tp["title"].String(), req.Detail), "info", "project", req.ProjectId)
	}
	return nil
}

// FinishTopic 人终验收：completed（对照 acceptance）或 abandoned
func (s *sProject) FinishTopic(ctx context.Context, req *api.TopicFinishReq) (err error) {
	tp, err := s.topicRow(ctx, req.ProjectId, req.Id)
	if err != nil {
		return err
	}
	if tp["status"].String() != "active" {
		return fmt.Errorf("专题已终态")
	}
	uid := perm.UserId(ctx)
	if uid != tp["created_by"].Int() && !perm.IsProjectOwner(ctx, uid, req.ProjectId) {
		return fmt.Errorf("仅专题创建者或项目管理员可终验收")
	}
	if _, err = g.DB().Model("topics").Ctx(ctx).Where("id", req.Id).Data(g.Map{
		"status": req.Result, "completed_at": time.Now().Format("2006-01-02 15:04:05"),
	}).Update(); err != nil {
		return liberr.WrapDb(ctx, err, "终验收失败")
	}
	s.recordActivity(ctx, uid, "topic."+req.Result, "topic", req.Id, tp["title"].String(), req.ProjectId, "")
	if req.Result == "completed" && tp["assignee_id"].Int() > 0 {
		notify.Send(ctx, tp["assignee_id"].Int(), "专题验收通过",
			fmt.Sprintf("「%s」已通过人验收，感谢推进", tp["title"].String()), "success", "project", req.ProjectId)
	}
	return nil
}
