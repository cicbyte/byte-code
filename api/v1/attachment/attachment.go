package attachment

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 附件操作 ====================

type AttachmentUploadReq struct {
	g.Meta     `path:"/v1/attachments/upload" method:"post" mime:"multipart/form-data" tags:"附件管理" summary:"上传附件"`
	EntityType string `p:"entityType" v:"required|in:task,requirement,doc,test_case,project#实体类型不能为空"`
	EntityId   int    `p:"entityId" v:"required#实体ID不能为空"`
	File       string `p:"file" type:"file" v:"required#文件不能为空"`
}

type AttachmentUploadRes struct {
	g.Meta `mime:"application/json"`
	Id     int    `json:"id"`
}

type AttachmentGetReq struct {
	g.Meta `path:"/v1/attachments/{id}" method:"get" tags:"附件管理" summary:"获取附件信息"`
	Id     int `json:"-" in:"path" v:"required#附件ID不能为空"`
}

type AttachmentGetRes struct {
	g.Meta `mime:"application/json"`
	AttachmentItem
}

type AttachmentDownloadReq struct {
	g.Meta `path:"/v1/attachments/{id}/download" method:"get" tags:"附件管理" summary:"获取下载URL"`
	Id     int `json:"-" in:"path" v:"required#附件ID不能为空"`
}

type AttachmentDownloadRes struct {
	g.Meta `mime:"application/json"`
	Url    string `json:"url"`
}

type AttachmentPreviewReq struct {
	g.Meta `path:"/v1/attachments/{id}/preview" method:"get" tags:"附件管理" summary:"获取预览URL"`
	Id     int `json:"-" in:"path" v:"required#附件ID不能为空"`
}

type AttachmentPreviewRes struct {
	g.Meta `mime:"application/json"`
	Url    string `json:"url"`
}

type AttachmentDeleteReq struct {
	g.Meta `path:"/v1/attachments/{id}" method:"delete" tags:"附件管理" summary:"删除附件"`
	Id     int `json:"-" in:"path" v:"required#附件ID不能为空"`
}

type AttachmentDeleteRes struct {
	g.Meta `mime:"application/json"`
}

type AttachmentListReq struct {
	g.Meta     `path:"/v1/attachments" method:"get" tags:"附件管理" summary:"查询实体附件列表"`
	EntityType string `json:"entityType" v:"required#实体类型不能为空"`
	EntityId   int    `json:"entityId" v:"required#实体ID不能为空"`
}

type AttachmentListRes struct {
	g.Meta `mime:"application/json"`
	List   []AttachmentItem `json:"list"`
}

type AttachmentUpdateReq struct {
	g.Meta      `path:"/v1/attachments/{id}" method:"put" tags:"附件管理" summary:"更新附件描述"`
	Id          int    `json:"-" in:"path" v:"required#附件ID不能为空"`
	Description string `json:"description"`
}

type AttachmentUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type AttachmentItem struct {
	Id            int    `json:"id"`
	S3Key         string `json:"s3Key"`
	OriginalName  string `json:"originalName"`
	FileSize      int64  `json:"fileSize"`
	MimeType      string `json:"mimeType"`
	FileExt       string `json:"fileExt"`
	EntityType    string `json:"entityType"`
	EntityId      int    `json:"entityId"`
	UploaderId    int    `json:"uploaderId"`
	Description   string `json:"description"`
	DownloadCount int    `json:"downloadCount"`
	CreatedAt     string `json:"createdAt"`
	UploaderName  string `json:"uploaderName,omitempty"`
}

// ==================== 存储配置管理 ====================

type StorageConfigGetReq struct {
	g.Meta `path:"/v1/admin/storage/config" method:"get" tags:"存储管理" summary:"获取存储配置"`
}

type StorageConfigGetRes struct {
	g.Meta   `mime:"application/json"`
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
	Region   string `json:"region"`
	UseSSL   bool   `json:"useSSL"`
}

type StorageConfigUpdateReq struct {
	g.Meta    `path:"/v1/admin/storage/config" method:"put" tags:"存储管理" summary:"更新存储配置"`
	Endpoint  string `json:"endpoint" v:"required#endpoint不能为空"`
	Bucket    string `json:"bucket" v:"required#bucket不能为空"`
	AccessKey string `json:"accessKey" v:"required#accessKey不能为空"`
	SecretKey string `json:"secretKey" v:"required#secretKey不能为空"`
	Region    string `json:"region"`
	UseSSL    bool   `json:"useSSL"`
}

type StorageConfigUpdateRes struct {
	g.Meta `mime:"application/json"`
}

type StorageTestReq struct {
	g.Meta `path:"/v1/admin/storage/test" method:"post" tags:"存储管理" summary:"测试S3连接"`
}

type StorageTestRes struct {
	g.Meta `mime:"application/json"`
	Ok     bool   `json:"ok"`
	Msg    string `json:"msg,omitempty"`
}
