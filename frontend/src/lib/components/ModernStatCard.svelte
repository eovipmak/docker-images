<script lang="ts">
    export let title: string;
    export let value: string | number;
    export let trend: number | null = null; // Percentage change + or -
    export let icon: string; // SVG path d
    export let color: string = "blue"; // emerald, blue, rose, etc.
    export let sparklineData: number[] = []; // Array of numbers 0-100

    // Color mapping for gradients
    const colors: Record<string, any> = {
        emerald: {
            bg: "bg-emerald-500/10",
            text: "text-emerald-500",
            stroke: "#10b981", 
            gradient: "from-emerald-500 to-teal-500"
        },
        blue: {
            bg: "bg-blue-500/10",
            text: "text-blue-500",
            stroke: "#3b82f6",
            gradient: "from-blue-500 to-indigo-500"
        },
        rose: {
            bg: "bg-rose-500/10",
            text: "text-rose-500",
            stroke: "#f43f5e",
            gradient: "from-rose-500 to-pink-500"
        },
        amber: {
            bg: "bg-amber-500/10",
            text: "text-amber-500",
            stroke: "#f59e0b",
            gradient: "from-amber-500 to-orange-500"
        },
        cyan: {
             bg: "bg-cyan-500/10",
             text: "text-cyan-400",
             stroke: "#22d3ee",
             gradient: "from-cyan-400 to-blue-500"
        },
        purple: {
            bg: "bg-purple-500/10",
            text: "text-purple-500",
            stroke: "#a855f7",
            gradient: "from-purple-500 to-fuchsia-500"
        },
         indigo: {
            bg: "bg-indigo-500/10",
            text: "text-indigo-500",
            stroke: "#6366f1",
            gradient: "from-indigo-500 to-violet-500"
        }
    };

    $: theme = colors[color] || colors.blue;

    // Generate simple SVG path for sparkline
    function getSparklinePath(data: number[], width: number, height: number): string {
        if (!data || data.length === 0) return "";
        const max = Math.max(...data, 1);
        const step = width / (data.length - 1);
        
        return data.map((val, i) => {
            const x = i * step;
            const y = height - (val / max) * height;
            return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
        }).join(' ');
    }

    // Mock data if none provided, just for visual
    $: displayData = sparklineData.length > 0 ? sparklineData : [40, 30, 45, 50, 40, 60, 55, 70, 60, 80];
</script>

<div class="relative overflow-hidden rounded-2xl bg-white dark:bg-[#1a1c2e] p-6 shadow-lg border border-gray-100 dark:border-white/5 transition-transform hover:scale-[1.02] duration-300">
    <div class="flex items-start justify-between mb-4">
        <div>
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">{title}</p>
            <h3 class="text-3xl font-bold text-gray-900 dark:text-white">{value}</h3>
        </div>
        <div class="p-3 rounded-xl {theme.bg} {theme.text}">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                {@html icon}
            </svg>
        </div>
    </div>

    <div class="flex items-end justify-between">
        {#if trend !== null}
            <div class="flex items-center gap-1 text-sm font-medium {trend >= 0 ? 'text-emerald-500' : 'text-rose-500'}">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4 {trend < 0 ? 'rotate-180' : ''}">
                    <path fill-rule="evenodd" d="M10 17a.75.75 0 01-.75-.75V5.612L5.29 9.77a.75.75 0 01-1.08-1.04l5.25-5.5a.75.75 0 011.08 0l5.25 5.5a.75.75 0 11-1.08 1.04l-3.96-4.158V16.25A.75.75 0 0110 17z" clip-rule="evenodd" />
                </svg>
                {Math.abs(trend)}%
            </div>
        {/if}

        {#if sparklineData.length > 0}
            <div class="h-10 w-24 ml-auto">
                 <svg width="100%" height="100%" viewBox="0 0 100 40" preserveAspectRatio="none" class="overflow-visible">
                    <!-- Glowing line shadow -->
                    <path d={getSparklinePath(sparklineData, 100, 40)} fill="none" stroke={theme.stroke} stroke-width="4" stroke-opacity="0.2" class="blur-[2px]" />
                    <!-- Actual line -->
                    <path d={getSparklinePath(sparklineData, 100, 40)} fill="none" stroke={theme.stroke} stroke-width="2" vector-effect="non-scaling-stroke" />
                 </svg>
            </div>
        {/if}
    </div>
    
    <!-- Decorative gradient glow -->
    <div class="absolute -top-10 -right-10 w-32 h-32 bg-gradient-to-br {theme.gradient} opacity-10 blur-2xl rounded-full pointer-events-none"></div>
</div>
