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

	let config: Record<string, any> = { url: '' };

	let errors: Record<string, string> = {};
	let isSubmitting = false;
	let lastChannelId: string | null = null;

	// Only update formData when channel actually changes (different channel or switching between create/edit)
	$: if (channel && channel.id !== lastChannelId) {
		formData = {
			name: channel.name || '',
			type: channel.type || 'webhook',
			enabled: channel.enabled !== undefined ? channel.enabled : true
		};
		config = channel.config || {};
		lastChannelId = channel?.id || null;
	} else if (!channel && lastChannelId !== null) {
		// Switching from edit to create mode
		formData = {
			name: '',
			type: 'webhook',
			enabled: true
		};
		config = { url: '' };
		lastChannelId = null;
	}

	$: isEditMode = !!channel;

	// Reset config when type changes (only in create mode)
	function handleTypeChange() {
		if (!isEditMode) {
			config = getDefaultConfig(formData.type);
		}
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
			} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(config.to)) {
				errors.config = 'Invalid email address';
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
			const endpoint = isEditMode ? `/api/v1/alert-channels/${channel.id}` : '/api/v1/alert-channels';
			const method = isEditMode ? 'PUT' : 'POST';

			const payload = {
				name: formData.name,
				type: formData.type,
				config: config,
				enabled: formData.enabled
			};

			const response = await fetchAPI(endpoint, {
				method,
				body: JSON.stringify(payload)
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
		config = { url: '' };
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
	<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
	<div
		class="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
		on:click={handleBackdropClick}
		on:keydown={(e) => { if (e.key === 'Escape') closeModal(); }}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto ring-1 ring-slate-900/5">
			<div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
				<h2 id="modal-title" class="text-lg font-semibold text-slate-900">
					{isEditMode ? 'Edit Alert Channel' : 'Create Alert Channel'}
				</h2>
				<button
					type="button"
					on:click={closeModal}
					class="text-slate-400 hover:text-slate-500 transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form on:submit|preventDefault={handleSubmit} class="p-6 space-y-6">
				{#if errors.submit}
					<div class="rounded-md bg-red-50 p-4 border border-red-200">
						<div class="flex">
							<div class="flex-shrink-0">
								<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
									<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
								</svg>
							</div>
							<div class="ml-3">
								<h3 class="text-sm font-medium text-red-800">Error</h3>
								<div class="mt-2 text-sm text-red-700">
									<p>{errors.submit}</p>
								</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Name -->
				<div>
					<label for="name" class="block text-sm font-medium leading-6 text-slate-900">
						Name <span class="text-red-500">*</span>
					</label>
					<div class="mt-2">
						<input
							type="text"
							id="name"
							bind:value={formData.name}
							class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
							placeholder="e.g., DevOps Team Discord"
						/>
					</div>
					{#if errors.name}
						<p class="mt-2 text-sm text-red-600">{errors.name}</p>
					{/if}
				</div>

				<!-- Type -->
				<div>
					<label for="type" class="block text-sm font-medium leading-6 text-slate-900">
						Channel Type <span class="text-red-500">*</span>
					</label>
					<div class="mt-2">
						<select
							id="type"
							bind:value={formData.type}
							on:change={handleTypeChange}
							class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
						>
							<option value="webhook">Webhook</option>
							<option value="discord">Discord</option>
							<option value="email">Email</option>
						</select>
					</div>
					{#if errors.type}
						<p class="mt-2 text-sm text-red-600">{errors.type}</p>
					{/if}
				</div>

				<!-- Configuration -->
				<div class="border-t border-slate-200 pt-6">
					<h3 class="text-base font-semibold leading-6 text-slate-900 mb-4">Configuration</h3>
					
					{#if formData.type === 'webhook'}
						<div>
							<label for="webhook_url" class="block text-sm font-medium leading-6 text-slate-900">
								Webhook URL <span class="text-red-500">*</span>
							</label>
							<div class="mt-2">
								<input
									type="url"
									id="webhook_url"
									bind:value={config.url}
									class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
									placeholder="https://api.example.com/webhook"
								/>
							</div>
							<p class="mt-2 text-sm text-slate-500">We'll send a POST request with JSON payload to this URL.</p>
						</div>
					{:else if formData.type === 'discord'}
						<div>
							<label for="discord_url" class="block text-sm font-medium leading-6 text-slate-900">
								Discord Webhook URL <span class="text-red-500">*</span>
							</label>
							<div class="mt-2">
								<input
									type="url"
									id="discord_url"
									bind:value={config.webhook_url}
									class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
									placeholder="https://discord.com/api/webhooks/..."
								/>
							</div>
						</div>
					{:else if formData.type === 'email'}
						<div>
							<label for="email_to" class="block text-sm font-medium leading-6 text-slate-900">
								Email Address <span class="text-red-500">*</span>
							</label>
							<div class="mt-2">
								<input
									type="email"
									id="email_to"
									bind:value={config.to}
									class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
									placeholder="alerts@example.com"
								/>
							</div>
						</div>
					{/if}

					{#if errors.config}
						<p class="mt-2 text-sm text-red-600">{errors.config}</p>
					{/if}
				</div>

				<!-- Enabled -->
				<div class="border-t border-slate-200 pt-6">
					<div class="relative flex gap-x-3">
						<div class="flex h-6 items-center">
							<input
								id="enabled"
								name="enabled"
								type="checkbox"
								bind:checked={formData.enabled}
								class="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
							/>
						</div>
						<div class="text-sm leading-6">
							<label for="enabled" class="font-medium text-slate-900">Enable this channel</label>
							<p class="text-slate-500">If disabled, no notifications will be sent to this channel.</p>
						</div>
					</div>
				</div>

				<!-- Actions -->
				<div class="mt-6 flex items-center justify-end gap-x-6 border-t border-slate-200 pt-6">
					<button
						type="button"
						on:click={closeModal}
						class="text-sm font-semibold leading-6 text-slate-900 hover:text-slate-700"
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Saving...' : isEditMode ? 'Update Channel' : 'Create Channel'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
