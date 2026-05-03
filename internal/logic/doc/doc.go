package doc

import (
	"context"
	"fmt"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/doc"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/activity"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/frame/g"
)

func init() {
	service.RegisterDoc(New())
}

func New() *sDoc {
	return &sDoc{}
}

type sDoc struct{}

// ==================== 文档 CRUD ====================

func (s *sDoc) Create(ctx context.Context, req *api.DocCreateReq) (id int, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		if uid == 0 {
			panic("用户未登录")
		}
		result, err := g.DB().Model("docs").Ctx(ctx).Insert(g.Map{
			"project_id":    req.ProjectId,
			"parent_id":     req.ParentId,
			"title":         req.Title,
			"content":       req.Content,
			"type":          req.Type,
			"sort_order":    req.SortOrder,
			"creator_id":    uid,
			"last_editor_id": uid,
			"version":       1,
			"status":        "active",
			"created_at":    time.Now().Format("2006-01-02 15:04:05"),
			"updated_at":    time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建文档失败")
		lastId, err := result.LastInsertId()
		liberr.ErrIsNil(ctx, err, "获取文档ID失败")
		id = int(lastId)

		// 创建初始版本记录
		_, err = g.DB().Model("doc_versions").Ctx(ctx).Insert(g.Map{
			"doc_id":         id,
			"version":        1,
			"content":        req.Content,
			"editor_id":      uid,
			"change_summary": "创建文档",
			"created_at":     time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建版本记录失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "doc.created",
			TargetType: "doc",
			TargetID:   id,
			TargetName: req.Title,
			ProjectID:  req.ProjectId,
			Detail:     fmt.Sprintf("创建文档: %s", req.Title),
		})
	})
	return
}

func (s *sDoc) Update(ctx context.Context, req *api.DocUpdateReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		// 获取当前文档
		var currentDoc struct {
			Id      int
			Version int
			Title   string
		}
		err := g.DB().Model("docs").Ctx(ctx).WherePri(req.Id).Scan(&currentDoc)
		liberr.ErrIsNil(ctx, err, "获取文档失败")
		if currentDoc.Id == 0 {
			panic("文档不存在")
		}

		newVersion := currentDoc.Version + 1
		data := g.Map{
			"last_editor_id": uid,
			"version":        newVersion,
			"updated_at":     time.Now().Format("2006-01-02 15:04:05"),
		}
		if req.Title != "" {
			data["title"] = req.Title
		}
		if req.Content != "" {
			data["content"] = req.Content
		}
		if req.Type != "" {
			data["type"] = req.Type
		}
		if req.SortOrder != 0 {
			data["sort_order"] = req.SortOrder
		}
		if req.Status != "" {
			data["status"] = req.Status
		}
		if req.ParentId != 0 {
			data["parent_id"] = req.ParentId
		}

		_, err = g.DB().Model("docs").Ctx(ctx).WherePri(req.Id).Update(data)
		liberr.ErrIsNil(ctx, err, "更新文档失败")

		// 创建版本记录（仅在有内容变更时）
		if req.Content != "" {
			_, err = g.DB().Model("doc_versions").Ctx(ctx).Insert(g.Map{
				"doc_id":         req.Id,
				"version":        newVersion,
				"content":        req.Content,
				"editor_id":      uid,
				"change_summary": fmt.Sprintf("更新文档至版本 %d", newVersion),
				"created_at":     time.Now().Format("2006-01-02 15:04:05"),
			})
			liberr.ErrIsNil(ctx, err, "创建版本记录失败")
		}

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "doc.updated",
			TargetType: "doc",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("更新文档: ID=%d, 版本=%d", req.Id, newVersion),
		})
	})
	return
}

