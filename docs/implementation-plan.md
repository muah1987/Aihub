# Implementation Plan

This document tracks the phased implementation of Aihub — from initial MVP through the full AI-native web platform.

---

## Phase 1: Core Foundation (MVP)

**Status: Complete**

The minimum viable product establishing authentication, project management, real-time chat, AI agent invocation, and Docker terminal sandbox.

### Backend

| Component | Description | Files |
|-----------|-------------|-------|
| Auth Service | Email/password registration, JWT access + refresh tokens, login/logout | `internal/auth/service.go`, `jwt.go`, `middleware.go` |
| Provider Service | Encrypted storage of API keys (OpenAI, Anthropic, GitHub), validation endpoint | `internal/provider/service.go`, `github.go` |
| Project Service | CRUD for projects mapped to GitHub repos, repo listing via GitHub API | `internal/project/service.go` |
| Chat Service | Message CRUD for project timeline, WebSocket hub for real-time broadcast | `internal/chat/service.go`, `hub.go` |
| Agent Service | Agent profiles with model/temperature/system prompt config, AI invocation via provider API | `internal/agent/service.go`, `provider.go` |
| Terminal Service | Docker container lifecycle management, WebSocket shell sessions | `internal/terminal/service.go`, `container.go` |
| Config | Environment-based configuration loading | `internal/config/config.go` |
| Database | PostgreSQL connection, GORM setup, migration runner | `internal/database/` |
| Middleware | CORS, request logging, rate limiting (10 req/s, 50 burst) | `internal/middleware/` |
| Router | Chi router with all API routes | `internal/router/router.go` |

### Database (Migration 000001)

| Table | Purpose |
|-------|---------|
| `users` | User accounts with email, password hash, display name |
| `provider_connections` | API keys for AI providers (encrypted via AES-256-GCM) |
| `projects` | GitHub repo mappings with metadata and settings |
| `messages` | Project chat timeline (user, agent, system messages) |
| `agents` | AI agent profiles with model config and token tracking |
| `terminal_sessions` | Docker sandbox container sessions |

### Frontend

| Component | Description |
|-----------|-------------|
| `LoginForm` / `RegisterForm` | Auth UI with form validation |
| `ProtectedRoute` | Route guard that redirects unauthenticated users |
| `ProjectList` / `ProjectCard` / `CreateProject` | Project management dashboard |
| `ProjectDashboard` | Tabbed project view (main application shell) |
| `ChatTimeline` / `ChatInput` / `MessageBubble` | Real-time project chat |
| `AgentPanel` | Agent CRUD, model selection, invoke modal |
| `Terminal` | xterm.js-based terminal with WebSocket connection |
| `AppShell` / `Header` / `Sidebar` / `TabBar` | Responsive layout |
| `Button` / `Card` / `Input` / `Modal` / `Toast` | Reusable UI primitives |
| `authStore` | Zustand store for auth tokens and user profile |
| API modules | `auth.ts`, `agents.ts`, `chat.ts`, `projects.ts`, `client.ts` |

