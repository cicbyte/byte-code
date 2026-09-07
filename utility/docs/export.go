package docs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/container/gvar"

	"github.com/cicbyte/byte-code/utility/dbinit"
)

// legacyExportFlag 导出完成标记（sys_config key），置 1 后不再重复导出
const legacyExportFlag = "legacy_docs_exported"

// legacyTargetPrefix 存量知识库导出的 vault 内目标前缀（space=knowledge）
const legacyTargetPrefix = KnowledgeDir

// ExportLegacyDocs 把存量 docs 表（旧知识库）一次性导出到各项目 vault。
// 幂等：sys_config 标记位 + 目标文件已存在即跳过，不覆盖 vault 中的人工修改。
func ExportLegacyDocs(ctx context.Context) error {
	v, _ := g.DB().Model("sys_config").Ctx(ctx).Where("key", legacyExportFlag).Value("value")
	if v.String() == "1" {
		return nil
	}
	// 全新安装没有 docs/doc_relations 表（表随旧版迁移早已移除）：无存量可导，
	// 直接置标记——否则每次启动都因表缺失告警且标记永远置不上
	if !tableExists(ctx, "docs") || !tableExists(ctx, "doc_relations") {
		markLegacyExported(ctx)
		return nil
	}

	rows, err := g.DB().Model("docs").Ctx(ctx).
		Where("status", "active").Order("parent_id ASC, sort_order ASC, id ASC").All()
	if err != nil {
		return err
	}
	relations, err := g.DB().Model("doc_relations").Ctx(ctx).All()
	if err != nil {
		return err
	}
	linkedByDoc := make(map[int64][]string)
	for _, r := range relations {
		t, id := r["target_type"].String(), r["target_id"].Int64()
		if id <= 0 {
			continue
		}
		var prefix string
		switch t {
		case "task":
			prefix = "task"
		case "requirement":
			prefix = "req"
		case "test_case":
			prefix = "tc"
		default: // project 关联无意义，跳过
			continue
		}
		linkedByDoc[r["doc_id"].Int64()] = append(linkedByDoc[r["doc_id"].Int64()], fmt.Sprintf("%s:%d", prefix, id))
	}

	// 按项目分组
	byProject := make(map[int64]gdb.Result)
	for _, r := range rows {
		pid := r["project_id"].Int64()
		if pid <= 0 { // 无项目归属的孤儿文档无处安放，跳过
			continue
		}
		byProject[pid] = append(byProject[pid], r)
	}

	for pid, docs := range byProject {
		if err := exportProjectDocs(ctx, pid, docs, linkedByDoc); err != nil {
			return fmt.Errorf("export project %d docs: %w", pid, err)
		}
	}

	markLegacyExported(ctx)
	g.Log().Infof(ctx, "legacy docs exported to vault: %d projects", len(byProject))
	return nil
}

// markLegacyExported 置完成标记（count-then-insert，与 setting.go 既有模式一致）
func markLegacyExported(ctx context.Context) {
	cnt, _ := g.DB().Model("sys_config").Ctx(ctx).Where("key", legacyExportFlag).Count()
	if cnt > 0 {
		_, _ = g.DB().Model("sys_config").Ctx(ctx).Where("key", legacyExportFlag).Data("value", "1").Update()
	} else {
		_, _ = g.DB().Model("sys_config").Ctx(ctx).Data(g.Map{"key": legacyExportFlag, "value": "1"}).Insert()
	}
}

// tableExists 跨方言表存在性探测（mysql: information_schema / sqlite: sqlite_master）
func tableExists(ctx context.Context, name string) bool {
	var (
		v   *gvar.Var
		err error
	)
	if dbinit.Dialect() == "mysql" {
		v, err = g.DB().GetValue(ctx,
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", name)
	} else {
		v, err = g.DB().GetValue(ctx,
			"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name)
	}
	return err == nil && v != nil && v.Int() > 0
}

// exportProjectDocs 导出单个项目：folder→目录、doc/template→md（frontmatter: space=knowledge）
func exportProjectDocs(ctx context.Context, projectId int64, docs gdb.Result, linkedByDoc map[int64][]string) error {
	// id -> record
	byId := make(map[int64]gdb.Record, len(docs))
	for _, d := range docs {
		byId[d["id"].Int64()] = d
	}
	// parent -> children（保持 sort_order 序）
	children := make(map[int64][]gdb.Record)
	var roots gdb.Result
	for _, d := range docs {
		p := d["parent_id"].Int64()
		if p > 0 {
			if _, ok := byId[p]; ok {
				children[p] = append(children[p], d)
				continue
			}
		}
		roots = append(roots, d)
	}
	usedNames := make(map[string]bool) // 同目录下防重名（递归时按目录前缀隔离）

	var walk func(nodes gdb.Result, dirRel string, visited map[int64]bool) error
	walk = func(nodes gdb.Result, dirRel string, visited map[int64]bool) error {
		sort.SliceStable(nodes, func(i, j int) bool {
			si, sj := nodes[i]["sort_order"].Int(), nodes[j]["sort_order"].Int()
			if si != sj {
				return si < sj
			}
			return nodes[i]["id"].Int64() < nodes[j]["id"].Int64()
		})
		for _, d := range nodes {
			id := d["id"].Int64()
			if visited[id] { // 环防护
				continue
			}
			visited[id] = true
			title := strings.TrimSpace(d["title"].String())
			if title == "" {
				title = fmt.Sprintf("untitled-%d", id)
			}
			name := sanitizeName(title)
			docType := d["type"].String()

			if docType == "folder" {
				folderKey := dirRel + "/" + name + "/"
				uniq := name
				for i := 1; usedNames[folderKey+uniq+"/"]; i++ {
					uniq = fmt.Sprintf("%s-%d", name, i)
				}
				usedNames[folderKey+uniq+"/"] = true
				subRel := dirRel + "/" + uniq
				if err := os.MkdirAll(filepath.Join(RootPath(projectId), filepath.FromSlash(legacyTargetPrefix+subRel)), 0o755); err != nil {
					return err
				}
				if err := walk(children[id], subRel, visited); err != nil {
					return err
				}
				continue
			}

			// doc / template → md 文件
			fileKey := dirRel + "/"
			uniq := name
			for i := 1; usedNames[fileKey+uniq+".md"]; i++ {
				uniq = fmt.Sprintf("%s-%d", name, i)
			}
			relSlash := legacyTargetPrefix + dirRel + "/" + uniq + ".md"
			usedNames[fileKey+uniq+".md"] = true

			abs, err := SafeJoin(projectId, relSlash)
			if err != nil {
				return err
			}
			if _, err := os.Stat(abs); err == nil {
				continue // 已存在（前次导出或人工放置），不覆盖
			}
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return err
			}
			fm := Frontmatter{
				Title:  title, // 原始标题保留在 frontmatter，文件名可能被 sanitize
				Space:  "knowledge",
				Type:   docType,
				Status: "published",
				Linked: linkedByDoc[id],
			}
			content := RenderFrontmatter(fm, d["content"].String())
			if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(roots, "", make(map[int64]bool))
}

var legacyInvalidChars = regexp.MustCompile(`[\\/:*?"<>|\r\n\t]`)

// sanitizeName Windows 非法字符替换、去首尾空白与点、长度截断
func sanitizeName(name string) string {
	name = legacyInvalidChars.ReplaceAllString(strings.TrimSpace(name), "_")
	name = strings.Trim(name, ". ")
	if name == "" {
		name = "untitled"
	}
	if len(name) > 80 {
		name = name[:80]
	}
	return name
}
