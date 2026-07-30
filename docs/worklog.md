# 工作日志

## 2026-08-29

- 修复 JWT 秘钥硬编码（安全审计 H-1）：`internal/logic/auth/jwt.go` 移除源码硬编码秘钥，改为按「环境变量 `JWT_SECRET` > 配置 `token.secret` > `resource/data/jwt.secret` 文件 > 首次启动随机生成并落盘」加载，保持零配置开箱可用且每个部署实例秘钥唯一。
- 涉及文件：`internal/logic/auth/jwt.go`、`manifest/config/config.yaml`（新增 `token.secret` 配置项）、`.gitignore`（忽略 jwt.secret）、`README.md`（配置表补一行）。
- 验收：`go build` / `go vet` 通过；隔离环境实测登录签发、鉴权访问、伪造 token 拒绝（401）、重启后秘钥复用（token 不失效）、环境变量优先级覆盖，全部通过。注意：升级后旧 token 全部失效需重新登录。
