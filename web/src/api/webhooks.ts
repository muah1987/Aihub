import client from './client';

export interface Webhook {
  id: string;
  project_id: string;
  vps_target_id?: string;
  branch: string;
  active: boolean;
  last_triggered_at?: string;
  created_at: string;
}

export const webhooksApi = {
  list: (projectId: string) =>
    client.get<{ webhooks: Webhook[] }>(`/projects/${projectId}/webhooks`),

  create: (projectId: string, vpsTargetId: string | null, branch: string) =>
    client.post<{ webhook: Webhook; secret: string; note: string }>(
      `/projects/${projectId}/webhooks`,
      { vps_target_id: vpsTargetId, branch },
    ),

  setActive: (projectId: string, webhookId: string, active: boolean) =>
    client.put(`/projects/${projectId}/webhooks/${webhookId}/active`, { active }),

  delete: (projectId: string, webhookId: string) =>
    client.delete(`/projects/${projectId}/webhooks/${webhookId}`),

  getSecret: (projectId: string, webhookId: string) =>
    client.get<{ secret: string }>(`/projects/${projectId}/webhooks/${webhookId}/secret`),
};

export const monitoringApi = {
  collect: (projectId: string, targetId: string) =>
    client.post<{ metric: ServerMetric }>(`/projects/${projectId}/deploy/${targetId}/metrics/collect`),

  latest: (projectId: string, targetId: string) =>
    client.get<{ metric: ServerMetric }>(`/projects/${projectId}/deploy/${targetId}/metrics/latest`),

  history: (projectId: string, targetId: string, limit = 20) =>
    client.get<{ metrics: ServerMetric[] }>(`/projects/${projectId}/deploy/${targetId}/metrics/history?limit=${limit}`),

  listStages: (projectId: string, targetId: string) =>
    client.get<{ stages: PipelineStage[] }>(`/projects/${projectId}/deploy/${targetId}/stages`),

  createStage: (projectId: string, targetId: string, data: { name: string; command: string; stage_order: number; on_failure?: string }) =>
    client.post<{ stage: PipelineStage }>(`/projects/${projectId}/deploy/${targetId}/stages`, data),

  deleteStage: (projectId: string, targetId: string, stageId: string) =>
    client.delete(`/projects/${projectId}/deploy/${targetId}/stages/${stageId}`),
};

export interface ServerMetric {
  id: string;
  vps_target_id: string;
  cpu_percent: number;
  mem_percent: number;
  disk_percent: number;
  load_avg: string;
  uptime_seconds: number;
  recorded_at: string;
}

export interface PipelineStage {
  id: string;
  vps_target_id: string;
  name: string;
  command: string;
  stage_order: number;
  on_failure: 'abort' | 'continue';
  created_at: string;
}
