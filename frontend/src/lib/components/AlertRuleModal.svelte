<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

	export let isOpen = false;
	export let rule: any = null;

	const dispatch = createEventDispatcher();

	interface FormData {
		name: string;
		monitor_id: string;
		trigger_type: string;
		threshold_value: number;
		enabled: boolean;
		channel_ids: string[];
	}

	let formData: FormData = {
		name: '',
		monitor_id: '',
		trigger_type: 'down',
		threshold_value: 3,
		enabled: true,
		channel_ids: []
	};

	let errors: Record<string, string> = {};
	let isSubmitting = false;
	let monitors: any[] = [];
	let channels: any[] = [];
	let isLoadingData = false;
	let lastRuleId: string | null = null;

	// Only update formData when rule actually changes (different rule or switching between create/edit)
	$: if (rule && rule.id !== lastRuleId) {
		let monitorId = '';
		if (rule.monitor_id && typeof rule.monitor_id === 'object' && 'String' in rule.monitor_id) {
			monitorId = rule.monitor_id.Valid ? rule.monitor_id.String : '';
		} else {
			monitorId = rule.monitor_id || '';
		}

		formData = {
			name: rule.name || '',
			monitor_id: monitorId,
			trigger_type: rule.trigger_type || 'down',
			threshold_value: rule.threshold_value || getDefaultThreshold(rule.trigger_type || 'down'),
			enabled: rule.enabled !== undefined ? rule.enabled : true,
			channel_ids: rule.channel_ids || []
		};
		lastRuleId = rule?.id || null;
	} else if (!rule && lastRuleId !== null) {
		// Switching from edit to create mode
		formData = {
			name: '',
			monitor_id: '',
			trigger_type: 'down',
			threshold_value: 3,
			enabled: true,
			channel_ids: []
		};
		lastRuleId = null;
	}

	$: isEditMode = !!rule;

	$: thresholdLabel = getThresholdLabel(formData.trigger_type);
	$: thresholdHelp = getThresholdHelp(formData.trigger_type);

	$: if (isOpen) {
		loadData();
	}

	// Get default threshold based on trigger type
	function getDefaultThreshold(triggerType: string): number {
		switch (triggerType) {
			case 'down':
				return 3;
			case 'slow_response':
				return 5000;
			case 'ssl_expiry':
				return 30;
			default:
				return 0;
		}
	}

	// Update threshold when trigger type changes


	async function loadData() {
		if (isLoadingData) return;
		isLoadingData = true;

		try {
			const [monitorsResponse, channelsResponse] = await Promise.all([
				fetchAPI('/api/v1/monitors'),
				fetchAPI('/api/v1/alert-channels')
			]);

			if (monitorsResponse.ok) {
				monitors = await monitorsResponse.json();
			}

			if (channelsResponse.ok) {
				channels = await channelsResponse.json();
			}
		} catch (error) {
			console.error('Error loading data:', error);
		} finally {
			isLoadingData = false;
		}
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Name is required';
		}

		if (!formData.trigger_type) {
			errors.trigger_type = 'Trigger type is required';
		}

		if (formData.threshold_value < 0) {
			errors.threshold_value = 'Threshold value must be positive';
		}

		// Validate threshold based on trigger type
		if (formData.trigger_type === 'down' && formData.threshold_value < 1) {
			errors.threshold_value = 'Must be at least 1 failed check';
		}

		if (formData.trigger_type === 'slow_response' && formData.threshold_value < 100) {
			errors.threshold_value = 'Must be at least 100ms';
		}

		if (formData.trigger_type === 'ssl_expiry' && formData.threshold_value < 1) {
			errors.threshold_value = 'Must be at least 1 day';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		isSubmitting = true;

		try {
			const endpoint = isEditMode ? `/api/v1/alert-rules/${rule.id}` : '/api/v1/alert-rules';
			const method = isEditMode ? 'PUT' : 'POST';

			const payload: any = {
				name: formData.name,
				trigger_type: formData.trigger_type,
				threshold_value: formData.threshold_value,
				enabled: formData.enabled,
				channel_ids: formData.channel_ids
			};

			// Handle monitor_id - send null for "All monitors"
			let monitorIdToSend = null;
			if (formData.monitor_id) {
				if (typeof formData.monitor_id === 'object' && formData.monitor_id !== null && 'String' in formData.monitor_id) {
					// @ts-ignore
					monitorIdToSend = formData.monitor_id.Valid ? formData.monitor_id.String : null;
				} else {
					monitorIdToSend = formData.monitor_id;
				}
			}
			payload.monitor_id = monitorIdToSend;

			const response = await fetchAPI(endpoint, {
				method,
				body: JSON.stringify(payload)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to save alert rule');
			}

			const savedRule = await response.json();
			dispatch('save', savedRule);
			closeModal();
		} catch (error: any) {
			errors.submit = error.message || 'An error occurred';
		} finally {
			isSubmitting = false;
		}
	}

	function closeModal() {
		isOpen = false;
		formData = {
			name: '',
			monitor_id: '',
			trigger_type: 'down',
			threshold_value: 3,
			enabled: true,
			channel_ids: []
		};
		errors = {};
		dispatch('close');
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeModal();
		}
	}

	function toggleChannel(channelId: string) {
		const index = formData.channel_ids.indexOf(channelId);
		if (index === -1) {
			formData.channel_ids = [...formData.channel_ids, channelId];
		} else {
			formData.channel_ids = formData.channel_ids.filter((id) => id !== channelId);
		}
	}

	function getThresholdLabel(triggerType?: string): string {
		const type = triggerType || formData.trigger_type;
		switch (type) {
			case 'down':
				return 'Failed Checks';
			case 'slow_response':
				return 'Response Time (ms)';
			case 'ssl_expiry':
				return 'Days Before Expiry';
			default:
				return 'Threshold';
		}
	}

	function getThresholdHelp(triggerType?: string): string {
		const type = triggerType || formData.trigger_type;
		switch (type) {
			case 'down':
				return 'Number of consecutive failed checks before triggering';
			case 'slow_response':
				return 'Response time threshold in milliseconds';
			case 'ssl_expiry':
				return 'Alert when SSL certificate expires within this many days';
			default:
				return '';
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
		on:click={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
			<div class="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
				<h2 id="modal-title" class="text-2xl font-bold text-gray-900">
					{isEditMode ? 'Edit Alert Rule' : 'Create Alert Rule'}
				</h2>
				<button
					type="button"
					on:click={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={handleSubmit} class="p-6 space-y-4">
				{#if errors.submit}
					<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
						{errors.submit}
					</div>
				{/if}

				<!-- Name -->
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 mb-1">
						Name <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="name"
						bind:value={formData.name}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Website Down Alert"
					/>
					{#if errors.name}
						<p class="text-sm text-red-600 mt-1">{errors.name}</p>
					{/if}
				</div>

				<!-- Monitor -->
				<div>
					<label for="monitor_id" class="block text-sm font-medium text-gray-700 mb-1">
						Monitor
					</label>
					<select
						id="monitor_id"
						bind:value={formData.monitor_id}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All monitors</option>
						{#each monitors as monitor}
							<option value={monitor.id}>{monitor.name}</option>
						{/each}
					</select>
					<p class="text-xs text-gray-500 mt-1">Leave empty to apply to all monitors</p>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<!-- Trigger Type -->
					<div>
						<label for="trigger_type" class="block text-sm font-medium text-gray-700 mb-1">
							Trigger Type <span class="text-red-500">*</span>
						</label>
						<select
							id="trigger_type"
							bind:value={formData.trigger_type}
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
							on:change={() => {
								formData.threshold_value = getDefaultThreshold(formData.trigger_type);
							}}
						>
							<option value="down">Down</option>
							<option value="slow_response">Slow Response</option>
							<option value="ssl_expiry">SSL Expiry</option>
						</select>
						{#if errors.trigger_type}
							<p class="text-sm text-red-600 mt-1">{errors.trigger_type}</p>
						{/if}
					</div>

					<!-- Threshold Value -->
					<div>
						<label for="threshold_value" class="block text-sm font-medium text-gray-700 mb-1">
							{thresholdLabel} <span class="text-red-500">*</span>
						</label>
						<input
							type="number"
							id="threshold_value"
							bind:value={formData.threshold_value}
							min="0"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
						{#if errors.threshold_value}
							<p class="text-sm text-red-600 mt-1">{errors.threshold_value}</p>
						{:else if thresholdHelp}
							<p class="text-xs text-gray-500 mt-1">{thresholdHelp}</p>
						{/if}
					</div>
				</div>

				<!-- Alert Channels -->
				<div class="border-t border-gray-200 pt-4">
					<h3 class="text-lg font-medium text-gray-900 mb-3">Alert Channels</h3>
					{#if channels.length === 0}
						<p class="text-sm text-gray-500">No alert channels configured yet</p>
					{:else}
						<div class="space-y-2 max-h-48 overflow-y-auto border border-gray-200 rounded-md p-3">
							{#each channels as channel}
								<div class="flex items-center">
									<input
										type="checkbox"
										id="channel-{channel.id}"
										checked={formData.channel_ids.includes(channel.id)}
										on:change={() => toggleChannel(channel.id)}
										class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
									/>
									<label for="channel-{channel.id}" class="ml-2 block text-sm text-gray-700">
										{channel.name}
										<span class="text-gray-500">({channel.type})</span>
									</label>
								</div>
							{/each}
						</div>
						<p class="text-xs text-gray-500 mt-1">
							Selected: {formData.channel_ids.length} channel{formData.channel_ids.length !== 1
								? 's'
								: ''}
						</p>
					{/if}
				</div>

				<!-- Enabled -->
				<div class="border-t border-gray-200 pt-4">
					<div class="flex items-center">
						<input
							type="checkbox"
							id="enabled"
							bind:checked={formData.enabled}
							class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
						/>
						<label for="enabled" class="ml-2 block text-sm text-gray-700">
							Enable this alert rule
						</label>
					</div>
				</div>

				<!-- Actions -->
				<div class="flex justify-end gap-3 pt-4 border-t border-gray-200">
					<button
						type="button"
						on:click={closeModal}
						class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors"
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Saving...' : isEditMode ? 'Update Rule' : 'Create Rule'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
