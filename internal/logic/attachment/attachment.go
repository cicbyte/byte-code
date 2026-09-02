package attachment

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	api "github.com/cicbyte/byte-code/api/v1/attachment"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/activity"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/cicbyte/byte-code/utility/storage"
	liberr "github.com/cicbyte/byte-code/library/liberr"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
)

func init() {
	service.RegisterAttachment(New())
}

func New() *sAttachment {
	return &sAttachment{}
}

type sAttachment struct{}

// 附件扩展名白名单与存储 MIME：MIME 由扩展名推导而非信任客户端 Content-Type，
// 防止伪造 text/html 等类型经预签名 URL 直开造成存储型 XSS；
// html/htm/svg/js 等可执行/可渲染类型一律禁止上传
var attachmentExtMime = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".webp": "image/webp", ".bmp": "image/bmp",
	".pdf": "application/pdf", ".txt": "text/plain", ".md": "text/markdown",
	".csv": "text/csv", ".log": "text/plain", ".json": "application/json",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".xls": "application/vnd.ms-excel",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".doc": "application/msword",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".ppt": "application/vnd.ms-powerpoint",
	".zip": "application/zip", ".rar": "application/vnd.rar",
	".7z": "application/x-7z-compressed", ".tar": "application/x-tar",
	".gz": "application/gzip",
}

// 单文件大小上限 20MB（须配合 server.clientMaxBodySize 配置）
const attachmentMaxSize = 20 << 20

func (s *sAttachment) Upload(ctx context.Context, req *api.AttachmentUploadReq) (id int, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)
		if uid == 0 {
			panic("用户未登录")
		}

		// 实体归属校验：附件必须挂在当前用户可访问的项目实体上
		if !attachmentEntityAccessible(ctx, uid, req.EntityType, req.EntityId) {
			panic("无权向该实体上传附件")
		}

		// 从请求中获取上传文件
		r := g.RequestFromCtx(ctx)
		if r == nil {
			panic("获取请求失败")
		}
		file := r.GetUploadFile("file")
		if file == nil {
			panic("上传文件不能为空")
		}

		// 类型白名单校验（在触达存储之前）
		fileExt := strings.ToLower(filepath.Ext(file.Filename))
		mime, ok := attachmentExtMime[fileExt]
		if !ok {
			panic("不支持的文件类型: " + fileExt)
		}
		if file.Size > attachmentMaxSize {
			panic("文件超过大小上限 20MB")
		}

		// 打开文件
		f, err := file.Open()
		if err != nil {
			g.Log().Errorf(ctx, "打开上传文件失败: %v", err)
			panic("打开上传文件失败")
		}
		defer f.Close()

		// 生成 S3 key
		monthStr := time.Now().Format("2006-01")
		s3Key := fmt.Sprintf("%s/%d/%s/%s%s", req.EntityType, req.EntityId, monthStr, uuid.New().String(), fileExt)

		// 获取 S3 存储实例
		s3Storage, err := storage.NewS3StorageFromCtx(ctx)
		if err != nil {
			// S3 错误可能含 endpoint/凭据校验细节，只进服务端日志
			g.Log().Errorf(ctx, "获取存储实例失败: %v", err)
			panic("存储服务不可用，请联系管理员检查存储配置")
		}

		// 上传到 S3
		// S3 上传用扩展名推导的安全 MIME（不信任客户端声明，防伪造 Content-Type 直开 XSS）
		err = s3Storage.Upload(ctx, s3Key, f, file.Size, mime)
		if err != nil {
			g.Log().Errorf(ctx, "上传文件到S3失败: %v", err)
			panic("上传文件失败")
		}

		// 创建附件记录（MIME 用扩展名推导值，不信任客户端声明）
		result, err := g.DB().Model("attachments").Ctx(ctx).Insert(g.Map{
			"s3_key":         s3Key,
			"original_name":  file.Filename,
			"file_size":      file.Size,
			"mime_type":      mime,
			"file_ext":       fileExt,
			"entity_type":    req.EntityType,
			"entity_id":      req.EntityId,
			"uploader_id":    uid,
			"description":    "",
			"download_count": 0,
			"created_at":     time.Now().Format("2006-01-02 15:04:05"),
		})
		liberr.ErrIsNil(ctx, err, "创建附件记录失败")

		lastId, err := result.LastInsertId()
		liberr.ErrIsNil(ctx, err, "获取附件ID失败")
		id = int(lastId)

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "attachment.uploaded",
			TargetType: "attachment",
			TargetID:   id,
			TargetName: file.Filename,
			Detail:     fmt.Sprintf("上传附件: %s (%s)", file.Filename, req.EntityType),
		})
	})
	return
}

