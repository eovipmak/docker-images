import { env } from '$env/dynamic/public';

const BASE_URL = env.PUBLIC_API_URL || 'http://localhost:8080';

interface FetchOptions extends RequestInit {
	token?: string;
}

/**
 * Wrapper around fetch API with Bearer Token support
 * @param endpoint - API endpoint (e.g., '/api/v1/users')
 * @param options - Fetch options including optional Bearer token
 * @returns Promise with the fetch response
 */
export async function fetchAPI(endpoint: string, options: FetchOptions = {}): Promise<Response> {
	const { token, ...fetchOptions } = options;

	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(fetchOptions.headers as Record<string, string>)
	};

	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const url = `${BASE_URL}${endpoint}`;

	return fetch(url, {
		...fetchOptions,
		headers
	});
}

/**
 * Get the base API URL
 */
export function getBaseURL(): string {
	return BASE_URL;
}
