# 工作日志

## 2026-08-29

- 修复 JWT 秘钥硬编码（安全审计 H-1）：`internal/logic/auth/jwt.go` 移除源码硬编码秘钥，改为按「环境变量 `JWT_SECRET` > 配置 `token.secret` > `resource/data/jwt.secret` 文件 > 首次启动随机生成并落盘」加载，保持零配置开箱可用且每个部署实例秘钥唯一。
- 涉及文件：`internal/logic/auth/jwt.go`、`manifest/config/config.yaml`（新增 `token.secret` 配置项）、`.gitignore`（忽略 jwt.secret）、`README.md`（配置表补一行）。
- 验收：`go build` / `go vet` 通过；隔离环境实测登录签发、鉴权访问、伪造 token 拒绝（401）、重启后秘钥复用（token 不失效）、环境变量优先级覆盖，全部通过。注意：升级后旧 token 全部失效需重新登录。

## 2026-08-29（修复批次 2：H-4）

- 修复 categories 列表 OrderBy SQL 注入（安全审计 H-4），双层防御：`model.PageReq.OrderBy` 增加 regex 格式校验 tag（覆盖全部嵌入该结构的列表接口）；新增 `model.SafeOrderBy()` 白名单映射函数，`internal/logic/categories/categories.go` 消费点限定五列白名单，非法输入回退默认排序。
- 涉及文件：`internal/model/common.go`、`internal/logic/categories/categories.go`。
- 验收：修复前 PoC 实证子查询注入可执行（排序随 CASE 表达式变化）；修复后注入载荷被校验层拦截（code 51）、`sort desc`/`name` 等白名单值排序正确、不带 orderBy 默认行为不变、role 列表等嵌入 PageReq 的接口无回归。`go build` / `go vet` 通过。

## 2026-08-29（修复批次 3：H-6）

- 修复默认弱口令无强制修改 + README 密码文档错误（安全审计 H-6）：新增迁移 39 为 `sys_users` 加 `must_change_password` 标记（仅默认口令哈希未改的账号置 1）；登录返回 `mustChangePassword`；TokenAuth 中间件拦截标记账号的业务请求（code 1001，仅放行改密/登出/用户信息）；`ChangePassword` 成功后清除标记；前端登录页移除默认账密预填、登录与 HTTP 拦截器处理强制改密跳转；README 更正默认口令为 admin/admin123 并说明强制改密。
- 涉及文件：`resource/sql/sqlite/39_alter_sys_users_must_change_password.sql`（新增）、`api/v1/auth/auth.go`、`internal/service/auth.go`、`internal/controller/auth.go`、`internal/logic/auth/auth.go`、`internal/logic/middleware/middleware.go`、`internal/logic/setting/setting.go`、`web/src/enums/httpEnum.ts`、`web/src/utils/http/alova/index.ts`、`web/src/views/login/index.vue`、`README.md`。
- 验收：隔离环境实测新库全流程（登录标记→业务拦截 1001→白名单放行→改密→同 token 立即放行→旧密码失效），存量库场景（已改密码后重跑迁移 39 不误标记）；`go build`/`go vet`、前端 `pnpm build` 通过。
- 事故记录：测试中发现 GoFrame 对配置中相对路径 SQLite link 的解析不是基于进程工作目录，导致 21:28-21:39 的"隔离"测试实例实际读写的是项目真实库 `resource/data/app.db`（该库当时无业务数据，仅种子数据）。影响：admin 密码被测试改为临时值、迁移 39 提前应用。已补救：admin 恢复默认口令哈希并按设计置 `must_change_password=1`，清除测试 token。后续测试改用绝对路径数据库链接实现真隔离（已验证生效）。
