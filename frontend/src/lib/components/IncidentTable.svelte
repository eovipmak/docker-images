<script lang="ts">
	import { goto } from '$app/navigation';
	import IncidentBadge from './IncidentBadge.svelte';

	export let incidents: any[] = [];

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
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

<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
	{#if incidents.length === 0}
		<div class="text-center py-12 px-4">
			<svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			<h3 class="mt-2 text-sm font-semibold text-slate-900">No incidents found</h3>
			<p class="mt-1 text-sm text-slate-500">Incidents will appear here when alerts are triggered.</p>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-slate-200">
				<thead class="bg-slate-50">
					<tr>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Monitor</th>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Alert Rule</th>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Started At</th>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Resolved At</th>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Duration</th>
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-slate-200">
					{#each incidents as incident (incident.id)}
						<tr
							class="hover:bg-slate-50 transition-colors cursor-pointer"
							on:click={() => handleRowClick(incident.id)}
							role="button"
							tabindex="0"
							on:keydown={(e) => e.key === 'Enter' && handleRowClick(incident.id)}
						>
							<td class="px-6 py-4">
								<div class="text-sm font-medium text-slate-900">{incident.monitor_name || 'Unknown'}</div>
								<div class="text-sm text-slate-500 truncate max-w-xs">{incident.monitor_url || ''}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-slate-900">{incident.alert_rule_name || 'Unknown'}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-slate-600">{formatDate(incident.started_at)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-slate-600">
									{incident.resolved_at && incident.resolved_at.Valid ? formatDate(incident.resolved_at.Time) : 'Ongoing'}
								</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm text-slate-600">{formatDuration(incident.duration)}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<IncidentBadge status={incident.status} severity={incident.severity || 'warning'} />
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
