<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import LineChart from '$lib/components/charts/LineChart.svelte';
	import DonutChart from '$lib/components/charts/DonutChart.svelte';
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

	// Auto-refresh settings
	let autoRefreshInterval = 300; // seconds (5 minutes)
	let autoRefreshTimer: ReturnType<typeof setInterval> | null = null;

	$: monitorId = $page.params.id || '';

	onMount(() => {
		loadMonitorData();
		startAutoRefresh();
	});

	onDestroy(() => {
		stopAutoRefresh();
	});

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

<div class="max-w-7xl mx-auto">
	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
			{error}
		</div>
		<button on:click={handleBack} class="text-blue-600 hover:text-blue-800 font-medium">
			← Back to Monitors
		</button>
	{:else if monitor}
		<div class="mb-4">
			<button
				on:click={handleBack}
				class="text-blue-600 hover:text-blue-800 font-medium mb-4 inline-flex items-center"
			>
				← Back to Monitors
			</button>
			<div class="flex justify-between items-start">
				<div>
					<h1 class="text-2xl font-bold text-gray-900 mb-2">{monitor.name}</h1>
					<p class="text-gray-600">{monitor.url}</p>
				</div>
				<div class="flex items-center gap-2">
					<label for="refresh-interval" class="text-sm text-gray-600">Auto-refresh:</label>
					<select 
						id="refresh-interval" 
						bind:value={autoRefreshInterval} 
						on:change={handleIntervalChange}
						class="px-3 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
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

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
			<!-- Status Overview -->
			<div class="bg-white rounded-lg shadow-md p-4">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Status</h3>
				<MonitorStatus status={getMonitorStatus()} />
			</div>
			<div class="bg-white rounded-lg shadow-md p-4">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Uptime (24h)</h3>
				<p class="text-2xl font-bold text-gray-900">{calculateUptime()}</p>
			</div>
			<div class="bg-white rounded-lg shadow-md p-4">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Avg Response Time</h3>
				<p class="text-2xl font-bold text-gray-900">{getAverageResponseTime()}</p>
			</div>
			
			<!-- SSL Information (conditional) -->
			{#if monitor.check_ssl && sslStatus}
				<div class="bg-white rounded-lg shadow-md p-4">
					<h3 class="text-sm font-medium text-gray-500 mb-2">SSL Status</h3>
					<div class="space-y-1">
						<p class="text-lg font-semibold {sslStatus.valid ? 'text-green-600' : 'text-red-600'}">
							{sslStatus.valid ? 'Valid' : 'Invalid'}
						</p>
						{#if sslStatus.expires_at}
							<p class="text-sm text-gray-600">
								Expires: {new Date(sslStatus.expires_at).toLocaleDateString()}
							</p>
						{/if}
					</div>
				</div>
			{:else if monitor.check_ssl}
				<div class="bg-white rounded-lg shadow-md p-4">
					<h3 class="text-sm font-medium text-gray-500 mb-2">SSL Status</h3>
					<p class="text-lg font-semibold text-gray-500">Checking...</p>
				</div>
			{/if}
		</div>

		<!-- History Charts -->
		<div class="bg-white rounded-lg shadow-md p-4 mb-4">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-bold text-gray-900">History</h2>
				<div class="flex bg-gray-100 rounded-lg p-1">
					<button
						class="px-3 py-1 text-sm rounded-md transition-colors {chartType === 'uptime' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						on:click={() => chartType = 'uptime'}
					>
						Uptime
					</button>
					<button
						class="px-3 py-1 text-sm rounded-md transition-colors {chartType === 'response' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						on:click={() => chartType = 'response'}
					>
						Response Time
					</button>
				</div>
			</div>

			{#if chartType === 'uptime'}
				<!-- Uptime History Chart -->
				{#if checks && checks.length > 0}
					<div class="flex items-end gap-1 h-16">
						{#each checks.slice(0, 48).reverse() as check}
							<div
								class="flex-1 rounded-t transition-all hover:opacity-75"
								class:bg-green-500={check.success}
								class:bg-red-500={!check.success}
								style="height: {check.success ? '100%' : '20%'}"
								title="{formatDate(check.checked_at)} - {check.success ? 'Up' : 'Down'}"
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-xs text-gray-500 mt-2">
						<span>24h ago</span>
						<span>Now</span>
					</div>
				{:else}
					<p class="text-gray-500">No check history available</p>
				{/if}
			{:else}
				<!-- Response Time Chart -->
				<div class="mb-4">
					<div class="flex bg-gray-50 rounded-lg p-1 w-fit">
						<button
							class="px-2 py-1 text-xs rounded-md transition-colors {responseTimePeriod === '1h' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
							on:click={() => responseTimePeriod = '1h'}
						>
							1h
						</button>
						<button
							class="px-2 py-1 text-xs rounded-md transition-colors {responseTimePeriod === '6h' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
							on:click={() => responseTimePeriod = '6h'}
						>
							6h
						</button>
						<button
							class="px-2 py-1 text-xs rounded-md transition-colors {responseTimePeriod === '12h' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
							on:click={() => responseTimePeriod = '12h'}
						>
							12h
						</button>
						<button
							class="px-2 py-1 text-xs rounded-md transition-colors {responseTimePeriod === '24h' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
							on:click={() => responseTimePeriod = '24h'}
						>
							24h
						</button>
						<button
							class="px-2 py-1 text-xs rounded-md transition-colors {responseTimePeriod === '1w' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
							on:click={() => responseTimePeriod = '1w'}
						>
							1w
						</button>
					</div>
				</div>
				{#if getResponseTimeData() && getResponseTimeData().length > 0}
					<div class="h-48">
						<LineChart 
							data={getResponseTimeData()} 
							label="Response Time" 
							color="#3B82F6"
							yAxisLabel="Response Time (ms)"
						/>
					</div>
				{:else if checks && checks.length > 0}
					{@const responseTimes = getFilteredResponseTimes()}
					{@const maxTime = Math.max(...responseTimes, 1)}
					<div class="flex items-end gap-1 h-16">
						{#each responseTimes as time}
							<div
								class="flex-1 bg-blue-500 rounded-t transition-all hover:opacity-75"
								style="height: {(time / maxTime) * 100}%"
								title="{time}ms"
							></div>
						{/each}
					</div>
					<div class="flex justify-between text-xs text-gray-500 mt-2">
						<span>{getPeriodLabel()} ago</span>
						<span>Max: {maxTime}ms</span>
						<span>Now</span>
					</div>
				{:else}
					<p class="text-gray-500">No response time data available</p>
				{/if}
			{/if}
		</div>

		<!-- Uptime Chart -->
		<div class="bg-white rounded-lg shadow-md p-4 mb-4">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-bold text-gray-900">Uptime</h2>
				<div class="flex bg-gray-100 rounded-lg p-1">
					<button
						class="px-3 py-1 text-sm rounded-md transition-colors {uptimePeriod === '7d' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						on:click={() => uptimePeriod = '7d'}
					>
						Last 7 Days
					</button>
					<button
						class="px-3 py-1 text-sm rounded-md transition-colors {uptimePeriod === '30d' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
						on:click={() => uptimePeriod = '30d'}
					>
						Last 30 Days
					</button>
				</div>
			</div>
			{#if (uptimePeriod === '7d' && metrics7d && metrics7d.uptime) || (uptimePeriod === '30d' && metrics30d && metrics30d.uptime)}
				<div class="h-48">
					<DonutChart 
						percentage={uptimePeriod === '7d' ? metrics7d.uptime.percentage : metrics30d.uptime.percentage} 
						label="Uptime"
					/>
				</div>
				<div class="mt-4 text-center text-sm text-gray-600">
					<p>
						{uptimePeriod === '7d' 
							? `${metrics7d.uptime.success_checks} successful / ${metrics7d.uptime.total_checks} total checks`
							: `${metrics30d.uptime.success_checks} successful / ${metrics30d.uptime.total_checks} total checks`
						}
					</p>
				</div>
			{:else}
				<p class="text-gray-500">No uptime data available</p>
			{/if}
		</div>

		{#if checks && checks.length > 0}
			<div class="bg-white rounded-lg shadow-md p-4">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Recent Checks</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Time
								</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Status
								</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Status Code
								</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Response Time
								</th>
								<th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Error
								</th>
							</tr>
						</thead>
						<tbody class="bg-white divide-y divide-gray-200">
							{#each checks.slice(0, 5) as check}
								<tr>
									<td class="px-4 py-2 whitespace-nowrap text-sm text-gray-900">
										{formatDate(check.checked_at)}
									</td>
									<td class="px-4 py-2 whitespace-nowrap">
										<MonitorStatus status={check.success ? 'up' : 'down'} showText={true} />
									</td>
									<td class="px-4 py-2 whitespace-nowrap text-sm text-gray-900">
										{extractInt64(check.status_code, 'N/A')}
									</td>
									<td class="px-4 py-2 whitespace-nowrap text-sm text-gray-900">
										{#if isValidSqlNull(check.response_time_ms)}
											{extractInt64(check.response_time_ms, 0)}ms
										{:else}
											N/A
										{/if}
									</td>
									<td class="px-4 py-2 text-sm text-red-600">
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
