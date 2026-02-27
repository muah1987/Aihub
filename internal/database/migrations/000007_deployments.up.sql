-- Project environment variables (secrets stored AES-encrypted)
CREATE TABLE project_env_vars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    is_secret BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, key)
);

-- VPS deployment targets per project
CREATE TABLE vps_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 22,
    username VARCHAR(255) NOT NULL,
    auth_type VARCHAR(20) DEFAULT 'key',
    ssh_key TEXT,
    ssh_password TEXT,
    deploy_path VARCHAR(500) NOT NULL DEFAULT '/app',
    pre_deploy_cmd TEXT,
    deploy_cmd TEXT NOT NULL DEFAULT 'git pull && npm install && pm2 restart all',
    post_deploy_cmd TEXT,
    status VARCHAR(20) DEFAULT 'idle',
    last_deployed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Deployment run history
CREATE TABLE deployment_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vps_target_id UUID NOT NULL REFERENCES vps_targets(id) ON DELETE CASCADE,
    project_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'running',
    triggered_by UUID REFERENCES users(id),
    log_output TEXT DEFAULT '',
    started_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX ON project_env_vars(project_id);
CREATE INDEX ON vps_targets(project_id);
CREATE INDEX ON deployment_runs(vps_target_id);
CREATE INDEX ON deployment_runs(project_id);
