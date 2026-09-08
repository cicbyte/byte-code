# 多阶段构建（参考 forks）：前端 → Go 二进制（版本注入）→ alpine 运行层。
# byte-code 特性：resource/（public+sql）运行时读取、SQLite 数据落 /data。

# ====== 前端 ======
FROM node:22-alpine AS frontend
RUN corepack enable && corepack prepare pnpm@latest-10 --activate
WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm run build

# ====== 后端 ======
FROM golang:1.25-alpine AS backend
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ARG VERSION=dev
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/cicbyte/byte-code/internal/version.Version=${VERSION}" -o bytecode .

# ====== 运行镜像 ======
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend /build/bytecode /app/bytecode
COPY --from=backend /build/resource/sql /app/resource/sql
COPY --from=frontend /web/dist /app/resource/public
COPY manifest/config/config.yaml /app/config.yaml
# 开发配置的 SQLite 是相对路径（./resource/data/app.db）；容器内重定向到
# /data 卷（升级镜像数据不丢）。backup/projects 等数据面同样收口到 /data。
# 生产镜像同时关闭 OpenAPI/Swagger 暴露（挂载自定义 config 覆盖者不受影响）
RUN sed -i 's|./resource/data/app.db|/data/app.db|; s|openapiPath: "/api.json"|openapiPath: ""|; s|swaggerPath: "/swagger"|swaggerPath: ""|' /app/config.yaml && \
    mkdir -p /data

# SQLite 库与备份默认落 /data
ENV TZ=Asia/Shanghai
VOLUME /data
EXPOSE 8000

ENTRYPOINT ["/app/bytecode"]
