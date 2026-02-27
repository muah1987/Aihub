import client from './client';

export interface AgentTool {
  id: string;
  project_id: string;
  name: string;
  description: string;
  tool_type: 'function' | 'http' | 'shell';
  definition: Record<string, unknown>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface ToolBinding {
  id: string;
  agent_id: string;
  tool_id: string;
  tool: AgentTool;
  created_at: string;
}

export interface ToolExecution {
  id: string;
  agent_id: string;
  tool_id: string;
  project_id: string;
  triggered_by?: string;
  input: Record<string, unknown>;
  output: string;
  status: 'success' | 'error';
  duration_ms: number;
  error_message?: string;
  created_at: string;
}

export const toolsApi = {
  list: (projectId: string) =>
    client.get<{ tools: AgentTool[] }>(`/projects/${projectId}/tools`),

  create: (projectId: string, data: { name: string; description: string; tool_type: string; definition: unknown }) =>
    client.post<{ tool: AgentTool }>(`/projects/${projectId}/tools`, data),

  update: (projectId: string, toolId: string, data: Partial<AgentTool>) =>
    client.put<{ tool: AgentTool }>(`/projects/${projectId}/tools/${toolId}`, data),

  delete: (projectId: string, toolId: string) =>
    client.delete(`/projects/${projectId}/tools/${toolId}`),

  listExecutions: (projectId: string, limit?: number) =>
    client.get<{ executions: ToolExecution[] }>(`/projects/${projectId}/tools/executions?limit=${limit || 50}`),

  // Agent bindings
  listBindings: (projectId: string, agentId: string) =>
    client.get<{ bindings: ToolBinding[] }>(`/projects/${projectId}/agents/${agentId}/tools`),

  bind: (projectId: string, agentId: string, toolId: string) =>
    client.post(`/projects/${projectId}/agents/${agentId}/tools`, { tool_id: toolId }),

  unbind: (projectId: string, agentId: string, toolId: string) =>
    client.delete(`/projects/${projectId}/agents/${agentId}/tools/${toolId}`),

  execute: (projectId: string, agentId: string, toolId: string, input: unknown) =>
    client.post<{ execution: ToolExecution }>(`/projects/${projectId}/agents/${agentId}/tools/${toolId}/execute`, { input }),
};
