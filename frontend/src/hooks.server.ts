import { env } from '$env/dynamic/private';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event, {
		transformPageChunk: ({ html }) => {
			// Replace PUBLIC_API_URL placeholder with actual value
			// This enables SSE connection from browser to backend
			const publicApiUrl = env.PUBLIC_API_URL || '';
			return html.replace(/%PUBLIC_API_URL%/g, publicApiUrl);
		}
	});
	
	return response;
};
