package project

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/library/liberr"
	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/cicbyte/byte-code/utility/storage"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ==================== 项目发布（Releases） ====================
// 发布是项目治理级动作：建/删收 maintainer 档；读与下载全成员 + 绑定 agent。
// 文件独立体系（release_files）：按 release/{projectId}/{version}/{文件名}
// 组织、单文件上限 2GB、公开令牌直链分享（吊销=清 token）

const releaseTimeLayout = "2006-01-02 15:04:05"

// releaseFileMaxSize 安装包动辄几百 MB（用户口径）：发布文件专属上限 2GB
//（附件通道仍 20MB 不变）；服务端 clientMaxBodySize 须同步抬高
const releaseFileMaxSize = int64(2) << 30

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
	rows, err := g.DB().Model("release_files").Ctx(ctx).
		WhereIn("release_id", ids).
		Group("release_id").
		Fields("release_id, COUNT(*) AS cnt, COALESCE(SUM(file_size),0) AS sz").All()
	if err != nil {
		return
	}
	agg := map[int]*api.ReleaseItem{}
	for i := range list {
		agg[list[i].Id] = &list[i]
	}
	for _, r := range rows {
		if it, ok := agg[r["release_id"].Int()]; ok {
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
	// 先取存储键（删行后无从查起），发布与文件记录同事务删除
	var files []struct{ StorageKey string }
	if err := g.DB().Model("release_files").Ctx(ctx).Where("release_id", id).
		Fields("storage_key").Scan(&files); err != nil {
		return liberr.WrapDb(ctx, err, "查询发布文件失败")
	}
	if err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Model("release_files").Where("release_id", id).Delete(); err != nil {
			return err
		}
		_, err := tx.Ctx(ctx).Model("project_releases").WherePri(id).Delete()
		return err
	}); err != nil {
		return liberr.WrapDb(ctx, err, "删除发布失败")
	}
	st, serr := storage.NewStorageFromCtx(ctx)
	if serr == nil {
		for _, f := range files {
			_ = st.Delete(ctx, f.StorageKey)
		}
	}
	return nil
}

// ---------- 发布文件 ----------

// releaseFileCtx 取文件行 + 所属发布 + 项目 id；供文件域各操作统一鉴权
type releaseFileRow struct {
	Id         int
	ReleaseId  int
	FileName   string
	StorageKey string
	Version    string
	ProjectId  int
}

func loadReleaseFile(ctx context.Context, id int) (*releaseFileRow, error) {
	row, err := g.DB().Model("release_files f").Ctx(ctx).
		InnerJoin("project_releases r", "f.release_id = r.id").
		Fields("f.id, f.release_id, f.file_name, f.storage_key, r.version, r.project_id").
		Where("f.id", id).One()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询发布文件失败")
	}
	if row.IsEmpty() {
		return nil, fmt.Errorf("文件不存在")
	}
	return &releaseFileRow{
		Id: row["id"].Int(), ReleaseId: row["release_id"].Int(),
		FileName: row["file_name"].String(), StorageKey: row["storage_key"].String(),
		Version: row["version"].String(), ProjectId: row["project_id"].Int(),
	}, nil
}

// sanitizeFileName 存储键用文件名：去路径分隔与 Windows 非法字符（防穿越/乱键）
func sanitizeFileName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`\/:*?"<>|`, r) {
			return '_'
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	return name
}

