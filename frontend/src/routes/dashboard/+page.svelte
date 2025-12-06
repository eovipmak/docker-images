<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks, latestIncidents, connectEventStream, disconnectEventStream } from '$lib/api/events';
	import ModernStatCard from '$lib/components/ModernStatCard.svelte';
	import { goto } from '$app/navigation';
	import type { MaintenanceWindow } from '$lib/types';

	interface DashboardStats {
		total_monitors: number;
		up_count: number;
		down_count: number;
		open_incidents: number;
		average_response_time: number;
		overall_uptime: number;
	}

	let stats: DashboardStats = {
		total_monitors: 0,
		up_count: 0,
		down_count: 0,
		open_incidents: 0,
		average_response_time: 0,
		overall_uptime: 0
	};
    
	let isLoading = true;

	async function loadDashboardData() {
		try {
			const response = await fetchAPI('/api/v1/dashboard');
			if (response.ok) {
				const data: any = await response.json();
				stats = data.stats;
			}
		} catch (err) {
			console.error('Error loading dashboard:', err);
		} finally {
			isLoading = false;
		}
	}

	onMount(async () => {
		await loadDashboardData();
		await connectEventStream();
	});

	onDestroy(() => {
		disconnectEventStream();
	});
</script>

<svelte:head>
	<title>Dashboard - V-Insight</title>
</svelte:head>

<div class="space-y-8 pb-10 pt-8">

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-10">
        
        <!-- Dashboard Overview -->
        <div>
            <h2 class="text-xl font-bold text-slate-900 dark:text-white mb-6 flex items-center gap-2">
                <span class="w-2 h-8 rounded-full bg-cyan-500"></span>
                Dashboard Overview
            </h2>
            
            <!-- Top Metrics Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                <ModernStatCard 
                    title="Total Monitors" 
                    value={stats.total_monitors} 
                    color="cyan"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />'
                />
                 <ModernStatCard 
                    title="Active Incidents" 
                    value={stats.open_incidents} 
                    color="rose"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.008v.008H12v-.008z" />'
                />
                 <ModernStatCard 
                    title="Uptime (24h)" 
                    value={`${stats.overall_uptime?.toFixed(2) || '100'}%`} 
                    color="emerald"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />'
                />
                 <ModernStatCard 
                    title="Avg Response" 
                    value={`${Math.round(stats.average_response_time || 0)}ms`} 
                    color="indigo"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />'
                />
            </div>
            
             <!-- Secondary Metrics Grid -->
             <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <ModernStatCard 
                    title="Operational Services" 
                    value={stats.up_count} 
                    color="emerald"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />'
                />
                <ModernStatCard 
                    title="Services Down" 
                    value={stats.down_count} 
                    color="rose"
                    icon='<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />'
                />
             </div>
        </div>

    </div>
</div>