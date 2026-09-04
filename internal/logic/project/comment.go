package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/notify"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

// 评论与 AI 执行日志

func (s *sProject) CreateComment(ctx context.Context, req *api.CommentCreateReq) (id int, err error) {
	userId := 0
	if uid := ctx.Value("userId"); uid != nil {
		userId = uid.(int)
	}

	// user_type 由服务端按账号类型推导（AI 账号经 Key 登录记为 ai），
	// 不信任客户端声明，防止伪造 AI 评论
	userType := "human"
	if v, _ := g.DB().Model("sys_users").Where("id", userId).Fields("type").Value(); v != nil && v.String() == "ai" {
		userType = "ai"
	}

	result, err := g.DB().Model("comments").Ctx(ctx).Insert(g.Map{
		"task_id":   req.TaskId,
		"parent_id": req.ParentId,
		"user_id":   userId,
		"content":   req.Content,
		"user_type": userType,
	})
	// 评论通知（评论本身不阻塞）：通知任务负责人与创建者（排除自己）；
	// 回复时额外通知被回复评论的作者；内容中 @用户名 精确匹配触发提及通知
	if err == nil {
		taskRow, _ := g.DB().Model("tasks").Ctx(ctx).Where("id", req.TaskId).
			Fields("title, assignee_id, creator_id").One()
		if !taskRow.IsEmpty() {
			title := taskRow["title"].String()
			notified := map[int]bool{userId: true}
			for _, target := range []int{taskRow["assignee_id"].Int(), taskRow["creator_id"].Int()} {
				if target > 0 && !notified[target] {
					notified[target] = true
					notify.Send(ctx, target, "任务新评论",
						fmt.Sprintf("任务「%s」有新评论", title), "info", "task", req.TaskId)
				}
			}
			if req.ParentId > 0 {
				if parentAuthor, _ := g.DB().Model("comments").Ctx(ctx).Where("id", req.ParentId).
					Fields("user_id").Value(); parentAuthor != nil {
					t := parentAuthor.Int()
					if t > 0 && !notified[t] {
						notify.Send(ctx, t, "评论被回复",
							fmt.Sprintf("你在任务「%s」的评论被回复", title), "info", "task", req.TaskId)
					}
				}
			}
			authorName := ""
			if an, _ := g.DB().Model("sys_users").Where("id", userId).
				Fields("COALESCE(NULLIF(real_name, ''), username)").Value(); an != nil {
				authorName = an.String()
			}
			notifyMentions(ctx, req.Content, title, authorName, req.TaskId, notified)
		}
	}
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建评论失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

// notifyMentions @提及通知：评论内容与 @用户名 做真实用户名的字符串精确比对
// （避免正则猜测用户名字符集带来的误报/漏报），作者本人与已通知对象不重复打扰
func notifyMentions(ctx context.Context, content, taskTitle, authorName string, taskId int, notified map[int]bool) {
	if !strings.Contains(content, "@") {
		return
	}
	users, err := g.DB().Model("sys_users").Ctx(ctx).
		Fields("id, username").Where("type", "human").All()
	if err != nil {
		return
	}
	for _, u := range users {
		uid := u["id"].Int()
		if uid <= 0 || notified[uid] {
			continue
		}
		if strings.Contains(content, "@"+u["username"].String()) {
			notified[uid] = true
			who := authorName
			if who == "" {
				who = "有人"
			}
			notify.Send(ctx, uid, "评论提及了你",
				fmt.Sprintf("%s 在任务「%s」的评论中提及了你", who, taskTitle), "info", "task", taskId)
		}
	}
}

func (s *sProject) ListComments(ctx context.Context, taskId int) (res *api.CommentListRes, err error) {
	res = &api.CommentListRes{}
	var list []api.CommentItem
	err = g.DB().Model("comments c").Ctx(ctx).
		LeftJoin("sys_users u", "c.user_id = u.id").
		Fields("c.id, c.task_id, c.user_id, u.username, COALESCE(u.real_name, '') as real_name, c.content, c.user_type, c.created_at").
		Where("c.task_id", taskId).
		Order("c.id ASC").
		Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询评论失败")
	}
	res.List = list
	return res, nil
}

// ==================== AI 执行日志 ====================

func (s *sProject) UpdateComment(ctx context.Context, req *api.CommentUpdateReq) (err error) {
	// 仅评论作者可编辑
	uid := perm.UserId(ctx)
	rec, err := g.DB().Model("comments").Ctx(ctx).Where("id", req.Id).Fields("user_id").Value()
	if err != nil || rec == nil {
		return fmt.Errorf("评论不存在")
	}
	if rec.Int() != uid {
		return fmt.Errorf("只能编辑自己的评论")
	}
	_, err = g.DB().Model("comments").Ctx(ctx).
		Where("id", req.Id).
		Data(g.Map{"content": *req.Content, "updated_at": time.Now().Format("2006-01-02 15:04:05")}).
		Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "编辑评论失败")
	}
	return nil
}

func (s *sProject) DeleteComment(ctx context.Context, id int) (err error) {
	// 评论作者或项目 owner 可删除
	uid := perm.UserId(ctx)
	rec, err := g.DB().Model("comments").Ctx(ctx).Where("id", id).Fields("user_id, task_id").One()
	if err != nil || rec == nil {
		return fmt.Errorf("评论不存在")
	}
	if rec["user_id"].Int() != uid {
		// 查任务所属项目，看是否 owner
		taskProject := perm.EntityProjectId(ctx, "tasks", rec["task_id"].Int())
		if !perm.IsProjectOwner(ctx, uid, taskProject) {
			return fmt.Errorf("只能删除自己的评论")
		}
	}
	// 删除子回复
	if _, err := g.DB().Model("comments").Ctx(ctx).Where("parent_id", id).Delete(); err != nil {
		return liberr.WrapDb(ctx, err, "删除子回复失败")
	}
	_, err = g.DB().Model("comments").Ctx(ctx).Where("id", id).Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "删除评论失败")
	}
	return nil
}
func (s *sProject) CreateAiLog(ctx context.Context, req *api.AiLogCreateReq) (id int, err error) {
	result, err := g.DB().Model("ai_execution_logs").Ctx(ctx).Insert(g.Map{
		"task_id":    req.TaskId,
		"ai_user_id": req.AiUserId,
		"action":     req.Action,
		"detail":     req.Detail,
		"status":     req.Status,
	})
	if err != nil {
		return 0, liberr.WrapDb(ctx, err, "创建AI执行日志失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), nil
}

func (s *sProject) ListAiLogs(ctx context.Context, taskId int) (res *api.AiLogListRes, err error) {
	res = &api.AiLogListRes{}
	var list []api.AiLogItem
	err = g.DB().Model("ai_execution_logs l").Ctx(ctx).
		LeftJoin("sys_users u", "u.id = l.ai_user_id").
		Fields("l.*, u.username AS ai_username").
		Where("l.task_id", taskId).
		Order("l.id DESC").
		Scan(&list)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询AI执行日志失败")
	}
	res.List = list
	return res, nil
}
