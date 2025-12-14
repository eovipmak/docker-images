<script lang="ts">
	import MonitorStatus from './MonitorStatus.svelte';
	import Favicon from './Favicon.svelte';
	import Card from './Card.svelte';
	import type { Monitor } from '$lib/types';
	import { createEventDispatcher } from 'svelte';

	export let monitor: Monitor;

	const dispatch = createEventDispatcher();

	// Get monitor type display label
	function getTypeLabel(type: string): string {
		switch (type) {
			case 'tcp': return 'TCP';
			case 'icmp': 
			case 'ping': return 'ICMP';
			default: return 'HTTP';
		}
	}

	// Get type badge color classes
	function getTypeBadgeClasses(type: string): string {
		switch (type) {
			case 'tcp': return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300';
			case 'icmp':
			case 'ping': return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300';
			default: return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300';
		}
	}

	// Get type indicator classes for the small badge
	function getTypeIndicatorClasses(type: string): string {
		switch (type) {
			case 'tcp': return 'bg-orange-500 text-white';
			case 'icmp':
			case 'ping': return 'bg-purple-500 text-white';
			default: return 'bg-blue-500 text-white';
		}
	}

	// Get type indicator letter
	function getTypeIndicator(type: string): string {
		switch (type) {
			case 'tcp': return 'T';
			case 'icmp':
			case 'ping': return 'I';
			default: return 'H';
		}
	}

	function handleClick() {
		console.debug('[MonitorCard] clicked', monitor?.id);
		dispatch('view', monitor);
	}

	function handleEdit(e: Event) {
		e.stopPropagation();
		dispatch('edit', monitor);
	}

	function handleDelete(e: Event) {
		e.stopPropagation();
		dispatch('delete', monitor);
	}
</script>

	<!-- Make the entire card a link so clicking navigates to the monitor detail page.
		 Keep the existing dispatched 'view' event for any parent component handlers. -->
	<a data-testid="monitor-card" href={`/user/monitors/${monitor?.id || ''}`} class="block">
	<Card className="cursor-pointer flex flex-col h-full" on:click={handleClick}>
		<div class="flex items-start justify-between">
			<div class="flex items-center gap-3">
				<div class="p-2.5 bg-slate-100 dark:bg-slate-700/60 rounded-lg text-slate-500 dark:text-slate-400 relative">
					<Favicon url={monitor.url} type={monitor.type} />
					<!-- Monitor type indicator -->
					<div class="absolute -top-1 -right-1 w-3 h-3 rounded-full flex items-center justify-center text-xs font-bold {getTypeIndicatorClasses(monitor.type)}">
						{getTypeIndicator(monitor.type)}
					</div>
				</div>
				<div class="min-w-0">
					<h3 class="text-sm font-semibold text-slate-900 dark:text-gray-100 truncate">{monitor.name}</h3>
					<div class="flex items-center gap-1">
						<p class="text-xs text-slate-500 dark:text-slate-400 truncate">{monitor.url}</p>
						<span class="px-1.5 py-0.5 rounded text-xs font-medium {getTypeBadgeClasses(monitor.type)}">
							{getTypeLabel(monitor.type)}
						</span>
					</div>
					{#if monitor.tags && monitor.tags.length > 0}
						<div class="flex flex-wrap gap-1 mt-1">
							{#each monitor.tags.slice(0, 3) as tag}
								<span class="px-1.5 py-0.5 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded">{tag}</span>
							{/each}
							{#if monitor.tags.length > 3}
								<span class="px-1.5 py-0.5 text-xs bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 rounded">+{monitor.tags.length - 3}</span>
							{/if}
						</div>
					{/if}
				</div>
			</div>		<MonitorStatus status={(monitor.status || (monitor.enabled ? 'up' : 'unknown'))} showText={false} />
	</div>

	<div class="mt-auto pt-3 border-t border-slate-100 dark:border-slate-700 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
		<div>
			{#if monitor.last_checked_at}
				<span>Checked {new Date(monitor.last_checked_at).toLocaleString()}</span>
			{:else}
				<span>Not checked yet</span>
			{/if}
		</div>
		<div class="flex gap-2 items-center">
			{#if monitor.last_check?.response_time_ms}
				<span class="text-slate-400 dark:text-slate-500">{Math.round(monitor.last_check.response_time_ms)}ms</span>
			{/if}
			<div class="flex gap-2 items-center opacity-0 group-hover:opacity-100 group-focus-within:opacity-100 transition-opacity">
				<button
					class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:bg-blue-50 rounded-md transition-colors outline-none"
					on:click|preventDefault|stopPropagation={handleEdit}
					title="Edit"
					aria-label="Edit"
				>
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
						<path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
					</svg>
				</button>
				<button
					class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:bg-red-50 rounded-md transition-colors outline-none"
					on:click|preventDefault|stopPropagation={handleDelete}
					title="Delete"
					aria-label="Delete"
				>
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
						<path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
					</svg>
				</button>
			</div>
		</div>
	</div>
	</Card>
	</a>
