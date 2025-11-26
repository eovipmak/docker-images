import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fetchAPI } from './client';

// Mock localStorage
const localStorageMock = (() => {
	let store: Record<string, string> = {};

	return {
		getItem: (key: string) => store[key] || null,
		setItem: (key: string, value: string) => {
			store[key] = value.toString();
		},
		removeItem: (key: string) => {
			delete store[key];
		},
		clear: () => {
			store = {};
		}
	};
})();

// Mock fetch
const fetchMock = vi.fn();

// Mock window.location
const locationMock = {
	href: '',
	pathname: '/'
};

// Setup global mocks
global.localStorage = localStorageMock as any;
global.fetch = fetchMock as any;
Object.defineProperty(window, 'location', {
	value: locationMock,
	writable: true
});

describe('fetchAPI', () => {
	beforeEach(() => {
		localStorageMock.clear();
		fetchMock.mockClear();
		locationMock.href = '';
		locationMock.pathname = '/';
	});

	describe('basic functionality', () => {
		it('should make a fetch request with default options', async () => {
			const mockResponse = { ok: true, json: async () => ({ data: 'test' }) };
			fetchMock.mockResolvedValueOnce(mockResponse);

			const response = await fetchAPI('/api/v1/test');

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json'
				}
			});
			expect(response).toBe(mockResponse);
		});

		it('should include custom headers', async () => {
			const mockResponse = { ok: true };
			fetchMock.mockResolvedValueOnce(mockResponse);

			await fetchAPI('/api/v1/test', {
				headers: {
					'X-Custom-Header': 'custom-value'
				}
			});

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json',
					'X-Custom-Header': 'custom-value'
				}
			});
		});

		it('should support different HTTP methods', async () => {
			const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'];

			for (const method of methods) {
				fetchMock.mockResolvedValueOnce({ ok: true });

				await fetchAPI('/api/v1/test', { method });

				expect(fetchMock).toHaveBeenCalledWith(
					'/api/v1/test',
					expect.objectContaining({ method })
				);
				fetchMock.mockClear();
			}
		});

		it('should include request body', async () => {
			fetchMock.mockResolvedValueOnce({ ok: true });

			const body = JSON.stringify({ data: 'test' });
			await fetchAPI('/api/v1/test', {
				method: 'POST',
				body
			});

			expect(fetchMock).toHaveBeenCalledWith(
				'/api/v1/test',
				expect.objectContaining({
					method: 'POST',
					body
				})
			);
		});
	});

	describe('authentication', () => {
		it('should automatically add auth token from localStorage', async () => {
			const token = 'test-token';
			localStorageMock.setItem('auth_token', token);

			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('/api/v1/test');

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`
				}
			});
		});

		it('should use provided token instead of localStorage token', async () => {
			localStorageMock.setItem('auth_token', 'stored-token');
			const providedToken = 'provided-token';

			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('/api/v1/test', { token: providedToken });

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${providedToken}`
				}
			});
		});

		it('should skip auth when skipAuth is true', async () => {
			localStorageMock.setItem('auth_token', 'test-token');

			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('/api/v1/test', { skipAuth: true });

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json'
				}
			});
		});

		it('should not add auth header when no token exists', async () => {
			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('/api/v1/test');

			expect(fetchMock).toHaveBeenCalledWith('/api/v1/test', {
				headers: {
					'Content-Type': 'application/json'
				}
			});
		});
	});

	describe('401 handling', () => {
		it('should clear token and redirect to login on 401', async () => {
			const token = 'expired-token';
			localStorageMock.setItem('auth_token', token);
			locationMock.pathname = '/dashboard';

			fetchMock.mockResolvedValueOnce({ status: 401, ok: false });

			await fetchAPI('/api/v1/test');

			expect(localStorageMock.getItem('auth_token')).toBeNull();
			expect(locationMock.href).toBe('/login');
		});

		it('should not redirect if already on login page', async () => {
			localStorageMock.setItem('auth_token', 'token');
			locationMock.pathname = '/login';

			fetchMock.mockResolvedValueOnce({ status: 401, ok: false });

			await fetchAPI('/api/v1/test');

			expect(locationMock.href).toBe('');
		});

		it('should not redirect if already on register page', async () => {
			localStorageMock.setItem('auth_token', 'token');
			locationMock.pathname = '/register';

			fetchMock.mockResolvedValueOnce({ status: 401, ok: false });

			await fetchAPI('/api/v1/test');

			expect(locationMock.href).toBe('');
		});

		it('should not handle 401 when skipAuth is true', async () => {
			localStorageMock.setItem('auth_token', 'token');
			locationMock.pathname = '/dashboard';

			fetchMock.mockResolvedValueOnce({ status: 401, ok: false });

			await fetchAPI('/api/v1/test', { skipAuth: true });

			// Token should not be cleared
			expect(localStorageMock.getItem('auth_token')).toBe('token');
			// Should not redirect
			expect(locationMock.href).toBe('');
		});
	});

	describe('error handling', () => {
		it('should propagate fetch errors', async () => {
			const error = new Error('Network error');
			fetchMock.mockRejectedValueOnce(error);

			await expect(fetchAPI('/api/v1/test')).rejects.toThrow('Network error');
		});

		it('should handle non-401 error responses normally', async () => {
			const responses = [
				{ status: 400, ok: false },
				{ status: 403, ok: false },
				{ status: 404, ok: false },
				{ status: 500, ok: false }
			];

			for (const response of responses) {
				fetchMock.mockResolvedValueOnce(response);

				const result = await fetchAPI('/api/v1/test');

				expect(result).toBe(response);
				expect(locationMock.href).toBe('');
				fetchMock.mockClear();
			}
		});
	});

	describe('edge cases', () => {
		it('should handle empty endpoint', async () => {
			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('');

			expect(fetchMock).toHaveBeenCalledWith('', expect.any(Object));
		});

		it('should preserve all fetch options', async () => {
			fetchMock.mockResolvedValueOnce({ ok: true });

			await fetchAPI('/api/v1/test', {
				method: 'POST',
				mode: 'cors',
				cache: 'no-cache',
				credentials: 'include',
				redirect: 'follow',
				referrerPolicy: 'no-referrer'
			});

			expect(fetchMock).toHaveBeenCalledWith(
				'/api/v1/test',
				expect.objectContaining({
					method: 'POST',
					mode: 'cors',
					cache: 'no-cache',
					credentials: 'include',
					redirect: 'follow',
					referrerPolicy: 'no-referrer'
				})
			);
		});
	});
});
