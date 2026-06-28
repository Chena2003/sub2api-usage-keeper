# Sub2API Usage Keeper

[English README](./README.en.md)

Sub2API Usage Keeper 是一个独立 sidecar dashboard，用于展示 Sub2API 上游账号状态、脱敏后的 quota/账号元数据，以及总体调用统计。

## 功能特性

- 读取现有 Sub2API PostgreSQL 数据库中的账号与 usage 聚合数据
- 展示总账号、可用账号、请求量、Token 消耗、成本和模型排行
- 展示所有上游账号的状态（Active/Paused/Error/Rate Limited/Overloaded/Temp Unschedulable）、计划、reset 倒计时和近 7 天用量
- 账号额度页面：三段式行布局展示身份/指标/双窗口额度条（5h + 7d），读取上游 `accounts.extra` 中的真实 utilization 使用率
- 浏览器 API 不返回 Sub2API 原始 credentials
- 本地 SQLite 存储 dashboard 自有数据、日志和备份
- 可选密码登录保护
- Docker / Docker Compose 部署

## Docker 接入现有 Sub2API

1. 确保 dashboard 容器和 `sub2api-postgres` 在同一个 Docker network。
2. 尽可能为 `SUB2API_DATABASE_URL` 配置只读 PostgreSQL 用户。
3. 默认将 dashboard 绑定到 `127.0.0.1:8082`，再通过 Caddy、Nginx 或 Cloudflare 暴露给用户。

```bash
cp .env.example .env
vim .env
docker build -t sub2api-usage-keeper:latest .
docker compose -f docker-compose.example.yml up -d
```

## 配置

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `SUB2API_DATABASE_URL` | 是 | - | Sub2API PostgreSQL 连接地址，例如 `postgres://sub2api:password@sub2api-postgres:5432/sub2api?sslmode=disable` |
| `AUTH_ENABLED` | 否 | `false` | 是否启用登录保护 |
| `LOGIN_PASSWORD` | 鉴权启用时必填 | - | Dashboard 登录密码 |
| `AUTH_SESSION_TTL` | 否 | `168h` | Session 生命周期 |
| `APP_PORT` | 否 | `8080` | HTTP 监听端口 |
| `APP_BASE_PATH` | 否 | 根路径 | 子路径部署前缀，例如 `/usage`；留空表示 `/` |
| `TZ` | 否 | `Asia/Shanghai` | 项目时区 |
| `WORK_DIR` | 否 | `./data` | 应用工作目录；数据库、日志和备份默认写在这里 |
| `PUBLIC_MODE` | 否 | `true` | 启用面向用户展示的公开 dashboard 模式 |
| `QUOTA_REFRESH_INTERVAL` | 否 | `5m` | quota/status 刷新间隔 |
| `REQUEST_TIMEOUT` | 否 | `30s` | 外部请求超时 |
| `LOG_LEVEL` | 否 | `info` | 日志级别 |
| `LOG_FILE_ENABLED` | 否 | `true` | 是否写入持久化日志文件 |
| `LOG_RETENTION_DAYS` | 否 | `7` | 日志保留天数；`0` 表示不自动清理 |
| `BACKUP_ENABLED` | 否 | `true` | 是否启用 SQLite 数据库备份 |
| `BACKUP_INTERVAL` | 否 | `24h` | 数据库备份间隔 |
| `BACKUP_RETENTION_DAYS` | 否 | `7` | 备份保留天数 |

## 安全说明

浏览器 API 不会返回 Sub2API 账号 credentials。账号 email、access token、refresh token、ID token、API key、password、session key 以及 token/secret 类字段会从响应中移除；当前仅暴露允许展示的 credential key，例如 `plan_type` 和 `model_mapping`。

生产部署建议：

- 使用只读 PostgreSQL 用户连接 Sub2API 数据库。

默认部署为无密码公开仪表盘（`AUTH_ENABLED=false`）。浏览器 API 会展示统计、账号状态和排行榜，但不会返回原始 token、密码、邮箱或 credential secret 字段。如果需要访问控制，建议在 Cloudflare Access、Nginx Basic Auth 或其他反代层实现。

- 在反向代理层配置 HTTPS。
- 仅将容器端口绑定到 `127.0.0.1`，不要直接公开到公网。

## 本地开发

### 前置依赖

- Go 1.22+
- Node.js 22+
- npm
- 可访问的 Sub2API PostgreSQL 数据库

### 启动

```bash
cp .env.example .env
go run ./cmd/server/main.go
```

前端开发服务器：

```bash
npm --prefix ./web ci
npm --prefix ./web run dev -- --host 127.0.0.1
```

## 验证

```bash
go test ./cmd/... ./internal/...
npm --prefix ./web run test
npm --prefix ./web run lint
npm --prefix ./web run typecheck
npm --prefix ./web run build
```

也可以运行：

```bash
make verify
```
