<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import { extractInt64, extractString, isValidSqlNull } from '$lib/utils/sqlNull';

	let monitorId: string = '';
	let monitor: any = null;
	let checks: any[] = [];
	let sslStatus: any = null;
	let metrics24h: any = null;
	let metrics7d: any = null;
	let metrics30d: any = null;
	let isLoading = true;
	let error = '';
	let uptimePeriod: '7d' | '30d' = '7d';
	let responseTimePeriod: '1h' | '6h' | '12h' | '24h' | '1w' = '24h';
	let chartType: 'uptime' | 'response' = 'uptime';

	// Lazy loaded chart components
	let LineChart: any = null;
	let DonutChart: any = null;
	let chartsLoaded = false;

	// Auto-refresh settings
	let autoRefreshInterval = 300; // seconds (5 minutes)
	let autoRefreshTimer: ReturnType<typeof setInterval> | null = null;

	$: monitorId = $page.params.id || '';

	// Computed uptime data based on selected period
	$: currentUptimeData = uptimePeriod === '7d' ? metrics7d?.uptime : metrics30d?.uptime;

	onMount(async () => {
		loadMonitorData();
		startAutoRefresh();
		// Lazy load chart components
		loadCharts();
	});

	onDestroy(() => {
		stopAutoRefresh();
	});

	async function loadCharts() {
		try {
			const [lineChartModule, donutChartModule] = await Promise.all([
				import('$lib/components/charts/LineChart.svelte'),
				import('$lib/components/charts/DonutChart.svelte')
			]);
			LineChart = lineChartModule.default;
			DonutChart = donutChartModule.default;
			chartsLoaded = true;
		} catch (err) {
			console.error('Failed to load chart components:', err);
		}
	}

	async function loadMonitorData() {
		isLoading = true;
		error = '';

		try {
			const monitorResponse = await fetchAPI(`/api/v1/monitors/${monitorId}`);
			if (!monitorResponse.ok) {
				throw new Error('Failed to load monitor');
			}
			monitor = await monitorResponse.json();

			const checksResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/checks?limit=100`);
			if (checksResponse.ok) {
				checks = await checksResponse.json();
			}

			// Load metrics for different periods
			const [metrics24hRes, metrics7dRes, metrics30dRes] = await Promise.all([
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=24h`),
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=7d`),
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=30d`)
			]);

			if (metrics24hRes.ok) {
				metrics24h = await metrics24hRes.json();
			}
			if (metrics7dRes.ok) {
				metrics7d = await metrics7dRes.json();
			}
			if (metrics30dRes.ok) {
				metrics30d = await metrics30dRes.json();
			}

			if (monitor.check_ssl && monitor.url.startsWith('https')) {
				const sslResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/ssl-status`);
				if (sslResponse.ok) {
					const data = await sslResponse.json();
					sslStatus = data.ssl_status;
				}
			}
		} catch (err: any) {
			console.error('Error loading monitor data:', err);
			error = err.message || 'Failed to load monitor data';
		} finally {
			isLoading = false;
		}
	}

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return 'Invalid Date';
		return date.toLocaleString();
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
			.filter((check) => isValidSqlNull(check.response_time_ms))
			.map((check) => extractInt64(check.response_time_ms, 0));

		if (responseTimes.length === 0) return 'N/A';

		const avg = responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length;
		return `${Math.round(avg)}ms`;
	}

	function handleBack() {
		goto('/domains');
	}

	// Start auto-refresh timer
	function startAutoRefresh() {
		if (autoRefreshTimer) {
			clearInterval(autoRefreshTimer);
		}
		autoRefreshTimer = setInterval(() => {
			console.log(`[Monitor Details] Auto-refreshing data (${autoRefreshInterval}s interval)`);
			loadMonitorData();
		}, autoRefreshInterval * 1000);
	}

	// Stop auto-refresh timer
	function stopAutoRefresh() {
		if (autoRefreshTimer) {
			clearInterval(autoRefreshTimer);
			autoRefreshTimer = null;
		}
	}

	// Handle interval change
	function handleIntervalChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		autoRefreshInterval = parseInt(target.value);
		startAutoRefresh(); // Restart with new interval
	}

	// Get response time data based on selected period
	function getResponseTimeData() {
		switch (responseTimePeriod) {
			case '1h':
				// For 1h, we might need to filter recent data or fetch from API
				// For now, return empty array to use fallback logic
				return [];
			case '6h':
				return [];
			case '12h':
				return [];
			case '24h':
				return metrics24h?.response_time_history || [];
			case '1w':
				return metrics7d?.response_time_history || [];
			default:
				return metrics24h?.response_time_history || [];
		}
	}

	// Get filtered response times based on period
	function getFilteredResponseTimes() {
		if (!checks || checks.length === 0) return [];

		let limit = 48; // Default for 24h (assuming 30min intervals)
		switch (responseTimePeriod) {
			case '1h':
				limit = 2; // Assuming 30min intervals
				break;
			case '6h':
				limit = 12;
				break;
			case '12h':
				limit = 24;
				break;
			case '24h':
				limit = 48;
				break;
			case '1w':
				limit = 336; // 7 days * 48 checks per day
				break;
		}

		return checks
			.filter((c) => isValidSqlNull(c.response_time_ms))
			.slice(0, limit)
			.reverse()
			.map((c) => extractInt64(c.response_time_ms, 0));
	}

	// Get period label for display
	function getPeriodLabel() {
		switch (responseTimePeriod) {
			case '1h':
				return '1h';
			case '6h':
				return '6h';
			case '12h':
				return '12h';
			case '24h':
				return '24h';
			case '1w':
				return '1w';
			default:
				return '24h';
		}
	}
</script>

<svelte:head>
	<title>{monitor?.name || 'Monitor Details'} - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto space-y-6">
	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-4 flex items-center">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
			</svg>
			{error}
		</div>
		<button on:click={handleBack} class="text-blue-600 hover:text-blue-800 font-medium flex items-center transition-colors">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
			</svg>
			Back to Monitors
		</button>
	{:else if monitor}
		<!-- Header -->
		<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
			<div class="flex items-center gap-4">
				<button 
					on:click={handleBack}
					class="p-2 rounded-lg hover:bg-slate-100 text-slate-500 transition-colors"
					aria-label="Back to monitors"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
					</svg>
				</button>
				<div>
					<div class="flex items-center gap-3">
						<h1 class="text-2xl font-bold text-slate-900">{monitor.name}</h1>
						<MonitorStatus status={getMonitorStatus()} />
					</div>
					<a href={monitor.url} target="_blank" rel="noopener noreferrer" class="text-slate-500 hover:text-blue-600 hover:underline text-sm mt-1 flex items-center gap-1">
						{monitor.url}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>
				</div>
			</div>
			
			<div class="flex items-center gap-3">
				<div class="flex items-center gap-2 bg-white px-3 py-1.5 rounded-lg border border-slate-200 shadow-sm">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
					<label for="refresh-interval" class="text-sm text-slate-600 font-medium">Refresh:</label>
					<select 
						id="refresh-interval" 
						bind:value={autoRefreshInterval} 
						on:change={handleIntervalChange}
						class="text-sm border-none p-0 focus:ring-0 text-slate-900 font-semibold bg-transparent cursor-pointer"
					>
						<option value={15}>15s</option>
						<option value={30}>30s</option>
						<option value={60}>1m</option>
						<option value={300}>5m</option>
						<option value={900}>15m</option>
					</select>
				</div>
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Status Overview -->
			<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
				<div class="flex items-center justify-between mb-2">
					<h3 class="text-sm font-medium text-slate-500">Current Status</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					<span class="text-2xl font-bold text-slate-900 capitalize">{getMonitorStatus()}</span>
				</div>
			</div>

			<!-- Uptime -->
			<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
				<div class="flex items-center justify-between mb-2">
					<h3 class="text-sm font-medium text-slate-500">Uptime (24h)</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					<span class="text-2xl font-bold text-slate-900">{calculateUptime()}</span>
				</div>
			</div>

			<!-- Response Time -->
			<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
				<div class="flex items-center justify-between mb-2">
					<h3 class="text-sm font-medium text-slate-500">Avg Response</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					<span class="text-2xl font-bold text-slate-900">{getAverageResponseTime()}</span>
				</div>
			</div>
			
			<!-- SSL Information -->
			<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-5">
				<div class="flex items-center justify-between mb-2">
					<h3 class="text-sm font-medium text-slate-500">SSL Status</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					</div>
				</div>
				{#if monitor.check_ssl && sslStatus}
					<div class="flex flex-col">
						<span class="text-2xl font-bold {sslStatus.valid ? 'text-emerald-600' : 'text-rose-600'}">
							{sslStatus.valid ? 'Valid' : 'Invalid'}
						</span>
						{#if sslStatus.expires_at}
							<span class="text-xs text-slate-500 mt-1">
								Expires: {new Date(sslStatus.expires_at).toLocaleDateString()}
							</span>
						{/if}
					</div>
				{:else if monitor.check_ssl}
					<span class="text-lg font-medium text-slate-500">Checking...</span>
				{:else}
					<span class="text-lg font-medium text-slate-400">Disabled</span>
				{/if}
			</div>
		</div>

		<!-- History Charts -->
		<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-lg font-bold text-slate-900">History</h2>
				<div class="flex bg-slate-100 rounded-lg p-1">
					<button
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {chartType === 'uptime' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
						on:click={() => chartType = 'uptime'}
					>
						Uptime
					</button>
					<button
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {chartType === 'response' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
						on:click={() => chartType = 'response'}
					>
						Response Time
					</button>
				</div>
			</div>

			{#if chartType === 'uptime'}
				<!-- Uptime History Chart -->
				{#if checks && checks.length > 0}
					<div class="flex items-end gap-1 h-24">
						{#each checks.slice(0, 48).reverse() as check}
							<div
								class="flex-1 rounded-sm transition-all hover:opacity-80"
								class:bg-emerald-500={check.success}
								class:bg-rose-500={!check.success}
								style="height: {check.success ? '100%' : '20%'}"
								title="{formatDate(check.checked_at)} - {check.success ? 'Up' : 'Down'}"
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-xs text-slate-400 mt-3 font-medium">
						<span>24h ago</span>
						<span>Now</span>
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center h-32 text-slate-400">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
						</svg>
						<p>No check history available</p>
					</div>
				{/if}
			{:else}
				<!-- Response Time Chart -->
				<div class="mb-6 flex justify-end">
					<div class="flex bg-slate-100 rounded-lg p-1 w-fit">
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '1h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => responseTimePeriod = '1h'}
						>
							1h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '6h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => responseTimePeriod = '6h'}
						>
							6h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '12h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => responseTimePeriod = '12h'}
						>
							12h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '24h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => responseTimePeriod = '24h'}
						>
							24h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '1w' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => responseTimePeriod = '1w'}
						>
							1w
						</button>
					</div>
				</div>
				{#if chartsLoaded && LineChart && getResponseTimeData() && getResponseTimeData().length > 0}
					<div class="h-64">
						<svelte:component 
							this={LineChart}
							data={getResponseTimeData()} 
							label="Response Time" 
							color="#3B82F6"
							yAxisLabel="Response Time (ms)"
						/>
					</div>
				{:else if checks && checks.length > 0}
					{@const responseTimes = getFilteredResponseTimes()}
					{@const maxTime = Math.max(...responseTimes, 1)}
					<div class="flex items-end gap-1 h-24">
						{#each responseTimes as time}
							<div
								class="flex-1 bg-blue-500 rounded-sm transition-all hover:opacity-80"
								style="height: {(time / maxTime) * 100}%"
								title="{time}ms"
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-xs text-slate-400 mt-3 font-medium">
						<span>{getPeriodLabel()} ago</span>
						<span>Max: {maxTime}ms</span>
						<span>Now</span>
					</div>
				{:else}
					<div class="flex flex-col items-center justify-center h-32 text-slate-400">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
						<p>No response time data available</p>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Uptime Chart -->
		<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-lg font-bold text-slate-900">Uptime Statistics</h2>
				<div class="flex bg-slate-100 rounded-lg p-1">
					<button
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {uptimePeriod === '7d' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
						on:click={() => uptimePeriod = '7d'}
					>
						Last 7 Days
					</button>
					<button
						class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {uptimePeriod === '30d' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
						on:click={() => uptimePeriod = '30d'}
					>
						Last 30 Days
					</button>
				</div>
			</div>
			{#if chartsLoaded && DonutChart && ((uptimePeriod === '7d' && metrics7d && metrics7d.uptime) || (uptimePeriod === '30d' && metrics30d && metrics30d.uptime))}
				<div class="h-64">
					<svelte:component 
						this={DonutChart}
						percentage={uptimePeriod === '7d' ? metrics7d.uptime.percentage : metrics30d.uptime.percentage} 
						label="Uptime"
					/>
				</div>
				<div class="mt-6 text-center">
					<p class="text-sm font-medium text-slate-600">
						{uptimePeriod === '7d' 
							? `${metrics7d.uptime.success_checks} successful / ${metrics7d.uptime.total_checks} total checks`
							: `${metrics30d.uptime.success_checks} successful / ${metrics30d.uptime.total_checks} total checks`
						}
					</p>
				</div>
			{:else if !chartsLoaded && ((uptimePeriod === '7d' && metrics7d && metrics7d.uptime) || (uptimePeriod === '30d' && metrics30d && metrics30d.uptime))}
				<div class="h-64 flex items-center justify-center">
					<div class="text-center">
						<div class="text-5xl font-bold text-slate-900 mb-2">
							{uptimePeriod === '7d' ? metrics7d.uptime.percentage.toFixed(1) : metrics30d.uptime.percentage.toFixed(1)}%
						</div>
						<div class="text-sm font-medium text-slate-500 uppercase tracking-wide">Uptime</div>
					</div>
				</div>
				<div class="mt-6 text-center">
					<p class="text-sm font-medium text-slate-600">
						{uptimePeriod === '7d' 
							? `${metrics7d.uptime.success_checks} successful / ${metrics7d.uptime.total_checks} total checks`
							: `${metrics30d.uptime.success_checks} successful / ${metrics30d.uptime.total_checks} total checks`
						}
					</p>
				</div>
			{:else}
				<div class="flex flex-col items-center justify-center h-32 text-slate-400">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<p>No uptime data available</p>
				</div>
			{/if}
		</div>

		{#if checks && checks.length > 0}
			<div class="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
				<div class="px-6 py-4 border-b border-slate-200">
					<h2 class="text-lg font-bold text-slate-900">Recent Checks</h2>
				</div>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-slate-200">
						<thead class="bg-slate-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Status
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Status Code
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Response Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Error
								</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-slate-200">
							{#each checks.slice(0, 10) as check}
								<tr class="hover:bg-slate-50 transition-colors">
									<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-900">
										{formatDate(check.checked_at)}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<MonitorStatus status={check.success ? 'up' : 'down'} showText={true} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600 font-mono">
										{extractInt64(check.status_code, 'N/A')}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-slate-600 font-mono">
										{#if isValidSqlNull(check.response_time_ms)}
											{extractInt64(check.response_time_ms, 0)}ms
										{:else}
											N/A
										{/if}
									</td>
									<td class="px-6 py-4 text-sm text-rose-600 max-w-xs truncate">
										{extractString(check.error_message, '-')}
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
