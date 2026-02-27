import client from './client';

export interface TeamMemory {
  id: string;
  project_id: string;
  category: string;
  key: string;
  content: string;
  content_type: string;
  pinned: boolean;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export const memoryApi = {
  list: (projectId: string, category?: string) => {
    const params = category ? `?category=${encodeURIComponent(category)}` : '';
    return client.get<{ memories: TeamMemory[] }>(`/projects/${projectId}/memory${params}`);
  },

  set: (projectId: string, data: { category?: string; key: string; content: string; content_type?: string; pinned?: boolean }) =>
    client.post<{ memory: TeamMemory }>(`/projects/${projectId}/memory`, data),

  search: (projectId: string, query: string) =>
    client.get<{ memories: TeamMemory[] }>(`/projects/${projectId}/memory/search?q=${encodeURIComponent(query)}`),

  get: (projectId: string, category: string, key: string) =>
    client.get<{ memory: TeamMemory }>(`/projects/${projectId}/memory/${encodeURIComponent(category)}/${encodeURIComponent(key)}`),

  delete: (projectId: string, memoryId: string) =>
    client.delete(`/projects/${projectId}/memory/${memoryId}`),
};
