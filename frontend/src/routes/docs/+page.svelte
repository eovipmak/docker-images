<script>
	import { onMount } from 'svelte';

	let loaded = false;

	onMount(() => {
		// Load Swagger UI CSS
		const link = document.createElement('link');
		link.rel = 'stylesheet';
		link.href = 'https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css';
		document.head.appendChild(link);

		// Load Swagger UI JS
		const script = document.createElement('script');
		script.src = 'https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js';
		script.onload = () => {
			// Initialize Swagger UI
			// @ts-ignore - SwaggerUIBundle is loaded from CDN
			window.SwaggerUIBundle({
				url: '/swagger/doc.json',
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					// @ts-ignore
					window.SwaggerUIBundle.presets.apis,
					// @ts-ignore
					window.SwaggerUIBundle.SwaggerUIStandalonePreset
				],
				layout: 'BaseLayout',
				defaultModelsExpandDepth: 1,
				defaultModelExpandDepth: 1,
				docExpansion: 'list',
				filter: true,
				persistAuthorization: true
			});
			loaded = true;
		};
		document.head.appendChild(script);
	});
</script>

<svelte:head>
	<title>API Documentation - V-Insight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50">
	<!-- Header -->
	<div class="bg-white border-b border-gray-200">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-3xl font-bold text-gray-900">API Documentation</h1>
					<p class="mt-2 text-sm text-gray-600">
						Complete API reference for V-Insight monitoring platform
					</p>
				</div>
				<a
					href="/dashboard"
					class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
				>
					Back to Dashboard
				</a>
			</div>
		</div>
	</div>

	<!-- Quick Links -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
		<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
			<h3 class="text-sm font-medium text-blue-900 mb-2">Quick Start</h3>
			<div class="text-sm text-blue-800 space-y-1">
				<p>
					<strong>1. Authentication:</strong> Use the <code
						class="bg-blue-100 px-1 rounded">/auth/login</code
					> endpoint to get a JWT token
				</p>
				<p>
					<strong>2. Authorization:</strong> Include the token in the Authorization header as
					<code class="bg-blue-100 px-1 rounded">Bearer &lt;token&gt;</code>
				</p>
				<p>
					<strong>3. Try it out:</strong> Use the "Authorize" button below to set your token and test
					endpoints directly
				</p>
			</div>
		</div>
	</div>

	<!-- Swagger UI Container -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-8">
		<div class="bg-white rounded-lg shadow">
			{#if !loaded}
				<div class="flex items-center justify-center p-12">
					<div class="text-center">
						<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600">
						</div>
						<p class="mt-4 text-gray-600">Loading API documentation...</p>
					</div>
				</div>
			{/if}
			<div id="swagger-ui"></div>
		</div>
	</div>
</div>

<style>
	/* Custom styles for Swagger UI */
	:global(#swagger-ui) {
		font-family: inherit;
	}

	:global(.swagger-ui .topbar) {
		display: none;
	}

	:global(.swagger-ui .info) {
		margin: 20px 0;
	}

	:global(.swagger-ui .scheme-container) {
		background: #fafafa;
		box-shadow: none;
		padding: 20px;
		border-radius: 4px;
	}
</style>
