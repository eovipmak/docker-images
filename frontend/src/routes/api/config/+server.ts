import { env as dynamicEnv } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// Try dynamic env first (from container), then fall back to process.env
const getPublicApiUrl = (): string => {
	// Check dynamic env (from docker environment)
	if (dynamicEnv.PUBLIC_API_URL) {
		return dynamicEnv.PUBLIC_API_URL;
	}
	// Fallback to process.env
	if (process.env.PUBLIC_API_URL) {
		return process.env.PUBLIC_API_URL;
	}
	// Default for development
	return 'http://localhost:8080';
};

export const GET: RequestHandler = async () => {
	const publicApiUrl = getPublicApiUrl();
	console.log('[Config API] PUBLIC_API_URL:', publicApiUrl);
	
	return json({
		publicApiUrl
	});
};
