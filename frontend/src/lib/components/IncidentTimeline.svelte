<script lang="ts">
	export let incident: any;

	interface TimelineEvent {
		timestamp: string;
		title: string;
		description: string;
		type: 'start' | 'resolve' | 'notify';
	}

	$: events = buildTimeline(incident);

	function buildTimeline(incident: any): TimelineEvent[] {
		if (!incident) return [];

		const timeline: TimelineEvent[] = [
			{
				timestamp: incident.started_at,
				title: 'Incident Started',
				description: `Alert "${incident.alert_rule_name || 'Unknown'}" triggered for monitor "${incident.monitor_name || 'Unknown'}"`,
				type: 'start'
			}
		];

		if (incident.notified_at && incident.notified_at.Valid && isValidDate(incident.notified_at.Time)) {
			let description = `Alert notifications were sent at ${formatDate(incident.notified_at.Time)}`;
			if (incident.channels && incident.channels.length > 0) {
				const channelNames = (incident.channels as any[]).map((ch: any) => `${ch.name} (${ch.type})`).join(', ');
				description += ` to: ${channelNames}`;
			}
			timeline.push({
				timestamp: incident.notified_at.Time,
				title: 'Notifications Sent',
				description: description,
				type: 'notify'
			});
		}

		if (incident.resolved_at && incident.resolved_at.Valid && isValidDate(incident.resolved_at.Time)) {
			timeline.push({
				timestamp: incident.resolved_at.Time,
				title: 'Incident Resolved',
				description: 'The incident was automatically resolved',
				type: 'resolve'
			});
		}

		return timeline.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
	}

	function isValidDate(dateString: string): boolean {
		const date = new Date(dateString);
		return date instanceof Date && !isNaN(date.getTime());
	}

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
</script>

<div class="bg-white shadow-sm ring-1 ring-slate-900/5 sm:rounded-lg overflow-hidden">
	<div class="px-4 py-5 sm:p-6">
		<h3 class="text-base font-semibold leading-6 text-slate-900 mb-6">Incident Timeline</h3>
		
		{#if events.length === 0}
			<p class="text-sm text-slate-500">No timeline events available</p>
		{:else}
			<div class="flow-root">
				<ul role="list" class="-mb-8">
					{#each events as event, index (index)}
						<li>
							<div class="relative pb-8">
								{#if index !== events.length - 1}
									<span class="absolute left-4 top-4 -ml-px h-full w-0.5 bg-slate-200" aria-hidden="true"></span>
								{/if}
								<div class="relative flex space-x-3">
									<div>
										{#if event.type === 'start'}
											<span class="h-8 w-8 rounded-full bg-red-100 flex items-center justify-center ring-8 ring-white">
												<svg class="h-5 w-5 text-red-600" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
													<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd" />
												</svg>
											</span>
										{:else if event.type === 'notify'}
											<span class="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center ring-8 ring-white">
												<svg class="h-5 w-5 text-blue-600" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
													<path d="M3 4a2 2 0 00-2 2v1.161l8.441 4.221a1.25 1.25 0 001.118 0L19 7.162V6a2 2 0 00-2-2H3z" />
													<path d="M19 8.839l-7.77 3.885a2.75 2.75 0 01-2.46 0L1 8.839V14a2 2 0 002 2h14a2 2 0 002-2V8.839z" />
												</svg>
											</span>
										{:else if event.type === 'resolve'}
											<span class="h-8 w-8 rounded-full bg-green-100 flex items-center justify-center ring-8 ring-white">
												<svg class="h-5 w-5 text-green-600" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
													<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clip-rule="evenodd" />
												</svg>
											</span>
										{/if}
									</div>
									<div class="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
										<div>
											<p class="text-sm text-slate-500">
												<span class="font-medium text-slate-900">{event.title}</span>
												<span class="block mt-1">{event.description}</span>
											</p>
										</div>
										<div class="whitespace-nowrap text-right text-sm text-slate-500">
											<time datetime={event.timestamp}>{formatDate(event.timestamp)}</time>
										</div>
									</div>
								</div>
							</div>
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</div>
</div>
