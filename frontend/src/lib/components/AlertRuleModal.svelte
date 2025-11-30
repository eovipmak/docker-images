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
	let lastMonitorId: string | null = null;

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
		lastMonitorId = monitorId;
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
		lastMonitorId = null;
	}

	$: isEditMode = !!rule;

	$: thresholdLabel = getThresholdLabel(formData.trigger_type);
	$: thresholdHelp = getThresholdHelp(formData.trigger_type);

	// Get selected monitor info
	$: selectedMonitor = formData.monitor_id ? monitors.find(m => m.id === formData.monitor_id) : null;

	// Get available trigger types based on selected monitor
	$: availableTriggerTypes = getAvailableTriggerTypes(selectedMonitor);

	// Handle monitor change - reset trigger type if needed
	function handleMonitorChange() {
		if (formData.monitor_id !== lastMonitorId && selectedMonitor) {
			lastMonitorId = formData.monitor_id;
			if (formData.trigger_type && !availableTriggerTypes.includes(formData.trigger_type)) {
				formData.trigger_type = availableTriggerTypes[0] || 'down';
				formData.threshold_value = getDefaultThreshold(formData.trigger_type);
			}
		}
	}

	// Watch for monitor changes
	$: if (selectedMonitor !== undefined) {
		handleMonitorChange();
	}

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

	// Get available trigger types based on selected monitor
	function getAvailableTriggerTypes(selectedMonitor: any): string[] {
		const baseTypes = ['down', 'slow_response'];
		
		// If no specific monitor selected (All monitors), allow all types
		if (!selectedMonitor) {
			return [...baseTypes, 'ssl_expiry'];
		}
		
		// If HTTP monitor, allow SSL expiry
		if (selectedMonitor.type === 'http') {
			return [...baseTypes, 'ssl_expiry'];
		}
		
		// If TCP monitor, only allow down and slow_response
		return baseTypes;
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

		// Validate SSL expiry rules
		if (formData.trigger_type === 'ssl_expiry') {
			const selectedMonitor = formData.monitor_id ? monitors.find(m => m.id === formData.monitor_id) : null;
			
			// If a specific TCP monitor is selected, SSL expiry is not allowed
			if (selectedMonitor && selectedMonitor.type === 'tcp') {
				errors.trigger_type = 'SSL Expiry rules cannot be created for TCP monitors';
			}
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
		class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
		on:click={handleBackdropClick}
		on:keydown={(e) => { if (e.key === 'Escape') closeModal(); }}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		tabindex="-1"
	>
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto ring-1 ring-slate-900/5 dark:ring-slate-700">
			<div class="px-6 py-4 border-b border-slate-100 dark:border-slate-700 flex justify-between items-center bg-slate-50/50 dark:bg-slate-950/40">
				<h2 id="modal-title" class="text-lg font-semibold text-slate-900 dark:text-gray-100">
					{isEditMode ? 'Edit Alert Rule' : 'Create Alert Rule'}
				</h2>
				<button
					type="button"
					on:click={closeModal}
					class="text-slate-400 hover:text-slate-500 dark:hover:text-slate-300 transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={handleSubmit} class="p-6 space-y-6">
				{#if errors.submit}
					<div class="rounded-md bg-red-50 dark:bg-red-900/20 p-4 border border-red-200 dark:border-red-800">
						<div class="flex">
							<div class="flex-shrink-0">
								<svg class="h-5 w-5 text-red-400 dark:text-red-300" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
									<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
								</svg>
							</div>
							<div class="ml-3">
								<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
								<div class="mt-2 text-sm text-red-700 dark:text-red-300">
									<p>{errors.submit}</p>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Name -->
				<div>
					<label for="name" class="block text-sm font-medium leading-6 text-slate-900 dark:text-slate-200">
						Name <span class="text-red-500">*</span>
					</label>
					<div class="mt-2">
						<input
							type="text"
							id="name"
							bind:value={formData.name}
							class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 dark:text-slate-100 dark:bg-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-700 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
							placeholder="e.g., Website Down Alert"
						/>
					</div>
					{#if errors.name}
						<p class="mt-2 text-sm text-red-600 dark:text-red-400">{errors.name}</p>
					{/if}
				</div>

				<!-- Monitor -->
				<div>
					<label for="monitor_id" class="block text-sm font-medium leading-6 text-slate-900 dark:text-slate-200">
						Monitor
					</label>
					<div class="mt-2">
						<select
							id="monitor_id"
							bind:value={formData.monitor_id}
							class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 dark:text-slate-100 dark:bg-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
						>
							<option value="">All monitors</option>
							{#each monitors as monitor}
								<option value={monitor.id}>{monitor.name}</option>
							{/each}
						</select>
					</div>
					<p class="mt-2 text-sm text-slate-500 dark:text-slate-400">Leave empty to apply to all monitors.</p>
				</div>

				<div class="grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-2">
					<!-- Trigger Type -->
					<div>
						<label for="trigger_type" class="block text-sm font-medium leading-6 text-slate-900 dark:text-slate-200">
							Trigger Type <span class="text-red-500">*</span>
						</label>
						<div class="mt-2">
							<select
								id="trigger_type"
								bind:value={formData.trigger_type}
								class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 dark:text-slate-100 dark:bg-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-700 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
								on:change={() => {
									formData.threshold_value = getDefaultThreshold(formData.trigger_type);
								}}
							>
								{#each availableTriggerTypes as triggerType}
									<option value={triggerType}>
										{triggerType === 'down' ? 'Down' : 
										 triggerType === 'slow_response' ? 'Slow Response' : 
										 triggerType === 'ssl_expiry' ? 'SSL Expiry' : triggerType}
									</option>
								{/each}
							</select>
						</div>
						{#if errors.trigger_type}
							<p class="mt-2 text-sm text-red-600 dark:text-red-400">{errors.trigger_type}</p>
						{/if}
					</div>

					<!-- Threshold Value -->
					<div>
						<label for="threshold_value" class="block text-sm font-medium leading-6 text-slate-900 dark:text-slate-200">
							{thresholdLabel} <span class="text-red-500">*</span>
						</label>
						<div class="mt-2">
							<input
								type="number"
								id="threshold_value"
								bind:value={formData.threshold_value}
								min="0"
								class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 dark:text-slate-100 dark:bg-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-700 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
							/>
						</div>
						{#if errors.threshold_value}
							<p class="mt-2 text-sm text-red-600 dark:text-red-400">{errors.threshold_value}</p>
						{:else if thresholdHelp}
							<p class="mt-2 text-sm text-slate-500 dark:text-slate-400">{thresholdHelp}</p>
						{/if}
					</div>
				</div>

				<!-- Alert Channels -->
				<div class="border-t border-slate-200 dark:border-slate-700 pt-6">
					<h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-slate-100 mb-4">Alert Channels</h3>
					{#if channels.length === 0}
						<p class="text-sm text-slate-500 dark:text-slate-400">No alert channels configured yet.</p>
					{:else}
						<div class="space-y-3 max-h-48 overflow-y-auto rounded-md border border-slate-200 dark:border-slate-700 p-4 bg-slate-50 dark:bg-slate-900/50">
							{#each channels as channel}
								<div class="relative flex items-start">
									<div class="flex h-6 items-center">
										<input
											id="channel-{channel.id}"
											type="checkbox"
											checked={formData.channel_ids.includes(channel.id)}
											on:change={() => toggleChannel(channel.id)}
											class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-800"
										/>
									</div>
									<div class="ml-3 text-sm leading-6">
										<label for="channel-{channel.id}" class="font-medium text-slate-900 dark:text-slate-200">
											{channel.name}
											<span class="font-normal text-slate-500 dark:text-slate-400">({channel.type})</span>
										</label>
									</div>
								</div>
							{/each}
						</div>
						<p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
							Selected: {formData.channel_ids.length} channel{formData.channel_ids.length !== 1 ? 's' : ''}
						</p>
					{/if}
				</div>

				<!-- Enabled -->
				<div class="border-t border-slate-200 dark:border-slate-700 pt-6">
					<div class="relative flex gap-x-3">
						<div class="flex h-6 items-center">
							<input
								id="enabled"
								name="enabled"
								type="checkbox"
								bind:checked={formData.enabled}
								class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-800"
							/>
						</div>
						<div class="text-sm leading-6">
							<label for="enabled" class="font-medium text-slate-900 dark:text-slate-200">Enable this alert rule</label>
							<p class="text-slate-500 dark:text-slate-400">If disabled, this rule will not trigger any alerts.</p>
						</div>
					</div>
				</div>

				<!-- Actions -->
				<div class="mt-6 flex items-center justify-end gap-x-6 border-t border-slate-200 dark:border-slate-700 pt-6">
					<button
						type="button"
						on:click={closeModal}
						class="text-sm font-semibold leading-6 text-slate-900 dark:text-slate-200 hover:text-slate-700 dark:hover:text-slate-400"
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Saving...' : isEditMode ? 'Update Rule' : 'Create Rule'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
