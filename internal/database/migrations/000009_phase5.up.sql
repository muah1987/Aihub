-- Phase 5: Agent Tool Ecosystem + Usage Analytics

-- Agent tools: reusable tool definitions attached to agents
CREATE TABLE agent_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tool_type VARCHAR(50) NOT NULL DEFAULT 'function', -- function, http, shell
    definition JSONB NOT NULL DEFAULT '{}',              -- schema / config for the tool
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, name)
);

CREATE INDEX idx_agent_tools_project ON agent_tools(project_id);

-- Many-to-many: which agents have which tools enabled
CREATE TABLE agent_tool_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tool_id  UUID NOT NULL REFERENCES agent_tools(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, tool_id)
);

-- Tool execution log: every time a tool runs during an agent invocation
CREATE TABLE tool_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tool_id  UUID NOT NULL REFERENCES agent_tools(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    triggered_by UUID REFERENCES users(id),
    input JSONB NOT NULL DEFAULT '{}',
    output TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'success',  -- success, error
    duration_ms INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tool_executions_agent  ON tool_executions(agent_id);
CREATE INDEX idx_tool_executions_project ON tool_executions(project_id);
CREATE INDEX idx_tool_executions_created ON tool_executions(created_at DESC);

-- Usage analytics: per-invocation token + cost tracking
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    input_tokens INT NOT NULL DEFAULT 0,
    output_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    cost_microcents BIGINT NOT NULL DEFAULT 0,  -- cost in 1/10000 of a cent for precision
    tool_calls_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_records_user    ON usage_records(user_id);
CREATE INDEX idx_usage_records_project ON usage_records(project_id);
CREATE INDEX idx_usage_records_created ON usage_records(created_at DESC);

-- Cost budgets per project (optional caps)
CREATE TABLE cost_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE UNIQUE,
    monthly_limit_microcents BIGINT NOT NULL DEFAULT 0,  -- 0 = unlimited
    alert_threshold_pct INT NOT NULL DEFAULT 80,          -- alert at 80% by default
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