func (s *sDoc) Delete(ctx context.Context, id int) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		// 检查是否有子文档
		count, err := g.DB().Model("docs").Ctx(ctx).Where("parent_id", id).Count()
		liberr.ErrIsNil(ctx, err, "检查子文档失败")
		if count > 0 {
			panic("该文档下有子文档，不能删除")
		}

		// 删除关联
		_, err = g.DB().Model("doc_relations").Ctx(ctx).Where("doc_id", id).Delete()
		liberr.ErrIsNil(ctx, err, "删除文档关联失败")

		// 删除版本
		_, err = g.DB().Model("doc_versions").Ctx(ctx).Where("doc_id", id).Delete()
		liberr.ErrIsNil(ctx, err, "删除文档版本失败")

		// 删除文档
		_, err = g.DB().Model("docs").Ctx(ctx).WherePri(id).Delete()
		liberr.ErrIsNil(ctx, err, "删除文档失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "doc.deleted",
			TargetType: "doc",
			TargetID:   id,
			Detail:     fmt.Sprintf("删除文档: ID=%d", id),
		})
	})
	return
}

func (s *sDoc) List(ctx context.Context, req *api.DocListReq) (total int, list []api.DocItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// Count 查询（不带 Fields，兼容 SQLite）
		countM := g.DB().Model("docs d").Ctx(ctx)
		if req.ProjectId != 0 {
			countM = countM.Where("d.project_id", req.ProjectId)
		}
		if req.ParentId != 0 {
			countM = countM.Where("d.parent_id", req.ParentId)
		}
		if req.Type != "" {
			countM = countM.Where("d.type", req.Type)
		}
		if req.Status != "" {
			countM = countM.Where("d.status", req.Status)
		}
		if req.Keyword != "" {
			countM = countM.Where("d.title LIKE ?", "%"+req.Keyword+"%")
		}

		total, err = countM.Count()
		liberr.ErrIsNil(ctx, err, "获取文档数量失败")

		// 数据查询
		m := g.DB().Model("docs d").Ctx(ctx).
			LeftJoin("sys_users cu", "d.creator_id = cu.id").
			LeftJoin("sys_users eu", "d.last_editor_id = eu.id").
			Fields("d.*, cu.real_name as creator_name, eu.real_name as editor_name")
		if req.ProjectId != 0 {
			m = m.Where("d.project_id", req.ProjectId)
		}
		if req.ParentId != 0 {
			m = m.Where("d.parent_id", req.ParentId)
		}
		if req.Type != "" {
			m = m.Where("d.type", req.Type)
		}
		if req.Status != "" {
			m = m.Where("d.status", req.Status)
		}
		if req.Keyword != "" {
			m = m.Where("d.title LIKE ?", "%"+req.Keyword+"%")
		}

		pageNum := req.PageNum
		if pageNum == 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize == 0 {
			pageSize = 10
		}
		err = m.Page(pageNum, pageSize).Order("d.sort_order asc, d.created_at desc").Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取文档列表失败")
	})
	return
}

func (s *sDoc) Get(ctx context.Context, id int) (res *api.DocDetailRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		res = &api.DocDetailRes{}
		err := g.DB().Model("docs d").Ctx(ctx).
			LeftJoin("sys_users cu", "d.creator_id = cu.id").
			LeftJoin("sys_users eu", "d.last_editor_id = eu.id").
			Fields("d.*, cu.real_name as creator_name, eu.real_name as editor_name").
			Where("d.id", id).
			Scan(&res.DocItem)
		liberr.ErrIsNil(ctx, err, "获取文档失败")
		if res.DocItem.Id == 0 {
			panic("文档不存在")
		}
	})
	return
}

func (s *sDoc) Move(ctx context.Context, req *api.DocMoveReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		// 不能移动到自身下
		if req.Id == req.ParentId {
			panic("不能将文档移动到自身下")
		}

		_, err = g.DB().Model("docs").Ctx(ctx).WherePri(req.Id).Update(g.Map{
			"parent_id":     req.ParentId,
			"last_editor_id": uid,
			"updated_at":    time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "移动文档失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "doc.moved",
			TargetType: "doc",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("移动文档 ID=%d 到 parent_id=%d", req.Id, req.ParentId),
		})
	})
	return
}

