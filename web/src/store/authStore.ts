import { create } from 'zustand';
import { authApi, type User, type TokenPair } from '../api/auth';

interface AuthState {
  user: User | null;
  tokens: TokenPair | null;
  loading: boolean;
  error: string | null;
  pendingTwoFactorToken: string | null;
  login: (email: string, password: string) => Promise<void>;
  verifyTwoFactor: (code: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => void;
  loadFromStorage: () => void;
  updateUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  tokens: null,
  loading: false,
  error: null,
  pendingTwoFactorToken: null,

  login: async (email, password) => {
    set({ loading: true, error: null, pendingTwoFactorToken: null });
    try {
      const { data } = await authApi.login(email, password);
      if (data.requires_2fa && data.pending_token) {
        set({ pendingTwoFactorToken: data.pending_token, loading: false });
        return;
      }
      if (data.user && data.tokens) {
        localStorage.setItem('user', JSON.stringify(data.user));
        localStorage.setItem('tokens', JSON.stringify(data.tokens));
        set({ user: data.user, tokens: data.tokens, loading: false });
      }
    } catch (err: unknown) {
      const message = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Login failed';
      set({ error: message, loading: false });
      throw new Error(message);
    }
  },

  verifyTwoFactor: async (code) => {
    const token = get().pendingTwoFactorToken;
    if (!token) throw new Error('No pending 2FA token');
    set({ loading: true, error: null });
    try {
      const { data } = await authApi.verifyLogin2FA(token, code);
      localStorage.setItem('user', JSON.stringify(data.user));
      localStorage.setItem('tokens', JSON.stringify(data.tokens));
      set({ user: data.user, tokens: data.tokens, pendingTwoFactorToken: null, loading: false });
    } catch (err: unknown) {
      const message = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Invalid code';
      set({ error: message, loading: false });
      throw new Error(message);
    }
  },

  register: async (email, password, displayName) => {
    set({ loading: true, error: null });
    try {
      const { data } = await authApi.register(email, password, displayName);
      localStorage.setItem('user', JSON.stringify(data.user));
      localStorage.setItem('tokens', JSON.stringify(data.tokens));
      set({ user: data.user, tokens: data.tokens, loading: false });
    } catch (err: unknown) {
      const message = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Registration failed';
      set({ error: message, loading: false });
      throw new Error(message);
    }
  },

  logout: () => {
    localStorage.removeItem('user');
    localStorage.removeItem('tokens');
    set({ user: null, tokens: null, pendingTwoFactorToken: null });
    authApi.logout().catch(() => {});
  },

  loadFromStorage: () => {
    const user = localStorage.getItem('user');
    const tokens = localStorage.getItem('tokens');
    if (user && tokens) {
      set({ user: JSON.parse(user), tokens: JSON.parse(tokens) });
    }
  },

  updateUser: (user) => {
    localStorage.setItem('user', JSON.stringify(user));
    set({ user });
  },
}));
