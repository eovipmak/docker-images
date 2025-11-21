<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import MonitorStatus from './MonitorStatus.svelte';

	export let monitors: any[] = [];

	const dispatch = createEventDispatcher();

	type SortField = 'name' | 'url' | 'status' | 'last_checked_at' | 'response_time';
	type SortDirection = 'asc' | 'desc';

	let sortField: SortField = 'name';
	let sortDirection: SortDirection = 'asc';
	let searchQuery = '';

	// Format date for display
	function formatDate(dateString?: string): string {
		if (!dateString) return 'Never';
		const date = new Date(dateString);
		return date.toLocaleString();
	}

	// Format relative time
	function formatRelativeTime(dateString?: string): string {
		if (!dateString) return 'Never';
		const date = new Date(dateString);
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

	// Get average response time (mock for now)
	function getResponseTime(monitor: any): string {
		// In a real implementation, this would come from the API
		return monitor.avg_response_time ? `${monitor.avg_response_time}ms` : 'N/A';
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
				case 'url':
					aVal = a.url.toLowerCase();
					bVal = b.url.toLowerCase();
					break;
				case 'last_checked_at':
					aVal = a.last_checked_at ? new Date(a.last_checked_at).getTime() : 0;
					bVal = b.last_checked_at ? new Date(b.last_checked_at).getTime() : 0;
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

	function handleEdit(monitor: any, event: Event) {
		event.stopPropagation();
		dispatch('edit', monitor);
	}

	function handleDelete(monitor: any, event: Event) {
		event.stopPropagation();
		dispatch('delete', monitor);
	}

	// Sort icon component
	function getSortIcon(field: SortField): string {
		if (sortField !== field) return '↕';
		return sortDirection === 'asc' ? '↑' : '↓';
	}
</script>

<div class="space-y-4">
	<!-- Search and Filter -->
	<div class="flex items-center gap-4">
		<div class="flex-1">
			<input
				type="text"
				placeholder="Search monitors..."
				bind:value={searchQuery}
				class="w-full px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
			/>
		</div>
		<div class="text-sm text-gray-600">
			{filteredAndSortedMonitors.length} monitor{filteredAndSortedMonitors.length !== 1 ? 's' : ''}
		</div>
	</div>

	<!-- Table -->
	<div class="bg-white rounded-lg shadow-md overflow-hidden">
		{#if filteredAndSortedMonitors.length === 0}
			<div class="text-center py-12 px-4">
				<p class="text-gray-500 mb-2">
					{searchQuery ? 'No monitors found matching your search' : 'No monitors configured yet'}
				</p>
				{#if !searchQuery}
					<p class="text-sm text-gray-400">Add your first monitor to start monitoring</p>
				{/if}
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th
								scope="col"
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
								on:click={() => sortMonitors('name')}
							>
								<div class="flex items-center gap-2">
									Name
									<span class="text-gray-400">{getSortIcon('name')}</span>
								</div>
							</th>
							<th
								scope="col"
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
								on:click={() => sortMonitors('url')}
							>
								<div class="flex items-center gap-2">
									URL
									<span class="text-gray-400">{getSortIcon('url')}</span>
								</div>
							</th>
							<th
								scope="col"
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Status
							</th>
							<th
								scope="col"
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
								on:click={() => sortMonitors('last_checked_at')}
							>
								<div class="flex items-center gap-2">
									Last Check
									<span class="text-gray-400">{getSortIcon('last_checked_at')}</span>
								</div>
							</th>
							<th
								scope="col"
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
							>
								Response Time
							</th>
							<th scope="col" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each filteredAndSortedMonitors as monitor (monitor.id)}
							<tr
								class="hover:bg-gray-50 cursor-pointer transition-colors"
								on:click={() => handleRowClick(monitor)}
							>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="flex items-center">
										<div>
											<div class="text-sm font-medium text-gray-900">{monitor.name}</div>
											{#if !monitor.enabled}
												<div class="text-xs text-gray-500">(Disabled)</div>
											{/if}
										</div>
									</div>
								</td>
								<td class="px-6 py-4">
									<div class="text-sm text-gray-900 truncate max-w-xs" title={monitor.url}>
										{monitor.url}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<MonitorStatus status={getMonitorStatus(monitor)} />
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm text-gray-900">
										{formatRelativeTime(monitor.last_checked_at)}
									</div>
									<div class="text-xs text-gray-500">
										{formatDate(monitor.last_checked_at)}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									{getResponseTime(monitor)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
									<button
										on:click={(e) => handleEdit(monitor, e)}
										class="text-blue-600 hover:text-blue-900 mr-3"
										title="Edit"
									>
										Edit
									</button>
									<button
										on:click={(e) => handleDelete(monitor, e)}
										class="text-red-600 hover:text-red-900"
										title="Delete"
									>
										Delete
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
