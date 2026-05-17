# Sub2API Usage Keeper

[中文说明](./README.md)

Sub2API Usage Keeper is a standalone sidecar dashboard for Sub2API account status, quota-safe account metadata, and global usage statistics.

## Features

- Reads account and aggregate usage data from an existing Sub2API PostgreSQL database
- Shows total accounts, active accounts, request volume, token usage, cost, and model ranking
- Shows upstream account status, plan, reset time, and recent 7-day usage
- Never returns raw Sub2API credentials to browser APIs
- Uses local SQLite for dashboard-owned data, logs, and backups
- Optional password login protection
- Docker / Docker Compose deployment

## Docker with existing Sub2API

1. Ensure the dashboard container can reach `sub2api-postgres` on the same Docker network.
2. Set `SUB2API_DATABASE_URL` to a read-only Postgres user when possible.
3. Start the dashboard on `127.0.0.1:8082` and expose it through Caddy, Nginx, or Cloudflare.

```bash
cp .env.example .env
vim .env
docker build -t sub2api-usage-keeper:latest .
docker compose -f docker-compose.example.yml up -d
```

## Configuration

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `SUB2API_DATABASE_URL` | Yes | - | Sub2API PostgreSQL connection string, for example `postgres://sub2api:password@sub2api-postgres:5432/sub2api?sslmode=disable` |
| `AUTH_ENABLED` | No | `false` | Enable login protection |
| `LOGIN_PASSWORD` | When auth is enabled | - | Dashboard login password |
| `AUTH_SESSION_TTL` | No | `168h` | Session lifetime |
| `APP_PORT` | No | `8080` | HTTP listen port |
| `APP_BASE_PATH` | No | root path | Subpath prefix such as `/usage`; empty means `/` |
| `TZ` | No | `Asia/Shanghai` | Project timezone |
| `WORK_DIR` | No | `./data` | Application work directory for database, logs, and backups |
| `PUBLIC_MODE` | No | `true` | Enable user-facing public dashboard mode |
| `QUOTA_REFRESH_INTERVAL` | No | `5m` | Quota/status refresh interval |
| `REQUEST_TIMEOUT` | No | `30s` | External request timeout |
| `LOG_LEVEL` | No | `info` | Log level |
| `LOG_FILE_ENABLED` | No | `true` | Write persistent log files |
| `LOG_RETENTION_DAYS` | No | `7` | Log retention days; `0` disables cleanup |
| `BACKUP_ENABLED` | No | `true` | Enable SQLite database backups |
| `BACKUP_INTERVAL` | No | `24h` | Database backup interval |
| `BACKUP_RETENTION_DAYS` | No | `7` | Backup retention days |

## Security notes

The browser API never returns Sub2API account credentials. Account emails, access tokens, refresh tokens, ID tokens, API keys, passwords, session keys, and token/secret-like fields are removed from responses. Only explicitly allowed display-safe credential keys, such as `plan_type` and `model_mapping`, are exposed.

Production recommendations:

- Use a read-only PostgreSQL user for the Sub2API database connection.
- Enable `AUTH_ENABLED=true`.
- Terminate HTTPS at your reverse proxy.
- Bind the container port to `127.0.0.1` instead of exposing it directly to the internet.

## Development

### Prerequisites

- Go 1.22+
- Node.js 22+
- npm
- A reachable Sub2API PostgreSQL database

### Run locally

```bash
cp .env.example .env
go run ./cmd/server/main.go
```

Frontend dev server:

```bash
npm --prefix ./web ci
npm --prefix ./web run dev -- --host 127.0.0.1
```

## Verification

```bash
go test ./cmd/... ./internal/...
npm --prefix ./web run test
npm --prefix ./web run lint
npm --prefix ./web run typecheck
npm --prefix ./web run build
```

Or run:

```bash
make verify
```
