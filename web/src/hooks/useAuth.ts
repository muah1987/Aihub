import { useEffect } from 'react';
import { useAuthStore } from '../store/authStore';

export function useAuth() {
  const store = useAuthStore();

  useEffect(() => {
    store.loadFromStorage();
  }, []);

  return {
    user: store.user,
    loading: store.loading,
    error: store.error,
    isAuthenticated: !!store.user,
    login: store.login,
    register: store.register,
    logout: store.logout,
  };
}
