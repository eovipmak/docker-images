import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';
import { visualizer } from 'rollup-plugin-visualizer';

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

// Check if bundle analysis is requested
const analyze = process.env.ANALYZE === 'true';

// Build plugins array - separate sveltekit and other plugins
const extraPlugins = [];

// Add visualizer plugin when analyzing bundle
if (analyze) {
extraPlugins.push(
visualizer({
filename: 'stats.html',
open: true,
gzipSize: true,
brotliSize: true,
template: 'treemap'
})
);
}

return {
plugins: [sveltekit(), ...extraPlugins],
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
},
build: {
// Enable source maps for production debugging (optional)
sourcemap: false,
// Minification settings - use terser for production builds
minify: 'terser',
terserOptions: {
compress: {
drop_console: mode === 'production',
drop_debugger: mode === 'production'
},
mangle: true
},
// Rollup options for code splitting and tree shaking
rollupOptions: {
output: {
// Manual chunks for better caching
manualChunks: (id) => {
// Vendor chunk for node_modules
if (id.includes('node_modules')) {
// Chart.js and related dependencies
if (id.includes('chart.js') || id.includes('chartjs-adapter-date-fns')) {
return 'vendor-charts';
}
// Date utilities
if (id.includes('date-fns')) {
return 'vendor-date';
}
// Other vendor libraries
return 'vendor';
}
},
// Optimize chunk names for caching
chunkFileNames: 'assets/[name]-[hash].js',
entryFileNames: 'assets/[name]-[hash].js',
assetFileNames: 'assets/[name]-[hash].[ext]'
},
// Tree shaking configuration
treeshake: {
moduleSideEffects: 'no-external',
propertyReadSideEffects: false
}
},
// Target modern browsers for smaller bundles
target: 'es2020',
// Chunk size warning limit
chunkSizeWarningLimit: 500
}
};
});
