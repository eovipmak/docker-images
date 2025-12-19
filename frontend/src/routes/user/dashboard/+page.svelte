<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { fetchAPI } from '$lib/api/client';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow,
		TableCaption
	} from '$lib/components/ui/table';
	import Sparkline from '$lib/components/Sparkline.svelte';
	import type { PageData } from './$types';

	interface DashboardStats {
		total_monitors: number;
		up_count: number;
		down_count: number;
		open_incidents: number;
		average_response_time?: number;
		overall_uptime?: number;
	}

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
		user_id: number;
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
		status: string;
		trigger_value?: string;
		notified_at?: string;
		created_at: string;
	}

	interface MonitorCheckWithMonitor {
		check: MonitorCheck;
		monitor: Monitor;
	}

	interface IncidentWithDetails {
		incident: Incident;
		monitor: Monitor;
	}

	interface MonitorWithSparkline {
		monitor: Monitor;
		recent_checks: MonitorCheck[];
		status: string;
	}

	interface DashboardPayload {
		stats: DashboardStats;
		recent_checks: MonitorCheckWithMonitor[];
		open_incidents: IncidentWithDetails[];
		monitors: MonitorWithSparkline[];
	}

	export let data: PageData;

	let stats: DashboardStats = data?.stats ?? {
		total_monitors: 0,
		up_count: 0,
		down_count: 0,
		open_incidents: 0,
		average_response_time: 0,
		overall_uptime: 0
	};
	let recentChecks: MonitorCheckWithMonitor[] = data?.recentChecks ?? [];
	let openIncidents: IncidentWithDetails[] = data?.openIncidents ?? [];
	let monitors: MonitorWithSparkline[] = [];
	let isLoading = false;
	let statCards: { title: string; value: number | string; description: string; accent: string }[] = [];

	$: statCards = [
		{
			title: 'Total Monitors',
			value: stats.total_monitors,
			description: 'Active across your workspace',
			accent: 'from-cyan-500 to-blue-500'
		},
		{
			title: 'Open Incidents',
			value: stats.open_incidents,
			description: 'Issues requiring attention',
			accent: 'from-rose-500 to-orange-500'
		},
		{
			title: 'Uptime',
			value: `${stats.overall_uptime?.toFixed(2) ?? '100.00'}%`,
			description: 'Last 24 hours',
			accent: 'from-emerald-500 to-teal-500'
		},
		{
			title: 'Avg Response',
			value: `${Math.round(stats.average_response_time ?? 0)} ms`,
			description: 'Latest ping window',
			accent: 'from-indigo-500 to-purple-500'
		},
		{
			title: 'Operational',
			value: stats.up_count,
			description: 'Healthy services',
			accent: 'from-emerald-500 to-lime-500'
		},
		{
			title: 'Down',
			value: stats.down_count,
			description: 'Investigate quickly',
			accent: 'from-red-500 to-amber-500'
		}
	];

	async function loadDashboardData() {
		isLoading = true;
		try {
			const response = await fetchAPI('/api/v1/dashboard');
			if (response.ok) {
				const payload: DashboardPayload = await response.json();
				stats = payload.stats ?? stats;
				recentChecks = payload.recent_checks ?? recentChecks;
				openIncidents = payload.open_incidents ?? openIncidents;
				monitors = payload.monitors ?? [];
			}
		} catch (err) {
			console.error('Error loading dashboard:', err);
		} finally {
			isLoading = false;
		}
	}

	function getSparklineData(checks: MonitorCheck[]) {
		return checks.map(check => ({
			time: new Date(check.checked_at),
			value: check.response_time_ms ?? 0,
			success: check.success
		})).sort((a, b) => a.time.getTime() - b.time.getTime());
	}

	onMount(async () => {
		await loadDashboardData();
	});
</script>

<svelte:head>
	<title>Dashboard - V-Insight</title>
</svelte:head>

