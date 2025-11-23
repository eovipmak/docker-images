<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

	export let isOpen = false;
	export let monitor: any = null;

	const dispatch = createEventDispatcher();

	interface FormData {
		name: string;
		url: string;
		check_interval: number;
		timeout: number;
		enabled: boolean;
		check_ssl: boolean;
		ssl_alert_days: number;
	}

	let formData: FormData = {
		name: '',
		url: '',
		check_interval: 300,
		timeout: 30,
		enabled: true,
		check_ssl: true,
		ssl_alert_days: 30
	};

	let errors: Record<string, string> = {};
	let isSubmitting = false;
	let lastMonitorId: string | null = null;

	// Only update formData when monitor actually changes (different monitor or switching between create/edit)
	$: if (monitor && monitor.id !== lastMonitorId) {
		formData = {
			name: monitor.name || '',
			url: monitor.url || '',
			check_interval: monitor.check_interval || 300,
			timeout: monitor.timeout || 30,
			enabled: monitor.enabled !== undefined ? monitor.enabled : true,
			check_ssl: monitor.check_ssl !== undefined ? monitor.check_ssl : true,
			ssl_alert_days: monitor.ssl_alert_days || 30
		};
		lastMonitorId = monitor?.id || null;
	} else if (!monitor && lastMonitorId !== null) {
		// Switching from edit to create mode
		formData = {
			name: '',
			url: '',
			check_interval: 300,
			timeout: 30,
			enabled: true,
			check_ssl: true,
			ssl_alert_days: 30
		};
		lastMonitorId = null;
	}

	$: isEditMode = !!monitor;

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Name is required';
		}

		if (!formData.url.trim()) {
			errors.url = 'URL is required';
		} else {
			try {
				new URL(formData.url);
			} catch {
				errors.url = 'Invalid URL format';
			}
		}

		if (formData.check_interval < 60) {
			errors.check_interval = 'Check interval must be at least 60 seconds';
		}

		if (formData.timeout < 5 || formData.timeout > 120) {
			errors.timeout = 'Timeout must be between 5 and 120 seconds';
		}

		if (formData.ssl_alert_days < 1) {
			errors.ssl_alert_days = 'SSL alert days must be at least 1';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		isSubmitting = true;

		try {
			const endpoint = isEditMode ? `/api/v1/monitors/${monitor.id}` : '/api/v1/monitors';
			const method = isEditMode ? 'PUT' : 'POST';

			const response = await fetchAPI(endpoint, {
				method,
				body: JSON.stringify(formData)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to save monitor');
			}

			const savedMonitor = await response.json();
			dispatch('save', savedMonitor);
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
			url: '',
			check_interval: 300,
			timeout: 30,
			enabled: true,
			check_ssl: true,
			ssl_alert_days: 30
		};
		errors = {};
		dispatch('close');
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeModal();
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
					{isEditMode ? 'Edit Monitor' : 'Add Monitor'}
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
						placeholder="My Website"
					/>
					{#if errors.name}
						<p class="text-sm text-red-600 mt-1">{errors.name}</p>
					{/if}
				</div>

				<!-- URL -->
				<div>
					<label for="url" class="block text-sm font-medium text-gray-700 mb-1">
						URL <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="url"
						bind:value={formData.url}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="https://example.com"
					/>
					{#if errors.url}
						<p class="text-sm text-red-600 mt-1">{errors.url}</p>
					{/if}
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<!-- Check Interval -->
					<div>
						<label for="check_interval" class="block text-sm font-medium text-gray-700 mb-1">
							Check Interval (seconds)
						</label>
						<input
							type="number"
							id="check_interval"
							bind:value={formData.check_interval}
							min="60"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
						{#if errors.check_interval}
							<p class="text-sm text-red-600 mt-1">{errors.check_interval}</p>
						{/if}
					</div>

					<!-- Timeout -->
					<div>
						<label for="timeout" class="block text-sm font-medium text-gray-700 mb-1">
							Timeout (seconds)
						</label>
						<input
							type="number"
							id="timeout"
							bind:value={formData.timeout}
							min="5"
							max="120"
							class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						/>
						{#if errors.timeout}
							<p class="text-sm text-red-600 mt-1">{errors.timeout}</p>
						{/if}
					</div>
				</div>

				<!-- SSL Settings -->
				<div class="border-t border-gray-200 pt-4">
					<h3 class="text-lg font-medium text-gray-900 mb-3">SSL/TLS Settings</h3>

					<div class="space-y-3">
						<!-- Check SSL -->
						<div class="flex items-center">
							<input
								type="checkbox"
								id="check_ssl"
								bind:checked={formData.check_ssl}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
							<label for="check_ssl" class="ml-2 block text-sm text-gray-700">
								Check SSL Certificate
							</label>
						</div>

						<!-- SSL Alert Days -->
						{#if formData.check_ssl}
							<div>
								<label for="ssl_alert_days" class="block text-sm font-medium text-gray-700 mb-1">
									Alert Before Expiry (days)
								</label>
								<input
									type="number"
									id="ssl_alert_days"
									bind:value={formData.ssl_alert_days}
									min="1"
									class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								/>
								{#if errors.ssl_alert_days}
									<p class="text-sm text-red-600 mt-1">{errors.ssl_alert_days}</p>
								{/if}
							</div>
						{/if}
					</div>
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
							Enable monitoring
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
						{isSubmitting ? 'Saving...' : isEditMode ? 'Update Monitor' : 'Create Monitor'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
