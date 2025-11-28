<script lang="ts">
	export let className = '';
	export let headerClass = '';
	export let bodyClass = '';
	// allow forwarding arbitrary attributes (e.g. data-testid)
	// NOTE: do not `declare` $$restProps in TS context to avoid redeclaration issues.
</script>

<div {...$$restProps} class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 overflow-hidden">
	{#if $$slots.header}
		<div class={`px-6 py-4 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-950/40 ${headerClass}`}>
			<slot name="header"></slot>
		</div>
	{/if}

	<div class={`p-4 ${bodyClass} ${className}`}>
		<slot />
	</div>
</div>
