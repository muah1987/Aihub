# Database Schema

Aihub uses PostgreSQL 16 with GORM as the ORM. All primary keys are UUIDs generated via `gen_random_uuid()`.

## Entity Relationship Overview

```
users ─────────┬──── provider_connections
               ├──── projects ──────────┬──── messages
               ├──── organizations      ├──── agents ──────────┬──── agent_tool_bindings ── agent_tools
               ├──── notifications      ├──── terminal_sessions├──── tool_executions
               └──── usage_records      ├──── team_memories    │
                                        ├──── agent_teams ─────┤
                                        │     └─ agent_team_members
                                        │     └─ agent_tasks
                                        ├──── project_env_vars
                                        ├──── vps_targets ─────┬──── deployment_runs
                                        │                      ├──── webhooks
                                        │                      ├──── pipeline_stages
                                        │                      └──── server_metrics
                                        ├──── agent_tools
                                        ├──── cost_budgets
                                        └──── webhooks
```

## Tables

### users

Core user accounts.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| email | VARCHAR(255) | Unique |
| password_hash | TEXT | bcrypt |
| name | VARCHAR(255) | |
| email_verified | BOOLEAN | Default false |
| two_factor_enabled | BOOLEAN | Default false |
| two_factor_secret | TEXT | Encrypted TOTP secret |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### provider_connections

Stores API keys for external providers (OpenAI, Anthropic, GitHub).

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| provider_type | VARCHAR(50) | `openai`, `anthropic`, `github` |
| provider_name | VARCHAR(255) | User-given label |
| access_token | TEXT | AES-256-GCM encrypted |
| status | VARCHAR(20) | `active`, `invalid` |
| created_at | TIMESTAMPTZ | |

### projects

Each project maps to one GitHub repository.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| name | VARCHAR(255) | |
| description | TEXT | |
| repo_provider | VARCHAR(50) | `github` |
| repo_owner | VARCHAR(255) | |
| repo_name | VARCHAR(255) | |
| repo_url | VARCHAR(500) | |
| repo_default_branch | VARCHAR(100) | Default `main` |
| provider_connection_id | UUID | FK → provider_connections (nullable) |
| organization_id | UUID | FK → organizations (nullable) |
| status | VARCHAR(20) | Default `active` |
| settings | JSONB | Default `{}` |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### messages

Project chat timeline.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| sender_id | UUID | User or agent ID |
| sender_name | VARCHAR(255) | |
| sender_type | VARCHAR(20) | `user`, `agent`, `system` |
| content | TEXT | |
| message_type | VARCHAR(50) | `user_message`, `agent_output`, `system_event`, etc. |
| metadata | JSONB | |
| created_at | TIMESTAMPTZ | |

### agents

AI agent profiles with model configuration.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| name | VARCHAR(100) | |
| role | VARCHAR(100) | |
| model | VARCHAR(100) | `gpt-4o`, `claude-sonnet-4-20250514`, etc. |
| provider_connection_id | UUID | FK → provider_connections (nullable) |
| system_prompt | TEXT | |
| temperature | DECIMAL(3,2) | Default 0.7 |
| max_tokens | INT | Default 4096 |
| status | VARCHAR(20) | `idle`, `running`, `error` |
| token_usage_total | BIGINT | Cumulative token count |
| last_heartbeat | TIMESTAMPTZ | |
| settings | JSONB | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### agent_tools

Reusable tool definitions scoped to a project.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| name | VARCHAR(100) | Unique per project |
| description | TEXT | |
| tool_type | VARCHAR(50) | `function`, `http`, `shell` |
| definition | JSONB | Schema/config for the tool |
| enabled | BOOLEAN | Default true |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### agent_tool_bindings

Many-to-many: which agents have which tools.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| agent_id | UUID | FK → agents |
| tool_id | UUID | FK → agent_tools |
| created_at | TIMESTAMPTZ | |

Unique constraint on `(agent_id, tool_id)`.

### tool_executions

Log of every tool invocation.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| agent_id | UUID | FK → agents |
| tool_id | UUID | FK → agent_tools |
| project_id | UUID | FK → projects |
| triggered_by | UUID | FK → users (nullable) |
| input | JSONB | |
| output | TEXT | |
| status | VARCHAR(20) | `success`, `error` |
| duration_ms | INT | Execution time |
| error_message | TEXT | Nullable |
| created_at | TIMESTAMPTZ | |

### usage_records

Per-invocation token and cost tracking.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| project_id | UUID | FK → projects |
| agent_id | UUID | FK → agents (nullable, SET NULL) |
| provider | VARCHAR(50) | `openai`, `anthropic` |
| model | VARCHAR(100) | |
| input_tokens | INT | |
| output_tokens | INT | |
| total_tokens | INT | |
| cost_microcents | BIGINT | 1/10000 of a cent |
| tool_calls_count | INT | |
| created_at | TIMESTAMPTZ | |

### cost_budgets

Optional monthly spend caps per project.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects (unique) |
| monthly_limit_microcents | BIGINT | 0 = unlimited |
| alert_threshold_pct | INT | Default 80 |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### agent_teams

Multi-agent team definitions.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| name | VARCHAR(255) | |
| strategy | VARCHAR(20) | `sequential`, `parallel` |
| description | TEXT | |
| leader_agent_id | UUID | FK → agents (nullable) |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### agent_team_members

