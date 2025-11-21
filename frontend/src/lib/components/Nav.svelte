<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import { onMount } from 'svelte';

	// Check authentication status on mount
	onMount(async () => {
		await authStore.checkAuth();
	});

	const publicNavItems = [
		{ href: '/login', label: 'Login' }
	];

	const authenticatedNavItems = [
		{ href: '/', label: 'Home' },
		{ href: '/dashboard', label: 'Dashboard' },
		{ href: '/domains', label: 'Domains' },
		{ href: '/alerts', label: 'Alerts' },
		{ href: '/incidents', label: 'Incidents' },
		{ href: '/settings', label: 'Settings' }
	];

	function handleLogout() {
		authStore.logout();
		window.location.href = '/login';
	}
</script>

<nav class="bg-blue-600 text-white shadow-lg">
	<div class="container mx-auto px-4">
		<div class="flex items-center justify-between h-16">
			<div class="flex items-center">
				<span class="text-xl font-bold">V-Insight</span>
			</div>
			<div class="flex space-x-4">
				{#if $authStore.isAuthenticated}
					{#each authenticatedNavItems as item}
						<a
							href={item.href}
							class="px-3 py-2 rounded-md text-sm font-medium transition-colors {$page.url
								.pathname === item.href
								? 'bg-blue-700'
								: 'hover:bg-blue-500'}"
						>
							{item.label}
						</a>
					{/each}
					<button
						on:click={handleLogout}
						class="px-3 py-2 rounded-md text-sm font-medium transition-colors hover:bg-blue-500"
					>
						Logout
					</button>
				{:else}
					{#each publicNavItems as item}
						<a
							href={item.href}
							class="px-3 py-2 rounded-md text-sm font-medium transition-colors {$page.url
								.pathname === item.href
								? 'bg-blue-700'
								: 'hover:bg-blue-500'}"
						>
							{item.label}
						</a>
					{/each}
				{/if}
			</div>
		</div>
	</div>
</nav>
