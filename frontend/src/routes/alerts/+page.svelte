<script lang="ts">
// Alerts list will be loaded here
let alerts: any[] = [];
</script>

<svelte:head>
<title>Alerts - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
<div class="max-w-7xl mx-auto">
<h1 class="text-3xl font-bold text-gray-900 mb-2">Alerts</h1>
<p class="text-gray-600 mb-8">View and manage monitoring alerts</p>

<div class="bg-white rounded-lg shadow-md p-6">
{#if alerts.length === 0}
<div class="text-center py-12">
<p class="text-gray-500 mb-4">No active alerts</p>
<p class="text-sm text-gray-400">All systems are operating normally</p>
</div>
{:else}
<div class="space-y-4">
{#each alerts as alert}
<div class="border-l-4 border-red-500 bg-red-50 p-4 rounded">
<p class="font-medium text-gray-900">{alert.message}</p>
</div>
{/each}
</div>
{/if}
</div>
</div>
</div>
