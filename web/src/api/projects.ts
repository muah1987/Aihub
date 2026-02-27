import client from './client';

export interface Project {
  id: string;
  user_id: string;
  name: string;
  description: string;
  repo_provider: string;
  repo_owner: string;
  repo_name: string;
  repo_url: string;
  repo_default_branch: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface GitHubRepo {
  id: number;
  name: string;
  full_name: string;
  description: string;
  html_url: string;
  clone_url: string;
  default_branch: string;
  private: boolean;
  language: string;
}

export const projectsApi = {
  list: () => client.get<{ projects: Project[] }>('/projects'),

  create: (data: {
    name: string;
    description?: string;
    repo_owner: string;
    repo_name: string;
    provider_connection_id: string;
  }) => client.post<{ project: Project }>('/projects', data),

  get: (id: string) => client.get<{ project: Project }>(`/projects/${id}`),

  update: (id: string, data: { name?: string; description?: string }) =>
    client.put<{ project: Project }>(`/projects/${id}`, data),

  delete: (id: string) => client.delete(`/projects/${id}`),

  listRepos: (providerConnectionId: string, page = 1) =>
    client.get<{ repos: GitHubRepo[] }>('/projects/repos', {
      params: { provider_connection_id: providerConnectionId, page },
    }),
};
