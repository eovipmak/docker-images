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
			let description = 'Alert notifications were sent';
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
		return date.toLocaleString();
	}

	function getTimelineIcon(type: string): string {
		switch (type) {
			case 'start':
				return '⚠️';
			case 'notify':
				return '📧';
			case 'resolve':
				return '✅';
			default:
				return '•';
		}
	}

	function getTimelineColor(type: string): string {
		switch (type) {
			case 'start':
				return 'border-red-500 bg-red-50';
			case 'notify':
				return 'border-blue-500 bg-blue-50';
			case 'resolve':
				return 'border-green-500 bg-green-50';
			default:
				return 'border-gray-500 bg-gray-50';
		}
	}
</script>

<div class="bg-white rounded-lg shadow-md p-6">
	<h3 class="text-lg font-semibold text-gray-900 mb-4">Incident Timeline</h3>
	
	{#if events.length === 0}
		<p class="text-gray-500">No timeline events available</p>
	{:else}
		<div class="relative">
			<!-- Vertical line -->
			<div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200"></div>
			
			<!-- Timeline events -->
			<div class="space-y-6">
				{#each events as event, index (index)}
					<div class="relative flex items-start gap-4">
						<!-- Icon -->
						<div class="flex-shrink-0 w-8 h-8 rounded-full border-2 {getTimelineColor(event.type)} flex items-center justify-center text-base z-10">
							{getTimelineIcon(event.type)}
						</div>
						
						<!-- Content -->
						<div class="flex-1 pt-0.5">
							<div class="flex items-center justify-between mb-1">
								<h4 class="text-sm font-medium text-gray-900">{event.title}</h4>
								<span class="text-xs text-gray-500">{formatDate(event.timestamp)}</span>
							</div>
							<p class="text-sm text-gray-600">{event.description}</p>
							{#if event.type === 'start' && incident.trigger_value}
								<p class="text-xs text-gray-500 mt-1">Trigger value: {incident.trigger_value}</p>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
