<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import { latestMonitorChecks } from '$lib/api/events';
	import type { Monitor } from '$lib/types';
	import MonitorCard from '$lib/components/MonitorCard.svelte';
	import MonitorList from '$lib/components/MonitorList.svelte';
	import Card from '$lib/components/Card.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let monitors: Monitor[] = [];
	let isLoading = true;
	let error = '';
	let isModalOpen = false;
	let selectedMonitor: Monitor | null = null;
	let searchQuery = '';
	let sortField: 'name' | 'status' = 'name';
	let sortDirection: 'asc' | 'desc' = 'asc';
	let useTable: boolean = false;
	let selectedTag: string = '';  // New: for tag filtering

	// Confirm modal state
	let isConfirmModalOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let onConfirmCallback: (() => void) | null = null;

	// Lazy loaded modal component
	let MonitorModal: any = null;
	let modalLoaded = false;

	// Subscribe to monitor check events
	let unsubscribe: (() => void) | null = null;

	// Get all unique tags from monitors
	$: allTags = [...new Set(monitors.flatMap(m => m.tags || []))].sort();

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
		console.debug('[Monitors page] view monitor', monitor?.id);
		if (!monitor || !monitor.id) {
			// visible feedback for debugging
			alert('Monitor or monitor.id missing (check console)');
			return;
		}

		// Try SPA navigation first; fall back to full navigation if it fails
		goto(`/user/monitors/${monitor.id}`).catch(() => {
			window.location.href = `/user/monitors/${monitor.id}`;
		});
	}

	async function handleEditMonitor(event: CustomEvent) {
		await loadModal();
		selectedMonitor = event.detail;
		isModalOpen = true;
	}

	async function handleDeleteMonitor(event: CustomEvent) {
		const monitor = event.detail;

		confirmTitle = 'Delete Monitor';
		confirmMessage = `Are you sure you want to delete "${monitor.name}"?`;
		onConfirmCallback = async () => {
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
		};
		isConfirmModalOpen = true;
	}

	function handleConfirmDelete() {
		if (onConfirmCallback) {
			onConfirmCallback();
		}
		isConfirmModalOpen = false;
	}

	function handleCancelDelete() {
		isConfirmModalOpen = false;
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

	// Format relative time (simple copy from MonitorTable)
	function formatRelativeTime(dateString?: string): string {
		if (!dateString) return 'Never';
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return 'Invalid Date';
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		return `${diffDays}d ago`;
	}

	// Sort monitors
	function sortMonitors(field: 'name' | 'status') {
		if (sortField === field) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortDirection = 'asc';
		}
	}

	// Filter and sort monitors (reactive)
	$: filteredAndSortedMonitors = monitors
		.filter((monitor) => {
			// Text search filter
			if (searchQuery) {
				const query = searchQuery.toLowerCase();
				const matchesSearch = monitor.name.toLowerCase().includes(query) ||
					monitor.url.toLowerCase().includes(query);
				if (!matchesSearch) return false;
			}
			// Tag filter
			if (selectedTag) {
				const hasTag = monitor.tags?.includes(selectedTag);
				if (!hasTag) return false;
			}
			return true;
		})
		.sort((a, b) => {
			let aVal: any, bVal: any;
			switch (sortField) {
				case 'name':
					aVal = a.name.toLowerCase();
					bVal = b.name.toLowerCase();
					break;
				case 'status':
					aVal = a.status || (a.enabled ? 'up' : 'unknown');
					bVal = b.status || (b.enabled ? 'up' : 'unknown');
					break;
				default:
					return 0;
			}

			if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
			if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
			return 0;
		});
</script>

