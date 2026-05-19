# Sub2API Usage Keeper Redesign Design

## Summary

Redesign the current monitoring site as **Sub2API Usage Keeper**: a no-password, public-readable dashboard that preserves CPA Usage Keeper's tabbed analysis experience while changing the statistics source to Sub2API request data. The redesign adds account quota windows and token consumption rankings without removing the original usage-analysis workflows.

## Goals

- Remove the application password login flow.
- Preserve CPA Usage Keeper's multi-tab dashboard structure and Chinese/English language switching.
- Replace CPA-oriented statistics with Sub2API request, model, account, API key, and user statistics.
- Add account quota visibility for 5-hour and weekly windows, including consumption and refresh time.
- Add token consumption rankings, defaulting to user ranking.
- Redesign the frontend with a clean light blue/white visual style aligned with Sub2API.
- Keep browser APIs from exposing raw credentials, tokens, passwords, or email fields.

## Non-goals

- Do not modify the core Sub2API service.
- Do not expose raw Sub2API credential fields in browser APIs.
- Do not keep the current standalone dark Sub2API dashboard as the primary product experience.
- Do not split CPA statistics and Sub2API statistics into two disconnected apps.

## Product structure

The product keeps a tabbed analytics shell similar to CPA Usage Keeper, but every statistical view is backed by Sub2API data.

Tabs:

1. **Overview / 总览**
   - Whole-site Sub2API request count, token volume, cost, active users, error/success rate, and time-series trends.
   - Includes model/user summaries and an account quota risk summary that links to Account Quotas.
2. **API & Models / API 与模型**
   - Model usage, API key usage, source distribution, pricing, latency, and related analysis using Sub2API data.
3. **Request Events / 请求事件**
   - Sub2API request events with time, user/API key, model, status, token, cost, duration, and backend account fields.
   - Supports filtering and pagination.
4. **Token Ranking / Token 排行榜**
   - Defaults to user token consumption ranking.
   - Supports switching between user, API key, model, and backend account dimensions.
5. **Account Quotas / 账号额度**
   - Shows backend account quota state, especially 5-hour and weekly windows.
   - Shows consumption, limit, remaining quota, refresh time, schedulable state, and risk labels.
6. **Health / Settings / 健康 / 设置**
   - Shows service health, sync state, refresh state, and local maintenance information.
   - Does not include an application password login entry.

The default landing tab is Overview. Account quota information appears as a summary on Overview and as a full dedicated tab.

## Visual design

Use the user-selected **clean light blue/white** direction:

- White and light gray surfaces.
- Blue primary color for active tabs, highlights, progress, and key actions.
- Subtle borders and shadows instead of heavy dark panels.
- Clear data-card hierarchy: primary metrics first, charts second, diagnostic tables below.
- Design should feel like a Sub2API admin/user dashboard extension, not a separate dark monitoring tool.

The top bar includes:

- Product name: Sub2API Usage Keeper.
- Language switcher preserving CPA Usage Keeper's Chinese/English behavior.
- Refresh or sync state.
- Time range selector where relevant.

## Backend data design

Use Sub2API PostgreSQL as the authoritative statistics source and keep local SQLite for the app's own runtime state, backup, logs, and optional snapshots.

Primary Sub2API sources:

- `accounts`
  - Account identity, provider/type, display name, status, schedulable state, session/window fields, rate limit/reset fields, and display-safe metadata.
- `usage_logs`
  - Request event details for Request Events, rankings, user/API key/model/account aggregation, status, duration, token, and cost analysis.
- `usage_dashboard_daily`
  - Daily overview and trend aggregation.
- `usage_dashboard_hourly`
  - Hourly overview and trend aggregation.

Local SQLite can store:

- Runtime and maintenance state.
- Backup state.
- Optional quota snapshots.
- Sync status.

New local SQLite tables must not store raw Sub2API credentials. Browser-facing APIs must never expose raw credentials from any source.

## API design

Keep the frontend API shape close to CPA Usage Keeper where practical so existing components can be adapted instead of replaced. New and adapted endpoints should be Sub2API-focused.

Expected endpoint groups:

- `GET /api/v1/sub2api/overview`
- `GET /api/v1/sub2api/models`
- `GET /api/v1/sub2api/events`
- `GET /api/v1/sub2api/rankings?dimension=user|api_key|model|account`
- `GET /api/v1/sub2api/account-quotas`
- Existing health and maintenance endpoints where still applicable.

Ranking behavior:

- Default dimension is `user`.
- Supported dimensions are `user`, `api_key`, `model`, and `account`.
- Each row includes request count, input tokens, output tokens, cache tokens, total tokens, cost, and share of total.

Account quota response behavior:

- Include account display name, provider, type, status, schedulable state, and last-used state.
- Include 5-hour window fields:
  - consumed amount.
  - limit if available.
  - remaining amount if available.
  - consumption ratio if available.
  - window start/end if available.
  - next refresh time if available.
