package project

// 项目关联实现：owner 管理关联。关联的用途是**反馈投递面**——
// A 关联 B 后，A 可向 B 投递跨项目反馈（线索），由 B 的准入 agent 阅读
// 分析后自行决定是否建任务；A 侧不直接向 B 的任务池写入。

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/perm"
)

func (s *sProject) ListRelations(ctx context.Context, projectId int) (res *api.RelationListRes, err error) {
	res = &api.RelationListRes{List: []api.RelationItem{}}
	rows, err := g.DB().Model("project_relations r").Ctx(ctx).
		LeftJoin("projects p", "p.id = r.related_project_id").
		Fields("r.id, r.related_project_id, p.name, r.created_at").
		Where("r.project_id", projectId).
		Order("r.id ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询关联项目失败")
	}
	for _, row := range rows {
		res.List = append(res.List, api.RelationItem{
			Id:        row["id"].Int(),
			ProjectId: row["related_project_id"].Int(),
			Name:      row["name"].String(),
			CreatedAt: row["created_at"].String(),
		})
	}
	return res, nil
}

func (s *sProject) AddRelation(ctx context.Context, req *api.RelationAddReq) (err error) {
	// 关联是项目级治理动作：仅 owner 可添加（不能任意项目都加）
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, req.ProjectId) {
		return fmt.Errorf("仅项目管理员可管理关联项目")
	}
	if req.RelatedProjectId == req.ProjectId {
		return fmt.Errorf("不能关联自身")
	}
	// 对方必须存在
	if cnt, _ := g.DB().Model("projects").Ctx(ctx).Where("id", req.RelatedProjectId).Count(); cnt == 0 {
		return fmt.Errorf("对方项目不存在")
	}
	// 操作者须对对方项目可访问（防探测：无权项目连名字都不该挂进来）
	if !perm.CanAccessProject(ctx, perm.UserId(ctx), req.RelatedProjectId) {
		return fmt.Errorf("无权关联该项目（需先成为其成员或获得 agent 准入）")
	}
	_, err = g.DB().Model("project_relations").Ctx(ctx).Insert(g.Map{
		"project_id": req.ProjectId, "related_project_id": req.RelatedProjectId,
		"created_by": perm.UserId(ctx),
	})
	if err != nil {
		return liberr.WrapDb(ctx, err, "添加关联失败")
	}
	s.recordActivity(ctx, perm.UserId(ctx), "project.relation_added", "project", req.ProjectId,
		fmt.Sprintf("#%d", req.RelatedProjectId), req.ProjectId, fmt.Sprintf("关联了项目 #%d", req.RelatedProjectId))
	return nil
}

func (s *sProject) RemoveRelation(ctx context.Context, projectId, relationId int) (err error) {
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, projectId) {
		return fmt.Errorf("仅项目管理员可管理关联项目")
	}
	result, err := g.DB().Model("project_relations").Ctx(ctx).
		Where("id", relationId).Where("project_id", projectId).Delete()
	if err != nil {
		return liberr.WrapDb(ctx, err, "移除关联失败")
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("关联不存在")
	}
	return nil
}
