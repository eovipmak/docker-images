<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks, latestIncidents, connectEventStream, disconnectEventStream } from '$lib/api/events';
	import StatCard from '$lib/components/StatCard.svelte';
	import IncidentBadge from '$lib/components/IncidentBadge.svelte';
	import Card from '$lib/components/Card.svelte';
	import { goto } from '$app/navigation';
	import type { MaintenanceWindow } from '$lib/types';

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
	let activeMaintenanceWindows: MaintenanceWindow[] = [];
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
		} catch (err) {
			console.error('Error loading dashboard:', err);
			error = 'An error occurred while loading dashboard data';
		} finally {
			isLoading = false;
		}
	}

	// Load maintenance windows and filter active ones
	async function loadMaintenanceWindows() {
		try {
			const response = await fetchAPI('/api/v1/maintenance-windows');
			if (response.ok) {
				const windows: MaintenanceWindow[] = await response.json();
				const now = new Date();
				// Filter to only active maintenance windows (start_time <= now <= end_time)
				activeMaintenanceWindows = windows.filter(w => {
					const startTime = new Date(w.start_time);
					const endTime = new Date(w.end_time);
					return startTime <= now && now <= endTime;
				});
			}
		} catch (err) {
			console.error('Error loading maintenance windows:', err);
		}
	}

	function handleViewAlert(rule: any) {
		goto('/alerts');
	}

	onMount(async () => {
		await Promise.all([
			loadDashboardData(),
			loadMaintenanceWindows()
		]);

		// Start SSE connection for real-time updates
		await connectEventStream();

		// Subscribe to monitor check events (for reference, but don't auto-refresh)
		unsubscribeChecks = latestMonitorChecks.subscribe((checks) => {
			// Don't auto-refresh dashboard - only monitor detail page uses SSE
		});

		// Subscribe to incident events (for reference, but don't auto-refresh)
		unsubscribeIncidents = latestIncidents.subscribe((incidents) => {
			// Don't auto-refresh dashboard - only monitor detail page uses SSE
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

	// Format time remaining for maintenance window
	function formatTimeRemaining(endTimeStr: string): string {
		const endTime = new Date(endTimeStr);
		const now = new Date();
		const diffMs = endTime.getTime() - now.getTime();
		
		if (diffMs <= 0) return 'Ending soon';
		
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffDays > 0) return `${diffDays}d ${diffHours % 24}h remaining`;
		if (diffHours > 0) return `${diffHours}h ${diffMins % 60}m remaining`;
		return `${diffMins}m remaining`;
	}
</script>

<svelte:head>
	<title>Dashboard - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">Dashboard</h1>
			<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Overview of your monitoring status and system metrics.</p>
		</div>
	</div>

	{#if isLoading}
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
            {#each Array(6) as _}
                <div class="h-24 sm:h-32 bg-slate-100 dark:bg-slate-800 rounded-xl animate-pulse"></div>
            {/each}
        </div>
	{/if}
	{#if !isLoading}
		<!-- Stats Cards -->
		<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 sm:gap-4">
			<StatCard 
                title="Total Monitors" 
                value={stats.total_monitors} 
                valueColor="text-slate-900 dark:text-white" 
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

		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6 lg:gap-8">
			<!-- Open Incidents -->
			<div class="lg:col-span-2 space-y-6">
				<Card>
					<div slot="header" class="px-4 sm:px-6 py-3 sm:py-4 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between bg-slate-50/50 dark:bg-slate-800/50">
						<h2 class="text-base sm:text-lg font-semibold text-slate-900 dark:text-white">Open Incidents</h2>
						{#if openIncidents.length > 0}
							<span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-rose-100 dark:bg-rose-900/30 text-rose-700 dark:text-rose-300 border border-rose-200 dark:border-rose-800">
								{openIncidents.length} Active
							</span>
						{/if}
					</div>
                    
					{#if openIncidents.length === 0}
						<div class="p-12 text-center">
                            <div class="inline-flex items-center justify-center w-12 h-12 rounded-full bg-emerald-100 dark:bg-emerald-900/30 mb-4">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 text-emerald-600 dark:text-emerald-400">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <h3 class="text-sm font-medium text-slate-900 dark:text-white">All systems operational</h3>
                            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">No active incidents reported at this time.</p>
                        </div>
                    {:else}
                        <div class="divide-y divide-slate-100 dark:divide-slate-700">
                            {#each openIncidents as { incident, monitor }}
                                <button 
                                    type="button"
                                    on:click={() => goto(`/incidents/${incident.id}`)}
                                    class="w-full text-left p-4 sm:p-6 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer">
                                    <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2 sm:gap-4">
                                        <div class="flex-1 min-w-0">
                                            <div class="flex flex-wrap items-center gap-2 sm:gap-3 mb-1">
                                                <IncidentBadge 
                                                    status={incident.status} 
                                                    severity={getIncidentSeverity(incident.trigger_value)} 
                                                />
                                                <span class="font-medium text-slate-900 dark:text-white truncate">{monitor.name}</span>
                                            </div>
                                            <p class="text-sm text-slate-500 dark:text-slate-400 truncate mb-2">{monitor.url}</p>
                                            {#if incident.trigger_value}
                                                <div class="inline-flex items-center px-2 py-1 rounded bg-slate-100 dark:bg-slate-700 text-xs text-slate-600 dark:text-slate-300">
                                                    <span class="font-medium mr-1">Trigger:</span> {incident.trigger_value}
                                                </div>
                                            {/if}
                                        </div>
                                        <div class="text-left sm:text-right flex-shrink-0">
                                            <p class="text-xs font-medium text-slate-500 dark:text-slate-400">
                                                Started {formatRelativeTime(incident.started_at)}
                                            </p>
                                            <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5 sm:mt-1 hidden sm:block">
                                                {formatDate(incident.started_at)}
                                            </p>
                                        </div>
                                    </div>
                                </button>
                            {/each}
                        </div>
					{/if}
				</Card>
			</div>
            
            <!-- Right column: Maintenance Windows -->
			<div class="space-y-6">
				<Card>
					<div slot="header" class="px-4 sm:px-6 py-3 sm:py-4 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between bg-slate-50/50 dark:bg-slate-800/50">
						<h2 class="text-base sm:text-lg font-semibold text-slate-900 dark:text-white">Active Maintenance</h2>
						{#if activeMaintenanceWindows.length > 0}
							<span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-300 border border-amber-200 dark:border-amber-800">
								{activeMaintenanceWindows.length} Active
							</span>
						{/if}
					</div>

					{#if activeMaintenanceWindows.length === 0}
						<div class="p-8 text-center">
							<div class="inline-flex items-center justify-center w-10 h-10 rounded-full bg-slate-100 dark:bg-slate-700 mb-3">
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-slate-400">
									<path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z" />
								</svg>
							</div>
							<h3 class="text-sm font-medium text-slate-900 dark:text-white">No active maintenance</h3>
							<p class="mt-1 text-xs text-slate-500 dark:text-slate-400">All systems are running normally.</p>
						</div>
					{:else}
						<div class="divide-y divide-slate-100 dark:divide-slate-700">
							{#each activeMaintenanceWindows as window}
								<button 
									type="button"
									on:click={() => goto('/settings/maintenance')}
									class="w-full text-left p-4 hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors cursor-pointer"
								>
									<div class="flex items-start gap-3">
										<div class="flex-shrink-0 mt-0.5">
											<div class="w-8 h-8 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
												<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 text-amber-600 dark:text-amber-400">
													<path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z" />
												</svg>
											</div>
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-slate-900 dark:text-white truncate">{window.name}</p>
											<p class="text-xs text-amber-600 dark:text-amber-400 mt-0.5">{formatTimeRemaining(window.end_time)}</p>
											{#if window.tags && window.tags.length > 0}
												<div class="flex flex-wrap gap-1 mt-1.5">
													{#each window.tags.slice(0, 2) as tag}
														<span class="px-1.5 py-0.5 text-xs bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 rounded">{tag}</span>
													{/each}
													{#if window.tags.length > 2}
														<span class="px-1.5 py-0.5 text-xs bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 rounded">+{window.tags.length - 2}</span>
													{/if}
												</div>
											{/if}
										</div>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</Card>
			</div>
		</div>

	{/if}
</div>