<svelte:head>
	<title>Monitors - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">Monitors</h1>
			  <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Manage your website and service monitors.</p>
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
		<div class="p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 flex items-center">
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
			</svg>
			{error}
		</div>
	{/if}

	{#if isLoading}
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 space-y-4">
            {#each Array(5) as _}
                <div class="h-12 bg-slate-100 dark:bg-slate-700 rounded-lg animate-pulse"></div>
            {/each}
        </div>
	{:else}
		<!-- Toolbar (search, tag filter & sort) -->
		<Card className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
			<div class="flex flex-col sm:flex-row gap-3 flex-1">
				<!-- Search -->
				<div class="relative max-w-md w-full">
					<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-slate-400">
							<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
						</svg>
					</div>
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Search monitors..."
						class="block w-full pl-10 pr-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg leading-5 bg-white dark:bg-slate-800 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition-shadow text-slate-900 dark:text-white"
					/>
				</div>
				<!-- Tag filter -->
				{#if allTags.length > 0}
					<div class="relative">
						<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
							<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 text-slate-400">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9.568 3H5.25A2.25 2.25 0 003 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 005.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 009.568 3z" />
								<path stroke-linecap="round" stroke-linejoin="round" d="M6 6h.008v.008H6V6z" />
							</svg>
						</div>
						<select 
							bind:value={selectedTag} 
							class="block w-full pl-9 pr-8 py-2 border border-slate-300 dark:border-slate-600 rounded-lg text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 bg-white dark:bg-slate-800 text-slate-900 dark:text-white"
						>
							<option value="">All tags</option>
							{#each allTags as tag}
								<option value={tag}>{tag}</option>
							{/each}
						</select>
					</div>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				<select bind:value={sortField} class="block w-full rounded-lg border-slate-300 dark:border-slate-600 py-2 pl-3 pr-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 sm:text-sm bg-white dark:bg-slate-800 text-slate-900 dark:text-white">
					<option value="name">Name</option>
					<option value="status">Status</option>
				</select>
				<button
					class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
					on:click={() => (sortDirection = sortDirection === 'asc' ? 'desc' : 'asc')}
					title={sortDirection === 'asc' ? 'Ascending' : 'Descending'}
				>
					{#if sortDirection === 'asc'}
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h9.75m4.5-4.5v12m0 0l-3.75-3.75M17.25 21L21 17.25" />
						</svg>
					{:else}
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
							<path stroke-linecap="round" stroke-linejoin="round" d="M3 4.5h14.25M3 9h9.75M3 13.5h5.25m5.25-.75L17.25 9m0 0L21 12.75M17.25 9v12" />
						</svg>
					{/if}
				</button>
			</div>
		</Card>

		<!-- Content: table or grid -->
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-4">
			<div class="flex items-center justify-between">
				<div class="text-sm text-slate-600 dark:text-slate-400">{filteredAndSortedMonitors.length} monitor{filteredAndSortedMonitors.length !== 1 ? 's' : ''}</div>
				<div class="flex items-center gap-2">
					<button
						class="inline-flex items-center px-2 py-1 rounded-md bg-slate-50 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-600"
						on:click={() => (useTable = !useTable)}
						title={useTable ? 'Switch to Grid' : 'Switch to Table'}
					>
						{#if useTable}
							Table
						{:else}
							Grid
						{/if}
					</button>
				</div>
			</div>
		</div>

		{#if filteredAndSortedMonitors.length === 0}
			<div class="mt-4 bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-12 text-center">
				<div class="flex flex-col items-center justify-center">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-12 h-12 text-slate-300 dark:text-slate-600 mb-4">
						<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
					</svg>
					<h3 class="text-lg font-medium text-slate-900 dark:text-white">No monitors found</h3>
					<p class="mt-1 text-slate-500 dark:text-slate-400">Try adjusting your search or add a new monitor.</p>
				</div>
			</div>
		{:else if useTable}
			<div class="mt-4">
				<MonitorList monitors={filteredAndSortedMonitors} useTable={true} on:view={handleViewMonitor} on:edit={handleEditMonitor} on:delete={handleDeleteMonitor} />
			</div>
		{:else}
			<!-- Grid (using MonitorList) -->
			<div class="mt-4">
				<MonitorList monitors={filteredAndSortedMonitors} on:view={handleViewMonitor} on:edit={handleEditMonitor} on:delete={handleDeleteMonitor} />
			</div>
		{/if}
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

<ConfirmModal
	isOpen={isConfirmModalOpen}
	title={confirmTitle}
	message={confirmMessage}
	on:confirm={handleConfirmDelete}
	on:cancel={handleCancelDelete}
/>
