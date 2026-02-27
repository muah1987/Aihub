-- Team memory entries (project-scoped knowledge base)
CREATE TABLE IF NOT EXISTS team_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL DEFAULT 'general',
    key VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text',
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_by_agent UUID REFERENCES agents(id),
    pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, category, key)
);

CREATE INDEX IF NOT EXISTS idx_team_memories_project ON team_memories(project_id);
CREATE INDEX IF NOT EXISTS idx_team_memories_project_category ON team_memories(project_id, category);
CREATE INDEX IF NOT EXISTS idx_team_memories_search ON team_memories USING GIN (to_tsvector('english', content));
