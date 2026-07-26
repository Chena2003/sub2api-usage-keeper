# Sub2API Usage Keeper

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19+-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

English | [中文](./README.md)

Sub2API Usage Keeper is a standalone sidecar dashboard for [Sub2API](https://github.com/xLmile/sub2api) usage statistics, account status monitoring, and ranking analysis.

This project is inspired by [CPA Usage Keeper](https://github.com/willxup/cpa-usage-keeper) and adapts its architecture and design patterns for the Sub2API data model and deployment scenarios.

The server queries Sub2API PostgreSQL directly through a read-only connection. It does not consume a Redis usage queue or copy Sub2API request events into local SQLite. SQLite is limited to the application's remaining local data, cleanup, and backups.

## Features

- 📊 **Usage Overview** — Request volume, token consumption, cost statistics with daily/hourly granularity
- 📈 **Trend Charts** — Top 12 usage trend line charts by model/user/API Key/account
- 🏆 **Rankings** — Token ranking across four dimensions: user, API Key, model, account
- 🖥️ **Account Status** — Upstream account states (Active/Paused/Error/Rate Limited, etc.), plan type, reset countdown
- 📋 **Account Quotas** — Three-segment layout showing identity/metrics/dual quota bars (5h + 7d) with real utilization data
- 🩺 **Health Monitoring** — Request success rate timeline, 15-minute granularity health blocks
- 📝 **Event Logs** — Per-request details with pagination and filtering
- 🔒 **Data Masking** — Browser APIs never expose raw credentials; emails and user identifiers are automatically masked
- 🌙 **Dark Mode** — Automatic light/dark theme switching
- 🌐 **Multi-language** — Chinese / English / Traditional Chinese
- 🐳 **Containerized** — One-command Docker deployment with multi-stage builds

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22, Gin, GORM, SQLite |
| Frontend | React 19, TypeScript, Vite, Chart.js, Recharts, Zustand |
| Data Source | Sub2API PostgreSQL (read-only connection) |
| Deployment | Docker, Docker Compose |

## Quick Start

### Docker Deployment (Recommended)

Ensure the dashboard container can reach the Sub2API PostgreSQL database (same Docker network).

```bash
# 1. Configure
cp .env.example .env
vim .env  # Set SUB2API_DATABASE_URL

# 2. Build and start
docker build -t sub2api-usage-keeper:latest .
docker compose -f docker-compose.example.yml up -d
```

### Attach to Existing Sub2API

```bash
docker run -d \
  --name sub2api-usage-keeper \
  --restart unless-stopped \
  --network <your-sub2api-network> \
  -p 127.0.0.1:8082:8080 \
  --env-file .env \
  -v ./data:/data \
  sub2api-usage-keeper:latest
```

> The current version has no built-in login protection. Bind it to `127.0.0.1` or a protected private network, and use upstream infrastructure such as Caddy, Nginx, or Cloudflare Access for HTTPS and access control.

## Configuration

| Variable | Required | Default | Description |
|----------|:--------:|---------|-------------|
| `SUB2API_DATABASE_URL` | ✅ | — | Sub2API PostgreSQL connection string |
| `APP_PORT` | | `8080` | HTTP listen port |
| `APP_BASE_PATH` | | `/` | Subpath prefix, e.g. `/usage` |
| `TZ` | | `Asia/Shanghai` | Timezone |
| `WORK_DIR` | | `./data` | Data directory (SQLite, logs, backups) |
| `LOG_LEVEL` | | `info` | Log level |
| `LOG_FILE_ENABLED` | | `true` | Write persistent log files |
| `LOG_RETENTION_DAYS` | | `7` | Log retention days |
| `BACKUP_ENABLED` | | `true` | Enable SQLite backups |
| `BACKUP_INTERVAL` | | `24h` | Backup interval |
| `BACKUP_RETENTION_DAYS` | | `7` | Backup retention days |

## Development

### Prerequisites

- Go 1.22+
- Node.js 22+
- npm
- A reachable Sub2API PostgreSQL database

### Run Backend

```bash
cp .env.example .env
# Edit SUB2API_DATABASE_URL to point at your database
go run ./cmd/server/main.go
```

### Run Frontend Dev Server

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

Visit http://127.0.0.1:5173

### Verification

```bash
# All-in-one
make verify

# Or run individually
go test ./cmd/... ./internal/...
npm --prefix ./web test -- --run
npm --prefix ./web run lint
npm --prefix ./web run typecheck
npm --prefix ./web run build
```

## Project Structure

```
sub2api-usage-keeper/
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── api/             # HTTP routes and handlers
│   ├── app/             # Application bootstrap and DI
│   ├── quota/           # Quota calculation and type definitions
│   ├── service/         # Business logic layer
│   └── sub2api/         # Sub2API database access and entities
├── web/
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── lib/         # API calls and utilities
│   │   ├── pages/       # Page components
│   │   └── stores/      # Zustand state management
│   └── ...
├── Dockerfile           # Multi-stage build
├── docker-compose.example.yml
└── .env.example
```

## Security

- Browser APIs **never return** raw credentials (emails, access tokens, refresh tokens, passwords, session keys, etc.)
- User emails are automatically masked: local part > 5 chars retains first 3 and last 2; short local parts keep only the first character; domain remains visible
- Non-email user identifiers are hash-masked
- There is no built-in login page or password session; do not expose the service directly to the public Internet
- Bind container ports to `127.0.0.1`, terminate HTTPS at the reverse proxy, and enforce access control there

### Use a read-only database user

The application only issues `SELECT` statements, but that is a code-level
self-restriction. Enforce read-only access at the database layer so that neither a
leaked connection string nor a future code change can write to production data:

```sql
CREATE USER sub2api_readonly WITH PASSWORD 'replace-with-strong-password';
GRANT CONNECT ON DATABASE sub2api TO sub2api_readonly;
GRANT USAGE ON SCHEMA public TO sub2api_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO sub2api_readonly;
-- Make future tables read-only as well
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO sub2api_readonly;
```

Then point `SUB2API_DATABASE_URL` at that user.

## Acknowledgments

- [CPA Usage Keeper](https://github.com/willxup/cpa-usage-keeper) — The reference implementation that inspired this project's architecture and frontend design
- [Sub2API](https://github.com/xLmile/sub2api) — The upstream AI API gateway

## License

[MIT](./LICENSE)
