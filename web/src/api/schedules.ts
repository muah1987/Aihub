import client from './client';

export interface ScheduledJob {
  id: string;
  project_id: string;
  name: string;
  description: string;
  cron_expression: string;
  action_type: string;
  action_config: Record<string, unknown>;
  enabled: boolean;
  last_run_at?: string;
  last_run_status?: string;
  next_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ScheduledJobRun {
  id: string;
  job_id: string;
  status: 'running' | 'completed' | 'failed';
  output: string;
  error_message: string;
  started_at: string;
  finished_at?: string;
}

export const schedulesApi = {
  list: (projectId: string) =>
    client.get<{ jobs: ScheduledJob[] }>(`/projects/${projectId}/schedules`),

  create: (projectId: string, data: { name: string; description?: string; cron_expression: string; action_type: string; action_config?: unknown }) =>
    client.post<{ job: ScheduledJob }>(`/projects/${projectId}/schedules`, data),

  update: (projectId: string, jobId: string, data: Partial<ScheduledJob>) =>
    client.put<{ job: ScheduledJob }>(`/projects/${projectId}/schedules/${jobId}`, data),

  delete: (projectId: string, jobId: string) =>
    client.delete(`/projects/${projectId}/schedules/${jobId}`),

  setEnabled: (projectId: string, jobId: string, enabled: boolean) =>
    client.put(`/projects/${projectId}/schedules/${jobId}/enabled`, { enabled }),

  listRuns: (projectId: string, jobId: string) =>
    client.get<{ runs: ScheduledJobRun[] }>(`/projects/${projectId}/schedules/${jobId}/runs`),
};
