<script lang="ts">
let settings = {
notifications: true,
emailAlerts: false,
theme: 'light'
};
</script>

<svelte:head>
<title>Settings - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
<div class="max-w-4xl mx-auto">
<h1 class="text-3xl font-bold text-gray-900 mb-2">Settings</h1>
<p class="text-gray-600 mb-8">Configure your account and monitoring preferences</p>

<div class="bg-white rounded-lg shadow-md p-6 mb-6">
<h2 class="text-xl font-semibold text-gray-800 mb-4">Notification Settings</h2>

<div class="space-y-4">
<div class="flex items-center justify-between">
<div>
<p class="font-medium text-gray-700">Push Notifications</p>
<p class="text-sm text-gray-500">Receive notifications in your browser</p>
</div>
<input
type="checkbox"
bind:checked={settings.notifications}
class="h-5 w-5 text-blue-600"
/>
</div>

<div class="flex items-center justify-between">
<div>
<p class="font-medium text-gray-700">Email Alerts</p>
<p class="text-sm text-gray-500">Receive alerts via email</p>
</div>
<input
type="checkbox"
bind:checked={settings.emailAlerts}
class="h-5 w-5 text-blue-600"
/>
</div>
</div>
</div>

<div class="bg-white rounded-lg shadow-md p-6">
<h2 class="text-xl font-semibold text-gray-800 mb-4">Appearance</h2>

<div>
<label for="theme" class="block text-sm font-medium text-gray-700 mb-2">
Theme
</label>
<select
id="theme"
bind:value={settings.theme}
class="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
>
<option value="light">Light</option>
<option value="dark">Dark</option>
<option value="auto">Auto</option>
</select>
</div>
</div>

<div class="mt-6">
<button
class="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 transition-colors font-medium"
>
Save Settings
</button>
</div>
</div>
</div>
