<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchAPI } from '$lib/api/client';
	import Card from '$lib/components/Card.svelte';

	interface StatusPage {
		id: string;
		tenant_id: number;
		slug: string;
		name: string;
		public_enabled: boolean;
		created_at: string;
		updated_at: string;
		monitor_count?: number;
	}

	let statusPages: StatusPage[] = [];
	let isLoading = true;
	let error = '';
	let isModalOpen = false;
	let selectedStatusPage: StatusPage | null = null;

	// Lazy loaded modal component
	let StatusPageModal: any = null;
	let modalLoaded = false;

	// Toast notification
	let showToast = false;
	let toastMessage = '';

	async function loadStatusPages() {
		try {
			isLoading = true;
			error = '';
			const response = await fetchAPI('/api/v1/status-pages');
			if (!response.ok) {
				throw new Error('Failed to load status pages');
			}
			const data = await response.json();
			const statusPagesData = data.status_pages || [];

			// Load monitor counts for each status page
			const statusPagesWithCounts = await Promise.all(
				statusPagesData.map(async (statusPage: StatusPage) => {
					try {
						const monitorResponse = await fetchAPI(`/api/v1/status-pages/${statusPage.id}/monitors`);
						if (monitorResponse.ok) {
							const monitorData = await monitorResponse.json();
							return { ...statusPage, monitor_count: (monitorData.monitors || []).length };
						}
					} catch (err) {
						console.error(`Failed to load monitors for status page ${statusPage.id}:`, err);
					}
					return { ...statusPage, monitor_count: 0 };
				})
			);

			statusPages = statusPagesWithCounts;
		} catch (err: any) {
			error = err.message || 'Failed to load status pages';
		} finally {
			isLoading = false;
		}
	}

	async function handleCreate() {
		selectedStatusPage = null;
		await loadModal();
		isModalOpen = true;
	}

	async function handleEdit(statusPage: StatusPage) {
		selectedStatusPage = statusPage;
		await loadModal();
		isModalOpen = true;
	}

	async function handleDelete(statusPage: StatusPage) {
		if (!confirm(`Are you sure you want to delete "${statusPage.name}"?`)) {
			return;
		}

		try {
			await fetchAPI(`/api/v1/status-pages/${statusPage.id}`, {
				method: 'DELETE'
			});
			await loadStatusPages();
		} catch (err: any) {
			alert(`Failed to delete status page: ${err.message}`);
		}
	}

	async function loadModal() {
		if (!modalLoaded) {
			const module = await import('$lib/components/StatusPageModal.svelte') as any;
			StatusPageModal = module.default;
			modalLoaded = true;
		}
	}

	function handleModalClose() {
		isModalOpen = false;
		selectedStatusPage = null;
	}

	async function handleModalSave() {
		await loadStatusPages();
		handleModalClose();
	}

	function getPublicUrl(slug: string): string {
		return `${window.location.origin}/status/${slug}`;
	}

	async function copyToClipboard(text: string) {
		try {
			if (navigator.clipboard && window.isSecureContext) {
				// Use the Clipboard API when available and in secure context
				await navigator.clipboard.writeText(text);
			} else {
				// Fallback for older browsers or non-HTTPS environments
				const textArea = document.createElement('textarea');
				textArea.value = text;
				textArea.style.position = 'fixed';
				textArea.style.left = '-999999px';
				textArea.style.top = '-999999px';
				document.body.appendChild(textArea);
				textArea.focus();
				textArea.select();
				document.execCommand('copy');
				document.body.removeChild(textArea);
			}
			// Show success toast
			showToastNotification('URL copied to clipboard!');
		} catch (err) {
			console.error('Failed to copy text: ', err);
			showToastNotification('Failed to copy URL');
		}
	}

	function showToastNotification(message: string) {
		toastMessage = message;
		showToast = true;
		// Hide toast after 3 seconds
		setTimeout(() => {
			showToast = false;
		}, 3000);
	}

	onMount(() => {
		loadStatusPages();
	});
</script>

<svelte:head>
	<title>Status Pages - V-Insight</title>
</svelte:head>

