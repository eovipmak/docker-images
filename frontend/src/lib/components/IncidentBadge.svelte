<script lang="ts">
	export let status: 'open' | 'resolved';
	export let severity: 'critical' | 'warning' | 'info' = 'warning';

	const statusConfig = {
		open: {
			critical: { color: 'text-rose-700', bg: 'bg-rose-50 border-rose-200', text: 'Critical' },
			warning: { color: 'text-amber-700', bg: 'bg-amber-50 border-amber-200', text: 'Warning' },
			info: { color: 'text-sky-700', bg: 'bg-sky-50 border-sky-200', text: 'Info' }
		},
		resolved: {
			critical: { color: 'text-emerald-700', bg: 'bg-emerald-50 border-emerald-200', text: 'Resolved' },
			warning: { color: 'text-emerald-700', bg: 'bg-emerald-50 border-emerald-200', text: 'Resolved' },
			info: { color: 'text-emerald-700', bg: 'bg-emerald-50 border-emerald-200', text: 'Resolved' }
		}
	};

	const config = statusConfig[status][severity];
</script>

<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border {config.bg} {config.color}">
	{config.text}
</span>
