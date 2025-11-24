<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import IncidentTimeline from '$lib/components/IncidentTimeline.svelte';

	let incidentId: string = '';
	let incident: any = null;
	let monitorChecks: any[] = [];
	let isLoading = true;
	let isResolving = false;
	let error = '';

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
		if (!confirm('Are you sure you want to manually resolve this incident?')) {
			return;
		}

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
	}

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString();
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

<div class="container mx-auto px-4 py-8">
	<div class="max-w-7xl mx-auto">
		<!-- Back button -->
		<button
			on:click={handleBack}
			class="mb-4 text-blue-600 hover:text-blue-800 font-medium flex items-center gap-2"
		>
			← Back to Incidents
		</button>

		{#if isLoading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
				{error}
			</div>
		{:else if incident}
			<!-- Header -->
			<div class="bg-white rounded-lg shadow-md p-6 mb-6">
				<div class="flex items-start justify-between mb-4">
					<div>
						<h1 class="text-2xl font-bold text-gray-900 mb-2">Incident Details</h1>
						<p class="text-gray-600">ID: {incident.id}</p>
					</div>
					<StatusBadge status={incident.status} size="md" />
				</div>

				<!-- Incident Info Grid -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<h3 class="text-sm font-medium text-gray-500 mb-1">Monitor</h3>
						<p class="text-base font-medium text-gray-900">{incident.monitor_name}</p>
						<p class="text-sm text-gray-600">{incident.monitor_url}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-gray-500 mb-1">Alert Rule</h3>
						<p class="text-base text-gray-900">{incident.alert_rule_name || 'N/A'}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-gray-500 mb-1">Started At</h3>
						<p class="text-base text-gray-900">{formatDate(incident.started_at)}</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-gray-500 mb-1">Resolved At</h3>
						<p class="text-base text-gray-900">
							{incident.resolved_at && incident.resolved_at.Valid ? formatDate(incident.resolved_at.Time) : 'Ongoing'}
						</p>
					</div>

					<div>
						<h3 class="text-sm font-medium text-gray-500 mb-1">Duration</h3>
						<p class="text-base text-gray-900">{formatDuration(incident.duration)}</p>
					</div>

					{#if incident.trigger_value}
						<div>
							<h3 class="text-sm font-medium text-gray-500 mb-1">Trigger Value</h3>
							<p class="text-base text-gray-900">{incident.trigger_value}</p>
						</div>
					{/if}
				</div>

				<!-- Manual resolve button -->
				{#if incident.status === 'open'}
					<div class="mt-6 pt-6 border-t border-gray-200">
						<button
							on:click={handleResolve}
							disabled={isResolving}
							class="bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{isResolving ? 'Resolving...' : 'Manually Resolve Incident'}
						</button>
					</div>
				{/if}
			</div>

			<!-- Timeline -->
			<div class="mb-6">
				<IncidentTimeline {incident} />
			</div>

			<!-- Recent Monitor Checks -->
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Recent Monitor Checks</h3>
				
				{#if monitorChecks.length === 0}
					<p class="text-gray-500">No recent checks available</p>
				{:else}
					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-gray-200">
							<thead class="bg-gray-50">
								<tr>
									<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
										Time
									</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
										Status
									</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
										Response Time
									</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
										Status Code
									</th>
									<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
										Error
									</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-gray-200">
								{#each monitorChecks as check}
									<tr>
										<td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
											{formatDate(check.checked_at)}
										</td>
										<td class="px-4 py-3 whitespace-nowrap">
											<span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full {check.success
												? 'bg-green-100 text-green-800'
												: 'bg-red-100 text-red-800'}">
												{check.success ? 'Success' : 'Failed'}
											</span>
										</td>
										<td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
											{check.response_time_ms && check.response_time_ms.Valid ? `${check.response_time_ms.Int64}ms` : 'N/A'}
										</td>
										<td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
											{check.status_code && check.status_code.Valid ? check.status_code.Int64 : 'N/A'}
										</td>
										<td class="px-4 py-3 text-sm text-gray-900">
											{check.error_message && check.error_message.Valid ? check.error_message.String : '-'}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>
