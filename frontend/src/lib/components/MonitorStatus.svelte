<script lang="ts">
	export let status: 'up' | 'down' | 'unknown';
	export let showText: boolean = true;

	const statusConfig = {
		up: {
			color: 'text-emerald-700',
			bg: 'bg-emerald-50 border-emerald-200',
            dot: 'bg-emerald-500',
			text: 'Operational'
		},
		down: {
			color: 'text-rose-700',
			bg: 'bg-rose-50 border-rose-200',
            dot: 'bg-rose-500',
			text: 'Downtime'
		},
		unknown: {
			color: 'text-slate-700',
			bg: 'bg-slate-50 border-slate-200',
            dot: 'bg-slate-400',
			text: 'Unknown'
		}
	};

	const config = statusConfig[status] || statusConfig.unknown;
</script>

<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border {config.bg} {config.color}">
    <span class="relative flex h-2 w-2 mr-2">
        {#if status === 'down'}
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full {config.dot} opacity-75"></span>
        {/if}
        <span class="relative inline-flex rounded-full h-2 w-2 {config.dot}"></span>
    </span>
	{#if showText}
		{config.text}
	{/if}
</span>
