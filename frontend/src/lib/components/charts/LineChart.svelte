<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
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

	export let data: { timestamp: string; value: number }[] = [];
	export let label: string = 'Response Time';
	export let color: string = '#3B82F6';
	export let yAxisLabel: string = 'Response Time (ms)';
	export let fillOpacity: number = 0.2; // Opacity for area fill (0-1)

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	// Helper function to convert hex color to rgba
	function hexToRgba(hex: string, alpha: number): string {
		const r = parseInt(hex.slice(1, 3), 16);
		const g = parseInt(hex.slice(3, 5), 16);
		const b = parseInt(hex.slice(5, 7), 16);
		return `rgba(${r}, ${g}, ${b}, ${alpha})`;
	}

	onMount(() => {
		createChart();
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
		}
	});

	$: if (chart && data) {
		updateChart();
	}

	function createChart() {
		if (!canvas) return;

		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		chart = new Chart(ctx, {
			type: 'line',
			data: {
				labels: data.map((d) => new Date(d.timestamp)),
				datasets: [
					{
						label,
						data: data.map((d) => d.value),
						borderColor: color,
						backgroundColor: hexToRgba(color, fillOpacity),
						borderWidth: 2,
						tension: 0.4,
						fill: true,
						pointRadius: 3,
						pointHoverRadius: 5
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: {
						display: false
					},
					tooltip: {
						mode: 'index',
						intersect: false,
						callbacks: {
							label: function (context) {
								let label = context.dataset.label || '';
								if (label) {
									label += ': ';
								}
								if (context.parsed.y !== null) {
									label += Math.round(context.parsed.y * 100) / 100;
								}
								return label;
							}
						}
					}
				},
				scales: {
					x: {
						type: 'time',
						time: {
							displayFormats: {
								hour: 'MMM d, HH:mm',
								day: 'MMM d'
							}
						},
						grid: {
							display: false
						}
					},
					y: {
						beginAtZero: true,
						title: {
							display: true,
							text: yAxisLabel
						},
						grid: {
							color: '#E5E7EB'
						}
					}
				},
				interaction: {
					mode: 'nearest',
					axis: 'x',
					intersect: false
				}
			}
		});
	}

	function updateChart() {
		if (!chart) return;

		chart.data.labels = data.map((d) => new Date(d.timestamp));
		chart.data.datasets[0].data = data.map((d) => d.value);
		chart.update();
	}
</script>

<div class="w-full h-full">
	<canvas bind:this={canvas}></canvas>
</div>
