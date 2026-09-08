package project

// QA 库实现：按问题去重的 upsert（agent 解决问题后重复沉淀同一问题
// 只会更新答案，不堆积条目）；检索按 hits 降序（高频先出）。

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/escape"
	"github.com/cicbyte/byte-code/utility/perm"
)

// UpsertQa 创建/更新：同问题（精确匹配）已存在则更新答案/标签并复活
func (s *sProject) UpsertQa(ctx context.Context, req *api.QaUpsertReq) (id int, updated bool, err error) {
	uid := perm.UserId(ctx)
	existing, _ := g.DB().Model("project_qas").Ctx(ctx).
		Where("project_id", req.ProjectId).Where("question", req.Question).One()
	if !existing.IsEmpty() {
		if _, err = g.DB().Model("project_qas").Ctx(ctx).Where("id", existing["id"].Int()).Data(g.Map{
			"answer": req.Answer, "tags": req.Tags,
			"status": "active", "updated_at": time.Now().Format("2006-01-02 15:04:05"),
			"updated_by": uid,
		}).Update(); err != nil {
			return 0, true, liberr.WrapDb(ctx, err, "更新 QA 失败")
		}
		return existing["id"].Int(), true, nil
	}
	result, err := g.DB().Model("project_qas").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId, "question": req.Question,
		"answer": req.Answer, "tags": req.Tags, "status": "active",
		"created_by": uid, "updated_by": uid,
	})
	if err != nil {
		return 0, false, liberr.WrapDb(ctx, err, "创建 QA 失败")
	}
	lastId, _ := result.LastInsertId()
	return int(lastId), false, nil
}

func (s *sProject) ListQas(ctx context.Context, req *api.QaListReq) (res *api.QaListRes, err error) {
	res = &api.QaListRes{List: []api.QaItem{}}
	m := g.DB().Model("project_qas q").Ctx(ctx).
		LeftJoin("sys_users u", "u.id = q.updated_by").
		Fields("q.id, q.question, q.answer, q.tags, q.hits, q.status, COALESCE(u.real_name, u.username) AS updater, q.updated_at").
		Where("q.project_id", req.ProjectId).
		Where("q.status", "active")
	if req.Keyword != "" {
		kw := "%" + escape.Like(req.Keyword) + "%"
		m = m.Where("(q.question LIKE ? ESCAPE '|' OR q.answer LIKE ? ESCAPE '|' OR q.tags LIKE ? ESCAPE '|')", kw, kw, kw)
	}
	if req.Tag != "" {
		m = m.Where("q.tags LIKE ?", "%"+escape.Like(req.Tag)+"%")
	}
	rows, err := m.Order("q.hits DESC, q.id DESC").Limit(req.Size).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询 QA 失败")
	}
	for _, r := range rows {
		res.List = append(res.List, api.QaItem{
			Id: r["id"].Int(), Question: r["question"].String(), Answer: r["answer"].String(),
			Tags: r["tags"].String(), Hits: r["hits"].Int(), Status: r["status"].String(),
			Updater: r["updater"].String(), UpdatedAt: r["updated_at"].String(),
		})
	}
	return res, nil
}

// HitQa 命中计数（agent 检索命中后调用；高频 QA 进开工包）
func (s *sProject) HitQa(ctx context.Context, projectId, id int) (err error) {
	if _, err = g.DB().Exec(ctx, "UPDATE project_qas SET hits = hits + 1 WHERE id = ? AND project_id = ?", id, projectId); err != nil {
		return liberr.WrapDb(ctx, err, "更新命中失败")
	}
	return nil
}

func (s *sProject) ArchiveQa(ctx context.Context, projectId, id int) (err error) {
	result, err := g.DB().Model("project_qas").Ctx(ctx).
		Where("id", id).Where("project_id", projectId).
		Data(g.Map{"status": "archived"}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "归档失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("QA 不存在")
	}
	return nil
}
