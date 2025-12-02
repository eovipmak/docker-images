<script lang="ts">
    import Card from './Card.svelte';
    export let title: string;
    export let value: string | number;
    export let valueColor: string = 'text-slate-900 dark:text-white';
    export let icon: string = ''; // Optional SVG path
</script>

<Card className="p-4 sm:p-6 transition-all duration-200 hover:shadow-md">
    <div class="flex items-center justify-between">
        <div class="min-w-0 flex-1">
            <p class="text-xs sm:text-sm font-medium text-slate-500 dark:text-slate-400 mb-1" title={title}>{title}</p>
            <h3 class="text-lg sm:text-2xl font-bold {valueColor} tracking-tight" title={String(value)}>{value}</h3>
        </div>
        {#if icon}
            <div class="p-2 sm:p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 text-slate-400 dark:text-slate-500 ml-2 flex-shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 sm:w-6 sm:h-6">
                    {@html icon}
                </svg>
            </div>
        {/if}
    </div>
</Card>
