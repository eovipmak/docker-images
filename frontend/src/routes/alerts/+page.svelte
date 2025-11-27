<script lang="ts">
    import { onMount } from 'svelte';
    import { fetchAPI } from '$lib/api/client';
    import AlertRuleModal from '$lib/components/AlertRuleModal.svelte';
    import AlertChannelModal from '$lib/components/AlertChannelModal.svelte';
    import AlertCard from '$lib/components/AlertCard.svelte';

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
        if (!confirm(`Are you sure you want to delete "${rule.name}"?`)) {
            return;
        }

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
        if (!confirm(`Are you sure you want to delete "${channel.name}"?`)) {
            return;
        }

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

<div class="px-4 sm:px-6 lg:px-8 py-8">
    <div class="sm:flex sm:items-center">
        <div class="sm:flex-auto">
            <h1 class="text-2xl font-semibold leading-6 text-slate-900">Alerts</h1>
            <p class="mt-2 text-sm text-slate-600">Manage alert rules and notification channels to stay informed about your infrastructure.</p>
        </div>
    </div>

    <!-- Tabs -->
    <div class="mt-6 border-b border-slate-200">
        <nav class="-mb-px flex space-x-8" aria-label="Tabs">
            <button
                on:click={() => (activeTab = 'rules')}
                class="{activeTab === 'rules'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700'} whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium transition-colors"
            >
                Alert Rules
            </button>
            <button
                on:click={() => (activeTab = 'channels')}
                class="{activeTab === 'channels'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700'} whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium transition-colors"
            >
                Alert Channels
            </button>
        </nav>
    </div>

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
            <div class="flex justify-between items-center mb-4">
                <p class="text-sm text-slate-600">
                    {alertRules.length} rule{alertRules.length !== 1 ? 's' : ''} configured
                </p>
                <button
                    on:click={handleCreateRule}
                    class="inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-colors"
                >
                    <svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-11.25a.75.75 0 00-1.5 0v2.5h-2.5a.75.75 0 000 1.5h2.5v2.5a.75.75 0 001.5 0v-2.5h2.5a.75.75 0 000-1.5h-2.5v-2.5z" clip-rule="evenodd" />
                    </svg>
                    Create Rule
                </button>
            </div>

            <div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
                {#if alertRules.length === 0}
                    <div class="text-center py-12 px-4">
                        <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                        <h3 class="mt-2 text-sm font-semibold text-slate-900">No alert rules</h3>
                        <p class="mt-1 text-sm text-slate-500">Get started by creating a new alert rule.</p>
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
                            {#each alertRules as rule (rule.id)}
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
                <p class="text-sm text-slate-600">
                    {alertChannels.length} channel{alertChannels.length !== 1 ? 's' : ''} configured
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

            <div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
                {#if alertChannels.length === 0}
                    <div class="text-center py-12 px-4">
                        <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                        </svg>
                        <h3 class="mt-2 text-sm font-semibold text-slate-900">No alert channels</h3>
                        <p class="mt-1 text-sm text-slate-500">Get started by creating a new notification channel.</p>
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
                    <div class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-slate-200">
                            <thead class="bg-slate-50">
                                <tr>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Type</th>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Configuration</th>
                                    <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
                                    <th scope="col" class="relative px-6 py-3">
                                        <span class="sr-only">Actions</span>
                                    </th>
                                </tr>
                            </thead>
                            <tbody class="bg-white divide-y divide-slate-200">
                                {#each alertChannels as channel (channel.id)}
                                    <tr class="hover:bg-slate-50 transition-colors">
                                        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-slate-900">{channel.name}</td>
                                        <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
                                            <span class="inline-flex items-center rounded-md bg-slate-50 px-2 py-1 text-xs font-medium text-slate-700 ring-1 ring-inset ring-slate-600/20">
                                                {getChannelTypeLabel(channel.type)}
                                            </span>
                                        </td>
                                        <td class="px-6 py-4 text-sm text-slate-500">
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
                                        <td class="px-6 py-4 whitespace-nowrap">
                                            <button
                                                on:click={() => handleToggleChannelEnabled(channel)}
                                                class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {channel.enabled ? 'bg-green-50 text-green-700 ring-1 ring-inset ring-green-600/20 hover:bg-green-100' : 'bg-slate-50 text-slate-600 ring-1 ring-inset ring-slate-500/10 hover:bg-slate-100'}"
                                            >
                                                {channel.enabled ? 'Enabled' : 'Disabled'}
                                            </button>
                                        </td>
                                        <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                            <div class="flex justify-end gap-3">
                                                <button
                                                    on:click={() => handleTestChannel(channel)}
                                                    class="text-purple-600 hover:text-purple-900 transition-colors"
                                                >
                                                    Test
                                                </button>
                                                <button
                                                    on:click={() => handleEditChannel(channel)}
                                                    class="text-blue-600 hover:text-blue-900 transition-colors"
                                                >
                                                    Edit
                                                </button>
                                                <button
                                                    on:click={() => handleDeleteChannel(channel)}
                                                    class="text-red-600 hover:text-red-900 transition-colors"
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
<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import AlertRuleModal from '$lib/components/AlertRuleModal.svelte';
	import AlertChannelModal from '$lib/components/AlertChannelModal.svelte';

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
		if (!confirm(`Are you sure you want to delete "${rule.name}"?`)) {
			return;
		}

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
		if (!confirm(`Are you sure you want to delete "${channel.name}"?`)) {
			return;
		}

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
					<div class="p-4">
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
				return 'Webhook';
								<div class="group">
									<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-4 hover:shadow-md transition-colors flex flex-col h-full">
										<div class="flex items-start justify-between gap-3 mb-3">
											<div class="min-w-0">
												<h3 class="text-sm font-semibold text-slate-900 truncate">{rule.name}</h3>
												<p class="text-xs text-slate-500 truncate">{getTriggerTypeLabel(rule.trigger_type)} • {getMonitorName(rule.monitor_id)}</p>
											</div>
											<div class="text-xs text-slate-500">{rule.enabled ? 'Enabled' : 'Disabled'}</div>
										</div>
										<div class="mt-auto pt-2 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
											<div>Threshold: {rule.threshold_value}</div>
											<div class="flex gap-2">
												<button on:click={() => handleEditRule(rule)} class="text-blue-600 hover:text-blue-900">Edit</button>
												<button on:click={() => handleDeleteRule(rule)} class="text-red-600 hover:text-red-900">Delete</button>
											</div>
										</div>
									</div>
								</div>
					: 'border-transparent text-slate-500 hover:border-slate-300 hover:text-slate-700'} whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium transition-colors"
			>
				Alert Channels
			</button>
		</nav>
	</div>

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
			<div class="flex justify-between items-center mb-4">
				<p class="text-sm text-slate-600">
					{alertRules.length} rule{alertRules.length !== 1 ? 's' : ''} configured
				</p>
				<button
					on:click={handleCreateRule}
					class="inline-flex items-center rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-colors"
				>
					<svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm.75-11.25a.75.75 0 00-1.5 0v2.5h-2.5a.75.75 0 000 1.5h2.5v2.5a.75.75 0 001.5 0v-2.5h2.5a.75.75 0 000-1.5h-2.5v-2.5z" clip-rule="evenodd" />
					</svg>
					Create Rule
				</button>
			</div>

			<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
				{#if alertRules.length === 0}
					<div class="text-center py-12 px-4">
						<svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<h3 class="mt-2 text-sm font-semibold text-slate-900">No alert rules</h3>
						<p class="mt-1 text-sm text-slate-500">Get started by creating a new alert rule.</p>
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
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
							<thead class="bg-slate-50">
								<tr>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Monitor</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Trigger</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Threshold</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Channels</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
									<th scope="col" class="relative px-6 py-3">
										<span class="sr-only">Actions</span>
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-slate-200">
								{#each alertRules as rule (rule.id)}
									<tr class="hover:bg-slate-50 transition-colors">
										<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-slate-900">{rule.name}</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">{getMonitorName(rule.monitor_id)}</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
											<span class="inline-flex items-center rounded-md bg-slate-50 px-2 py-1 text-xs font-medium text-slate-700 ring-1 ring-inset ring-slate-600/20">
												{getTriggerTypeLabel(rule.trigger_type)}
											</span>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">{rule.threshold_value}</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
											{rule.channel_ids?.length || 0} channel{rule.channel_ids?.length !== 1 ? 's' : ''}
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<button
												on:click={() => handleToggleRuleEnabled(rule)}
												class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {rule.enabled ? 'bg-green-50 text-green-700 ring-1 ring-inset ring-green-600/20 hover:bg-green-100' : 'bg-slate-50 text-slate-600 ring-1 ring-inset ring-slate-500/10 hover:bg-slate-100'}"
											>
												{rule.enabled ? 'Enabled' : 'Disabled'}
											</button>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
											<div class="flex justify-end gap-3">
												<button
													on:click={() => handleEditRule(rule)}
													class="text-blue-600 hover:text-blue-900 transition-colors"
												>
													Edit
												</button>
												<button
													on:click={() => handleDeleteRule(rule)}
													class="text-red-600 hover:text-red-900 transition-colors"
												>
													Delete
												</button>
											</div>
										</td>
									</tr>
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
				<p class="text-sm text-slate-600">
					{alertChannels.length} channel{alertChannels.length !== 1 ? 's' : ''} configured
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

			<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
				{#if alertChannels.length === 0}
					<div class="text-center py-12 px-4">
						<svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<h3 class="mt-2 text-sm font-semibold text-slate-900">No alert channels</h3>
						<p class="mt-1 text-sm text-slate-500">Get started by creating a new notification channel.</p>
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
					<div class="p-4">
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
							<thead class="bg-slate-50">
								<tr>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Type</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Configuration</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
									<th scope="col" class="relative px-6 py-3">
										<span class="sr-only">Actions</span>
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-slate-200">
								{#each alertChannels as channel (channel.id)}
									<tr class="hover:bg-slate-50 transition-colors">
										<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-slate-900">{channel.name}</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
											<span class="inline-flex items-center rounded-md bg-slate-50 px-2 py-1 text-xs font-medium text-slate-700 ring-1 ring-inset ring-slate-600/20">
												{getChannelTypeLabel(channel.type)}
											</span>
										</td>
										<td class="px-6 py-4 text-sm text-slate-500">
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
										<td class="px-6 py-4 whitespace-nowrap">
											<button
												on:click={() => handleToggleChannelEnabled(channel)}
												class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {channel.enabled ? 'bg-green-50 text-green-700 ring-1 ring-inset ring-green-600/20 hover:bg-green-100' : 'bg-slate-50 text-slate-600 ring-1 ring-inset ring-slate-500/10 hover:bg-slate-100'}"
											>
												{channel.enabled ? 'Enabled' : 'Disabled'}
											</button>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
											<div class="flex justify-end gap-3">
												<button
													on:click={() => handleTestChannel(channel)}
													class="text-purple-600 hover:text-purple-900 transition-colors"
												>
													Test
												</button>
												<button
													on:click={() => handleEditChannel(channel)}
													class="text-blue-600 hover:text-blue-900 transition-colors"
												>
													Edit
												</button>
												<button
													on:click={() => handleDeleteChannel(channel)}
													class="text-red-600 hover:text-red-900 transition-colors"
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
