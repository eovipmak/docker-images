<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import LineChart from './LineChart.svelte';

	export let monitorId: string;

	let stats: { timestamp: string; response_time_ms: number }[] = [];
	let isLoading = true;
	let error = '';

	// Transform stats data for LineChart component
	$: chartData = stats.map(stat => ({
		timestamp: stat.timestamp,
		value: stat.response_time_ms
	}));

	$: hasData = chartData.length > 0;

	onMount(async () => {
		await loadStats();
	});

	async function loadStats() {
		isLoading = true;
		error = '';

		try {
			const response = await fetchAPI(`/api/v1/monitors/${monitorId}/stats`);
			if (!response.ok) {
				throw new Error('Failed to load response time statistics');
			}
			const data = await response.json();
			stats = data || [];
		} catch (err: any) {
			console.error('Error loading stats:', err);
			error = err.message || 'Failed to load statistics';
		} finally {
			isLoading = false;
		}
	}
</script>

{#if isLoading}
	<div class="flex items-center justify-center h-64">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
	</div>
{:else if error}
	<div class="flex flex-col items-center justify-center h-64 text-slate-400">
		<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
		</svg>
		<p class="text-sm">{error}</p>
	</div>
{:else if hasData}
	<div class="h-64">
		<LineChart
			data={chartData}
			label="Response Time"
			color="#3B82F6"
			yAxisLabel="Response Time (ms)"
			timeRange={{ unit: 'hour' }}
			timeUnit="hour"
			spanGapsProp={true}
			suggestedMax={undefined}
		/>
	</div>
{:else}
	<div class="flex flex-col items-center justify-center h-64 text-slate-400">
		<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mb-2 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
		</svg>
		<p class="text-sm">No response time data available for the last 24 hours</p>
	</div>
{/if}