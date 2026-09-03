// Package vault 项目记忆与文档中枢 logic 层。
// 文档操作：磁盘为真相源，索引（project_document_index）在每次写后增量重扫收口。
package vault

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	api "github.com/cicbyte/byte-code/api/v1/vault"
	service "github.com/cicbyte/byte-code/internal/service"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/vault"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type sVault struct{}

func init() {
	service.RegisterVault(New())
}

// New 创建文档/记忆中枢 logic 实例
func New() *sVault {
	return &sVault{}
}

// textExt 是否文本文件（正文可读/可写 frontmatter）
func textExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".md", ".markdown", ".mdx", ".txt":
		return true
	}
	return false
}

// ==================== 目录树 ====================

func (s *sVault) Tree(ctx context.Context, projectId int64, space string) ([]api.VaultTreeNode, error) {
	root := vault.RootPath(projectId)
	if _, err := os.Stat(root); err != nil {
		return []api.VaultTreeNode{}, nil
	}
	idxMap := indexByPath(ctx, projectId)
	nodes, err := buildTree(root, "", idxMap)
	if err != nil {
		return nil, err
	}
	if space != "" {
		nodes = filterTreeSpace(nodes, space)
	}
	return nodes, nil
}

// buildTree 递归读磁盘目录树；文件元数据优先取索引，索引缺失时实时解析 frontmatter
func buildTree(root, rel string, idxMap map[string]*vault.DocIndex) ([]api.VaultTreeNode, error) {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	var nodes []api.VaultTreeNode
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}
		if e.IsDir() {
			children, err := buildTree(root, childRel, idxMap)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, api.VaultTreeNode{
				Name: name, Path: childRel, IsDir: true, Children: children,
			})
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		node := api.VaultTreeNode{
			Name: name, Path: childRel, Size: info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		}
		if idx, ok := idxMap[childRel]; ok {
			node.Meta = &api.FileMeta{
				Title: idx.Title, Space: idx.Space, Type: idx.Type, Status: idx.Status,
				Tags: splitCsv(idx.Tags), Linked: splitCsv(idx.Linked),
			}
		} else if textExt(path.Ext(name)) {
			if fm, _, err := readFrontmatter(filepath.Join(root, filepath.FromSlash(childRel))); err == nil {
				node.Meta = fmToFileMeta(fm, childRel)
			}
		}
		nodes = append(nodes, node)
	}
	// 目录在前，同级按名称排序（稳定展示）
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

// filterTreeSpace 按空间过滤：knowledge 只留 知识库/knowledge 顶层，work 留其余
func filterTreeSpace(nodes []api.VaultTreeNode, space string) []api.VaultTreeNode {
	var out []api.VaultTreeNode
	for _, n := range nodes {
		if n.IsDir {
			if s := spaceOfTop(n.Name); s == space {
				out = append(out, n)
			}
			continue
		}
		if s := spaceOfTop(n.Name); s == space {
			out = append(out, n)
		}
	}
	return out
}

func spaceOfTop(topDirOrFile string) string {
	if topDirOrFile == vault.KnowledgeDir || strings.EqualFold(topDirOrFile, "knowledge") {
		return "knowledge"
	}
	return "work"
}

func fmToFileMeta(fm vault.Frontmatter, relSlash string) *api.FileMeta {
	meta := &api.FileMeta{
		Title:  fm.Title,
		Space:  fm.Space,
		Type:   fm.Type,
		Status: fm.Status,
		Tags:   fm.Tags,
		Linked: fm.Linked,
	}
	if meta.Title == "" {
		meta.Title = strings.TrimSuffix(path.Base(relSlash), path.Ext(relSlash))
	}
	if meta.Space == "" {
		meta.Space = spaceOfTop(strings.SplitN(relSlash, "/", 2)[0])
	}
	if meta.Status == "" {
		meta.Status = "published"
	}
	return meta
}

// ==================== 读取 ====================

func (s *sVault) ReadFile(ctx context.Context, projectId int64, rel string) (*api.VaultFileGetRes, error) {
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return nil, gerror.New(err.Error())
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, gerror.New("文件不存在")
	}
	if info.IsDir() {
		return nil, gerror.New("路径是目录，请读取具体文件")
	}
	res := &api.VaultFileGetRes{Path: rel, Size: info.Size()}
	if textExt(path.Ext(rel)) {
		data, err := os.ReadFile(abs)
		if err != nil {
			return nil, gerror.New("读取文件失败")
		}
		fm, _, _ := vault.ParseFrontmatter(string(data))
		res.Content = string(data)
		res.FileMeta = fmToFileMeta(fm, rel)
		return res, nil
	}
	res.Binary = true
	if idx := indexOne(ctx, projectId, rel); idx != nil {
		res.FileMeta = &api.FileMeta{
			Title: idx.Title, Space: idx.Space, Type: idx.Type, Status: idx.Status,
			Tags: splitCsv(idx.Tags), Linked: splitCsv(idx.Linked),
		}
	}
	return res, nil
}

