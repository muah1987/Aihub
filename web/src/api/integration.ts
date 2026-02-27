import client from './client';

export interface IntegrationConnection {
  id: string;
  project_id: string;
  created_by?: string;
  platform: 'slack' | 'discord' | 'webhook' | 'zapier';
  name: string;
  config: Record<string, string>;
  channel_id?: string;
  enabled: boolean;
  status: 'pending' | 'connected' | 'error';
  error_message?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationRule {
  id: string;
  user_id: string;
  project_id?: string;
  event_type: string;
  channel: string;
  enabled: boolean;
  min_severity: string;
  quiet_hours_start?: number;
  quiet_hours_end?: number;
  created_at: string;
  updated_at: string;
}

export interface EmailDigest {
  id: string;
  user_id: string;
  project_id?: string;
  frequency: 'daily' | 'weekly' | 'none';
  include_deployments: boolean;
  include_chat_summary: boolean;
  include_agent_activity: boolean;
  include_monitoring: boolean;
  last_sent_at?: string;
  next_send_at?: string;
  enabled: boolean;
  created_at: string;
}

export interface OutboundEvent {
  id: string;
  project_id: string;
  integration_id?: string;
  event_type: string;
  payload: Record<string, unknown>;
  status: 'pending' | 'sent' | 'failed' | 'retrying';
  http_status?: number;
  attempts: number;
  max_attempts: number;
  sent_at?: string;
  created_at: string;
}

export const integrationApi = {
  // Connections
  listConnections: (projectId: string, platform?: string) => {
    const q = platform ? `?platform=${platform}` : '';
    return client.get<{ connections: IntegrationConnection[] }>(`/projects/${projectId}/integrations${q}`);
  },

  createConnection: (projectId: string, data: { platform: string; name: string; config: Record<string, string>; channel_id?: string }) =>
    client.post<{ connection: IntegrationConnection }>(`/projects/${projectId}/integrations`, data),

  updateConnection: (projectId: string, connId: string, data: Record<string, unknown>) =>
    client.put(`/projects/${projectId}/integrations/${connId}`, data),

  deleteConnection: (projectId: string, connId: string) =>
    client.delete(`/projects/${projectId}/integrations/${connId}`),

  testConnection: (projectId: string, connId: string) =>
    client.post<{ result: string; error?: string }>(`/projects/${projectId}/integrations/${connId}/test`),

  // Events
  listEvents: (projectId: string, limit?: number) =>
    client.get<{ events: OutboundEvent[] }>(`/projects/${projectId}/integrations/events?limit=${limit || 50}`),

  // Notification Rules
  listRules: (projectId?: string) => {
    const q = projectId ? `?project_id=${projectId}` : '';
    return client.get<{ rules: NotificationRule[] }>(`/notification-rules${q}`);
  },

  createRule: (data: { project_id?: string; event_type: string; channel: string; min_severity?: string }) =>
    client.post<{ rule: NotificationRule }>('/notification-rules', data),

  updateRule: (ruleId: string, data: Record<string, unknown>) =>
    client.put(`/notification-rules/${ruleId}`, data),

  deleteRule: (ruleId: string) =>
    client.delete(`/notification-rules/${ruleId}`),

  // Email Digests
  listDigests: () =>
    client.get<{ digests: EmailDigest[] }>('/email-digests'),

  upsertDigest: (data: {
    project_id?: string;
    input: { frequency: string; include_deployments: boolean; include_chat_summary: boolean; include_agent_activity: boolean; include_monitoring: boolean };
  }) => client.post<{ digest: EmailDigest }>('/email-digests', data),

  getDigest: (projectId?: string) => {
    const q = projectId ? `?project_id=${projectId}` : '';
    return client.get<{ digest: EmailDigest }>(`/email-digests/current${q}`);
  },

  deleteDigest: (digestId: string) =>
    client.delete(`/email-digests/${digestId}`),
};
