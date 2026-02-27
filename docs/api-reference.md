# API Reference

Base URL: `/api/v1`

All protected endpoints require a `Authorization: Bearer <access_token>` header.

---

## Authentication

### POST /auth/register

Create a new user account. Sends a verification email if SMTP is configured.

```json
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe"
}
```

**Response** `201`

```json
{
  "user": { "id": "uuid", "email": "...", "name": "...", "email_verified": false },
  "tokens": { "access_token": "...", "refresh_token": "..." }
}
```

### POST /auth/login

```json
{ "email": "user@example.com", "password": "securepassword" }
```

**Response** `200` — returns tokens, or `pending_2fa_token` if 2FA is enabled.

### POST /auth/refresh

```json
{ "refresh_token": "..." }
```

### POST /auth/verify-email

```json
{ "token": "verification-token-from-email" }
```

### POST /auth/2fa/verify-login

```json
{ "pending_token": "...", "code": "123456" }
```

### GET /auth/me `Protected`

Returns the current user profile.

### POST /auth/2fa/setup `Protected`

Returns TOTP secret and QR code data URI.

### POST /auth/2fa/confirm `Protected`

```json
{ "code": "123456" }
```

Enables 2FA and returns backup codes.

### POST /auth/2fa/disable `Protected`

```json
{ "code": "123456" }
```

---

## Provider Connections

### GET /providers `Protected`

List all provider connections for the current user.

### POST /providers `Protected`

```json
{
  "provider_type": "openai",
  "provider_name": "My OpenAI Key",
  "access_token": "sk-..."
}
```

### DELETE /providers/{id} `Protected`

### POST /providers/{id}/validate `Protected`

Validates the stored API key against the provider.

---

## Projects

### GET /projects `Protected`

List all projects for the current user.

### POST /projects `Protected`

```json
{
  "name": "My Project",
  "repo_url": "https://github.com/user/repo",
  "repo_provider": "github",
  "repo_owner": "user",
  "repo_name": "repo",
  "provider_connection_id": "uuid"
}
```

### GET /projects/{id} `Protected`

### PUT /projects/{id} `Protected`

### DELETE /projects/{id} `Protected`

### GET /projects/{id}/repos `Protected`

List GitHub repos from the project's provider connection.

---

## Chat Messages

### GET /projects/{id}/messages `Protected`

Returns the message history for the project.

### POST /projects/{id}/messages `Protected`

```json
{ "content": "Hello!", "message_type": "user_message" }
```

### WebSocket: /projects/{id}/chat/ws?token={access_token}

Real-time chat. Messages are broadcast to all connected clients for the project.

---

## Agents

### GET /projects/{id}/agents `Protected`

### POST /projects/{id}/agents `Protected`

```json
{
  "name": "Code Reviewer",
  "role": "developer",
  "model": "gpt-4o",
  "system_prompt": "You are a code reviewer...",
  "temperature": 0.7,
  "max_tokens": 4096,
  "provider_connection_id": "uuid"
}
```

### PUT /projects/{id}/agents/{agentId} `Protected`

### DELETE /projects/{id}/agents/{agentId} `Protected`

### POST /projects/{id}/agents/{agentId}/invoke `Protected`

```json
{ "prompt": "Review this function for bugs..." }
```

Response is also posted to the project chat.

---

## Agent Tools

### GET /projects/{id}/tools `Protected`

List all tools in the project.

### POST /projects/{id}/tools `Protected`

```json
{
  "name": "fetch_weather",
  "description": "Fetches weather for a given city",
  "tool_type": "http",
  "definition": {
    "url": "https://api.weather.example/v1/current",
    "method": "GET",
    "headers": { "Authorization": "Bearer xxx" }
  }
}
```

Tool types:
- `function` — Schema-only, used by AI model natively
- `http` — Makes an HTTP request with optional body template (`{{.input}}` placeholder)
- `shell` — Runs a shell command with optional timeout

