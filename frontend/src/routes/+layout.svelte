<script lang="ts">
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
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

<div class="min-h-screen bg-gray-50">
	<Nav />
	<main>
		<slot />
	</main>
</div>

