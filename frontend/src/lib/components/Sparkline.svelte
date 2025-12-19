<script lang="ts">
	export let data: { time: Date; value: number; success: boolean }[] = [];
	export let width = 120;
	export let height = 32;

	$: points = calculatePoints(data, width, height);

	function calculatePoints(data: { time: Date; value: number; success: boolean }[], w: number, h: number): string {
		if (data.length === 0) return '';
		
		const values = data.map(d => d.value);
		const maxValue = Math.max(...values, 1);
		const minValue = Math.min(...values, 0);
		const range = maxValue - minValue || 1;
		
		const stepX = w / (data.length - 1 || 1);
		
		return data.map((point, i) => {
			const x = i * stepX;
			const y = h - ((point.value - minValue) / range) * h;
			return `${x.toFixed(1)},${y.toFixed(1)}`;
		}).join(' ');
	}
</script>

<svg {width} {height} class="sparkline" viewBox="0 0 {width} {height}">
	<polyline
		points={points}
		fill="none"
		stroke="currentColor"
		stroke-width="1.5"
		stroke-linecap="round"
		stroke-linejoin="round"
		class="opacity-70"
	/>
	{#each data as point, i}
		{#if i === data.length - 1}
			{@const x = (i * width) / (data.length - 1 || 1)}
			{@const values = data.map(d => d.value)}
			{@const maxValue = Math.max(...values, 1)}
			{@const minValue = Math.min(...values, 0)}
			{@const range = maxValue - minValue || 1}
			{@const y = height - ((point.value - minValue) / range) * height}
			<circle
				cx={x}
				cy={y}
				r="2.5"
				fill={point.success ? 'rgb(34, 197, 94)' : 'rgb(248, 113, 113)'}
				class="drop-shadow-sm"
			/>
		{/if}
	{/each}
</svg>

<style>
	.sparkline {
		overflow: visible;
	}
</style>
