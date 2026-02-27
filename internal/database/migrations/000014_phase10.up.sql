-- Phase 10: Platform Polish & Scale

-- User preferences (theme, UI settings)
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) NOT NULL DEFAULT 'system',        -- 'light', 'dark', 'system'
    accent_color VARCHAR(20) DEFAULT 'blue',
    sidebar_collapsed BOOLEAN DEFAULT false,
    compact_mode BOOLEAN DEFAULT false,
    editor_font_size INTEGER DEFAULT 14,
    notifications_sound BOOLEAN DEFAULT true,
    locale VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    settings JSONB DEFAULT '{}',                         -- extensible key/value
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id)
);

-- UNIQUE(user_id) already creates an implicit index; no separate index needed.

-- Pinned/favorite projects
CREATE TABLE pinned_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    pin_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, project_id)
);

CREATE INDEX idx_pinned_projects_user ON pinned_projects(user_id, pin_order);

-- Agent prompt versions (A/B testing)
CREATE TABLE prompt_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    version_label VARCHAR(100) NOT NULL,
    system_prompt TEXT NOT NULL,
    model_config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT false,
    total_invocations INTEGER DEFAULT 0,
    avg_rating DECIMAL(3,2) DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_prompt_versions_agent ON prompt_versions(agent_id);
CREATE INDEX idx_prompt_versions_active ON prompt_versions(agent_id) WHERE is_active = true;