### API Endpoints

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
GET    /api/v1/auth/me
POST   /api/v1/auth/logout
GET    /api/v1/providers
POST   /api/v1/providers
DELETE /api/v1/providers/{id}
POST   /api/v1/providers/{id}/validate
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/{id}
PUT    /api/v1/projects/{id}
DELETE /api/v1/projects/{id}
GET    /api/v1/projects/{id}/repos
GET    /api/v1/projects/{id}/messages
POST   /api/v1/projects/{id}/messages
GET    /api/v1/projects/{id}/agents
POST   /api/v1/projects/{id}/agents
PUT    /api/v1/projects/{id}/agents/{agentId}
DELETE /api/v1/projects/{id}/agents/{agentId}
POST   /api/v1/projects/{id}/agents/{agentId}/invoke
GET    /api/v1/projects/{id}/terminal/sessions
POST   /api/v1/projects/{id}/terminal/sessions
DELETE /api/v1/projects/{id}/terminal/sessions/{sessionId}
WS     /api/v1/projects/{id}/chat/ws
WS     /api/v1/projects/{id}/terminal/ws/{sessionId}
```

---

## Phase 2: Security, Collaboration & Multi-Agent Intelligence

**Status: Complete**

Adds email verification, two-factor authentication, role-based access control with organizations, multi-agent team orchestration, and project knowledge base.

### Backend

| Component | Description | Files |
|-----------|-------------|-------|
| Email Verification | Token generation, SMTP sending, verification endpoint | `internal/auth/service.go` (extended), `internal/email/service.go` |
| Two-Factor Auth | TOTP setup/confirm/disable, backup codes, login verification | `internal/auth/totp.go` |
| RBAC Service | Role enforcement (owner, admin, member) at organization level | `internal/rbac/service.go` |
| Organization Service | Org CRUD, member management, email invitations with tokens | `internal/organization/service.go` |
| Memory Service | Project-scoped knowledge base with categories, pinning, GIN full-text search | `internal/memory/service.go` |
| Team Service | Multi-agent team CRUD with sequential/parallel strategies | `internal/team/service.go` |
| Orchestrator | Hierarchical task decomposition, inter-agent delegation, result synthesis | `internal/team/orchestrator.go` |

### Database (Migrations 000002–000006)

| Table | Purpose |
|-------|---------|
| `email_verification_tokens` | Time-limited tokens for email verification |
| `two_factor_backup_codes` | One-time recovery codes for 2FA |
| `organizations` | Team organizations with owner and settings |
| `organization_members` | Membership with role (owner/admin/member) |
| `organization_invitations` | Pending invitations with expiry |
| `team_memories` | Knowledge base entries with category/key/content/pinned |
| `agent_teams` | Team definitions with strategy (sequential/parallel) and leader |
| `agent_team_members` | Agent-to-team membership with role and priority |
| `agent_tasks` | Orchestration tasks with parent relationships and status tracking |

### Frontend

| Component | Description |
|-----------|-------------|
| `TwoFactorForm` | TOTP code input during login |
| `EmailVerificationBanner` | Prompts unverified users to check email |
| `SettingsPage` | 2FA setup/disable with backup codes display |
| `OrgPanel` | Organization management (inline in Settings/Dashboard) |
| `MemoryPanel` | Knowledge base CRUD with search and pin/unpin |
| `TeamPanel` | Team creation, member management, invoke, task history |
| API modules | `organizations.ts`, `memory.ts`, `teams.ts` |

### API Endpoints (New)

```
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/resend-verification
POST   /api/v1/auth/2fa/setup
POST   /api/v1/auth/2fa/confirm
POST   /api/v1/auth/2fa/disable
POST   /api/v1/auth/2fa/verify-login
GET    /api/v1/organizations
POST   /api/v1/organizations
GET    /api/v1/organizations/{orgId}
PUT    /api/v1/organizations/{orgId}
DELETE /api/v1/organizations/{orgId}
GET    /api/v1/organizations/{orgId}/members
POST   /api/v1/organizations/{orgId}/invite
DELETE /api/v1/organizations/{orgId}/members/{userId}
PUT    /api/v1/organizations/{orgId}/members/{userId}/role
POST   /api/v1/invitations/{token}/accept
GET    /api/v1/projects/{id}/memory
POST   /api/v1/projects/{id}/memory
GET    /api/v1/projects/{id}/memory/search
GET    /api/v1/projects/{id}/memory/{category}/{key}
DELETE /api/v1/projects/{id}/memory/{memoryId}
GET    /api/v1/projects/{id}/teams
POST   /api/v1/projects/{id}/teams
GET    /api/v1/projects/{id}/teams/{teamId}
PUT    /api/v1/projects/{id}/teams/{teamId}
DELETE /api/v1/projects/{id}/teams/{teamId}
POST   /api/v1/projects/{id}/teams/{teamId}/invoke
GET    /api/v1/projects/{id}/teams/{teamId}/members
POST   /api/v1/projects/{id}/teams/{teamId}/members
DELETE /api/v1/projects/{id}/teams/{teamId}/members/{agentId}
PUT    /api/v1/projects/{id}/teams/{teamId}/leader
GET    /api/v1/projects/{id}/teams/{teamId}/tasks
GET    /api/v1/projects/{id}/teams/{teamId}/tasks/{taskId}
```

### Sidebar Tabs Added

- Teams (Users icon)
- Memory (Brain icon)

---

## Phase 3: VPS Deployment & Environment Secrets

**Status: Complete**

Adds SSH-based deployment to VPS targets, AES-encrypted environment variable management, and deployment history with logs.

### Backend

| Component | Description | Files |
|-----------|-------------|-------|
| Deployment Service | VPS target CRUD, env var management (AES-encrypted), async SSH deployment | `internal/deployment/service.go` |
| SSH Execution | SSH key & password auth, script execution via stdin pipe, env var injection | `internal/deployment/service.go` (`sshConnect`, `runScript`) |
| Deployment Handler | REST API for targets, env vars, deployment triggers (202 Accepted), run history | `internal/deployment/handler.go` |

### Database (Migration 000007)

| Table | Purpose |
|-------|---------|
| `project_env_vars` | AES-encrypted key/value environment variables per project |
| `vps_targets` | SSH deployment targets with host, auth, commands, deploy path |
| `deployment_runs` | Deployment execution log with status, stdout/stderr, timestamps |

### Key Implementation Details

- **Env var injection**: Shell `export KEY='VALUE'` lines prepended to deploy script
- **SSH auth**: Supports `key` (PEM private key) and `password` authentication
- **Script execution**: Single SSH session with `session.Stdin = strings.NewReader(script)` then `session.Run("bash -s")`
- **Async deploy**: Creates `DeploymentRun` record, goroutine SSHes and updates log/status
- **Encryption**: Same AES-256-GCM pattern as provider tokens for SSH keys, passwords, and env values

### Frontend

| Component | Description |
|-----------|-------------|
| `ProjectSettings` | Dual-tab: Environment Variables (table with secret masking) + VPS Deployment (target config, deploy button, run history) |
| `DevOpsPanel` | Deployment target status cards, one-click deploy, deployment history with expandable logs |
| `LogsPanel` | Split-pane deployment log viewer with target filter and log download |
| API module | `deployment.ts` |

### API Endpoints (New)

```
GET    /api/v1/projects/{id}/env
POST   /api/v1/projects/{id}/env
DELETE /api/v1/projects/{id}/env/{envId}
GET    /api/v1/projects/{id}/deploy
POST   /api/v1/projects/{id}/deploy
PUT    /api/v1/projects/{id}/deploy/{targetId}
DELETE /api/v1/projects/{id}/deploy/{targetId}
POST   /api/v1/projects/{id}/deploy/{targetId}/trigger
GET    /api/v1/projects/{id}/deploy/runs
GET    /api/v1/projects/{id}/deploy/{targetId}/runs
GET    /api/v1/projects/{id}/runs/{runId}
```

### Sidebar Tabs Added

- DevOps (Wrench icon)
- Logs (FileText icon)
- Settings (Settings icon)

---

## Phase 4: Webhooks, Monitoring, Notifications & Pipeline

**Status: Complete**

Adds GitHub webhook auto-deploy, SSH-based server monitoring, real-time WebSocket notifications, and ordered pipeline stages.

### Backend

| Component | Description | Files |
|-----------|-------------|-------|
| Webhook Service | GitHub HMAC-SHA256 validation, branch matching, auto-deploy trigger | `internal/webhook/service.go` |
| Webhook Handler | Public ingress endpoint (no auth, HMAC-verified), authenticated CRUD, secret shown once on create | `internal/webhook/handler.go` |
| Notification Service | DB persistence + live push via WebSocket hub | `internal/notification/service.go` |
| Notification Hub | Per-user connection map (`map[uuid.UUID]map[*websocket.Conn]struct{}`), concurrent-safe | `internal/notification/service.go` |
| Notification Handler | REST CRUD + WebSocket upgrade with JWT query-param auth | `internal/notification/handler.go` |
| Monitoring Service | SSH-based metric collection (compound bash command for CPU/mem/disk/load/uptime), pipeline stage CRUD | `internal/monitoring/service.go` |
| SSHRunner Interface | `monitoring.SSHRunner` interface implemented by `deployment.Service.RunSSHCommand` | `internal/deployment/service.go` |
| JWT Helper | `auth.GetJWTServiceFromContext()` for WebSocket auth | `internal/auth/middleware.go` |

### Database (Migration 000008)

| Table | Purpose |
|-------|---------|
| `webhooks` | GitHub webhook config with HMAC secret, branch filter, VPS target link |
| `pipeline_stages` | Ordered shell commands per VPS target with on-failure behavior |
| `notifications` | User notification inbox with type, title, message, read status, JSONB data |
| `server_metrics` | Point-in-time VPS health snapshots (CPU, memory, disk, load, uptime) |

### Key Implementation Details

- **Webhook verification**: `X-Hub-Signature-256` header with `sha256=<hex>` HMAC format
- **Webhook secret**: Generated as 64-char hex on create, shown once in response, encrypted at rest
- **Auto-deploy flow**: GitHub push → HMAC verify → branch match → async deploy + notification
- **Metric collection**: SSH compound bash command parsing `/proc/stat`, `free`, `df`, `/proc/loadavg`, `/proc/uptime`
- **Pipeline stages**: `on_failure` can be `abort` or `continue`
- **WebSocket notifications**: JWT token passed as `?token=` query param, hub broadcasts to all user connections

### Frontend

| Component | Description |
|-----------|-------------|
| `NotificationBell` | Header bell icon with unread badge, WebSocket for live push, dropdown notification list |
| `notificationStore` | Zustand store with `addLive`, `markRead`, `markAllRead`, `setUnreadCount` |
| `DevOpsPanel` (rewritten) | 4 sub-tabs: Overview (deploy status + history), Metrics (SSH gauge cards + CPU bar chart), Webhooks (copy URL, reveal secret, toggle active), Pipeline (ordered stage list) |
| API modules | `notifications.ts`, `webhooks.ts` (includes `monitoringApi`) |

### API Endpoints (New)

```
POST   /webhooks/github/{webhookId}              (Public, HMAC-verified)
GET    /api/v1/notifications
POST   /api/v1/notifications/{id}/read
POST   /api/v1/notifications/read-all
DELETE /api/v1/notifications/{id}
WS     /api/v1/notifications/ws
GET    /api/v1/projects/{id}/webhooks
POST   /api/v1/projects/{id}/webhooks
PUT    /api/v1/projects/{id}/webhooks/{webhookId}/active
DELETE /api/v1/projects/{id}/webhooks/{webhookId}
GET    /api/v1/projects/{id}/webhooks/{webhookId}/secret
POST   /api/v1/projects/{id}/deploy/{targetId}/metrics/collect
GET    /api/v1/projects/{id}/deploy/{targetId}/metrics/latest
GET    /api/v1/projects/{id}/deploy/{targetId}/metrics/history
GET    /api/v1/projects/{id}/deploy/{targetId}/stages
POST   /api/v1/projects/{id}/deploy/{targetId}/stages
DELETE /api/v1/projects/{id}/deploy/{targetId}/stages/{stageId}
```

---

## Phase 5: Agent Tool Ecosystem & Usage Analytics

**Status: Complete**

Adds a tool registry for agents (function, HTTP, shell tools), agent-tool bindings, tool execution logging, per-invocation usage analytics with cost tracking, and monthly budget management.

### Backend

| Component | Description | Files |
|-----------|-------------|-------|
| Tools Service | Tool CRUD (function/http/shell types), agent-tool binding, execution engine | `internal/tools/service.go` |
| HTTP Tool Runner | HTTP request executor with URL, method, headers, body template (`{{.input}}`) | `internal/tools/service.go` (`runHTTPTool`) |
| Shell Tool Runner | Bash command executor with timeout, `{{.input}}` substitution | `internal/tools/service.go` (`runShellTool`) |
| Tools Handler | REST API for tool CRUD, bindings, execution, execution log | `internal/tools/handler.go` |
| Analytics Service | Usage recording with model-specific cost tables, daily summaries, model breakdowns, budget management | `internal/analytics/service.go` |
| Analytics Handler | REST API for summary, daily, models, recent, budget CRUD | `internal/analytics/handler.go` |
| Agent Integration | `UsageRecorder` and `ToolProvider` interfaces on agent.Service (avoids import cycles), tool descriptions injected into system prompt, analytics recorded per invocation | `internal/agent/service.go` |

### Database (Migration 000009)

| Table | Purpose |
|-------|---------|
| `agent_tools` | Tool definitions with name (unique per project), type, JSONB definition schema |
| `agent_tool_bindings` | Many-to-many agent-to-tool links (unique constraint) |
| `tool_executions` | Execution log with input (JSONB), output (text), status, duration_ms, error_message |
| `usage_records` | Per-invocation: user, project, agent, provider, model, input/output tokens, cost in microcents, tool call count |
| `cost_budgets` | Monthly spend cap per project with alert threshold percentage |

### Key Implementation Details

- **Tool types**: `function` (AI-native, schema-only), `http` (full HTTP client), `shell` (bash with timeout)
- **Tool injection**: Bound tools are described in the agent's system prompt so the AI model knows what's available
- **Cost calculation**: Model-specific lookup table (microcents per token for input/output), covers GPT-4o, GPT-4, GPT-3.5, Claude Opus/Sonnet/Haiku variants
- **Budget check**: `CheckBudget()` returns (currentSpend, limit, overBudget) for current calendar month
- **Interface pattern**: `agent.UsageRecorder` and `agent.ToolProvider` interfaces avoid circular imports between agent → analytics and agent → tools

### Model Cost Table

| Model | Input (microcents/tok) | Output (microcents/tok) |
|-------|----------------------|------------------------|
| gpt-4o | 25 | 100 |
| gpt-4o-mini | 2 | 6 |
| gpt-4-turbo | 100 | 300 |
| gpt-4 | 300 | 600 |
| gpt-3.5-turbo | 5 | 15 |
| claude-opus-4-20250514 | 150 | 750 |
| claude-sonnet-4-20250514 | 30 | 150 |
| claude-3-5-sonnet-20241022 | 30 | 150 |
| claude-3-5-haiku-20241022 | 10 | 50 |
| claude-3-opus-20240229 | 150 | 750 |
| claude-3-haiku-20240307 | 2 | 12 |

### Frontend

| Component | Description |
|-----------|-------------|
| `ToolsPanel` | 3 sub-tabs: Tool Registry (CRUD + enable/disable toggle), Agent Bindings (bind/unbind per agent), Execution Log (status, duration, output) |
| `AnalyticsPanel` | 4 sub-tabs: Overview (stat cards + daily bar chart + budget progress bar), Models (per-model cost breakdown with proportion bars), History (usage table), Budget (monthly limit + alert threshold config) |
| API modules | `tools.ts`, `analytics.ts` |

### API Endpoints (New)

```
GET    /api/v1/projects/{id}/tools
POST   /api/v1/projects/{id}/tools
PUT    /api/v1/projects/{id}/tools/{toolId}
DELETE /api/v1/projects/{id}/tools/{toolId}
GET    /api/v1/projects/{id}/tools/executions
GET    /api/v1/projects/{id}/agents/{agentId}/tools
POST   /api/v1/projects/{id}/agents/{agentId}/tools
DELETE /api/v1/projects/{id}/agents/{agentId}/tools/{toolId}
POST   /api/v1/projects/{id}/agents/{agentId}/tools/{toolId}/execute
GET    /api/v1/projects/{id}/analytics/summary
GET    /api/v1/projects/{id}/analytics/daily
GET    /api/v1/projects/{id}/analytics/models
GET    /api/v1/projects/{id}/analytics/recent
GET    /api/v1/projects/{id}/analytics/budget
POST   /api/v1/projects/{id}/analytics/budget
```

### Sidebar Tabs Added

- Tools (Puzzle icon)
- Analytics (BarChart3 icon)

---

## Future Phases (Planned)

### Phase 6: Workflow Automation & Scheduling

- Visual workflow builder (drag-and-drop agent chains)
- Cron job scheduling for recurring team invocations
- Trigger-action patterns (on deploy success → run tests → notify)
- State machine definitions for complex multi-step workflows
- Webhook outbound events (notify external systems on events)

### Phase 7: Advanced Knowledge Management

- Vector embeddings for semantic memory search (pgvector)
- Document ingestion (PDF, Markdown, Word) with chunking and indexing
- Memory versioning with diff and rollback
- Cross-project knowledge sharing within organizations
- Auto-memory extraction from chat conversations

### Phase 8: Multi-Channel Notifications & Integrations

- Slack integration (incoming + outgoing)
- Discord bot integration
- Email digest summaries (daily/weekly)
- Custom webhook outbound events
- Notification rules and filtering (per-user preferences)
- Zapier/Make.com webhook compatibility

### Phase 9: Enterprise & Compliance

- Audit logging (who did what, when, from where)
- Data export/import (full project backup as JSON/ZIP)
- Advanced RBAC (resource-level permissions, custom roles)
- Activity feeds per project (timeline of all actions)
- Data retention policies with automatic cleanup
- SSO / SAML integration

### Phase 10: Platform Polish & Scale

- Dark mode toggle with system preference detection
- Streaming agent responses (SSE/WebSocket streaming from AI providers)
- Agent prompt versioning and A/B testing
- Workspace favorites and pinned projects
- Bulk operations (multi-select agents, tools, env vars)
- Advanced search across all project data
- Self-hosted deployment guide with Helm chart / Docker Compose production config
- Horizontal scaling documentation (multiple server instances with shared PostgreSQL)

---

## Cumulative Statistics

| Metric | Count |
|--------|-------|
| Database tables | 27 |
| Database migrations | 9 |
| Backend service packages | 18 (15 domain + 3 infrastructure) |
| API endpoints | 60+ |
| WebSocket hubs | 3 (chat, terminal, notifications) |
| Frontend components | 16 directories |
| Frontend API modules | 11 |
| Sidebar navigation tabs | 10 |
| Zustand stores | 2 (auth, notifications) |

---

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Go + Chi | High performance, simple dependency injection, stdlib-compatible middleware |
| GORM | Productive ORM for rapid development, auto-migration on startup |
| PostgreSQL 16 | JSONB for flexible schemas, GIN for full-text search, `gen_random_uuid()` for UUIDs |
| React 19 + Vite 7 | Fast HMR, modern React features, TypeScript-first |
| Tailwind CSS v4 | Utility-first styling with CSS custom properties for theming |
| Zustand | Minimal global state management without boilerplate |
| gorilla/websocket | Battle-tested WebSocket library for Go |
| AES-256-GCM | Industry-standard authenticated encryption for secrets at rest |
| Interface injection | Consumer-defined interfaces (e.g., `agent.UsageRecorder`) to avoid circular imports |
| Async deployment | Goroutine-based SSH execution with immediate 202 response |
