<script lang="ts">
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';

	let isSubmitting = false;
	let error = '';
	let form = {
		email: '',
		password: '',
		role: 'user'
	};

	async function handleSubmit() {
		if (!form.email || !form.password) {
			error = 'Email and password are required';
			return;
		}

		isSubmitting = true;
		error = '';

		try {
			const response = await fetchAPI('/api/v1/admin/users', {
				method: 'POST',
				body: JSON.stringify(form)
			});

			if (response.ok) {
				goto('/admin/users');
			} else {
				const errorData = await response.json();
				error = errorData.error || 'Failed to create user';
			}
		} catch (err: any) {
			error = err.message || 'An error occurred';
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>Create User - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-md mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Create User</h1>
			<p class="text-gray-500 dark:text-gray-400">Add a new user to the system.</p>
		</div>
		<button
			on:click={() => goto('/admin/users')}
			class="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-slate-800 hover:bg-gray-50 dark:hover:bg-slate-700"
		>
			<svg class="-ml-1 mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
			</svg>
			Back to Users
		</button>
	</div>

	<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-black ring-opacity-5 dark:ring-slate-700 rounded-lg">
		<form on:submit|preventDefault={handleSubmit} class="space-y-6 p-6">
			{#if error}
				<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg">
					{error}
				</div>
			{/if}

			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Email</label>
				<input
					type="email"
					id="email"
					bind:value={form.email}
					class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
					required
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
				<input
					type="password"
					id="password"
					bind:value={form.password}
					class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
					required
					minlength="6"
				/>
			</div>

			<div>
				<label for="role" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Role</label>
				<select
					id="role"
					bind:value={form.role}
					class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
				>
					<option value="user">User</option>
					<option value="admin">Admin</option>
				</select>
			</div>

			<div class="flex justify-end space-x-3">
				<button
					type="button"
					on:click={() => goto('/admin/users')}
					class="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-slate-800 hover:bg-gray-50 dark:hover:bg-slate-700"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={isSubmitting}
					class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
				>
					{#if isSubmitting}
						<svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
					{/if}
					Create User
				</button>
			</div>
		</form>
	</div>
</div>