func (s *sProject) UploadReleaseFile(ctx context.Context, req *api.ReleaseFileUploadReq) (id int, err error) {
	rel := perm.EntityProjectId(ctx, "project_releases", req.Id)
	if rel == 0 {
		return 0, fmt.Errorf("发布不存在")
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), rel) {
		return 0, fmt.Errorf("仅项目管理员可上传发布文件")
	}
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return 0, fmt.Errorf("文件不能为空")
	}
	if file.Size > releaseFileMaxSize {
		return 0, fmt.Errorf("文件超过大小上限 2GB")
	}
	name := sanitizeFileName(file.Filename)
	// 版本内同名拒传：直链语义按文件名稳定，重传先删（也防误覆盖）
	if cnt, _ := g.DB().Model("release_files").Ctx(ctx).
		Where("release_id", req.Id).Where("file_name", name).Count(); cnt > 0 {
		return 0, fmt.Errorf("同名文件 %s 已存在（请先删除旧文件）", name)
	}
	vrow, _ := g.DB().Model("project_releases").Ctx(ctx).WherePri(req.Id).Fields("version").Value()
	key := fmt.Sprintf("release/%d/%s/%s", rel, sanitizeFileName(vrow.String()), name)

	f, oerr := file.Open()
	if oerr != nil {
		return 0, fmt.Errorf("打开上传文件失败")
	}
	defer f.Close()
	st, serr := storage.NewStorageFromCtx(ctx)
	if serr != nil {
		g.Log().Errorf(ctx, "获取存储实例失败: %v", serr)
		return 0, fmt.Errorf("存储服务不可用，请联系管理员检查存储配置")
	}
	// MIME 不设白名单：发布包形态开放（exe/msi/dmg/apk/镜像…），
	// 下载恒带 attachment 头 + 服务端推导值，无直开 XSS 面
	mime := mimeByExt(filepath.Ext(name))
	if uerr := st.Upload(ctx, key, f, file.Size, mime); uerr != nil {
		g.Log().Errorf(ctx, "上传发布文件失败: %v", uerr)
		return 0, fmt.Errorf("上传失败")
	}
	result, ierr := g.DB().Model("release_files").Ctx(ctx).Insert(g.Map{
		"release_id":     req.Id,
		"file_name":      name,
		"file_size":      file.Size,
		"mime_type":      mime,
		"storage_key":    key,
		"uploader_id":    perm.UserId(ctx),
		"download_count": 0,
		"created_at":     time.Now().Format(releaseTimeLayout),
	})
	if ierr != nil {
		_ = st.Delete(ctx, key)
		return 0, liberr.WrapDb(ctx, ierr, "登记发布文件失败")
	}
	lastId, _ := result.LastInsertId()

	activity.Record(ctx, activity.ActivityInput{
		ActorID:    perm.UserId(ctx),
		ActorType:  "human",
		Action:     "release.file_uploaded",
		TargetType: "release_file",
		TargetID:   int(lastId),
		TargetName: name,
		ProjectID:  rel,
		Detail:     fmt.Sprintf("上传发布文件: %s (%d bytes)", name, file.Size),
	})
	return int(lastId), nil
}

func mimeByExt(ext string) string {
	m := map[string]string{
		".zip": "application/zip", ".gz": "application/gzip", ".tar": "application/x-tar",
		".rar": "application/vnd.rar", ".7z": "application/x-7z-compressed",
		".exe": "application/x-msdownload", ".msi": "application/x-msi",
		".dmg": "application/x-apple-diskimage", ".pkg": "application/x-newton-compatible-pkg",
		".apk": "application/vnd.android.package-archive", ".deb": "application/vnd.debian.binary-package",
		".rpm": "application/x-rpm", ".bin": "application/octet-stream", ".img": "application/octet-stream",
		".iso": "application/x-iso9660-image", ".whl": "application/python-wheel",
	}
	if v, ok := m[strings.ToLower(ext)]; ok {
		return v
	}
	return "application/octet-stream"
}

func (s *sProject) ListReleaseFiles(ctx context.Context, releaseId int) (res *api.ReleaseFileListRes, err error) {
	res = &api.ReleaseFileListRes{List: []api.ReleaseFileItem{}}
	rows, err := g.DB().Model("release_files f").Ctx(ctx).
		LeftJoin("sys_users u", "f.uploader_id = u.id").
		Fields("f.id, f.release_id, f.file_name, f.file_size, f.mime_type, f.uploader_id, "+
			"COALESCE(NULLIF(u.real_name,''), u.username) AS uploader_name, "+
			"f.download_count, f.share_token, f.share_expires_at, f.created_at").
		Where("f.release_id", releaseId).
		Order("f.id ASC").All()
	if err != nil {
		return nil, liberr.WrapDb(ctx, err, "查询发布文件失败")
	}
	for _, r := range rows {
		item := api.ReleaseFileItem{
			Id: r["id"].Int(), ReleaseId: r["release_id"].Int(),
			FileName: r["file_name"].String(), FileSize: r["file_size"].Int64(),
			MimeType: r["mime_type"].String(), UploaderId: r["uploader_id"].Int(),
			UploaderName:  r["uploader_name"].String(),
			DownloadCount: r["download_count"].Int(),
			ShareExpiresAt: r["share_expires_at"].String(),
			CreatedAt:     r["created_at"].String(),
		}
		item.Shared = r["share_token"].String() != ""
		if !item.Shared {
			item.ShareExpiresAt = ""
		}
		res.List = append(res.List, item)
	}
	return res, nil
}

