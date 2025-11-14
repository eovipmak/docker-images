import axios from 'axios';
import type { SSLCheckResponse } from '../types';

// Get base path from environment or use default
const getBasePath = () => {
  // Support runtime configuration for subpath deployments
  const base = import.meta.env.VITE_BASE_PATH || '';
  // Treat empty string or '/' as root (empty string)
  // For non-empty bases, remove trailing slash
  if (!base || base === '/') {
    return '';
  }
  return base.endsWith('/') ? base.slice(0, -1) : base;
};

const api = axios.create({
  baseURL: getBasePath(),
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

// Add response interceptor to handle 401 errors and auto-refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // If 401 and we haven't tried to refresh yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Try to refresh the token
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          const response = await api.post('/auth/jwt/refresh');
          
          if (response.data.access_token) {
            localStorage.setItem('auth_token', response.data.access_token);
            if (response.data.refresh_token) {
              localStorage.setItem('refresh_token', response.data.refresh_token);
            }
            
            // Retry the original request with new token
            originalRequest.headers.Authorization = `Bearer ${response.data.access_token}`;
            return api(originalRequest);
          }
        }
      } catch (refreshError) {
        // Refresh failed, use redirect guard and clear tokens
        if (!isRedirecting) {
          isRedirecting = true;
          localStorage.removeItem('auth_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user');
          window.location.href = getBasePath() + '/login';
        }
        return Promise.reject(refreshError);
      }
    }
    
    // For other 401 errors or if refresh failed, use redirect guard
    if (error.response?.status === 401) {
      if (!isRedirecting) {
        isRedirecting = true;
        localStorage.removeItem('auth_token');
        localStorage.removeItem('refresh_token');
        localStorage.removeItem('user');
        window.location.href = getBasePath() + '/login';
      }
    }
    
    return Promise.reject(error);
  }
);

// Helper to determine if input is an IP address
const isIPAddress = (host: string): boolean => {
  const ipv4Pattern = /^(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])$/;
  const ipv6Pattern = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;
  
  return ipv4Pattern.test(host) || ipv6Pattern.test(host);
};

// Parse target input to extract host and port
export const parseTarget = (target: string): { host: string; port: number } | null => {
  const trimmed = target.trim();
  if (!trimmed) {
    return null;
  }

  // Check for bracketed IPv6 address with port: [IPv6]:port
  const ipv6BracketMatch = trimmed.match(/^\[([^\]]+)\]:(\d+)$/);
  if (ipv6BracketMatch) {
    const host = ipv6BracketMatch[1];
    const port = parseInt(ipv6BracketMatch[2], 10);
    
    if (port < 1 || port > 65535) {
      throw new Error('Port must be between 1 and 65535');
    }
    
    return { host, port };
  }

  // Check for bracketed IPv6 address without port: [IPv6]
  const ipv6BracketOnlyMatch = trimmed.match(/^\[([^\]]+)\]$/);
  if (ipv6BracketOnlyMatch) {
    return {
      host: ipv6BracketOnlyMatch[1],
      port: 443,
    };
  }

  // Check for non-bracketed input with port: host:port (IPv4 or hostname)
  const hostPortMatch = trimmed.match(/^([^:]+):(\d+)$/);
  
  if (hostPortMatch) {
    const host = hostPortMatch[1];
    const port = parseInt(hostPortMatch[2], 10);
    
    if (port < 1 || port > 65535) {
      throw new Error('Port must be between 1 and 65535');
    }
    
    return { host, port };
  }

  // No port specified, use default
  return {
    host: trimmed,
    port: 443,
  };
};

export const checkSSL = async (target: string): Promise<SSLCheckResponse> => {
  const parsed = parseTarget(target);
  
  if (!parsed) {
    throw new Error('Invalid target');
  }

  const params: Record<string, string | number> = { port: parsed.port };
  
  if (isIPAddress(parsed.host)) {
    params.ip = parsed.host;
  } else {
    params.domain = parsed.host;
  }

  const response = await api.get<SSLCheckResponse>('/api/check', { params });
  return response.data;
};

// Get monitoring statistics
export const getStats = async () => {
  const response = await api.get('/api/stats');
  return response.data;
};

// Get check history
export const getHistory = async (domain?: string, limit: number = 50) => {
  const params: Record<string, string | number> = { limit };
  if (domain) {
    params.domain = domain;
  }
  const response = await api.get('/api/history', { params });
  return response.data;
};

// Get list of monitored domains
export const getDomains = async (limit: number = 100) => {
  const response = await api.get('/api/domains', { params: { limit } });
  return response.data;
};

// Add a new domain to monitor
export const addDomain = async (domain: string, port: number = 443) => {
  const response = await api.post('/api/domains', { domain, port });
  return response.data;
};

// Alert Configuration APIs
export const getAlertConfig = async () => {
  const response = await api.get('/api/alert-config');
  return response.data;
};

export const updateAlertConfig = async (config: any) => {
  const response = await api.post('/api/alert-config', config);
  return response.data;
};

// Alerts APIs
export const getAlerts = async (unreadOnly = false, unresolvedOnly = false, limit = 50) => {
  const params: Record<string, boolean | number> = { limit };
  if (unreadOnly) params.unread_only = true;
  if (unresolvedOnly) params.unresolved_only = true;
  
  const response = await api.get('/api/alerts', { params });
  return response.data;
};

export const markAlertRead = async (alertId: number) => {
  const response = await api.patch(`/api/alerts/${alertId}/read`);
  return response.data;
};

export const markAlertResolved = async (alertId: number) => {
  const response = await api.patch(`/api/alerts/${alertId}/resolve`);
  return response.data;
};

export const deleteAlert = async (alertId: number) => {
  const response = await api.delete(`/api/alerts/${alertId}`);
  return response.data;
};

export default api;
