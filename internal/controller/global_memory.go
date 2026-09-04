package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/docs"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/perm"
)

// GlobalMemories 全局记忆（project_memories 表 project_id=0 行）：
// 独立控制器而非挂在 vaultController 上——后者被路由整体绑定在认证组，
// 全局记忆需要读写分组（读全员、写仅管理员），必须逐方法注册
var GlobalMemories = globalMemoryController{}

type globalMemoryController struct{}

func (c *globalMemoryController) List(ctx context.Context, req *api.GlobalMemoryListReq) (res *api.GlobalMemoryListRes, err error) {
	res = new(api.GlobalMemoryListRes)
	res.List, err = service.Docs().MemList(ctx, 0, req.Prefix, req.Include)
	return
}

func (c *globalMemoryController) Get(ctx context.Context, req *api.GlobalMemoryGetReq) (res *api.GlobalMemoryGetRes, err error) {
	r, err := service.Docs().MemGet(ctx, 0, req.Key)
	if err != nil {
		return nil, err
	}
	return &api.GlobalMemoryGetRes{MemoryItem: r.MemoryItem}, nil
}

func (c *globalMemoryController) Set(ctx context.Context, req *api.GlobalMemorySetReq) (res *api.GlobalMemorySetRes, err error) {
	res = new(api.GlobalMemorySetRes)
	err = service.Docs().MemSet(ctx, 0, req.Key, &api.MemorySetReq{
		Value: req.Value, Ttl: req.Ttl, Status: req.Status,
	})
	return
}

func (c *globalMemoryController) Verify(ctx context.Context, req *api.GlobalMemoryVerifyReq) (res *api.GlobalMemoryVerifyRes, err error) {
	res = new(api.GlobalMemoryVerifyRes)
	err = service.Docs().MemVerify(ctx, 0, req.Key, int64(perm.UserId(ctx)))
	return
}

func (c *globalMemoryController) Expire(ctx context.Context, req *api.GlobalMemoryExpireReq) (res *api.GlobalMemoryExpireRes, err error) {
	res = new(api.GlobalMemoryExpireRes)
	err = service.Docs().MemExpire(ctx, 0, req.Key)
	return
}

func (c *globalMemoryController) Delete(ctx context.Context, req *api.GlobalMemoryDeleteReq) (res *api.GlobalMemoryDeleteRes, err error) {
	res = new(api.GlobalMemoryDeleteRes)
	err = service.Docs().MemDelete(ctx, 0, req.Key)
	return
}
