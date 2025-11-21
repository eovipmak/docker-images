<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import IncidentTable from '$lib/components/IncidentTable.svelte';

	let incidents: any[] = [];
	let monitors: any[] = [];
	let isLoading = true;
	let error = '';

	// Filter states
	let statusFilter = ''; // '', 'open', or 'resolved'
	let monitorFilter = '';
	let fromDate = '';
	let toDate = '';

	onMount(() => {
		loadMonitors();
		loadIncidents();
	});

	async function loadMonitors() {
		try {
			const response = await fetchAPI('/api/v1/monitors');
			if (!response.ok) {
				throw new Error('Failed to load monitors');
			}
			monitors = await response.json();
		} catch (err: any) {
			console.error('Error loading monitors:', err);
			// Don't block on monitor load failure
		}
	}

	async function loadIncidents() {
		isLoading = true;
		error = '';

		try {
			// Build query string
			const params = new URLSearchParams();
			if (statusFilter) params.append('status', statusFilter);
			if (monitorFilter) params.append('monitor_id', monitorFilter);
			if (fromDate) {
				const fromDateTime = new Date(fromDate).toISOString();
				params.append('from', fromDateTime);
			}
			if (toDate) {
				const toDateTime = new Date(toDate + 'T23:59:59').toISOString();
				params.append('to', toDateTime);
			}
			params.append('limit', '100');

			const queryString = params.toString();
			const url = `/api/v1/incidents${queryString ? '?' + queryString : ''}`;

			const response = await fetchAPI(url);
			if (!response.ok) {
				throw new Error('Failed to load incidents');
			}

			incidents = await response.json();
		} catch (err: any) {
			console.error('Error loading incidents:', err);
			error = err.message || 'Failed to load incidents';
		} finally {
			isLoading = false;
		}
	}

	function handleFilterChange() {
		loadIncidents();
	}

	function handleClearFilters() {
		statusFilter = '';
		monitorFilter = '';
		fromDate = '';
		toDate = '';
		loadIncidents();
	}
</script>

<svelte:head>
	<title>Incidents - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-7xl mx-auto">
		<div class="mb-6">
			<h1 class="text-3xl font-bold text-gray-900 mb-2">Incidents History</h1>
			<p class="text-gray-600">View and manage all incident history</p>
		</div>

		<!-- Filters -->
		<div class="bg-white rounded-lg shadow-md p-4 mb-6">
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				<!-- Status filter -->
				<div>
					<label for="status-filter" class="block text-sm font-medium text-gray-700 mb-1">
						Status
					</label>
					<select
						id="status-filter"
						bind:value={statusFilter}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All</option>
						<option value="open">Open</option>
						<option value="resolved">Resolved</option>
					</select>
				</div>

				<!-- Monitor filter -->
				<div>
					<label for="monitor-filter" class="block text-sm font-medium text-gray-700 mb-1">
						Monitor
					</label>
					<select
						id="monitor-filter"
						bind:value={monitorFilter}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All Monitors</option>
						{#each monitors as monitor}
							<option value={monitor.id}>{monitor.name}</option>
						{/each}
					</select>
				</div>

				<!-- From date -->
				<div>
					<label for="from-date" class="block text-sm font-medium text-gray-700 mb-1">
						From Date
					</label>
					<input
						type="date"
						id="from-date"
						bind:value={fromDate}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<!-- To date -->
				<div>
					<label for="to-date" class="block text-sm font-medium text-gray-700 mb-1">
						To Date
					</label>
					<input
						type="date"
						id="to-date"
						bind:value={toDate}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
			</div>

			<!-- Clear filters button -->
			{#if statusFilter || monitorFilter || fromDate || toDate}
				<div class="mt-4">
					<button
						on:click={handleClearFilters}
						class="text-sm text-blue-600 hover:text-blue-800 font-medium"
					>
						Clear all filters
					</button>
				</div>
			{/if}
		</div>

		<!-- Results count -->
		<div class="mb-4">
			<p class="text-sm text-gray-600">
				{incidents.length} incident{incidents.length !== 1 ? 's' : ''} found
			</p>
		</div>

		<!-- Incidents table -->
		{#if isLoading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
				{error}
			</div>
		{:else}
			<IncidentTable {incidents} />
		{/if}
	</div>
</div>
