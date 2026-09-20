package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 项目发布（Releases，#551） ====================
// 发布是项目治理级动作：建/删收 maintainer 档；读与下载全成员 + 绑定 agent
//（文件走附件通道，附件门禁按 release→project 归属解析）

const releaseTimeLayout = "2006-01-02 15:04:05"

func (s *sProject) CreateRelease(ctx context.Context, req *api.ReleaseCreateReq) (id int, err error) {
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), req.ProjectId) {
		return 0, fmt.Errorf("仅项目管理员可创建发布")
	}
	if req.Title == "" {
		req.Title = req.Version
	}
	now := time.Now().Format(releaseTimeLayout)
	result, err := g.DB().Model("project_releases").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId,
		"version":    req.Version,
		"title":      req.Title,
		"notes":      req.Notes,
		"channel":    req.Channel,
		"created_by": perm.UserId(ctx),
		"created_at": now,
		"updated_at": now,
	})
	if err != nil {
		// 项目内版本号唯一：同版本重复发布明确报出
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry") {
			return 0, fmt.Errorf("版本 %s 已存在", req.Version)
		}
		return 0, liberr.WrapDb(ctx, err, "创建发布失败")
	}
	lastId, _ := result.LastInsertId()
	id = int(lastId)

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    perm.UserId(ctx),
		ActorType:  "human",
		Action:     "release.published",
		TargetType: "release",
		TargetID:   id,
		TargetName: req.Version,
		ProjectID:  req.ProjectId,
		Detail:     fmt.Sprintf("发布 %s（%s）", req.Version, req.Channel),
	})
	return id, nil
}

func (s *sProject) ListReleases(ctx context.Context, req *api.ReleaseListReq) (total int, list []api.ReleaseItem, err error) {
	m := g.DB().Model("project_releases").Ctx(ctx).Where("project_id", req.ProjectId)
	if req.Channel != "" {
		m = m.Where("channel", req.Channel)
	}
	total, err = m.Count()
	if err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询发布数量失败")
	}
	pageNum, pageSize := req.PageNum, req.PageSize
	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if err = m.Page(pageNum, pageSize).Order("id DESC").Scan(&list); err != nil {
		return 0, nil, liberr.WrapDb(ctx, err, "查询发布列表失败")
	}
	if list == nil {
		list = []api.ReleaseItem{}
	}
	fillReleaseAggregates(ctx, list)
	fillReleaseUserNames(ctx, list)
	return total, list, nil
}

// fillReleaseAggregates 批量回填各发布的文件数与总大小（一次 IN 聚合替代逐行查询）
func fillReleaseAggregates(ctx context.Context, list []api.ReleaseItem) {
	if len(list) == 0 {
		return
	}
	ids := make([]int, 0, len(list))
	for _, r := range list {
		ids = append(ids, r.Id)
	}
	rows, err := g.DB().Model("attachments").Ctx(ctx).
		Where("entity_type", "release").
		WhereIn("entity_id", ids).
		Group("entity_id").
		Fields("entity_id, COUNT(*) AS cnt, COALESCE(SUM(file_size),0) AS sz").All()
	if err != nil {
		return
	}
	agg := map[int]*api.ReleaseItem{}
	for i := range list {
		agg[list[i].Id] = &list[i]
	}
	for _, r := range rows {
		if it, ok := agg[r["entity_id"].Int()]; ok {
			it.FileCount = r["cnt"].Int()
			it.TotalSizeBytes = r["sz"].Int64()
		}
	}
}

func fillReleaseUserNames(ctx context.Context, list []api.ReleaseItem) {
	if len(list) == 0 {
		return
	}
	ids := make([]int, 0, len(list))
	seen := map[int]bool{}
	for _, r := range list {
		if r.CreatedBy > 0 && !seen[r.CreatedBy] {
			seen[r.CreatedBy] = true
			ids = append(ids, r.CreatedBy)
		}
	}
	if len(ids) == 0 {
		return
	}
	rows, err := g.DB().Model("sys_users").Ctx(ctx).WhereIn("id", ids).
		Fields("id, COALESCE(NULLIF(real_name,''), username) AS name").All()
	if err != nil {
		return
	}
	names := map[int]string{}
	for _, r := range rows {
		names[r["id"].Int()] = r["name"].String()
	}
	for i := range list {
		list[i].CreatedByName = names[list[i].CreatedBy]
	}
}

func (s *sProject) UpdateRelease(ctx context.Context, req *api.ReleaseUpdateReq) (err error) {
	pid := perm.EntityProjectId(ctx, "project_releases", req.Id)
	if pid == 0 {
		return fmt.Errorf("发布不存在")
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), pid) {
		return fmt.Errorf("仅项目管理员可更新发布")
	}
	data := g.Map{"updated_at": time.Now().Format(releaseTimeLayout)}
	if req.Title != "" {
		data["title"] = req.Title
	}
	// 说明允许清空（发错想撤回说明的场景）
	if req.Notes != "" {
		data["notes"] = req.Notes
	}
	if req.Channel != "" {
		data["channel"] = req.Channel
	}
	if _, err = g.DB().Model("project_releases").Ctx(ctx).WherePri(req.Id).Update(data); err != nil {
		return liberr.WrapDb(ctx, err, "更新发布失败")
	}
	return nil
}

func (s *sProject) DeleteRelease(ctx context.Context, id int) (err error) {
	pid := perm.EntityProjectId(ctx, "project_releases", id)
	if pid == 0 {
		return fmt.Errorf("发布不存在")
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), pid) {
		return fmt.Errorf("仅项目管理员可删除发布")
	}
	// 发布与其附件记录同事务删除
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Exec(
			"DELETE FROM attachments WHERE entity_type = 'release' AND entity_id = ?", id,
		); err != nil {
			return err
		}
		_, err := tx.Ctx(ctx).Model("project_releases").WherePri(id).Delete()
		return err
	})
}
