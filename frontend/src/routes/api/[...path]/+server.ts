import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

const BACKEND_URL = env.BACKEND_API_URL || 'http://backend:8080';

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
				key.toLowerCase() !== 'keep-alive'
			) {
				responseHeaders.set(key, value);
			}
		});

		return new Response(response.body, {
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
