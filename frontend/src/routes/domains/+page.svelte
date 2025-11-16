<script lang="ts">
// Domain list will be loaded here
let domains: any[] = [];
</script>

<svelte:head>
<title>Domains - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
<div class="max-w-7xl mx-auto">
<div class="flex justify-between items-center mb-6">
<div>
<h1 class="text-3xl font-bold text-gray-900 mb-2">Domains</h1>
<p class="text-gray-600">Manage and monitor your domains</p>
</div>
<button
class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors font-medium"
>
Add Domain
</button>
</div>

<div class="bg-white rounded-lg shadow-md p-6">
{#if domains.length === 0}
<div class="text-center py-12">
<p class="text-gray-500 mb-4">No domains configured yet</p>
<p class="text-sm text-gray-400">Add your first domain to start monitoring</p>
</div>
{:else}
<div class="space-y-4">
{#each domains as domain}
<div class="border border-gray-200 rounded-lg p-4">
<p class="font-medium text-gray-900">{domain.name}</p>
</div>
{/each}
</div>
{/if}
</div>
</div>
</div>
