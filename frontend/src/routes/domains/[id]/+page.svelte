<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';
	import LineChart from '$lib/components/charts/LineChart.svelte';
	import DonutChart from '$lib/components/charts/DonutChart.svelte';
	import BarChart from '$lib/components/charts/BarChart.svelte';
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

	// Auto-refresh settings
	let autoRefreshInterval = 60; // seconds
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

		<div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Status</h3>
				<MonitorStatus status={getMonitorStatus()} />
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<h3 class="text-sm font-medium text-gray-500 mb-2">Uptime (24h)</h3>
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

		<div class="bg-white rounded-lg shadow-md p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Uptime History (Last 24 Hours)</h2>
			{#if checks && checks.length > 0}
				<div class="flex items-end gap-1 h-48">
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
		</div>

		<div class="bg-white rounded-lg shadow-md p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Response Time (Last 24 Hours)</h2>
			{#if metrics24h && metrics24h.response_time_history && metrics24h.response_time_history.length > 0}
				<div class="h-64">
					<LineChart 
						data={metrics24h.response_time_history} 
						label="Response Time" 
						color="#3B82F6"
						yAxisLabel="Response Time (ms)"
					/>
				</div>
			{:else if checks && checks.length > 0}
				{@const responseTimes = checks
					.filter((c) => isValidSqlNull(c.response_time_ms))
					.slice(0, 48)
					.reverse()
					.map((c) => extractInt64(c.response_time_ms, 0))}
				{@const maxTime = Math.max(...responseTimes, 1)}
				<div class="flex items-end gap-1 h-48">
					{#each responseTimes as time}
						<div
							class="flex-1 bg-blue-500 rounded-t transition-all hover:opacity-75"
							style="height: {(time / maxTime) * 100}%"
							title="{time}ms"
						></div>
					{/each}
				</div>
				<div class="flex justify-between text-xs text-gray-500 mt-2">
					<span>24h ago</span>
					<span>Max: {maxTime}ms</span>
					<span>Now</span>
				</div>
			{:else}
				<p class="text-gray-500">No response time data available</p>
			{/if}
		</div>

		<!-- Uptime Charts -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Uptime - Last 7 Days</h2>
				{#if metrics7d && metrics7d.uptime}
					<div class="h-64">
						<DonutChart 
							percentage={metrics7d.uptime.percentage} 
							label="Uptime"
						/>
					</div>
					<div class="mt-4 text-center text-sm text-gray-600">
						<p>{metrics7d.uptime.success_checks} successful / {metrics7d.uptime.total_checks} total checks</p>
					</div>
				{:else}
					<p class="text-gray-500">No uptime data available</p>
				{/if}
			</div>
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Uptime - Last 30 Days</h2>
				{#if metrics30d && metrics30d.uptime}
					<div class="h-64">
						<DonutChart 
							percentage={metrics30d.uptime.percentage} 
							label="Uptime"
						/>
					</div>
					<div class="mt-4 text-center text-sm text-gray-600">
						<p>{metrics30d.uptime.success_checks} successful / {metrics30d.uptime.total_checks} total checks</p>
					</div>
				{:else}
					<p class="text-gray-500">No uptime data available</p>
				{/if}
			</div>
		</div>

		<!-- Status Code Distribution -->
		<div class="bg-white rounded-lg shadow-md p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Status Code Distribution (Last 24 Hours)</h2>
			{#if metrics24h && metrics24h.status_code_distribution && metrics24h.status_code_distribution.length > 0}
				<div class="h-64">
					<BarChart data={metrics24h.status_code_distribution} />
				</div>
			{:else}
				<p class="text-gray-500">No status code data available</p>
			{/if}
		</div>

		{#if monitor.check_ssl && monitor.url.startsWith('https') && sslStatus}
			<div class="bg-white rounded-lg shadow-md p-6 mb-8">
				<h2 class="text-xl font-bold text-gray-900 mb-4">SSL Certificate</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<p class="text-sm text-gray-500">Valid</p>
						<p class="font-medium text-gray-900">
							{sslStatus.valid ? 'Yes' : 'No'}
						</p>
					</div>
					{#if sslStatus.expires_at}
						<div>
							<p class="text-sm text-gray-500">Expires At</p>
							<p class="font-medium text-gray-900">{formatDate(sslStatus.expires_at)}</p>
						</div>
					{/if}
					{#if sslStatus.issuer}
						<div>
							<p class="text-sm text-gray-500">Issuer</p>
							<p class="font-medium text-gray-900">{sslStatus.issuer}</p>
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

		{#if checks && checks.length > 0}
			<div class="bg-white rounded-lg shadow-md p-6">
				<h2 class="text-xl font-bold text-gray-900 mb-4">Recent Checks</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200">
						<thead class="bg-gray-50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Status
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Status Code
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Response Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
									Error
								</th>
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
										{extractInt64(check.status_code, 'N/A')}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
										{#if isValidSqlNull(check.response_time_ms)}
											{extractInt64(check.response_time_ms, 0)}ms
										{:else}
											N/A
										{/if}
									</td>
									<td class="px-6 py-4 text-sm text-red-600">
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
