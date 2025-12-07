<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';
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

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">Incidents</h1>
			<p class="mt-1 text-sm text-slate-600 dark:text-slate-400">View and manage incidents detected by your monitors.</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-6 shadow-none">
		<form on:submit|preventDefault={handleFilterSubmit} class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
			<div class="sm:col-span-2">
				<label for="status" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Status</label>
				<div class="mt-2">
					<select
						id="status"
						bind:value={statusFilter}
						class="block w-full rounded-lg border border-slate-500 dark:border-slate-600 bg-white dark:bg-slate-800 py-2 px-3 text-sm text-slate-900 dark:text-white focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All Statuses</option>
						<option value="open">Open</option>
						<option value="resolved">Resolved</option>
					</select>
				</div>
			</div>

			<div class="sm:col-span-2">
				<label for="monitor" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Monitor</label>
				<div class="mt-2">
					<select
						id="monitor"
						bind:value={monitorFilter}
						class="block w-full rounded-lg border border-slate-500 dark:border-slate-600 bg-white dark:bg-slate-800 py-2 px-3 text-sm text-slate-900 dark:text-white focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All Monitors</option>
						{#each monitors as monitor}
							<option value={monitor.id}>{monitor.name}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="sm:col-span-1">
				<label for="from-date" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">From</label>
				<div class="mt-2">
					<input
						type="date"
						id="from-date"
						bind:value={fromDate}
						class="block w-full rounded-lg border border-slate-500 dark:border-slate-600 bg-white dark:bg-slate-800 py-2 px-3 text-sm text-slate-900 dark:text-white focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
			</div>

			<div class="sm:col-span-1">
				<label for="to-date" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">To</label>
				<div class="mt-2">
					<input
						type="date"
						id="to-date"
						bind:value={toDate}
						class="block w-full rounded-lg border border-slate-500 dark:border-slate-600 bg-white dark:bg-slate-800 py-2 px-3 text-sm text-slate-900 dark:text-white focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>
			</div>

			<div class="sm:col-span-6 flex justify-end gap-3 pt-2">
				<button
					type="button"
					on:click={handleResetFilters}
					class="inline-flex items-center justify-center px-4 py-2 border border-slate-200 dark:border-slate-700 text-sm font-medium rounded-lg text-slate-700 dark:text-slate-200 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
				>
					Reset
				</button>
				<button
					type="submit"
					class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
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
			<div class="rounded-md bg-red-50 dark:bg-red-900/30 p-4 border border-red-200 dark:border-red-800">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error loading incidents</h3>
						<div class="mt-2 text-sm text-red-700 dark:text-red-300">
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
