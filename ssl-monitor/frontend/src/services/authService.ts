import axios from 'axios';

// Get base URL from environment or use default
const getBaseURL = () => {
  // Support runtime configuration for subpath deployments
  const base = import.meta.env.VITE_BASE_PATH || '';
  return base === '/' ? '' : base;
};

const api = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000,
});

// Global flag to prevent multiple simultaneous redirects (race condition guard)
let isRedirecting = false;

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    // WARNING: localStorage is vulnerable to XSS attacks
    // TODO: Migrate to HTTP-only, Secure cookies for production
    // This requires backend changes to support cookie-based authentication
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Use redirect guard to prevent race condition when multiple 401s occur
      if (!isRedirecting) {
        isRedirecting = true;
        // Clear token and redirect to login
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user');
        window.location.href = getBaseURL() + '/login';
      }
    }
    return Promise.reject(error);
  }
);

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
}

export interface User {
  id: number;
  email: string;
  is_active: boolean;
  is_superuser: boolean;
  is_verified: boolean;
}

export interface AuthResponse {
  access_token: string;
  refresh_token?: string;
  token_type: string;
}

// Authentication API calls
export const authService = {
  async register(data: RegisterData): Promise<User> {
    const response = await api.post('/auth/register', data);
    return response.data;
  },

  async login(email: string, password: string): Promise<AuthResponse> {
    const formData = new FormData();
    formData.append('username', email);
    formData.append('password', password);

    const response = await api.post('/auth/jwt/login', formData);
    
    // Store the tokens
    if (response.data.access_token) {
      localStorage.setItem('auth_token', response.data.access_token);
    }
    if (response.data.refresh_token) {
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }
    
    return response.data;
  },

  async logout(): Promise<void> {
    try {
      await api.post('/auth/jwt/logout');
    } finally {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('user');
    }
  },

  async getCurrentUser(): Promise<User> {
    const response = await api.get('/auth/me');
    localStorage.setItem('user', JSON.stringify(response.data));
    return response.data;
  },

  async refreshToken(): Promise<AuthResponse> {
    const response = await api.post('/auth/jwt/refresh');
    
    // Store the new tokens
    if (response.data.access_token) {
      localStorage.setItem('auth_token', response.data.access_token);
    }
    if (response.data.refresh_token) {
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }
    
    return response.data;
  },

  async requestVerification(): Promise<void> {
    await api.post('/auth/request-verify-token');
  },

  async verify(token: string): Promise<void> {
    await api.post('/auth/verify', { token });
  },

  async forgotPassword(email: string): Promise<void> {
    await api.post('/auth/forgot-password', { email });
  },

  async resetPassword(token: string, password: string): Promise<void> {
    await api.post('/auth/reset-password', { token, password });
  },

  getToken(): string | null {
    return localStorage.getItem('auth_token');
  },

  isAuthenticated(): boolean {
    return !!this.getToken();
  },
};

export default authService;
