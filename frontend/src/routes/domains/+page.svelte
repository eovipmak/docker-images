<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks } from '$lib/api/events';
	import MonitorTable from '$lib/components/MonitorTable.svelte';

	let monitors: any[] = [];
	let isLoading = true;
	let error = '';
	let isModalOpen = false;
	let selectedMonitor: any = null;

	// Lazy loaded modal component
	let MonitorModal: any = null;
	let modalLoaded = false;

	// Subscribe to monitor check events
	let unsubscribe: (() => void) | null = null;

	onMount(() => {
		loadMonitors();

		// Subscribe to SSE events to update monitor status in real-time
		unsubscribe = latestMonitorChecks.subscribe((checks) => {
			// Update monitors with latest check data
			monitors = monitors.map((monitor) => {
				const latestCheck = checks.get(monitor.id);
				if (latestCheck) {
					return {
						...monitor,
						last_check: latestCheck,
						status: latestCheck.success ? 'up' : 'down'
					};
				}
				return monitor;
			});
		});
	});

	onDestroy(() => {
		if (unsubscribe) {
			unsubscribe();
		}
	});

	async function loadModal() {
		if (!modalLoaded) {
			try {
				const module = await import('$lib/components/MonitorModal.svelte');
				MonitorModal = module.default;
				modalLoaded = true;
			} catch (err) {
				console.error('Failed to load MonitorModal:', err);
			}
		}
	}

	async function loadMonitors() {
		isLoading = true;
		error = '';

		try {
			const response = await fetchAPI('/api/v1/monitors');

			if (!response.ok) {
				throw new Error('Failed to load monitors');
			}

			monitors = await response.json();
		} catch (err: any) {
			console.error('Error loading monitors:', err);
			error = err.message || 'Failed to load monitors';
		} finally {
			isLoading = false;
		}
	}

	async function handleAddMonitor() {
		await loadModal();
		selectedMonitor = null;
		isModalOpen = true;
	}

	function handleViewMonitor(event: CustomEvent) {
		const monitor = event.detail;
		goto(`/domains/${monitor.id}`);
	}

	async function handleEditMonitor(event: CustomEvent) {
		await loadModal();
		selectedMonitor = event.detail;
		isModalOpen = true;
	}

	async function handleDeleteMonitor(event: CustomEvent) {
		const monitor = event.detail;

		if (!confirm(`Are you sure you want to delete "${monitor.name}"?`)) {
			return;
		}

		try {
			const response = await fetchAPI(`/api/v1/monitors/${monitor.id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				throw new Error('Failed to delete monitor');
			}

			monitors = monitors.filter((m) => m.id !== monitor.id);
		} catch (err: any) {
			console.error('Error deleting monitor:', err);
			alert(err.message || 'Failed to delete monitor');
		}
	}

	function handleMonitorSaved(event: CustomEvent) {
		const savedMonitor = event.detail;
		const index = monitors.findIndex((m) => m.id === savedMonitor.id);

		if (index !== -1) {
			monitors[index] = savedMonitor;
		} else {
			monitors = [savedMonitor, ...monitors];
		}
	}
</script>

<svelte:head>
	<title>Monitors - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto space-y-8">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-slate-900">Monitors</h1>
			<p class="mt-1 text-sm text-slate-500">Manage your website and service monitors.</p>
		</div>
		<button
			on:click={handleAddMonitor}
			class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
		>
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
			</svg>
			Add Monitor
		</button>
	</div>

	{#if error}
		<div class="p-4 rounded-lg bg-red-50 border border-red-200 text-red-700 flex items-center">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
			</svg>
			{error}
		</div>
	{/if}

	{#if isLoading}
		<div class="bg-white rounded-xl shadow-sm border border-slate-200 p-6 space-y-4">
            {#each Array(5) as _}
                <div class="h-12 bg-slate-100 rounded-lg animate-pulse"></div>
            {/each}
        </div>
	{:else}
		<MonitorTable
			{monitors}
			on:view={handleViewMonitor}
			on:edit={handleEditMonitor}
			on:delete={handleDeleteMonitor}
		/>
	{/if}
</div>

{#if modalLoaded}
	<svelte:component
		this={MonitorModal}
		isOpen={isModalOpen}
		monitor={selectedMonitor}
		on:close={() => (isModalOpen = false)}
		on:save={handleMonitorSaved}
	/>
{/if}