func (s *sVault) ServeRaw(ctx context.Context, r *ghttp.Request, projectId int64, rel string) error {
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return gerror.New(err.Error())
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return gerror.New("文件不存在")
	}
	r.Response.Header().Set("Content-Disposition",
		`inline; filename="`+strings.ReplaceAll(path.Base(rel), `"`, `_`)+`"`)
	r.Response.ServeFileDownload(abs, path.Base(rel))
	return nil
}

// ==================== 写入 ====================

// writeWithSnapshot 写前快照旧版，随后覆盖写入
func writeWithSnapshot(projectId int64, rel, content string) (int64, error) {
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return 0, gerror.New(err.Error())
	}
	if err := vault.Snapshot(projectId, rel); err != nil {
		return 0, gerror.Newf("创建版本快照失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return 0, gerror.New("创建目录失败")
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return 0, gerror.New("写入文件失败")
	}
	return int64(len(content)), nil
}

// demoteKnowledgeDraft 发布流守卫：knowledge 空间文件被修改时，published 自动回 draft 等人审。
// 返回调整后的内容；explicitStatus=true（meta 显式指定状态，即人审发布动作）时不降级。
func demoteKnowledgeDraft(oldContent, newContent string, explicitStatus bool) string {
	if explicitStatus {
		return newContent
	}
	oldFm, _, err1 := vault.ParseFrontmatter(oldContent)
	if err1 != nil || oldFm.Space != "knowledge" || oldFm.Status != "published" {
		return newContent
	}
	newFm, body, err2 := vault.ParseFrontmatter(newContent)
	if err2 != nil {
		return newContent
	}
	if newFm.Space == "" {
		newFm.Space = oldFm.Space
	}
	if newFm.Title == "" {
		newFm.Title = oldFm.Title
	}
	newFm.Status = "draft"
	return vault.RenderFrontmatter(newFm, body)
}

func (s *sVault) WriteFile(ctx context.Context, projectId int64, rel, content string) (*api.VaultFileWriteRes, error) {
	if !textExt(path.Ext(rel)) {
		return nil, gerror.New("非文本文件请走上传接口")
	}
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return nil, gerror.New(err.Error())
	}
	old, _ := os.ReadFile(abs)
	content = demoteKnowledgeDraft(string(old), content, false)
	size, err := writeWithSnapshot(projectId, rel, content)
	if err != nil {
		return nil, err
	}
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return &api.VaultFileWriteRes{Path: rel, Size: size}, nil
}

func (s *sVault) PatchFile(ctx context.Context, projectId int64, rel string, ops []api.VaultPatchOperation) (*api.VaultFilePatchRes, error) {
	if !textExt(path.Ext(rel)) {
		return nil, gerror.New("补丁仅支持文本文件")
	}
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return nil, gerror.New(err.Error())
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, gerror.New("文件不存在")
	}
	content := string(raw)
	res := &api.VaultFilePatchRes{}
	for i, op := range ops {
		item := api.VaultPatchItem{Index: i, Type: op.Type}
		switch op.Type {
		case "replace":
			if op.Search == "" {
				item.Reason = "search 为空"
			} else if strings.Contains(content, op.Search) {
				content = strings.Replace(content, op.Search, op.Replace, 1)
				item.Applied = true
			} else {
				item.Reason = "search 未命中"
			}
		case "append":
			content += op.Content
			item.Applied = true
		case "prepend":
			fm, body, _ := vault.ParseFrontmatter(content)
			if fm.Title != "" || fm.Space != "" { // 有 frontmatter：正文头插入，保持头部完整
				content = vault.RenderFrontmatter(fm, op.Content+body)
			} else {
				content = op.Content + content
			}
			item.Applied = true
		default:
			item.Reason = "未知操作类型 " + op.Type
		}
		if item.Applied {
			res.Applied++
		} else {
			res.Skipped++
		}
		res.Items = append(res.Items, item)
	}
	if res.Applied == 0 {
		return res, nil // 全部未命中：不落盘，调用方可依据 items 判断
	}
	content = demoteKnowledgeDraft(string(raw), content, false)
	if _, err := writeWithSnapshot(projectId, rel, content); err != nil {
		return nil, err
	}
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return res, nil
}

func (s *sVault) CreateFolder(ctx context.Context, projectId int64, rel string) error {
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return gerror.New(err.Error())
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return gerror.New("创建目录失败")
	}
	return nil
}

