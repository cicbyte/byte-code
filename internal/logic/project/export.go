package project

// 项目数据导出：数据所有权出口（记忆中枢定位下用户必须能带走全部数据）

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/docs"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *sProject) ExportProject(ctx context.Context, projectId int) (*api.ProjectExportRes, error) {
	// 导出属管理级操作：数据含评论作者、记忆等全量内容
	if uid := perm.UserId(ctx); !perm.IsProjectOwner(ctx, uid, projectId) {
		return nil, fmt.Errorf("仅项目管理员可导出项目数据")
	}

	res := &api.ProjectExportRes{
		Version:    "1.1",
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 任一表读取失败即整体报错：静默吞错会产出"看似成功"的残缺导出包。
	// （v1.0 的实证：documentIndex 表无 id 列，统一按 id 排序查询恒报错
	// 被吞，导出的 documentIndex 一直是空数组）
	var ferr error
	fetch := func(table, where, order string, args ...interface{}) (out []map[string]interface{}) {
		m := g.DB().Model(table).Ctx(ctx)
		if where != "" {
			m = m.Where(where, args...)
		}
		if order == "" {
			order = "id ASC"
		}
		rows, err := m.Order(order).All()
		if err != nil {
			if ferr == nil {
				ferr = liberr.WrapDb(ctx, err, "导出读取失败："+table)
			}
			return nil
		}
		out = make([]map[string]interface{}, 0, len(rows))
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

	// 附属实体归属条件：attachments/entity_tags 按 (entity_type, entity_id)
	// 关联，取项目内四大实体 id 并集；db_columns 经 db_tables 两跳
	entityScope := func(table string) string {
		return fmt.Sprintf(`(%s.entity_type = 'task' AND %s.entity_id IN (SELECT id FROM tasks WHERE project_id = ?)
OR %s.entity_type = 'requirement' AND %s.entity_id IN (SELECT id FROM requirements WHERE project_id = ?)
OR %s.entity_type = 'test_plan' AND %s.entity_id IN (SELECT id FROM test_plans WHERE project_id = ?)
OR %s.entity_type = 'test_case' AND %s.entity_id IN (SELECT id FROM test_cases WHERE project_id = ?))`,
			table, table, table, table, table, table, table, table)
	}
	entityArgs := []interface{}{projectId, projectId, projectId, projectId}

	res.Members = fetch("project_members", "project_id", "", projectId)
	res.Requirements = fetch("requirements", "project_id", "", projectId)
	res.Milestones = fetch("milestones", "project_id", "", projectId)
	res.Sprints = fetch("sprints", "project_id", "", projectId)
	res.Tasks = fetch("tasks", "project_id", "", projectId)
	res.Comments = fetch("comments c", "c.task_id IN (SELECT id FROM tasks WHERE project_id = ?)", "", projectId)
	res.TestPlans = fetch("test_plans", "project_id", "", projectId)
	res.TestCases = fetch("test_cases", "project_id", "", projectId)
	res.TestPlanCases = fetch("test_plan_cases", "test_plan_id IN (SELECT id FROM test_plans WHERE project_id = ?)", "", projectId)
	res.Memories = fetch("project_memories", "project_id", "", projectId)
	// 复合主键 (project_id, path)，无 id 列——按 path 排序
	res.DocumentIndex = fetch("project_document_index", "project_id", "path ASC", projectId)
	res.Attachments = fetch("attachments", entityScope("attachments"), "", entityArgs...)
	res.EntityTags = fetch("entity_tags", entityScope("entity_tags"), "", entityArgs...)
	res.Activities = fetch("activities", "project_id", "", projectId)
	res.AiExecutionLogs = fetch("ai_execution_logs", "task_id IN (SELECT id FROM tasks WHERE project_id = ?)", "", projectId)
	res.Databases = fetch("project_databases", "project_id", "", projectId)
	res.DbTables = fetch("db_tables", "project_id", "", projectId)
	res.DbColumns = fetch("db_columns", "table_id IN (SELECT id FROM db_tables WHERE project_id = ?)", "", projectId)
	res.SchemaVersions = fetch("schema_versions", "project_id", "", projectId)
	if ferr != nil {
		return nil, ferr
	}

	// 文档正文：vault 当前版本收录（.history 快照排除——版本体积大，
	// 索引+当前正文已构成可迁移的完整文档集）
	res.DocContents = map[string]string{}
	root := docs.RootPath(int64(projectId))
	if _, err := os.Stat(root); err == nil {
		walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, rerr := filepath.Rel(root, p)
			if rerr != nil {
				return rerr
			}
			rel = filepath.ToSlash(rel)
			// 历史快照目录整棵跳过
			if d.IsDir() {
				if rel == ".history" || strings.HasPrefix(rel, ".history/") {
					return fs.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(rel, ".history/") || filepath.Ext(p) != ".md" {
				return nil
			}
			data, rerr := os.ReadFile(p)
			if rerr != nil {
				return rerr
			}
			res.DocContents[rel] = string(data)
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("读取文档正文失败: %w", walkErr)
		}
	}
	return res, nil
}
