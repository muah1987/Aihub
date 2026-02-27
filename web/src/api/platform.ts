import client from './client';

export interface UserPreference {
  id: string;
  user_id: string;
  theme: 'light' | 'dark' | 'system';
  accent_color: string;
  sidebar_collapsed: boolean;
  compact_mode: boolean;
  editor_font_size: number;
  notifications_sound: boolean;
  locale: string;
  timezone: string;
  settings: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface PinnedProject {
  id: string;
  user_id: string;
  project_id: string;
  pin_order: number;
  created_at: string;
  project?: {
    id: string;
    name: string;
    repo_owner: string;
    repo_name: string;
    status: string;
  };
}

export interface SearchResult {
  type: 'message' | 'memory' | 'agent' | 'document' | 'workflow';
  id: string;
  project_id: string;
  title: string;
  snippet: string;
  score: number;
}

export interface PromptVersion {
  id: string;
  agent_id: string;
  version_label: string;
  system_prompt: string;
  model_config: Record<string, unknown>;
  is_active: boolean;
  total_invocations: number;
  avg_rating: number;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export const platformApi = {
  // Preferences
  getPreferences: () =>
    client.get<{ preferences: UserPreference }>('/preferences'),

  updatePreferences: (data: Partial<Pick<UserPreference, 'theme' | 'accent_color' | 'sidebar_collapsed' | 'compact_mode' | 'editor_font_size' | 'notifications_sound' | 'locale' | 'timezone'>>) =>
    client.put<{ preferences: UserPreference }>('/preferences', data),

  // Pinned projects
  listPinned: () =>
    client.get<{ pinned: PinnedProject[] }>('/pinned'),

  pinProject: (projectId: string) =>
    client.post<{ pinned: PinnedProject }>(`/projects/${projectId}/pin`),

  unpinProject: (projectId: string) =>
    client.delete(`/projects/${projectId}/pin`),

  reorderPins: (projectIds: string[]) =>
    client.post('/pinned/reorder', { project_ids: projectIds }),

  // Global search
  search: (query: string) =>
    client.get<{ results: SearchResult[] }>(`/search?q=${encodeURIComponent(query)}`),

  // Prompt versions
  listPromptVersions: (projectId: string, agentId: string) =>
    client.get<{ versions: PromptVersion[] }>(`/projects/${projectId}/agents/${agentId}/prompts`),

  createPromptVersion: (projectId: string, agentId: string, data: { label: string; system_prompt: string }) =>
    client.post<{ version: PromptVersion }>(`/projects/${projectId}/agents/${agentId}/prompts`, data),

  activatePromptVersion: (projectId: string, agentId: string, versionId: string) =>
    client.put(`/projects/${projectId}/agents/${agentId}/prompts/${versionId}/activate`),

  deletePromptVersion: (projectId: string, agentId: string, versionId: string) =>
    client.delete(`/projects/${projectId}/agents/${agentId}/prompts/${versionId}`),
};