func (s *sAttachment) Get(ctx context.Context, id int) (res *api.AttachmentGetRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		res = &api.AttachmentGetRes{}
		err := g.DB().Model("attachments a").Ctx(ctx).
			LeftJoin("sys_users u", "a.uploader_id = u.id").
			Fields("a.*, u.real_name as uploader_name").
			Where("a.id", id).
			Scan(&res.AttachmentItem)
		liberr.ErrIsNil(ctx, err, "获取附件信息失败")
		if res.AttachmentItem.Id == 0 {
			panic("附件不存在")
		}
	})
	return
}

func (s *sAttachment) DownloadURL(ctx context.Context, id int) (url string, err error) {
	if !attachmentByIdAccessible(ctx, perm.UserId(ctx), id) {
		return "", fmt.Errorf("无权访问该附件")
	}
	err = g.Try(ctx, func(ctx context.Context) {
		var item struct {
			S3Key string
		}
		err := g.DB().Model("attachments").Ctx(ctx).WherePri(id).Scan(&item)
		liberr.ErrIsNil(ctx, err, "获取附件失败")
		if item.S3Key == "" {
			panic("附件不存在")
		}

		s3Storage, err := storage.NewS3StorageFromCtx(ctx)
		liberr.ErrIsNil(ctx, err, "获取存储实例失败")

		url, err = s3Storage.DownloadURL(ctx, item.S3Key, 1*time.Hour)
		liberr.ErrIsNil(ctx, err, "生成下载URL失败")

		// 增加下载计数
		_, _ = g.DB().Model("attachments").Ctx(ctx).WherePri(id).Increment("download_count", 1)
	})
	return
}

func (s *sAttachment) PreviewURL(ctx context.Context, id int) (url string, err error) {
	if !attachmentByIdAccessible(ctx, perm.UserId(ctx), id) {
		return "", fmt.Errorf("无权访问该附件")
	}
	err = g.Try(ctx, func(ctx context.Context) {
		var item struct {
			S3Key string
		}
		err := g.DB().Model("attachments").Ctx(ctx).WherePri(id).Scan(&item)
		liberr.ErrIsNil(ctx, err, "获取附件失败")
		if item.S3Key == "" {
			panic("附件不存在")
		}

		s3Storage, err := storage.NewS3StorageFromCtx(ctx)
		liberr.ErrIsNil(ctx, err, "获取存储实例失败")

		url, err = s3Storage.DownloadURL(ctx, item.S3Key, 30*time.Minute)
		liberr.ErrIsNil(ctx, err, "生成预览URL失败")
	})
	return
}

func (s *sAttachment) Delete(ctx context.Context, id int) (err error) {
	if !attachmentByIdAccessible(ctx, perm.UserId(ctx), id) {
		return fmt.Errorf("无权删除该附件")
	}
	err = g.Try(ctx, func(ctx context.Context) {
		userId := ctx.Value("userId")
		uid, _ := userId.(int)

		var item struct {
			S3Key        string
			OriginalName string
		}
		err := g.DB().Model("attachments").Ctx(ctx).WherePri(id).Scan(&item)
		liberr.ErrIsNil(ctx, err, "获取附件失败")
		if item.S3Key == "" {
			panic("附件不存在")
		}

		// 从 S3 删除文件
		s3Storage, err := storage.NewS3StorageFromCtx(ctx)
		if err == nil {
			_ = s3Storage.Delete(ctx, item.S3Key)
		}

		// 删除数据库记录
		_, err = g.DB().Model("attachments").Ctx(ctx).WherePri(id).Delete()
		liberr.ErrIsNil(ctx, err, "删除附件记录失败")

		activity.Record(ctx, activity.ActivityInput{
			ActorID:    uid,
			ActorType:  "human",
			Action:     "attachment.deleted",
			TargetType: "attachment",
			TargetID:   id,
			TargetName: item.OriginalName,
			Detail:     fmt.Sprintf("删除附件: %s", item.OriginalName),
		})
	})
	return
}

func (s *sAttachment) List(ctx context.Context, req *api.AttachmentListReq) (list []api.AttachmentItem, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		err := g.DB().Model("attachments a").Ctx(ctx).
			LeftJoin("sys_users u", "a.uploader_id = u.id").
			Fields("a.*, u.real_name as uploader_name").
			Where("a.entity_type", req.EntityType).
			Where("a.entity_id", req.EntityId).
			Order("a.created_at desc").
			Scan(&list)
		liberr.ErrIsNil(ctx, err, "获取附件列表失败")
	})
	return
}

