<script lang="ts">
	export let status: 'up' | 'down' | 'unknown';
	export let showText: boolean = true;

	const statusConfig = {
		up: {
			color: 'text-green-600',
			bg: 'bg-green-100',
			text: 'Up'
		},
		down: {
			color: 'text-red-600',
			bg: 'bg-red-100',
			text: 'Down'
		},
		unknown: {
			color: 'text-gray-600',
			bg: 'bg-gray-100',
			text: 'Unknown'
		}
	};

	const config = statusConfig[status] || statusConfig.unknown;
</script>

<span class="inline-flex items-center gap-2">
	<span class="flex h-3 w-3 relative">
		<span class="animate-ping absolute inline-flex h-full w-full rounded-full {config.bg} opacity-75"></span>
		<span class="relative inline-flex rounded-full h-3 w-3 {config.bg}"></span>
	</span>
	{#if showText}
		<span class="text-sm font-medium {config.color}">{config.text}</span>
	{/if}
</span>
