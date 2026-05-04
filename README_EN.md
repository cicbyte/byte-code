# ByteCode

> AI-Native 项目管理平台，集成产品管理、任务追踪、测试管理、知识库与 AI 智能体，开箱即用、零外部依赖。

[中文](README.md) | **English**


## Features

### Project & Product Management
- **Product Line Management** — Multi-product project organization with requirements and milestone tracking
- **PingCode-style Navigation** — Left sidebar modules + top entity sub-navigation for a drill-down experience
- **Database Model Management** — Project-level database connection config, visual table/field editor, full schema change history

### Tasks & Sprints
- **Task Board** — Visual task status workflow
- **Task List** — Multi-dimensional filtering, sorting, batch operations
- **Sprint Management** — Iteration planning, burndown charts, task import
- **AI Claim & Complete** — AI agents auto-claim, execute, and submit results

### Team Collaboration
- **Member Management** — Project-level roles and permissions
- **Comment System** — Unified comment stream for human and AI users
- **Activity Feed & Notifications** — Real-time project updates
- **Tagging System** — Cross-entity tag classification

### Platform Capabilities
- **Test Management** — Test cases, test plans, execution records
- **Knowledge Base** — Document management with Markdown editing
- **Attachment Management** — S3-compatible storage (RustFS / MinIO)
- **System Administration** — Users, roles, menus, audit logs

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm

### Clone & Run

```bash
git clone https://github.com/cicbyte/byte-code.git
cd byte-code

# Start backend (auto-migrates SQLite database)
go run main.go

# Frontend dev server (new terminal)
cd web
pnpm install
pnpm dev
```

Backend runs at [http://localhost:8000](http://localhost:8000), frontend dev server at [http://localhost:8002](http://localhost:8002).

Default admin credentials: `admin` / `123456`

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3 + TypeScript + Naive UI |
| State Management | Pinia |
| HTTP Client | Alova v3 |
| Backend | Go + GoFrame v2 |
| Database | SQLite (zero config, data file at `resource/data/app.db`) |
| Authentication | JWT Token |
| File Storage | S3-compatible (RustFS / MinIO) |

## Project Structure

```
byte-code/
├── api/                    # API request/response struct definitions
│   └── v1/                 # v1 API (project, product, database...)
├── internal/
│   ├── cmd/                # Entry point, database migration
│   ├── controller/         # Controller layer (thin proxy)
│   ├── logic/              # Business logic implementation
│   ├── model/              # Data models
│   ├── service/            # Service interface definitions
│   └── router/             # Route registration
├── resource/
│   ├── sql/sqlite/         # Database migration files (sequentially numbered)
│   ├── data/               # SQLite data file
│   └── public/             # Frontend build output (production mode)
├── web/                    # Vue 3 frontend
│   ├── src/
│   │   ├── api/            # Frontend API layer
│   │   ├── config/         # Navigation config, etc.
│   │   ├── layout/         # Layout components (Sidebar, Header, EntityNavBar)
│   │   ├── router/         # Frontend routes
│   │   ├── store/          # Pinia state management
│   │   └── views/          # Page components
│   └── package.json
├── main.go                 # Entry file
└── Makefile                # Build commands
```

## Configuration

Config file: `manifest/config/config.yaml`

| Setting | Description | Default |
|---|---|---|
| `server.address` | Backend listen address | `:8000` |
| `server.openapiPath` | OpenAPI doc path | `/api.json` |
| `server.swaggerPath` | Swagger UI path | `/swagger` |
| `database.default.link` | Database connection | `sqlite::@file(./resource/data/app.db)` |

## Build

```bash
# Build backend
go build -o byte-code .

# Build frontend
cd web && pnpm build

# Run in production (backend serves frontend static files)
./byte-code
```

## Development Commands

```bash
make build        # Build binary
make dao          # Generate DAO layer code
make service      # Generate Service interfaces
make ctrl         # Generate Controller code
```

## API Documentation

After starting the backend, access Swagger UI at: [http://localhost:8000/swagger](http://localhost:8000/swagger)

## Contributing

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/xxx`)
3. Commit your changes (`git commit -m 'feat: xxx'`)
4. Push the branch (`git push origin feature/xxx`)
5. Create a Pull Request

## License

MIT License
