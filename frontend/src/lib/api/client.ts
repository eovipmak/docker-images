import { browser } from '$app/environment';

interface FetchOptions extends RequestInit {
	token?: string;
	skipAuth?: boolean;
}

/**
 * Get the auth token from localStorage
 */
function getAuthToken(): string | null {
	if (browser) {
		return localStorage.getItem('auth_token');
	}
	return null;
}

/**
 * Wrapper around fetch API with automatic Bearer Token support
 * All API requests are proxied through SvelteKit server to avoid CORS issues
 * @param endpoint - API endpoint (e.g., '/api/v1/users')
 * @param options - Fetch options including optional Bearer token
 * @returns Promise with the fetch response
 */
export async function fetchAPI(endpoint: string, options: FetchOptions = {}): Promise<Response> {
	const { token, skipAuth = false, ...fetchOptions } = options;

	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(fetchOptions.headers as Record<string, string>)
	};

	// Automatically add auth token if available and not skipped
	if (!skipAuth) {
		const authToken = token || getAuthToken();
		if (authToken) {
			headers['Authorization'] = `Bearer ${authToken}`;
		}
	}

	const response = await fetch(endpoint, {
		...fetchOptions,
		headers
	});

	// Handle 401 Unauthorized - redirect to login
	if (response.status === 401 && browser && !skipAuth) {
		// Clear invalid token
		localStorage.removeItem('auth_token');
		
		// Redirect to login if not already there
		if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
			window.location.href = '/login';
		}
	}

	return response;
}
