<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks, latestIncidents, connectEventStream, disconnectEventStream } from '$lib/api/events';
	import StatCard from '$lib/components/StatCard.svelte';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import IncidentBadge from '$lib/components/IncidentBadge.svelte';
	import MonitorCard from '$lib/components/MonitorCard.svelte';
	import AlertCard from '$lib/components/AlertCard.svelte';
	import Card from '$lib/components/Card.svelte';
	import { goto } from '$app/navigation';

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
		average_response_time: number;
		overall_uptime: number;
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
		open_incidents: 0,
		average_response_time: 0,
		overall_uptime: 0
	};
	let recentChecks: MonitorCheckWithMonitor[] = [];
	let openIncidents: IncidentWithDetails[] = [];
    let monitors: Monitor[] = [];
    let alertRules: any[] = [];
	let isLoading = true;
	let error = '';

	// Subscribe to SSE events
	let unsubscribeChecks: (() => void) | null = null;
	let unsubscribeIncidents: (() => void) | null = null;

	// Load dashboard data
	async function loadDashboardData() {
		try {
			isLoading = true;
			error = '';
			const response = await fetchAPI('/api/v1/dashboard');

			if (!response.ok) {
				error = 'Failed to load dashboard data';
				console.error('Failed to load dashboard data:', response.status);
				return;
			}

			const data: DashboardData = await response.json();
			stats = data.stats;
			recentChecks = data.recent_checks || [];
			openIncidents = data.open_incidents || [];
			// Additionally fetch a small set of monitors and alert rules for dashboard preview
			try {
				const [monResp, rulesResp] = await Promise.all([
					fetchAPI('/api/v1/monitors'),
					fetchAPI('/api/v1/alert-rules')
				]);
				if (monResp.ok) {
					const monList = await monResp.json();
					monitors = monList.slice(0, 6);
				}
				if (rulesResp.ok) {
					const rulesList = await rulesResp.json();
					alertRules = rulesList.slice(0, 6);
				}
			} catch (err) {
				// Not critical; dashboard still works without previews
				console.debug('Failed to fetch monitors/alert rules for previews', err);
			}
		} catch (err) {
			console.error('Error loading dashboard:', err);
			error = 'An error occurred while loading dashboard data';
		} finally {
			isLoading = false;
		}
	}

	function handleViewMonitor(monitor: Monitor) {
		goto(`/monitors/${monitor.id}`);
	}

	function handleViewAlert(rule: any) {
		goto('/alerts');
	}

	onMount(async () => {
		await loadDashboardData();

		// Start SSE connection for real-time updates
		await connectEventStream();

		// Subscribe to monitor check events
		unsubscribeChecks = latestMonitorChecks.subscribe((checks) => {
			// When SSE events arrive, refresh dashboard data to get updated stats
			// This ensures we have accurate data while still benefiting from real-time notifications
			if (!isLoading) {
				loadDashboardData();
			}
		});

		// Subscribe to incident events
		unsubscribeIncidents = latestIncidents.subscribe((incidents) => {
			// When incident events arrive, refresh dashboard data
			if (!isLoading) {
				loadDashboardData();
			}
		});
	});

	onDestroy(() => {
		// Disconnect SSE when leaving dashboard
		disconnectEventStream();

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

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-slate-900">Dashboard</h1>
			<p class="mt-1 text-sm text-slate-500">Overview of your monitoring status and system metrics.</p>
		</div>
	</div>

	{#if isLoading}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
            {#each Array(6) as _}
                <div class="h-32 bg-slate-100 rounded-xl animate-pulse"></div>
            {/each}
        </div>
	{:else if error}
		<div class="p-4 rounded-lg bg-red-50 border border-red-200 text-red-700 flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
            </svg>
			{error}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
			<StatCard 
                title="Total Monitors" 
                value={stats.total_monitors} 
                valueColor="text-slate-900" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" />'
            />
			<StatCard 
                title="Operational" 
                value={stats.up_count} 
                valueColor="text-emerald-600" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />'
            />
			<StatCard 
                title="Downtime" 
                value={stats.down_count} 
                valueColor="text-rose-600" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />'
            />
			<StatCard 
                title="Open Incidents" 
                value={stats.open_incidents} 
                valueColor="text-amber-600" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />'
            />
			<StatCard 
				title="Avg Response" 
				value={stats.average_response_time ? `${Math.round(stats.average_response_time)}ms` : 'N/A'} 
				valueColor="text-sky-600" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />'
			/>
			<StatCard 
				title="Uptime (24h)" 
				value={stats.overall_uptime ? `${stats.overall_uptime.toFixed(2)}%` : 'N/A'} 
				valueColor="text-emerald-600" 
                icon='<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />'
			/>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
			<!-- Open Incidents -->
			<div class="lg:col-span-2 space-y-6">
				<Card>
					<div slot="header" class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
						<h2 class="text-lg font-semibold text-slate-900">Open Incidents</h2>
						{#if openIncidents.length > 0}
							<span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-rose-100 text-rose-700 border border-rose-200">
								{openIncidents.length} Active
							</span>
						{/if}
					</div>
                    
					{#if openIncidents.length === 0}
						<div class="p-12 text-center">
                            <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-emerald-100 mb-4">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 text-emerald-600">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <h3 class="text-sm font-medium text-slate-900">All systems operational</h3>
                            <p class="mt-1 text-sm text-slate-500">No active incidents reported at this time.</p>
                        </div>
                    {:else}
                        <div class="divide-y divide-slate-100">
                            {#each openIncidents as { incident, monitor }}
                                <div class="p-4 sm:p-6 hover:bg-slate-50 transition-colors">
                                    <div class="flex items-start justify-between">
                                        <div class="flex-1 min-w-0">
                                            <div class="flex items-center gap-3 mb-1">
                                                <IncidentBadge 
                                                    status={incident.status} 
                                                    severity={getIncidentSeverity(incident.trigger_value)} 
                                                />
                                                <span class="font-medium text-slate-900 truncate">{monitor.name}</span>
                                            </div>
                                            <p class="text-sm text-slate-500 truncate mb-2">{monitor.url}</p>
                                            {#if incident.trigger_value}
                                                <div class="inline-flex items-center px-2 py-1 rounded bg-slate-100 text-xs text-slate-600">
                                                    <span class="font-medium mr-1">Trigger:</span> {incident.trigger_value}
                                                </div>
                                            {/if}
                                        </div>
                                        <div class="text-right ml-4 flex-shrink-0">
                                            <p class="text-xs font-medium text-slate-500">
                                                Started {formatRelativeTime(incident.started_at)}
                                            </p>
                                            <p class="text-xs text-slate-400 mt-1">
                                                {formatDate(incident.started_at)}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>
					{/if}
				</Card>
			</div>
            
            <!-- Recent Checks (Optional, if we want to show it) -->
            <!-- For now, leaving empty or adding a placeholder for future widgets -->
		</div>

		<!-- Monitors & Alert Rules Overview -->
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
			<!-- Monitors List (Preview) -->
			<div class="lg:col-span-2">
				<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
					<div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
						<h2 class="text-lg font-semibold text-slate-900">Monitors</h2>
						{#if monitors.length > 0}
							<span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700 border border-blue-200">{monitors.length} visible</span>
						{/if}
					</div>
					<div class="p-4" data-testid="dashboard-monitors-preview">
						{#if monitors.length === 0}
							<div class="p-6 text-center text-sm text-slate-500">No monitors to display</div>
						{:else}
							<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
								{#each monitors as mon (mon.id)}
									<MonitorCard monitor={mon} on:view={(e) => handleViewMonitor(e.detail)} />
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Alerts List (Preview) -->
			<div>
				<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
					<div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
						<h2 class="text-lg font-semibold text-slate-900">Alert Rules</h2>
						{#if alertRules.length > 0}
							<span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-700 border border-amber-200">{alertRules.length} visible</span>
						{/if}
					</div>
					<div class="p-4">
						<div data-testid="dashboard-alerts-preview">
						{#if alertRules.length === 0}
							<div class="p-6 text-center text-sm text-slate-500">No alert rules to display</div>
						{:else}
							<div class="grid grid-cols-1 gap-4">
								{#each alertRules as rule (rule.id)}
									<div on:click={() => handleViewAlert(rule)}>
										<AlertCard rule={rule} />
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

