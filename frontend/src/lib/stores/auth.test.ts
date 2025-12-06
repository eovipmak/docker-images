import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { authStore } from './auth';

declare const global: any;

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

// Setup global mocks
global.localStorage = localStorageMock as any;
global.fetch = fetchMock as any;

describe('authStore', () => {
	beforeEach(() => {
		localStorageMock.clear();
		fetchMock.mockClear();
	});

	describe('initialization', () => {
		it('should initialize with unauthenticated state when no token exists', () => {
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});

		it('should initialize with authenticated state when token exists', () => {
			localStorageMock.setItem('auth_token', 'test-token');
			// Note: We can't test this directly as the store is already initialized
			// This would require dynamic store creation
		});
	});

	describe('login', () => {
		it('should set token in localStorage and update state', async () => {
			const token = 'test-auth-token';
			const userData = { id: 1, email: 'test@example.com', user_id: 1 };

			fetchMock.mockResolvedValueOnce({
				ok: true,
				json: async () => userData
			});

			await authStore.login(token);

			expect(localStorageMock.getItem('auth_token')).toBe(token);
			
			// Wait a bit for the async user fetch
			await new Promise(resolve => setTimeout(resolve, 50));
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
			expect(state.currentUser).toEqual(userData);
		});

		it('should handle user fetch failure gracefully', async () => {
			const token = 'test-auth-token';

			fetchMock.mockRejectedValueOnce(new Error('Network error'));

			await authStore.login(token);

			expect(localStorageMock.getItem('auth_token')).toBe(token);
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
			// User should be null due to fetch failure
			expect(state.currentUser).toBeNull();
		});

		it('should set authenticated even if user fetch returns non-ok', async () => {
			const token = 'test-auth-token';

			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 500
			});

			await authStore.login(token);

			expect(localStorageMock.getItem('auth_token')).toBe(token);
			
			await new Promise(resolve => setTimeout(resolve, 50));
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
		});
	});

	describe('logout', () => {
		it('should clear token and reset state', () => {
			localStorageMock.setItem('auth_token', 'test-token');
			
			authStore.logout();

			expect(localStorageMock.getItem('auth_token')).toBeNull();
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});

		it('should work even when no token exists', () => {
			authStore.logout();

			expect(localStorageMock.getItem('auth_token')).toBeNull();
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});
	});

	describe('checkAuth', () => {
		it('should validate token and fetch user data', async () => {
			const token = 'valid-token';
			const userData = { id: 1, email: 'test@example.com', user_id: 1 };

			localStorageMock.setItem('auth_token', token);

			fetchMock.mockResolvedValueOnce({
				ok: true,
				json: async () => userData
			});

			await authStore.checkAuth();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(true);
			expect(state.currentUser).toEqual(userData);
		});

		it('should clear invalid token and reset state', async () => {
			const token = 'invalid-token';
			localStorageMock.setItem('auth_token', token);

			fetchMock.mockResolvedValueOnce({
				ok: false,
				status: 401
			});

			await authStore.checkAuth();

			expect(localStorageMock.getItem('auth_token')).toBeNull();
			
			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});

		it('should handle network errors gracefully', async () => {
			const token = 'valid-token';
			localStorageMock.setItem('auth_token', token);

			fetchMock.mockRejectedValueOnce(new Error('Network error'));

			await authStore.checkAuth();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});

		it('should set unauthenticated when no token exists', async () => {
			await authStore.checkAuth();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.currentUser).toBeNull();
		});
	});

	describe('getToken', () => {
		it('should return token from localStorage', () => {
			const token = 'test-token';
			localStorageMock.setItem('auth_token', token);

			const retrievedToken = authStore.getToken();
			expect(retrievedToken).toBe(token);
		});

		it('should return null when no token exists', () => {
			const retrievedToken = authStore.getToken();
			expect(retrievedToken).toBeNull();
		});
	});
});
