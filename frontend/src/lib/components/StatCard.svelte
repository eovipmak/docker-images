<script lang="ts">
	export let title: string;
	export let value: string | number;
	export let valueColor: string = 'text-gray-900';
	export let bgColor: string = 'bg-white';
</script>

<div class="{bgColor} rounded-lg shadow-md p-6">
	<h3 class="text-sm font-medium text-gray-500 mb-2">{title}</h3>
	<p class="text-3xl font-bold {valueColor}">{value}</p>
</div>
