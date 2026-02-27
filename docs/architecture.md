# Architecture Overview

Aihub is a monorepo with a Go backend API server and a React single-page application frontend.

## System Diagram

```
┌─────────────┐      ┌──────────────────────────────────────────────────┐
│   Browser    │◄────►│              React SPA (Vite)                    │
│              │      │  Zustand stores, Axios client, WebSocket hooks  │
└──────┬───────┘      └──────────────────────────────────────────────────┘
       │ HTTP / WS
       ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        Go API Server (Chi)                           │
│                                                                      │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Auth  │ │ Agent  │ │  Chat  │ │ Terminal │ │   Deployment     │ │
│  │  JWT   │ │ Invoke │ │  Hub   │ │ Docker   │ │   SSH / VPS      │ │
│  │  2FA   │ │ Tools  │ │  WS    │ │ Sandbox  │ │   Env Secrets    │ │
│  └────────┘ └────────┘ └────────┘ └──────────┘ └──────────────────┘ │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌─────────┐ ┌───────────────┐  │
│  │ Teams  │ │ Memory │ │ Webhook  │ │ Monitor │ │ Notifications │  │
│  │ Orch.  │ │  KB    │ │ GitHub   │ │ Metrics │ │ WS Hub        │  │
│  └────────┘ └────────┘ └──────────┘ └─────────┘ └───────────────┘  │
│  ┌────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                  │
│  │ Tools  │ │Analytics │ │   RBAC   │ │   Org    │                  │
│  │Registry│ │  Usage   │ │  Roles   │ │  Teams   │                  │
│  └────────┘ └──────────┘ └──────────┘ └──────────┘                  │
└──────────────────────────────┬───────────────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   PostgreSQL 16     │
                    │   27 tables         │
                    │   GIN full-text     │
                    │   JSONB columns     │
                    └─────────────────────┘
```

## Backend Architecture

### Service Layer Pattern

Every domain follows the same three-file pattern:

```
internal/<domain>/
  ├── service.go   # Business logic, DB queries, core operations
  ├── handler.go   # HTTP handlers, request parsing, response writing
  └── (optional)   # hub.go for WebSocket hubs, orchestrator.go, provider.go
```

Services are wired together in `cmd/server/main.go` using constructor injection. Interfaces are used to avoid import cycles (e.g., `agent.UsageRecorder` is satisfied by `analytics.Service`).

### Request Flow

```
HTTP Request
  → Chi Router (middleware: CORS, Logger, RateLimiter, JWT injection)
    → Auth Middleware (validates JWT, injects user ID into context)
      → Handler (parses request, calls service)
        → Service (business logic, DB operations)
          → GORM / PostgreSQL
```

### WebSocket Architecture

Three independent WebSocket hubs:

| Hub | Purpose | Auth | Routing |
|-----|---------|------|---------|
| Chat Hub | Project chat rooms | JWT query param | `/api/v1/projects/{id}/chat/ws` |
| Terminal Hub | Docker shell sessions | JWT query param | `/api/v1/projects/{id}/terminal/ws/{sessionId}` |
| Notification Hub | Per-user notifications | JWT query param | `/api/v1/notifications/ws` |

Each hub maintains a concurrent-safe map of connections and broadcasts messages to relevant clients.

### Encryption

Sensitive values are encrypted at rest using AES-256-GCM:

- Provider API tokens (`provider_connections.access_token`)
- Environment variable values (`project_env_vars.value`)
- SSH keys and passwords (`vps_targets.ssh_key`, `vps_targets.ssh_password`)
- Webhook secrets (`webhooks.secret`)

The encryption key is configured via the `ENCRYPTION_KEY` environment variable (32-byte hex string).

### Authentication Flow

```
Register → Email verification token sent → Verify email
Login → JWT access + refresh tokens returned
       → If 2FA enabled: pending token returned, must verify TOTP
Access token (15m) → Sent as Bearer header
Refresh token (7d) → Used to rotate access tokens
```

## Frontend Architecture

### Component Organization

```
web/src/
├── api/           # Typed Axios client modules (one per backend domain)
├── components/
│   ├── layout/    # AppShell, Header, Sidebar, TabBar
│   ├── auth/      # Login, Register, 2FA, Email verification
│   ├── project/   # Project list, dashboard, settings
│   ├── chat/      # Chat timeline, message bubbles, input
│   ├── agent/     # Agent panel (CRUD + invoke)
│   ├── team/      # Team panel (orchestration)
│   ├── memory/    # Memory panel (knowledge base)
│   ├── tools/     # Tool registry, bindings, execution log
│   ├── analytics/ # Usage analytics dashboard
│   ├── terminal/  # xterm.js terminal view
│   ├── devops/    # Deployment, webhooks, metrics, pipeline
│   ├── log-viewer/# Deployment log viewer
│   ├── notifications/ # Notification bell with WebSocket
│   ├── organization/  # Organization management
│   └── ui/        # Reusable primitives (Button, Card, Input, Modal, Toast)
├── pages/         # Dashboard, Login, VerifyEmail, Organization, Settings
└── store/         # Zustand stores (authStore, notificationStore)
```

### State Management

- **Zustand** for global state (auth tokens, user profile, notifications)
- **React local state** for component-level UI state
- **Axios interceptors** for automatic token refresh on 401

### Routing

The app uses a tab-based navigation model:

1. **Unauthenticated**: Login / Register pages
2. **Authenticated**: Dashboard → Project list
3. **Project selected**: ProjectDashboard with tabbed panels
   - Chat, CLI, Agents, Teams, Memory, Tools, Analytics, DevOps, Logs, Settings

## Data Flow Examples

### Agent Invocation with Tools

```
User clicks "Invoke" on Agent panel
  → POST /api/v1/projects/{id}/agents/{agentId}/invoke
    → agent.Service.Invoke()
      1. Fetch agent + provider connection
      2. Decrypt provider API token
      3. Load team memory → inject into system prompt
      4. Load bound tools → inject descriptions into system prompt
      5. Call AI provider API (OpenAI/Anthropic)
      6. Record usage analytics (tokens, cost, tool count)
      7. Update agent token_usage_total
      8. Post response to project chat
      9. Chat hub broadcasts to WebSocket clients
```

### VPS Deployment

```
User clicks "Deploy" (or GitHub webhook fires)
  → POST /api/v1/projects/{id}/deploy/{targetId}/trigger
    → deployment.Service.Deploy()
      1. Create DeploymentRun record (status: running)
      2. Spawn goroutine:
         a. SSH connect to VPS (key or password auth)
         b. Build shell script:
            - Export project env vars
            - cd to deploy path
            - Run pre-deploy, deploy, post-deploy commands
         c. Execute via single SSH session (stdin pipe)
         d. Capture stdout/stderr into log_output
         e. Update DeploymentRun (status: success/failed)
      3. Return 202 Accepted immediately
```

## Security Model

| Concern | Implementation |
|---------|---------------|
| Authentication | JWT with short-lived access tokens (15m) + refresh tokens (7d) |
| 2FA | TOTP (RFC 6238) with encrypted backup codes |
| Authorization | RBAC with organization-level roles (owner, admin, member) |
| Secrets | AES-256-GCM encryption at rest |
| Webhooks | HMAC-SHA256 signature verification |
| Terminal | Docker container isolation with memory/CPU limits |
| Rate limiting | Token bucket (10 req/s, 50 burst) |
| CORS | Configurable allowed origins |
| Inputs | JSON schema validation, parameterized SQL via GORM |
