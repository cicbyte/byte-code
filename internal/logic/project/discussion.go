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
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

// 讨论区：想法/议题线程，论坛式回复；不进任务工作流，
// 成熟后 convert 转任务（status=converted + converted_task_id 血缘互链）。
// 权限：读=项目成员+绑定 agent（中间件一跳解析）；发起/回复 agent 走
// discuss 能力位；编辑/删除/归档/转化 = 作者或 maintainer。

// actorName 取账号显示名（real_name 优先，回退 username）
func actorName(ctx context.Context, userId int) string {
	if userId <= 0 {
		return ""
	}
	v, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", userId).
		Fields("COALESCE(NULLIF(real_name, ''), username)").Value()
	if v == nil {
		return ""
	}
	return v.String()
}

// actorType 服务端按账号类型推导（ai 账号经 Key 登录记为 ai），不信任客户端
func actorType(ctx context.Context, userId int) string {
	v, _ := g.DB().Model("sys_users").Ctx(ctx).Where("id", userId).Fields("type").Value()
	if v != nil && v.String() == "ai" {
		return "ai"
	}
	return "human"
}

// discussionGate 讨论写操作门禁：agent 需 discuss 能力位（人类直通）
func discussionGate(ctx context.Context, projectId int) error {
	return perm.AgentTaskGate(ctx, projectId, "discuss")
}

// discussionEditGate 编辑/删除/归档/转化：作者或 maintainer
func discussionEditGate(ctx context.Context, row gdb.Record) error {
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

func (s *sProject) CreateDiscussion(ctx context.Context, req *api.DiscussionCreateReq) (id int, err error) {
	if err := discussionGate(ctx, req.ProjectId); err != nil {
		return 0, err
	}
	userId := ctx.Value("userId").(int)
	result, err := g.DB().Model("discussions").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"title":      req.Title,
		"body":       req.Body,
		"status":     "open",
		"author_id":  userId,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建讨论失败")
	}
	lastId, _ := result.LastInsertId()
	activity.Record(ctx, activity.ActivityInput{
		ActorID: userId, ActorType: actorType(ctx, userId), ActorName: actorName(ctx, userId),
		Action: "discussion.created", TargetType: "discussion", TargetID: int(lastId),
		TargetName: req.Title, ProjectID: req.ProjectId,
	})
	return int(lastId), nil
}

func (s *sProject) ListDiscussions(ctx context.Context, req *api.DiscussionListReq) (total int, list []api.DiscussionItem, err error) {
	m := g.DB().Model("discussions").Ctx(ctx).Where("project_id", req.ProjectId)
	if req.Status != "" {
		m = m.Where("status", req.Status)
	}
	total, err = m.Count()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "统计讨论失败")
	}
	pageNum, pageSize := req.PageNum, req.PageSize
	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	rows, err := m.Fields("*").Order("id DESC").
		Page(pageNum, pageSize).All()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	list = []api.DiscussionItem{}
	for _, r := range rows {
		item := api.DiscussionItem{
			Id: r["id"].Int(), ProjectId: r["project_id"].Int(),
			Title: r["title"].String(), Body: r["body"].String(),
			Status: r["status"].String(), AuthorId: r["author_id"].Int(),
			CreatedAt: r["created_at"].String(), UpdatedAt: r["updated_at"].String(),
			ConvertedTaskId: r["converted_task_id"].Int(),
		}
		// 回复数一次聚合 + 作者署名（避免 N+1）
		item.AuthorName = actorName(ctx, item.AuthorId)
		item.AuthorType = actorType(ctx, item.AuthorId)
		list = append(list, item)
	}
	if len(list) > 0 {
		ids := make([]int, 0, len(list))
		for _, it := range list {
			ids = append(ids, it.Id)
		}
		cntRows, cerr := g.DB().Model("discussion_replies").Ctx(ctx).
			Fields("discussion_id, COUNT(*) AS cnt").
			WhereIn("discussion_id", ids).Group("discussion_id").All()
		if cerr == nil {
			cntMap := map[int]int{}
			for _, cr := range cntRows {
				cntMap[cr["discussion_id"].Int()] = cr["cnt"].Int()
			}
			for i := range list {
				list[i].ReplyCount = cntMap[list[i].Id]
			}
		}
	}
	return total, list, nil
}

