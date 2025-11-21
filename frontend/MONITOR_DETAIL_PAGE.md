# Monitor Detail Page Setup

This file contains instructions for setting up the monitor detail page.

## Quick Setup

Run this command from the repository root:

```bash
mkdir -p frontend/src/routes/domains/\[id\] && cp /tmp/monitor-detail-page.svelte frontend/src/routes/domains/\[id\]/+page.svelte
```

Or manually:

1. Create directory: `mkdir -p frontend/src/routes/domains/[id]`
2. Copy file: `cp /tmp/monitor-detail-page.svelte frontend/src/routes/domains/[id]/+page.svelte`

## Note

The file content has been created at `/tmp/monitor-detail-page.svelte` and needs to be moved to the correct location in the SvelteKit routes directory.

## Features

The monitor detail page includes:
- Monitor status and statistics
- Uptime history visualization (last 24 hours)
- Response time chart
- SSL certificate information (for HTTPS monitors)
- Monitor settings display
- Recent checks table with detailed information

---

## Original File Content (for reference)

```svelte
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import {
		Chart,
		LineController,
		LineElement,
		PointElement,
		LinearScale,
		TimeScale,
		Title,
		Tooltip,
		Legend,
		CategoryScale
	} from 'chart.js';
	import 'chartjs-adapter-date-fns';

	// Register Chart.js components
	Chart.register(
		LineController,
		LineElement,
		PointElement,
		LinearScale,
		TimeScale,
		CategoryScale,
		Title,
		Tooltip,
		Legend
	);

	let monitorId: string;
	let monitor: any = null;
	let checks: any[] = [];
	let sslStatus: any = null;
	let isLoading = true;
	let error = '';

	let checksChart: Chart | null = null;
	let responseTimeChart: Chart | null = null;
	let checksCanvas: HTMLCanvasElement;
	let responseTimeCanvas: HTMLCanvasElement;

	$: monitorId = $page.params.id;

	onMount(() => {
		loadMonitorData();
	});

	onDestroy(() => {
		if (checksChart) checksChart.destroy();
		if (responseTimeChart) responseTimeChart.destroy();
	});

	async function loadMonitorData() {
		isLoading = true;
		error = '';

		try {
			// Load monitor details
			const monitorResponse = await fetchAPI(`/api/v1/monitors/${monitorId}`);
			if (!monitorResponse.ok) {
				throw new Error('Failed to load monitor');
			}
			monitor = await monitorResponse.json();

			// Load check history
			const checksResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/checks`);
			if (checksResponse.ok) {
				checks = await checksResponse.json();
			}

			// Load SSL status if applicable
			if (monitor.check_ssl && monitor.url.startsWith('https')) {
				const sslResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/ssl-status`);
				if (sslResponse.ok) {
					sslStatus = await sslResponse.json();
				}
			}

			// Render charts after data is loaded
			setTimeout(() => {
				renderCharts();
			}, 100);
		} catch (err: any) {
			console.error('Error loading monitor data:', err);
			error = err.message || 'Failed to load monitor data';
		} finally {
			isLoading = false;
		}
	}

	function renderCharts() {
		if (!checks || checks.length === 0) return;

		// Prepare data - sort by timestamp
		const sortedChecks = [...checks].sort(
			(a, b) => new Date(a.checked_at).getTime() - new Date(b.checked_at).getTime()
		);

		// Status chart data
		const statusData = sortedChecks.map((check) => ({
			x: new Date(check.checked_at),
			y: check.success ? 1 : 0
		}));

		// Response time chart data
		const responseTimeData = sortedChecks
			.filter((check) => check.response_time_ms !== null && check.response_time_ms !== undefined)
			.map((check) => ({
				x: new Date(check.checked_at),
				y: check.response_time_ms
			}));

		// Render status chart
		if (checksCanvas && checksChart) {
			checksChart.destroy();
		}
		if (checksCanvas) {
			checksChart = new Chart(checksCanvas, {
				type: 'line',
				data: {
					datasets: [
						{
							label: 'Status',
							data: statusData,
							borderColor: 'rgb(34, 197, 94)',
							backgroundColor: 'rgba(34, 197, 94, 0.1)',
							stepped: true,
							fill: true,
							pointRadius: 0,
							borderWidth: 2
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						title: {
							display: true,
							text: 'Uptime Status (Last 24 Hours)'
						},
						legend: {
							display: false
						},
						tooltip: {
							callbacks: {
								label: function (context) {
									return context.parsed.y === 1 ? 'Up' : 'Down';
								}
							}
						}
					},
					scales: {
						x: {
							type: 'time',
							time: {
								unit: 'hour',
								displayFormats: {
									hour: 'HH:mm'
								}
							},
							title: {
								display: true,
								text: 'Time'
							}
						},
						y: {
							min: 0,
							max: 1,
							ticks: {
								callback: function (value) {
									return value === 1 ? 'Up' : 'Down';
								},
								stepSize: 1
							}
						}
					}
				}
			});
		}

		// Render response time chart
		if (responseTimeCanvas && responseTimeChart) {
			responseTimeChart.destroy();
		}
		if (responseTimeCanvas && responseTimeData.length > 0) {
			responseTimeChart = new Chart(responseTimeCanvas, {
				type: 'line',
				data: {
					datasets: [
						{
							label: 'Response Time (ms)',
							data: responseTimeData,
							borderColor: 'rgb(59, 130, 246)',
							backgroundColor: 'rgba(59, 130, 246, 0.1)',
							fill: true,
							tension: 0.4,
							pointRadius: 2,
							borderWidth: 2
						}
					]
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						title: {
							display: true,
							text: 'Response Time (Last 24 Hours)'
						},
						legend: {
							display: false
						}
					},
					scales: {
						x: {
							type: 'time',
							time: {
								unit: 'hour',
								displayFormats: {
									hour: 'HH:mm'
								}
							},
							title: {
								display: true,
								text: 'Time'
							}
						},
						y: {
							beginAtZero: true,
							title: {
								display: true,
								text: 'Response Time (ms)'
							}
						}
					}
				}
			});
		}
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleString();
	}

	function getMonitorStatus(): 'up' | 'down' | 'unknown' {
		if (!checks || checks.length === 0) return 'unknown';
		const lastCheck = checks[0];
		return lastCheck.success ? 'up' : 'down';
	}

	function calculateUptime(): string {
		if (!checks || checks.length === 0) return 'N/A';
		const successCount = checks.filter((check) => check.success).length;
		const percentage = (successCount / checks.length) * 100;
		return `${percentage.toFixed(2)}%`;
	}

	function getAverageResponseTime(): string {
		if (!checks || checks.length === 0) return 'N/A';
		const responseTimes = checks
			.filter((check) => check.response_time_ms !== null && check.response_time_ms !== undefined)
			.map((check) => check.response_time_ms);

		if (responseTimes.length === 0) return 'N/A';

		const avg = responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length;
		return `${Math.round(avg)}ms`;
	}

	function handleBack() {
		goto('/domains');
	}
</script>

<svelte:head>
	<title>{monitor?.name || 'Monitor Details'} - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
			{error}
		</div>
		<button
			on:click={handleBack}
			class="text-blue-600 hover:text-blue-800 font-medium"
		>
			← Back to Monitors
		</button>
	{:else if monitor}
		<!-- Header -->
		<div class="mb-6">
			<button
				on:click={handleBack}
				class="text-blue-600 hover:text-blue-800 font-medium mb-4 inline-flex items-center"
			>
				← Back to Monitors
			</button>
			<div class="flex justify-between items-start">
				<div>
					<h1 class="text-3xl font-bold text-gray-900 mb-2">{monitor.name}</h1>
					<p class="text-gray-600">{monitor.url}</p>
				</div>
				<MonitorStatus status={getMonitorStatus()} />
			</div>
		</div>

		<!-- Stats Cards -->
		<div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Status</h3>
				<p class="text-2xl font-bold text-gray-900">
					{monitor.enabled ? 'Enabled' : 'Disabled'}
				</p>
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Uptime</h3>
				<p class="text-2xl font-bold text-gray-900">{calculateUptime()}</p>
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Avg Response Time</h3>
				<p class="text-2xl font-bold text-gray-900">{getAverageResponseTime()}</p>
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Check Interval</h3>
				<p class="text-2xl font-bold text-gray-900">{monitor.check_interval}s</p>
			</div>
		</div>

		<!-- Charts -->
		<div class="grid grid-cols-1 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow-md p-6">
				<div style="height: 300px;">
					<canvas bind:this={checksCanvas}></canvas>
				</div>
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<div style="height: 300px;">
					<canvas bind:this={responseTimeCanvas}></canvas>
				</div>
			</div>
		</div>

		<!-- SSL Information -->
		{#if monitor.check_ssl && monitor.url.startsWith('https') && sslStatus}
			<div class="bg-white rounded-lg shadow-md p-6 mb-8">
				<h2 class="text-xl font-bold text-gray-900 mb-4">SSL Certificate</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<p class="text-sm text-gray-500">Valid</p>
						<p class="font-medium text-gray-900">
							{sslStatus.ssl_valid ? 'Yes' : 'No'}
						</p>
					</div>
					{#if sslStatus.ssl_expires_at}
						<div>
							<p class="text-sm text-gray-500">Expires At</p>
							<p class="font-medium text-gray-900">{formatDate(sslStatus.ssl_expires_at)}</p>
						</div>
					{/if}
					{#if sslStatus.ssl_issuer}
						<div>
							<p class="text-sm text-gray-500">Issuer</p>
							<p class="font-medium text-gray-900">{sslStatus.ssl_issuer}</p>
						</div>
					{/if}
					{#if sslStatus.error_message}
						<div class="col-span-2">
							<p class="text-sm text-gray-500">Error</p>
							<p class="font-medium text-red-600">{sslStatus.error_message}</p>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Monitor Settings -->
		<div class="bg-white rounded-lg shadow-md p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Settings</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<p class="text-sm text-gray-500">Timeout</p>
					<p class="font-medium text-gray-900">{monitor.timeout}s</p>
				</div>
				<div>
					<p class="text-sm text-gray-500">SSL Checks</p>
					<p class="font-medium text-gray-900">{monitor.check_ssl ? 'Enabled' : 'Disabled'}</p>
				</div>
				{#if monitor.check_ssl}
					<div>
						<p class="text-sm text-gray-500">SSL Alert (Days Before Expiry)</p>
						<p class="font-medium text-gray-900">{monitor.ssl_alert_days} days</p>
					</div>
				{/if}
				<div>
					<p class="text-sm text-gray-500">Last Checked</p>
					<p class="font-medium text-gray-900">
						{monitor.last_checked_at ? formatDate(monitor.last_checked_at) : 'Never'}
					</p>
				</div>
			</div>
		</div>

		<!-- Recent Checks -->
		{#if checks && checks.length > 0}
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Recent Checks</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Time</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status Code</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Response Time</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Error</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each checks.slice(0, 10) as check}
								<tr>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
										{formatDate(check.checked_at)}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<MonitorStatus status={check.success ? 'up' : 'down'} showText={true} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
										{check.status_code || 'N/A'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
										{check.response_time_ms ? `${check.response_time_ms}ms` : 'N/A'}
									</td>
									<td class="px-6 py-4 text-sm text-red-600">
										{check.error_message || '-'}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	{/if}
</div>
```
