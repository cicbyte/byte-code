package docs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// textExts 解析 frontmatter 的文本扩展名；其余类型仅索引元数据
var textExts = map[string]bool{".md": true, ".markdown": true, ".txt": true, ".mdx": true}

// DocIndex 索引行（project_document_index 表结构映射）
type DocIndex struct {
	ProjectId int64  `json:"project_id"`
	Path      string `json:"path"`
	Space     string `json:"space"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Tags      string `json:"tags"`
	Linked    string `json:"linked"`
	Ext       string `json:"ext"`
	Size      int64  `json:"size"`
	Checksum  string `json:"checksum"`
}

// ScanProject 增量扫描项目 vault 并同步索引表；返回 (新增+变更数, 删除数)
func ScanProject(ctx context.Context, projectId int64) (changed, deleted int, err error) {
	root := RootPath(projectId)

	// 现有索引：path -> checksum
	existing := make(map[string]string)
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).Fields("path", "checksum").All()
	if err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		existing[r["path"].String()] = r["checksum"].String()
	}

	seen := make(map[string]bool)
	if gfileExists(root) {
		err = filepath.WalkDir(root, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			name := d.Name()
			if d.IsDir() {
				// .history 与隐藏目录不入索引
				if p != root && (strings.HasPrefix(name, ".") || name == HistoryDir) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(name, ".") {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			relSlash := filepath.ToSlash(rel)
			seen[relSlash] = true

			sum, err := Checksum(p)
			if err != nil {
				g.Log().Warningf(ctx, "vault scan checksum %s: %v", relSlash, err)
				return nil
			}
			if old, ok := existing[relSlash]; ok && old == sum {
				return nil
			}

			idx := buildIndex(projectId, relSlash, p, info.Size(), sum)
			row := g.Map{
				"space": idx.Space, "title": idx.Title, "type": idx.Type,
				"status": idx.Status, "tags": idx.Tags, "linked": idx.Linked,
				"ext": idx.Ext, "size": idx.Size, "checksum": idx.Checksum,
				"updated_at": gtime.Now(),
			}
			if _, ok := existing[relSlash]; ok {
				if _, err := g.DB().Model("project_document_index").Ctx(ctx).
					Where("project_id", projectId).Where("path", relSlash).
					Data(row).Update(); err != nil {
					return err
				}
			} else {
				// Data 为覆盖语义，两次调用会丢字段：主键列与索引列必须并入同一 map
				insert := g.Map{"project_id": projectId, "path": relSlash}
				for k, v := range row {
					insert[k] = v
				}
				if _, err := g.DB().Model("project_document_index").Ctx(ctx).
					Data(insert).Insert(); err != nil {
					return err
				}
			}
			changed++
			return nil
		})
		if err != nil {
			return changed, deleted, fmt.Errorf("walk vault: %w", err)
		}
	}

	// 磁盘消失的文件删除索引行
	for path := range existing {
		if !seen[path] {
			if _, err := g.DB().Model("project_document_index").Ctx(ctx).
				Where("project_id", projectId).Where("path", path).Delete(); err != nil {
				return changed, deleted, err
			}
			deleted++
		}
	}
	return changed, deleted, nil
}

// ScanAll 全量扫描所有项目 vault（服务启动 / 手动全量刷新时调用）
func ScanAll(ctx context.Context) error {
	ids, err := g.DB().Model("projects").Ctx(ctx).Array("id")
	if err != nil {
		return err
	}
	for _, v := range ids {
		id := v.Int64()
		changed, deleted, err := ScanProject(ctx, id)
		if err != nil {
			g.Log().Warningf(ctx, "vault scan project %d: %v", id, err)
			continue
		}
		if changed > 0 || deleted > 0 {
			g.Log().Infof(ctx, "vault scanned project %d: %d changed, %d deleted", id, changed, deleted)
		}
	}
	return nil
}

// buildIndex 从文件构建索引行：文本类型解析 frontmatter，二进制仅元数据
func buildIndex(projectId int64, relSlash, abs string, size int64, sum string) DocIndex {
	ext := strings.ToLower(filepath.Ext(relSlash))
	base := strings.TrimSuffix(filepath.Base(relSlash), filepath.Ext(relSlash))
	idx := DocIndex{
		ProjectId: projectId,
		Path:      relSlash,
		Title:     base,
		Status:    "published",
		Space:     defaultSpace(relSlash),
		Ext:       ext,
		Size:      size,
		Checksum:  sum,
	}
	if !textExts[ext] {
		return idx
	}
	content, err := os.ReadFile(abs)
	if err != nil {
		return idx
	}
	fm, _, err := ParseFrontmatter(string(content))
	if err != nil {
		return idx
	}
	if fm.Title != "" {
		idx.Title = fm.Title
	}
	if fm.Space == "knowledge" || fm.Space == "work" {
		idx.Space = fm.Space
	}
	if fm.Type != "" {
		idx.Type = fm.Type
	}
	if fm.Status == "draft" || fm.Status == "published" {
		idx.Status = fm.Status
	}
	idx.Tags = strings.Join(fm.Tags, ",")
	idx.Linked = strings.Join(fm.Linked, ",")
	return idx
}

// defaultSpace space 缺省约定：顶层目录为 知识库/knowledge 时 knowledge，否则 work
func defaultSpace(relSlash string) string {
	top := relSlash
	if i := strings.Index(relSlash, "/"); i >= 0 {
		top = relSlash[:i]
	}
	if top == KnowledgeDir || strings.EqualFold(top, "knowledge") {
		return "knowledge"
	}
	return "work"
}

func gfileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
