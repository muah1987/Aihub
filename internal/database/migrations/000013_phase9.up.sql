-- Phase 9: Enterprise & Compliance

-- Audit logs (comprehensive action tracking)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,            -- 'project.create', 'deployment.trigger', 'member.invite', etc.
    resource_type VARCHAR(50) NOT NULL,      -- 'project', 'deployment', 'agent', 'user', 'org', etc.
    resource_id UUID,
    details JSONB DEFAULT '{}',              -- action-specific data (old/new values, params)
    ip_address VARCHAR(45),
    user_agent TEXT,
    severity VARCHAR(20) DEFAULT 'info',     -- 'info', 'warning', 'critical'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_org ON audit_logs(organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_project ON audit_logs(project_id, created_at DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- Data exports (project backup jobs)
CREATE TABLE data_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    export_type VARCHAR(30) NOT NULL DEFAULT 'full',   -- 'full', 'partial', 'memories', 'chat'
    format VARCHAR(20) NOT NULL DEFAULT 'json',        -- 'json', 'csv'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',     -- 'pending', 'processing', 'ready', 'expired', 'error'
    file_path TEXT,
    file_size BIGINT DEFAULT 0,
    include_chat BOOLEAN DEFAULT true,
    include_memories BOOLEAN DEFAULT true,
    include_agents BOOLEAN DEFAULT true,
    include_workflows BOOLEAN DEFAULT true,
    include_settings BOOLEAN DEFAULT true,
    error_message TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_data_exports_project ON data_exports(project_id, created_at DESC);
CREATE INDEX idx_data_exports_status ON data_exports(status) WHERE status = 'ready';

-- Custom roles (beyond built-in owner/admin/member/viewer)
CREATE TABLE custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    permissions JSONB NOT NULL DEFAULT '[]', -- array of permission strings
    is_system BOOLEAN DEFAULT false,         -- true for built-in roles
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, name)
);

CREATE INDEX idx_custom_roles_org ON custom_roles(organization_id);

-- Resource-level permission overrides
CREATE TABLE resource_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,      -- 'project', 'agent', 'deployment', 'workflow'
    resource_id UUID NOT NULL,
    permission VARCHAR(50) NOT NULL,         -- 'read', 'write', 'admin', 'execute', 'delete'
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, resource_type, resource_id, permission)
);

CREATE INDEX idx_resource_permissions_user ON resource_permissions(user_id, resource_type);
CREATE INDEX idx_resource_permissions_resource ON resource_permissions(resource_type, resource_id);

-- Data retention policies (automatic cleanup)
CREATE TABLE retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,      -- 'audit_logs', 'messages', 'notifications', 'outbound_events', 'server_metrics'
    retention_days INTEGER NOT NULL DEFAULT 90,
    enabled BOOLEAN DEFAULT true,
    last_run_at TIMESTAMPTZ,
    records_deleted INTEGER DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(COALESCE(organization_id, '00000000-0000-0000-0000-000000000000'), COALESCE(project_id, '00000000-0000-0000-0000-000000000000'), resource_type)
);

CREATE INDEX idx_retention_policies_org ON retention_policies(organization_id);
CREATE INDEX idx_retention_policies_project ON retention_policies(project_id);
