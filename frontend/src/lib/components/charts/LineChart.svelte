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

	export let data: { timestamp: string; value: number | null; min?: number | null; max?: number | null }[] = [];
	export let label: string = 'Response Time';
	export let color: string = '#3B82F6';
	export let yAxisLabel: string = 'Response Time (ms)';
export let timeRange: { min?: string | Date; max?: string | Date; unit?: 'minute' | 'hour' | 'day' } | null = null;
export let timeUnit: 'minute' | 'hour' | 'day' | undefined = undefined;
export let spanGapsProp: boolean | undefined = undefined; // allow caller to request spanning gaps
	export let suggestedMax: number | undefined = undefined;
	export let fillOpacity: number = 0.2; // Opacity for area fill (0-1)

	let canvas: HTMLCanvasElement;
	let chart: any = null;

	// Helper function to convert hex color to rgba
	function hexToRgba(hex: string, alpha: number): string {
		const r = parseInt(hex.slice(1, 3), 16);
		const g = parseInt(hex.slice(3, 5), 16);
		const b = parseInt(hex.slice(5, 7), 16);
		return `rgba(${r}, ${g}, ${b}, ${alpha})`;
	}

	function getDatasets() {
		const hasMinMax = data.some(d => d.min !== undefined && d.max !== undefined);
		const datasets = [];

		if (hasMinMax) {
			// Max line
			datasets.push({
				label: 'Max',
				data: data.map((d) => ({ x: new Date(d.timestamp).getTime(), y: d.max })),
				borderColor: hexToRgba(color, 0.6),
				backgroundColor: 'transparent',
				borderWidth: 1,
				tension: 0.2,
				spanGaps: typeof spanGapsProp === 'boolean' ? spanGapsProp : (timeUnit === 'minute'),
				fill: false,
				pointRadius: 0,
				pointHoverRadius: 3
			});
			// Avg line
			datasets.push({
				label: 'Avg',
				data: data.map((d) => ({ x: new Date(d.timestamp).getTime(), y: d.value })),
				borderColor: color,
				backgroundColor: hexToRgba(color, fillOpacity),
				borderWidth: 2,
				tension: 0.2,
				spanGaps: typeof spanGapsProp === 'boolean' ? spanGapsProp : (timeUnit === 'minute'),
				fill: true,
				pointRadius: data.length > 100 ? 0 : 3,
				pointHoverRadius: 5
			});
			// Min line
			datasets.push({
				label: 'Min',
				data: data.map((d) => ({ x: new Date(d.timestamp).getTime(), y: d.min })),
				borderColor: hexToRgba(color, 0.4),
				backgroundColor: 'transparent',
				borderWidth: 1,
				tension: 0.2,
				spanGaps: typeof spanGapsProp === 'boolean' ? spanGapsProp : (timeUnit === 'minute'),
				fill: false,
				pointRadius: 0,
				pointHoverRadius: 3
			});
		} else {
			// Single line
			datasets.push({
				label,
				data: data.map((d) => ({ x: new Date(d.timestamp).getTime(), y: d.value })),
				borderColor: color,
				backgroundColor: hexToRgba(color, fillOpacity),
				borderWidth: 2,
				tension: (timeUnit === 'day' || data.length > 200) ? 0 : 0.12,
				cubicInterpolationMode: 'monotone',
				spanGaps: typeof spanGapsProp === 'boolean' ? spanGapsProp : (timeUnit === 'minute'),
				fill: true,
				pointRadius: data.length > 100 ? 0 : 3,
				pointHoverRadius: 5
			});
		}
		return datasets;
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

	// Also update when timeRange changes
	$: if (chart && timeRange) {
		updateChart();
	}

	function createChart() {
		if (!canvas) return;

		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		const datasets = getDatasets();
		// Chart.js typings are strict about dataset generics; cast to any here to avoid complex generic plumbing
		chart = new Chart(ctx as any, {
			type: 'line',
			data: {
				datasets
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: {
						display: datasets.length > 1
					},
					tooltip: {
						mode: 'index',
						intersect: false,
						callbacks: {
							label: function (context: any) {
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
							unit: timeUnit || (timeRange?.unit as any) || undefined,
							displayFormats: {
								hour: 'MMM d, HH:mm',
								day: 'MMM d'
							}
						},
						grid: {
							display: false
						},
						// Optionally set min/max if provided (to control time window)
						...(timeRange ? { min: timeRange.min ? (new Date(timeRange.min).getTime()) : undefined, max: timeRange.max ? (new Date(timeRange.max).getTime()) : undefined } : {})
					},
					y: {
						beginAtZero: true,
						suggestedMax: suggestedMax || undefined,
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
		} as any);
	}

	function updateChart() {
		if (!chart) return;

		const newDatasets = getDatasets();
		chart.data.datasets = newDatasets;
		if (chart.options && chart.options.plugins && chart.options.plugins.legend) {
			chart.options.plugins.legend.display = newDatasets.length > 1;
		}
		
		// Update timeRange before calling chart.update()
		if (timeRange && chart.options && chart.options.scales && chart.options.scales.x) {
			const xScale: any = chart.options.scales.x;
			xScale.min = timeRange.min ? new Date(timeRange.min).getTime() : undefined;
			xScale.max = timeRange.max ? new Date(timeRange.max).getTime() : undefined;
		}
		
		chart.update();
	}
</script>

<div class="w-full h-full">
	<canvas bind:this={canvas}></canvas>
</div>
