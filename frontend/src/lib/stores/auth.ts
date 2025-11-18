import { writable } from 'svelte/store';
import { browser } from '$app/environment';

function createAuthStore() {
	// Initialize from localStorage if in browser
	const initialValue = browser ? !!localStorage.getItem('auth_token') : false;
	
	const { subscribe, set } = writable<boolean>(initialValue);

	return {
		subscribe,
		login: (token: string) => {
			if (browser) {
				localStorage.setItem('auth_token', token);
			}
			set(true);
		},
		logout: () => {
			if (browser) {
				localStorage.removeItem('auth_token');
			}
			set(false);
		},
		checkAuth: () => {
			if (browser) {
				const hasToken = !!localStorage.getItem('auth_token');
				set(hasToken);
			}
		}
	};
}

export const isAuthenticated = createAuthStore();