### PUT /projects/{id}/tools/{toolId} `Protected`

### DELETE /projects/{id}/tools/{toolId} `Protected`

### GET /projects/{id}/tools/executions `Protected`

Query: `?limit=50` — List tool execution logs.

### GET /projects/{id}/agents/{agentId}/tools `Protected`

List tools bound to an agent.

### POST /projects/{id}/agents/{agentId}/tools `Protected`

```json
{ "tool_id": "uuid" }
```

Bind a tool to an agent.

### DELETE /projects/{id}/agents/{agentId}/tools/{toolId} `Protected`

Unbind a tool from an agent.

### POST /projects/{id}/agents/{agentId}/tools/{toolId}/execute `Protected`

```json
{ "input": { "city": "London" } }
```

Manually execute a tool and log the result.

---

## Agent Teams

### GET /projects/{id}/teams `Protected`

### POST /projects/{id}/teams `Protected`

```json
{
  "name": "Research Team",
  "strategy": "sequential",
  "description": "Researches and summarizes topics"
}
```

Strategies: `sequential`, `parallel`

### GET /projects/{id}/teams/{teamId} `Protected`

### PUT /projects/{id}/teams/{teamId} `Protected`

### DELETE /projects/{id}/teams/{teamId} `Protected`

### POST /projects/{id}/teams/{teamId}/invoke `Protected`

```json
{ "prompt": "Research the latest trends in AI safety" }
```

### GET /projects/{id}/teams/{teamId}/members `Protected`

### POST /projects/{id}/teams/{teamId}/members `Protected`

```json
{ "agent_id": "uuid" }
```

### DELETE /projects/{id}/teams/{teamId}/members/{agentId} `Protected`

### PUT /projects/{id}/teams/{teamId}/leader `Protected`

```json
{ "agent_id": "uuid" }
```

### GET /projects/{id}/teams/{teamId}/tasks `Protected`

### GET /projects/{id}/teams/{teamId}/tasks/{taskId} `Protected`

---

## Team Memory

### GET /projects/{id}/memory `Protected`

### POST /projects/{id}/memory `Protected`

```json
{
  "category": "architecture",
  "key": "database-schema",
  "content": "We use PostgreSQL 16 with 27 tables...",
  "pinned": true
}
```

### GET /projects/{id}/memory/search?q=keyword `Protected`

Full-text search via GIN indexes.

### GET /projects/{id}/memory/{category}/{key} `Protected`

### DELETE /projects/{id}/memory/{memoryId} `Protected`

---

## Terminal Sessions

### GET /projects/{id}/terminal/sessions `Protected`

### POST /projects/{id}/terminal/sessions `Protected`

Creates a Docker sandbox container.

### DELETE /projects/{id}/terminal/sessions/{sessionId} `Protected`

### WebSocket: /projects/{id}/terminal/ws/{sessionId}?token={access_token}

Interactive terminal session. Sends/receives raw terminal data.

---

## Environment Variables

### GET /projects/{id}/env `Protected`

List env vars (secret values are masked).

### POST /projects/{id}/env `Protected`

```json
{ "key": "DATABASE_URL", "value": "postgres://...", "is_secret": true }
```

Upserts. Values are AES-256-GCM encrypted at rest.

### DELETE /projects/{id}/env/{envId} `Protected`

---

## VPS Deployment

### GET /projects/{id}/deploy `Protected`

List VPS deployment targets.

### POST /projects/{id}/deploy `Protected`

```json
{
  "name": "Production VPS",
  "host": "192.168.1.100",
  "port": 22,
  "username": "deploy",
  "auth_type": "key",
  "ssh_key": "-----BEGIN OPENSSH PRIVATE KEY-----...",
  "deploy_path": "/var/www/myapp",
  "pre_deploy_cmd": "git pull origin main",
  "deploy_cmd": "docker compose up -d --build",
  "post_deploy_cmd": "docker system prune -f"
}
```

