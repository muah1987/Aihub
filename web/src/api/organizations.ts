import client from './client';

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  owner_id: string;
  created_at: string;
}

export interface OrgMember {
  id: string;
  organization_id: string;
  user_id: string;
  role: string;
  joined_at: string;
  user?: {
    id: string;
    email: string;
    display_name: string;
  };
}

export const organizationsApi = {
  list: () => client.get<{ organizations: Organization[] }>('/organizations'),

  create: (data: { name: string; slug: string; description?: string }) =>
    client.post<{ organization: Organization }>('/organizations', data),

  get: (orgId: string) =>
    client.get<{ organization: Organization }>(`/organizations/${orgId}`),

  update: (orgId: string, data: { name?: string; description?: string }) =>
    client.put<{ organization: Organization }>(`/organizations/${orgId}`, data),

  delete: (orgId: string) =>
    client.delete(`/organizations/${orgId}`),

  listMembers: (orgId: string) =>
    client.get<{ members: OrgMember[] }>(`/organizations/${orgId}/members`),

  inviteMember: (orgId: string, email: string, role: string) =>
    client.post(`/organizations/${orgId}/invite`, { email, role }),

  removeMember: (orgId: string, userId: string) =>
    client.delete(`/organizations/${orgId}/members/${userId}`),

  updateMemberRole: (orgId: string, userId: string, role: string) =>
    client.put(`/organizations/${orgId}/members/${userId}/role`, { role }),

  acceptInvitation: (token: string) =>
    client.post(`/invitations/${token}/accept`),
};
