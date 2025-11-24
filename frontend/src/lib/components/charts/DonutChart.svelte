<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart,
		DoughnutController,
		ArcElement,
		Tooltip,
		Legend
	} from 'chart.js';

	// Register Chart.js components
	Chart.register(DoughnutController, ArcElement, Tooltip, Legend);

	export let percentage: number = 0;
	export let label: string = 'Uptime';
	export let successColor: string = '#10B981';
	export let failureColor: string = '#EF4444';

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	onMount(() => {
		createChart();
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
		}
	});

	$: if (chart && percentage !== undefined) {
		updateChart();
	}

	function createChart() {
		if (!canvas) return;

		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		// Ensure percentage is a valid number
		const validPercentage = isNaN(percentage) || percentage === undefined ? 0 : percentage;
		const success = Math.max(0, Math.min(100, validPercentage));
		const failure = 100 - success;

		chart = new Chart(ctx, {
			type: 'doughnut',
			data: {
				labels: ['Up', 'Down'],
				datasets: [
					{
						data: [success, failure],
						backgroundColor: [successColor, failureColor],
						borderWidth: 0
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: '70%',
				plugins: {
					legend: {
						display: true,
						position: 'bottom'
					},
					tooltip: {
						callbacks: {
							label: function (context) {
								let label = context.label || '';
								if (label) {
									label += ': ';
								}
								if (context.parsed !== null) {
									label += context.parsed.toFixed(2) + '%';
								}
								return label;
							}
						}
					}
				}
			}
		});
	}

	function updateChart() {
		if (!chart) return;

		// Ensure percentage is a valid number
		const validPercentage = isNaN(percentage) || percentage === undefined ? 0 : percentage;
		const success = Math.max(0, Math.min(100, validPercentage));
		const failure = 100 - success;

		chart.data.datasets[0].data = [success, failure];
		chart.update();
	}
</script>

<div class="w-full h-full relative">
	<canvas bind:this={canvas}></canvas>
	<div class="absolute inset-0 flex items-center justify-center pointer-events-none">
		<div class="text-center">
			<div class="text-2xl font-bold text-gray-900">
				{isNaN(percentage) || percentage === undefined ? '0.0' : percentage.toFixed(1)}%
			</div>
			<div class="text-sm text-gray-500">{label}</div>
		</div>
	</div>
</div>
