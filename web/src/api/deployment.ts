import client from './client';

export interface EnvVar {
  id: string;
  project_id: string;
  key: string;
  value: string;
  is_secret: boolean;
  created_at: string;
  updated_at: string;
}

export interface VPSTarget {
  id: string;
  project_id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  auth_type: 'key' | 'password';
  deploy_path: string;
  pre_deploy_cmd: string;
  deploy_cmd: string;
  post_deploy_cmd: string;
  status: 'idle' | 'deploying' | 'success' | 'failed';
  last_deployed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DeploymentRun {
  id: string;
  vps_target_id: string;
  project_id: string;
  status: 'running' | 'success' | 'failed';
  triggered_by?: string;
  log_output: string;
  started_at: string;
  finished_at?: string;
  created_at: string;
  target?: VPSTarget;
}

export interface TargetInput {
  name: string;
  host: string;
  port?: number;
  username: string;
  auth_type: 'key' | 'password';
  ssh_key?: string;
  ssh_password?: string;
  deploy_path: string;
  pre_deploy_cmd?: string;
  deploy_cmd: string;
  post_deploy_cmd?: string;
}

export const deploymentApi = {
  // Env vars
  listEnvVars: (projectId: string) =>
    client.get<{ env_vars: EnvVar[] }>(`/projects/${projectId}/env`),

  setEnvVar: (projectId: string, key: string, value: string, isSecret = false) =>
    client.post<{ env_var: EnvVar }>(`/projects/${projectId}/env`, {
      key,
      value,
      is_secret: isSecret,
    }),

  deleteEnvVar: (projectId: string, envId: string) =>
    client.delete(`/projects/${projectId}/env/${envId}`),

  // VPS targets
  listTargets: (projectId: string) =>
    client.get<{ targets: VPSTarget[] }>(`/projects/${projectId}/deploy`),

  createTarget: (projectId: string, input: TargetInput) =>
    client.post<{ target: VPSTarget }>(`/projects/${projectId}/deploy`, input),

  updateTarget: (projectId: string, targetId: string, input: TargetInput) =>
    client.put<{ target: VPSTarget }>(`/projects/${projectId}/deploy/${targetId}`, input),

  deleteTarget: (projectId: string, targetId: string) =>
    client.delete(`/projects/${projectId}/deploy/${targetId}`),

  // Deploy
  deploy: (projectId: string, targetId: string) =>
    client.post<{ run: DeploymentRun }>(`/projects/${projectId}/deploy/${targetId}/trigger`),

  // Runs
  listRuns: (projectId: string, targetId?: string, limit = 20) => {
    const base = targetId
      ? `/projects/${projectId}/deploy/${targetId}/runs`
      : `/projects/${projectId}/deploy/runs`;
    return client.get<{ runs: DeploymentRun[] }>(`${base}?limit=${limit}`);
  },

  getRun: (projectId: string, runId: string) =>
    client.get<{ run: DeploymentRun }>(`/projects/${projectId}/runs/${runId}`),
};
