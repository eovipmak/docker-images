<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import IncidentBadge from '$lib/components/IncidentBadge.svelte';
	import IncidentTimeline from '$lib/components/IncidentTimeline.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let incidentId: string = '';
	let incident: any = null;
	let monitorChecks: any[] = [];
	let isLoading = true;
	let isResolving = false;
	let error = '';

	// Confirm modal state
	let isConfirmModalOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let onConfirmCallback: (() => void) | null = null;

	$: incidentId = $page.params.id || '';

	onMount(() => {
		if (incidentId) {
			loadIncidentDetails();
		}
	});

	async function loadIncidentDetails() {
		isLoading = true;
		error = '';

		try {
			const response = await fetchAPI(`/api/v1/incidents/${incidentId}`);
			if (!response.ok) {
				if (response.status === 404) {
					error = 'Incident not found';
				} else {
					throw new Error('Failed to load incident');
				}
				return;
			}

			incident = await response.json();
			// Load monitor checks after incident is loaded
			loadMonitorChecks();
		} catch (err: any) {
			console.error('Error loading incident:', err);
			error = err.message || 'Failed to load incident';
		} finally {
			isLoading = false;
		}
	}

	async function loadMonitorChecks() {
		if (!incident) return;

		try {
			// Load recent checks for the monitor during the incident period
			const response = await fetchAPI(`/api/v1/monitors/${incident.monitor_id}/checks?limit=20`);
			if (response.ok) {
				monitorChecks = await response.json();
			}
		} catch (err) {
			console.error('Error loading monitor checks:', err);
			// Don't block on check load failure
		}
	}

	async function handleResolve() {
		confirmTitle = 'Resolve Incident';
		confirmMessage = 'Are you sure you want to manually resolve this incident?';
		onConfirmCallback = async () => {
			isResolving = true;
			try {
				const response = await fetchAPI(`/api/v1/incidents/${incidentId}/resolve`, {
					method: 'POST'
				});

				if (!response.ok) {
					throw new Error('Failed to resolve incident');
				}

				// Reload incident details
				await loadIncidentDetails();
			} catch (err: any) {
				console.error('Error resolving incident:', err);
				alert(err.message || 'Failed to resolve incident');
			} finally {
				isResolving = false;
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

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}

	function formatDuration(seconds: number | null): string {
		if (!seconds) return 'N/A';

		const days = Math.floor(seconds / 86400);
		const hours = Math.floor((seconds % 86400) / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);

		if (days > 0) {
			return `${days} day${days > 1 ? 's' : ''} ${hours} hour${hours !== 1 ? 's' : ''} ${minutes} minute${minutes !== 1 ? 's' : ''}`;
		} else if (hours > 0) {
			return `${hours} hour${hours !== 1 ? 's' : ''} ${minutes} minute${minutes !== 1 ? 's' : ''}`;
		} else if (minutes > 0) {
			return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
		} else {
			return `${Math.floor(seconds)} second${Math.floor(seconds) !== 1 ? 's' : ''}`;
		}
	}

	function handleBack() {
		goto('/incidents');
	}
</script>

<svelte:head>
	<title>Incident Details - V-Insight</title>
</svelte:head>

<div class="px-4 sm:px-6 lg:px-8 py-8">
	<!-- Back button -->
	<button
		on:click={handleBack}
		class="mb-6 text-sm font-medium text-slate-500 hover:text-slate-700 flex items-center gap-1 transition-colors"
	>
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
		</svg>
		Back to Incidents
	</button>

	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
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
						<p>{error}</p>
					</div>
				</div>
			</div>
		</div>
	{:else if incident}
		<!-- Header -->
		<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden mb-6">
			<div class="px-4 py-5 sm:p-6">
				<div class="flex items-start justify-between mb-6">
					<div>
						<h1 class="text-2xl font-bold text-slate-900 dark:text-white">Incident Details</h1>
						<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">ID: {incident.id}</p>
					</div>
					<IncidentBadge status={incident.status} severity={incident.severity || 'warning'} />
				</div>

				<!-- Incident Info Grid -->
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					<div>
						<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Monitor</h3>
						<p class="text-base font-medium text-slate-900 dark:text-white">{incident.monitor_name}</p>
						<p class="text-sm text-slate-500 dark:text-slate-400 truncate">{incident.monitor_url}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Alert Rule</h3>
						<p class="text-base text-slate-900 dark:text-white">{incident.alert_rule_name || 'N/A'}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Started At</h3>
						<p class="text-base text-slate-900 dark:text-white">{formatDate(incident.started_at)}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Resolved At</h3>
						<p class="text-base text-slate-900 dark:text-white">
							{incident.resolved_at && incident.resolved_at.Valid ? formatDate(incident.resolved_at.Time) : 'Ongoing'}
						</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Duration</h3>
						<p class="text-base text-slate-900 dark:text-white">{formatDuration(incident.duration)}</p>
					</div>

					{#if incident.trigger_value}
						<div>
							<h3 class="text-sm font-medium text-slate-500 dark:text-slate-400 mb-1">Trigger Value</h3>
							<p class="text-base text-slate-900 dark:text-white">{incident.trigger_value}</p>
						</div>
					{/if}
				</div>

				<!-- Manual resolve button -->
				{#if incident.status === 'open'}
					<div class="mt-8 pt-6 border-t border-slate-100 dark:border-slate-700">
						<button
							on:click={handleResolve}
							disabled={isResolving}
							class="inline-flex items-center rounded-md bg-green-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
						>
							{#if isResolving}
								<svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Resolving...
							{:else}
								<svg class="-ml-0.5 mr-1.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
									<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clip-rule="evenodd" />
								</svg>
								Manually Resolve Incident
							{/if}
						</button>
					</div>
				{/if}
			</div>
		</div>

		<!-- Timeline -->
		<div class="mb-6">
			<IncidentTimeline {incident} />
		</div>

		<!-- Recent Monitor Checks -->
		<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden">
			<div class="px-4 py-5 sm:p-6">
				<h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-white mb-4">Recent Monitor Checks</h3>
				
				{#if monitorChecks.length === 0}
					<p class="text-sm text-slate-500 dark:text-slate-400">No recent checks available</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
							<thead class="bg-slate-50 dark:bg-slate-900">
								<tr>
									<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Time</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Status</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Response Time</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Status Code</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Error</th>
								</tr>
							</thead>
							<tbody class="bg-white dark:bg-slate-800 divide-y divide-slate-200 dark:divide-slate-700">
								{#each monitorChecks as check}
									<tr class="hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors">
										<td class="px-4 py-3 whitespace-nowrap text-sm text-slate-900 dark:text-white">
											{formatDate(check.checked_at)}
										</td>
										<td class="px-4 py-3 whitespace-nowrap">
											<span class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ring-1 ring-inset {check.success ? 'bg-green-50 text-green-700 ring-green-600/20' : 'bg-red-50 text-red-700 ring-red-600/10'}">
												{check.success ? 'Success' : 'Failed'}
											</span>
										</td>
										<td class="px-4 py-3 whitespace-nowrap text-sm text-slate-600 dark:text-slate-300">
											{check.response_time_ms && check.response_time_ms.Valid ? `${check.response_time_ms.Int64}ms` : 'N/A'}
										</td>
										<td class="px-4 py-3 whitespace-nowrap text-sm text-slate-600 dark:text-slate-300">
											{check.status_code && check.status_code.Valid ? check.status_code.Int64 : 'N/A'}
										</td>
										<td class="px-4 py-3 text-sm text-slate-600 dark:text-slate-300 max-w-xs truncate" title={check.error_message && check.error_message.Valid ? check.error_message.String : ''}>
											{check.error_message && check.error_message.Valid ? check.error_message.String : '-'}
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

<ConfirmModal
	isOpen={isConfirmModalOpen}
	title={confirmTitle}
	message={confirmMessage}
	on:confirm={handleConfirmDelete}
	on:cancel={handleCancelDelete}
/>