<div class="px-4 sm:px-6 lg:px-8 py-8">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-2xl font-semibold leading-6 text-slate-900 dark:text-white">Status Pages</h1>
			<p class="mt-2 text-sm text-slate-600 dark:text-slate-400">Create and manage public status pages for your services.</p>
		</div>
		<div class="mt-4 sm:ml-16 sm:mt-0 sm:flex-none">
			<button
				on:click={handleCreate}
				class="block rounded-md bg-blue-600 px-3 py-2 text-center text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
			>
				Create Status Page
			</button>
		</div>
	</div>

	{#if error}
		<div class="mt-8 rounded-md bg-red-50 dark:bg-red-900/20 p-4">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">
						<p>{error}</p>
					</div>
				</div>
			</div>
		</div>
	{:else if isLoading}
		<div class="mt-8 flex justify-center">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else}
		<div class="mt-8 grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each statusPages as statusPage}
				<Card>
					<div class="p-6">
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<h3 class="text-lg font-medium text-slate-900 dark:text-white">{statusPage.name}</h3>
								<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">/{statusPage.slug}</p>
								{#if statusPage.monitor_count !== undefined}
									<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
										{statusPage.monitor_count} monitor{statusPage.monitor_count !== 1 ? 's' : ''}
									</p>
								{/if}
							</div>
							<div class="flex items-center space-x-2">
								{#if statusPage.public_enabled}
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
										Public
									</span>
								{:else}
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200">
										Private
									</span>
								{/if}
							</div>
						</div>

						{#if statusPage.public_enabled}
							<div class="mt-4">
								<div class="block text-sm font-medium text-slate-700 dark:text-slate-300">Public URL</div>
								<div class="mt-1 flex">
									<input
										type="text"
										readonly
										value={getPublicUrl(statusPage.slug)}
										class="block w-full rounded-l-md border-slate-300 dark:border-slate-600 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm text-slate-900 dark:text-slate-100"
									/>
									<button
										on:click={() => copyToClipboard(getPublicUrl(statusPage.slug))}
										class="inline-flex items-center rounded-r-md border border-l-0 border-slate-300 dark:border-slate-600 bg-slate-50 dark:bg-slate-800 px-3 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700"
									>
										<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
										</svg>
									</button>
								</div>
							</div>
						{/if}

						<div class="mt-6 flex space-x-3">
							<button
								on:click={() => handleEdit(statusPage)}
								class="flex-1 rounded-md bg-white dark:bg-slate-800 px-3 py-2 text-sm font-semibold text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700"
							>
								Edit
							</button>
							<button
								on:click={() => handleDelete(statusPage)}
								class="flex-1 rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500"
							>
								Delete
							</button>
						</div>
					</div>
				</Card>
			{/each}

			{#if statusPages.length === 0}
				<div class="col-span-full">
					<Card>
						<div class="p-6 text-center">
							<svg class="mx-auto h-12 w-12 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192L12.888 3.75H8.25A2.25 2.25 0 006 6v11.25A2.25 2.25 0 008.25 19.5H12M8.25 19.5h3.75v-3.75H8.25v3.75z" />
							</svg>
							<h3 class="mt-2 text-sm font-medium text-slate-900 dark:text-white">No status pages</h3>
							<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Get started by creating your first status page.</p>
							<div class="mt-6">
								<button
									on:click={handleCreate}
									class="inline-flex items-center rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500"
								>
									Create Status Page
								</button>
							</div>
						</div>
					</Card>
				</div>
			{/if}
		</div>
	{/if}
</div>

{#if isModalOpen && StatusPageModal}
	<svelte:component
		this={StatusPageModal}
		isOpen={isModalOpen}
		statusPage={selectedStatusPage}
		on:close={handleModalClose}
		on:save={handleModalSave}
	/>
{/if}

<!-- Toast notification -->
{#if showToast}
	<div class="fixed bottom-4 right-4 z-50">
		<div class="bg-green-600 text-white px-4 py-2 rounded-md shadow-lg flex items-center space-x-2">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
			</svg>
			<span>{toastMessage}</span>
		</div>
	</div>
{/if}