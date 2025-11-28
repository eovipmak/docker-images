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

<div class="px-4 sm:px-6 lg:px-8 py-8">
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
					   <div class="relative flex gap-x-3">
						   <div class="flex h-6 items-center">
							   <input
								   id="emailAlerts"
								   name="emailAlerts"
								   type="checkbox"
								   bind:checked={settings.emailAlerts}
								   class="h-4 w-4 rounded border-slate-300 dark:border-slate-600 text-blue-600 focus:ring-blue-600 dark:bg-slate-700"
							   />
						   </div>
						   <div class="text-sm leading-6">
							   <label for="emailAlerts" class="font-medium text-slate-900 dark:text-white">Email Alerts</label>
							   <p class="text-slate-500 dark:text-slate-400">Receive a daily summary of your monitoring status via email.</p>
						   </div>
					   </div>
				   </div>
			   </div>
		   </div>
	</div>
</div>
