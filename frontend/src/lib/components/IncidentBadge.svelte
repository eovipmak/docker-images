<script lang="ts">
	export let status: 'open' | 'resolved';
	export let severity: 'critical' | 'warning' | 'info' = 'warning';

	const statusConfig = {
		open: {
			critical: { color: 'text-red-700', bg: 'bg-red-100', text: 'Open - Critical' },
			warning: { color: 'text-yellow-700', bg: 'bg-yellow-100', text: 'Open - Warning' },
			info: { color: 'text-blue-700', bg: 'bg-blue-100', text: 'Open - Info' }
		},
		resolved: {
			critical: { color: 'text-green-700', bg: 'bg-green-100', text: 'Resolved' },
			warning: { color: 'text-green-700', bg: 'bg-green-100', text: 'Resolved' },
			info: { color: 'text-green-700', bg: 'bg-green-100', text: 'Resolved' }
		}
	};

	const config = statusConfig[status][severity];
</script>

<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {config.bg} {config.color}">
	{config.text}
</span>
