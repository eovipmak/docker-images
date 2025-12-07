<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
    import { fade } from 'svelte/transition';

    let monitorId = $page.params.id;
    let monitor: any = null;
    let isLoading = true;
    let error = '';

    onMount(() => {
        loadMonitor();
    });

    async function loadMonitor() {
        isLoading = true;
        try {
            const response = await fetchAPI(`/api/v1/admin/monitors/${monitorId}`);
            if (response.ok) {
                monitor = await response.json();
            } else {
                error = 'Failed to load monitor details';
            }
        } catch (err: any) {
             error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
	<title>Monitor Details - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8" in:fade>
    <div class="mb-6 flex items-center justify-between">
        <div>
            <div class="flex items-center gap-2 mb-1">
                <a href="/admin/monitors" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
                    Monitors
                </a>
                <span class="text-gray-300">/</span>
                <span class="text-gray-500 dark:text-gray-400">Details</span>
            </div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Monitor Details</h1>
        </div>
        <button on:click={loadMonitor} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 shadow-sm">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
        </button>
    </div>

    {#if error}
        <div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
            {error}
        </div>
    {/if}

    {#if isLoading}
        <div class="animate-pulse space-y-6">
            <div class="h-64 bg-gray-200 dark:bg-slate-800 rounded-xl"></div>
             <div class="h-32 bg-gray-200 dark:bg-slate-800 rounded-xl"></div>
        </div>
    {:else if monitor}
        <div class="space-y-6">
            <!-- Main Info Card -->
            <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50 flex justify-between items-center">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Configuration</h3>
                     <span class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium {monitor.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
                        {monitor.enabled ? 'Active' : 'Disabled'}
                    </span>
                </div>
                <div class="px-6 py-5">
                    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-6">
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Monitor Name</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white font-semibold">{monitor.name}</dd>
                        </div>
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">URL</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white truncate">
                                <a href="{monitor.url}" target="_blank" class="text-blue-500 hover:text-blue-400 underline">{monitor.url}</a>
                            </dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Type</dt>
                             <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                <span class="uppercase font-mono bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded text-xs">{monitor.type}</span>
                            </dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Owner ID</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                <a href="/admin/users" class="text-blue-500 hover:text-blue-400 underline">#{monitor.user_id}</a>
                            </dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Keyword</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">{monitor.keyword || '-'}</dd>
                        </div>
                          <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Check SSL</dt>
                             <dd class="mt-1 text-sm text-gray-900 dark:text-white">{monitor.check_ssl ? 'Yes' : 'No'}</dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created At</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">{new Date(monitor.created_at).toLocaleString()}</dd>
                        </div>
                    </dl>
                </div>
            </div>
            
             <!-- Stats Check Card -->
             <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Settings</h3>
                </div>
                 <div class="px-6 py-5">
                     <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 text-center">
                        <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-700">
                             <dt class="text-xs uppercase tracking-wider font-bold text-gray-500 dark:text-gray-400 mb-1">Interval</dt>
                             <dd class="text-lg font-medium text-brand-orange">{monitor.check_interval}s</dd>
                        </div>
                         <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-700">
                             <dt class="text-xs uppercase tracking-wider font-bold text-gray-500 dark:text-gray-400 mb-1">Timeout</dt>
                             <dd class="text-lg font-medium text-brand-orange">{monitor.timeout}s</dd>
                        </div>
                     </div>
                 </div>
             </div>
        </div>
    {:else}
        <div class="text-center py-12 bg-white dark:bg-slate-800 rounded-lg border border-gray-200 dark:border-gray-700">
            <p class="text-gray-500 dark:text-gray-400">Monitor not found</p>
            <a href="/admin/monitors" class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline">Return to list</a>
        </div>
    {/if}
</div>
