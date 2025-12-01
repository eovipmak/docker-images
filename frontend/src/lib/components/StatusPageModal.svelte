<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

	export let isOpen = false;
	export let statusPage: any = null;

	const dispatch = createEventDispatcher();

	interface FormData {
		slug: string;
		name: string;
		public_enabled: boolean;
		monitor_ids: string[];
	}

	interface Monitor {
		id: string;
		name: string;
		url: string;
		type: string;
		enabled: boolean;
	}

	let formData: FormData = {
		slug: '',
		name: '',
		public_enabled: false,
		monitor_ids: []
	};

	let monitors: Monitor[] = [];
	let errors: Record<string, string> = {};
	let isSubmitting = false;
	let lastStatusPageId: string | null = null;
	let monitorsLoaded = false;

	// Load monitors when component mounts
	onMount(() => {
		console.log('StatusPageModal mounted');
	});

	// Only update formData when statusPage actually changes (different statusPage or switching between create/edit)
	$: if (statusPage && statusPage.id !== lastStatusPageId) {
		loadStatusPageMonitors(statusPage.id);
		formData = {
			slug: statusPage.slug || '',
			name: statusPage.name || '',
			public_enabled: statusPage.public_enabled || false,
			monitor_ids: []
		};
		lastStatusPageId = statusPage?.id || null;
	} else if (!statusPage && lastStatusPageId !== null) {
		// Switching from edit to create mode
		formData = {
			slug: '',
			name: '',
			public_enabled: false,
			monitor_ids: []
		};
		lastStatusPageId = null;
	}

	// Load monitors when modal opens
	$: if (isOpen) {
		console.log('Modal opened, loading monitors...');
		loadMonitors();
	}

	function closeModal() {
		dispatch('close');
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeModal();
		}
	}

	async function loadMonitors() {
		console.log('Loading monitors...');
		try {
			const response = await fetchAPI('/api/v1/monitors');
			console.log('Monitors API response status:', response.status);
			if (response.ok) {
				const data = await response.json();
				if (Array.isArray(data)) {
					monitors = data;
				} else {
					monitors = data.monitors || [];
				}
				console.log('Loaded monitors:', monitors.length, monitors);
			} else {
				console.error('Failed to load monitors - bad response:', response.status, response.statusText);
				// Try to get error details
				try {
					const errorData = await response.text();
					console.error('Error response:', errorData);
				} catch (e) {
					console.error('Could not read error response');
				}
			}
		} catch (err) {
			console.error('Failed to load monitors:', err);
		}
	}

	async function loadStatusPageMonitors(statusPageId: string) {
		try {
			const response = await fetchAPI(`/api/v1/status-pages/${statusPageId}/monitors`);
			if (response.ok) {
				const data = await response.json();
				formData.monitor_ids = (data.monitors || []).map((m: Monitor) => m.id);
				console.log('Loaded status page monitors:', formData.monitor_ids);
			} else {
				console.error('Failed to load status page monitors - bad response:', response.status);
			}
		} catch (err) {
			console.error('Failed to load status page monitors:', err);
		}
	}

	function toggleMonitor(monitorId: string) {
		if (formData.monitor_ids.includes(monitorId)) {
			formData.monitor_ids = formData.monitor_ids.filter(id => id !== monitorId);
		} else {
			formData.monitor_ids = [...formData.monitor_ids, monitorId];
		}
	}

	async function handleSubmit() {
		errors = {};
		isSubmitting = true;

		try {
			// Basic validation
			if (!formData.name.trim()) {
				errors.name = 'Name is required';
			}
			if (!formData.slug.trim()) {
				errors.slug = 'Slug is required';
			} else if (!/^[a-zA-Z0-9_-]+$/.test(formData.slug)) {
				errors.slug = 'Slug can only contain letters, numbers, hyphens, and underscores';
			}

			if (Object.keys(errors).length > 0) {
				return;
			}

			const payload = {
				slug: formData.slug.trim(),
				name: formData.name.trim(),
				public_enabled: formData.public_enabled
			};

			let statusPageId: string;

			if (statusPage) {
				// Update existing status page
				await fetchAPI(`/api/v1/status-pages/${statusPage.id}`, {
					method: 'PUT',
					body: JSON.stringify(payload)
				});
				statusPageId = statusPage.id;
			} else {
				// Create new status page
				const response = await fetchAPI('/api/v1/status-pages', {
					method: 'POST',
					body: JSON.stringify(payload)
				});
				const data = await response.json();
				statusPageId = data.id;
			}

			// Manage monitor associations
			await updateMonitorAssociations(statusPageId, formData.monitor_ids);

			dispatch('save');
		} catch (err: any) {
			if (err.errors) {
				errors = err.errors;
			} else {
				errors.submit = err.message || 'An error occurred';
			}
		} finally {
			isSubmitting = false;
		}
	}

	async function updateMonitorAssociations(statusPageId: string, selectedMonitorIds: string[]) {
		// Get current monitors
		const currentResponse = await fetchAPI(`/api/v1/status-pages/${statusPageId}/monitors`);
		const currentData = await currentResponse.json();
		const currentMonitorIds = (currentData.monitors || []).map((m: Monitor) => m.id);

		// Add new monitors
		for (const monitorId of selectedMonitorIds) {
			if (!currentMonitorIds.includes(monitorId)) {
				await fetchAPI(`/api/v1/status-pages/${statusPageId}/monitors/${monitorId}`, {
					method: 'POST'
				});
			}
		}

		// Remove monitors that are no longer selected
		for (const monitorId of currentMonitorIds) {
			if (!selectedMonitorIds.includes(monitorId)) {
				await fetchAPI(`/api/v1/status-pages/${statusPageId}/monitors/${monitorId}`, {
					method: 'DELETE'
				});
			}
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
					{statusPage ? 'Edit Status Page' : 'Create Status Page'}
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
							placeholder="e.g., API Status"
						/>
					</div>
					{#if errors.name}
						<p class="mt-2 text-sm text-red-600 dark:text-red-400">{errors.name}</p>
					{/if}
				</div>

				<!-- Slug -->
				<div>
					<label for="slug" class="block text-sm font-medium leading-6 text-slate-900 dark:text-slate-200">
						Slug <span class="text-red-500">*</span>
					</label>
					<div class="mt-2">
						<input
							type="text"
							id="slug"
							bind:value={formData.slug}
							class="block w-full rounded-md border-0 px-3 py-2 text-slate-900 dark:text-slate-100 dark:bg-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-700 placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
							placeholder="e.g., api-status"
						/>
					</div>
					<p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
						URL-friendly identifier. Only letters, numbers, hyphens, and underscores allowed.
					</p>
					{#if errors.slug}
						<p class="mt-2 text-sm text-red-600 dark:text-red-400">{errors.slug}</p>
					{/if}
				</div>

				<!-- Public Enabled -->
				<div class="border-t border-slate-200 dark:border-slate-700 pt-6">
					<div class="relative flex gap-x-3">
						<div class="flex h-6 items-center">
							<input
								id="public_enabled"
								name="public_enabled"
								type="checkbox"
								bind:checked={formData.public_enabled}
								class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-800"
							/>
						</div>
						<div class="text-sm leading-6">
							<label for="public_enabled" class="font-medium text-slate-900 dark:text-slate-200">
								Make this status page public
							</label>
							<p class="text-slate-500 dark:text-slate-400">
								When enabled, anyone with the URL can view this status page without authentication.
							</p>
						</div>
					</div>
				<!-- Monitors -->
				<div class="border-t border-slate-200 dark:border-slate-700 pt-6">
					<h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-slate-100 mb-4">Monitors</h3>
					{#if monitors.length === 0}
						<div class="text-sm text-slate-500 dark:text-slate-400">
							<p class="mb-2">No monitors available.</p>
							<p><a href="/monitors" class="text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300 underline">Create monitors first</a> to associate them with this status page.</p>
						</div>
					{:else}
						<div class="space-y-3 max-h-48 overflow-y-auto rounded-md border border-slate-200 dark:border-slate-700 p-4 bg-slate-50 dark:bg-slate-900/50">
							{#each monitors as monitor}
								<div class="relative flex items-start">
									<div class="flex h-6 items-center">
										<input
											id="monitor-{monitor.id}"
											name="monitor-{monitor.id}"
											type="checkbox"
											checked={formData.monitor_ids.includes(monitor.id)}
											on:change={() => toggleMonitor(monitor.id)}
											class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-800"
										/>
									</div>
									<div class="ml-3 text-sm leading-6">
										<label for="monitor-{monitor.id}" class="font-medium text-slate-900 dark:text-slate-200">
											{monitor.name}
										</label>
										<p class="text-slate-500 dark:text-slate-400">
											{monitor.url} ({monitor.type})
											{#if !monitor.enabled}
												<span class="text-red-500 dark:text-red-400">- Disabled</span>
											{/if}
										</p>
									</div>
								</div>
							{/each}
						</div>
						<p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
							Selected: {formData.monitor_ids.length} monitor{formData.monitor_ids.length !== 1 ? 's' : ''}
						</p>
					{/if}
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
						{isSubmitting ? 'Saving...' : statusPage ? 'Update Status Page' : 'Create Status Page'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}