func (s *sVault) Upload(ctx context.Context, projectId int64, dir string, file *ghttp.UploadFile) (*api.VaultUploadRes, error) {
	if file == nil {
		return nil, gerror.New("文件不能为空")
	}
	name := sanitizeFileName(file.Filename)
	if dir == "" {
		dir = ""
	}
	rel := dir
	if strings.HasSuffix(dir, ".") || !strings.Contains(path.Base(dir), ".") {
		// dir 视为目标目录
		rel = strings.Trim(dir, "/") + "/" + name
	}
	rel = strings.Trim(rel, "/")
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return nil, gerror.New(err.Error())
	}
	if _, err := os.Stat(abs); err == nil {
		if err := vault.Snapshot(projectId, rel); err != nil {
			return nil, gerror.Newf("创建版本快照失败: %v", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, gerror.New("创建目录失败")
	}
	if _, err := file.Save(filepath.Dir(abs), false); err != nil {
		return nil, gerror.New("保存文件失败")
	}
	// Save 用原文件名落盘，与 rel 不一致时重命名对齐
	saved := filepath.Join(filepath.Dir(abs), file.Filename)
	if filepath.Base(saved) != filepath.Base(abs) {
		if err := os.Rename(saved, abs); err != nil {
			return nil, gerror.New("文件重命名失败")
		}
	}
	info, _ := os.Stat(abs)
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	return &api.VaultUploadRes{Path: rel, Size: size}, nil
}

func (s *sVault) Move(ctx context.Context, projectId int64, from, to string) error {
	fromAbs, err := vault.SafeJoin(projectId, from)
	if err != nil {
		return gerror.New(err.Error())
	}
	toAbs, err := vault.SafeJoin(projectId, to)
	if err != nil {
		return gerror.New(err.Error())
	}
	if _, err := os.Stat(fromAbs); err != nil {
		return gerror.New("源路径不存在")
	}
	if _, err := os.Stat(toAbs); err == nil {
		return gerror.New("目标路径已存在")
	}
	if strings.HasPrefix(filepath.Clean(toAbs)+string(filepath.Separator), filepath.Clean(fromAbs)+string(filepath.Separator)) {
		return gerror.New("不能移动到自身子目录")
	}
	if err := os.MkdirAll(filepath.Dir(toAbs), 0o755); err != nil {
		return gerror.New("创建目标目录失败")
	}
	if err := os.Rename(fromAbs, toAbs); err != nil {
		return gerror.New("移动失败")
	}
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return nil
}

func (s *sVault) Delete(ctx context.Context, projectId int64, rel string) error {
	abs, err := vault.SafeJoin(projectId, rel)
	if err != nil {
		return gerror.New(err.Error())
	}
	info, err := os.Stat(abs)
	if err != nil {
		return gerror.New("路径不存在")
	}
	if info.IsDir() {
		// 非空目录拒绝删除：防止误删整棵子树，需先清空（空目录无快照价值）
		entries, err := os.ReadDir(abs)
		if err != nil {
			return gerror.New("读取目录失败")
		}
		if len(entries) > 0 {
			return gerror.New("目录非空，请先清空后再删除")
		}
		if err := os.Remove(abs); err != nil {
			return gerror.New("删除失败")
		}
	} else {
		if err := vault.Snapshot(projectId, rel); err != nil {
			return gerror.Newf("创建版本快照失败: %v", err)
		}
		if err := os.Remove(abs); err != nil {
			return gerror.New("删除失败")
		}
	}
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return nil
}

func (s *sVault) UpdateMeta(ctx context.Context, projectId int64, req *api.VaultMetaUpdateReq) error {
	if !textExt(path.Ext(req.Path)) {
		return gerror.New("仅文本文件支持 frontmatter 元数据")
	}
	abs, err := vault.SafeJoin(projectId, req.Path)
	if err != nil {
		return gerror.New(err.Error())
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return gerror.New("文件不存在")
	}
	fm, body, err := vault.ParseFrontmatter(string(raw))
	if err != nil {
		return gerror.New("frontmatter 解析失败")
	}
	if fm.Title == "" {
		fm.Title = strings.TrimSuffix(path.Base(req.Path), path.Ext(req.Path))
	}
	if fm.Space == "" {
		fm.Space = spaceOfTop(strings.SplitN(req.Path, "/", 2)[0])
	}
	if req.Title != nil {
		fm.Title = *req.Title
	}
	if req.Type != nil {
		fm.Type = *req.Type
	}
	if req.Tags != nil {
		fm.Tags = req.Tags
	}
	if req.Linked != nil {
		fm.Linked = req.Linked
	}
	if req.Status != nil {
		fm.Status = *req.Status // 人审发布动作
	}
	content := vault.RenderFrontmatter(fm, body)
	content = demoteKnowledgeDraft(string(raw), content, req.Status != nil)
	if _, err := writeWithSnapshot(projectId, req.Path, content); err != nil {
		return err
	}
	if _, _, err := vault.ScanProject(ctx, projectId); err != nil {
		return liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return nil
}

// ==================== 查询 ====================

func (s *sVault) Search(ctx context.Context, projectId int64, q, space string) ([]api.VaultSearchItem, error) {
	like := "%" + q + "%"
	model := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).
		Where("(title LIKE ? OR tags LIKE ? OR path LIKE ?)", like, like, like)
	if space != "" {
		model = model.Where("space", space)
	}
	rows, err := model.Order("updated_at DESC").Limit(50).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	seen := make(map[string]bool)
	var items []api.VaultSearchItem
	for _, r := range rows {
		p := r["path"].String()
		seen[p] = true
		items = append(items, idxRowToItem(r))
	}
	// md 正文包含匹配（索引未覆盖内容维度；量级允许扫盘）
	contentHits := searchContent(ctx, projectId, q)
	for _, r := range contentHits {
		p := r["path"].String()
		if seen[p] || (space != "" && r["space"].String() != space) {
			continue
		}
		seen[p] = true
		items = append(items, idxRowToItem(r))
		if len(items) >= 50 {
			break
		}
	}
	return items, nil
}

func (s *sVault) Linked(ctx context.Context, projectId int64, target string) ([]api.VaultSearchItem, error) {
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).
		Where("','||linked||',' LIKE ?", "%,"+target+",%").
		Limit(100).All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	var items []api.VaultSearchItem
	for _, r := range rows {
		items = append(items, idxRowToItem(r))
	}
	return items, nil
}

