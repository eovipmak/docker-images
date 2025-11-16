<script lang="ts">
	import { page } from '$app/stores';

	const navItems = [
		{ href: '/', label: 'Home' },
		{ href: '/login', label: 'Login' },
		{ href: '/dashboard', label: 'Dashboard' },
		{ href: '/domains', label: 'Domains' },
		{ href: '/alerts', label: 'Alerts' },
		{ href: '/settings', label: 'Settings' }
	];
</script>

<nav class="bg-blue-600 text-white shadow-lg">
	<div class="container mx-auto px-4">
		<div class="flex items-center justify-between h-16">
			<div class="flex items-center">
				<span class="text-xl font-bold">V-Insight</span>
			</div>
			<div class="flex space-x-4">
				{#each navItems as item}
					<a
						href={item.href}
						class="px-3 py-2 rounded-md text-sm font-medium transition-colors {$page.url
							.pathname === item.href
							? 'bg-blue-700'
							: 'hover:bg-blue-500'}"
					>
						{item.label}
					</a>
				{/each}
			</div>
		</div>
	</div>
</nav>
