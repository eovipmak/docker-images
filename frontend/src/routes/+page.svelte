<script lang="ts">
	import { onMount } from 'svelte';
	import { env } from '$env/dynamic/public';

	let apiStatus = 'Checking...';
	let apiVersion = '';

	onMount(async () => {
		try {
			const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080';
			const response = await fetch(`${apiUrl}/api/v1`);
			if (response.ok) {
				const data = await response.json();
				apiStatus = 'Connected ✓';
				apiVersion = data.version;
			} else {
				apiStatus = 'Error';
			}
		} catch (error) {
			apiStatus = 'Disconnected';
		}
	});
</script>

<svelte:head>
	<title>V-Insight - Multi-tenant Monitoring SaaS</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-3xl mx-auto">
		<h1 class="text-4xl font-bold text-gray-900 mb-4">Welcome to V-Insight</h1>
		<p class="text-xl text-gray-600 mb-8">Multi-tenant Monitoring SaaS Platform</p>

		<div class="bg-white rounded-lg shadow-md p-6">
			<h2 class="text-2xl font-semibold text-gray-800 mb-4">System Status</h2>
			<div class="space-y-2">
				<p class="text-gray-700">
					Backend API: <span class="font-semibold text-blue-600">{apiStatus}</span>
				</p>
				{#if apiVersion}
					<p class="text-gray-700">
						API Version: <span class="font-semibold text-blue-600">{apiVersion}</span>
					</p>
				{/if}
			</div>
		</div>
	</div>
</div>

