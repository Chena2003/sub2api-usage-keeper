# Sub2API Usage Keeper

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19+-61DAFB?logo=react&logoColor=black)](https://react.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](./Dockerfile)

[English](./README.en.md) | 中文

Sub2API Usage Keeper 是一个独立的 sidecar 仪表盘服务，用于 [Sub2API](https://github.com/xLmile/sub2api) 的用量统计、账号状态监控和排行分析。

本项目参考了 [CPA Usage Keeper](https://github.com/willxup/cpa-usage-keeper) 的架构设计和实现方式，针对 Sub2API 的数据模型和部署场景进行了适配开发。

服务端通过只读连接直接查询 Sub2API PostgreSQL，不消费 Redis usage 队列，也不把 Sub2API 请求事件复制到本地 SQLite。SQLite 仅用于应用自身仍保留的本地数据、清理和备份。

## 功能特性

- 📊 **用量总览** — 请求量、Token 消耗、成本统计，支持日/时维度切换
- 📈 **趋势图表** — 模型/用户/API Key/账号 的 Top 12 使用趋势折线图
- 🏆 **排行榜** — 按用户、API Key、模型、账号四个维度统计 Token 排行
- 🖥️ **账号状态** — 上游账号状态（Active/Paused/Error/Rate Limited 等）、计划类型、reset 倒计时
- 📋 **账号额度** — 三段式布局展示身份/指标/双窗口额度条（5h + 7d），读取真实 utilization 数据
- 🩺 **健康监控** — 请求成功率时间线、15 分钟粒度健康块
- 📝 **事件日志** — 逐条请求明细、分页、筛选
- 🔒 **安全脱敏** — 浏览器 API 不返回原始 credentials；email/用户标识自动脱敏
- 🌙 **暗色主题** — 支持明/暗主题自动切换
- 🌐 **多语言** — 中文 / English / 繁體中文
- 🐳 **容器化** — Docker 一键部署，多阶段构建

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.22、Gin、GORM、SQLite |
| 前端 | React 19、TypeScript、Vite、Chart.js、Recharts、Zustand |
| 数据源 | Sub2API PostgreSQL（只读连接） |
| 部署 | Docker、Docker Compose |

## 快速开始

### Docker 部署（推荐）

确保 dashboard 容器能访问 Sub2API 的 PostgreSQL 数据库（同一 Docker network）。

```bash
# 1. 准备配置
cp .env.example .env
vim .env  # 填写 SUB2API_DATABASE_URL

# 2. 构建并启动
docker build -t sub2api-usage-keeper:latest .
docker compose -f docker-compose.example.yml up -d
```

### 接入已有 Sub2API

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

> 当前版本没有内置登录，任何能访问到它的人都能看到看板内容。这是有意的：本服务面向"公开只读看板"场景。若你想公开，直接经入口（如 Cloudflare Tunnel / 反向代理）对外即可；若想私有，则在入口层加访问控制或仅绑定内网。无论哪种，务必配合下方「安全说明」里的只读数据库用户。

## 配置

| 变量 | 必填 | 默认值 | 说明 |
|------|:----:|--------|------|
| `SUB2API_DATABASE_URL` | ✅ | — | Sub2API PostgreSQL 连接串 |
| `APP_PORT` | | `8080` | HTTP 监听端口 |
| `APP_BASE_PATH` | | `/` | 子路径前缀，如 `/usage` |
| `TZ` | | `Asia/Shanghai` | 时区 |
| `WORK_DIR` | | `./data` | 数据目录（SQLite、日志、备份） |
| `LOG_LEVEL` | | `info` | 日志级别 |
| `LOG_FILE_ENABLED` | | `true` | 写入持久化日志文件 |
| `LOG_RETENTION_DAYS` | | `7` | 日志保留天数 |
| `BACKUP_ENABLED` | | `true` | 启用 SQLite 备份 |
| `BACKUP_INTERVAL` | | `24h` | 备份间隔 |
| `BACKUP_RETENTION_DAYS` | | `7` | 备份保留天数 |

## 本地开发

### 前置依赖

- Go 1.22+
- Node.js 22+
- npm
- 可访问的 Sub2API PostgreSQL 数据库

### 启动后端

```bash
cp .env.example .env
# 修改 SUB2API_DATABASE_URL 指向你的数据库
go run ./cmd/server/main.go
```

### 启动前端开发服务器

```bash
cd web
npm ci
npm run dev -- --host 127.0.0.1
```

访问 http://127.0.0.1:5173

### 验证

```bash
# 一键验证
make verify

# 或分步执行
go test ./cmd/... ./internal/...
npm --prefix ./web test -- --run
npm --prefix ./web run lint
npm --prefix ./web run typecheck
npm --prefix ./web run build
```

## 项目结构

```
sub2api-usage-keeper/
├── cmd/server/          # 程序入口
├── internal/
│   ├── api/             # HTTP 路由与处理器
│   ├── app/             # 应用启动与依赖注入
│   ├── quota/           # 额度计算与类型定义
│   ├── service/         # 业务逻辑层
│   └── sub2api/         # Sub2API 数据库访问与实体
├── web/
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── lib/         # API 调用与工具函数
│   │   ├── pages/       # 页面组件
│   │   └── stores/      # Zustand 状态管理
│   └── ...
├── Dockerfile           # 多阶段构建
├── docker-compose.example.yml
└── .env.example
```

## 安全说明

本服务设计为**可公开访问的只读看板**：没有内置登录，任何能访问到它的人都能看到页面内容。因此安全模型不是"挡住访问者"，而是"确保对外暴露的只有脱敏后的聚合数据"。

浏览器 API 的输出边界（无论公开与否都成立）：

- **不返回**原始 credentials（email、access token、refresh token、password、session key 等）
- 用户邮箱自动脱敏：本地部分 > 5 字符保留首 3 尾 2；短本地部分仅保留首字符；域名可见
- 非邮箱用户标识进行哈希脱敏
- 账号原始错误消息、临时不可调度原因等文本不外泄

部署时你需要自己权衡的一点：看板会公开展示**聚合成本、token 用量、账号数量、模型分布**等运营数据。个人项目通常无所谓；若你不希望这些数字对外，可在前端相应面板中不渲染成本列，或在入口层（如 Cloudflare Access）加访问控制。**不需要**为了"加锁"而套 Access——那是给本该私有的后台补登录用的，与公开看板的定位冲突。

### 使用只读数据库用户（公开部署下为必做项）

应用代码本身只执行 `SELECT`，但这只是代码层面的自我约束。**当看板公开可达时，任何人都能触发它的数据库查询**，因此务必在数据库层面强制只读，这样最坏情况也只是被读，不可能写入或删除 Sub2API 生产数据：

```sql
CREATE USER sub2api_readonly WITH PASSWORD 'replace-with-strong-password';
GRANT CONNECT ON DATABASE sub2api TO sub2api_readonly;
GRANT USAGE ON SCHEMA public TO sub2api_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO sub2api_readonly;
-- 让后续新建的表也自动只读
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO sub2api_readonly;
-- 双保险：即使某处漏授权，该用户的事务也强制只读
ALTER USER sub2api_readonly SET default_transaction_read_only = on;
```

然后把 `SUB2API_DATABASE_URL` 指向该用户。

## 致谢

- [CPA Usage Keeper](https://github.com/willxup/cpa-usage-keeper) — 本项目的参考实现，架构设计和前端框架均受其启发
- [Sub2API](https://github.com/xLmile/sub2api) — 上游 AI API 网关

## License

[MIT](./LICENSE)
