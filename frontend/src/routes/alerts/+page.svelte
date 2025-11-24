<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

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

	// Lazy loaded modal components
	let AlertRuleModal: any = null;
	let AlertChannelModal: any = null;
	let ruleModalLoaded = false;
	let channelModalLoaded = false;

	onMount(() => {
		loadData();
	});

	async function loadRuleModal() {
		if (!ruleModalLoaded) {
			try {
				const module = await import('$lib/components/AlertRuleModal.svelte');
				AlertRuleModal = module.default;
				ruleModalLoaded = true;
			} catch (err) {
				console.error('Failed to load AlertRuleModal:', err);
			}
		}
	}

	async function loadChannelModal() {
		if (!channelModalLoaded) {
			try {
				const module = await import('$lib/components/AlertChannelModal.svelte');
				AlertChannelModal = module.default;
				channelModalLoaded = true;
			} catch (err) {
				console.error('Failed to load AlertChannelModal:', err);
			}
		}
	}

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
	async function handleCreateRule() {
		await loadRuleModal();
		selectedRule = null;
		isRuleModalOpen = true;
	}

	async function handleEditRule(rule: any) {
		await loadRuleModal();
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
	async function handleCreateChannel() {
		await loadChannelModal();
		selectedChannel = null;
		isChannelModalOpen = true;
	}

	async function handleEditChannel(channel: any) {
		await loadChannelModal();
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

<div class="container mx-auto px-4 py-8">
<div class="max-w-7xl mx-auto">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-3xl font-bold text-gray-900 mb-2">Alerts</h1>
			<p class="text-gray-600">Manage alert rules and notification channels</p>
		</div>
	</div>

	<!-- Tabs -->
	<div class="border-b border-gray-200 mb-6">
		<nav class="-mb-px flex space-x-8">
			<button
				on:click={() => (activeTab = 'rules')}
				class="border-b-2 py-4 px-1 text-sm font-medium transition-colors {activeTab === 'rules'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
			>
				Alert Rules
			</button>
			<button
				on:click={() => (activeTab = 'channels')}
				class="border-b-2 py-4 px-1 text-sm font-medium transition-colors {activeTab ===
				'channels'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
			>
				Alert Channels
			</button>
		</nav>
	</div>

	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
			{error}
		</div>
	{:else if activeTab === 'rules'}
		<!-- Alert Rules Tab -->
		<div class="space-y-4">
			<div class="flex justify-between items-center">
				<p class="text-sm text-gray-600">
					{alertRules.length} rule{alertRules.length !== 1 ? 's' : ''}
				</p>
				<button
					on:click={handleCreateRule}
					class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors font-medium"
				>
					Create Rule
				</button>
			</div>

			<div class="bg-white rounded-lg shadow-md overflow-hidden">
				{#if alertRules.length === 0}
					<div class="text-center py-12 px-4">
						<p class="text-gray-500 mb-2">No alert rules configured yet</p>
						<p class="text-sm text-gray-400">Create your first rule to start receiving alerts</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-gray-200">
							<thead class="bg-gray-50">
								<tr>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Name
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Monitor
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Trigger Type
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Threshold
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Channels
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Status
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Actions
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-gray-200">
								{#each alertRules as rule (rule.id)}
									<tr class="hover:bg-gray-50 transition-colors">
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm font-medium text-gray-900">{rule.name}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm text-gray-900">{getMonitorName(rule.monitor_id)}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm text-gray-900">{getTriggerTypeLabel(rule.trigger_type)}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm text-gray-900">{rule.threshold_value}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm text-gray-900">
												{rule.channel_ids?.length || 0} channel{rule.channel_ids?.length !== 1
													? 's'
													: ''}
											</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<button
												on:click={() => handleToggleRuleEnabled(rule)}
												class="inline-flex items-center"
											>
												<span
													class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {rule.enabled
														? 'bg-green-100 text-green-800'
														: 'bg-gray-100 text-gray-800'}"
												>
													{rule.enabled ? 'Enabled' : 'Disabled'}
												</span>
											</button>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
											<button
												on:click={() => handleEditRule(rule)}
												class="text-blue-600 hover:text-blue-900 mr-3"
												title="Edit"
											>
												Edit
											</button>
											<button
												on:click={() => handleDeleteRule(rule)}
												class="text-red-600 hover:text-red-900"
												title="Delete"
											>
												Delete
											</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<!-- Alert Channels Tab -->
		<div class="space-y-4">
			<div class="flex justify-between items-center">
				<p class="text-sm text-gray-600">
					{alertChannels.length} channel{alertChannels.length !== 1 ? 's' : ''}
				</p>
				<button
					on:click={handleCreateChannel}
					class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors font-medium"
				>
					Create Channel
				</button>
			</div>

			<div class="bg-white rounded-lg shadow-md overflow-hidden">
				{#if alertChannels.length === 0}
					<div class="text-center py-12 px-4">
						<p class="text-gray-500 mb-2">No alert channels configured yet</p>
						<p class="text-sm text-gray-400">Create your first channel to receive notifications</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-gray-200">
							<thead class="bg-gray-50">
								<tr>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Name
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Type
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Configuration
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Status
									</th>
									<th
										scope="col"
										class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
									>
										Actions
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-gray-200">
								{#each alertChannels as channel (channel.id)}
									<tr class="hover:bg-gray-50 transition-colors">
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm font-medium text-gray-900">{channel.name}</div>
										</td>
										<td class="px-6 py-4 whitespace-nowrap">
											<div class="text-sm text-gray-900">{getChannelTypeLabel(channel.type)}</div>
										</td>
										<td class="px-6 py-4">
											<div class="text-sm text-gray-500 truncate max-w-xs">
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
												class="inline-flex items-center"
											>
												<span
													class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {channel.enabled
														? 'bg-green-100 text-green-800'
														: 'bg-gray-100 text-gray-800'}"
												>
													{channel.enabled ? 'Enabled' : 'Disabled'}
												</span>
											</button>
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
											<button
												on:click={() => handleTestChannel(channel)}
												class="text-purple-600 hover:text-purple-900 mr-3"
												title="Test Channel"
											>
												Test
											</button>
											<button
												on:click={() => handleEditChannel(channel)}
												class="text-blue-600 hover:text-blue-900 mr-3"
												title="Edit"
											>
												Edit
											</button>
											<button
												on:click={() => handleDeleteChannel(channel)}
												class="text-red-600 hover:text-red-900"
												title="Delete"
											>
												Delete
											</button>
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
</div>

{#if ruleModalLoaded && AlertRuleModal}
	<svelte:component
		this={AlertRuleModal}
		bind:isOpen={isRuleModalOpen}
		rule={selectedRule}
		on:save={handleRuleModalSave}
		on:close={handleRuleModalClose}
	/>
{/if}

{#if channelModalLoaded && AlertChannelModal}
	<svelte:component
		this={AlertChannelModal}
		bind:isOpen={isChannelModalOpen}
		channel={selectedChannel}
		on:save={handleChannelModalSave}
		on:close={handleChannelModalClose}
	/>
{/if}