<div class="space-y-8 pb-10 pt-4">
	<div class="space-y-1">
		<p class="text-sm text-muted-foreground">Trust &amp; Visual</p>
		<h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
		<p class="text-muted-foreground">
			At-a-glance health of your monitors with live status.
		</p>
	</div>

	<!-- Top Metrics Grid -->
	<div class="grid gap-6 lg:grid-cols-3">
		{#each statCards as card (card.title)}
			<Card class="relative overflow-hidden border border-border/70">
				<div class="absolute inset-0 opacity-60 blur-3xl bg-gradient-to-r {card.accent}"></div>
				<CardHeader class="relative flex flex-row items-start justify-between pb-2">
					<div>
						<CardDescription>{card.description}</CardDescription>
						<CardTitle class="text-xl">{card.title}</CardTitle>
					</div>
					<div
						class={`h-11 w-11 shrink-0 rounded-full bg-gradient-to-br ${card.accent} shadow-lg ring-1 ring-white/30 dark:ring-black/40 flex items-center justify-center text-white text-sm font-semibold`}
					>
						{card.value}
					</div>
				</CardHeader>
				<CardContent class="relative pt-2">
					<p class="text-3xl font-semibold leading-tight">{card.value}</p>
					{#if isLoading}
						<p class="text-xs text-muted-foreground mt-1">Refreshing...</p>
					{/if}
				</CardContent>
			</Card>
		{/each}
	</div>

	<!-- All Monitors Grid - Replaces Ping History Chart -->
	<Card>
		<CardHeader>
			<CardTitle>All Monitors</CardTitle>
			<CardDescription>Monitor status with response time trends</CardDescription>
		</CardHeader>
		<CardContent class="p-0">
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Monitor</TableHead>
						<TableHead>URL</TableHead>
						<TableHead>Status</TableHead>
						<TableHead>Last Response</TableHead>
						<TableHead>Trend (20 checks)</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{#if monitors.length === 0}
						<TableRow>
							<TableCell colspan={5} class="text-muted-foreground text-center py-8">
								No monitors configured yet. Add your first monitor to get started.
							</TableCell>
						</TableRow>
					{:else}
						{#each monitors as item}
							{@const sparklineData = getSparklineData(item.recent_checks)}
							{@const lastCheck = item.recent_checks.length > 0 ? item.recent_checks[0] : null}
							<TableRow>
								<TableCell class="font-medium">{item.monitor.name}</TableCell>
								<TableCell class="text-sm text-muted-foreground max-w-xs truncate">
									{item.monitor.url}
								</TableCell>
								<TableCell>
									<span class={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold ${
										item.status === 'up'
											? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200'
											: item.status === 'down'
											? 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-200'
											: 'bg-slate-100 text-slate-700 dark:bg-slate-500/15 dark:text-slate-200'
									}`}>
										{item.status}
									</span>
								</TableCell>
								<TableCell class="text-sm">
									{lastCheck?.response_time_ms ? `${lastCheck.response_time_ms} ms` : '—'}
								</TableCell>
								<TableCell>
									{#if sparklineData.length > 0}
										<div class="text-cyan-500 dark:text-cyan-400">
											<Sparkline data={sparklineData} width={120} height={32} />
										</div>
									{:else}
										<span class="text-xs text-muted-foreground">No data</span>
									{/if}
								</TableCell>
							</TableRow>
						{/each}
					{/if}
				</TableBody>
			</Table>
		</CardContent>
	</Card>

	<!-- Recent Checks and Open Incidents - Expanded to fill space -->
	<div class="grid gap-6 lg:grid-cols-2">
		<Card class="col-span-1">
			<CardHeader>
				<CardTitle>Recent checks</CardTitle>
				<CardDescription>Latest monitor responses</CardDescription>
			</CardHeader>
			<CardContent class="p-0">
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>Monitor</TableHead>
							<TableHead>Status</TableHead>
							<TableHead>Response</TableHead>
							<TableHead>Checked at</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#if recentChecks.length === 0}
							<TableRow>
								<TableCell colspan={4} class="text-muted-foreground">
									No checks recorded yet.
								</TableCell>
							</TableRow>
						{:else}
							{#each recentChecks.slice(0, 10) as item}
								<TableRow>
									<TableCell class="font-medium">{item.monitor.name}</TableCell>
									<TableCell>
										<span class={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold ${
											item.check.success
												? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-200'
												: 'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-200'
										}`}>
											{item.check.success ? 'Up' : 'Down'}
										</span>
									</TableCell>
									<TableCell class="text-sm">
										{item.check.response_time_ms ? `${item.check.response_time_ms} ms` : '—'}
									</TableCell>
									<TableCell class="text-sm text-muted-foreground">
										{new Date(item.check.checked_at).toLocaleString()}
									</TableCell>
								</TableRow>
							{/each}
						{/if}
					</TableBody>
				</Table>
			</CardContent>
		</Card>

		<Card class="col-span-1">
			<CardHeader>
				<CardTitle>Open incidents</CardTitle>
				<CardDescription>Active issues that need attention</CardDescription>
			</CardHeader>
			<CardContent class="p-0">
				<Table>
					<TableCaption class="text-left px-4">
						{openIncidents.length
							? 'Live incident stream'
							: 'All clear — no open incidents.'}
					</TableCaption>
					<TableHeader>
						<TableRow>
							<TableHead>Monitor</TableHead>
							<TableHead>Started</TableHead>
							<TableHead>Status</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{#if openIncidents.length === 0}
							<TableRow>
								<TableCell colspan={3} class="text-muted-foreground">
									Everything looks stable right now.
								</TableCell>
							</TableRow>
						{:else}
							{#each openIncidents.slice(0, 10) as incident}
								<TableRow>
									<TableCell class="font-medium">
										{incident.monitor.name}
									</TableCell>
									<TableCell class="text-sm text-muted-foreground">
										{new Date(incident.incident.started_at).toLocaleString()}
									</TableCell>
									<TableCell>
										<span class="inline-flex items-center rounded-full bg-rose-100 px-2 py-0.5 text-xs font-medium text-rose-700 dark:bg-rose-500/15 dark:text-rose-200">
											{incident.incident.status}
										</span>
									</TableCell>
								</TableRow>
							{/each}
						{/if}
					</TableBody>
				</Table>
			</CardContent>
		</Card>
	</div>
</div>
