import client from './client';

export interface AuditLog {
  id: string;
  organization_id?: string;
  project_id?: string;
  user_id?: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  details: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
  severity: 'info' | 'warning' | 'critical';
  created_at: string;
}

export interface DataExport {
  id: string;
  project_id: string;
  requested_by?: string;
  export_type: string;
  format: string;
  status: 'pending' | 'processing' | 'ready' | 'expired' | 'error';
  file_size: number;
  include_chat: boolean;
  include_memories: boolean;
  include_agents: boolean;
  include_workflows: boolean;
  include_settings: boolean;
  error_message?: string;
  expires_at?: string;
  created_at: string;
  completed_at?: string;
}

export interface CustomRole {
  id: string;
  organization_id: string;
  name: string;
  description: string;
  permissions: string[];
  is_system: boolean;
  created_at: string;
}

export interface ResourcePermission {
  id: string;
  organization_id: string;
  user_id: string;
  resource_type: string;
  resource_id: string;
  permission: string;
  granted_by?: string;
  created_at: string;
}

export interface RetentionPolicy {
  id: string;
  organization_id?: string;
  project_id?: string;
  resource_type: string;
  retention_days: number;
  enabled: boolean;
  last_run_at?: string;
  records_deleted: number;
  created_at: string;
}

export const complianceApi = {
  // Audit logs
  listAuditLogs: (projectId: string, params?: { action?: string; resource_type?: string; severity?: string; limit?: number; offset?: number }) => {
    const q = new URLSearchParams();
    if (params?.action) q.set('action', params.action);
    if (params?.resource_type) q.set('resource_type', params.resource_type);
    if (params?.severity) q.set('severity', params.severity);
    if (params?.limit) q.set('limit', String(params.limit));
    if (params?.offset) q.set('offset', String(params.offset));
    return client.get<{ audit_logs: AuditLog[]; total: number }>(`/projects/${projectId}/audit-logs?${q.toString()}`);
  },

  // Data exports
  listExports: (projectId: string) =>
    client.get<{ exports: DataExport[] }>(`/projects/${projectId}/exports`),

  createExport: (projectId: string, data: {
    export_type?: string; format?: string;
    include_chat?: boolean; include_memories?: boolean; include_agents?: boolean;
    include_workflows?: boolean; include_settings?: boolean;
  }) => client.post<{ export: DataExport }>(`/projects/${projectId}/exports`, data),

  downloadExport: (projectId: string, exportId: string) =>
    client.get<Blob>(`/projects/${projectId}/exports/${exportId}/download`, { responseType: 'blob' as 'json' }),

  deleteExport: (projectId: string, exportId: string) =>
    client.delete(`/projects/${projectId}/exports/${exportId}`),

  // Custom roles (org-scoped)
  listRoles: (orgId: string) =>
    client.get<{ roles: CustomRole[] }>(`/organizations/${orgId}/roles`),

  createRole: (orgId: string, data: { name: string; description?: string; permissions: string[] }) =>
    client.post<{ role: CustomRole }>(`/organizations/${orgId}/roles`, data),

  updateRole: (orgId: string, roleId: string, data: { name: string; description?: string; permissions: string[] }) =>
    client.put(`/organizations/${orgId}/roles/${roleId}`, data),

  deleteRole: (orgId: string, roleId: string) =>
    client.delete(`/organizations/${orgId}/roles/${roleId}`),

  // Retention policies
  listRetentionPolicies: (projectId: string) =>
    client.get<{ policies: RetentionPolicy[] }>(`/projects/${projectId}/retention`),

  upsertRetentionPolicy: (projectId: string, data: { resource_type: string; retention_days: number; enabled: boolean }) =>
    client.post<{ policy: RetentionPolicy }>(`/projects/${projectId}/retention`, data),

  deleteRetentionPolicy: (projectId: string, policyId: string) =>
    client.delete(`/projects/${projectId}/retention/${policyId}`),
};
