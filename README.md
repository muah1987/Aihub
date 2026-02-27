# Aihub

AI-native web platform for managing multi-agent AI workflows, team orchestration, VPS deployment, and real-time collaboration — all from a unified project dashboard.

## Features

- **Multi-Model Agents** — Create agents backed by OpenAI or Anthropic, each with custom system prompts, temperature, and token limits
- **Multi-Agent Teams** — Orchestrate agent teams with sequential or parallel execution strategies and hierarchical task decomposition
- **Agent Tool Ecosystem** — Define function, HTTP, and shell tools that agents can invoke during conversations
- **Team Memory** — Project-scoped knowledge base with categories, pinning, and full-text search (GIN indexes)
- **VPS Deployment** — SSH-based deployment to multiple targets with pre/deploy/post commands and encrypted secrets
- **CI/CD Pipeline** — GitHub webhook auto-deploy with HMAC-SHA256 verification, ordered pipeline stages
- **Server Monitoring** — SSH-based metric collection (CPU, memory, disk, load, uptime) with historical storage
- **Real-Time Notifications** — WebSocket-powered notification hub with per-user live push
- **Usage Analytics** — Per-invocation token and cost tracking with daily charts, model breakdowns, and monthly budget caps
- **Terminal Sandbox** — Docker-containerized shell sessions per project
- **Organizations & RBAC** — Team organizations with role-based access control (owner, admin, member)
- **2FA / TOTP** — Two-factor authentication with backup codes
- **Email Verification** — SMTP-based email verification flow

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Chi router, GORM, PostgreSQL 16 |
| Frontend | React 19, TypeScript 5.9, Vite 7, Tailwind CSS v4 |
| Real-Time | gorilla/websocket (chat, terminal, notifications) |
| Auth | JWT (access + refresh), TOTP 2FA, AES-256-GCM encryption |
| Deployment | SSH via golang.org/x/crypto/ssh |
| Containers | Docker SDK for terminal sandbox |
| State | Zustand (frontend) |

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 16
- Docker (optional, for terminal sandbox)

### 1. Clone and install

```bash
git clone https://github.com/muah1987/Aihub.git
cd Aihub
make setup
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your database URL, JWT secret, encryption key, etc.
```

### 3. Start database

```bash
# Option A: Docker
make docker-up

# Option B: Local PostgreSQL
createdb aihub
```

### 4. Run the server

```bash
make run
# Server starts on http://localhost:8080
```

### 5. Run the frontend

```bash
make web-dev
# Frontend starts on http://localhost:5173
```

### Full Docker Setup

```bash
docker compose up -d
# API: http://localhost:8080
# Web: http://localhost:5173
```

## Project Structure

```
Aihub/
├── cmd/server/          # Server entrypoint
├── internal/
│   ├── agent/           # AI agent CRUD + invocation
│   ├── analytics/       # Usage tracking + cost budgets
│   ├── auth/            # JWT, 2FA, email verification
│   ├── chat/            # Project chat + WebSocket hub
│   ├── config/          # Environment configuration
│   ├── database/        # DB connection + migrations
│   ├── deployment/      # VPS deployment + SSH
│   ├── email/           # SMTP service
│   ├── memory/          # Team memory knowledge base
│   ├── middleware/       # CORS, logging, rate limiting
│   ├── monitoring/      # Server metrics + pipeline stages
│   ├── notification/    # In-app notifications + WebSocket hub
│   ├── organization/    # Org management + invitations
│   ├── project/         # Project CRUD
│   ├── provider/        # AI provider connections
│   ├── rbac/            # Role-based access control
│   ├── router/          # API routing
│   ├── team/            # Multi-agent teams + orchestrator
│   ├── terminal/        # Docker sandbox sessions
│   ├── tools/           # Agent tool registry + execution
│   └── webhook/         # GitHub webhook ingress
├── web/
│   └── src/
│       ├── api/         # Typed API client modules
│       ├── components/  # React components
│       ├── pages/       # Page-level views
│       └── store/       # Zustand stores
├── docker/              # Dockerfiles + nginx config
├── docs/                # Documentation
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Documentation

- [Architecture Overview](docs/architecture.md)
- [API Reference](docs/api-reference.md)
- [Deployment Guide](docs/deployment.md)
- [Development Guide](docs/development.md)
- [Database Schema](docs/database.md)

## Environment Variables

See [`.env.example`](.env.example) for all configuration options with descriptions.

## License

Private — all rights reserved.
