<script lang="ts">
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
	import { authStore } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';

	// Public routes that don't require authentication
	const publicRoutes = ['/', '/login', '/register'];

	// Check if current route is public
	function isPublicRoute(pathname: string): boolean {
		return publicRoutes.includes(pathname);
	}

	// Check authentication on mount and when route changes
	onMount(async () => {
		// Check authentication status
		await authStore.checkAuth();

		// Handle route protection
		if (browser) {
			const unsubscribe = page.subscribe(($page) => {
				const pathname = $page.url.pathname;
				
				// If not authenticated and trying to access protected route
				authStore.subscribe((state) => {
					if (!state.isAuthenticated && !isPublicRoute(pathname)) {
						window.location.href = '/login';
					}
				})();
			});

			return () => {
				unsubscribe();
			};
		}
	});
</script>

<div class="min-h-screen bg-gray-50">
	<Nav />
	<main>
		<slot />
	</main>
</div>

