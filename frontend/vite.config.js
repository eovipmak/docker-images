import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	optimizeDeps: {
		include: ['svelte', '@sveltejs/kit', 'vite']
	},
	server: {
		host: '0.0.0.0',
		port: 3000,
		allowedHosts: ['localhost', '127.0.0.1', 'monit.24-7.top'],
		watch: {
			usePolling: true
		}
	}
});
