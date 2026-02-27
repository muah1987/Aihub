# Development Guide

## Prerequisites

- Go 1.24+
- Node.js 20+ / npm 10+
- PostgreSQL 16 (or Docker)
- Docker (optional, for terminal sandbox feature)

## Initial Setup

```bash
git clone https://github.com/muah1987/Aihub.git
cd Aihub

# Install all dependencies
make setup

# Copy and configure environment
cp .env.example .env
```

## Running Locally

### Start the database

```bash
# Option A: Docker (recommended)
make docker-up
# Starts PostgreSQL on port 5432

# Option B: Local PostgreSQL
createdb aihub
```

### Start the backend

```bash
make run
# Server starts on http://localhost:8080
# Migrations run automatically
```

### Start the frontend

```bash
make web-dev
# Vite dev server on http://localhost:5173 with HMR
```

## Project Structure

```
cmd/server/main.go       # Server entrypoint, service wiring
internal/
  <domain>/
    service.go            # Business logic
    handler.go            # HTTP handlers
web/
  src/
    api/<domain>.ts       # Typed API client
    components/<domain>/  # React components
    pages/                # Page views
    store/                # Zustand stores
```

## Adding a New Feature

### Backend

1. **Migration** — Create `internal/database/migrations/000XXX_feature.up.sql`
2. **Model** — Add GORM model to `internal/models/`
3. **Service** — Create `internal/<feature>/service.go` with business logic
4. **Handler** — Create `internal/<feature>/handler.go` with HTTP handlers
5. **Router** — Add routes in `internal/router/router.go`
6. **Bootstrap** — Wire service in `cmd/server/main.go`

### Frontend

1. **API module** — Create `web/src/api/<feature>.ts`
2. **Component** — Create `web/src/components/<feature>/`
3. **Integration** — Add to `ProjectDashboard.tsx` switch and `Sidebar.tsx` tabs

## Code Conventions

### Backend

- **Error handling**: Return `fmt.Errorf("context: %w", err)` — always wrap with context
- **Handler pattern**: Parse request → validate → call service → write JSON response
- **JSON responses**: Always wrap in an object: `{"agents": [...]}`, not bare arrays
- **UUIDs**: Use `github.com/google/uuid` for all IDs, `gen_random_uuid()` in PostgreSQL
- **Encryption**: Use `crypto/aes` + `crypto/cipher` GCM mode for secrets
- **Interfaces**: Define in the consumer package to avoid import cycles

### Frontend

- **TypeScript**: Strict mode, explicit types for API responses
- **Components**: Functional components with hooks
- **Styling**: Tailwind CSS v4 with CSS custom properties (`var(--color-*)`)
- **State**: Zustand for global state, `useState` for local component state
- **API calls**: All go through typed modules in `web/src/api/`

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build Go binary to `bin/aihub-server` |
| `make run` | Run server with `go run` |
| `make test` | Run all Go tests |
| `make lint` | Run golangci-lint |
| `make docker-up` | Start PostgreSQL via Docker Compose |
| `make docker-down` | Stop Docker Compose services |
| `make docker-build` | Build all Docker images |
| `make web-dev` | Start Vite dev server |
| `make web-build` | Build frontend for production |
| `make web-install` | Install npm dependencies |
| `make setup` | Full initial setup |
| `make clean` | Remove build artifacts and node_modules |

## Type Checking

### Go

```bash
go build ./...
go vet ./...
```

### TypeScript

```bash
cd web && npx tsc --noEmit
```

## Database Migrations

Migrations are in `internal/database/migrations/` and run automatically on server start via GORM's AutoMigrate pattern.

Migration files follow the naming convention:

```
000001_init.up.sql
000002_email_verification.up.sql
...
000009_phase5.up.sql
```

Current tables (27 total):

| Migration | Tables |
|-----------|--------|
| 000001 | `users`, `provider_connections`, `projects`, `messages`, `agents`, `terminal_sessions` |
| 000002 | `email_verification_tokens` |
| 000003 | `two_factor_backup_codes` |
| 000004 | `organizations`, `organization_members`, `organization_invitations` |
| 000005 | `team_memories` |
| 000006 | `agent_teams`, `agent_team_members`, `agent_tasks` |
| 000007 | `project_env_vars`, `vps_targets`, `deployment_runs` |
| 000008 | `webhooks`, `pipeline_stages`, `notifications`, `server_metrics` |
| 000009 | `agent_tools`, `agent_tool_bindings`, `tool_executions`, `usage_records`, `cost_budgets` |

## WebSocket Development

Three WebSocket endpoints are available for real-time features:

```
Chat:          ws://localhost:8080/api/v1/projects/{id}/chat/ws?token=<jwt>
Terminal:      ws://localhost:8080/api/v1/projects/{id}/terminal/ws/{sessionId}?token=<jwt>
Notifications: ws://localhost:8080/api/v1/notifications/ws?token=<jwt>
```

Authentication is via JWT token passed as a `?token=` query parameter.

## Testing API Endpoints

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"password123","name":"Test"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@test.com","password":"password123"}'

# Use token
export TOKEN="<access_token_from_login>"

# List projects
curl http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN"
```

## Docker Sandbox Setup

The terminal sandbox feature requires a Docker image:

```bash
cd docker/sandbox
docker build -t aihub-sandbox:latest .
```

If Docker is unavailable, the server starts normally with terminal features disabled (logged as a warning).
