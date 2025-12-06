<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

    let alertRules: any[] = [];
    let isLoading = true;
    let error = '';

    onMount(() => {
        loadAlertRules();
    });

    async function loadAlertRules() {
        isLoading = true;
        error = '';
        try {
            const response = await fetchAPI('/api/v1/admin/alert-rules');
            if (response.ok) {
                const data = await response.json();
                alertRules = data || [];
            } else {
                error = 'Failed to load alert rules';
            }
        } catch (err: any) {
            error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
	<title>System Alert Rules - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex justify-between items-center mb-6">
        <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">System Alert Rules</h1>
            <p class="text-gray-500 dark:text-gray-400">View all alert rules across the system.</p>
        </div>
        <button on:click={loadAlertRules} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
        </button>
    </div>

    {#if error}
        <div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
            {error}
        </div>
    {/if}

    <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-black ring-opacity-5 dark:ring-slate-700 rounded-lg overflow-hidden">
        <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-300 dark:divide-slate-700">
                <thead class="bg-gray-50 dark:bg-slate-900">
                    <tr>
                        <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white sm:pl-6">ID</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Name</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Condition</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Threshold</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">User ID</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Status</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-slate-700 bg-white dark:bg-slate-800">
                    {#if isLoading}
                        {#each Array(5) as _}
                            <tr>
                                <td class="py-4 pl-4 pr-3 sm:pl-6"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-8 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-32 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-32 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-16 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-12 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-16 animate-pulse"></div></td>
                            </tr>
                        {/each}
                    {:else if alertRules.length === 0}
                         <tr>
                            <td colspan="6" class="px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400">No alert rules found</td>
                        </tr>
                    {:else}
                        {#each alertRules as rule}
                            <tr>
                                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">{rule.id}</td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900 dark:text-white">{rule.name}</td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                    {rule.monitor_type} {rule.condition_type}
                                </td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                    {rule.threshold}
                                </td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">{rule.user_id}</td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {rule.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
                                        {rule.enabled ? 'Enabled' : 'Disabled'}
                                    </span>
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>
