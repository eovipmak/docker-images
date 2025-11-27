<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Card from './Card.svelte';

	export let rule: any;

	const dispatch = createEventDispatcher();

	function handleEdit() {
		dispatch('edit', rule);
	}

	function handleDelete() {
		dispatch('delete', rule);
	}
</script>

<Card className="cursor-pointer flex flex-col h-full" data-testid="alert-card">
	<div class="flex items-start justify-between gap-3">
		<div class="min-w-0">
			<h3 class="text-sm font-semibold text-slate-900 truncate">{rule.name}</h3>
			<p class="text-xs text-slate-500 truncate">{rule.trigger_type} • Target: {rule.monitor_id ? 'Monitor' : 'All'}</p>
		</div>
		<div class="text-xs text-slate-500">
			{rule.enabled ? 'Enabled' : 'Disabled'}
		</div>
	</div>

	<div class="mt-auto pt-3 border-t border-slate-100 flex items-center justify-between">
		<div class="text-xs text-slate-500">Threshold: {String(rule.threshold_value)}</div>
		<div class="flex items-center gap-2">
			<button on:click={handleEdit} class="text-blue-600 hover:text-blue-900 text-sm">Edit</button>
			<button on:click={handleDelete} class="text-red-600 hover:text-red-900 text-sm">Delete</button>
		</div>
	</div>
	</div>
</Card>
