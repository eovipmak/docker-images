<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import MonitorStatus from './MonitorStatus.svelte';
	import Favicon from './Favicon.svelte';

	export let monitors: any[] = [];

	const dispatch = createEventDispatcher();

	type SortField = 'name' | 'status';
	type SortDirection = 'asc' | 'desc';

	let sortField: SortField = 'name';
	let sortDirection: SortDirection = 'asc';
	let searchQuery = '';

	// Format relative time
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

	// Get monitor status based on last check
	function getMonitorStatus(monitor: any): 'up' | 'down' | 'unknown' {
		if (!monitor.last_checked_at) return 'unknown';
		// This is a simplified status - in reality you'd check the last check result
		return monitor.enabled ? 'up' : 'unknown';
	}

	// Sort monitors
	function sortMonitors(field: SortField) {
		if (sortField === field) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortDirection = 'asc';
		}
	}

	// Filter and sort monitors
	$: filteredAndSortedMonitors = monitors
		.filter((monitor) => {
			if (!searchQuery) return true;
			const query = searchQuery.toLowerCase();
			return (
				monitor.name.toLowerCase().includes(query) ||
				monitor.url.toLowerCase().includes(query)
			);
		})
		.sort((a, b) => {
			let aVal, bVal;

			switch (sortField) {
				case 'name':
					aVal = a.name.toLowerCase();
					bVal = b.name.toLowerCase();
					break;
				case 'status':
					// Simple sort by status string for now
					aVal = getMonitorStatus(a);
					bVal = getMonitorStatus(b);
					break;
				default:
					return 0;
			}

			if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
			if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
			return 0;
		});

	function handleRowClick(monitor: any) {
		dispatch('view', monitor);
	}

	function handleEdit(monitor: any) {
		dispatch('edit', monitor);
	}

	function handleDelete(monitor: any) {
		dispatch('delete', monitor);
	}
</script>

<div class="flex flex-col gap-6">
	<!-- Toolbar -->
	<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div class="relative max-w-md w-full">
			<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-slate-400 dark:text-slate-500">
					<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
				</svg>
			</div>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search monitors..."
				class="block w-full pl-10 pr-3 py-2 border border-slate-300 dark:border-slate-600 rounded-lg leading-5 bg-white dark:bg-slate-900/50 placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition-shadow text-slate-900 dark:text-gray-100"
			/>
		</div>
		
		<div class="flex items-center gap-3">
			<select
				bind:value={sortField}
				class="block w-full rounded-lg border border-slate-300 dark:border-slate-600 py-2 pl-3 pr-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 sm:text-sm bg-white dark:bg-slate-900/50 text-slate-900 dark:text-gray-100"
			>
				<option value="name">Name</option>
				<option value="status">Status</option>
			</select>
			<button
				class="p-2 text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors"
				on:click={() => sortDirection = sortDirection === 'asc' ? 'desc' : 'asc'}
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
	</div>

	<!-- Grid -->
	{#if filteredAndSortedMonitors.length === 0}
		<div class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-12 text-center">
			<div class="flex flex-col items-center justify-center">
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-12 h-12 text-slate-300 dark:text-slate-600 mb-4">
					<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
				</svg>
				<h3 class="text-lg font-medium text-slate-900 dark:text-gray-100">No monitors found</h3>
				<p class="mt-1 text-slate-500 dark:text-slate-400">Try adjusting your search or add a new monitor.</p>
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each filteredAndSortedMonitors as monitor (monitor.id)}
				<div 
					role="button"
					tabindex="0"
					class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-5 hover:shadow-md hover:border-blue-300 dark:hover:border-blue-600 transition-all cursor-pointer group flex flex-col"
					on:click={() => handleRowClick(monitor)}
					on:keydown={(e) => e.key === 'Enter' && handleRowClick(monitor)}
				>
					<div class="flex justify-between items-start mb-4">
						<div class="flex items-center gap-3">
							<div class="p-2.5 bg-slate-100 dark:bg-slate-700/70 rounded-lg text-slate-500 dark:text-slate-400 group-hover:bg-blue-50 dark:group-hover:bg-blue-900/30 group-hover:text-blue-600 transition-colors shrink-0">
								<Favicon url={monitor.url} />
							</div>
							<div>
								<h3 class="font-semibold text-slate-900 dark:text-gray-100 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{monitor.name}</h3>
								<div class="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1 mt-0.5">
									<span class="truncate max-w-[120px]">{monitor.url}</span>
									<span class="px-1.5 py-0.5 rounded text-xs font-medium 
										{monitor.type === 'tcp' ? 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300' : 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'}">
										{monitor.type === 'tcp' ? 'TCP' : 'HTTP'}
									</span>
								</div>
							</div>
						</div>
						<MonitorStatus status={getMonitorStatus(monitor)} showText={false} />
					</div>

					<div class="mt-auto pt-4 border-t border-slate-100 dark:border-slate-700 flex justify-between items-center">
						<div class="text-xs text-slate-400 dark:text-slate-500">
							{#if monitor.last_checked_at}
								Checked {formatRelativeTime(monitor.last_checked_at)}
							{:else}
								Not checked yet
							{/if}
						</div>
						<div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
							<button
								class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:text-blue-400 dark:hover:bg-blue-900/20 rounded-md transition-colors"
								on:click|stopPropagation={() => handleEdit(monitor)}
								title="Edit"
							>
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
									<path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
								</svg>
							</button>
							<button
								class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:text-red-400 dark:hover:bg-red-900/20 rounded-md transition-colors"
								on:click|stopPropagation={() => handleDelete(monitor)}
								title="Delete"
							>
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
									<path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
								</svg>
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
