package controller

import (
	"github.com/gogf/gf/v2/frame/g"
	"context"

	api "github.com/cicbyte/byte-code/api/v1/attachment"
	service "github.com/cicbyte/byte-code/internal/service"
)

var AttachmentCtrl = attachmentController{}

type attachmentController struct {
	BaseController
}

func (c *attachmentController) AttachmentUpload(ctx context.Context, req *api.AttachmentUploadReq) (res *api.AttachmentUploadRes, err error) {
	res = new(api.AttachmentUploadRes)
	id, err := service.Attachment().Upload(ctx, req)
	res.Id = id
	return
}

func (c *attachmentController) AttachmentGet(ctx context.Context, req *api.AttachmentGetReq) (res *api.AttachmentGetRes, err error) {
	return service.Attachment().Get(ctx, req.Id)
}

func (c *attachmentController) AttachmentDownload(ctx context.Context, req *api.AttachmentDownloadReq) (res *api.AttachmentDownloadRes, err error) {
	res = new(api.AttachmentDownloadRes)
	url, err := service.Attachment().DownloadURL(ctx, req.Id)
	res.Url = url
	return
}

func (c *attachmentController) AttachmentPreview(ctx context.Context, req *api.AttachmentPreviewReq) (res *api.AttachmentPreviewRes, err error) {
	res = new(api.AttachmentPreviewRes)
	url, err := service.Attachment().PreviewURL(ctx, req.Id)
	res.Url = url
	return
}

func (c *attachmentController) AttachmentDelete(ctx context.Context, req *api.AttachmentDeleteReq) (res *api.AttachmentDeleteRes, err error) {
	res = new(api.AttachmentDeleteRes)
	err = service.Attachment().Delete(ctx, req.Id)
	return
}

func (c *attachmentController) AttachmentList(ctx context.Context, req *api.AttachmentListReq) (res *api.AttachmentListRes, err error) {
	res = new(api.AttachmentListRes)
	list, err := service.Attachment().List(ctx, req)
	res.List = list
	return
}

func (c *attachmentController) AttachmentUpdate(ctx context.Context, req *api.AttachmentUpdateReq) (res *api.AttachmentUpdateRes, err error) {
	res = new(api.AttachmentUpdateRes)
	err = service.Attachment().Update(ctx, req)
	return
}

// ==================== 存储配置管理 ====================

func (c *attachmentController) StorageConfigGet(ctx context.Context, req *api.StorageConfigGetReq) (res *api.StorageConfigGetRes, err error) {
	return service.Attachment().GetStorageConfig(ctx)
}

func (c *attachmentController) StorageConfigUpdate(ctx context.Context, req *api.StorageConfigUpdateReq) (res *api.StorageConfigUpdateRes, err error) {
	res = new(api.StorageConfigUpdateRes)
	err = service.Attachment().UpdateStorageConfig(ctx, req)
	return
}

func (c *attachmentController) StorageTest(ctx context.Context, req *api.StorageTestReq) (res *api.StorageTestRes, err error) {
	res = new(api.StorageTestRes)
	ok, msg, err := service.Attachment().TestStorage(ctx)
	res.Ok = ok
	res.Msg = msg
	return
}


// AttachmentFile 本地存储附件直出（S3 后端走预签名 URL 不经此端点）
func (c *attachmentController) AttachmentFile(ctx context.Context, req *api.AttachmentFileReq) (res *api.AttachmentFileRes, err error) {
	if serr := service.Attachment().ServeFile(ctx, g.RequestFromCtx(ctx), req.Id); serr != nil {
		g.RequestFromCtx(ctx).Response.WriteStatus(404, serr.Error())
	}
	return
}
