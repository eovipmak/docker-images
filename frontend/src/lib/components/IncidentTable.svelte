<script lang="ts">
	import { goto } from '$app/navigation';
	import StatusBadge from './StatusBadge.svelte';

	export let incidents: any[] = [];

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString();
	}

	function formatDuration(seconds: number | null): string {
		if (!seconds) return 'N/A';
		
		const days = Math.floor(seconds / 86400);
		const hours = Math.floor((seconds % 86400) / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);

		if (days > 0) {
			return `${days}d ${hours}h ${minutes}m`;
		} else if (hours > 0) {
			return `${hours}h ${minutes}m`;
		} else if (minutes > 0) {
			return `${minutes}m`;
		} else {
			return `${Math.floor(seconds)}s`;
		}
	}

	function handleRowClick(id: string) {
		goto(`/incidents/${id}`);
	}
</script>

<div class="bg-white rounded-lg shadow-md overflow-hidden">
	{#if incidents.length === 0}
		<div class="text-center py-12 px-4">
			<p class="text-gray-500 mb-2">No incidents found</p>
			<p class="text-sm text-gray-400">Incidents will appear here when alerts are triggered</p>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Monitor
						</th>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Alert Rule
						</th>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Started At
						</th>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Resolved At
						</th>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Duration
						</th>
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Status
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each incidents as incident (incident.id)}
						<tr
							class="hover:bg-gray-50 transition-colors cursor-pointer"
							on:click={() => handleRowClick(incident.id)}
							role="button"
							tabindex="0"
							on:keydown={(e) => e.key === 'Enter' && handleRowClick(incident.id)}
						>
							<td class="px-6 py-4">
								<div class="text-sm font-medium text-gray-900">{incident.monitor_name || 'Unknown'}</div>
								<div class="text-sm text-gray-500 truncate max-w-xs">{incident.monitor_url || ''}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-900">{incident.alert_rule_name || 'Unknown'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-900">{formatDate(incident.started_at)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-900">
									{incident.resolved_at && incident.resolved_at.Valid ? formatDate(incident.resolved_at.Time) : 'Ongoing'}
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-gray-900">{formatDuration(incident.duration)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<StatusBadge status={incident.status} size="sm" />
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
