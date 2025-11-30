<script lang="ts">
	export let url: string;
	export let className: string = "w-6 h-6";
	export let type: string = 'http';

	let hasError = false;
	let domain = '';

	$: {
		try {
			domain = new URL(url).hostname;
			// Reset error when url changes, but only if the domain actually changed to avoid loops if something weird happens
			// simpler: just reset hasError.
			hasError = false; 
		} catch {
			domain = '';
		}
	}

	$: faviconUrl = domain ? `https://t1.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=https://${domain}&size=64` : null;
</script>

{#if faviconUrl && !hasError}
	<img 
		src={faviconUrl} 
		alt="favicon" 
		class="{className} object-contain"
		on:error={() => hasError = true}
	/>
{:else}
	<div class="{className} flex items-center justify-center text-slate-400">
		{#if type === 'tcp'}
			<span class="text-xs font-bold text-orange-500">TCP</span>
		{:else}
			<slot>
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-full h-full">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" />
				</svg>
			</slot>
		{/if}
	</div>
{/if}