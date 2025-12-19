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
	import { Line } from 'svelte-chartjs';
	import { Chart, type ChartData, type ChartOptions, type ScatterDataPoint, registerables } from 'chart.js';
	import 'chartjs-adapter-date-fns';
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

	interface DashboardPayload {
		stats: DashboardStats;
		recent_checks: MonitorCheckWithMonitor[];
		open_incidents: IncidentWithDetails[];
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
	let isLoading = false;
	let chartReady = browser;
	let pingHistory: { time: Date; response: number; monitor: string; success: boolean }[] = [];
	let chartData: ChartData<'line', ScatterDataPoint[]>;
	let statCards: { title: string; value: number | string; description: string; accent: string }[] = [];

	if (browser) {
		Chart.register(...registerables);
	}

	const chartOptions: ChartOptions<'line'> = {
		responsive: true,
		maintainAspectRatio: false,
		interaction: { mode: 'index', intersect: false },
		scales: {
			x: {
				type: 'time',
				time: {
					tooltipFormat: 'PPpp'
				},
				grid: { display: false },
				title: { display: true, text: 'Checked at' }
			},
			y: {
				beginAtZero: true,
				title: { display: true, text: 'Response time (ms)' },
				grid: { color: 'rgba(148, 163, 184, 0.2)' }
			}
		},
		plugins: {
			legend: { display: false },
			tooltip: {
				callbacks: {
					label: (context) => `${context.parsed?.y ?? 0} ms`
				}
			}
		}
	};

	$: pingHistory = recentChecks
		.map(({ check, monitor }) => ({
			time: new Date(check.checked_at),
			response: check.response_time_ms ?? 0,
			monitor: monitor.name,
			success: check.success
		}))
		.sort((a, b) => a.time.getTime() - b.time.getTime())
		.slice(-20);

	$: chartData = {
		datasets: [
			{
				label: 'Ping response',
				data: pingHistory.map((point) => ({ x: point.time.getTime(), y: point.response })),
				borderColor: 'rgb(56, 189, 248)',
				backgroundColor: 'rgba(56, 189, 248, 0.18)',
				tension: 0.35,
				borderWidth: 2,
				fill: true,
				pointRadius: 4,
				pointHoverRadius: 5,
				pointBackgroundColor: pingHistory.map((point) =>
					point.success ? 'rgb(34, 197, 94)' : 'rgb(248, 113, 113)'
				)
			}
		]
	} satisfies ChartData<'line', ScatterDataPoint[]>;

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
			title: 'Uptime (24h)',
			value: `${stats.overall_uptime?.toFixed(2) ?? '100.00'}%`,
			description: 'Weighted by traffic',
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
			}
		} catch (err) {
			console.error('Error loading dashboard:', err);
		} finally {
			isLoading = false;
		}
	}

	onMount(async () => {
		if (!chartReady && browser) {
			Chart.register(...registerables);
			chartReady = true;
		}
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
			At-a-glance health of your monitors with live ping history.
		</p>
	</div>

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

	<div class="grid gap-6 lg:grid-cols-3">
		<Card class="lg:col-span-2">
			<CardHeader class="flex flex-row items-center justify-between">
				<div>
					<CardTitle>Ping history</CardTitle>
					<CardDescription>Latest response times across your monitors</CardDescription>
				</div>
			</CardHeader>
			<CardContent class="h-[320px]">
				{#if chartReady && pingHistory.length}
					<Line data={chartData} options={chartOptions} />
				{:else if chartReady}
					<div class="flex h-full items-center justify-center rounded-md border border-dashed">
						<p class="text-sm text-muted-foreground">
							No ping history available yet. Checks will appear here automatically.
						</p>
					</div>
				{:else}
					<div class="flex h-full items-center justify-center rounded-md border border-dashed">
						<p class="text-sm text-muted-foreground">
							Chart will render automatically in the browser.
						</p>
					</div>
				{/if}
			</CardContent>
		</Card>

		<Card>
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
							{#each openIncidents.slice(0, 6) as incident}
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
							{#each recentChecks.slice(0, 8) as item}
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
				<CardTitle>Summary</CardTitle>
				<CardDescription>Performance snapshot</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="flex items-center justify-between rounded-lg border bg-muted/50 px-4 py-3">
					<div>
						<p class="text-sm text-muted-foreground">Avg response</p>
						<p class="text-xl font-semibold">
							{Math.round(stats.average_response_time ?? 0)} ms
						</p>
					</div>
					<span class="text-xs text-emerald-600 dark:text-emerald-300">Stable</span>
				</div>
				<div class="flex items-center justify-between rounded-lg border bg-muted/50 px-4 py-3">
					<div>
						<p class="text-sm text-muted-foreground">24h uptime</p>
						<p class="text-xl font-semibold">
							{stats.overall_uptime?.toFixed(2) ?? '100.00'}%
						</p>
					</div>
					<span class="text-xs text-indigo-600 dark:text-indigo-300">Auto-updated</span>
				</div>
				<div class="flex items-center justify-between rounded-lg border bg-muted/50 px-4 py-3">
					<div>
						<p class="text-sm text-muted-foreground">Monitors</p>
						<p class="text-xl font-semibold">{stats.total_monitors}</p>
					</div>
					<span class="text-xs text-cyan-600 dark:text-cyan-300">
						{stats.up_count} up / {stats.down_count} down
					</span>
				</div>
			</CardContent>
		</Card>
	</div>
</div>
