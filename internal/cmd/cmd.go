package cmd

import (
	"context"
	"strings"

	_ "github.com/cicbyte/byte-code/internal/logic"
	logicAiengine "github.com/cicbyte/byte-code/internal/logic/aiengine"
	"github.com/cicbyte/byte-code/internal/router"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/auditwriter"
	"github.com/cicbyte/byte-code/utility/dbclean"
	"github.com/cicbyte/byte-code/utility/dbbackup"
	"github.com/cicbyte/byte-code/utility/dbinit"
	"github.com/cicbyte/byte-code/utility/docs"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gcron"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 自动数据库迁移
			if err := dbinit.AutoMigrate(ctx); err != nil {
				g.Log().Fatalf(ctx, "Database migration failed: %v", err)
			}

			// 记忆/文档中枢：存量一次性导出 + 索引首扫 + 记忆物化全部异步首跑——
			// vault 很大时同步前置会让 HTTP 迟迟不监听（健康检查/容器编排判死）。
			// 迁移必须同步（schema 就绪），其余失败只告警不阻塞
			go func() {
				bgCtx := context.Background()
				if err := docs.ExportLegacyDocs(bgCtx); err != nil {
					g.Log().Warningf(bgCtx, "legacy docs export failed: %v", err)
				}
				if err := docs.ScanAll(bgCtx); err != nil {
					g.Log().Warningf(bgCtx, "docs initial scan failed: %v", err)
				}
				if err := service.Docs().MemMaterialize(bgCtx); err != nil {
					g.Log().Warningf(bgCtx, "memory materialize failed: %v", err)
				}
			}()
			if _, err := gcron.AddSingleton(ctx, "0 10 3 * * *", func(ctx context.Context) {
				if err := docs.ScanAll(ctx); err != nil {
					g.Log().Warningf(ctx, "docs scheduled scan failed: %v", err)
				}
			}); err != nil {
				g.Log().Warningf(ctx, "schedule docs scan failed: %v", err)
			}
			if _, err := gcron.AddSingleton(ctx, "0 20 3 * * *", func(ctx context.Context) {
				if err := service.Docs().MemMaterialize(ctx); err != nil {
					g.Log().Warningf(ctx, "memory materialize failed: %v", err)
				}
			}); err != nil {
				g.Log().Warningf(ctx, "schedule memory materialize failed: %v", err)
			}

			// AI 执行引擎自启：开关持久化为开则恢复运行（配置在 sys_config）
			logicAiengine.RestoreEngine(ctx)

			// 只增表定期清理：启动即执行一次，此后每天 03:00 执行
			dbclean.Run(ctx)
			if _, err := gcron.AddSingleton(ctx, "0 0 3 * * *", func(ctx context.Context) {
				dbclean.Run(ctx)
			}); err != nil {
				g.Log().Warningf(ctx, "schedule dbclean failed: %v", err)
			}

			// 数据库在线备份：启动即执行一次，此后每天 03:30 执行（错开清理任务）
			dbbackup.Run(ctx)
			if _, err := gcron.AddSingleton(ctx, "0 30 3 * * *", func(ctx context.Context) {
				dbbackup.Run(ctx)
			}); err != nil {
				g.Log().Warningf(ctx, "schedule dbbackup failed: %v", err)
			}

			// 审计日志异步落盘协程
			auditwriter.Start()

			s := g.Server()
			// 前端构建产物静态服务：resource/public 作为站点静态根
			s.SetServerRoot("resource/public")
			s.Group("/", func(group *ghttp.RouterGroup) {
				// 审计中间件须挂在 HandlerResponse 之外：标准控制器的响应体由
				// HandlerResponse 在链条 unwind 时写入，内层恢复时读不到（创建类
				// 操作无法从响应体补全新实体 id）
				group.Middleware(service.Middleware().MiddlewareAuditLog)
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				r := &router.Router{}
				r.BindController(ctx, group)

			// 添加SPA路由回退支持，处理Vue Router的HTML5 History模式
			group.Hook("/*", ghttp.HookBeforeServe, func(r *ghttp.Request) {
				path := r.URL.Path

				// 如果是API请求，跳过SPA回退
				if strings.HasPrefix(path, "/api/") {
					return
				}

				// 如果是静态资源文件（有文件扩展名），跳过SPA回退
				if strings.Contains(path, ".") && !strings.HasSuffix(path, "/") {
					return
				}

				// 其余路径（含站点根路径 /）都返回 index.html，让 Vue Router 处理
				// index.html 必须禁缓存：它引用带 hash 的资源文件，缓存会导致发版后
				// 用户一直加载旧 bundle（hash 资源自身可长缓存）
				r.Response.Header().Set("Cache-Control", "no-cache")
				r.Response.ServeFile("resource/public/index.html")
				r.ExitAll()
			})
			})
			s.Run()
			// ghttp 优雅关停完成后 Run 返回：排空审计缓冲再退出
			// （否则最后 flushInterval 窗口内的审计随进程一起丢）
			auditwriter.Stop()
			return nil
		},
	}
)