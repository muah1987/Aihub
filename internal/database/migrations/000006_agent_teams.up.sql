-- Agent Teams
CREATE TABLE IF NOT EXISTS agent_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    leader_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    strategy VARCHAR(50) DEFAULT 'sequential',
    settings JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'idle',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_teams_project ON agent_teams(project_id);

-- Agent Team Memberships
CREATE TABLE IF NOT EXISTS agent_team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    role_in_team VARCHAR(100) DEFAULT 'member',
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(team_id, agent_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_team_members_team ON agent_team_members(team_id);

-- Task tracking for team orchestration
CREATE TABLE IF NOT EXISTS agent_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES agent_teams(id) ON DELETE CASCADE,
    parent_task_id UUID REFERENCES agent_tasks(id),
    assigned_agent_id UUID REFERENCES agents(id),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    prompt TEXT,
    result TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_tasks_team ON agent_tasks(team_id);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_agent ON agent_tasks(assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_parent ON agent_tasks(parent_task_id);
CREATE INDEX IF NOT EXISTS idx_agent_tasks_status ON agent_tasks(status);
