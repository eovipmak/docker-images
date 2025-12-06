<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
    import { authStore } from '$lib/stores/auth';

    let stats = {
        users: 0,
        monitors: 0,
        alertRules: 0
    };
    let isLoading = true;
    let error = '';

    onMount(async () => {
        try {
            const [usersRes, monitorsRes, alertsRes] = await Promise.all([
                fetchAPI('/api/v1/admin/users'),
                fetchAPI('/api/v1/admin/monitors'),
                fetchAPI('/api/v1/admin/alert-rules')
            ]);

            if (usersRes.ok && monitorsRes.ok && alertsRes.ok) {
                const users = await usersRes.json();
                const monitors = await monitorsRes.json();
                const alerts = await alertsRes.json();

                stats = {
                    users: users.length,
                    monitors: monitors.length,
                    alertRules: alerts.length
                };
            } else {
                error = 'Failed to load system statistics';
            }
        } catch (err: any) {
             error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    });

    const quickLinks = [
        { name: 'Manage Users', path: '/admin/users', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z', color: 'bg-blue-500', href: '/admin/users' },
        { name: 'System Monitors', path: '/admin/monitors', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z', color: 'bg-emerald-500', href: '/admin/monitors' },
        { name: 'System Alert Rules', path: '/admin/alert-rules', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9', color: 'bg-orange-500', href: '/admin/alert-rules' }
    ];
</script>

<svelte:head>
	<title>Admin Dashboard - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">System Administration</h1>
    <p class="text-gray-500 dark:text-gray-400 mb-8">Overview of system health and resources.</p>
    
    {#if error}
        <div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
            {error}
        </div>
    {/if}

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <!-- Users Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 relative overflow-hidden group">
            <div class="absolute -right-6 -top-6 w-24 h-24 bg-blue-500/10 rounded-full group-hover:scale-110 transition-transform duration-500"></div>
            <div class="relative">
                <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total Users</h3>
                {#if isLoading}
                    <div class="h-10 w-24 bg-gray-200 dark:bg-slate-700 animate-pulse rounded mt-1"></div>
                {:else}
                    <p class="text-4xl font-extrabold text-gray-900 dark:text-white mt-2">{stats.users}</p>
                {/if}
            </div>
            <div class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-700 flex justify-between items-center">
                <a href="/admin/users" class="text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300">View Details →</a>
                <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg text-blue-600 dark:text-blue-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
                </div>
            </div>
        </div>

        <!-- Monitors Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 relative overflow-hidden group">
            <div class="absolute -right-6 -top-6 w-24 h-24 bg-emerald-500/10 rounded-full group-hover:scale-110 transition-transform duration-500"></div>
            <div class="relative">
                <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total Monitors</h3>
                {#if isLoading}
                    <div class="h-10 w-24 bg-gray-200 dark:bg-slate-700 animate-pulse rounded mt-1"></div>
                {:else}
                    <p class="text-4xl font-extrabold text-gray-900 dark:text-white mt-2">{stats.monitors}</p>
                {/if}
            </div>
            <div class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-700 flex justify-between items-center">
                <a href="/admin/monitors" class="text-sm font-medium text-emerald-600 dark:text-emerald-400 hover:text-emerald-800 dark:hover:text-emerald-300">View Details →</a>
                <div class="p-2 bg-emerald-100 dark:bg-emerald-900/30 rounded-lg text-emerald-600 dark:text-emerald-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
                </div>
            </div>
        </div>

        <!-- Alerts Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 relative overflow-hidden group">
            <div class="absolute -right-6 -top-6 w-24 h-24 bg-orange-500/10 rounded-full group-hover:scale-110 transition-transform duration-500"></div>
            <div class="relative">
                <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Alert Rules</h3>
                {#if isLoading}
                    <div class="h-10 w-24 bg-gray-200 dark:bg-slate-700 animate-pulse rounded mt-1"></div>
                {:else}
                    <p class="text-4xl font-extrabold text-gray-900 dark:text-white mt-2">{stats.alertRules}</p>
                {/if}
            </div>
            <div class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-700 flex justify-between items-center">
                <a href="/admin/alert-rules" class="text-sm font-medium text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-300">View Details →</a>
                <div class="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg text-orange-600 dark:text-orange-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg>
                </div>
            </div>
        </div>
    </div>

    <h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Quick Actions</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each quickLinks as link}
            <a href={link.href} class="flex items-center p-4 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl shadow-sm hover:shadow-md transition-shadow">
                <div class="p-3 rounded-lg text-white {link.color}">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={link.icon}></path></svg>
                </div>
                <div class="ml-4">
                    <h3 class="text-base font-semibold text-gray-900 dark:text-white">{link.name}</h3>
                    <p class="text-sm text-gray-500 dark:text-gray-400">Manage {link.name.toLowerCase()}</p>
                </div>
            </a>
        {/each}
    </div>
</div>
