# Workspace Maintenance Documentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the broken deployment-notes link and establish accurate maintenance documentation for the Sub2API workspace and current Sub2API Usage Keeper implementation.

**Architecture:** A root `MAINTENANCE.md` is the source of truth for maintainers and AI agents. Concise instruction files point to it, repository README/templates describe only active user-facing capabilities, and dated deployment notes remain historical records behind a compatibility symbolic link.

**Tech Stack:** Markdown, POSIX symbolic links, Git, Go/React project metadata, Docker Compose examples

---

### Task 1: Create The Workspace Maintenance Manual

**Files:**
- Create: `/Users/laijiachen.1/Documents/sub2api/MAINTENANCE.md`
- Modify: `/Users/laijiachen.1/Documents/sub2api/AGENTS.md`
- Modify: `/Users/laijiachen.1/Documents/sub2api/CLAUDE.md`

- [ ] **Step 1: Write the maintenance manual**

Create `MAINTENANCE.md` with these explicit sections and facts:

```markdown
# Sub2API Workspace Maintenance Manual

## Purpose And Reading Order
## Workspace Map
## System Boundaries
## Sub2API Usage Keeper Architecture
## Data Sources And API Flow
## Security And Privacy Invariants
## Frontend Structure
## Local Development And Verification
## Production Deployment Assumptions
## Known Gaps And Legacy Surfaces
## Documentation Map
```

Document the three-repository ownership model, the PostgreSQL-to-normalization-to-Gin-to-Zustand flow, the local SQLite role, real health and quota sources, masking requirements, exact verification commands, loopback deployment rule, missing application login, inactive parsed configuration, and links to design/session/deployment/incident records.

- [ ] **Step 2: Update the instruction-file reading order**

Add `MAINTENANCE.md` to the workspace map in both instruction files and state:

```markdown
- Read `MAINTENANCE.md` before architecture, deployment, security, or cross-repository work.
- `sub2api_dosc/` intentionally keeps its misspelled compatibility name and points to the current personal deployment-note directory.
```

- [ ] **Step 3: Verify synchronized instructions and valid paths**

Run:

```bash
cmp -s AGENTS.md CLAUDE.md
test -f MAINTENANCE.md
rg -n 'MAINTENANCE.md|sub2api_dosc' AGENTS.md CLAUDE.md MAINTENANCE.md
```

Expected: `cmp` exits 0, the maintenance file exists, and all three documents contain the new reading references.

### Task 2: Correct Sub2API Usage Keeper Public Documentation

**Files:**
- Modify: `/Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/README.md`
- Modify: `/Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/README.en.md`
- Modify: `/Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/.env.example`
- Modify: `/Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/docker-compose.example.yml`

- [ ] **Step 1: Correct the architecture and stack descriptions**

Describe the application as a read-only Sub2API PostgreSQL dashboard with local SQLite used for application-local storage and backups. Change the frontend stack from React 18 to React 19 and include Recharts where the stack table enumerates chart libraries.

- [ ] **Step 2: Remove deleted authentication claims**

Remove `AUTH_ENABLED`, `LOGIN_PASSWORD`, and `AUTH_SESSION_TTL` from README configuration tables and deployment examples. Add a security note that the application currently has no built-in login screen and must be bound to loopback or protected by upstream access control.

- [ ] **Step 3: Remove inactive user-facing variables**

Remove `PUBLIC_MODE`, `QUOTA_REFRESH_INTERVAL`, and `REQUEST_TIMEOUT` from `.env.example`, `docker-compose.example.yml`, and README configuration tables because they do not affect the current main runtime path. Do not remove their Go config fields in this documentation-only change.

- [ ] **Step 4: Verify Chinese and English documents agree**

Run:

```bash
rg -n 'React 18|AUTH_ENABLED|LOGIN_PASSWORD|AUTH_SESSION_TTL|PUBLIC_MODE|QUOTA_REFRESH_INTERVAL|REQUEST_TIMEOUT' README.md README.en.md .env.example docker-compose.example.yml
rg -n 'React 19|127\.0\.0\.1|PostgreSQL|SQLite' README.md README.en.md
```

Expected: the first search returns no matches; both READMEs contain the corrected stack, data-source, and protection guidance.

- [ ] **Step 5: Commit repository documentation corrections**

