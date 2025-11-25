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

			const queryString = params.toString();
			const url = `/api/v1/incidents${queryString ? `?${queryString}` : ''}`;

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

	function handleFilterSubmit() {
		loadIncidents();
	}

	function handleResetFilters() {
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

<div class="px-4 sm:px-6 lg:px-8 py-8">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-2xl font-semibold leading-6 text-slate-900">Incidents</h1>
			<p class="mt-2 text-sm text-slate-600">View and manage incidents detected by your monitors.</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="mt-8 bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg p-6">
		<form on:submit|preventDefault={handleFilterSubmit} class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
			<div class="sm:col-span-2">
				<label for="status" class="block text-sm font-medium leading-6 text-slate-900">Status</label>
				<div class="mt-2">
					<select
						id="status"
						bind:value={statusFilter}
						class="block w-full rounded-md border-0 py-1.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
					>
						<option value="">All Statuses</option>
						<option value="open">Open</option>
						<option value="resolved">Resolved</option>
					</select>
				</div>
			</div>

			<div class="sm:col-span-2">
				<label for="monitor" class="block text-sm font-medium leading-6 text-slate-900">Monitor</label>
				<div class="mt-2">
					<select
						id="monitor"
						bind:value={monitorFilter}
						class="block w-full rounded-md border-0 py-1.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
					>
						<option value="">All Monitors</option>
						{#each monitors as monitor}
							<option value={monitor.id}>{monitor.name}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="sm:col-span-1">
				<label for="from-date" class="block text-sm font-medium leading-6 text-slate-900">From</label>
				<div class="mt-2">
					<input
						type="date"
						id="from-date"
						bind:value={fromDate}
						class="block w-full rounded-md border-0 py-1.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
					/>
				</div>
			</div>

			<div class="sm:col-span-1">
				<label for="to-date" class="block text-sm font-medium leading-6 text-slate-900">To</label>
				<div class="mt-2">
					<input
						type="date"
						id="to-date"
						bind:value={toDate}
						class="block w-full rounded-md border-0 py-1.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:text-sm sm:leading-6"
					/>
				</div>
			</div>

			<div class="sm:col-span-6 flex justify-end gap-3 pt-2">
				<button
					type="button"
					on:click={handleResetFilters}
					class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 hover:bg-slate-50"
				>
					Reset
				</button>
				<button
					type="submit"
					class="rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
				>
					Apply Filters
				</button>
			</div>
		</form>
	</div>

	<!-- Incidents List -->
	<div class="mt-8">
		{#if isLoading}
			<div class="flex items-center justify-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div class="rounded-md bg-red-50 p-4 border border-red-200">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800">Error loading incidents</h3>
						<div class="mt-2 text-sm text-red-700">
							<p>{error}</p>
						</div>
					</div>
				</div>
			</div>
		{:else}
			<IncidentTable {incidents} />
		{/if}
	</div>
</div>
