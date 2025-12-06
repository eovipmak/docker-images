<script lang="ts">
	import Navbar from '$lib/components/Navbar.svelte';
	import { authStore } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
    import { browser } from '$app/environment';

    const navItems = [
		{ name: 'Dashboard', path: '/user/dashboard' },
		{ name: 'Monitors', path: '/user/monitors' },
		{ name: 'Alerts', path: '/user/alerts' },
		{ name: 'Incidents', path: '/user/incidents' },
        { name: 'Settings', path: '/user/settings' }
	];

    let authInitialized = false;

    onMount(async () => {
        await authStore.checkAuth();
        authInitialized = true;
    });

    $: if (browser && authInitialized && !$authStore.isAuthenticated) {
        window.location.href = '/login';
    }
</script>

{#if $authStore.isAuthenticated}
    <div class="min-h-screen bg-gray-50 dark:bg-[#0b0c15]">
        <Navbar {navItems} homeLink="/user/dashboard" />
        <main class="w-full">
            <slot />
        </main>
    </div>
{/if}
