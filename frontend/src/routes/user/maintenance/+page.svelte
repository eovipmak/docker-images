<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import type { MaintenanceWindow, Monitor } from '$lib/types';
	import Card from '$lib/components/Card.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let maintenanceWindows: MaintenanceWindow[] = [];
	let monitors: Monitor[] = [];
	let isLoading = true;
	let error = '';
	let isModalOpen = false;
	let editingWindow: MaintenanceWindow | null = null;

	// Form state
	let formName = '';
	let formStartTime = '';
	let formEndTime = '';
	let formRepeatInterval = 0;
	let formMonitorIds: string[] = [];
	let formTags: string[] = [];
	let formTagInput = '';

	// Confirm modal state
	let isConfirmModalOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let onConfirmCallback: (() => void) | null = null;

	// Repeat interval options
	const repeatOptions = [
		{ value: 0, label: 'One-time (no repeat)' },
		{ value: 3600, label: 'Hourly' },
		{ value: 86400, label: 'Daily' },
		{ value: 604800, label: 'Weekly' },
		{ value: 2592000, label: 'Monthly (30 days)' }
	];

	onMount(() => {
		loadData();
	});

	async function loadData() {
		isLoading = true;
		error = '';
		try {
			const [windowsRes, monitorsRes] = await Promise.all([
				fetchAPI('/api/v1/maintenance-windows'),
				fetchAPI('/api/v1/monitors')
			]);

			if (!windowsRes.ok) {
				const errorData = await windowsRes.json().catch(() => ({}));
				throw new Error(errorData.error || 'Failed to load maintenance windows');
			}
			if (!monitorsRes.ok) {
				const errorData = await monitorsRes.json().catch(() => ({}));
				throw new Error(errorData.error || 'Failed to load monitors');
			}

			const windowsData = await windowsRes.json();
			const monitorsData = await monitorsRes.json();
			
			// Ensure arrays even if null is returned
			maintenanceWindows = Array.isArray(windowsData) ? windowsData : [];
			monitors = Array.isArray(monitorsData) ? monitorsData : [];
		} catch (err: any) {
			error = err.message || 'Failed to load data';
			console.error('Error loading maintenance data:', err);
		} finally {
			isLoading = false;
		}
	}

	function openCreateModal() {
		editingWindow = null;
		formName = '';
		formStartTime = '';
		formEndTime = '';
		formRepeatInterval = 0;
		formMonitorIds = [];
		formTags = [];
		formTagInput = '';
		isModalOpen = true;
	}

	function openEditModal(window: MaintenanceWindow) {
		editingWindow = window;
		formName = window.name;
		formStartTime = formatDateTimeLocal(window.start_time);
		formEndTime = formatDateTimeLocal(window.end_time);
		formRepeatInterval = window.repeat_interval;
		formMonitorIds = window.monitor_ids || [];
		formTags = window.tags || [];
		formTagInput = '';
		isModalOpen = true;
	}

	function formatDateTimeLocal(isoString: string): string {
		const date = new Date(isoString);
		const offset = date.getTimezoneOffset();
		const localDate = new Date(date.getTime() - offset * 60 * 1000);
		return localDate.toISOString().slice(0, 16);
	}

	function closeModal() {
		isModalOpen = false;
		editingWindow = null;
	}

	async function handleSubmit() {
		if (!formName.trim() || !formStartTime || !formEndTime) {
			error = 'Please fill in all required fields';
			return;
		}

		const payload = {
			name: formName.trim(),
			start_time: new Date(formStartTime).toISOString(),
			end_time: new Date(formEndTime).toISOString(),
			repeat_interval: formRepeatInterval,
			monitor_ids: formMonitorIds,
			tags: formTags
		};

		try {
			const url = editingWindow
				? `/api/v1/maintenance-windows/${editingWindow.id}`
				: '/api/v1/maintenance-windows';
			const method = editingWindow ? 'PUT' : 'POST';

			const response = await fetchAPI(url, {
				method,
				body: JSON.stringify(payload)
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to save maintenance window');
			}

			closeModal();
			await loadData();
		} catch (err: any) {
			error = err.message;
		}
	}

	function confirmDelete(window: MaintenanceWindow) {
		confirmTitle = 'Delete Maintenance Window';
		confirmMessage = `Are you sure you want to delete "${window.name}"?`;
		onConfirmCallback = () => deleteWindow(window.id);
		isConfirmModalOpen = true;
	}

	async function deleteWindow(id: string) {
		try {
			const response = await fetchAPI(`/api/v1/maintenance-windows/${id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				throw new Error('Failed to delete maintenance window');
			}

			await loadData();
		} catch (err: any) {
			error = err.message;
		}
	}

	function addTag() {
		const tag = formTagInput.trim();
		if (tag && !formTags.includes(tag)) {
			formTags = [...formTags, tag];
			formTagInput = '';
		}
	}

	function removeTag(tag: string) {
		formTags = formTags.filter((t) => t !== tag);
	}

	function toggleMonitor(monitorId: string) {
		if (formMonitorIds.includes(monitorId)) {
			formMonitorIds = formMonitorIds.filter((id) => id !== monitorId);
		} else {
			formMonitorIds = [...formMonitorIds, monitorId];
		}
	}

	function formatDateTime(isoString: string): string {
		return new Date(isoString).toLocaleString();
	}

	function getRepeatLabel(seconds: number): string {
		const option = repeatOptions.find((o) => o.value === seconds);
		return option?.label || 'Custom';
	}

	function isActive(window: MaintenanceWindow): boolean {
		const now = new Date();
		const start = new Date(window.start_time);
		const end = new Date(window.end_time);
		return now >= start && now <= end;
	}

	function getMonitorNames(ids: string[]): string {
		if (!ids || ids.length === 0) return 'All monitors';
		return ids
			.map((id) => monitors.find((m) => m.id === id)?.name || id)
			.join(', ');
	}
</script>

<svelte:head>
	<title>Maintenance Windows - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-slate-900 dark:text-white">Maintenance Windows</h1>
			<p class="text-slate-600 dark:text-slate-400 mt-1">
				Schedule maintenance periods to suppress alerts
			</p>
		</div>
		<button
			on:click={openCreateModal}
			class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				fill="none"
				viewBox="0 0 24 24"
				stroke-width="1.5"
				stroke="currentColor"
				class="w-5 h-5"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
			</svg>
			Schedule Maintenance
		</button>
	</div>

	{#if error}
		<div class="mb-4 p-4 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-lg">
			{error}
		</div>
	{/if}

	{#if isLoading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else if maintenanceWindows.length === 0}
		<Card>
			<div class="text-center py-12">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="w-12 h-12 mx-auto text-slate-400 mb-4"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z"
					/>
				</svg>
				<h3 class="text-lg font-medium text-slate-900 dark:text-white mb-2">
					No maintenance windows scheduled
				</h3>
				<p class="text-slate-500 dark:text-slate-400 mb-4">
					Schedule maintenance windows to suppress alerts during planned downtime.
				</p>
				<button
					on:click={openCreateModal}
					class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
				>
					Schedule Your First Maintenance
				</button>
			</div>
		</Card>
	{:else}
		<div class="grid gap-4">
			{#each maintenanceWindows as window}
				<Card>
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-3 mb-2">
								<h3 class="text-lg font-semibold text-slate-900 dark:text-white">
									{window.name}
								</h3>
								{#if isActive(window)}
									<span
										class="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 rounded-full"
									>
										Active
									</span>
								{:else if new Date(window.start_time) > new Date()}
									<span
										class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400 rounded-full"
									>
										Scheduled
									</span>
								{:else}
									<span
										class="px-2 py-1 text-xs font-medium bg-slate-100 text-slate-800 dark:bg-slate-700 dark:text-slate-300 rounded-full"
									>
										Completed
									</span>
								{/if}
							</div>

							<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-slate-600 dark:text-slate-400">
								<div>
									<span class="font-medium">Start:</span>
									{formatDateTime(window.start_time)}
								</div>
								<div>
									<span class="font-medium">End:</span>
									{formatDateTime(window.end_time)}
								</div>
								<div>
									<span class="font-medium">Repeat:</span>
									{getRepeatLabel(window.repeat_interval)}
								</div>
								<div>
									<span class="font-medium">Monitors:</span>
									{getMonitorNames(window.monitor_ids)}
								</div>
							</div>

							{#if window.tags && window.tags.length > 0}
								<div class="mt-3 flex flex-wrap gap-2">
									{#each window.tags as tag}
										<span
											class="px-2 py-1 text-xs bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded"
										>
											{tag}
										</span>
									{/each}
								</div>
							{/if}
						</div>

						<div class="flex gap-2 ml-4">
							<button
								on:click={() => openEditModal(window)}
								class="p-2 text-slate-500 hover:text-blue-600 dark:text-slate-400 dark:hover:text-blue-400 transition-colors"
								title="Edit"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="1.5"
									stroke="currentColor"
									class="w-5 h-5"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10"
									/>
								</svg>
							</button>
							<button
								on:click={() => confirmDelete(window)}
								class="p-2 text-slate-500 hover:text-red-600 dark:text-slate-400 dark:hover:text-red-400 transition-colors"
								title="Delete"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									fill="none"
									viewBox="0 0 24 24"
									stroke-width="1.5"
									stroke="currentColor"
									class="w-5 h-5"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
									/>
								</svg>
							</button>
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Create/Edit Modal -->
{#if isModalOpen}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
		on:click={closeModal}
		on:keydown={(e) => e.key === 'Escape' && closeModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white dark:bg-slate-800 rounded-xl shadow-xl max-w-lg w-full max-h-[90vh] overflow-y-auto"
			on:click|stopPropagation
			on:keydown|stopPropagation
			role="document"
		>
			<div class="p-6 border-b border-slate-200 dark:border-slate-700">
				<h2 class="text-xl font-semibold text-slate-900 dark:text-white">
					{editingWindow ? 'Edit Maintenance Window' : 'Schedule Maintenance Window'}
				</h2>
			</div>

			<form on:submit|preventDefault={handleSubmit} class="p-6 space-y-4">
				<div>
					<label
						for="name"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
					>
						Name *
					</label>
					<input
						type="text"
						id="name"
						bind:value={formName}
						class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Weekly server maintenance"
						required
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label
							for="startTime"
							class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
						>
							Start Time *
						</label>
						<input
							type="datetime-local"
							id="startTime"
							bind:value={formStartTime}
							class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							required
						/>
					</div>
					<div>
						<label
							for="endTime"
							class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
						>
							End Time *
						</label>
						<input
							type="datetime-local"
							id="endTime"
							bind:value={formEndTime}
							class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							required
						/>
					</div>
				</div>

				<div>
					<label
						for="repeatInterval"
						class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1"
					>
						Repeat
					</label>
					<select
						id="repeatInterval"
						bind:value={formRepeatInterval}
						class="w-full px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					>
						{#each repeatOptions as option}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
						Apply to Monitors
					</label>
					<p class="text-xs text-slate-500 dark:text-slate-400 mb-2">
						Leave empty to apply to all monitors
					</p>
					<div class="max-h-40 overflow-y-auto border border-slate-300 dark:border-slate-600 rounded-lg p-2 space-y-1">
						{#each monitors as monitor}
							<label class="flex items-center gap-2 p-2 hover:bg-slate-50 dark:hover:bg-slate-700 rounded cursor-pointer">
								<input
									type="checkbox"
									checked={formMonitorIds.includes(monitor.id)}
									on:change={() => toggleMonitor(monitor.id)}
									class="rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-500"
								/>
								<span class="text-sm text-slate-700 dark:text-slate-300">{monitor.name}</span>
							</label>
						{/each}
						{#if monitors.length === 0}
							<p class="text-sm text-slate-500 dark:text-slate-400 p-2">No monitors available</p>
						{/if}
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
						Tags
					</label>
					<div class="flex gap-2">
						<input
							type="text"
							bind:value={formTagInput}
							on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
							class="flex-1 px-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="Add a tag"
						/>
						<button
							type="button"
							on:click={addTag}
							class="px-3 py-2 bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
						>
							Add
						</button>
					</div>
					{#if formTags.length > 0}
						<div class="mt-2 flex flex-wrap gap-2">
							{#each formTags as tag}
								<span
									class="px-2 py-1 text-sm bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded flex items-center gap-1"
								>
									{tag}
									<button
										type="button"
										on:click={() => removeTag(tag)}
										class="hover:text-red-600 dark:hover:text-red-400"
									>
										×
									</button>
								</span>
							{/each}
						</div>
					{/if}
				</div>

				<div class="flex justify-end gap-3 pt-4">
					<button
						type="button"
						on:click={closeModal}
						class="px-4 py-2 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
					>
						{editingWindow ? 'Update' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<ConfirmModal
	isOpen={isConfirmModalOpen}
	title={confirmTitle}
	message={confirmMessage}
	on:confirm={() => {
		isConfirmModalOpen = false;
		if (onConfirmCallback) onConfirmCallback();
	}}
	on:cancel={() => (isConfirmModalOpen = false)}
/>
