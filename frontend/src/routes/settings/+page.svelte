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
			<h1 class="text-2xl font-semibold leading-6 text-slate-900">Settings</h1>
			<p class="mt-2 text-sm text-slate-600">Configure your account and monitoring preferences.</p>
		</div>
	</div>

	<div class="mt-8 space-y-6">
		<!-- Notification Settings -->
		<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
			<div class="px-4 py-5 sm:p-6">
				<h3 class="text-base font-semibold leading-6 text-slate-900">Notifications</h3>
				<div class="mt-2 max-w-xl text-sm text-slate-500">
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
								class="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
							/>
						</div>
						<div class="text-sm leading-6">
							<label for="notifications" class="font-medium text-slate-900">Push Notifications</label>
							<p class="text-slate-500">Receive notifications in your browser when incidents occur.</p>
						</div>
					</div>
					<div class="relative flex gap-x-3">
						<div class="flex h-6 items-center">
							<input
								id="emailAlerts"
								name="emailAlerts"
								type="checkbox"
								bind:checked={settings.emailAlerts}
								class="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
							/>
						</div>
						<div class="text-sm leading-6">
							<label for="emailAlerts" class="font-medium text-slate-900">Email Alerts</label>
							<p class="text-slate-500">Receive a daily summary of your monitoring status via email.</p>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Appearance Settings -->
		<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
			<div class="px-4 py-5 sm:p-6">
				<h3 class="text-base font-semibold leading-6 text-slate-900">Appearance</h3>
				<div class="mt-2 max-w-xl text-sm text-slate-500">
					<p>Customize the look and feel of the dashboard.</p>
				</div>
				<div class="mt-6">
					<div class="flex items-center gap-x-3">
						<label for="theme" class="block text-sm font-medium leading-6 text-slate-900">Theme</label>
						<select
							id="theme"
							name="theme"
							bind:value={settings.theme}
							class="block w-full rounded-md border-0 py-1.5 text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 focus:ring-2 focus:ring-inset focus:ring-blue-600 sm:max-w-xs sm:text-sm sm:leading-6"
						>
							<option value="light">Light</option>
							<option value="dark">Dark</option>
							<option value="system">System</option>
						</select>
					</div>
				</div>
			</div>
		</div>

		<div class="flex justify-end">
			<button
				type="button"
				class="rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
			>
				Save Changes
			</button>
		</div>
	</div>
</div>
