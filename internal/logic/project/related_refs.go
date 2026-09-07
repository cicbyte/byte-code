package project

// 任务跨项目引用的解析与校验（迁移 54 / dev-docs 决策：第一档引用+通知）

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/cicbyte/byte-code/utility/perm"
)

// relatedRef 单条引用：指向另一项目的实体（当前仅 task 档；title 为快照）
type relatedRef struct {
	Type      string `json:"type"`
	ProjectId int64  `json:"projectId"`
	Id        int64  `json:"id"`
	Title     string `json:"title"`
}

func (r relatedRef) refKey() string {
	return fmt.Sprintf("%s:%d:%d", r.Type, r.ProjectId, r.Id)
}

// parseRelatedRefs 校验引用 JSON：数组结构合法、type 受限、被引实体存在、
// 且引用者对被引项目有读权限（CanAccessProject——防跨项目探测，P0-4 同款边界）
func parseRelatedRefs(ctx context.Context, raw string) ([]relatedRef, error) {
	var refs []relatedRef
	if raw == "" {
		raw = "[]"
	}
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return nil, fmt.Errorf("relatedRefs 必须是 JSON 数组")
	}
	uid := perm.UserId(ctx)
	seen := map[string]bool{}
	for i := range refs {
		r := &refs[i]
		if r.Type != "task" {
			return nil, fmt.Errorf("relatedRefs[%d].type 仅支持 task", i)
		}
		if r.ProjectId <= 0 || r.Id <= 0 {
			return nil, fmt.Errorf("relatedRefs[%d] 的 projectId/id 必须为正", i)
		}
		if !perm.CanAccessProject(ctx, uid, int(r.ProjectId)) {
			return nil, fmt.Errorf("无权引用项目 %d 的实体", r.ProjectId)
		}
		// 实体存在性 + title 快照回填（调用方传空或不准时以库内为准）
		title, err := g.DB().Model("tasks").Ctx(ctx).Where("id", r.Id).Fields("title").Value()
		if err != nil || title == nil {
			return nil, fmt.Errorf("被引任务 %d 不存在", r.Id)
		}
		r.Title = title.String()
		if seen[r.refKey()] {
			return nil, fmt.Errorf("relatedRefs 存在重复引用 %s", r.refKey())
		}
		seen[r.refKey()] = true
	}
	return refs, nil
}

// refProjectName 引用通知里的项目名（取不到退回 id 展示）
func (s *sProject) refProjectName(ctx context.Context, projectId int) string {
	if v, err := g.DB().Model("projects").Ctx(ctx).Where("id", projectId).Fields("name").Value(); err == nil && v != nil {
		return fmt.Sprintf("「%s」", v.String())
	}
	return fmt.Sprintf("#%d", projectId)
}

func (s *sProject) taskTitle(ctx context.Context, taskId int) string {
	if v, err := g.DB().Model("tasks").Ctx(ctx).Where("id", taskId).Fields("title").Value(); err == nil && v != nil {
		return v.String()
	}
	return fmt.Sprintf("#%d", taskId)
}

func (s *sProject) taskProjectId(ctx context.Context, taskId int) int {
	if v, err := g.DB().Model("tasks").Ctx(ctx).Where("id", taskId).Fields("project_id").Value(); err == nil && v != nil {
		return v.Int()
	}
	return 0
}