func (s *sAttachment) Update(ctx context.Context, req *api.AttachmentUpdateReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		_, err := g.DB().Model("attachments").Ctx(ctx).WherePri(req.Id).Update(g.Map{
			"description": req.Description,
		})
		liberr.ErrIsNil(ctx, err, "更新附件描述失败")
	})
	return
}

// ==================== 存储配置管理 ====================

func (s *sAttachment) GetStorageConfig(ctx context.Context) (res *api.StorageConfigGetRes, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		config, err := storage.GetStorageConfig(ctx)
		liberr.ErrIsNil(ctx, err, "获取存储配置失败")
		res = &api.StorageConfigGetRes{
			Endpoint: config.Endpoint,
			Bucket:   config.Bucket,
			Region:   config.Region,
			UseSSL:   config.UseSSL,
		}
	})
	return
}

func (s *sAttachment) UpdateStorageConfig(ctx context.Context, req *api.StorageConfigUpdateReq) (err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		config := storage.StorageConfig{
			Endpoint:  req.Endpoint,
			Bucket:    req.Bucket,
			AccessKey: req.AccessKey,
			SecretKey: req.SecretKey,
			Region:    req.Region,
			UseSSL:    req.UseSSL,
		}
		configJson := g.Map{
			"endpoint":  config.Endpoint,
			"bucket":    config.Bucket,
			"accessKey": config.AccessKey,
			"secretKey": config.SecretKey,
			"region":    config.Region,
			"useSSL":    config.UseSSL,
		}
		configStr := gconvExport(configJson)

		// 检查是否已有配置
		count, err := g.DB().Model("sys_config").Ctx(ctx).Where("`key`", "storage_config").Count()
		liberr.ErrIsNil(ctx, err, "检查配置失败")

		if count > 0 {
			_, err = g.DB().Model("sys_config").Ctx(ctx).
				Where("`key`", "storage_config").
				Update(g.Map{
					"value":      configStr,
					"updated_at": gtime.Now(),
				})
		} else {
			_, err = g.DB().Model("sys_config").Ctx(ctx).Insert(g.Map{
				"key":        "storage_config",
				"value":      configStr,
				"created_at": gtime.Now(),
				"updated_at": gtime.Now(),
			})
		}
		liberr.ErrIsNil(ctx, err, "保存存储配置失败")
	})
	return
}

func (s *sAttachment) TestStorage(ctx context.Context) (ok bool, msg string, err error) {
	err = g.Try(ctx, func(ctx context.Context) {
		config, err := storage.GetStorageConfig(ctx)
		liberr.ErrIsNil(ctx, err, "获取存储配置失败")

		testErr := storage.TestConfig(*config)
		if testErr != nil {
			ok = false
			msg = testErr.Error()
		} else {
			ok = true
			msg = "连接成功"
		}
	})
	return
}

func gconvExport(m g.Map) string {
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// attachmentEntityAccessible 校验当前用户是否有权操作指定实体的附件。
// entityType 必须与白名单一致，且目标实体所属项目须为当前用户可访问的项目。
func attachmentEntityAccessible(ctx context.Context, userId int, entityType string, entityId int) bool {
	if entityId <= 0 || userId <= 0 {
		return false
	}
	// 实体类型到含 project_id 列的表映射（与上传白名单一致）
	entityTable := map[string]string{
		"task":        "tasks",
		"requirement": "requirements",
		"doc":         "docs",
		"test_case":   "test_cases",
		"project":     "projects",
	}
	table, ok := entityTable[entityType]
	if !ok {
		return false
	}
	v, err := g.DB().Model(table).Where("id", entityId).Fields("project_id").Value()
	if err != nil || v == nil {
		return false
	}
	// project 类型自身即目标，直接校验项目成员
	pid := v.Int()
	if entityType == "project" {
		pid = entityId
	}
	return perm.CanAccessProject(ctx, userId, pid)
}

// attachmentByIdAccessible 按附件 id 解析其实体归属再校验
func attachmentByIdAccessible(ctx context.Context, userId, attachmentId int) bool {
	v, err := g.DB().Model("attachments").Where("id", attachmentId).Fields("entity_type", "entity_id").Value()
	_ = v
	// 取两个字段需用 One
	rec, err := g.DB().Model("attachments").Where("id", attachmentId).Fields("entity_type, entity_id").One()
	if err != nil || rec == nil {
		return false
	}
	return attachmentEntityAccessible(ctx, userId, rec["entity_type"].String(), rec["entity_id"].Int())
}
