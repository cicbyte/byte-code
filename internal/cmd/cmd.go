package cmd

import (
	"context"
	"github.com/cicbyte/byte-code/internal/version"
	"runtime/debug"
	"strings"
	"time"

	_ "github.com/cicbyte/byte-code/internal/logic"
	logicAiengine "github.com/cicbyte/byte-code/internal/logic/aiengine"
	"github.com/cicbyte/byte-code/internal/logic/project"
	"github.com/cicbyte/byte-code/internal/router"
	"github.com/cicbyte/byte-code/internal/service"
	"github.com/cicbyte/byte-code/utility/auditwriter"
	"github.com/cicbyte/byte-code/utility/dbbackup"
	"github.com/cicbyte/byte-code/utility/dbclean"
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
			// 裸 goroutine 不受 gcron 的 panic recover 保护，走 safeGo（见文件尾）
			safeGo(func(bgCtx context.Context) {
				if err := docs.ExportLegacyDocs(bgCtx); err != nil {
					g.Log().Warningf(bgCtx, "legacy docs export failed: %v", err)
				}
				// 文档扫描/记忆物化/清理/备份统一走下方 runMaintenance（含启动补跑），
				// 此处再跑一遍会与维护 goroutine 并发叠跑
				if err := project.ScanDueTasks(bgCtx); err != nil {
					g.Log().Warningf(bgCtx, "task due scan failed: %v", err)
				}
			})
			// 任务租约释放：每小时扫一次失联 agent 的认领（2h 无进展回 open）
			if _, err := gcron.AddSingleton(ctx, "0 0 * * * *", func(ctx context.Context) {
				if err := project.ReleaseStaleClaims(ctx); err != nil {
					g.Log().Warningf(ctx, "schedule lease release failed: %v", err)
				}
			}); err != nil {
				g.Log().Warningf(ctx, "schedule lease release failed: %v", err)
			}

			// 任务到期提醒：每天 09:00（工作时段推送；启动首跑见上方异步块）
			if _, err := gcron.AddSingleton(ctx, "0 0 9 * * *", func(ctx context.Context) {
				if err := project.ScanDueTasks(ctx); err != nil {
					g.Log().Warningf(ctx, "schedule task due scan failed: %v", err)
				}
			}); err != nil {
				g.Log().Warningf(ctx, "schedule task due scan failed: %v", err)
			}

			// AI 执行引擎自启：开关持久化为开则恢复运行（配置在 sys_config）
			logicAiengine.RestoreEngine(ctx)

			// 03:00 维护窗口单入口串行执行：清理 → 文档扫描 → 记忆物化 → 备份。
			// 原先四个独立 cron 各自错峰 10 分钟，但互不防叠——前序任务跑超时
			// 即与备份（VACUUM INTO）并发写，SQLite 单写者下只是紧张而非死锁，
			// 串行化后彻底消除叠跑窗口；单步失败告警不阻断后续步骤
			maintenance := []struct {
				name string
				run  func(context.Context) error
			}{
				{"dbclean", func(c context.Context) error { dbclean.Run(c); return nil }},
				{"docs scan", func(c context.Context) error { return docs.ScanAll(c) }},
				{"mem materialize", func(c context.Context) error { return service.Docs().MemMaterialize(c) }},
				{"db backup", func(c context.Context) error { dbbackup.Run(c); return nil }},
			}
			// 每步限时：串行链任一步卡死会把后续步骤（尤其备份）无限推迟且无
			// 告警；超时告警后放行下一步
			runMaintenance := func(c context.Context) {
				for _, step := range maintenance {
					stepCtx, cancel := context.WithTimeout(c, 30*time.Minute)
					if err := step.run(stepCtx); err != nil {
						g.Log().Warningf(stepCtx, "maintenance %s failed: %v", step.name, err)
					}
					cancel()
				}
			}
			// 启动即跑一次（错过夜间窗口的补执行）；gcron 的 recover 罩不到裸 goroutine
			safeGo(runMaintenance)
			if _, err := gcron.AddSingleton(ctx, "0 0 3 * * *", runMaintenance); err != nil {
				g.Log().Warningf(ctx, "schedule maintenance failed: %v", err)
			}

			// 审计日志异步落盘协程
			auditwriter.Start()

			g.Log().Infof(ctx, "bytecode %s starting", version.Version)
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

// safeGo 带 recover 的后台 goroutine：gcron 内置 panic 恢复，但裸 go 出去的
// 任务（启动补跑/维护链首跑）无人兜底——任一 panic 会直接崩掉整个进程
func safeGo(fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(context.Background(), "background goroutine panicked: %v, stack: %s", r, debug.Stack())
			}
		}()
		fn(context.Background())
	}()
}
