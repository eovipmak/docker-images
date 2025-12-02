import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

// BACKEND_API_URL: for server-side proxy (Docker internal network)
const BACKEND_URL = env.BACKEND_API_URL || env.PUBLIC_API_URL || 'http://localhost:8080';

export const GET: RequestHandler = async ({ params, url, request }) => {
	const backendUrl = `${BACKEND_URL}/swagger/${params.path}${url.search}`;

	const headers: HeadersInit = {};
	request.headers.forEach((value, key) => {
		if (key.toLowerCase() !== 'host' && key.toLowerCase() !== 'connection') {
			headers[key] = value;
		}
	});

	try {
		const response = await fetch(backendUrl, {
			method: 'GET',
			headers
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

		const contentType = response.headers.get('content-type') || 'application/json';
		responseHeaders.set('content-type', contentType);

		const responseBody = await response.arrayBuffer();

		return new Response(responseBody, {
			status: response.status,
			statusText: response.statusText,
			headers: responseHeaders
		});
	} catch (error) {
		console.error('Swagger proxy error:', error);
		return new Response(JSON.stringify({ error: 'Backend service unavailable' }), {
			status: 503,
			headers: { 'Content-Type': 'application/json' }
		});
	}
};