### PUT /projects/{id}/deploy/{targetId} `Protected`

### DELETE /projects/{id}/deploy/{targetId} `Protected`

### POST /projects/{id}/deploy/{targetId}/trigger `Protected`

Triggers an async deployment. Returns `202 Accepted`.

### GET /projects/{id}/deploy/runs `Protected`

List all deployment runs.

### GET /projects/{id}/deploy/{targetId}/runs `Protected`

List runs for a specific target.

### GET /projects/{id}/runs/{runId} `Protected`

Get a single run with full log output.

---

## Webhooks

### POST /webhooks/github/{webhookId} `Public`

GitHub webhook ingress. Validates `X-Hub-Signature-256` HMAC header. Triggers auto-deploy if branch matches.

### GET /projects/{id}/webhooks `Protected`

### POST /projects/{id}/webhooks `Protected`

```json
{ "vps_target_id": "uuid", "branch": "main" }
```

Returns the generated secret (shown once).

### PUT /projects/{id}/webhooks/{webhookId}/active `Protected`

```json
{ "active": true }
```

### DELETE /projects/{id}/webhooks/{webhookId} `Protected`

### GET /projects/{id}/webhooks/{webhookId}/secret `Protected`

Retrieve the webhook HMAC secret.

---

## Server Monitoring

### POST /projects/{id}/deploy/{targetId}/metrics/collect `Protected`

SSH into the VPS and collect current metrics.

### GET /projects/{id}/deploy/{targetId}/metrics/latest `Protected`

### GET /projects/{id}/deploy/{targetId}/metrics/history?limit=20 `Protected`

### GET /projects/{id}/deploy/{targetId}/stages `Protected`

List pipeline stages for a target.

### POST /projects/{id}/deploy/{targetId}/stages `Protected`

```json
{
  "name": "Run tests",
  "command": "npm test",
  "stage_order": 1,
  "on_failure": "abort"
}
```

### DELETE /projects/{id}/deploy/{targetId}/stages/{stageId} `Protected`

---

## Notifications

### GET /notifications `Protected`

### POST /notifications/{id}/read `Protected`

### POST /notifications/read-all `Protected`

### DELETE /notifications/{id} `Protected`

### WebSocket: /notifications/ws?token={access_token}

Events:
- `connected` — initial payload with `unread_count`
- `notification` — new notification pushed in real-time

---

## Usage Analytics

### GET /projects/{id}/analytics/summary?days=30 `Protected`

```json
{
  "summary": {
    "total_tokens": 125000,
    "total_cost_microcents": 3750000,
    "total_invocations": 42,
    "tool_calls_total": 15
  }
}
```

### GET /projects/{id}/analytics/daily?days=30 `Protected`

Daily breakdown of tokens, cost, and invocations.

### GET /projects/{id}/analytics/models?days=30 `Protected`

Per-model breakdown.

### GET /projects/{id}/analytics/recent?limit=50 `Protected`

Recent usage records.

### GET /projects/{id}/analytics/budget `Protected`

Returns current budget, spend, and over-budget status.

### POST /projects/{id}/analytics/budget `Protected`

```json
{
  "monthly_limit_microcents": 10000000,
  "alert_threshold_pct": 80
}
```

---

## Organizations

### GET /organizations `Protected`

### POST /organizations `Protected`

```json
{ "name": "My Team", "description": "Our development team" }
```

### GET /organizations/{orgId} `Protected`

### PUT /organizations/{orgId} `Protected`

### DELETE /organizations/{orgId} `Protected`

### GET /organizations/{orgId}/members `Protected`

### POST /organizations/{orgId}/invite `Protected`

```json
{ "email": "teammate@example.com", "role": "member" }
```

### DELETE /organizations/{orgId}/members/{userId} `Protected`

### PUT /organizations/{orgId}/members/{userId}/role `Protected`

```json
{ "role": "admin" }
```

### POST /invitations/{token}/accept `Protected`
