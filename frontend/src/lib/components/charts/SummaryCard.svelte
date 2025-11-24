<script lang="ts">
	export let data: { status_code: number; count: number }[] = [];

	function getColorForStatusCode(statusCode: number): string {
		if (statusCode >= 200 && statusCode < 300) return '#10B981'; // Green for success
		if (statusCode >= 300 && statusCode < 400) return '#3B82F6'; // Blue for redirects
		if (statusCode >= 400 && statusCode < 500) return '#F59E0B'; // Yellow for client errors
		if (statusCode >= 500) return '#EF4444'; // Red for server errors
		return '#6B7280'; // Gray for unknown
	}

	function getStatusDescription(statusCode: number): string {
		if (statusCode >= 200 && statusCode < 300) return 'Success';
		if (statusCode >= 300 && statusCode < 400) return 'Redirect';
		if (statusCode >= 400 && statusCode < 500) return 'Client Error';
		if (statusCode >= 500) return 'Server Error';
		return 'Unknown';
	}

	$: if (data.length === 1) {
		const item = data[0];
		const color = getColorForStatusCode(item.status_code);
		const description = getStatusDescription(item.status_code);
	}
</script>

{#if data.length === 1}
	{@const item = data[0]}
	{@const color = getColorForStatusCode(item.status_code)}
	{@const description = getStatusDescription(item.status_code)}
	<div class="flex items-center justify-center p-8 bg-gray-50 rounded-lg">
		<div class="text-center">
			<div class="inline-flex items-center justify-center w-16 h-16 rounded-full mb-4" style="background-color: {color}20;">
				<span class="text-2xl font-bold" style="color: {color};">{item.status_code}</span>
			</div>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">{description}</h3>
			<p class="text-sm text-gray-600 mb-2">Status Code: {item.status_code}</p>
			<p class="text-2xl font-bold text-gray-900">{item.count} requests</p>
		</div>
	</div>
{/if}