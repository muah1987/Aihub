import client from './client';

export interface Message {
  id: string;
  project_id: string;
  sender_type: 'human' | 'agent' | 'system' | 'build' | 'repo';
  sender_id?: string;
  sender_name: string;
  content: string;
  message_type: string;
  metadata: Record<string, unknown>;
  created_at: string;
}

export const chatApi = {
  getHistory: (projectId: string, before?: string, limit = 50) =>
    client.get<{ messages: Message[] }>(`/projects/${projectId}/messages`, {
      params: { before, limit },
    }),

  sendMessage: (projectId: string, content: string, messageType = 'text') =>
    client.post<{ message: Message }>(`/projects/${projectId}/messages`, {
      content,
      message_type: messageType,
    }),
};
