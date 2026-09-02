package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/vault"
	service "github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/perm"
	"github.com/gogf/gf/v2/frame/g"
)

var Vault = vaultController{}

type vaultController struct {
	BaseController
}

func (c *vaultController) VaultTree(ctx context.Context, req *api.VaultTreeReq) (res *api.VaultTreeRes, err error) {
	res = new(api.VaultTreeRes)
	res.Tree, err = service.Vault().Tree(ctx, req.Id, req.Space)
	return
}

func (c *vaultController) VaultFileGet(ctx context.Context, req *api.VaultFileGetReq) (res *api.VaultFileGetRes, err error) {
	return service.Vault().ReadFile(ctx, req.Id, req.Path)
}

func (c *vaultController) VaultFileRaw(ctx context.Context, req *api.VaultFileRawReq) (res *api.VaultFileRawRes, err error) {
	// 二进制流直出：不走统一 JSON 响应，控制器内直接写 Response
	if err := service.Vault().ServeRaw(ctx, g.RequestFromCtx(ctx), req.Id, req.Path); err != nil {
		g.RequestFromCtx(ctx).Response.WriteStatus(404, err.Error())
	}
	return
}

func (c *vaultController) VaultFileWrite(ctx context.Context, req *api.VaultFileWriteReq) (res *api.VaultFileWriteRes, err error) {
	return service.Vault().WriteFile(ctx, req.Id, req.Path, req.Content)
}

func (c *vaultController) VaultFilePatch(ctx context.Context, req *api.VaultFilePatchReq) (res *api.VaultFilePatchRes, err error) {
	return service.Vault().PatchFile(ctx, req.Id, req.Path, req.Operations)
}

func (c *vaultController) VaultFolderCreate(ctx context.Context, req *api.VaultFolderCreateReq) (res *api.VaultFolderCreateRes, err error) {
	res = new(api.VaultFolderCreateRes)
	err = service.Vault().CreateFolder(ctx, req.Id, req.Path)
	return
}

func (c *vaultController) VaultUpload(ctx context.Context, req *api.VaultUploadReq) (res *api.VaultUploadRes, err error) {
	return service.Vault().Upload(ctx, req.Id, req.Path, req.File)
}

func (c *vaultController) VaultFileMove(ctx context.Context, req *api.VaultFileMoveReq) (res *api.VaultFileMoveRes, err error) {
	res = new(api.VaultFileMoveRes)
	err = service.Vault().Move(ctx, req.Id, req.From, req.To)
	return
}

func (c *vaultController) VaultFileDelete(ctx context.Context, req *api.VaultFileDeleteReq) (res *api.VaultFileDeleteRes, err error) {
	res = new(api.VaultFileDeleteRes)
	err = service.Vault().Delete(ctx, req.Id, req.Path)
	return
}

func (c *vaultController) VaultMetaUpdate(ctx context.Context, req *api.VaultMetaUpdateReq) (res *api.VaultMetaUpdateRes, err error) {
	res = new(api.VaultMetaUpdateRes)
	err = service.Vault().UpdateMeta(ctx, req.Id, req)
	return
}

func (c *vaultController) VaultSearch(ctx context.Context, req *api.VaultSearchReq) (res *api.VaultSearchRes, err error) {
	res = new(api.VaultSearchRes)
	res.Items, err = service.Vault().Search(ctx, req.Id, req.Q, req.Space)
	return
}

func (c *vaultController) VaultLinked(ctx context.Context, req *api.VaultLinkedReq) (res *api.VaultLinkedRes, err error) {
	res = new(api.VaultLinkedRes)
	target := ""
	switch {
	case req.Task > 0:
		target = "task:" + intToString(req.Task)
	case req.Req > 0:
		target = "req:" + intToString(req.Req)
	case req.Tc > 0:
		target = "tc:" + intToString(req.Tc)
	default:
		return res, nil
	}
	res.Items, err = service.Vault().Linked(ctx, req.Id, target)
	return
}

func (c *vaultController) VaultRefresh(ctx context.Context, req *api.VaultRefreshReq) (res *api.VaultRefreshRes, err error) {
	return service.Vault().Refresh(ctx, req.Id)
}

// ==================== KV 记忆 ====================

func (c *vaultController) MemoryList(ctx context.Context, req *api.MemoryListReq) (res *api.MemoryListRes, err error) {
	res = new(api.MemoryListRes)
	res.List, err = service.Vault().MemList(ctx, req.Id, req.Prefix, req.Include)
	return
}

func (c *vaultController) MemoryGet(ctx context.Context, req *api.MemoryGetReq) (res *api.MemoryGetRes, err error) {
	return service.Vault().MemGet(ctx, req.Id, req.Key)
}

func (c *vaultController) MemorySet(ctx context.Context, req *api.MemorySetReq) (res *api.MemorySetRes, err error) {
	res = new(api.MemorySetRes)
	err = service.Vault().MemSet(ctx, req.Id, req.Key, req)
	return
}

func (c *vaultController) MemoryVerify(ctx context.Context, req *api.MemoryVerifyReq) (res *api.MemoryVerifyRes, err error) {
	res = new(api.MemoryVerifyRes)
	err = service.Vault().MemVerify(ctx, req.Id, req.Key, int64(perm.UserId(ctx)))
	return
}

func (c *vaultController) MemoryExpire(ctx context.Context, req *api.MemoryExpireReq) (res *api.MemoryExpireRes, err error) {
	res = new(api.MemoryExpireRes)
	err = service.Vault().MemExpire(ctx, req.Id, req.Key)
	return
}

func (c *vaultController) MemoryDelete(ctx context.Context, req *api.MemoryDeleteReq) (res *api.MemoryDeleteRes, err error) {
	res = new(api.MemoryDeleteRes)
	err = service.Vault().MemDelete(ctx, req.Id, req.Key)
	return
}

func intToString(n int) string {
	return g.NewVar(n).String()
}
