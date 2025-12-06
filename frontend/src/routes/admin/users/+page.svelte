<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fetchAPI } from '$lib/api/client';
    import { authStore } from '$lib/stores/auth';
    import ConfirmModal from '$lib/components/ConfirmModal.svelte';

    let users: any[] = [];
    let isLoading = true;
    let error = '';
    let isDeleting = false;

    // Confirm modal
    let isConfirmOpen = false;
    let confirmTitle = '';
    let confirmMessage = '';
    let onConfirm: () => void;

    onMount(() => {
        loadUsers();
    });

    async function loadUsers() {
        isLoading = true;
        error = '';
        try {
            const response = await fetchAPI('/api/v1/admin/users');
            if (response.ok) {
                users = await response.json();
            } else {
                error = 'Failed to load users';
            }
        } catch (err: any) {
            error = err.message || 'An error occurred';
        } finally {
            isLoading = false;
        }
    }

    function handleDelete(user: any) {
        confirmTitle = 'Delete User';
        confirmMessage = `Are you sure you want to delete user ${user.email}? This action cannot be undone.`;
        onConfirm = async () => {
            isDeleting = true;
            try {
                const response = await fetchAPI(`/api/v1/admin/users/${user.id}`, {
                    method: 'DELETE'
                });
                if (response.ok) {
                    await loadUsers();
                } else {
                    alert('Failed to delete user');
                }
            } catch (err) {
                console.error(err);
                alert('An error occurred');
            } finally {
                isDeleting = false;
                isConfirmOpen = false;
            }
        };
        isConfirmOpen = true;
    }
</script>

<svelte:head>
	<title>Manage Users - Admin - V-Insight</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="flex justify-between items-center mb-6">
        <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">User Management</h1>
            <p class="text-gray-500 dark:text-gray-400">View and manage system users.</p>
        </div>
        <div class="flex space-x-3">
            <button on:click={loadUsers} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
            </button>
            <button
                on:click={() => goto('/admin/create-user')}
                class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
                <svg class="-ml-1 mr-2 h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
                </svg>
                Create User
            </button>
        </div>
    </div>

    {#if error}
        <div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg mb-6">
            {error}
        </div>
    {/if}

    <div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-black ring-opacity-5 dark:ring-slate-700 rounded-lg overflow-hidden">
        <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-300 dark:divide-slate-700">
                <thead class="bg-gray-50 dark:bg-slate-900">
                    <tr>
                        <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white sm:pl-6">ID</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Email</th>
                        <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Role</th>
                         <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Created At</th>
                        <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                            <span class="sr-only">Actions</span>
                        </th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-slate-700 bg-white dark:bg-slate-800">
                    {#if isLoading}
                        {#each Array(5) as _}
                            <tr>
                                <td class="py-4 pl-4 pr-3 sm:pl-6"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-8 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-32 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-16 animate-pulse"></div></td>
                                <td class="px-3 py-4"><div class="h-4 bg-gray-200 dark:bg-slate-700 rounded w-24 animate-pulse"></div></td>
                                <td class="relative py-4 pl-3 pr-4 sm:pr-6"></td>
                            </tr>
                        {/each}
                    {:else if users.length === 0}
                         <tr>
                            <td colspan="5" class="px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400">No users found</td>
                        </tr>
                    {:else}
                        {#each users as user}
                            <tr class="cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-700" on:click={() => goto(`/admin/view-user/${user.id}`)}>
                                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">{user.id}</td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">{user.email}</td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">
                                    <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {user.role === 'admin' ? 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300' : 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'}">
                                        {user.role}
                                    </span>
                                </td>
                                <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-300">{new Date(user.created_at).toLocaleDateString()}</td>
                                <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                                    {#if user.id !== $authStore.currentUser?.id}
                                        <button 
                                            on:click|stopPropagation={() => handleDelete(user)}
                                            class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 disabled:opacity-50"
                                            disabled={isDeleting}
                                        >
                                            Delete<span class="sr-only">, {user.email}</span>
                                        </button>
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                    {/if}
                </tbody>
            </table>
        </div>
    </div>
</div>

<ConfirmModal 
    isOpen={isConfirmOpen}
    title={confirmTitle}
    message={confirmMessage}
    on:confirm={onConfirm}
    on:cancel={() => isConfirmOpen = false}
/>
