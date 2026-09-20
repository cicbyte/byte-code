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
			controller.Auth.ForgotPassword,
			controller.Auth.ResetPassword,
			controller.AiUserCtrl.AiLogin,
			// 头像文件直读（<img> 无法携带认证头；头像属非敏感展示数据）
			controller.Setting.Avatar,
		)
		// Agent 自助注册（纯身份零权限；公开是有意的——权限由项目接入码把关）
		group.Group("/v1", func(group *ghttp.RouterGroup) {
			group.Bind(
				controller.AgentCtl.Register,
				// 发布分享直链（#552）：凭令牌免鉴权下载——分享即链接，
				// 吊销/过期即 404
				controller.ReleaseCtrl.ReleaseFilePublic,
			)
		})
	})

	// 需要 Token 认证的 API（所有登录用户）
	group.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(service.Middleware().MiddlewareCORS)
		group.Middleware(service.Middleware().MiddlewareTokenAuth)
		group.Middleware(service.Middleware().MiddlewareProjectAuth)
		// 使用埋点：认证后读+写全量（uid 依赖 TokenAuth；项目归属复用 ProjectAuth 的解析）
		group.Middleware(service.Middleware().MiddlewareUsageTrack)

		group.Bind(
			controller.Auth.AdminInfo,
			controller.Auth.Logout,
			// 用户搜索（添加成员选择器）：认证人类可用（controller 内拒 agent）
			controller.UserCtrl.Search,
			controller.Menu.Menus,
			controller.DashboardCtrl.Console,
			controller.Setting.GetProfile,
			controller.Setting.UpdateProfile,
			controller.Setting.UpdateAvatar,
			controller.Setting.ChangePassword,
		)

		// v1 版本 API（需要认证）
		group.Group("/v1", func(group *ghttp.RouterGroup) {
			group.Bind(
				controller.Health,
			)

			// 项目管理（含需求/里程碑，B 类化后挂在项目下）
			group.Bind(
				controller.ProjectCtrl,
			)
			// 项目导出单独绑（同控制器整绑也行，但导出是重接口，显式列出便于审计路由表）

			// 审核中心（跨项目聚合待审 + 批量审核；动作语义复用任务 ReviewTask）
			group.Bind(
				controller.Review,
			)

			// 数据库模型管理
			group.Bind()

			// 测试管理
			group.Bind(
				controller.Test,
			)

			// 记忆/文档中枢（磁盘真相源 + KV 记忆；项目归属由路径 /projects/{id}/ 中间件校验）
			group.Bind(
				controller.Docs,
			)

			// 全局记忆读取（project_id=0，跨项目通用约定；全员可读，
			// 写操作在下方管理组——不能整绑控制器，须逐方法注册）
			group.Bind(
				controller.GlobalMemories.List,
				controller.GlobalMemories.Get,
				controller.GlobalMemories.Propose,
			)

			// Agent 接入协议（认证组：bc key 或人类 token；
			// Join/Session/Tasks 业务层限定 agent 身份，接入码生成限项目 owner）
			group.Bind(
				controller.AgentCtl.Join,
				controller.AgentCtl.SessionCreate,
				controller.AgentCtl.AgentProjects,
				controller.AgentCtl.AgentTasks,
				controller.AgentCtl.AgentDocsTree,
				controller.AgentCtl.AgentDocsFile,
				controller.AgentCtl.AgentDocsWrite,
				controller.AgentCtl.AgentDocsSearch,
				controller.AgentCtl.JoinCodeCreate,
				controller.AgentCtl.AgentRemoveProject,
				controller.AgentCtl.AgentCapabilitiesUpdate,
			)

			// 附件：上传/下载/列表对所有登录用户开放（存储配置在管理组）
			group.Bind(
				controller.AttachmentCtrl.AttachmentUpload,
				controller.AttachmentCtrl.AttachmentGet,
				controller.AttachmentCtrl.AttachmentDownload,
				controller.AttachmentCtrl.AttachmentFile,
				controller.AttachmentCtrl.AttachmentPreview,
				controller.AttachmentCtrl.AttachmentDelete,
				controller.AttachmentCtrl.AttachmentList,
				controller.AttachmentCtrl.AttachmentUpdate,
			)

			// 平台功能（活动流、通知、搜索、统计；审计日志在管理组）。
			// 标签读/挂/摘全员；标签定义（建/改/删）改为权限字典放行
			// （perm.RequireMenuPerm platform_tags，role2 预绑——登录即可用标签）
			group.Bind(
				controller.PlatformCtrl.ListTags,
				controller.PlatformCtrl.CreateTag,
				controller.PlatformCtrl.UpdateTag,
				controller.PlatformCtrl.DeleteTag,
				controller.PlatformCtrl.AttachTag,
				controller.PlatformCtrl.DetachTag,
				controller.PlatformCtrl.GetTagEntities,
				controller.PlatformCtrl.ListActivities,
				controller.PlatformCtrl.ListNotifications,
				controller.PlatformCtrl.ReadNotification,
				controller.PlatformCtrl.ReadAllNotifications,
				controller.PlatformCtrl.UnreadCount,
				controller.PlatformCtrl.NotificationStream,
				controller.PlatformCtrl.Search,
				controller.PlatformCtrl.DashboardStats,
			)
		})

		// 自动绑定定义的控制器
		if err := libRouter.RouterAutoBind(ctx, router, group); err != nil {
			panic(err)
		}
	})

	// 管理 API（需要认证 + 超级管理员）
	group.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(service.Middleware().MiddlewareCORS)
		group.Middleware(service.Middleware().MiddlewareTokenAuth)
		group.Middleware(service.Middleware().MiddlewareAdminAuth)

		group.Bind(
			controller.Menu.MenuList,
			controller.Role.List,
			controller.Role.Create,
			controller.Role.Update,
			controller.Role.Delete,
			controller.Role.UpdateMenus,
			controller.Setting.GetSystemConfig,
			controller.Setting.UpdateSystemConfig,
		)

		group.Group("/v1", func(group *ghttp.RouterGroup) {
			// 全局记忆写操作 + 提案审核（影响所有项目的 AI 上下文，仅管理员）
			group.Bind(
				controller.GlobalMemories.Set,
				controller.GlobalMemories.ProposalList,
				controller.GlobalMemories.ProposalReview,
				controller.GlobalMemories.Verify,
				controller.GlobalMemories.Expire,
				controller.GlobalMemories.Delete,
			)

			// AI 用户管理（AiLogin 已在公开组）
			group.Bind(
				controller.AiUserCtrl.Create,
				controller.AiUserCtrl.Update,
				controller.AiUserCtrl.Delete,
				controller.AiUserCtrl.List,
				controller.AiUserCtrl.ResetKey,
				controller.AiUserCtrl.AgentBindings,
			)

			// S3 存储配置（含凭据）
			group.Bind(
				controller.AttachmentCtrl.StorageConfigGet,
				controller.AttachmentCtrl.StorageConfigUpdate,
				controller.AttachmentCtrl.StorageTest,
			)

			// 审计日志 + 使用分析 + 审计导出（均管理员）
			group.Bind(
				controller.PlatformCtrl.ListAuditLogs,
				controller.PlatformCtrl.AuditLogExport,
				controller.PlatformCtrl.UsageOverview,
				controller.PlatformCtrl.UsageReport,
			)

			// 用户管理
			group.Bind(controller.UserCtrl)

			// AI 引擎管理
			group.Bind(
				controller.Setting.GetAiEngineConfig,
				controller.Setting.UpdateAiEngineConfig,
				controller.Setting.ToggleAiEngine,
			)
		})
	})
}
