<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { authStore } from '$lib/stores/auth';
	import { themeStore } from '$lib/stores/theme';
	import { sidebarOpen, toggleSidebar } from '$lib/stores/sidebar';
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

<div class="flex h-screen bg-gray-50 dark:bg-slate-950 overflow-hidden font-sans">
	{#if $authStore.isAuthenticated || !isPublicRoute($page.url.pathname)}
		<Sidebar />
	{/if}
	
	<div class="flex-1 flex flex-col overflow-hidden relative">
		<!-- Mobile Header with Hamburger Menu -->
		{#if $authStore.isAuthenticated || !isPublicRoute($page.url.pathname)}
			<header class="lg:hidden flex items-center justify-between h-14 px-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex-shrink-0">
				<button
					on:click={toggleSidebar}
					class="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
					aria-label="Open menu"
				>
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
					</svg>
				</button>
				<span class="text-lg font-bold tracking-wider text-blue-500 dark:text-blue-400">V-INSIGHT</span>
				<div class="w-10"></div>
			</header>
		{/if}
		<main class="flex-1 overflow-x-hidden overflow-y-auto bg-gray-50 dark:bg-slate-950 p-4 md:p-8">
			<slot />
		</main>
	</div>
</div>

