<script lang="ts">
// Dashboard data will be loaded here
let stats = {
totalDomains: 0,
activeAlerts: 0,
uptime: '0%'
};
</script>

<svelte:head>
<title>Dashboard - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto">
<h1 class="text-3xl font-bold text-gray-900 mb-6">Dashboard</h1>
<p class="text-gray-600 mb-8">Monitor your domains and view system metrics</p>

<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
<div class="bg-white rounded-lg shadow-md p-6">
<h3 class="text-sm font-medium text-gray-500 mb-2">Total Domains</h3>
<p class="text-3xl font-bold text-gray-900">{stats.totalDomains}</p>
</div>

<div class="bg-white rounded-lg shadow-md p-6">
<h3 class="text-sm font-medium text-gray-500 mb-2">Active Alerts</h3>
<p class="text-3xl font-bold text-red-600">{stats.activeAlerts}</p>
</div>

<div class="bg-white rounded-lg shadow-md p-6">
<h3 class="text-sm font-medium text-gray-500 mb-2">Uptime</h3>
<p class="text-3xl font-bold text-green-600">{stats.uptime}</p>
</div>
</div>

<div class="bg-white rounded-lg shadow-md p-6">
<h2 class="text-xl font-semibold text-gray-800 mb-4">Recent Activity</h2>
<p class="text-gray-500">No recent activity to display</p>
</div>
</div>
