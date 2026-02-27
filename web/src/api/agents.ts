import client from './client';

export interface Agent {
  id: string;
  project_id: string;
  name: string;
  role: string;
  model: string;
  system_prompt: string;
  temperature: number;
  max_tokens: number;
  status: 'idle' | 'running' | 'error';
  token_usage_total: number;
  last_heartbeat?: string;
  created_at: string;
}

export const agentsApi = {
  list: (projectId: string) =>
    client.get<{ agents: Agent[] }>(`/projects/${projectId}/agents`),

  create: (projectId: string, data: {
    name: string;
    role: string;
    model: string;
    system_prompt?: string;
    temperature?: number;
    max_tokens?: number;
    provider_connection_id: string;
  }) => client.post<{ agent: Agent }>(`/projects/${projectId}/agents`, data),

  update: (projectId: string, agentId: string, data: Partial<Agent>) =>
    client.put<{ agent: Agent }>(`/projects/${projectId}/agents/${agentId}`, data),

  delete: (projectId: string, agentId: string) =>
    client.delete(`/projects/${projectId}/agents/${agentId}`),

  invoke: (projectId: string, agentId: string, prompt: string) =>
    client.post<{ message: unknown }>(`/projects/${projectId}/agents/${agentId}/invoke`, { prompt }),
};
