<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/components/Navbar.svelte';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';

	// Subscribe to theme store to keep it active
	$: if (browser && $themeStore !== undefined) {
		// Theme store subscription is handled in the store itself
	}

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

// Redirect authenticated users away from public routes (landing, login, register)
$: if (browser && authInitialized && $authStore.isAuthenticated && isPublicRoute($page.url.pathname)) {
	// If user is authenticated and on a public route, send them to the dashboard
	// Avoid causing a redirect loop if they're already on /dashboard
	if ($page.url.pathname !== '/dashboard') {
		window.location.href = '/dashboard';
	}
}

	// Check authentication on mount
	onMount(async () => {
		await authStore.checkAuth();
		authInitialized = true;
	});
</script>

<div class="min-h-screen bg-gray-50 dark:bg-[#0b0c15] text-slate-900 dark:text-slate-100 font-sans transition-colors duration-300 selection:bg-indigo-500/30">
	{#if $authStore.isAuthenticated || !isPublicRoute($page.url.pathname)}
		<Navbar />
	{/if}
	
	<main class="w-full">
		<slot />
	</main>
</div>
