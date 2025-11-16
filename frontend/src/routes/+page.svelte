<script>
	import { onMount } from 'svelte';

	let apiStatus = 'Checking...';
	let apiVersion = '';

	onMount(async () => {
		try {
			const response = await fetch('http://localhost:8080/api/v1');
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

<div class="container">
	<h1>Welcome to V-Insight</h1>
	<p>Multi-tenant Monitoring SaaS Platform</p>
	
	<div class="status">
		<h2>System Status</h2>
		<p>Backend API: <strong>{apiStatus}</strong></p>
		{#if apiVersion}
			<p>API Version: <strong>{apiVersion}</strong></p>
		{/if}
	</div>
</div>

<style>
	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 2rem;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
	}

	h1 {
		color: #333;
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
	}

	p {
		color: #666;
		font-size: 1.2rem;
	}

	.status {
		margin-top: 2rem;
		padding: 1.5rem;
		background: #f5f5f5;
		border-radius: 8px;
	}

	.status h2 {
		margin-top: 0;
		color: #333;
	}

	strong {
		color: #2563eb;
	}
</style>
