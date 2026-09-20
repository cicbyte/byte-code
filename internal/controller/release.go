package controller

import (
	"context"

	api "github.com/cicbyte/byte-code/api/v1/project"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/gogf/gf/v2/frame/g"
)

// ReleaseCtrl 发布公开访问控制器：分享直链下载免鉴权。
// 独立成控制器而非挂在 ProjectCtrl 上——ProjectCtrl 整绑在认证组，
// 公开端点须只进公开路由组（避免重复注册/被鉴权拦截）。
var ReleaseCtrl = releaseController{}

type releaseController struct {
	BaseController
}

// ReleaseFilePublic 公开直链下载（免鉴权）：直接写响应流/302，不走 JSON 壳
func (c *releaseController) ReleaseFilePublic(ctx context.Context, req *api.ReleaseFilePublicReq) (res *api.ReleaseFilePublicRes, err error) {
	service.Project().ServeReleaseFilePublic(ctx, g.RequestFromCtx(ctx), req.Token)
	return nil, nil
}
