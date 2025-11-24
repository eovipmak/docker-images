<script lang="ts">
	import BarChart from './BarChart.svelte';
	import PieChart from './PieChart.svelte';
	import SummaryCard from './SummaryCard.svelte';

	export let data: { status_code: number; count: number }[] = [];
	export let label: string = 'Status Code Distribution';

	// Determine the visualization type based on the number of unique status codes
	$: visualizationType = data.length === 1 ? 'summary' : data.length <= 3 ? 'pie' : 'bar';
</script>

{#if visualizationType === 'summary'}
	<SummaryCard {data} />
{:else if visualizationType === 'pie'}
	<div class="h-64">
		<PieChart {data} {label} />
	</div>
{:else}
	<div class="h-64">
		<BarChart {data} {label} />
	</div>
{/if}