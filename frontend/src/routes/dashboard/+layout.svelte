<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { connectEventStream, disconnectEventStream } from '$lib/api/events';

	// Connect to SSE stream when dashboard is mounted
	onMount(async () => {
		await connectEventStream();
	});

	// Disconnect when dashboard is unmounted
	onDestroy(() => {
		disconnectEventStream();
	});
</script>

<div class="container mx-auto px-4 py-8">
	<slot />
</div>