func (s *sDoc) Tree(ctx context.Context, projectId int) (tree []api.DocTreeNode, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		var allNodes []api.DocTreeNode
		err := g.DB().Model("docs").Ctx(ctx).
			Where("project_id", projectId).
			Where("status", "active").
			Fields("id, parent_id, title, type, sort_order").
			Order("sort_order asc, id asc").
			Scan(&allNodes)
		liberr.ErrIsNil(ctx, err, "获取文档树失败")

		tree = buildTree(allNodes, 0)
	})
	return
}

func buildTree(nodes []api.DocTreeNode, parentId int) []api.DocTreeNode {
	var result []api.DocTreeNode
	for _, node := range nodes {
		if node.ParentId == parentId {
			children := buildTree(nodes, node.Id)
			node.Children = children
			result = append(result, node)
		}
	}
	return result
}

// ==================== 文档版本 ====================

func (s *sDoc) Versions(ctx context.Context, req *api.DocVersionListReq) (total int, list []api.DocVersionItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// Count 查询（不带 Fields，兼容 SQLite）
		total, err = g.DB().Model("doc_versions dv").Ctx(ctx).
			Where("dv.doc_id", req.Id).Count()
		liberr.ErrIsNil(ctx, err, "获取版本数量失败")

		pageNum := req.PageNum
		if pageNum == 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize == 0 {
			pageSize = 10
		}

		// 数据查询
		m := g.DB().Model("doc_versions dv").Ctx(ctx).
			LeftJoin("sys_users u", "dv.editor_id = u.id").
			Fields("dv.*, u.real_name as editor_name").
			Where("dv.doc_id", req.Id)
		err = m.Page(pageNum, pageSize).Order("dv.version desc").Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取版本列表失败")
	})
	return
}

func (s *sDoc) Revert(ctx context.Context, req *api.DocRevertReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		// 获取目标版本内容
		var version struct {
			Content string
		}
		err := g.DB().Model("doc_versions").Ctx(ctx).
			Where("doc_id", req.Id).
			Where("version", req.Version).
			Scan(&version)
		liberr.ErrIsNil(ctx, err, "获取目标版本失败")
		if version.Content == "" && req.Version != 1 {
			panic("目标版本不存在")
		}

		// 获取当前文档
		var currentDoc struct {
			Version int
		}
		err = g.DB().Model("docs").Ctx(ctx).WherePri(req.Id).Scan(&currentDoc)
		liberr.ErrIsNil(ctx, err, "获取文档失败")

		newVersion := currentDoc.Version + 1
		summary := req.Summary
		if summary == "" {
			summary = fmt.Sprintf("回退到版本 %d", req.Version)
		}

		// 更新文档
		_, err = g.DB().Model("docs").Ctx(ctx).WherePri(req.Id).Update(g.Map{
			"content":        version.Content,
			"version":        newVersion,
			"last_editor_id": uid,
			"updated_at":     time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "回退文档失败")

		// 创建版本记录
		_, err = g.DB().Model("doc_versions").Ctx(ctx).Insert(g.Map{
			"doc_id":         req.Id,
			"version":        newVersion,
			"content":        version.Content,
			"editor_id":      uid,
			"change_summary": summary,
			"created_at":     time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建版本记录失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "doc.reverted",
			TargetType: "doc",
			TargetID:   req.Id,
			Detail:     fmt.Sprintf("回退文档到版本 %d (新版本 %d)", req.Version, newVersion),
		})
	})
	return
}

// ==================== 文档关联 ====================

func (s *sDoc) CreateRelation(ctx context.Context, req *api.DocRelationCreateReq) (id int, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		// 检查是否已存在
		count, err := g.DB().Model("doc_relations").Ctx(ctx).
			Where("doc_id", req.Id).
			Where("target_type", req.TargetType).
			Where("target_id", req.TargetId).
			Count()
		liberr.ErrIsNil(ctx, err, "检查关联失败")
		if count > 0 {
			panic("关联已存在")
		}

		result, err := g.DB().Model("doc_relations").Ctx(ctx).Insert(g.Map{
			"doc_id":      req.Id,
			"target_type": req.TargetType,
			"target_id":   req.TargetId,
		})
		liberr.ErrIsNil(ctx, err, "创建关联失败")
		lastId, err := result.LastInsertId()
		liberr.ErrIsNil(ctx, err, "获取关联ID失败")
		id = int(lastId)
	})
	return
}

