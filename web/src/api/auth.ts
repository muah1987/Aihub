import client from './client';

export interface User {
  id: string;
  email: string;
  display_name: string;
  email_verified: boolean;
  two_factor_enabled?: boolean;
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

export interface LoginResponse {
  user?: User;
  tokens?: TokenPair;
  requires_2fa?: boolean;
  pending_token?: string;
}

export interface TwoFASetup {
  secret: string;
  qr_url: string;
  backup_codes: string[];
}

export const authApi = {
  register: (email: string, password: string, display_name: string) =>
    client.post<AuthResponse>('/auth/register', { email, password, display_name }),

  login: (email: string, password: string) =>
    client.post<LoginResponse>('/auth/login', { email, password }),

  refresh: (refresh_token: string) =>
    client.post<AuthResponse>('/auth/refresh', { refresh_token }),

  me: () => client.get<{ user: User }>('/auth/me'),

  logout: () => client.post('/auth/logout'),

  // Email verification
  verifyEmail: (token: string) =>
    client.post('/auth/verify-email', { token }),

  resendVerification: () =>
    client.post('/auth/resend-verification'),

  // 2FA
  setup2FA: () =>
    client.post<TwoFASetup>('/auth/2fa/setup'),

  confirm2FA: (code: string) =>
    client.post('/auth/2fa/confirm', { code }),

  disable2FA: (code: string) =>
    client.post('/auth/2fa/disable', { code }),

  verifyLogin2FA: (token: string, code: string) =>
    client.post<AuthResponse>('/auth/2fa/verify-login', { token, code }),
};
