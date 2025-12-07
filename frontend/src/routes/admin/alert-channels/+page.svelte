<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { fade } from 'svelte/transition';
    import { authStore } from '$lib/stores/auth';
    import { fetchAPI } from '$lib/api/client';
    import { goto } from '$app/navigation';

    let channels: any[] = [];
    let isLoading = true;
    let error = '';
    let unsubscribe: () => void;

    // Subscribe to auth store to ensure user is admin
    $: if ($authStore.isAuthenticated && $authStore.currentUser?.role !== 'admin') {
         // This is handled by main page logic usually, but here for safety
    }

    onMount(() => {
        unsubscribe = authStore.subscribe(state => {
           if (state.isAuthenticated && state.currentUser?.role === 'admin') {
               loadChannels();
           }
        });
        
         // Initial check if store is already populated
        if ($authStore.isAuthenticated && $authStore.currentUser?.role === 'admin') {
            loadChannels();
        }
    });
    
    onDestroy(() => {
        if (unsubscribe) unsubscribe();
    });

    async function loadChannels() {
        try {
            const response = await fetchAPI('/api/v1/admin/alert-channels');
            if (response.ok) {
                channels = await response.json() || [];
            } else {
                error = 'Failed to load alert channels';
            }
        } catch (err: any) {
            error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }
</script>

<svelte:head>
    <title>System Alert Channels - Admin - V-Insight</title>
</svelte:head>

<div class="px-4 sm:px-6 lg:px-8 py-8" in:fade>
    <div class="sm:flex sm:items-center">
        <div class="sm:flex-auto">
            <h1 class="text-xl font-semibold text-gray-900 dark:text-white">System Alert Channels</h1>
            <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">View all alert channels across the system.</p>
        </div>
        <div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
             <button on:click={loadChannels} class="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto">
                Refresh
            </button>
        </div>
    </div>
    
    <div class="mt-8 flex flex-col">
        <div class="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
            <div class="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                    {#if isLoading}
                         <div class="p-12 text-center text-gray-500 dark:text-gray-400">Loading alert channels...</div>
                    {:else if error}
                        <div class="p-12 text-center text-red-500">{error}</div>
                    {:else if channels.length === 0}
                         <div class="p-12 text-center text-gray-500 dark:text-gray-400">No alert channels found</div>
                    {:else}
                        <table class="min-w-full divide-y divide-gray-300 dark:divide-slate-700">
                            <thead class="bg-gray-50 dark:bg-slate-800">
                                <tr>
                                    <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white sm:pl-6">ID</th>
                                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Name</th>
                                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Type</th>
                                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">User ID</th>
                                    <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Status</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-200 dark:divide-slate-700 bg-white dark:bg-slate-800">
                                {#each channels as channel}
                                    <tr class="hover:bg-gray-50 dark:hover:bg-slate-700/50 cursor-pointer" on:click={() => goto(`/admin/alert-channels/${channel.id}`)}>
                                        <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">{channel.id}</td>
                                        <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-900 dark:text-white">{channel.name}</td>
                                        <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                            <span class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300">
                                                {channel.type}
                                            </span>
                                        </td>
                                        <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                            <a href="/admin/users/{channel.user_id}" on:click|stopPropagation class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300">
                                                {channel.user_id}
                                            </a>
                                        </td>
                                        <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                            <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {channel.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'}">
                                                {channel.enabled ? 'Enabled' : 'Disabled'}
                                            </span>
                                        </td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    {/if}
                </div>
            </div>
        </div>
    </div>
</div>
