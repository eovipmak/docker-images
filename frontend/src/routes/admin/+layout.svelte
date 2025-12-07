<script lang="ts">
	import Navbar from '$lib/components/Navbar.svelte';
	import { authStore } from '$lib/stores/auth';
	import { onMount } from 'svelte';
    import { browser } from '$app/environment';

    const navItems = [
		{ name: 'Dashboard', path: '/admin' },
		{ name: 'Users', path: '/admin/users' },
		{ name: 'Monitors', path: '/admin/monitors' },
        { name: 'Alert Rules', path: '/admin/alert-rules' },
        { name: 'Alert Channels', path: '/admin/alert-channels' }
	];

    let authInitialized = false;

    onMount(async () => {
        await authStore.checkAuth();
        authInitialized = true;
    });

    $: if (browser && authInitialized) {
        if (!$authStore.isAuthenticated) {
            window.location.href = '/login';
        } else if ($authStore.currentUser?.role !== 'admin') {
            window.location.href = '/user/dashboard';
        }
    }
</script>

{#if $authStore.isAuthenticated && $authStore.currentUser?.role === 'admin'}
    <div class="min-h-screen bg-gray-50 dark:bg-[#0b0c15]">
        <div class="bg-red-600 text-white text-xs font-bold text-center py-1">ADMIN MODE</div>
        <Navbar {navItems} homeLink="/admin" isAdmin={true} />
        <main class="w-full">
            <slot />
        </main>
    </div>
{/if}
