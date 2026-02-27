import client from './client';

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  message: string;
  read: boolean;
  data?: Record<string, unknown>;
  created_at: string;
}

export const notificationsApi = {
  list: (unreadOnly = false, limit = 30) =>
    client.get<{ notifications: Notification[]; unread_count: number }>(
      `/notifications?unread=${unreadOnly}&limit=${limit}`,
    ),

  markRead: (id: string) =>
    client.post(`/notifications/${id}/read`),

  markAllRead: () =>
    client.post('/notifications/read-all'),

  delete: (id: string) =>
    client.delete(`/notifications/${id}`),
};
