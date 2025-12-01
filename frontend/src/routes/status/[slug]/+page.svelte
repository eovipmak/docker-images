<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<svelte:head>
	<title>{data.statusPage.name} - Status</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-slate-950">
	<div class="max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
		<div class="text-center mb-12">
			<h1 class="text-4xl font-bold text-gray-900 dark:text-white mb-4">
				{data.statusPage.name}
			</h1>
			<p class="text-lg text-gray-600 dark:text-gray-400">
				Service Status Overview
			</p>
		</div>

		<div class="bg-white dark:bg-slate-900 shadow rounded-lg overflow-hidden">
			<div class="px-4 py-5 sm:p-6">
				<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-6">
					Monitor Status
				</h2>

				{#if data.monitors.length === 0}
					<p class="text-gray-500 dark:text-gray-400">No monitors configured.</p>
				{:else}
					<div class="space-y-4">
						{#each data.monitors as monitor}
							<div class="flex items-center justify-between p-4 border border-gray-200 dark:border-slate-700 rounded-lg">
								<div class="flex-1">
									<h3 class="text-sm font-medium text-gray-900 dark:text-white">
										{monitor.name}
									</h3>
									<p class="text-sm text-gray-500 dark:text-gray-400">
										{monitor.url} ({monitor.type})
									</p>
									{#if !monitor.enabled}
										<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">
											Monitor disabled
										</p>
									{/if}
								</div>
								<div class="ml-4">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {monitor.status === 'up' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : monitor.status === 'down' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'}">
										{monitor.status === 'up' ? 'Operational' : monitor.status === 'down' ? 'Down' : 'Unknown'}
									</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<div class="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
			<p>Status page powered by V-Insight</p>
		</div>
	</div>
</div>