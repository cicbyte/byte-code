package cmd

import (
	"context"
	"strings"

	_ "github.com/cicbyte/byte-code/internal/logic"
	"github.com/cicbyte/byte-code/internal/router"
	"github.com/cicbyte/byte-code/utility/dbclean"
	"github.com/cicbyte/byte-code/utility/dbinit"
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

			// 只增表定期清理：启动即执行一次，此后每天 03:00 执行
			dbclean.Run(ctx)
			if _, err := gcron.AddSingleton(ctx, "0 0 3 * * *", func(ctx context.Context) {
				dbclean.Run(ctx)
			}); err != nil {
				g.Log().Warningf(ctx, "schedule dbclean failed: %v", err)
			}

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
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

					// 对于其他所有路径，都返回index.html，让Vue Router处理
					if path != "/" && !strings.HasPrefix(path, "/api/") {
						r.Response.ServeFile("resource/public/html/index.html")
						r.ExitAll()
					}
				})
			})
			s.Run()
			return nil
		},
	}
)