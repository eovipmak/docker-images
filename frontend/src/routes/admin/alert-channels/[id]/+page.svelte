<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
    import { fade } from 'svelte/transition';

    let channelId = $page.params.id;
    let channel: any = null;
    let isLoading = true;
    let error = '';

    onMount(() => {
        loadChannel();
    });

    async function loadChannel() {
        isLoading = true;
        try {
            const response = await fetchAPI(`/api/v1/admin/alert-channels/${channelId}`);
            if (response.ok) {
                channel = await response.json();
            } else {
                error = 'Failed to load alert channel details';
            }
        } catch (err: any) {
             error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
	<title>Alert Channel Details - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8" in:fade>
    <div class="mb-6 flex items-center justify-between">
        <div>
            <div class="flex items-center gap-2 mb-1">
                <a href="/admin/alert-channels" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
                    Alert Channels
                </a>
                <span class="text-gray-300">/</span>
                <span class="text-gray-500 dark:text-gray-400">Details</span>
            </div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Alert Channel Details</h1>
        </div>
        <button on:click={loadChannel} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 shadow-sm">
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
    {:else if channel}
        <div class="space-y-6">
            <!-- Main Info Card -->
            <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50 flex justify-between items-center">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Configuration</h3>
                    <span class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium {channel.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
                        {channel.enabled ? 'Active' : 'Disabled'}
                    </span>
                </div>
                <div class="px-6 py-5">
                    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-6">
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Channel Name</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white font-semibold">{channel.name}</dd>
                        </div>
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Owner ID</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                <a href="/admin/users" class="text-blue-500 hover:text-blue-400 underline">#{channel.user_id}</a>
                            </dd>
                        </div>
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Type</dt>
                             <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                <span class="uppercase font-mono bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded text-xs">{channel.type}</span>
                            </dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created At</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">{new Date(channel.created_at).toLocaleString()}</dd>
                        </div>
                    </dl>
                </div>
            </div>

            <!-- Config Card -->
            <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Channel Configuration</h3>
                </div>
                <div class="px-6 py-5">
                    <pre class="bg-gray-50 dark:bg-slate-900 p-4 rounded-lg overflow-x-auto text-sm text-gray-800 dark:text-gray-300 font-mono">{JSON.stringify(channel.config, null, 2)}</pre>
                </div>
            </div>
            
        </div>
    {:else}
        <div class="text-center py-12 bg-white dark:bg-slate-800 rounded-lg border border-gray-200 dark:border-gray-700">
            <p class="text-gray-500 dark:text-gray-400">Alert channel not found</p>
            <a href="/admin/alert-channels" class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline">Return to list</a>
        </div>
    {/if}
</div>
