<script lang="ts">
	let settings = {
		notifications: true,
		theme: 'light'
	};
</script>

<svelte:head>
	<title>Settings - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-8 py-8">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-2xl font-semibold leading-6 text-slate-900 dark:text-white">Settings</h1>
			<p class="mt-2 text-sm text-slate-600 dark:text-slate-400">Configure your account and monitoring preferences.</p>
		</div>
	</div>

	<div class="mt-8 space-y-6">
		   <!-- Notification Settings -->
		   <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden">
			   <div class="px-4 py-5 sm:p-6">
				   <h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-white">Notifications</h3>
				   <div class="mt-2 max-w-xl text-sm text-slate-500 dark:text-slate-400">
					   <p>Manage how you receive alerts and updates.</p>
				   </div>
				   <div class="mt-6 space-y-6">
					   <div class="relative flex gap-x-3">
						   <div class="flex h-6 items-center">
							   <input
								   id="notifications"
								   name="notifications"
								   type="checkbox"
								   bind:checked={settings.notifications}
								   class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-700"
							   />
						   </div>
						   <div class="text-sm leading-6">
							   <label for="notifications" class="font-medium text-slate-900 dark:text-white">Push Notifications</label>
							   <p class="text-slate-500 dark:text-slate-400">Receive notifications in your browser when incidents occur.</p>
						   </div>
					   </div>
				   </div>
			   </div>
		   </div>

		   <!-- API Documentation -->
		   <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg overflow-hidden">
			   <div class="px-4 py-5 sm:p-6">
				   <h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-white">API Documentation</h3>
				   <div class="mt-2 max-w-xl text-sm text-slate-500 dark:text-slate-400">
					   <p>Access the complete API reference to integrate V-Insight with your applications.</p>
				   </div>
				   <div class="mt-5">
					   <a
						   href="/docs"
						   class="inline-flex items-center gap-x-2 rounded-md bg-blue-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
					   >
						   <svg class="-ml-0.5 h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
							   <path fill-rule="evenodd" d="M4.5 2A1.5 1.5 0 003 3.5v13A1.5 1.5 0 004.5 18h11a1.5 1.5 0 001.5-1.5V7.621a1.5 1.5 0 00-.44-1.06l-4.12-4.122A1.5 1.5 0 0011.378 2H4.5zm2.25 8.5a.75.75 0 000 1.5h6.5a.75.75 0 000-1.5h-6.5zm0 3a.75.75 0 000 1.5h6.5a.75.75 0 000-1.5h-6.5z" clip-rule="evenodd" />
						   </svg>
						   View API Documentation
					   </a>
				   </div>
			   </div>
		   </div>
	</div>
</div>
