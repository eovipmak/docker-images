import { writable } from 'svelte/store';
import { browser } from '$app/environment';

interface User {
	id: number;
	email: string;
	tenant_id: number;
}

interface AuthState {
	isAuthenticated: boolean;
	currentUser: User | null;
}

function createAuthStore() {
	// Initialize from localStorage if in browser
	const initialState: AuthState = {
		isAuthenticated: browser ? !!localStorage.getItem('auth_token') : false,
		currentUser: null
	};
	
	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,
		login: async (token: string) => {
			if (browser) {
				localStorage.setItem('auth_token', token);
			}
			
			// Update state immediately
			update(state => ({
				...state,
				isAuthenticated: true
			}));

			// Fetch user info in the background
			try {
				const response = await fetch('/api/v1/auth/me', {
					headers: {
						'Authorization': `Bearer ${token}`
					}
				});
				
				if (response.ok) {
					const userData = await response.json();
					update(state => ({
						...state,
						currentUser: userData
					}));
				}
			} catch (error) {
				console.error('Failed to fetch user info:', error);
			}
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('auth_token');
			}
			set({ isAuthenticated: false, currentUser: null });
		},
		checkAuth: async () => {
			if (browser) {
				const token = localStorage.getItem('auth_token');
				if (token) {
					try {
						// Fetch current user info from backend
						const response = await fetch('/api/v1/auth/me', {
							headers: {
								'Authorization': `Bearer ${token}`
							}
						});
						
						if (response.ok) {
							const userData = await response.json();
							update(state => ({
								...state,
								isAuthenticated: true,
								currentUser: userData
							}));
						} else {
							// Token is invalid, clear it
							localStorage.removeItem('auth_token');
							set({ isAuthenticated: false, currentUser: null });
						}
					} catch (error) {
						console.error('Auth check failed:', error);
						set({ isAuthenticated: false, currentUser: null });
					}
				} else {
					set({ isAuthenticated: false, currentUser: null });
				}
			}
		},
		getToken: (): string | null => {
			if (browser) {
				return localStorage.getItem('auth_token');
			}
			return null;
		}
	};
}

export const authStore = createAuthStore();

// For backward compatibility with existing code
export const isAuthenticated = {
	subscribe: (fn: (value: boolean) => void) => {
		return authStore.subscribe((state) => fn(state.isAuthenticated));
	},
	login: authStore.login,
	logout: authStore.logout,
	checkAuth: authStore.checkAuth
};