func (s *sProject) DiscussionDetail(ctx context.Context, id int) (res *api.DiscussionDetailRes, err error) {
	row, err := g.DB().Model("discussions").Ctx(ctx).Where("id", id).One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if row.IsEmpty() {
		return nil, fmt.Errorf("讨论不存在")
	}
	res = &api.DiscussionDetailRes{
		Id: row["id"].Int(), ProjectId: row["project_id"].Int(),
		Title: row["title"].String(), Body: row["body"].String(),
		Status: row["status"].String(), AuthorId: row["author_id"].Int(),
		AuthorName: actorName(ctx, row["author_id"].Int()),
		AuthorType: actorType(ctx, row["author_id"].Int()),
		CreatedAt:  row["created_at"].String(), UpdatedAt: row["updated_at"].String(),
		ConvertedTaskId: row["converted_task_id"].Int(),
		Replies:         []api.DiscussionReplyItem{},
	}
	replies, err := g.DB().Model("discussion_replies").Ctx(ctx).
		Where("discussion_id", id).Order("id ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询回复失败")
	}
	for _, r := range replies {
		res.Replies = append(res.Replies, api.DiscussionReplyItem{
			Id: r["id"].Int(), DiscussId: r["discussion_id"].Int(),
			UserId: r["user_id"].Int(), UserName: actorName(ctx, r["user_id"].Int()),
			UserType: r["user_type"].String(), Content: r["content"].String(),
			CreatedAt: r["created_at"].String(),
		})
	}
	return res, nil
}

func (s *sProject) UpdateDiscussion(ctx context.Context, req *api.DiscussionUpdateReq) (err error) {
	row, err := g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if row.IsEmpty() {
		return fmt.Errorf("讨论不存在")
	}
	if err := discussionEditGate(ctx, row); err != nil {
		return err
	}
	updates := g.Map{"updated_at": time.Now().Format("2006-01-02 15:04:05")}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Body != "" {
		updates["body"] = req.Body
	}
	if _, err := g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).Data(updates).Update(); err != nil {
		return liberr.WrapDb(ctx, err, "更新讨论失败")
	}
	return nil
}

func (s *sProject) DeleteDiscussion(ctx context.Context, id int) (err error) {
	row, err := g.DB().Model("discussions").Ctx(ctx).Where("id", id).One()
	if err != nil {
		return liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if row.IsEmpty() {
		return fmt.Errorf("讨论不存在")
	}
	if err := discussionEditGate(ctx, row); err != nil {
		return err
	}
	if _, err := g.DB().Model("discussions").Ctx(ctx).Where("id", id).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "删除讨论失败")
	}
	if _, err := g.DB().Model("discussion_replies").Ctx(ctx).Where("discussion_id", id).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "删除回复失败")
	}
	return nil
}

