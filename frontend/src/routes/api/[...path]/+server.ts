import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

// BACKEND_API_URL: for server-side proxy (Docker internal network)
// Falls back to PUBLIC_API_URL or localhost for compatibility
const BACKEND_URL = env.BACKEND_API_URL || env.PUBLIC_API_URL || 'http://localhost:8080';

export const GET: RequestHandler = async ({ params, url, request }) => {
	return proxyRequest('GET', params.path, url, request);
};

export const POST: RequestHandler = async ({ params, url, request }) => {
	return proxyRequest('POST', params.path, url, request);
};

export const PUT: RequestHandler = async ({ params, url, request }) => {
	return proxyRequest('PUT', params.path, url, request);
};

export const PATCH: RequestHandler = async ({ params, url, request }) => {
	return proxyRequest('PATCH', params.path, url, request);
};

export const DELETE: RequestHandler = async ({ params, url, request }) => {
	return proxyRequest('DELETE', params.path, url, request);
};

async function proxyRequest(
	method: string,
	path: string,
	url: URL,
	request: Request
): Promise<Response> {
	const backendUrl = `${BACKEND_URL}/api/${path}${url.search}`;

	const headers: HeadersInit = {};
	request.headers.forEach((value, key) => {
		if (key.toLowerCase() !== 'host' && key.toLowerCase() !== 'connection') {
			headers[key] = value;
		}
	});

	try {
		const body = method !== 'GET' && method !== 'HEAD' ? await request.text() : undefined;

		const response = await fetch(backendUrl, {
			method,
			headers,
			body
		});

		const responseHeaders = new Headers();
		response.headers.forEach((value, key) => {
			if (
				key.toLowerCase() !== 'transfer-encoding' &&
				key.toLowerCase() !== 'connection' &&
				key.toLowerCase() !== 'keep-alive' &&
				key.toLowerCase() !== 'content-encoding' &&
				key.toLowerCase() !== 'content-length'
			) {
				responseHeaders.set(key, value);
			}
		});

		// Add CORS headers
		responseHeaders.set('access-control-allow-origin', '*');
		responseHeaders.set('access-control-allow-methods', 'GET, POST, PUT, PATCH, DELETE, OPTIONS');
		responseHeaders.set('access-control-allow-headers', '*');

		const responseBody = await response.text();

		return new Response(responseBody, {
			status: response.status,
			statusText: response.statusText,
			headers: responseHeaders
		});
	} catch (error) {
		console.error('Proxy error:', error);
		return new Response(JSON.stringify({ error: 'Backend service unavailable' }), {
			status: 503,
			headers: { 'Content-Type': 'application/json' }
		});
	}
}
