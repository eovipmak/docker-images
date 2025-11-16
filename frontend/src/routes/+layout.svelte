<script lang="ts">
	import '../app.css';
	import Nav from '$lib/components/Nav.svelte';
</script>

<div class="min-h-screen bg-gray-50">
	<Nav />
	<main>
		<slot />
	</main>
</div>