func (s *sDoc) DeleteRelation(ctx context.Context, docId int, targetType string, targetId int) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		_, err = g.DB().Model("doc_relations").Ctx(ctx).
			Where("doc_id", docId).
			Where("target_type", targetType).
			Where("target_id", targetId).
			Delete()
		liberr.ErrIsNil(ctx, err, "删除关联失败")
	})
	return
}

func (s *sDoc) ListRelations(ctx context.Context, docId int) (list []api.DocRelationItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		err := g.DB().Model("doc_relations").Ctx(ctx).
			Where("doc_id", docId).
			Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取关联列表失败")

		// 填充目标标题
		for i, rel := range list {
			title := s.getTargetTitle(ctx, rel.TargetType, rel.TargetId)
			list[i].TargetTitle = title
		}
	})
	return
}

func (s *sDoc) getTargetTitle(ctx context.Context, targetType string, targetId int) string {
	var tableName string
	switch targetType {
	case "task":
		tableName = "tasks"
	case "requirement":
		tableName = "requirements"
	case "test_case":
		tableName = "test_cases"
	case "doc":
		tableName = "docs"
	default:
		return ""
	}
	var title string
	err := g.DB().Model(tableName).Ctx(ctx).WherePri(targetId).Fields("title").Scan(&title)
	if err != nil {
		return ""
	}
	return title
}

// ==================== 全局搜索 ====================

func (s *sDoc) Search(ctx context.Context, req *api.SearchReq) (total int, list []api.SearchResult, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		keyword := req.Keyword
		list = make([]api.SearchResult, 0)

		// 搜索文档（使用 FTS5）
		s.searchDocs(ctx, keyword, &list)

		// 搜索任务
		s.searchTable(ctx, "tasks", "task", keyword, &list)

		// 搜索需求
		s.searchTable(ctx, "requirements", "requirement", keyword, &list)

		// 搜索测试用例
		s.searchTable(ctx, "test_cases", "test_case", keyword, &list)

		total = len(list)

		// 分页处理
		pageNum := req.PageNum
		if pageNum == 0 {
			pageNum = 1
		}
		pageSize := req.PageSize
		if pageSize == 0 {
			pageSize = 20
		}
		start := (pageNum - 1) * pageSize
		if start >= total {
			list = make([]api.SearchResult, 0)
			return
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		list = list[start:end]
	})
	return
}

func (s *sDoc) searchDocs(ctx context.Context, keyword string, results *[]api.SearchResult) {
	// 尝试使用 FTS5
	var docs []struct {
		Id    int
		Title string
	}
	err := g.DB().Model("docs").Ctx(ctx).
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Fields("id, title").
		Limit(50).
		Scan(&docs)
	if err != nil {
		return
	}
	for _, d := range docs {
		*results = append(*results, api.SearchResult{
			Module:  "doc",
			Id:      d.Id,
			Title:   d.Title,
			Summary: keyword,
		})
	}
}

func (s *sDoc) searchTable(ctx context.Context, tableName, module, keyword string, results *[]api.SearchResult) {
	var items []struct {
		Id    int
		Title string
	}
	err := g.DB().Model(tableName).Ctx(ctx).
		Where("title LIKE ?", "%"+keyword+"%").
		Fields("id, title").
		Limit(50).
		Scan(&items)
	if err != nil {
		return
	}
	for _, item := range items {
		*results = append(*results, api.SearchResult{
			Module:  module,
			Id:      item.Id,
			Title:   item.Title,
			Summary: keyword,
		})
	}
}