func (s *sVault) Refresh(ctx context.Context, projectId int64) (*api.VaultRefreshRes, error) {
	changed, deleted, err := vault.ScanProject(ctx, projectId)
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "索引同步失败")
	}
	return &api.VaultRefreshRes{Changed: changed, Deleted: deleted}, nil
}

// ==================== 内部工具 ====================

func indexByPath(ctx context.Context, projectId int64) map[string]*vault.DocIndex {
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).All()
	if err != nil {
		return map[string]*vault.DocIndex{}
	}
	m := make(map[string]*vault.DocIndex, len(rows))
	for i := range rows {
		r := rows[i]
		m[r["path"].String()] = &vault.DocIndex{
			ProjectId: projectId, Path: r["path"].String(), Space: r["space"].String(),
			Title: r["title"].String(), Type: r["type"].String(), Status: r["status"].String(),
			Tags: r["tags"].String(), Linked: r["linked"].String(), Ext: r["ext"].String(),
			Size: r["size"].Int64(), Checksum: r["checksum"].String(),
		}
	}
	return m
}

func indexOne(ctx context.Context, projectId int64, rel string) *vault.DocIndex {
	r, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).Where("path", rel).One()
	if err != nil || r.IsEmpty() {
		return nil
	}
	return &vault.DocIndex{
		ProjectId: projectId, Path: rel, Space: r["space"].String(), Title: r["title"].String(),
		Type: r["type"].String(), Status: r["status"].String(), Tags: r["tags"].String(),
		Linked: r["linked"].String(), Ext: r["ext"].String(), Size: r["size"].Int64(),
	}
}

func readFrontmatter(abs string) (vault.Frontmatter, string, error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return vault.Frontmatter{}, "", err
	}
	return vault.ParseFrontmatter(string(data))
}

func searchContent(ctx context.Context, projectId int64, q string) gdb.Result {
	rows, err := g.DB().Model("project_document_index").Ctx(ctx).
		Where("project_id", projectId).
		Where("ext IN (?)", g.Slice{".md", ".markdown", ".mdx", ".txt"}).
		Limit(500).All()
	if err != nil {
		return nil
	}
	var hits gdb.Result
	root := vault.RootPath(projectId)
	for _, r := range rows {
		p := r["path"].String()
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			continue
		}
		if strings.Contains(string(data), q) {
			hits = append(hits, r)
		}
	}
	return hits
}

func idxRowToItem(r gdb.Record) api.VaultSearchItem {
	return api.VaultSearchItem{
		Path:  r["path"].String(),
		Title: r["title"].String(),
		Space: r["space"].String(),
		Type:  r["type"].String(),
		Tags:  splitCsv(r["tags"].String()),
	}
}

func splitCsv(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

var windowsInvalid = `/:*?"<>|`

func sanitizeFileName(name string) string {
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(windowsInvalid, r) || r == '\r' || r == '\n' || r == '\t' {
			return '_'
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" {
		return "untitled"
	}
	return name
}