- Include weekly window fields with the same shape.
- Include risk state: normal, near limit, exhausted, not schedulable, rate limited, unknown, or not configured.
- If Sub2API does not expose a reliable field for a quota limit or refresh time, return `unknown` or `not configured` rather than fabricating a value.

## Authentication and public access

The application should no longer require password login.

- Default or deployment configuration sets `AUTH_ENABLED=false`.
- The frontend no longer routes users through LoginPage before showing the dashboard.
- Dashboard APIs are public-readable by default.
- Health checks remain public.
- The service should still be deployable behind a reverse proxy or Cloudflare Access if external access control is desired later.

## Security boundary

The user chose a full public display for dashboard data. This expands visibility for statistics and display identities, but it does not permit raw secret exposure.

Browser APIs must never return raw values for:

- `access_token`
- `refresh_token`
- `id_token`
- `api_key`
- `password`
- `session_key`
- email fields
- token-like fields
- secret-like fields

Allowed display data includes:

- Non-secret user or display identifiers when available.
- Redacted or display-safe API key labels.
- Account display name, provider, type, status, quota windows, and usage statistics.
- Model names, request counts, token counts, cost, latency, trends, and ranking aggregates.

Credential key names should not be exposed arbitrarily. Display-safe credential keys remain allowlisted, currently `plan_type` and `model_mapping`, unless the implementation finds additional non-sensitive display fields and explicitly adds them.

## Frontend implementation design

Reuse and adapt the existing CPA Usage Keeper frontend instead of replacing it with a single Sub2API page.

Frontend changes:

- Restore the multi-tab UsagePage-style experience as the main app shell.
- Merge the current Sub2ApiDashboardPage functionality into the tabbed dashboard.
- Add Account Quotas components:
  - summary cards.
  - quota risk labels.
  - 5-hour and weekly progress bars.
  - refresh time display.
  - account table.
- Add Token Ranking components:
  - dimension switcher.
  - ranking table.
  - token/cost/share columns.
- Adapt Request Events to Sub2API `usage_logs` data.
- Adapt API & Models to Sub2API model/API key/source/latency/cost data.
- Keep the language switcher and add translations for every new tab, table header, state label, empty state, and error message.
- Remove or bypass the LoginPage route from the default app flow. Auth code can remain internally if removal would cause unnecessary churn.

## Backend implementation design

Backend changes:

- Keep the app/service/repository layering.
- Keep Sub2API PostgreSQL repository as the source for dashboard data.
- Add repository methods for events, ranking aggregations, and quota-window data.
- Keep existing query parameter normalization patterns for days, hours, limits, pagination, and dimensions.
- Keep local app maintenance and backup behavior where still relevant.
- Avoid restoring CPA network-sync behavior if it no longer serves Sub2API statistics.
- Keep sensitive-field filtering centralized in normalization code.

## Internationalization

Preserve CPA Usage Keeper's Chinese/English switching.

New translation coverage includes:

- Account Quotas / 账号额度
- Token Ranking / Token 排行榜
- 5-hour window / 5 小时窗口
- Weekly window / 周限窗口
- Refresh time / 刷新时间
- Consumed / 已消耗
- Remaining / 剩余
- Near limit / 接近限制
- Exhausted / 已达限制
- Not schedulable / 不可调度
- Rate limited / 已限流
- Unknown / 未知
- Not configured / 未配置

## Verification plan

Backend verification:

- Config defaults or deployment config disable auth.
- Sensitive credential fields are not serialized in browser API responses.
- Account quota normalization handles known, unknown, and not-configured window data.
- Ranking endpoint defaults to `user` and validates supported dimensions.
- Events endpoint handles filtering and pagination boundaries.
- Overview/model endpoints use Sub2API data.

Frontend verification:

- Opening the site shows the dashboard without password login.
- Overview renders Sub2API summary and trends.
- Tab switching works.
- Language switching works for existing and new strings.
- Token Ranking defaults to user ranking and switches dimensions.
- Account Quotas renders 5-hour and weekly windows, refresh time, status, and empty states.
- Request Events renders Sub2API events with filters and pagination.

Build and test verification:

- Run Go tests.
- Run frontend lint.
- Run frontend typecheck.
- Run frontend tests.
- Run frontend build.
- Start the app locally and manually verify the main dashboard flows in a browser.

## Acceptance criteria

- The website no longer asks for a password before showing the dashboard.
- The visual style is clean light blue/white and feels aligned with Sub2API.
- CPA Usage Keeper's core tabbed analytics experience is preserved.
- Statistics describe Sub2API requests rather than CPA project requests.
- Account Quotas shows 5-hour and weekly consumption and refresh information when available.
- Token Ranking defaults to user token consumption and supports other ranking dimensions.
- Chinese/English switching remains available.
- Browser APIs do not expose raw credentials, tokens, passwords, or email fields.
