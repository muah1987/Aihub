import axios from 'axios';

const API_BASE = '/api/v1';

const client = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
});

// Request interceptor: attach access token
client.interceptors.request.use((config) => {
  const tokens = localStorage.getItem('tokens');
  if (tokens) {
    const { access_token } = JSON.parse(tokens);
    config.headers.Authorization = `Bearer ${access_token}`;
  }
  return config;
});

// Response interceptor: handle 401 with token refresh
client.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;
      const tokens = localStorage.getItem('tokens');
      if (tokens) {
        try {
          const { refresh_token } = JSON.parse(tokens);
          const res = await axios.post(`${API_BASE}/auth/refresh`, { refresh_token });
          localStorage.setItem('tokens', JSON.stringify(res.data.tokens));
          original.headers.Authorization = `Bearer ${res.data.tokens.access_token}`;
          return client(original);
        } catch {
          localStorage.removeItem('tokens');
          localStorage.removeItem('user');
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

export default client;
