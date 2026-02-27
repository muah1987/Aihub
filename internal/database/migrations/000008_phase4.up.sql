-- GitHub webhooks for auto-deploy on push
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    vps_target_id UUID REFERENCES vps_targets(id) ON DELETE SET NULL,
    secret TEXT NOT NULL,
    branch VARCHAR(255) DEFAULT 'main',
    active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Pipeline stages per VPS target (ordered steps before/after deploy)
CREATE TABLE pipeline_stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vps_target_id UUID NOT NULL REFERENCES vps_targets(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    command TEXT NOT NULL,
    stage_order INTEGER NOT NULL DEFAULT 0,
    on_failure VARCHAR(20) DEFAULT 'abort',  -- 'abort' | 'continue'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- In-app notification inbox
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT false,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- VPS server metrics snapshots (polled via SSH)
CREATE TABLE server_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vps_target_id UUID NOT NULL REFERENCES vps_targets(id) ON DELETE CASCADE,
    cpu_percent REAL DEFAULT 0,
    mem_percent REAL DEFAULT 0,
    disk_percent REAL DEFAULT 0,
    load_avg VARCHAR(50) DEFAULT '',
    uptime_seconds BIGINT DEFAULT 0,
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX ON webhooks(project_id);
CREATE INDEX ON pipeline_stages(vps_target_id);
CREATE INDEX ON pipeline_stages(stage_order);
CREATE INDEX ON notifications(user_id, read);
CREATE INDEX ON notifications(created_at DESC);
CREATE INDEX ON server_metrics(vps_target_id, recorded_at DESC);