Run:

```bash
git add README.md README.en.md .env.example docker-compose.example.yml
git commit -m "docs: align usage keeper documentation with runtime"
```

Expected: one commit containing only the four documentation/template files.

### Task 3: Refresh The Workspace Session Record

**Files:**
- Modify: `/Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper-session-notes.md`

- [ ] **Step 1: Add a current-state section dated 2026-07-10**

Record the `8810492` legacy-module removal, real `usage_logs` plus `ops_error_logs` health path, ranking trends, local-calendar seven-day health window, silent refresh behavior, current public-dashboard authentication boundary, and the fact that older legacy-health notes are historical.

- [ ] **Step 2: Preserve historical commit notes**

Do not delete the original June notes. Add explicit supersession labels to the old health-empty explanation and the old claim that settings contain API key/pricing cards.

- [ ] **Step 3: Verify the session record contains no active stale claim**

Run:

```bash
rg -n '2026-07-10|8810492|ops_error_logs|无内置登录|历史说明' sub2api-usage-keeper-session-notes.md
```

Expected: every current-state marker is present and dated.

### Task 4: Rebuild And Index Deployment Notes

**Files:**
- Replace symlink: `/Users/laijiachen.1/Documents/sub2api/sub2api_dosc`
- Create: `/Users/laijiachen.1/Documents/Note/个人开发项目/sub2api/README.md`
- Modify: `/Users/laijiachen.1/Documents/Note/个人开发项目/sub2api/todo.md`

- [ ] **Step 1: Repoint the compatibility link**

Run:

```bash
ln -sfn '/Users/laijiachen.1/Documents/Note/个人开发项目/sub2api' sub2api_dosc
```

Expected: `readlink sub2api_dosc` prints the new absolute target.

- [ ] **Step 2: Create the deployment-record index**

Create a Chinese `README.md` that identifies `sub2api.md` as the 2026-05-06 SMTP/Cloudflare record, `sub2api_2026-05-25.md` as the Cloudflare 524 and bind-mount record, and `todo.md` as the dashboard follow-up record. State that dated files are historical evidence, secrets must not be copied into public docs, and live production state requires direct verification.

- [ ] **Step 3: Close the resolved dashboard todo items**

Rewrite `todo.md` as a dated completion record. Mark all five items complete and map them to current behavior: 30-second silent refresh, local-calendar seven-day health window, pricing i18n and readable table text, health independence from range selection, and aligned model-detail tables.

- [ ] **Step 4: Verify the deployment-note link and index**

Run:

```bash
readlink sub2api_dosc
test -f sub2api_dosc/README.md
rg -n '历史|2026-05-06|2026-05-25|已完成' sub2api_dosc/README.md sub2api_dosc/todo.md
```

Expected: the new target is printed, the index is reachable through the link, and historical/completion labels are present.

### Task 5: Final Documentation Verification

**Files:**
- Verify all files changed in Tasks 1-4.

- [ ] **Step 1: Check formatting and repository status**

Run:

```bash
git -C sub2api-usage-keeper diff --check
git -C sub2api-usage-keeper status --short --branch
cmp -s AGENTS.md CLAUDE.md
```

Expected: no whitespace errors, synchronized instruction files, and only intentional repository state remains.

- [ ] **Step 2: Scan for stale capability claims**

Run:

```bash
rg -n 'React 18|AUTH_ENABLED|LOGIN_PASSWORD|AUTH_SESSION_TTL' \
  sub2api-usage-keeper/README.md \
  sub2api-usage-keeper/README.en.md \
  sub2api-usage-keeper/.env.example \
  sub2api-usage-keeper/docker-compose.example.yml
```

Expected: no matches.

- [ ] **Step 3: Verify the complete documentation map**

Run:

```bash
test -f MAINTENANCE.md
test -f design.md
test -f sub2api-usage-keeper-session-notes.md
test -f redis-security-incident-technical.md
test -f sub2api_dosc/README.md
```

Expected: every command exits 0.

- [ ] **Step 4: Review diffs for secrets and unintended history rewrites**

Inspect repository and workspace diffs. Confirm that no new credential value, token, database URL, SMTP password, private key, or production secret was introduced and that the dated SMTP and 524 records were not rewritten.
