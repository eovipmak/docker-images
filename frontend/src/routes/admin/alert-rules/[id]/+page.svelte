<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
    import { fade } from 'svelte/transition';

    let ruleId = $page.params.id;
    let rule: any = null;
    let isLoading = true;
    let error = '';

    onMount(() => {
        loadRule();
    });

    async function loadRule() {
        isLoading = true;
        try {
            const response = await fetchAPI(`/api/v1/admin/alert-rules/${ruleId}`);
            if (response.ok) {
                rule = await response.json();
            } else {
                error = 'Failed to load alert rule details';
            }
        } catch (err: any) {
             error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
	<title>Alert Rule Details - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8" in:fade>
    <div class="mb-6 flex items-center justify-between">
        <div>
            <div class="flex items-center gap-2 mb-1">
                <a href="/admin/alert-rules" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
                    Alert Rules
                </a>
                <span class="text-gray-300">/</span>
                <span class="text-gray-500 dark:text-gray-400">Details</span>
            </div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Alert Rule Details</h1>
        </div>
        <button on:click={loadRule} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 shadow-sm">
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
    {:else if rule}
        <div class="space-y-6">
            <!-- Main Info Card -->
            <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50 flex justify-between items-center">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Configuration</h3>
                    <span class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium {rule.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
                        {rule.enabled ? 'Active' : 'Disabled'}
                    </span>
                </div>
                <div class="px-6 py-5">
                    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-6">
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Rule Name</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white font-semibold">{rule.name}</dd>
                        </div>
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Owner ID</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                <a href="/admin/view-user/{rule.user_id}" class="text-blue-500 hover:text-blue-400 underline">#{rule.user_id}</a>
                            </dd>
                        </div>
                        <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Monitor ID</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                                {#if rule.monitor_id}
                                    <a href="/admin/monitors/{rule.monitor_id}" class="text-blue-500 hover:text-blue-400 underline">#{rule.monitor_id}</a>
                                {:else}
                                    <span class="text-gray-400">N/A</span>
                                {/if}
                            </dd>
                        </div>
                         <div class="sm:col-span-1">
                            <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created At</dt>
                            <dd class="mt-1 text-sm text-gray-900 dark:text-white">{new Date(rule.created_at).toLocaleString()}</dd>
                        </div>
                    </dl>
                </div>
            </div>

            <!-- Conditions Card -->
            <div class="bg-white dark:bg-slate-800 shadow-sm rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
                <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
                    <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Trigger Logic</h3>
                </div>
                <div class="px-6 py-5">
                    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 text-center">
                        <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-700">
                             <dt class="text-xs uppercase tracking-wider font-bold text-gray-500 dark:text-gray-400 mb-1">Trigger Type</dt>
                             <dd class="text-lg font-medium text-brand-orange">{rule.trigger_type}</dd>
                        </div>
                        <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-700">
                             <dt class="text-xs uppercase tracking-wider font-bold text-gray-500 dark:text-gray-400 mb-1">Threshold Value</dt>
                             <dd class="text-lg font-medium text-brand-orange">{rule.threshold_value}</dd>
                        </div>
                    </div>
                </div>
            </div>
            
        </div>
    {:else}
        <div class="text-center py-12 bg-white dark:bg-slate-800 rounded-lg border border-gray-200 dark:border-gray-700">
            <p class="text-gray-500 dark:text-gray-400">Alert rule not found</p>
            <a href="/admin/alert-rules" class="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline">Return to list</a>
        </div>
    {/if}
</div>
