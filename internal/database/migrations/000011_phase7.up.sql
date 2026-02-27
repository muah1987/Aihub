-- Phase 7: Advanced Knowledge Management
-- Documents, chunks, memory versions, shared memories

-- Knowledge documents (uploaded files)
CREATE TABLE knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(500) NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    file_type VARCHAR(50) NOT NULL,          -- 'pdf', 'markdown', 'text', 'docx', 'code'
    file_size BIGINT NOT NULL DEFAULT 0,
    content_hash VARCHAR(64),                -- SHA-256 for dedup
    chunk_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'processing', 'ready', 'error'
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_knowledge_documents_project ON knowledge_documents(project_id);
CREATE INDEX idx_knowledge_documents_status ON knowledge_documents(project_id, status);

-- Document chunks (split text for retrieval)
CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',             -- page, section, heading, etc.
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_document_chunks_document ON document_chunks(document_id, chunk_index);
CREATE INDEX idx_document_chunks_project ON document_chunks(project_id);
CREATE INDEX idx_document_chunks_fts ON document_chunks USING GIN (to_tsvector('english', content));

-- Memory versions (track changes to team_memories)
CREATE TABLE memory_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL REFERENCES team_memories(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text',
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL DEFAULT 'update',  -- 'create', 'update', 'rollback'
    diff_summary TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_versions_memory ON memory_versions(memory_id, version_number DESC);

-- Shared memories (cross-project knowledge via organizations)
CREATE TABLE shared_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    source_project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_memory_id UUID NOT NULL REFERENCES team_memories(id) ON DELETE CASCADE,
    shared_by UUID REFERENCES users(id) ON DELETE SET NULL,
    access_level VARCHAR(20) NOT NULL DEFAULT 'read',  -- 'read', 'write'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, source_memory_id)
);

CREATE INDEX idx_shared_memories_org ON shared_memories(organization_id);
CREATE INDEX idx_shared_memories_project ON shared_memories(source_project_id);

-- Auto-extracted memories from chat
CREATE TABLE auto_extractions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    extracted_content TEXT NOT NULL,
    extraction_type VARCHAR(50) NOT NULL,    -- 'decision', 'requirement', 'architecture', 'bug', 'action_item'
    confidence DECIMAL(3,2) NOT NULL DEFAULT 0.5,
    accepted BOOLEAN DEFAULT NULL,           -- null = pending review, true = accepted, false = rejected
    memory_id UUID REFERENCES team_memories(id) ON DELETE SET NULL,  -- linked if accepted
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_auto_extractions_project ON auto_extractions(project_id, created_at DESC);
CREATE INDEX idx_auto_extractions_pending ON auto_extractions(project_id) WHERE accepted IS NULL;
