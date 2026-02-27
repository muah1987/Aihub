import client from './client';

export interface ActivityLog {
  id: string;
  project_id: string;
  user_id: string;
  action: string;
  resource_type: string;
  resource_id: string;
  details: Record<string, unknown>;
  created_at: string;
}

export const activityApi = {
  list: (projectId: string, params?: { limit?: number; offset?: number; action?: string; resource_type?: string }) => {
    const q = new URLSearchParams();
    if (params?.limit) q.set('limit', String(params.limit));
    if (params?.offset) q.set('offset', String(params.offset));
    if (params?.action) q.set('action', params.action);
    if (params?.resource_type) q.set('resource_type', params.resource_type);
    return client.get<{ activity: ActivityLog[]; total?: number }>(`/projects/${projectId}/activity?${q.toString()}`);
  },
};
