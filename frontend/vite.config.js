import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	// Load env file based on `mode` in the current working directory.
	// Set the third parameter to '' to load all env regardless of the `VITE_` prefix.
	const env = loadEnv(mode, process.cwd(), '');
	
	// Parse VITE_ALLOWED_HOSTS from environment variable
	// Expected format: comma-separated list of hosts
	// Example: VITE_ALLOWED_HOSTS=localhost,127.0.0.1,monit.24-7.top
	const allowedHosts = env.VITE_ALLOWED_HOSTS 
		? env.VITE_ALLOWED_HOSTS.split(',').map(host => host.trim()).filter(Boolean)
		: ['localhost', '127.0.0.1'];

	return {
		plugins: [sveltekit()],
		optimizeDeps: {
			include: ['svelte', '@sveltejs/kit', 'vite']
		},
		server: {
			host: '0.0.0.0',
			port: 3000,
			allowedHosts,
			watch: {
				usePolling: true
			}
		}
	};
});
