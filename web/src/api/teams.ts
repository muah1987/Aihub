import client from './client';

export interface AgentTeam {
  id: string;
  project_id: string;
  name: string;
  description: string;
  leader_agent_id?: string;
  strategy: 'sequential' | 'parallel' | 'adaptive';
  status: string;
  created_at: string;
}

export interface TeamMember {
  id: string;
  team_id: string;
  agent_id: string;
  role_in_team: string;
  priority: number;
  agent?: {
    id: string;
    name: string;
    role: string;
    model: string;
    status: string;
  };
}

export interface AgentTask {
  id: string;
  team_id: string;
  title: string;
  description: string;
  assigned_agent_id?: string;
  status: string;
  result?: string;
  created_at: string;
  started_at?: string;
  completed_at?: string;
  assigned_agent?: {
    id: string;
    name: string;
    role: string;
  };
}

export const teamsApi = {
  list: (projectId: string) =>
    client.get<{ teams: AgentTeam[] }>(`/projects/${projectId}/teams`),

  create: (projectId: string, data: { name: string; description?: string; strategy?: string }) =>
    client.post<{ team: AgentTeam }>(`/projects/${projectId}/teams`, data),

  get: (projectId: string, teamId: string) =>
    client.get<{ team: AgentTeam }>(`/projects/${projectId}/teams/${teamId}`),

  update: (projectId: string, teamId: string, data: { name?: string; description?: string; strategy?: string }) =>
    client.put<{ team: AgentTeam }>(`/projects/${projectId}/teams/${teamId}`, data),

  delete: (projectId: string, teamId: string) =>
    client.delete(`/projects/${projectId}/teams/${teamId}`),

  invoke: (projectId: string, teamId: string, prompt: string) =>
    client.post<{ message: unknown; tasks: AgentTask[] }>(`/projects/${projectId}/teams/${teamId}/invoke`, { prompt }),

  listMembers: (projectId: string, teamId: string) =>
    client.get<{ members: TeamMember[] }>(`/projects/${projectId}/teams/${teamId}/members`),

  addMember: (projectId: string, teamId: string, agentId: string, roleInTeam?: string) =>
    client.post(`/projects/${projectId}/teams/${teamId}/members`, { agent_id: agentId, role_in_team: roleInTeam || 'member' }),

  removeMember: (projectId: string, teamId: string, agentId: string) =>
    client.delete(`/projects/${projectId}/teams/${teamId}/members/${agentId}`),

  setLeader: (projectId: string, teamId: string, agentId: string) =>
    client.put(`/projects/${projectId}/teams/${teamId}/leader`, { agent_id: agentId }),

  listTasks: (projectId: string, teamId: string) =>
    client.get<{ tasks: AgentTask[] }>(`/projects/${projectId}/teams/${teamId}/tasks`),
};