func (s *sProject) ArchiveDiscussion(ctx context.Context, id int) (status string, err error) {
	row, err := g.DB().Model("discussions").Ctx(ctx).Where("id", id).One()
	if err != nil {
		return "", liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if row.IsEmpty() {
		return "", fmt.Errorf("讨论不存在")
	}
	if err := discussionEditGate(ctx, row); err != nil {
		return "", err
	}
	next := "archived"
	if row["status"].String() == "archived" {
		next = "open"
	}
	if _, err := g.DB().Model("discussions").Ctx(ctx).Where("id", id).
		Data(g.Map{"status": next, "updated_at": time.Now().Format("2006-01-02 15:04:05")}).Update(); err != nil {
		return "", liberr.WrapDb(ctx, err, "更新讨论失败")
	}
	return next, nil
}

func (s *sProject) CreateDiscussionReply(ctx context.Context, req *api.DiscussionReplyCreateReq) (id int, err error) {
	drow, err := g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).One()
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if drow.IsEmpty() {
		return 0, fmt.Errorf("讨论不存在")
	}
	projectId := drow["project_id"].Int()
	userId := ctx.Value("userId").(int)
	if err := discussionGate(ctx, projectId); err != nil {
		return 0, err
	}
	result, err := g.DB().Model("discussion_replies").Ctx(ctx).Insert(g.Map{
		"discussion_id": req.Id,
		"user_id":       userId,
		"content":       req.Content,
		"user_type":     actorType(ctx, userId),
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "回复失败")
	}
	// 归一 updated_at（MySQL 侧 ON UPDATE 自带，SQLite 侧手动）
	g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).
		Data(g.Map{"updated_at": time.Now().Format("2006-01-02 15:04:05")}).Update()
	// 通知作者（排除自己）
	if author := drow["author_id"].Int(); author > 0 && author != userId {
		notify.Send(ctx, author, "讨论新回复",
			fmt.Sprintf("讨论「%s」有新回复", drow["title"].String()), "info", "discussion", req.Id)
	}
	lastId, _ := result.LastInsertId()
	activity.Record(ctx, activity.ActivityInput{
		ActorID: userId, ActorType: actorType(ctx, userId), ActorName: actorName(ctx, userId),
		Action: "discussion.replied", TargetType: "discussion", TargetID: req.Id,
		TargetName: drow["title"].String(), ProjectID: projectId,
	})
	return int(lastId), nil
}

func (s *sProject) ConvertDiscussion(ctx context.Context, req *api.DiscussionConvertReq) (taskId int, err error) {
	row, err := g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).One()
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "查询讨论失败")
	}
	if row.IsEmpty() {
		return 0, fmt.Errorf("讨论不存在")
	}
	if row["status"].String() == "converted" {
		return 0, fmt.Errorf("该讨论已转为任务 #%d", row["converted_task_id"].Int())
	}
	if err := discussionEditGate(ctx, row); err != nil {
		return 0, err
	}
	userId := ctx.Value("userId").(int)
	projectId := row["project_id"].Int()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = row["title"].String()
	}
	// 任务正文 = 讨论正文 + 血缘标注（任务侧可反查完整讨论）
	desc := row["body"].String() + fmt.Sprintf("\n\n---\n来自讨论 #%d（/project/%d/discussions）", req.Id, projectId)
	result, err := g.DB().Model("tasks").Ctx(ctx).Insert(g.Map{
		"project_id":   projectId,
		"title":        title,
		"description":  desc,
		"type":         req.Type,
		"priority":     3,
		"status":       "open",
		"creator_id":   userId,
		"source":       "human",
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建任务失败")
	}
	lastId, _ := result.LastInsertId()
	taskId = int(lastId)
	if _, err := g.DB().Model("discussions").Ctx(ctx).Where("id", req.Id).
		Data(g.Map{"status": "converted", "converted_task_id": taskId,
			"updated_at": time.Now().Format("2006-01-02 15:04:05")}).Update(); err != nil {
		return 0, liberr.WrapDb(ctx, err, "回写讨论状态失败")
	}
	// 通知讨论作者（转化人=作者时不打扰）
	if author := row["author_id"].Int(); author > 0 && author != userId {
		notify.Send(ctx, author, "讨论已转任务",
			fmt.Sprintf("讨论「%s」已转为任务 #%d", row["title"].String(), taskId), "info", "task", taskId)
	}
	activity.Record(ctx, activity.ActivityInput{
		ActorID: userId, ActorType: actorType(ctx, userId), ActorName: actorName(ctx, userId),
		Action: "discussion.converted", TargetType: "discussion", TargetID: req.Id,
		TargetName: row["title"].String(), ProjectID: projectId,
		Detail: fmt.Sprintf("转为任务 #%d", taskId),
	})
	return taskId, nil
}
