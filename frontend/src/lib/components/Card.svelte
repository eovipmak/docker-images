<script lang="ts">
	export let className = '';
	export let headerClass = '';
	export let bodyClass = '';
</script>

<div class={`bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden ${className}`}>
	{#if $$slots.header}
		<div class={`px-6 py-4 border-b border-slate-100 bg-slate-50/50 ${headerClass}`}>
			<slot name="header"></slot>
		</div>
	{/if}

	<div class={`p-4 ${bodyClass}`}>
		<slot />
	</div>
</div>
