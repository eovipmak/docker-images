<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';

	let userId: number;
	let user: any = null;
	let isLoading = true;
	let error = '';
	let isEditing = false;
	let editForm = {
		email: '',
		password: '',
		role: ''
	};

	// Confirm modal
	let isConfirmOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let onConfirm: () => void;

	// Subscribe to page store to get params
	const unsubscribe = page.subscribe(($page) => {
		const id = $page.params.userid;
		if (id) {
			userId = parseInt(id);
		}
	});

	onMount(() => {
		loadUser();
	});

	async function loadUser() {
		isLoading = true;
		error = '';
		try {
			const response = await fetchAPI(`/api/v1/admin/users/${userId}`);
			if (response.ok) {
				user = await response.json();
			} else {
				error = 'Failed to load user';
			}
		} catch (err: any) {
			error = err.message || 'An error occurred';
		} finally {
			isLoading = false;
		}
	}

	function startEdit() {
		isEditing = true;
		editForm = {
			email: user.email,
			password: '',
			role: user.role
		};
	}

	function cancelEdit() {
		isEditing = false;
		editForm = { email: '', password: '', role: '' };
	}

	async function saveEdit() {
		try {
			const updateData: any = {
				email: editForm.email,
				role: editForm.role
			};
			if (editForm.password) {
				updateData.password = editForm.password;
			}

			const response = await fetchAPI(`/api/v1/admin/users/${userId}`, {
				method: 'PUT',
				body: JSON.stringify(updateData)
			});

			if (response.ok) {
				await loadUser();
				isEditing = false;
			} else {
				const errorData = await response.json();
				alert(errorData.error || 'Failed to update user');
			}
		} catch (err: any) {
			alert('An error occurred: ' + err.message);
		}
	}

	function handleDelete() {
		confirmTitle = 'Delete User';
		confirmMessage = `Are you sure you want to delete user ${user.email}? This action cannot be undone.`;
		onConfirm = async () => {
			try {
				const response = await fetchAPI(`/api/v1/admin/users/${userId}`, {
					method: 'DELETE'
				});
				if (response.ok) {
					goto('/admin/users');
				} else {
					alert('Failed to delete user');
				}
			} catch (err) {
				console.error(err);
				alert('An error occurred');
			} finally {
				isConfirmOpen = false;
			}
		};
		isConfirmOpen = true;
	}
</script>

<svelte:head>
	<title>View User - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">User Details</h1>
			<p class="text-gray-500 dark:text-gray-400">View and manage user information.</p>
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

	{#if error}
		<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
			{error}
		</div>
	{/if}

	{#if isLoading}
		<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-black ring-opacity-5 dark:ring-slate-700 rounded-lg p-6">
			<div class="animate-pulse">
				<div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-1/4 mb-4"></div>
				<div class="space-y-3">
					<div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-1/2"></div>
					<div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-1/3"></div>
					<div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-1/4"></div>
					<div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-1/6"></div>
				</div>
			</div>
		</div>
	{:else if user}
		<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-black ring-opacity-5 dark:ring-slate-700 rounded-lg overflow-hidden">
			<div class="px-4 py-5 sm:p-6">
				{#if isEditing}
					<div class="space-y-6">
						<div>
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">Edit User</h3>
							<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Update user information and password.</p>
						</div>

						<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
							<div>
								<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Email</label>
								<input
									type="email"
									id="email"
									bind:value={editForm.email}
									class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
									required
								/>
							</div>

							<div>
								<label for="role" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Role</label>
								<select
									id="role"
									bind:value={editForm.role}
									class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
								>
									<option value="user">User</option>
									<option value="admin">Admin</option>
								</select>
							</div>

							<div class="sm:col-span-2">
								<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">New Password (leave empty to keep current)</label>
								<input
									type="password"
									id="password"
									bind:value={editForm.password}
									class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-slate-700 dark:text-white sm:text-sm"
									placeholder="Enter new password"
								/>
							</div>
						</div>

						<div class="flex justify-end space-x-3">
							<button
								on:click={cancelEdit}
								class="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-200 bg-white dark:bg-slate-800 hover:bg-gray-50 dark:hover:bg-slate-700"
							>
								Cancel
							</button>
							<button
								on:click={saveEdit}
								class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
							>
								Save Changes
							</button>
						</div>
					</div>
				{:else}
					<div class="space-y-6">
						<div>
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">User Information</h3>
							<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Details about this user account.</p>
						</div>

						<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">ID</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">{user.id}</dd>
							</div>

							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Email</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">{user.email}</dd>
							</div>

							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Role</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">
									<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {user.role === 'admin' ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300' : 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'}">
										{user.role}
									</span>
								</dd>
							</div>

							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Created At</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">{new Date(user.created_at).toLocaleString()}</dd>
							</div>

							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Updated At</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">{new Date(user.updated_at).toLocaleString()}</dd>
							</div>
						</div>

						<div class="flex justify-end space-x-3">
							<button
								on:click={startEdit}
								class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
							>
								<svg class="-ml-1 mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
								</svg>
								Edit User
							</button>
							<button
								on:click={handleDelete}
								class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
							>
								<svg class="-ml-1 mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
								</svg>
								Delete User
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<ConfirmModal
	isOpen={isConfirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	on:confirm={onConfirm}
	on:cancel={() => isConfirmOpen = false}
/>