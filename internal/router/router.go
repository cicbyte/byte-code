package router

import (
	"context"

	controller "github.com/cicbyte/byte-code/internal/controller"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/library/libRouter"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Router struct{}

func (router *Router) BindController(ctx context.Context, group *ghttp.RouterGroup) {
	// 公开 API（无需认证）
	group.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(service.Middleware().MiddlewareCORS)
		group.Bind(
			controller.Auth.Login,
			controller.AiUserCtrl.AiLogin,
		)
	})

	// 需要 Token 认证的 API
	group.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(service.Middleware().MiddlewareCORS)
		group.Middleware(service.Middleware().MiddlewareTokenAuth)

		group.Bind(
			controller.Auth.AdminInfo,
			controller.Auth.Logout,
			controller.Menu.Menus,
			controller.Menu.MenuList,
			controller.Role.List,
			controller.DashboardCtrl.Console,
			controller.Setting.GetProfile,
			controller.Setting.UpdateProfile,
			controller.Setting.ChangePassword,
			controller.Setting.GetSystemConfig,
			controller.Setting.UpdateSystemConfig,
		)

		// v1 版本 API（需要认证）
		group.Group("/v1", func(group *ghttp.RouterGroup) {
			group.Bind(
				controller.Categories,
				controller.Health,
			)

			// 项目管理（含需求/里程碑，B 类化后挂在项目下）
			group.Bind(
				controller.ProjectCtrl,
			)

			// 数据库模型管理
			group.Bind(
				controller.DatabaseCtrl,
			)

			// 测试管理
			group.Bind(
				controller.Test,
			)

			// 知识库
			group.Bind(
				controller.Doc,
			)

			// 附件管理
			group.Bind(
				controller.AttachmentCtrl,
			)

			// AI 用户管理
			group.Bind(
				controller.AiUserCtrl,
			)

			// 平台功能（标签、活动流、通知、搜索、仪表盘、审计日志）
			group.Bind(
				controller.PlatformCtrl,
			)
		})

		// 自动绑定定义的控制器
		if err := libRouter.RouterAutoBind(ctx, router, group); err != nil {
			panic(err)
		}
	})
}
