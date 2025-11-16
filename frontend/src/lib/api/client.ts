interface FetchOptions extends RequestInit {
	token?: string;
}

/**
 * Wrapper around fetch API with Bearer Token support
 * All API requests are proxied through SvelteKit server to avoid CORS issues
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

	return fetch(endpoint, {
		...fetchOptions,
		headers
	});
}
