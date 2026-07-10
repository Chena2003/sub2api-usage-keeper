# Workspace Maintenance Documentation Design

## Purpose

Create a reliable maintenance entry point for the Sub2API workspace. The primary readers are the repository owner and AI coding agents that need to understand the workspace before changing code or deployment assumptions.

The documentation must describe current implementation facts, clearly separate historical deployment records from active guidance, and avoid introducing secrets or private production values.

## Scope

The workspace contains three related repositories and several root-level records:

- `sub2api/` is the upstream AI API gateway and the source of PostgreSQL usage and account data.
- `sub2api-usage-keeper/` is the actively maintained read-only Sub2API dashboard sidecar.
- `cpa-usage-keeper/` is the reference implementation and must not be treated as the active product.
- `sub2api_dosc` is a compatibility link used by existing instructions for deployment notes.
- Root-level design, session, and security-incident documents provide supporting context.

The change covers documentation and the deployment-notes symbolic link only. It does not change application behavior, database schemas, production services, or source code.

## Documentation Architecture

The root `MAINTENANCE.md` becomes the authoritative workspace overview for maintainers and AI agents. It will contain:

1. Workspace map and repository ownership.
2. Current Sub2API Usage Keeper architecture and startup flow.
3. PostgreSQL, SQLite, API, normalization, and frontend data flow.
4. Security and masking invariants.
5. Development and verification commands.
6. Production deployment assumptions and operational cautions.
7. Known implementation/documentation gaps and inherited legacy surfaces.
8. A reading map pointing to design, session, deployment, and incident records.

`AGENTS.md` and `CLAUDE.md` remain synchronized concise instruction files. They will point readers to `MAINTENANCE.md` and explain that `sub2api_dosc` keeps its misspelled name for compatibility.

The deployment-notes directory will gain a `README.md` index. Historical SMTP and Cloudflare 524 records remain unchanged except where an explicit status note is needed; their dates and observations are evidence from the time they were written, not current configuration guarantees.

## Symbolic Link

Keep the workspace link name:

```text
sub2api_dosc
```

Replace its obsolete target with:

```text
/Users/laijiachen.1/Documents/Note/个人开发项目/sub2api
```

Preserving the link name avoids breaking references in existing instructions and session records. The maintenance manual will call out the compatibility spelling so future agents do not rename it casually.

## Files To Update

### Workspace Root

- Create `MAINTENANCE.md` as the complete maintenance manual.
- Update `AGENTS.md` and `CLAUDE.md` with the new reading order, corrected deployment-note location, and current project boundaries.
- Update `sub2api-usage-keeper-session-notes.md` with the post-June-28 refactor and the current health, ranking, authentication, and deployment state.
- Repoint `sub2api_dosc` to the new deployment-notes directory.

### Sub2API Usage Keeper Repository

- Update `README.md` and `README.en.md` to match the current React 19 stack and direct PostgreSQL read architecture.
- Remove claims that password login protection exists in the current application.
- Describe the current public-dashboard deployment model and the requirement for loopback binding plus reverse-proxy access control.
- Remove deleted authentication variables from `.env.example` and `docker-compose.example.yml`.
- Remove inactive variables from the user-facing templates and README configuration tables. Record in `MAINTENANCE.md` that the Go configuration layer still parses them even though the main runtime path does not use them.

### Deployment Notes

- Create `/Users/laijiachen.1/Documents/Note/个人开发项目/sub2api/README.md` as an index of deployment records, their dates, and whether they are historical or current guidance.
- Update `todo.md` to mark the five dashboard items as completed by the current branch and reference the relevant current behavior.
- Do not rewrite the SMTP or Cloudflare 524 incident narratives as if they were current live-state inventories.

## Current-Fact Baseline

The documentation must reflect these verified implementation facts:

- The Go process opens local SQLite and the Sub2API PostgreSQL database, then serves Gin APIs and embedded React assets.
- The primary dashboard path reads `accounts`, `accounts.extra`, `usage_logs`, `ops_error_logs`, `usage_dashboard_daily`, and `usage_dashboard_hourly` from Sub2API PostgreSQL.
- SQLite remains open for local application data, cleanup, and backup; inherited generic tables and APIs still exist, but the removed CPA poller no longer ingests Sub2API usage into SQLite.
- Browser-facing account, event, and ranking responses are normalized before serialization.
- Email local parts are masked, non-email user identifiers and API key identifiers are hash-masked, raw credential values are withheld, and raw error or temporary-unschedulable messages are not exposed.
- Quota utilization comes from real values persisted in `accounts.extra`; missing values produce unknown or empty states rather than fabricated limits.
- Total Tokens is the sum of input, output, cache-creation, and cache-read tokens.
- Request health is derived from real successful usage events and operational error logs in 15-minute blocks over a seven-day window aligned to local calendar days.
- The active frontend uses React 19, TypeScript, Vite, Zustand, Chart.js, and Recharts.
- The password-session authentication module was removed. Current production protection depends on binding the service to loopback and enforcing access control at the reverse proxy or surrounding infrastructure.

## Security Rules

- Do not add secret values from `.env`, PostgreSQL URLs, SMTP credentials, Cloudflare tokens, API keys, OAuth tokens, or password files to repository documentation.
- Do not reproduce private values merely because they already exist in a personal historical note.
- Public repository documentation may name environment-variable keys and server-side file paths when operationally necessary, but must use placeholders for values.
- Deployment examples must bind the dashboard to `127.0.0.1` unless the text explicitly describes an alternative protected network boundary.
- Database and cache ports must not be documented as public `0.0.0.0` bindings.

## Validation

The documentation update is complete when all of the following checks pass:

1. `readlink sub2api_dosc` returns the new deployment-notes directory.
2. The link target can be listed and its new `README.md` is readable through the link.
3. `AGENTS.md` and `CLAUDE.md` remain byte-for-byte identical.
4. Repository searches no longer claim that the removed login module is active.
5. React version and data-source descriptions match the current package and Go wiring.
6. Every local path referenced by `MAINTENANCE.md` exists.
7. No new secret values appear in the workspace diff.
8. `git diff --check` passes in `sub2api-usage-keeper`.

Application tests and a development server are not required because this change does not modify executable code. Configuration examples receive static consistency checks against `internal/config/config.go` and the current application wiring.

## Non-Goals

- Reintroducing application-level authentication.
- Removing inherited SQLite entities, migrations, or generic APIs.
- Changing the dashboard design or implementing unresolved UI work.
- Editing the upstream `sub2api` or reference `cpa-usage-keeper` source code.
- Treating dated production notes as a live inventory without direct server verification.
