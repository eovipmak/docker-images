<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart,
		BarController,
		BarElement,
		CategoryScale,
		LinearScale,
		Title,
		Tooltip,
		Legend
	} from 'chart.js';

	// Register Chart.js components
	Chart.register(BarController, BarElement, CategoryScale, LinearScale, Title, Tooltip, Legend);

	export let data: { status_code: number; count: number }[] = [];
	export const label: string = 'Status Code Distribution';

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

	$: if (chart && data) {
		updateChart();
	}

	function getColorForStatusCode(statusCode: number): string {
		if (statusCode >= 200 && statusCode < 300) return '#10B981'; // Green for success
		if (statusCode >= 300 && statusCode < 400) return '#3B82F6'; // Blue for redirects
		if (statusCode >= 400 && statusCode < 500) return '#F59E0B'; // Yellow for client errors
		if (statusCode >= 500) return '#EF4444'; // Red for server errors
		return '#6B7280'; // Gray for unknown
	}

	function createChart() {
		if (!canvas) return;

		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		chart = new Chart(ctx, {
			type: 'bar',
			data: {
				labels: data.map((d) => d.status_code.toString()),
				datasets: [
					{
						label: 'Count',
						data: data.map((d) => d.count),
						backgroundColor: data.map((d) => getColorForStatusCode(d.status_code)),
						borderWidth: 0,
						borderRadius: 4
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
						callbacks: {
							title: function (context) {
								return 'Status Code: ' + context[0].label;
							},
							label: function (context) {
								return 'Count: ' + context.parsed.y;
							}
						}
					}
				},
				scales: {
					x: {
						title: {
							display: true,
							text: 'Status Code'
						},
						grid: {
							display: false
						}
					},
					y: {
						beginAtZero: true,
						title: {
							display: true,
							text: 'Count'
						},
						grid: {
							color: window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'rgba(255, 255, 255, 0.1)' : '#E5E7EB'
						},
						ticks: {
							precision: 0,
                            color: window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? '#9ca3af' : '#6b7280'
						}
					}
				}
			}
		});
	}

	function updateChart() {
		if (!chart) return;

		chart.data.labels = data.map((d) => d.status_code.toString());
		chart.data.datasets[0].data = data.map((d) => d.count);
		chart.data.datasets[0].backgroundColor = data.map((d) => getColorForStatusCode(d.status_code));
		chart.update();
	}
</script>

<div class="w-full h-full">
	<canvas bind:this={canvas}></canvas>
</div>
