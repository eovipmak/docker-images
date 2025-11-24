import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter({
			// Enable compression for static assets
			precompress: true
		})
	},

	// Compiler options for optimized output
	compilerOptions: {
		// Enable dev mode warnings only in development
		dev: process.env.NODE_ENV !== 'production',
		// Preserve whitespace only in development
		preserveWhitespace: false,
		// Enable CSS optimization
		css: 'injected'
	}
};

export default config;
