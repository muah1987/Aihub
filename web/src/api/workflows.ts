import client from './client';

export interface Workflow {
  id: string;
  project_id: string;
  name: string;
  description: string;
  trigger_type: 'manual' | 'webhook' | 'schedule' | 'deploy_success' | 'deploy_failure';
  trigger_config: Record<string, unknown>;
  enabled: boolean;
  steps: WorkflowStep[];
  created_at: string;
  updated_at: string;
}

export interface WorkflowStep {
  id: string;
  workflow_id: string;
  name: string;
  step_type: 'invoke_agent' | 'invoke_team' | 'run_tool' | 'http_request' | 'shell_command' | 'deploy' | 'notify' | 'condition';
  step_order: number;
  config: Record<string, unknown>;
  on_failure: 'abort' | 'continue';
  timeout_seconds: number;
  created_at: string;
}

export interface WorkflowRun {
  id: string;
  workflow_id: string;
  triggered_by: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled';
  started_at: string;
  finished_at?: string;
  step_runs: WorkflowStepRun[];
  created_at: string;
}

export interface WorkflowStepRun {
  id: string;
  run_id: string;
  step_id: string;
  step_name: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped';
  output: string;
  error_message: string;
  started_at?: string;
  finished_at?: string;
}

export const workflowsApi = {
  list: (projectId: string) =>
    client.get<{ workflows: Workflow[] }>(`/projects/${projectId}/workflows`),

  get: (projectId: string, workflowId: string) =>
    client.get<{ workflow: Workflow }>(`/projects/${projectId}/workflows/${workflowId}`),

  create: (projectId: string, data: { name: string; description?: string; trigger_type: string; trigger_config?: unknown }) =>
    client.post<{ workflow: Workflow }>(`/projects/${projectId}/workflows`, data),

  update: (projectId: string, workflowId: string, data: Partial<Workflow>) =>
    client.put<{ workflow: Workflow }>(`/projects/${projectId}/workflows/${workflowId}`, data),

  delete: (projectId: string, workflowId: string) =>
    client.delete(`/projects/${projectId}/workflows/${workflowId}`),

  setEnabled: (projectId: string, workflowId: string, enabled: boolean) =>
    client.put(`/projects/${projectId}/workflows/${workflowId}/enabled`, { enabled }),

  trigger: (projectId: string, workflowId: string) =>
    client.post<{ run: WorkflowRun }>(`/projects/${projectId}/workflows/${workflowId}/trigger`),

  listRuns: (projectId: string, workflowId: string) =>
    client.get<{ runs: WorkflowRun[] }>(`/projects/${projectId}/workflows/${workflowId}/runs`),

  getRun: (projectId: string, workflowId: string, runId: string) =>
    client.get<{ run: WorkflowRun }>(`/projects/${projectId}/workflows/${workflowId}/runs/${runId}`),

  addStep: (projectId: string, workflowId: string, data: { name: string; step_type: string; step_order: number; config?: unknown; on_failure?: string; timeout_seconds?: number }) =>
    client.post<{ step: WorkflowStep }>(`/projects/${projectId}/workflows/${workflowId}/steps`, data),

  deleteStep: (projectId: string, workflowId: string, stepId: string) =>
    client.delete(`/projects/${projectId}/workflows/${workflowId}/steps/${stepId}`),
};
