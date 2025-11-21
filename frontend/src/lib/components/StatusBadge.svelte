<script lang="ts">
	export let status: 'open' | 'resolved';
	export let size: 'sm' | 'md' = 'md';

	const statusConfig = {
		open: {
			color: 'text-red-700',
			bg: 'bg-red-100',
			text: 'Open'
		},
		resolved: {
			color: 'text-green-700',
			bg: 'bg-green-100',
			text: 'Resolved'
		}
	};

	const sizeConfig = {
		sm: 'px-2 py-0.5 text-xs',
		md: 'px-2.5 py-0.5 text-sm'
	};

	const config = statusConfig[status];
	const sizeClass = sizeConfig[size];
</script>

<span class="inline-flex items-center rounded-full font-medium {config.bg} {config.color} {sizeClass}">
	{config.text}
</span>
