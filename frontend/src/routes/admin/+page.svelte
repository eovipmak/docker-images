<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
    import { authStore } from '$lib/stores/auth';
    import { fade } from 'svelte/transition';

    let stats = {
        users: 0,
        monitors: 0,
        alertRules: 0,
        alertChannels: 0
    };
    let isLoading = true;
    let error = '';
    
    // Auth state handling
    let email = '';
	let password = '';
    let loginError = '';
    let isLoginLoading = false;

    // Subscribe to auth store
    $: isAuthenticated = $authStore.isAuthenticated;
    $: currentUser = $authStore.currentUser;
    $: isAdmin = currentUser?.role === 'admin';

    // Load data when authenticated as admin
    $: if (isAuthenticated && isAdmin) {
        loadData();
    }

    async function loadData() {
        isLoading = true;
        try {
            const [usersRes, monitorsRes, alertsRes, channelsRes] = await Promise.all([
                fetchAPI('/api/v1/admin/users'),
                fetchAPI('/api/v1/admin/monitors'),
                fetchAPI('/api/v1/admin/alert-rules'),
                fetchAPI('/api/v1/admin/alert-channels')
            ]);

            if (usersRes.ok && monitorsRes.ok && alertsRes.ok && channelsRes.ok) {
                const users = await usersRes.json();
                const monitors = await monitorsRes.json();
                const alerts = await alertsRes.json();
                const channels = await channelsRes.json();

                stats = {
                    users: Array.isArray(users) ? users.length : 0,
                    monitors: Array.isArray(monitors) ? monitors.length : 0,
                    alertRules: Array.isArray(alerts) ? alerts.length : 0,
                    alertChannels: Array.isArray(channels) ? channels.length : 0
                };
            } else {
                error = 'Failed to load system statistics';
            }
        } catch (err: any) {
             error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }

    async function handleLogin() {
		loginError = '';
		isLoginLoading = true;

		try {
			const response = await fetch('/api/v1/auth/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					email,
					password
				})
			});

			const data = await response.json();

			if (!response.ok) {
				loginError = data.error || 'Login failed';
				return;
			}

			// Store the token and update auth store
			if (data.token) {
				await authStore.login(data.token);
                // No redirect needed, reactive statements will handle showing the dashboard
			}
		} catch (err) {
			loginError = 'An error occurred. Please try again.';
			console.error('Login error:', err);
		} finally {
			isLoginLoading = false;
		}
	}

    const quickLinks = [
        { name: 'Manage Users', path: '/admin/users', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z', color: 'bg-blue-500', href: '/admin/users' },
        { name: 'System Monitors', path: '/admin/monitors', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z', color: 'bg-emerald-500', href: '/admin/monitors' },
        { name: 'System Alert Rules', path: '/admin/alert-rules', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9', color: 'bg-orange-500', href: '/admin/alert-rules' },
        { name: 'System Alert Channels', path: '/admin/alert-channels', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9', color: 'bg-purple-500', href: '/admin/alert-channels' }
    ];
</script>

<svelte:head>
	<title>Admin Dashboard - V-Insight</title>
</svelte:head>

{#if !isAuthenticated}
    <!-- Login Form -->
    <div class="min-h-screen flex items-center justify-center bg-dark-950/50 py-12 px-4 sm:px-6 lg:px-8 relative overflow-hidden font-sans">
        <div class="max-w-md w-full space-y-8 relative z-10">
            <div class="text-center">
                <h2 class="mt-6 text-3xl font-bold tracking-widest text-white uppercase">Admin <span class="text-brand-orange">Access</span></h2>
                <p class="mt-2 text-sm text-gray-400">
                    Restricted Area. Authorized Personnel Only.
                </p>
            </div>

            <div class="bg-dark-900/50 backdrop-blur-sm py-8 px-4 shadow-2xl sm:rounded-2xl sm:px-10 border border-white/10">
                {#if loginError}
                    <div class="mb-6 p-4 rounded-lg bg-red-900/20 border border-red-500/50 text-sm text-red-300 flex items-center">
                        {loginError}
                    </div>
                {/if}

                <form class="space-y-6" on:submit|preventDefault={handleLogin}>
                    <div>
                        <label for="email" class="block text-xs font-bold uppercase tracking-wider text-brand-orange mb-2">
                            Email address
                        </label>
                        <div class="mt-1">
                            <input
                                id="email"
                                name="email"
                                type="email"
                                autocomplete="email"
                                required
                                bind:value={email}
                                disabled={isLoginLoading}
                                class="appearance-none block w-full px-4 py-3 bg-dark-950 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-orange focus:border-brand-orange sm:text-sm transition-all"
                                placeholder="admin@example.com"
                            />
                        </div>
                    </div>

                    <div>
                        <label for="password" class="block text-xs font-bold uppercase tracking-wider text-brand-orange mb-2">
                            Password
                        </label>
                        <div class="mt-1">
                            <input
                                id="password"
                                name="password"
                                type="password"
                                autocomplete="current-password"
                                required
                                bind:value={password}
                                disabled={isLoginLoading}
                                class="appearance-none block w-full px-4 py-3 bg-dark-950 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-orange focus:border-brand-orange sm:text-sm transition-all"
                                placeholder="••••••••"
                            />
                        </div>
                    </div>

                    <div>
                        <button
                            type="submit"
                            disabled={isLoginLoading}
                            class="w-full flex justify-center py-4 px-4 border border-transparent rounded-lg shadow-lg text-sm font-bold uppercase tracking-widest text-white bg-gradient-to-r from-brand-orange to-red-600 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand-orange disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                        >
                            {#if isLoginLoading}
                                <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                </svg>
                                AUTHENTICATING...
                            {:else}
                                ADMIN LOGIN
                            {/if}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
{:else if !isAdmin}
    <!-- 403 Forbidden -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8" in:fade>
		<div class="text-center">
			<div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 dark:bg-red-900/30">
				<svg class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
				</svg>
			</div>
			<h1 class="mt-2 text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-5xl">403 Forbidden</h1>
			<p class="mt-2 text-base text-gray-500 dark:text-gray-400">You don't have permission to access the admin area.</p>
			<div class="mt-6">
				<a href="/user/dashboard" class="text-base font-medium text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300">
					Go back to dashboard
					<span aria-hidden="true"> &rarr;</span>
				</a>
			</div>
		</div>
	</div>
{:else}
    <!-- Admin Dashboard -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8" in:fade>
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">System Administration</h1>
		<p class="text-gray-500 dark:text-gray-400 mb-8">Overview of system health and resources.</p>
		
		{#if error}
			<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
				{error}
			</div>
		{/if}

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {#each [
                { label: 'Total Users', value: stats.users, color: 'blue', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
                { label: 'Total Monitors', value: stats.monitors, color: 'emerald', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
                { label: 'Alert Rules', value: stats.alertRules, color: 'orange', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9' },
                { label: 'Alert Channels', value: stats.alertChannels, color: 'purple', icon: 'M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9' }
            ] as stat}
                <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 flex justify-between items-center relative overflow-hidden group">
                     <div class="absolute -right-6 -top-6 w-24 h-24 bg-{stat.color}-500/10 rounded-full group-hover:scale-110 transition-transform duration-500"></div>
                     <div class="relative z-10">
                        <h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{stat.label}</h3>
                        {#if isLoading}
                            <div class="h-8 w-16 bg-gray-200 dark:bg-slate-700 animate-pulse rounded mt-1"></div>
                        {:else}
                            <p class="text-3xl font-extrabold text-gray-900 dark:text-white mt-1">{stat.value}</p>
                        {/if}
                    </div>
                    <div class="relative z-10 p-3 bg-{stat.color}-100 dark:bg-{stat.color}-900/30 rounded-xl text-{stat.color}-600 dark:text-{stat.color}-400">
                        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={stat.icon}></path></svg>
                    </div>
                </div>
            {/each}
        </div>

		<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Quick Actions</h2>
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
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
{/if}