func (s *sProject) DeleteReleaseFile(ctx context.Context, id int) (err error) {
	row, err := loadReleaseFile(ctx, id)
	if err != nil {
		return err
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), row.ProjectId) {
		return fmt.Errorf("仅项目管理员可删除发布文件")
	}
	if _, derr := g.DB().Model("release_files").Ctx(ctx).WherePri(id).Delete(); derr != nil {
		return liberr.WrapDb(ctx, derr, "删除发布文件失败")
	}
	if st, serr := storage.NewStorageFromCtx(ctx); serr == nil {
		_ = st.Delete(ctx, row.StorageKey)
	}
	return nil
}

// ShareReleaseFile 分享直链：幂等——未过期令牌直接复用；days>0 重新生成并设期
func (s *sProject) ShareReleaseFile(ctx context.Context, req *api.ReleaseFileShareReq) (res *api.ReleaseFileShareRes, err error) {
	row, err := loadReleaseFile(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), row.ProjectId) {
		return nil, fmt.Errorf("仅项目管理员可分享发布文件")
	}
	cur, _ := g.DB().Model("release_files").Ctx(ctx).WherePri(req.Id).
		Fields("share_token, share_expires_at").One()
	token := cur["share_token"].String()
	expires := cur["share_expires_at"].String()
	valid := token != "" && (expires == "" || expires > time.Now().Format(releaseTimeLayout))
	if !valid || req.Days > 0 {
		buf := make([]byte, 24)
		if _, rerr := rand.Read(buf); rerr != nil {
			return nil, fmt.Errorf("生成令牌失败")
		}
		token = hex.EncodeToString(buf)
		expires = ""
		if req.Days > 0 {
			expires = time.Now().AddDate(0, 0, req.Days).Format(releaseTimeLayout)
		}
		if _, uerr := g.DB().Model("release_files").Ctx(ctx).WherePri(req.Id).
			Data(g.Map{"share_token": token, "share_expires_at": expires}).Update(); uerr != nil {
			return nil, liberr.WrapDb(ctx, uerr, "保存分享令牌失败")
		}
	}
	return &api.ReleaseFileShareRes{
		Url:       fmt.Sprintf("/api/v1/release-files/public/%s", token),
		ExpiresAt: expires,
	}, nil
}

func (s *sProject) RevokeReleaseFileShare(ctx context.Context, id int) (err error) {
	row, err := loadReleaseFile(ctx, id)
	if err != nil {
		return err
	}
	if !perm.IsProjectMaintainer(ctx, perm.UserId(ctx), row.ProjectId) {
		return fmt.Errorf("仅项目管理员可吊销分享")
	}
	_, err = g.DB().Model("release_files").Ctx(ctx).WherePri(id).
		Data(g.Map{"share_token": nil, "share_expires_at": ""}).Update()
	if err != nil {
		return liberr.WrapDb(ctx, err, "吊销分享失败")
	}
	return nil
}

// ServeReleaseFilePublic 公开直链下载（免鉴权）：本地后端流式直出，
// S3 后端 302 预签名（1h）；过期/吊销令牌一律 404
func (s *sProject) ServeReleaseFilePublic(ctx context.Context, r *ghttp.Request, token string) {
	row, err := g.DB().Model("release_files").Ctx(ctx).
		Where("share_token", token).Fields("id, file_name, storage_key, share_expires_at, mime_type").One()
	if err != nil || row.IsEmpty() {
		r.Response.WriteStatus(404, "链接无效或已被吊销")
		return
	}
	if e := row["share_expires_at"].String(); e != "" && e <= time.Now().Format(releaseTimeLayout) {
		r.Response.WriteStatus(404, "链接已过期")
		return
	}
	_, _ = g.DB().Model("release_files").Ctx(ctx).WherePri(row["id"].Int()).Increment("download_count", 1)

	st, serr := storage.NewStorageFromCtx(ctx)
	if serr != nil {
		r.Response.WriteStatus(500, "存储服务不可用")
		return
	}
	url, derr := st.DownloadURL(ctx, row["storage_key"].String(), time.Hour)
	if derr != nil {
		r.Response.WriteStatus(500, "生成下载失败")
		return
	}
	if strings.HasPrefix(url, "local:") {
		local := storage.NewLocalStorage()
		abs, oerr := local.OpenPath(row["storage_key"].String())
		if oerr != nil {
			r.Response.WriteStatus(404, "文件缺失")
			return
		}
		if _, serr := os.Stat(abs); serr != nil {
			r.Response.WriteStatus(404, "文件缺失")
			return
		}
		name := strings.ReplaceAll(row["file_name"].String(), `"`, `_`)
		r.Response.Header().Set("Content-Type", row["mime_type"].String())
		r.Response.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
		r.Response.ServeFileDownload(abs, row["file_name"].String())
		return
	}
	r.Response.Header().Set("Location", url)
	r.Response.WriteStatus(302)
}
