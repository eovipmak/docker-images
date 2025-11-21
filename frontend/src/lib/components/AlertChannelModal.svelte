<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

	export let isOpen = false;
	export let channel: any = null;

	const dispatch = createEventDispatcher();

	interface FormData {
		name: string;
		type: string;
		enabled: boolean;
	}

	let formData: FormData = {
		name: '',
		type: 'webhook',
		enabled: true
	};

	let config: Record<string, any> = {};

	let errors: Record<string, string> = {};
	let isSubmitting = false;

	$: if (channel) {
		formData = {
			name: channel.name || '',
			type: channel.type || 'webhook',
			enabled: channel.enabled !== undefined ? channel.enabled : true
		};
		config = channel.config || {};
	}

	$: isEditMode = !!channel;

	// Reset config when type changes (only in create mode)
	$: if (formData.type && !isEditMode) {
		config = getDefaultConfig(formData.type);
	}

	function getDefaultConfig(type: string): Record<string, any> {
		switch (type) {
			case 'webhook':
				return { url: '' };
			case 'discord':
				return { webhook_url: '' };
			case 'email':
				return { to: '' };
			default:
				return {};
		}
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Name is required';
		}

		if (!formData.type) {
			errors.type = 'Type is required';
		}

		// Validate config based on type
		if (formData.type === 'webhook') {
			if (!config.url || !config.url.trim()) {
				errors.config = 'Webhook URL is required';
			} else {
				try {
					new URL(config.url);
				} catch {
					errors.config = 'Invalid URL format';
				}
			}
		}

		if (formData.type === 'discord') {
			if (!config.webhook_url || !config.webhook_url.trim()) {
				errors.config = 'Discord Webhook URL is required';
			} else {
				try {
					const url = new URL(config.webhook_url);
					if (!url.hostname.includes('discord.com')) {
						errors.config = 'Must be a valid Discord webhook URL';
					}
				} catch {
					errors.config = 'Invalid URL format';
				}
			}
		}

		if (formData.type === 'email') {
			if (!config.to || !config.to.trim()) {
				errors.config = 'Email address is required';
			} else {
				const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
				if (!emailRegex.test(config.to)) {
					errors.config = 'Invalid email address';
				}
			}
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		isSubmitting = true;

		try {
			const endpoint = isEditMode
				? `/api/v1/alert-channels/${channel.id}`
				: '/api/v1/alert-channels';
			const method = isEditMode ? 'PUT' : 'POST';

			const response = await fetchAPI(endpoint, {
				method,
				body: JSON.stringify({
					...formData,
					config
				})
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to save alert channel');
			}

			const savedChannel = await response.json();
			dispatch('save', savedChannel);
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
			type: 'webhook',
			enabled: true
		};
		config = {};
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
			<div class="px-6 py-4 border-b border-gray-200">
				<h2 id="modal-title" class="text-2xl font-bold text-gray-900">
					{isEditMode ? 'Edit Alert Channel' : 'Create Alert Channel'}
				</h2>
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
						placeholder="My Webhook Channel"
					/>
					{#if errors.name}
						<p class="text-sm text-red-600 mt-1">{errors.name}</p>
					{/if}
				</div>

				<!-- Type -->
				<div>
					<label for="type" class="block text-sm font-medium text-gray-700 mb-1">
						Type <span class="text-red-500">*</span>
					</label>
					<select
						id="type"
						bind:value={formData.type}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						disabled={isEditMode}
					>
						<option value="webhook">Webhook</option>
						<option value="discord">Discord</option>
						<option value="email">Email</option>
					</select>
					{#if errors.type}
						<p class="text-sm text-red-600 mt-1">{errors.type}</p>
					{/if}
					{#if isEditMode}
						<p class="text-xs text-gray-500 mt-1">Type cannot be changed after creation</p>
					{/if}
				</div>

				<!-- Configuration -->
				<div class="border-t border-gray-200 pt-4">
					<h3 class="text-lg font-medium text-gray-900 mb-3">Configuration</h3>

					{#if formData.type === 'webhook'}
						<div>
							<label for="webhook_url" class="block text-sm font-medium text-gray-700 mb-1">
								Webhook URL <span class="text-red-500">*</span>
							</label>
							<input
								type="url"
								id="webhook_url"
								bind:value={config.url}
								class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								placeholder="https://example.com/webhook"
							/>
							<p class="text-xs text-gray-500 mt-1">
								The URL where alert notifications will be sent via POST request
							</p>
						</div>
					{:else if formData.type === 'discord'}
						<div>
							<label for="discord_webhook_url" class="block text-sm font-medium text-gray-700 mb-1">
								Discord Webhook URL <span class="text-red-500">*</span>
							</label>
							<input
								type="url"
								id="discord_webhook_url"
								bind:value={config.webhook_url}
								class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								placeholder="https://discord.com/api/webhooks/..."
							/>
							<p class="text-xs text-gray-500 mt-1">
								Get this from your Discord channel settings → Integrations → Webhooks
							</p>
						</div>
					{:else if formData.type === 'email'}
						<div>
							<label for="email_to" class="block text-sm font-medium text-gray-700 mb-1">
								Email Address <span class="text-red-500">*</span>
							</label>
							<input
								type="email"
								id="email_to"
								bind:value={config.to}
								class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
								placeholder="alerts@example.com"
							/>
							<p class="text-xs text-gray-500 mt-1">Email notifications will be sent to this address</p>
						</div>
					{/if}

					{#if errors.config}
						<p class="text-sm text-red-600 mt-1">{errors.config}</p>
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
							Enable this channel
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
						{isSubmitting ? 'Saving...' : isEditMode ? 'Update Channel' : 'Create Channel'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
