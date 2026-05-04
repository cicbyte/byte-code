# ByteCode

> AI-Native 项目管理平台，集成产品管理、任务追踪、测试管理、知识库与 AI 智能体，开箱即用、零外部依赖。

**中文** | [English](README_EN.md)


## 功能特性

### 项目与产品管理
- **产品线管理** — 多产品线下的项目组织，需求、里程碑追踪
- **PingCode 风格导航** — 左侧模块选择 + 顶部实体子导航，钻入式体验
- **数据库模型管理** — 项目级数据库连接配置，可视化表/字段编辑器，完整 schema 变更历史

### 任务与 Sprint
- **任务看板** — 可视化任务状态流转
- **任务列表** — 多维度筛选、排序、批量操作
- **Sprint 管理** — 迭代规划、燃尽图、任务导入
- **AI 认领与完成** — AI 智能体自动认领、执行、提交结果

### 团队协作
- **成员管理** — 项目级角色权限
- **评论系统** — 人工与 AI 用户统一评论流
- **活动流 & 通知** — 实时项目动态
- **标签体系** — 跨实体标签分类

### 平台能力
- **测试管理** — 测试用例、测试计划、执行记录
- **知识库** — 文档管理，Markdown 编辑
- **附件管理** — S3 兼容存储（RustFS/MinIO）
- **系统管理** — 用户、角色、菜单、审计日志

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- pnpm

### 克隆并运行

```bash
git clone https://github.com/cicbyte/byte-code.git
cd byte-code

# 启动后端（自动迁移 SQLite 数据库）
go run main.go

# 前端开发（新终端）
cd web
pnpm install
pnpm dev
```

后端运行在 [http://localhost:8000](http://localhost:8000)，前端开发服务器默认 [http://localhost:8002](http://localhost:8002)。

默认管理员账号：`admin` / `123456`

## 技术栈

| 层级 | 技术 |
|---|---|
| 前端 | Vue 3 + TypeScript + Naive UI |
| 状态管理 | Pinia |
| HTTP 客户端 | Alova v3 |
| 后端 | Go + GoFrame v2 |
| 数据库 | SQLite（零配置，数据文件 `resource/data/app.db`） |
| 认证 | JWT Token |
| 文件存储 | S3 兼容（RustFS / MinIO） |

## 项目结构

```
byte-code/
├── api/                    # API 请求/响应结构定义
│   └── v1/                 # v1 版本 API（project, product, database...）
├── internal/
│   ├── cmd/                # 启动入口、数据库迁移
│   ├── controller/         # 控制器层（薄代理）
│   ├── logic/              # 业务逻辑实现
│   ├── model/              # 数据模型
│   ├── service/            # 服务接口定义
│   └── router/             # 路由注册
├── resource/
│   ├── sql/sqlite/         # 数据库迁移文件（编号递增）
│   ├── data/               # SQLite 数据文件
│   └── public/             # 前端构建产物（生产模式）
├── web/                    # Vue 3 前端
│   ├── src/
│   │   ├── api/            # 前端 API 层
│   │   ├── config/         # 导航配置等
│   │   ├── layout/         # 布局组件（侧边栏、Header、EntityNavBar）
│   │   ├── router/         # 前端路由
│   │   ├── store/          # Pinia 状态管理
│   │   └── views/          # 页面组件
│   └── package.json
├── main.go                 # 入口文件
└── Makefile                # 构建命令
```

## 配置

配置文件：`manifest/config/config.yaml`

| 配置项 | 说明 | 默认值 |
|---|---|---|
| `server.address` | 后端监听地址 | `:8000` |
| `server.openapiPath` | OpenAPI 文档路径 | `/api.json` |
| `server.swaggerPath` | Swagger UI 路径 | `/swagger` |
| `database.default.link` | 数据库连接 | `sqlite::@file(./resource/data/app.db)` |

## 构建

```bash
# 构建后端
go build -o byte-code .

# 构建前端
cd web && pnpm build

# 生产模式运行（后端直接托管前端静态文件）
./byte-code
```

## 开发命令

```bash
make build        # 构建二进制
make dao          # 生成 DAO 层代码
make service      # 生成 Service 接口
make ctrl         # 生成 Controller 代码
```

## API 文档

启动后端后访问 Swagger UI：[http://localhost:8000/swagger](http://localhost:8000/swagger)

## 参与贡献

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/xxx`)
3. 提交变更 (`git commit -m 'feat: xxx'`)
4. 推送分支 (`git push origin feature/xxx`)
5. 创建 Pull Request

## 开源许可证

MIT License
