<script lang="ts">
	export let title: string;
	export let value: string | number;
	export let valueColor: string = 'text-slate-900';
	export let bgColor: string = 'bg-white';
    export let icon: string = ''; // Optional SVG path
</script>

<div class="{bgColor} rounded-xl shadow-sm border border-slate-200 p-6 transition-all duration-200 hover:shadow-md">
    <div class="flex items-center justify-between">
        <div>
            <p class="text-sm font-medium text-slate-500 mb-1">{title}</p>
            <h3 class="text-2xl font-bold {valueColor} tracking-tight">{value}</h3>
        </div>
        {#if icon}
            <div class="p-3 rounded-lg bg-slate-50 text-slate-400">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                    {@html icon}
                </svg>
            </div>
        {/if}
    </div>
</div>
