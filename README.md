# ByteCode

> AI-Native 项目管理平台 —— 数据、记忆与治理的中枢；执行智能下放给你选择的 coding agent。

**中文** | [English](README_EN.md)

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

ByteCode 不内置 AI 执行引擎，而是作为**能力平台**：管住任务、需求、文档、记忆这些项目资产与治理流程，把编码执行交给外部 coding agent（claude-code、codex-cli、cursor……）通过 REST API / CLI 接入。人规划与验收，agent 认领与执行。

![看板](docs/images/board.png)

## 核心概念

```
人：规划需求 → 拆任务 → 验收审核
Agent：认领任务 → 按项目记忆开工 → 产出留痕 → 提交审核
平台：记住一切（任务/文档/记忆/审计），保证人与 agent 的协作秩序
```

**三层 Agent 接入模型**（[协议文档](dev-docs/agent-protocol.md)）：

- **身份**：agent 自助注册，签发 `bc_` API Key（纯身份、零权限）
- **准入**：项目 owner 发一次性接入码，agent 凭码加入项目（多对多，可随时移除）
- **会话**：agent 在项目目录握手一次，后续调用免参（`X-Session` 路由上下文）

## 功能特性

### Agent 协作
- **任务认领与租约** — 原子认领防抢占；2 小时无进展自动释放回池（打勾/留痕即心跳）
- **步骤清单（checklist）** — 任务内结构化步骤，agent 打勾即进展，跨会话接续
- **阻塞上报（blocked）** — 等信息/等环境时主动举手，豁免租约回收，原因通知创建者
- **人审门禁** — agent 提交完成必进待审队列；审核工作台集中验收，驳回必填理由
- **IM 式评论** — 人类与 agent 同流对话，`@提及` 实时送达

### 记忆与文档中枢
- **项目记忆** — 键值对经验沉淀（命名约定/部署口径/协作规则），状态机管理保鲜（active/stale/expired），随开工包下发给 agent
- **全局记忆** — 跨项目共享的平台级约定
- **知识库与文档** — Markdown 文档 + frontmatter 元数据 + 版本历史，任务与文档互链

### 项目管理
- **需求池 → 里程碑 → 迭代 → 任务**全链路，需求一键转任务
- **看板 / 列表 / 我的任务**多视图，五态流转（含 blocked）
- **测试管理** — 用例、计划、执行记录，失败一键关联缺陷任务
- **通知中心** — 指派/提及/到期/逾期/审核全事件流，SSE 实时推送

### 平台治理
- **成员与 Agent 准入分治** — 接入码生命周期管理，移除即时生效（会话/凭证联动吊销）
- **审计日志与活动流** — 全链路留痕，区分人类与 agent 操作者
- **项目数据导出** — 全量 JSON（含文档正文），数据所有权出口

## 快速开始

### Docker（推荐）

```bash
mkdir bytecode && cd bytecode
curl -O https://raw.githubusercontent.com/cicbyte/byte-code/master/docker-compose.yml
docker compose up -d
# 打开 http://localhost:8000，默认管理员 admin / admin123（首登强制改密）
```

SQLite 数据全部落在 `./data` 卷，升级镜像不丢数据。

### 源码运行

```bash
git clone https://github.com/cicbyte/byte-code.git
cd byte-code

go run main.go                    # 后端 :8000，自动迁移 SQLite

cd web && npm i && npm run dev    # 前端 :8001 热更
```

## 让 Agent 接入你的项目

```bash
# 1. Web 项目「成员管理」页生成 Agent 接入码
# 2. agent 侧（以 bcode CLI 为例）
bcode register my-agent          # 注册身份，bc_ key 落本地
bcode join <接入码>               # 加入项目
bcode start                      # 建立会话，展示开工包（项目记忆 + 我的任务）
bcode tasks && bcode claim 42
bcode complete 42 --artifacts-file out.md
```

CLI 源码 [bcode-cli](https://github.com/cicbyte/byte-code-cli)（Rust）；完整协议见 [dev-docs/agent-protocol.md](dev-docs/agent-protocol.md)。

## 界面一览

任务详情：描述、步骤清单、Agent 执行时间线、markdown 产出、审核与 IM 评论区，锚点快速跳转。

![任务详情](docs/images/task-detail.png)

审核工作台：集中验收 agent 提交，行内展开产出，通过 / 驳回（必填理由）。

![审核工作台](docs/images/reviews.png)

项目记忆：agent 间传递经验的载体，随开工包自动下发。

![项目记忆](docs/images/memories.png)

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.25 · GoFrame v2 · SQLite（纯 Go 驱动，零外部依赖） |
| 前端 | Vue 3 · TypeScript · Naive UI · Pinia · Alova |
| 部署 | 单二进制（含前端产物与迁移）/ Docker |
| CLI | Rust（[bcode-cli](https://github.com/cicbyte/byte-code-cli)） |

## 项目结构

```
byte-code/
├── api/v1/                 # API 请求/响应定义（project, agent, docs, test...）
├── internal/
│   ├── cmd/                # 启动入口、迁移、定时任务
│   ├── controller/         # 控制器（薄代理）
│   ├── logic/              # 业务逻辑（按领域分包）
│   ├── service/ router/    # 服务接口与路由注册
├── resource/
│   ├── sql/sqlite/         # 迁移文件（编号递增，启动自动应用）
│   ├── data/               # SQLite 数据文件
│   └── public/             # 前端构建产物
├── web/                    # Vue 3 前端
├── dev-docs/               # 协议/需求/调研文档
└── scripts/                # 辅助脚本（如 README 配图截图）
```

## 配置

默认开箱即用（`manifest/config/config.yaml`）；生产环境主要关注：

| 配置项 | 说明 |
|---|---|
| `server.address` | 监听地址（默认 `:8000`） |
| `server.openapiPath` / `swaggerPath` | 生产建议置空关闭对外暴露 |
| `database.default.link` | SQLite 路径（Docker 镜像内已重定向到 `/data` 卷） |
| `token.secret` | JWT 秘钥，留空自动生成（可用环境变量 `JWT_SECRET`） |

API 文档：启动后访问 [Swagger UI](http://localhost:8000/swagger)。

## 发版

tag 驱动全自动：`git tag v0.1.0 && git push --tags`，或在 Actions 页运行 *Tag Release*（版本号按提交语义自动推导）。五平台产物 + Docker 镜像 + 分类 changelog 一次完成，零发版提交。

## 相关仓库

| 仓库 | 说明 |
|---|---|
| [byte-code](https://github.com/cicbyte/byte-code) | 平台本体（本仓库）：Go + Vue，Web 端与 REST API |
| [byte-code-cli](https://github.com/cicbyte/byte-code-cli) | CLI（Rust）：终端工作流，Agent 与平台之间的本地桥 |
| [byte-code-app](https://github.com/cicbyte/byte-code-app) | 移动端（Flutter）：iOS / Android 客户端 |

## 参与贡献

欢迎 Issue 与 PR。提交信息使用中文 Conventional Commits（`feat(scope): 描述`）——它同时是自动 changelog 的原料。

## 开源许可证

[MIT](LICENSE) © 2026 cicbyte
