<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import MonitorTable from '$lib/components/MonitorTable.svelte';
	import MonitorModal from '$lib/components/MonitorModal.svelte';

	let monitors: any[] = [];
	let isLoading = true;
	let error = '';
	let isModalOpen = false;
	let selectedMonitor: any = null;

	onMount(() => {
		loadMonitors();
	});

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

	function handleAddMonitor() {
		selectedMonitor = null;
		isModalOpen = true;
	}

	function handleEditMonitor(event: CustomEvent) {
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

			// Reload monitors after deletion
			await loadMonitors();
		} catch (err: any) {
			console.error('Error deleting monitor:', err);
			alert(err.message || 'Failed to delete monitor');
		}
	}

	function handleViewMonitor(event: CustomEvent) {
		const monitor = event.detail;
		goto(`/domains/${monitor.id}`);
	}

	function handleModalSave() {
		isModalOpen = false;
		selectedMonitor = null;
		loadMonitors();
	}

	function handleModalClose() {
		isModalOpen = false;
		selectedMonitor = null;
	}
</script>

<svelte:head>
	<title>Monitors - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-3xl font-bold text-gray-900 mb-2">Monitors</h1>
			<p class="text-gray-600">Manage and monitor your websites and services</p>
		</div>
		<button
			on:click={handleAddMonitor}
			class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors font-medium"
		>
			Add Monitor
		</button>
	</div>

	{#if isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
			{error}
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

<MonitorModal
	bind:isOpen={isModalOpen}
	monitor={selectedMonitor}
	on:save={handleModalSave}
	on:close={handleModalClose}
/>