Team membership.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| team_id | UUID | FK → agent_teams |
| agent_id | UUID | FK → agents |
| joined_at | TIMESTAMPTZ | |

### agent_tasks

Tasks created during team orchestration.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| team_id | UUID | FK → agent_teams |
| agent_id | UUID | FK → agents (nullable) |
| parent_task_id | UUID | Self-referential FK (nullable) |
| prompt | TEXT | |
| result | TEXT | |
| status | VARCHAR(20) | `pending`, `running`, `completed`, `failed` |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### team_memories

Project knowledge base entries.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| category | VARCHAR(100) | |
| key | VARCHAR(255) | |
| content | TEXT | |
| pinned | BOOLEAN | Default false |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

Unique constraint on `(project_id, category, key)`. GIN index on `content` for full-text search.

### project_env_vars

Encrypted environment variables per project.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| key | VARCHAR(255) | |
| value | TEXT | AES-256-GCM encrypted |
| is_secret | BOOLEAN | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

Unique constraint on `(project_id, key)`.

### vps_targets

SSH deployment target configurations.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| name | VARCHAR(255) | |
| host | VARCHAR(255) | |
| port | INT | Default 22 |
| username | VARCHAR(255) | |
| auth_type | VARCHAR(20) | `key` or `password` |
| ssh_key | TEXT | AES encrypted (nullable) |
| ssh_password | TEXT | AES encrypted (nullable) |
| deploy_path | VARCHAR(500) | |
| pre_deploy_cmd | TEXT | |
| deploy_cmd | TEXT | |
| post_deploy_cmd | TEXT | |
| status | VARCHAR(20) | `idle`, `deploying`, `success`, `failed` |
| last_deployed_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### deployment_runs

History of deployment executions.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| vps_target_id | UUID | FK → vps_targets |
| project_id | UUID | FK → projects |
| status | VARCHAR(20) | `running`, `success`, `failed` |
| triggered_by | UUID | FK → users (nullable) |
| log_output | TEXT | Full stdout/stderr |
| started_at | TIMESTAMPTZ | |
| finished_at | TIMESTAMPTZ | |

### webhooks

GitHub webhook configurations for auto-deploy.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| vps_target_id | UUID | FK → vps_targets (nullable) |
| secret | TEXT | HMAC secret, AES encrypted |
| branch | VARCHAR(255) | Default `main` |
| active | BOOLEAN | Default true |
| last_triggered_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### pipeline_stages

Ordered shell steps in a deploy pipeline.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| vps_target_id | UUID | FK → vps_targets |
| name | VARCHAR(255) | |
| command | TEXT | |
| stage_order | INT | |
| on_failure | VARCHAR(20) | `abort` or `continue` |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### notifications

In-app user notifications.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| type | VARCHAR(50) | `deploy`, `error`, `alert`, etc. |
| title | VARCHAR(255) | |
| message | TEXT | |
| read | BOOLEAN | Default false |
| data | JSONB | Extra metadata |
| created_at | TIMESTAMPTZ | |

### server_metrics

Point-in-time VPS health snapshots.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| vps_target_id | UUID | FK → vps_targets |
| cpu_percent | FLOAT | |
| mem_percent | FLOAT | |
| disk_percent | FLOAT | |
| load_avg | VARCHAR(50) | |
| uptime_seconds | BIGINT | |
| recorded_at | TIMESTAMPTZ | |

### organizations

Team organizations.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| name | VARCHAR(255) | |
| description | TEXT | |
| owner_id | UUID | FK → users |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

### organization_members

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| organization_id | UUID | FK → organizations |
| user_id | UUID | FK → users |
| role | VARCHAR(20) | `owner`, `admin`, `member` |
| joined_at | TIMESTAMPTZ | |

### organization_invitations

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| organization_id | UUID | FK → organizations |
| email | VARCHAR(255) | |
| role | VARCHAR(20) | |
| token | VARCHAR(255) | Unique invitation token |
| accepted | BOOLEAN | |
| expires_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |

### email_verification_tokens

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| token | VARCHAR(255) | Unique |
| expires_at | TIMESTAMPTZ | |
| created_at | TIMESTAMPTZ | |

### two_factor_backup_codes

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| code | VARCHAR(20) | |
| used | BOOLEAN | Default false |
| created_at | TIMESTAMPTZ | |

### terminal_sessions

Docker sandbox sessions.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects |
| container_id | VARCHAR(255) | Docker container ID |
| status | VARCHAR(20) | `running`, `stopped` |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

## Indexes

Key indexes beyond primary keys:

- `users.email` — Unique
- `projects.user_id` — FK lookup
- `messages.project_id` — Chat history queries
- `agents.project_id` — Agent listing
- `team_memories.content` — GIN index for full-text search
- `agent_tools(project_id, name)` — Unique per project
- `agent_tool_bindings(agent_id, tool_id)` — Unique binding
- `tool_executions.created_at DESC` — Execution log ordering
- `usage_records.created_at DESC` — Analytics queries
- `usage_records.project_id` — Project-scoped analytics
- `usage_records.user_id` — User-scoped analytics
- `notifications.user_id` — User notification inbox
- `server_metrics.vps_target_id` — Metric history
