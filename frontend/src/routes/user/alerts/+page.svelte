<script lang="ts">
    import { onMount } from 'svelte';
    import { fetchAPI } from '$lib/api/client';
    import AlertRuleModal from '$lib/components/AlertRuleModal.svelte';
    import AlertChannelModal from '$lib/components/AlertChannelModal.svelte';
    import AlertCard from '$lib/components/AlertCard.svelte';
    import Card from '$lib/components/Card.svelte';
    import ConfirmModal from '$lib/components/ConfirmModal.svelte';

    type Tab = 'rules' | 'channels';

    let activeTab: Tab = 'rules';
    let alertRules: any[] = [];
    let alertChannels: any[] = [];
    let monitors: any[] = [];
    let isLoading = true;
    let error = '';

    let isRuleModalOpen = false;
    let isChannelModalOpen = false;
    let selectedRule: any = null;
    let selectedChannel: any = null;

    // Search and sort
    let searchQuery = '';
    let sortBy = 'name';
    let sortDirection: 'asc' | 'desc' = 'asc';

    // Confirm modal state
    let isConfirmModalOpen = false;
    let confirmTitle = '';
    let confirmMessage = '';
    let onConfirmCallback: (() => void) | null = null;

    function applySort(a: any, b: any) {
        let cmp = 0;
        if (sortBy === 'name') {
            cmp = a.name.localeCompare(b.name);
        } else if (sortBy === 'status') {
            cmp = (a.enabled ? 1 : 0) - (b.enabled ? 1 : 0);
        }
        return sortDirection === 'asc' ? cmp : -cmp;
    }

    $: filteredRules = alertRules
        .filter(rule => rule.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
                        getTriggerTypeLabel(rule.trigger_type).toLowerCase().includes(searchQuery.toLowerCase()))
        .sort(applySort);

    $: filteredChannels = alertChannels
        .filter(channel => channel.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
                           getChannelTypeLabel(channel.type).toLowerCase().includes(searchQuery.toLowerCase()))
        .sort(applySort);

    onMount(() => {
        loadData();
    });

    async function loadData() {
        isLoading = true;
        error = '';

        try {
            const [rulesResponse, channelsResponse, monitorsResponse] = await Promise.all([
                fetchAPI('/api/v1/alert-rules'),
                fetchAPI('/api/v1/alert-channels'),
                fetchAPI('/api/v1/monitors')
            ]);

            if (!rulesResponse.ok || !channelsResponse.ok || !monitorsResponse.ok) {
                throw new Error('Failed to load data');
            }

            alertRules = await rulesResponse.json();
            alertChannels = await channelsResponse.json();
            monitors = await monitorsResponse.json();
        } catch (err: any) {
            console.error('Error loading data:', err);
            error = err.message || 'Failed to load data';
        } finally {
            isLoading = false;
        }
    }

    // Alert Rules handlers
    function handleCreateRule() {
        selectedRule = null;
        isRuleModalOpen = true;
    }

    function handleEditRule(rule: any) {
        selectedRule = rule;
        isRuleModalOpen = true;
    }

    async function handleDeleteRule(rule: any) {
        confirmTitle = 'Delete Alert Rule';
        confirmMessage = `Are you sure you want to delete "${rule.name}"?`;
        onConfirmCallback = async () => {
            try {
                const response = await fetchAPI(`/api/v1/alert-rules/${rule.id}`, {
                    method: 'DELETE'
                });

                if (!response.ok) {
                    throw new Error('Failed to delete alert rule');
                }

                await loadData();
            } catch (err: any) {
                console.error('Error deleting alert rule:', err);
                alert(err.message || 'Failed to delete alert rule');
            }
        };
        isConfirmModalOpen = true;
    }

    async function handleToggleRuleEnabled(rule: any) {
        try {
            const response = await fetchAPI(`/api/v1/alert-rules/${rule.id}`, {
                method: 'PUT',
                body: JSON.stringify({ enabled: !rule.enabled })
            });

            if (!response.ok) {
                throw new Error('Failed to toggle alert rule');
            }

            await loadData();
        } catch (err: any) {
            console.error('Error toggling alert rule:', err);
            alert(err.message || 'Failed to toggle alert rule');
        }
    }

    // Alert Channels handlers
    function handleCreateChannel() {
        selectedChannel = null;
        isChannelModalOpen = true;
    }

    function handleEditChannel(channel: any) {
        selectedChannel = channel;
        isChannelModalOpen = true;
    }

    async function handleDeleteChannel(channel: any) {
        confirmTitle = 'Delete Alert Channel';
        confirmMessage = `Are you sure you want to delete "${channel.name}"?`;
        onConfirmCallback = async () => {
            try {
                const response = await fetchAPI(`/api/v1/alert-channels/${channel.id}`, {
                    method: 'DELETE'
                });

                if (!response.ok) {
                    throw new Error('Failed to delete alert channel');
                }

                await loadData();
            } catch (err: any) {
                console.error('Error deleting alert channel:', err);
                alert(err.message || 'Failed to delete alert channel');
            }
        };
        isConfirmModalOpen = true;
    }

    function handleConfirmDelete() {
        if (onConfirmCallback) {
            onConfirmCallback();
        }
        isConfirmModalOpen = false;
    }

    function handleCancelDelete() {
        isConfirmModalOpen = false;
    }

    async function handleToggleChannelEnabled(channel: any) {
        try {
            const response = await fetchAPI(`/api/v1/alert-channels/${channel.id}`, {
                method: 'PUT',
                body: JSON.stringify({ enabled: !channel.enabled })
            });

            if (!response.ok) {
                throw new Error('Failed to toggle alert channel');
            }

            await loadData();
        } catch (err: any) {
            console.error('Error toggling alert channel:', err);
            alert(err.message || 'Failed to toggle alert channel');
        }
    }

    async function handleTestChannel(channel: any) {
        try {
            const response = await fetchAPI(`/api/v1/alert-channels/${channel.id}/test`, {
                method: 'POST'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to test alert channel');
            }

            alert('Test notification sent successfully!');
        } catch (err: any) {
            console.error('Error testing alert channel:', err);
            alert(err.message || 'Failed to test alert channel');
        }
    }

    function handleRuleModalSave() {
        isRuleModalOpen = false;
        selectedRule = null;
        loadData();
    }

    function handleRuleModalClose() {
        isRuleModalOpen = false;
        selectedRule = null;
    }

    function handleChannelModalSave() {
        isChannelModalOpen = false;
        selectedChannel = null;
        loadData();
    }

    function handleChannelModalClose() {
        isChannelModalOpen = false;
        selectedChannel = null;
    }

    function getMonitorName(monitorId: any): string {
        if (!monitorId) return 'All monitors';

        let id = monitorId;
        if (typeof monitorId === 'object' && 'String' in monitorId) {
            if (!monitorId.Valid) return 'All monitors';
            id = monitorId.String;
        }

        const monitor = monitors.find((m) => m.id === id);
        return monitor ? monitor.name : 'Unknown';
    }

    function getTriggerTypeLabel(triggerType: string): string {
        switch (triggerType) {
            case 'down':
                return 'Down';
            case 'slow_response':
                return 'Slow Response';
            case 'ssl_expiry':
                return 'SSL Expiry';
            default:
                return triggerType;
        }
    }

    function getChannelTypeLabel(type: string): string {
        switch (type) {
            case 'webhook':
                return 'Webhook';
            case 'discord':
                return 'Discord';
            case 'email':
                return 'Email';
            default:
                return type;
        }
    }
</script>

<svelte:head>
    <title>Alerts - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
            <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100">Alerts</h1>
            <p class="mt-1 text-sm text-slate-600 dark:text-slate-400">Manage alert rules and notification channels to stay informed about your infrastructure.</p>
        </div>
        <div class="flex items-center gap-3">
            <button
                on:click={() => (activeTab = 'rules')}
                class={`px-3 py-2 text-sm font-medium rounded-lg border ${activeTab === 'rules' ? 'border-blue-200 text-blue-700 bg-blue-50 dark:border-blue-900/50 dark:bg-blue-900/20' : 'border-transparent text-slate-600 dark:text-slate-300 hover:border-slate-200 dark:hover:border-slate-700'}`}
            >
                Rules
            </button>
            <button
                on:click={() => (activeTab = 'channels')}
                class={`px-3 py-2 text-sm font-medium rounded-lg border ${activeTab === 'channels' ? 'border-blue-200 text-blue-700 bg-blue-50 dark:border-blue-900/50 dark:bg-blue-900/20' : 'border-transparent text-slate-600 dark:text-slate-300 hover:border-slate-200 dark:hover:border-slate-700'}`}
            >
                Channels
            </button>
            <button
                on:click={activeTab === 'rules' ? handleCreateRule : handleCreateChannel}
                class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
            >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                </svg>
                {activeTab === 'rules' ? 'Add Rule' : 'Add Channel'}
            </button>
        </div>
    </div>

    <Card className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="flex flex-col sm:flex-row gap-3 flex-1">
            <div class="relative max-w-md w-full">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-slate-400">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"></path>
                    </svg>
                </div>
                <input 
                    bind:value={searchQuery}
                    type="text" 
                    placeholder={`Search ${activeTab}...`} 
                    class="block w-full pl-10 pr-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg leading-5 bg-white dark:bg-slate-800 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition-shadow text-slate-900 dark:text-white"
                >
            </div>
        </div>
        <div class="flex items-center gap-3">
            <select 
                bind:value={sortBy}
                class="block w-full rounded-lg border-slate-300 dark:border-slate-600 py-2 pl-3 pr-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 sm:text-sm bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
            >
                <option value="name">Name</option>
                <option value="status">Status</option>
            </select>
            <button
                class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
                on:click={() => (sortDirection = sortDirection === 'asc' ? 'desc' : 'asc')}
                title={sortDirection === 'asc' ? 'Ascending' : 'Descending'}
            >
                {#if sortDirection === 'asc'}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h9.75m4.5-4.5v12m0 0l-3.75-3.75M17.25 21L21 17.25"></path>
                    </svg>
                {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h9.75m4.5-8.25L17.25 9m0 0L21 12.75M17.25 9v12"></path>
                    </svg>
                {/if}
            </button>
        </div>
    </Card>

    {#if isLoading}
        <div class="flex items-center justify-center py-12">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
    {:else if error}
        <div class="mt-6 rounded-md bg-red-50 p-4 border border-red-200">
            <div class="flex">
                <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
                    </svg>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800">Error loading data</h3>
                    <div class="mt-2 text-sm text-red-700">
                        <p>{error}</p>
                    </div>
                </div>
            </div>
        </div>
    {:else if activeTab === 'rules'}
        <!-- Alert Rules Tab -->
        <div class="mt-6">
            <div class="flex items-center mb-4">
                <p class="text-sm text-slate-600">
                    {filteredRules.length} rule{filteredRules.length !== 1 ? 's' : ''} configured
                </p>
            </div>

            <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden">
                {#if filteredRules.length === 0}
                    <div class="text-center py-12 px-4">
                        <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                        <h3 class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">No alert rules</h3>
                        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Get started by creating a new alert rule.</p>
                        <div class="mt-6">
                            <button
                                on:click={handleCreateRule}
                                class="inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                            >
                                <svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-11.25a.75.75 0 00-1.5 0v2.5h-2.5a.75.75 0 000 1.5h2.5v2.5a.75.75 0 001.5 0v-2.5h2.5a.75.75 0 000-1.5h-2.5v-2.5z" clip-rule="evenodd" />
                                </svg>
                                Create Rule
                            </button>
                        </div>
                    </div>
                {:else}
                    <div class="p-4">
                        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                            {#each filteredRules as rule (rule.id)}
                                <div class="group">
                                    <AlertCard {rule} on:edit={() => handleEditRule(rule)} on:delete={() => handleDeleteRule(rule)} />
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}
            </div>
        </div>
    {:else}
        <!-- Alert Channels Tab -->
        <div class="mt-6">
            <div class="flex justify-between items-center mb-4">
                <p class="text-sm text-slate-600 dark:text-slate-400">
                    {filteredChannels.length} channel{filteredChannels.length !== 1 ? 's' : ''} configured
                </p>
                <button
                    on:click={handleCreateChannel}
                    class="inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-colors"
                >
                    <svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-11.25a.75.75 0 00-1.5 0v2.5h-2.5a.75.75 0 000 1.5h2.5v2.5a.75.75 0 001.5 0v-2.5h2.5a.75.75 0 000-1.5h-2.5v-2.5z" clip-rule="evenodd" />
                    </svg>
                    Create Channel
                </button>
            </div>

            <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden">
                {#if filteredChannels.length === 0}
                    <div class="text-center py-12 px-4">
                        <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                        <h3 class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">No alert channels</h3>
                        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Get started by creating a new notification channel.</p>
                        <div class="mt-6">
                            <button
                                on:click={handleCreateChannel}
                                class="inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
                            >
                                <svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-11.25a.75.75 0 00-1.5 0v2.5h-2.5a.75.75 0 000 1.5h2.5v2.5a.75.75 0 001.5 0v-2.5h2.5a.75.75 0 000-1.5h-2.5v-2.5z" clip-rule="evenodd" />
                                </svg>
                                Create Channel
                            </button>
                        </div>
                    </div>
                {:else}
                    <!-- Mobile Card View -->
                    <div class="block md:hidden divide-y divide-slate-200 dark:divide-slate-700">
                        {#each filteredChannels as channel (channel.id)}
                            <div class="p-4">
                                <div class="flex items-start justify-between mb-2">
                                    <div>
                                        <div class="text-sm font-medium text-slate-900 dark:text-white">{channel.name}</div>
                                        <span class="inline-flex items-center mt-1 rounded-md bg-slate-50 dark:bg-slate-700/50 px-2 py-0.5 text-xs font-medium text-slate-700 dark:text-slate-300">
                                            {getChannelTypeLabel(channel.type)}
                                        </span>
                                    </div>
                                    <button
                                        on:click={() => handleToggleChannelEnabled(channel)}
                                        class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {channel.enabled ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400' : 'bg-slate-50 dark:bg-slate-700/50 text-slate-600 dark:text-slate-400'}"
                                    >
                                        {channel.enabled ? 'Enabled' : 'Disabled'}
                                    </button>
                                </div>
                                <div class="text-xs text-slate-500 dark:text-slate-400 truncate mb-3">
                                    {#if channel.type === 'webhook'}
                                        {channel.config.url || 'N/A'}
                                    {:else if channel.type === 'discord'}
                                        {channel.config.webhook_url || 'N/A'}
                                    {:else if channel.type === 'email'}
                                        {channel.config.to || 'N/A'}
                                    {:else}
                                        N/A
                                    {/if}
                                </div>
                                <div class="flex gap-4 text-sm">
                                    <button
                                        on:click={() => handleTestChannel(channel)}
                                        class="text-purple-600 dark:text-purple-400"
                                    >
                                        Test
                                    </button>
                                    <button
                                        on:click={() => handleEditChannel(channel)}
                                        class="text-blue-600 dark:text-blue-400"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        on:click={() => handleDeleteChannel(channel)}
                                        class="text-red-600 dark:text-red-400"
                                    >
                                        Delete
                                    </button>
                                </div>
                            </div>
                        {/each}
                    </div>

                    <!-- Desktop Table View -->
                    <div class="hidden md:block overflow-x-auto">
                        <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
                            <thead class="bg-slate-50 dark:bg-slate-900/50">
                                <tr>
                                    <th scope="col" class="px-4 lg:px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Name</th>
                                    <th scope="col" class="px-4 lg:px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Type</th>
                                    <th scope="col" class="px-4 lg:px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider hidden lg:table-cell">Configuration</th>
                                    <th scope="col" class="px-4 lg:px-6 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Status</th>
                                    <th scope="col" class="relative px-4 lg:px-6 py-3">
                                        <span class="sr-only">Actions</span>
                                    </th>
                                </tr>
                            </thead>
                            <tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-200 dark:divide-slate-700">
                                {#each filteredChannels as channel (channel.id)}
                                    <tr class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
                                        <td class="px-4 lg:px-6 py-4 whitespace-nowrap text-sm font-medium text-slate-900 dark:text-white">{channel.name}</td>
                                        <td class="px-4 lg:px-6 py-4 whitespace-nowrap text-sm text-slate-600 dark:text-slate-300">
                                            <span class="inline-flex items-center rounded-md bg-slate-50 dark:bg-slate-700/50 px-2 py-1 text-xs font-medium text-slate-700 dark:text-slate-300 ring-1 ring-inset ring-slate-600/20 dark:ring-slate-500/30">
                                                {getChannelTypeLabel(channel.type)}
                                            </span>
                                        </td>
                                        <td class="px-4 lg:px-6 py-4 text-sm text-slate-500 dark:text-slate-400 hidden lg:table-cell">
                                            <div class="truncate max-w-xs">
                                                {#if channel.type === 'webhook'}
                                                    {channel.config.url || 'N/A'}
                                                {:else if channel.type === 'discord'}
                                                    {channel.config.webhook_url || 'N/A'}
                                                {:else if channel.type === 'email'}
                                                    {channel.config.to || 'N/A'}
                                                {:else}
                                                    N/A
                                                {/if}
                                            </div>
                                        </td>
                                        <td class="px-4 lg:px-6 py-4 whitespace-nowrap">
                                            <button
                                                on:click={() => handleToggleChannelEnabled(channel)}
                                                class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {channel.enabled ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 ring-1 ring-inset ring-green-600/20 dark:ring-green-500/30 hover:bg-green-100 dark:hover:bg-green-900/50' : 'bg-slate-50 dark:bg-slate-700/50 text-slate-600 dark:text-slate-400 ring-1 ring-inset ring-slate-500/10 dark:ring-slate-500/30 hover:bg-slate-100 dark:hover:bg-slate-700/70'}"
                                            >
                                                {channel.enabled ? 'Enabled' : 'Disabled'}
                                            </button>
                                        </td>
                                        <td class="px-4 lg:px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            <div class="flex justify-end gap-2 lg:gap-3">
                                                <button
                                                    on:click={() => handleTestChannel(channel)}
                                                    class="text-purple-600 dark:text-purple-400 hover:text-purple-900 dark:hover:text-purple-300 transition-colors"
                                                >
                                                    Test
                                                </button>
                                                <button
                                                    on:click={() => handleEditChannel(channel)}
                                                    class="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300 transition-colors"
                                                >
                                                    Edit
                                                </button>
                                                <button
                                                    on:click={() => handleDeleteChannel(channel)}
                                                    class="text-red-600 dark:text-red-400 hover:text-red-900 dark:hover:text-red-300 transition-colors"
                                                >
                                                    Delete
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>

<AlertRuleModal
    bind:isOpen={isRuleModalOpen}
    rule={selectedRule}
    on:save={handleRuleModalSave}
    on:close={handleRuleModalClose}
/> 

<AlertChannelModal
    bind:isOpen={isChannelModalOpen}
    channel={selectedChannel}
    on:save={handleChannelModalSave}
    on:close={handleChannelModalClose}
/> 

<ConfirmModal
    isOpen={isConfirmModalOpen}
    title={confirmTitle}
    message={confirmMessage}
    on:confirm={handleConfirmDelete}
    on:cancel={handleCancelDelete}
/>
