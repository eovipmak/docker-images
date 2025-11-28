<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks, connectEventStream, disconnectEventStream, type MonitorCheckEvent } from '$lib/api/events';
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

	// Get a cutoff Date for the provided period
	function getCutoffForPeriod(period = responseTimePeriod) {
		const now = new Date();
		const cutoff = new Date(now);
		switch (period) {
			case '1h':
				cutoff.setTime(now.getTime() - 1 * 60 * 60 * 1000);
				break;
			case '6h':
				cutoff.setTime(now.getTime() - 6 * 60 * 60 * 1000);
				break;
			case '12h':
				cutoff.setTime(now.getTime() - 12 * 60 * 60 * 1000);
				break;
			case '24h':
				cutoff.setTime(now.getTime() - 24 * 60 * 60 * 1000);
				break;
			case '1w':
				cutoff.setTime(now.getTime() - 7 * 24 * 60 * 60 * 1000);
				break;
		}
		return cutoff;
	}

	// Get filtered response times based on period
	function getFilteredResponseTimes(period = responseTimePeriod) {
		if (!checks || checks.length === 0) return [];

		const cutoff = getCutoffForPeriod(period);

		// debug removed

		const filtered = checks
			.filter((c) => {
				const checkTime = new Date(c.checked_at);
				const valid = checkTime > cutoff && isValidSqlNull(c.response_time_ms);
				// removed debug logging
				return valid;
			})
			.reverse() // Since checks are newest first, reverse to oldest first
			.map((c) => extractInt64(c.response_time_ms, 0));

		// filtered response times length: ${filtered.length}

		return filtered;
	}

	// Force reactivity on checks by passing checks.length as a dependency signal
	$: responseTimes = (checks?.length ?? -1) >= 0 ? getFilteredResponseTimes(responseTimePeriod) : [];
	$: maxTime = responseTimes.length > 0 ? Math.max(...responseTimes) : 1;
	// Compute a conservative suggested max (95th percentile) and add padding to avoid outlier distortion
	$: p95Time = (() => {
		if (!responseTimes || responseTimes.length === 0) return maxTime;
		const sorted = [...responseTimes].sort((a,b) => a - b);
		const idx = Math.floor(sorted.length * 0.95);
		return sorted[Math.min(idx, sorted.length-1)];
	})();

	// Reactive current timeRange for the selected period
	$: currentTimeRange = getTimeRangeForPeriod(responseTimePeriod);

	// Lazy loaded chart components
	let LineChart: any = null;
	let DonutChart: any = null;
	let chartsLoaded = false;

	// SSE subscription
	let unsubscribeChecks: (() => void) | null = null;

	$: monitorId = $page.params.id || '';

	// Computed uptime data based on selected period
	$: currentUptimeData = uptimePeriod === '7d'
		? (metrics7d?.uptime ?? { percentage: 0, total_checks: 0, success_checks: 0, failed_checks: 0 })
		: (metrics30d?.uptime ?? { percentage: 0, total_checks: 0, success_checks: 0, failed_checks: 0 });

	// Force reactivity on checks and responseTimePeriod by using them in the expression
	$: responseTimeSeries = ((checks?.length ?? -1) >= 0 && responseTimePeriod) ? getResponseTimeData() : [];
	$: hasResponseTimeSeries = responseTimeSeries && responseTimeSeries.filter(d => d.value !== null).length > 0;
	$: bucketTotal = responseTimeSeries ? responseTimeSeries.length : 0;
	$: bucketNonEmpty = responseTimeSeries ? responseTimeSeries.filter(d => d.value !== null).length : 0;

	onMount(async () => {
		loadMonitorData();
		// Lazy load chart components
		loadCharts();
		
		// Start SSE connection for real-time updates
		await connectEventStream();
		
		// Subscribe to monitor check events for this specific monitor
		unsubscribeChecks = latestMonitorChecks.subscribe((latestChecks) => {
			const check = latestChecks.get(monitorId);
			if (check && !isLoading) {
				addCheckFromSSE(check);
			}
		});
	});

	onDestroy(() => {
		// Disconnect SSE when leaving monitor details
		disconnectEventStream();
		
		if (unsubscribeChecks) {
			unsubscribeChecks();
		}
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

	// Build a timeRange object to pass to LineChart so x axis covers data with appropriate padding
	function getTimeRangeForPeriod(period = responseTimePeriod) {
		const now = new Date();
		// Decide time unit for x axis ticks
		let unit: 'minute' | 'hour' | 'day' = 'hour';
		switch (period) {
			case '1h':
				unit = 'minute';
				break;
			case '6h':
			case '12h':
			case '24h':
				unit = 'hour';
				break;
			case '1w':
				unit = 'day';
				break;
		}
		
		// For time range, we want to fit the data with some padding
		// Don't force min to period cutoff - let it auto-fit based on actual data
		// This prevents large empty spaces when data only covers a portion of the period
		return { 
			max: now.toISOString(), 
			unit 
		};
	}

	// Bucket checks into fixed intervals between start and end
	function bucketChecks(checksArr: any[], startTime: Date, endTime: Date, intervalSeconds: number) {
		const buckets: { timestamp: string; values: number[] }[] = [];
		// Align to interval boundaries
		let current = new Date(Math.floor(startTime.getTime() / (intervalSeconds * 1000)) * intervalSeconds * 1000);
		while (current < endTime) {
			buckets.push({ timestamp: new Date(current).toISOString(), values: [] });
			current = new Date(current.getTime() + intervalSeconds * 1000);
		}

		// Assign checks to buckets
		for (const c of checksArr) {
			if (!isValidSqlNull(c.response_time_ms)) continue;
			const t = new Date(c.checked_at).getTime();
			if (t < startTime.getTime() || t > endTime.getTime()) continue;
			const bucketIndex = Math.floor((t - startTime.getTime()) / (intervalSeconds * 1000));
			if (bucketIndex >= 0 && bucketIndex < buckets.length) {
				buckets[bucketIndex].values.push(extractInt64(c.response_time_ms, 0));
			}
		}

		// Compute average per bucket (return null for empty buckets to keep timeline continuity)
		const result: { timestamp: string; value: number | null }[] = buckets.map((b) => ({
			timestamp: b.timestamp,
			value: b.values.length > 0 ? b.values.reduce((s, v) => s + v, 0) / b.values.length : null
		}));

		// Trim leading empty buckets so chart focuses on actual data
		// This prevents large empty spaces at the start of the chart
		const firstNonEmptyIndex = result.findIndex(b => b.value !== null);
		if (firstNonEmptyIndex > 0) {
			return result.slice(firstNonEmptyIndex);
		}

		return result;
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

			const checksResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/checks?limit=1000`);
			if (checksResponse.ok) {
				const checksData = await checksResponse.json();
				checks = checksData || [];
			} else {
				checks = [];
			}

			// Load metrics for different periods
			const [metrics24hRes, metrics7dRes, metrics30dRes] = await Promise.all([
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=24h`),
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=7d`),
				fetchAPI(`/api/v1/monitors/${monitorId}/metrics?period=30d`)
			]);

			// Helper to return a minimal uptime payload when none is available
			const defaultUptime = () => ({
				uptime: { percentage: 0, total_checks: 0, success_checks: 0, failed_checks: 0 },
				response_time_history: [],
				status_code_distribution: []
			});

			if (metrics24hRes.ok) {
				try {
					const parsed = await metrics24hRes.json();
					metrics24h = { ...(parsed || {}), uptime: parsed?.uptime || defaultUptime().uptime };
				} catch (err) {
					console.error('Failed to parse metrics24h JSON', err);
					metrics24h = defaultUptime();
				}
			} else {
				metrics24h = defaultUptime();
			}

			if (metrics7dRes.ok) {
				try {
					const parsed = await metrics7dRes.json();
					metrics7d = { ...(parsed || {}), uptime: parsed?.uptime || defaultUptime().uptime };
				} catch (err) {
					console.error('Failed to parse metrics7d JSON', err);
					metrics7d = defaultUptime();
				}
			} else {
				metrics7d = defaultUptime();
			}

			if (metrics30dRes.ok) {
				try {
					const parsed = await metrics30dRes.json();
					metrics30d = { ...(parsed || {}), uptime: parsed?.uptime || defaultUptime().uptime };
				} catch (err) {
					console.error('Failed to parse metrics30d JSON', err);
					metrics30d = defaultUptime();
				}
			} else {
				metrics30d = defaultUptime();
			}

			// Debug logging for metrics payloads
			console.debug('Loaded metrics:', {
				metrics24h: metrics24h?.uptime ?? metrics24h,
				metrics7d: metrics7d?.uptime ?? metrics7d,
				metrics30d: metrics30d?.uptime ?? metrics30d
			});

			// Loaded metrics sizes: metrics24h: ${metrics24h?.response_time_history?.length || 0} metrics7d: ${metrics7d?.response_time_history?.length || 0} metrics30d: ${metrics30d?.response_time_history?.length || 0}

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
		goto('/monitors');
	}

	function getResponseTimeSeriesFromTimes(times: number[], period: '1h' | '6h' | '12h' | '24h' | '1w') {
		if (times.length === 0) return [];
		const now = new Date();
		const cutoff = getCutoffForPeriod(period);
		const interval = (now.getTime() - cutoff.getTime()) / times.length;
		return times.map((time, i) => ({
			timestamp: new Date(cutoff.getTime() + i * interval).toISOString(),
			value: time
		}));
	}

	// Get response time data based on selected period
	// Build Chart.js-compatible data for the selected period
	function getResponseTimeData() {
		if (!checks || !Array.isArray(checks) || checks.length === 0) {
			return [];
		}
		const maxPointsRenderable = 500;
		switch (responseTimePeriod) {
			case '1h': {
				// Return full raw points for 1h (no downsampling), including checks without response time as null
				const start = getCutoffForPeriod('1h');
				const end = new Date();
				const rawData = checks
					.filter(c => {
						const t = new Date(c.checked_at);
						return t >= start && t <= end;
					})
					.sort((a, b) => new Date(a.checked_at).getTime() - new Date(b.checked_at).getTime())
					.map(c => ({
						timestamp: c.checked_at,
						value: isValidSqlNull(c.response_time_ms) ? extractInt64(c.response_time_ms, 0) : null
					}));
				return rawData;
			}
			case '6h': {
				// For 6h, use a larger window size for smoother chart
				const start = getCutoffForPeriod('6h');
				const end = new Date();
				const totalSeconds = (end.getTime() - start.getTime()) / 1000;
				const windowSizeSeconds = Math.max(120, Math.ceil(totalSeconds / 200)); // Aim for ~200 points for smoother display
				const buckets = bucketChecks(checks, start, end, windowSizeSeconds);
				return buckets;
			}
			case '12h': {
				// For 12h, use a larger window size for smoother chart
				const start = getCutoffForPeriod('12h');
				const end = new Date();
				const totalSeconds = (end.getTime() - start.getTime()) / 1000;
				const windowSizeSeconds = Math.max(180, Math.ceil(totalSeconds / 200)); // Aim for ~200 points for smoother display
				const buckets = bucketChecks(checks, start, end, windowSizeSeconds);
				return buckets;
			}
			case '24h':
			case '1w': {
				// Apply aggregation for larger time frames
				const start = getCutoffForPeriod(responseTimePeriod);
				const end = new Date();
				const totalSeconds = (end.getTime() - start.getTime()) / 1000;
				const windowSizeSeconds = Math.max(60, Math.ceil(totalSeconds / maxPointsRenderable));
				const buckets = bucketChecks(checks, start, end, windowSizeSeconds);
				return buckets;
			}
			default:
				return [];
		}
	}

	// Get period label for display
	function addCheckFromSSE(check: MonitorCheckEvent) {
		// Convert SSE check to API format
		const newCheck = {
			id: `sse-${Date.now()}`, // Generate temporary ID
			monitor_id: check.monitor_id,
			checked_at: check.checked_at,
			status_code: check.status_code !== undefined ? { Int64: check.status_code, Valid: true } : null,
			response_time_ms: check.response_time_ms !== undefined ? { Int64: check.response_time_ms, Valid: true } : null,
			ssl_valid: check.ssl_valid !== undefined ? { Bool: check.ssl_valid, Valid: true } : null,
			ssl_expires_at: check.ssl_expires_at ? { Time: check.ssl_expires_at, Valid: true } : null,
			error_message: check.error_message ? { String: check.error_message, Valid: true } : null,
			success: check.success
		};

		// Add to beginning of checks array, keep only last 100
		checks = [newCheck, ...checks.slice(0, 99)];

		// Update SSL status if SSL data is available
		if (check.ssl_valid !== undefined && monitor?.check_ssl) {
			sslStatus = {
				valid: check.ssl_valid,
				expires_at: check.ssl_expires_at || null
			};
		}

		// Trigger reactivity
		checks = [...checks];
	}

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
						<h1 class="text-2xl font-bold text-slate-900 dark:text-white drop-shadow-sm">{monitor.name}</h1>
						<MonitorStatus status={getMonitorStatus()} />
					</div>
					<a href={monitor.url} target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:underline text-sm mt-1 flex items-center gap-1">
						{monitor.url}
						<svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
						</svg>
					</a>
				</div>
			</div>
			
			<div class="flex items-center gap-3">
				<!-- Real-time updates via SSE -->
			</div>
		</div>

		<!-- Stats Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Status Overview -->
			   <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5">
				<div class="flex items-center justify-between mb-2">
					   <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400">Current Status</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					   <span class="text-2xl font-bold text-slate-900 dark:text-slate-100 capitalize">{getMonitorStatus()}</span>
				</div>
			</div>

			<!-- Uptime -->
			   <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5">
				<div class="flex items-center justify-between mb-2">
					   <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400">Uptime (24h)</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					   <span class="text-2xl font-bold text-slate-900 dark:text-slate-100">{calculateUptime()}</span>
				</div>
			</div>

			<!-- Response Time -->
			   <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5">
				<div class="flex items-center justify-between mb-2">
					   <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400">Avg Response</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
					</div>
				</div>
				<div class="flex items-baseline gap-2">
					   <span class="text-2xl font-bold text-slate-900 dark:text-slate-100">{getAverageResponseTime()}</span>
				</div>
			</div>
			
			<!-- SSL Information -->
			   <div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5">
				<div class="flex items-center justify-between mb-2">
					   <h3 class="text-sm font-medium text-slate-500 dark:text-slate-400">SSL Status</h3>
					<div class="p-2 bg-slate-50 rounded-lg">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					</div>
				</div>
				{#if monitor.check_ssl && sslStatus}
					<div class="flex flex-col">
						   <span class="text-2xl font-bold {sslStatus.valid ? 'text-emerald-600' : 'text-rose-600'} dark:text-slate-100">
							{sslStatus.valid ? 'Valid' : 'Invalid'}
						</span>
						{#if sslStatus.expires_at}
							   <span class="text-xs text-slate-500 dark:text-slate-400 mt-1">
								Expires: {new Date(sslStatus.expires_at).toLocaleDateString()}
							</span>
						{/if}
					</div>
				{:else if monitor.check_ssl}
					   <span class="text-lg font-medium text-slate-500 dark:text-slate-400">Checking...</span>
				{:else}
					   <span class="text-lg font-medium text-slate-400 dark:text-slate-500">Disabled</span>
				{/if}
			</div>
		</div>

		<!-- History Charts -->
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6">
			<div class="flex justify-between items-center mb-6">
				   <h2 class="text-lg font-bold text-slate-900 dark:text-slate-100">History</h2>
				   <div class="flex bg-slate-100 dark:bg-slate-700 rounded-lg p-1">
					<button
						   class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {chartType === 'uptime' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
						on:click={() => chartType = 'uptime'}
					>
						Uptime
					</button>
					<button
						   class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {chartType === 'response' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
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
					<div class="flex justify-between text-xs text-slate-400 dark:text-blue-300 mt-3 font-medium">
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
							on:click={() => { responseTimePeriod = '1h'; }}
						>
							1h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '6h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => { responseTimePeriod = '6h'; }}
						>
							6h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '12h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => { responseTimePeriod = '12h'; }}
						>
							12h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '24h' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => { responseTimePeriod = '24h'; }}
						>
							24h
						</button>
						<button
							class="px-3 py-1 text-xs font-medium rounded-md transition-all {responseTimePeriod === '1w' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'}"
							on:click={() => { responseTimePeriod = '1w'; }}
						>
							1w
						</button>
					</div>
				</div>
				{#if chartsLoaded && LineChart && responseTimeSeries && hasResponseTimeSeries}
                    
					{#key responseTimePeriod}
					<div class="h-64">
						<svelte:component 
							this={LineChart}
							data={responseTimeSeries} 
							label="Response Time" 
							color="#3B82F6"
							yAxisLabel="Response Time (ms)"
							timeRange={currentTimeRange}
							timeUnit={currentTimeRange?.unit}
							suggestedMax={Math.ceil(Math.max(p95Time * 1.25, maxTime * 1.05))}
							spanGapsProp={false}
						/>
					</div>
					{/key}
				{:else}
					<div class="flex flex-col items-center justify-center h-32 text-slate-400">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
						</svg>
						<p>No response time data available for the selected period. Try selecting a longer time range like 24h or 1w.</p>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Uptime Chart -->
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6">
			<div class="flex justify-between items-center mb-6">
				   <h2 class="text-lg font-bold text-slate-900 dark:text-slate-100">Uptime Statistics</h2>
				   <div class="flex bg-slate-100 dark:bg-slate-700 rounded-lg p-1">
					<button
						   class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {uptimePeriod === '7d' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
						on:click={() => uptimePeriod = '7d'}
					>
						Last 7 Days
					</button>
					<button
						   class="px-3 py-1.5 text-sm font-medium rounded-md transition-all {uptimePeriod === '30d' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 shadow-sm' : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'}"
						on:click={() => uptimePeriod = '30d'}
					>
						Last 30 Days
					</button>
				</div>
			</div>
			{#if chartsLoaded && DonutChart && currentUptimeData && currentUptimeData.total_checks > 0}
				<div class="h-64">
					<svelte:component 
						this={DonutChart}
						percentage={currentUptimeData.percentage} 
						label="Uptime"
						key={uptimePeriod}
					/>
				</div>
				<div class="mt-6 text-center">
					<p class="text-sm font-medium text-slate-600">
						{currentUptimeData.success_checks} successful / {currentUptimeData.total_checks} total checks
					</p>
				</div>
			{:else if !chartsLoaded && currentUptimeData && currentUptimeData.total_checks > 0}
				<div class="h-64 flex items-center justify-center">
					<div class="text-center">
						<div class="text-5xl font-extrabold text-slate-900 dark:text-blue-400 mb-2 drop-shadow">
							{currentUptimeData.percentage.toFixed(1)}%
						</div>
						<div class="text-sm font-semibold text-slate-500 dark:text-blue-300 uppercase tracking-wide">Uptime</div>
					</div>
				</div>
				<div class="mt-6 text-center">
					<p class="text-sm font-semibold text-slate-600 dark:text-blue-300 drop-shadow">
						{currentUptimeData.success_checks} successful / {currentUptimeData.total_checks} total checks
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
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden">
				   <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between">
					   <h2 class="text-base font-semibold text-slate-900 dark:text-slate-100">Recent Checks</h2>
					   <div class="text-xs text-slate-500 dark:text-slate-400">
						Last {Math.min(5, checks.length)} checks
					</div>
				</div>
				   <div class="overflow-x-auto">
					   <table class="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
						   <thead class="bg-slate-50 dark:bg-slate-900">
							<tr>
								   <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
									Time
								</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Status
								</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Code
								</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Response
								</th>
								<th class="px-4 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
									Error
								</th>
							</tr>
						</thead>
						   <tbody class="bg-white dark:bg-slate-800">
							{#each checks.slice(0, 5) as check, index}
								   <tr class="transition-colors {index % 2 === 0 ? 'bg-white dark:bg-slate-800' : 'bg-slate-50/50 dark:bg-slate-900/50'} hover:bg-slate-100 dark:hover:bg-slate-700">
									   <td class="px-4 py-3 whitespace-nowrap text-sm text-slate-900 dark:text-slate-100">
										<div class="flex items-center gap-2">
											<div class="w-2 h-2 rounded-full {check.success ? 'bg-emerald-500' : 'bg-rose-500'}"></div>
											<span>{formatDate(check.checked_at)}</span>
										</div>
									</td>
									   <td class="px-4 py-3 whitespace-nowrap">
										   <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {check.success ? 'bg-emerald-100 dark:bg-emerald-900 text-emerald-800 dark:text-emerald-300' : 'bg-rose-100 dark:bg-rose-900 text-rose-800 dark:text-rose-300'}">
											{check.success ? 'Up' : 'Down'}
										</span>
									</td>
									   <td class="px-4 py-3 whitespace-nowrap text-sm font-mono">
										   <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-slate-100 dark:bg-slate-900 text-slate-800 dark:text-slate-100">
											{extractInt64(check.status_code, 'N/A')}
										</span>
									</td>
									   <td class="px-4 py-3 whitespace-nowrap text-sm font-mono">
										{#if isValidSqlNull(check.response_time_ms)}
											   <span class="text-slate-700 dark:text-slate-300">{extractInt64(check.response_time_ms, 0)}ms</span>
										{:else}
											   <span class="text-slate-400 dark:text-slate-500">N/A</span>
										{/if}
									</td>
									   <td class="px-4 py-3 text-sm max-w-xs truncate">
										{#if extractString(check.error_message, '')}
											   <span class="text-rose-600 dark:text-rose-400" title={extractString(check.error_message, '')}>
												{extractString(check.error_message, '').length > 30 ? extractString(check.error_message, '').substring(0, 30) + '...' : extractString(check.error_message, '')}
											</span>
										{:else}
											   <span class="text-slate-400 dark:text-slate-500">-</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
				   <div class="px-4 py-3 bg-slate-50 dark:bg-slate-900 border-t border-slate-200 dark:border-slate-700">
					   <div class="flex items-center justify-between text-xs text-slate-600 dark:text-slate-400">
						<span>
							{checks.slice(0, 5).filter(c => c.success).length} successful, {checks.slice(0, 5).filter(c => !c.success).length} failed
						</span>
						<span>
							Avg: {Math.round(checks.slice(0, 5).filter(c => isValidSqlNull(c.response_time_ms)).reduce((sum, c) => sum + extractInt64(c.response_time_ms, 0), 0) / checks.slice(0, 5).filter(c => isValidSqlNull(c.response_time_ms)).length) || 0}ms
						</span>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
