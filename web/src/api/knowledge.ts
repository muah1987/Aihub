import client from './client';

export interface KnowledgeDocument {
  id: string;
  project_id: string;
  uploaded_by?: string;
  title: string;
  file_name: string;
  file_type: string;
  file_size: number;
  chunk_count: number;
  status: 'pending' | 'processing' | 'ready' | 'error';
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface DocumentChunk {
  id: string;
  document_id: string;
  project_id: string;
  chunk_index: number;
  content: string;
  token_count: number;
  created_at: string;
}

export interface MemoryVersion {
  id: string;
  memory_id: string;
  version_number: number;
  content: string;
  content_type: string;
  changed_by?: string;
  change_type: 'create' | 'update' | 'rollback';
  diff_summary?: string;
  created_at: string;
}

export interface SharedMemory {
  id: string;
  organization_id: string;
  source_project_id: string;
  source_memory_id: string;
  shared_by?: string;
  access_level: 'read' | 'write';
  memory?: { id: string; category: string; key: string; content: string };
  project?: { id: string; name: string };
  created_at: string;
}

export interface AutoExtraction {
  id: string;
  project_id: string;
  message_id?: string;
  extracted_content: string;
  extraction_type: 'decision' | 'requirement' | 'architecture' | 'bug' | 'action_item';
  confidence: number;
  accepted?: boolean | null;
  memory_id?: string;
  created_at: string;
}

export const knowledgeApi = {
  // Documents
  listDocuments: (projectId: string) =>
    client.get<{ documents: KnowledgeDocument[] }>(`/projects/${projectId}/knowledge/documents`),

  createDocument: (projectId: string, data: { title: string; file_name: string; file_type: string; content: string }) =>
    client.post<{ document: KnowledgeDocument }>(`/projects/${projectId}/knowledge/documents`, data),

  getDocument: (projectId: string, docId: string) =>
    client.get<{ document: KnowledgeDocument }>(`/projects/${projectId}/knowledge/documents/${docId}`),

  deleteDocument: (projectId: string, docId: string) =>
    client.delete(`/projects/${projectId}/knowledge/documents/${docId}`),

  getChunks: (projectId: string, docId: string) =>
    client.get<{ chunks: DocumentChunk[] }>(`/projects/${projectId}/knowledge/documents/${docId}/chunks`),

  searchChunks: (projectId: string, query: string, limit?: number) =>
    client.get<{ chunks: DocumentChunk[] }>(`/projects/${projectId}/knowledge/documents/search?q=${encodeURIComponent(query)}&limit=${limit || 20}`),

  // Versioning
  listVersions: (projectId: string, memoryId: string) =>
    client.get<{ versions: MemoryVersion[] }>(`/projects/${projectId}/knowledge/versions/${memoryId}`),

  getVersion: (projectId: string, memoryId: string, versionNumber: number) =>
    client.get<{ version: MemoryVersion }>(`/projects/${projectId}/knowledge/versions/${memoryId}/${versionNumber}`),

  rollback: (projectId: string, memoryId: string, versionNumber: number) =>
    client.post(`/projects/${projectId}/knowledge/versions/${memoryId}/rollback`, { version_number: versionNumber }),

  // Sharing
  shareMemory: (projectId: string, data: { organization_id: string; memory_id: string; access_level?: string }) =>
    client.post<{ shared_memory: SharedMemory }>(`/projects/${projectId}/knowledge/share`, data),

  getSharedForProject: (projectId: string) =>
    client.get<{ shared_memories: SharedMemory[] }>(`/projects/${projectId}/knowledge/shared`),

  listSharedMemories: (orgId: string) =>
    client.get<{ shared_memories: SharedMemory[] }>(`/organizations/${orgId}/shared-memories`),

  unshareMemory: (orgId: string, sharedId: string) =>
    client.delete(`/organizations/${orgId}/shared-memories/${sharedId}`),

  // Extractions
  listExtractions: (projectId: string, status?: string, limit?: number) => {
    const q = new URLSearchParams();
    if (status) q.set('status', status);
    if (limit) q.set('limit', String(limit));
    return client.get<{ extractions: AutoExtraction[] }>(`/projects/${projectId}/knowledge/extractions?${q.toString()}`);
  },

  acceptExtraction: (projectId: string, extractionId: string, saveToMemory: boolean) =>
    client.post(`/projects/${projectId}/knowledge/extractions/${extractionId}/accept`, { save_to_memory: saveToMemory }),

  rejectExtraction: (projectId: string, extractionId: string) =>
    client.post(`/projects/${projectId}/knowledge/extractions/${extractionId}/reject`),
};
