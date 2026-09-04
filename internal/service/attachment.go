package service

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"context"

	api "github.com/cicbyte/byte-code/api/v1/attachment"
)

type IAttachment interface {
	Upload(ctx context.Context, req *api.AttachmentUploadReq) (id int, err error)
	Get(ctx context.Context, id int) (res *api.AttachmentGetRes, err error)
	DownloadURL(ctx context.Context, id int) (url string, err error)
	ServeFile(ctx context.Context, r *ghttp.Request, id int) error
	PreviewURL(ctx context.Context, id int) (url string, err error)
	Delete(ctx context.Context, id int) (err error)
	List(ctx context.Context, req *api.AttachmentListReq) (list []api.AttachmentItem, err error)
	Update(ctx context.Context, req *api.AttachmentUpdateReq) (err error)
	// 存储配置
	GetStorageConfig(ctx context.Context) (res *api.StorageConfigGetRes, err error)
	UpdateStorageConfig(ctx context.Context, req *api.StorageConfigUpdateReq) (err error)
	TestStorage(ctx context.Context) (ok bool, msg string, err error)
}

var localAttachment IAttachment

func Attachment() IAttachment {
	if localAttachment == nil {
		panic("implement not found for interface IAttachment, forgot register?")
	}
	return localAttachment
}

func RegisterAttachment(i IAttachment) {
	localAttachment = i
}
