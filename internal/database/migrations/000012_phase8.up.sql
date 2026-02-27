-- Phase 8: Multi-Channel Notifications & Integrations

-- Integration connections (Slack, Discord, generic webhook)
CREATE TABLE integration_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    platform VARCHAR(30) NOT NULL,         -- 'slack', 'discord', 'webhook', 'zapier'
    name VARCHAR(200) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',    -- platform-specific encrypted config
    credentials TEXT,                       -- AES-encrypted token/secret
    channel_id VARCHAR(200),               -- target channel/room
    enabled BOOLEAN DEFAULT true,
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'connected', 'error'
    error_message TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_integration_connections_project ON integration_connections(project_id);
CREATE INDEX idx_integration_connections_platform ON integration_connections(project_id, platform);

-- Notification rules (per-user filter preferences)
CREATE TABLE notification_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,  -- null = global
    event_type VARCHAR(100) NOT NULL,       -- 'deployment', 'chat', 'agent', 'webhook', 'monitoring', '*'
    channel VARCHAR(30) NOT NULL,           -- 'in_app', 'email', 'slack', 'discord'
    enabled BOOLEAN DEFAULT true,
    min_severity VARCHAR(20) DEFAULT 'info', -- 'info', 'warning', 'critical'
    quiet_hours_start SMALLINT,             -- 0-23, null = no quiet hours
    quiet_hours_end SMALLINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, project_id, event_type, channel)
);

CREATE INDEX idx_notification_rules_user ON notification_rules(user_id);
CREATE INDEX idx_notification_rules_project ON notification_rules(user_id, project_id);

-- Email digest preferences
CREATE TABLE email_digests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,  -- null = all projects
    frequency VARCHAR(20) NOT NULL DEFAULT 'daily',  -- 'daily', 'weekly', 'none'
    include_deployments BOOLEAN DEFAULT true,
    include_chat_summary BOOLEAN DEFAULT true,
    include_agent_activity BOOLEAN DEFAULT true,
    include_monitoring BOOLEAN DEFAULT true,
    last_sent_at TIMESTAMPTZ,
    next_send_at TIMESTAMPTZ,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, project_id)
);

CREATE INDEX idx_email_digests_user ON email_digests(user_id);
CREATE INDEX idx_email_digests_pending ON email_digests(next_send_at) WHERE enabled = true;

-- Outbound events (webhook dispatches to external systems)
CREATE TABLE outbound_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    integration_id UUID REFERENCES integration_connections(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'sent', 'failed', 'retrying'
    http_status INTEGER,
    response_body TEXT,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    next_retry_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_outbound_events_project ON outbound_events(project_id, created_at DESC);
CREATE INDEX idx_outbound_events_pending ON outbound_events(status, next_retry_at) WHERE status IN ('pending', 'retrying');
