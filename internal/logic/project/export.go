package project

// 项目数据导出：数据所有权出口（记忆中枢定位下用户必须能带走全部数据）

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *sProject) ExportProject(ctx context.Context, projectId int) (*api.ProjectExportRes, error) {
	// 导出属管理级操作：数据含评论作者、记忆等全量内容
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, projectId) {
		return nil, fmt.Errorf("仅项目管理员可导出项目数据")
	}

	res := &api.ProjectExportRes{
		Version:    "1.0",
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	fetch := func(table, where string, args ...interface{}) []map[string]interface{} {
		m := g.DB().Model(table).Ctx(ctx)
		if where != "" {
			m = m.Where(where, args...)
		}
		rows, err := m.Order("id ASC").All()
		if err != nil {
			return []map[string]interface{}{}
		}
		out := make([]map[string]interface{}, 0, len(rows))
		for _, r := range rows {
			out = append(out, r.Map())
		}
		return out
	}

	pRow, err := g.DB().Model("projects").Ctx(ctx).Where("id", projectId).One()
	if err != nil || pRow.IsEmpty() {
		return nil, fmt.Errorf("项目不存在")
	}
	res.Project = pRow.Map()

	res.Requirements = fetch("requirements", "project_id", projectId)
	res.Milestones = fetch("milestones", "project_id", projectId)
	res.Sprints = fetch("sprints", "project_id", projectId)
	res.Tasks = fetch("tasks", "project_id", projectId)
	res.Comments = fetch("comments c", "c.task_id IN (SELECT id FROM tasks WHERE project_id = ?)", projectId)
	res.TestPlans = fetch("test_plans", "project_id", projectId)
	res.TestCases = fetch("test_cases", "project_id", projectId)
	res.TestPlanCases = fetch("test_plan_cases", "test_plan_id IN (SELECT id FROM test_plans WHERE project_id = ?)", projectId)
	res.Memories = fetch("project_memories", "project_id", projectId)
	res.DocumentIndex = fetch("project_document_index", "project_id", projectId)
	return res, nil
}
