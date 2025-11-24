<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks, latestIncidents } from '$lib/api/events';
	import StatCard from '$lib/components/StatCard.svelte';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import IncidentBadge from '$lib/components/IncidentBadge.svelte';

	interface MonitorCheck {
		id: string;
		monitor_id: string;
		checked_at: string;
		status_code?: number;
		response_time_ms?: number;
		ssl_valid?: boolean;
		ssl_expires_at?: string;
		error_message?: string;
		success: boolean;
	}

	interface Monitor {
		id: string;
		tenant_id: number;
		name: string;
		url: string;
		check_interval: number;
		timeout: number;
		enabled: boolean;
		check_ssl: boolean;
		ssl_alert_days: number;
		last_checked_at?: string;
		created_at: string;
		updated_at: string;
	}

	interface Incident {
		id: string;
		monitor_id: string;
		alert_rule_id: string;
		started_at: string;
		resolved_at?: string;
		status: 'open' | 'resolved';
		trigger_value?: string;
		notified_at?: string;
		created_at: string;
	}

	interface DashboardStats {
		total_monitors: number;
		up_count: number;
		down_count: number;
		open_incidents: number;
	}

	interface MonitorCheckWithMonitor {
		check: MonitorCheck;
		monitor: Monitor;
	}

	interface IncidentWithDetails {
		incident: Incident;
		monitor: Monitor;
	}

	interface DashboardData {
		stats: DashboardStats;
		recent_checks: MonitorCheckWithMonitor[];
		open_incidents: IncidentWithDetails[];
	}

	let stats: DashboardStats = {
		total_monitors: 0,
		up_count: 0,
		down_count: 0,
		open_incidents: 0
	};
	let recentChecks: MonitorCheckWithMonitor[] = [];
	let openIncidents: IncidentWithDetails[] = [];
	let isLoading = true;
	let error = '';

	// Subscribe to SSE events
	let unsubscribeChecks: (() => void) | null = null;
	let unsubscribeIncidents: (() => void) | null = null;

	onMount(async () => {
		try {
			const response = await fetchAPI('/api/v1/dashboard');

			if (!response.ok) {
				error = 'Failed to load dashboard data';
				return;
			}

			const data: DashboardData = await response.json();
			stats = data.stats;
			recentChecks = data.recent_checks || [];
			openIncidents = data.open_incidents || [];
		} catch (err) {
			console.error('Error loading dashboard:', err);
			error = 'An error occurred while loading dashboard data';
		} finally {
			isLoading = false;
		}

		// Subscribe to monitor check events
		unsubscribeChecks = latestMonitorChecks.subscribe((checks) => {
			// Update recent checks and stats based on SSE events
			if (checks.size > 0) {
				// Recalculate stats
				let upCount = 0;
				let downCount = 0;
				
				checks.forEach((check) => {
					if (check.success) {
						upCount++;
					} else {
						downCount++;
					}
				});

				stats = {
					...stats,
					up_count: upCount,
					down_count: downCount,
					total_monitors: checks.size
				};
			}
		});

		// Subscribe to incident events
		unsubscribeIncidents = latestIncidents.subscribe((incidents) => {
			// Update open incidents count
			stats = {
				...stats,
				open_incidents: incidents.length
			};
		});
	});

	onDestroy(() => {
		if (unsubscribeChecks) {
			unsubscribeChecks();
		}
		if (unsubscribeIncidents) {
			unsubscribeIncidents();
		}
	});

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

	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-8">
			{error}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
			<StatCard title="Total Monitors" value={stats.total_monitors} valueColor="text-gray-900" />
			<StatCard title="Monitors Up" value={stats.up_count} valueColor="text-green-600" />
			<StatCard title="Monitors Down" value={stats.down_count} valueColor="text-red-600" />
			<StatCard title="Open Incidents" value={stats.open_incidents} valueColor="text-yellow-600" />
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-1 gap-6 mb-8">
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
	{/if}
</div>

