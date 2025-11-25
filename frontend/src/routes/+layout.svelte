<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { authStore } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';

	// Public routes that don't require authentication
	const publicRoutes = ['/', '/login', '/register'];

	// Track if auth initialization is complete
	let authInitialized = false;

	// Check if current route is public
	function isPublicRoute(pathname: string): boolean {
		return publicRoutes.includes(pathname);
	}

	// Reactive statement to handle route protection
	$: if (browser && authInitialized && !$authStore.isAuthenticated && !isPublicRoute($page.url.pathname)) {
		window.location.href = '/login';
	}

	// Check authentication on mount
	onMount(async () => {
		await authStore.checkAuth();
		authInitialized = true;
	});
</script>

<div class="flex h-screen bg-gray-50 overflow-hidden font-sans">
	{#if $authStore.isAuthenticated || !isPublicRoute($page.url.pathname)}
		<Sidebar />
	{/if}
	
	<div class="flex-1 flex flex-col overflow-hidden relative">
		<main class="flex-1 overflow-x-hidden overflow-y-auto bg-gray-50 p-4 md:p-8">
			<slot />
		</main>
	</div>
</div>

