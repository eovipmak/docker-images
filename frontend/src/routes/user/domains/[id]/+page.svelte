<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorStatus from '$lib/components/MonitorStatus.svelte';

	let monitor: any = null;
	let checks: any[] = [];
	let sslStatus: any = null;
	let isLoading = true;
	let error = '';

	$: monitorId = $page.params.id as string;

	onMount(() => {
		loadMonitorData();
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

			if (monitor.check_ssl && monitor.url.startsWith('https')) {
				const sslResponse = await fetchAPI(`/api/v1/monitors/${monitorId}/ssl-status`);
				if (sslResponse.ok) {
					sslStatus = await sslResponse.json();
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
		goto('/user/domains');
	}
</script>

<svelte:head>
	<title>{monitor?.name || 'Monitor Details'} - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 dark:bg-red-900/30 border border-red-400 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded mb-4">
			{error}
		</div>
		<button on:click={handleBack} class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium">
			← Back to Monitors
		</button>
	{:else if monitor}
		<div class="mb-6">
			<button
				on:click={handleBack}
				class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium mb-4 inline-flex items-center"
			>
				← Back to Monitors
			</button>
			<div class="flex justify-between items-start">
				<div>
					<h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">{monitor.name}</h1>
					<p class="text-gray-600 dark:text-gray-300">{monitor.url}</p>
				</div>
				<MonitorStatus status={getMonitorStatus()} />
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6">
				<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Status</h3>
				<p class="text-2xl font-bold text-gray-900 dark:text-white">
					{monitor.enabled ? 'Enabled' : 'Disabled'}
				</p>
			</div>
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6">
				<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Uptime (24h)</h3>
				<p class="text-2xl font-bold text-gray-900 dark:text-white">{calculateUptime()}</p>
			</div>
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6">
				<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Avg Response Time</h3>
				<p class="text-2xl font-bold text-gray-900 dark:text-white">{getAverageResponseTime()}</p>
			</div>
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6">
				<h3 class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">Check Interval</h3>
				<p class="text-2xl font-bold text-gray-900 dark:text-white">{monitor.check_interval}s</p>
			</div>
		</div>

		<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Uptime History (Last 24 Hours)</h2>
			{#if checks && checks.length > 0}
				<div class="flex items-end gap-1 h-48">
					{#each checks.slice(0, 48) as check}
						<div
							class="flex-1 rounded-t transition-all hover:opacity-75"
							class:bg-green-500={check.success}
							class:bg-red-500={!check.success}
							style="height: {check.success ? '100%' : '20%'}"
							title="{formatDate(check.checked_at)} - {check.success ? 'Up' : 'Down'}"
						></div>
					{/each}
				</div>
				<div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-2">
					<span>24h ago</span>
					<span>Now</span>
				</div>
			{:else}
				<p class="text-gray-500 dark:text-gray-400">No check history available</p>
			{/if}
		</div>

		<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Response Time (Last 24 Hours)</h2>
			{#if checks && checks.length > 0}
				{@const responseTimes = checks
					.filter((c) => c.response_time_ms)
					.slice(0, 48)
					.map((c) => c.response_time_ms)}
				{@const maxTime = Math.max(...responseTimes, 1)}
				<div class="flex items-end gap-1 h-48">
					{#each responseTimes as time}
						<div
							class="flex-1 bg-blue-500 dark:bg-blue-600 rounded-t transition-all hover:opacity-75"
							style="height: {(time / maxTime) * 100}%"
							title="{time}ms"
						></div>
					{/each}
				</div>
				<div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-2">
					<span>24h ago</span>
					<span>Max: {maxTime}ms</span>
					<span>Now</span>
				</div>
			{:else}
				<p class="text-gray-500 dark:text-gray-400">No response time data available</p>
			{/if}
		</div>

		{#if monitor.check_ssl && monitor.url.startsWith('https') && sslStatus}
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6 mb-8">
				<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">SSL Certificate</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">Valid</p>
						<p class="font-medium text-gray-900 dark:text-white">
							{sslStatus.ssl_valid ? 'Yes' : 'No'}
						</p>
					</div>
					{#if sslStatus.ssl_expires_at}
						<div>
							<p class="text-sm text-gray-500 dark:text-gray-400">Expires At</p>
							<p class="font-medium text-gray-900 dark:text-white">{formatDate(sslStatus.ssl_expires_at)}</p>
						</div>
					{/if}
					{#if sslStatus.ssl_issuer}
						<div>
							<p class="text-sm text-gray-500 dark:text-gray-400">Issuer</p>
							<p class="font-medium text-gray-900 dark:text-white">{sslStatus.ssl_issuer}</p>
						</div>
					{/if}
					{#if sslStatus.error_message}
						<div class="col-span-2">
							<p class="text-sm text-gray-500 dark:text-gray-400">Error</p>
							<p class="font-medium text-red-600 dark:text-red-400">{sslStatus.error_message}</p>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6 mb-8">
			<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Settings</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Timeout</p>
					<p class="font-medium text-gray-900 dark:text-white">{monitor.timeout}s</p>
				</div>
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">SSL Checks</p>
					<p class="font-medium text-gray-900 dark:text-white">{monitor.check_ssl ? 'Enabled' : 'Disabled'}</p>
				</div>
				{#if monitor.check_ssl}
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">SSL Alert (Days Before Expiry)</p>
						<p class="font-medium text-gray-900 dark:text-white">{monitor.ssl_alert_days} days</p>
					</div>
				{/if}
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Last Checked</p>
					<p class="font-medium text-gray-900 dark:text-white">
						{monitor.last_checked_at ? formatDate(monitor.last_checked_at) : 'Never'}
					</p>
				</div>
			</div>
		</div>

		{#if checks && checks.length > 0}
			<div class="bg-white dark:bg-slate-800 rounded-lg shadow-md dark:shadow-none border dark:border-slate-700 p-6">
				<h2 class="text-xl font-bold text-gray-900 dark:text-white mb-4">Recent Checks</h2>
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200 dark:divide-slate-700">
						<thead class="bg-gray-50 dark:bg-slate-900/50">
							<tr>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									Status
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									Status Code
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									Response Time
								</th>
								<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									Error
								</th>
							</tr>
						</thead>
						<tbody class="bg-white dark:bg-slate-800 divide-y divide-gray-200 dark:divide-slate-700">
							{#each checks.slice(0, 10) as check}
								<tr>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
										{formatDate(check.checked_at)}
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<MonitorStatus status={check.success ? 'up' : 'down'} showText={true} />
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
										{check.status_code || 'N/A'}
									</td>
									<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
										{check.response_time_ms ? `${check.response_time_ms}ms` : 'N/A'}
									</td>
									<td class="px-6 py-4 text-sm text-red-600 dark:text-red-400">
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
