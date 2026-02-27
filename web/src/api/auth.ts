import client from './client';

export interface User {
  id: string;
  email: string;
  display_name: string;
  email_verified: boolean;
  role: string;
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
}

export const authApi = {
  register: (email: string, password: string, display_name: string) =>
    client.post<AuthResponse>('/auth/register', { email, password, display_name }),

  login: (email: string, password: string) =>
    client.post<AuthResponse>('/auth/login', { email, password }),

  refresh: (refresh_token: string) =>
    client.post<AuthResponse>('/auth/refresh', { refresh_token }),

  me: () => client.get<{ user: User }>('/auth/me'),

  logout: () => client.post('/auth/logout'),
};
