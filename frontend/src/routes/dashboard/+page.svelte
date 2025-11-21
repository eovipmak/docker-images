<script lang="ts">
	import type { PageData } from './$types';
	import StatCard from '$lib/components/StatCard.svelte';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import IncidentBadge from '$lib/components/IncidentBadge.svelte';

	export let data: PageData;

	$: stats = data.stats;
	$: recentChecks = data.recentChecks || [];
	$: openIncidents = data.openIncidents || [];

	// Format date for display
	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString();
	}

	// Format relative time (e.g., "2 minutes ago")
	function formatRelativeTime(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
		if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
		return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
	}

	// Determine incident severity based on trigger type
	function getIncidentSeverity(triggerValue?: string): 'critical' | 'warning' | 'info' {
		// For now, treat all as warning, but this can be enhanced
		return 'warning';
	}
</script>

<svelte:head>
	<title>Dashboard - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<h1 class="text-3xl font-bold text-gray-900 mb-6">Dashboard</h1>
	<p class="text-gray-600 mb-8">Monitor your domains and view system metrics</p>

	<!-- Stats Cards -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
		<StatCard title="Total Monitors" value={stats.total_monitors} valueColor="text-gray-900" />
		<StatCard title="Monitors Up" value={stats.up_count} valueColor="text-green-600" />
		<StatCard title="Monitors Down" value={stats.down_count} valueColor="text-red-600" />
		<StatCard title="Open Incidents" value={stats.open_incidents} valueColor="text-yellow-600" />
	</div>

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
		<!-- Recent Checks -->
		<div class="bg-white rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold text-gray-800 mb-4">Recent Checks</h2>
			{#if recentChecks.length === 0}
				<p class="text-gray-500">No recent checks to display</p>
			{:else}
				<div class="space-y-4">
					{#each recentChecks as { check, monitor }}
						<div class="border-b border-gray-200 pb-3 last:border-0">
							<div class="flex items-start justify-between">
								<div class="flex-1">
									<div class="flex items-center gap-2 mb-1">
										<MonitorStatus status={check.success ? 'up' : 'down'} />
										<span class="font-medium text-gray-900">{monitor.name}</span>
									</div>
									<p class="text-sm text-gray-600 truncate">{monitor.url}</p>
									{#if check.response_time_ms}
										<p class="text-xs text-gray-500 mt-1">
											Response time: {check.response_time_ms}ms
										</p>
									{/if}
									{#if check.error_message}
										<p class="text-xs text-red-600 mt-1">
											{check.error_message}
										</p>
									{/if}
								</div>
								<div class="text-right">
									<p class="text-xs text-gray-500">
										{formatRelativeTime(check.checked_at)}
									</p>
									{#if check.status_code}
										<p class="text-xs text-gray-600 mt-1">
											Status: {check.status_code}
										</p>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Open Incidents -->
		<div class="bg-white rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold text-gray-800 mb-4">Open Incidents</h2>
			{#if openIncidents.length === 0}
				<p class="text-gray-500">No open incidents</p>
			{:else}
				<div class="space-y-4">
					{#each openIncidents as { incident, monitor }}
						<div class="border-b border-gray-200 pb-3 last:border-0">
							<div class="flex items-start justify-between">
								<div class="flex-1">
									<div class="flex items-center gap-2 mb-1">
										<IncidentBadge 
											status={incident.status} 
											severity={getIncidentSeverity(incident.trigger_value)} 
										/>
										<span class="font-medium text-gray-900">{monitor.name}</span>
									</div>
									<p class="text-sm text-gray-600 truncate">{monitor.url}</p>
									{#if incident.trigger_value}
										<p class="text-xs text-gray-500 mt-1">
											Trigger: {incident.trigger_value}
										</p>
									{/if}
								</div>
								<div class="text-right">
									<p class="text-xs text-gray-500">
										Started {formatRelativeTime(incident.started_at)}
									</p>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

