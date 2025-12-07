<script lang="ts">
	import { fetchAPI } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth';

	let user = $authStore.currentUser;
	let currentPassword = '';
	let newPassword = '';
	let confirmPassword = '';
	let statusMessage = '';
	let statusType: 'success' | 'error' | '' = '';
	let isSaving = false;
	let showPasswordModal = false;

	$: user = $authStore.currentUser;

	const resetStatus = () => {
		statusMessage = '';
		statusType = '';
	};

	const handleChangePassword = async () => {
		resetStatus();

		if (!currentPassword || !newPassword) {
			statusType = 'error';
			statusMessage = 'Please enter your current password and a new password.';
			return;
		}

		if (newPassword.length < 6) {
			statusType = 'error';
			statusMessage = 'New password must be at least 6 characters.';
			return;
		}

		if (newPassword !== confirmPassword) {
			statusType = 'error';
			statusMessage = 'New password and confirmation do not match.';
			return;
		}

		isSaving = true;

		try {
			const response = await fetchAPI('/api/v1/auth/change-password', {
				method: 'POST',
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword
				})
			});

			const data = await response.json().catch(() => ({}));

			if (!response.ok) {
				throw new Error(data.error || 'Unable to change password.');
			}

			statusType = 'success';
			statusMessage = data.message || 'Password updated successfully.';
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			showPasswordModal = false;
		} catch (error) {
			statusType = 'error';
			statusMessage = (error as Error).message;
		} finally {
			isSaving = false;
		}
	};
</script>

<svelte:head>
	<title>Your Profile - V-Insight</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
	<div class="sm:flex sm:items-center">
		<div class="sm:flex-auto">
			<h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">Your Profile</h1>
			<p class="mt-1 text-sm text-slate-600 dark:text-slate-400">Manage your account information and preferences.</p>
		</div>
	</div>

	<div class="max-w-3xl space-y-6">
		<!-- Profile Information + Change Password CTA -->
		<div class="bg-white dark:bg-slate-800 shadow-sm ring-1 ring-slate-900/5 dark:ring-slate-700 sm:rounded-lg">
			<div class="px-6 py-5 sm:px-8 sm:py-6">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<h3 class="text-base font-semibold leading-6 text-slate-900 dark:text-white">Profile Information</h3>
						<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Your account details.</p>
					</div>
					<button
						type="button"
						class="inline-flex items-center justify-center rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
						on:click={() => {
							resetStatus();
							showPasswordModal = true;
						}}
					>
						Change password
					</button>
				</div>

				{#if statusMessage && !showPasswordModal}
					<div
						class={`mt-4 rounded-md px-4 py-3 text-sm ${statusType === 'success'
							? 'bg-green-50 text-green-800 ring-1 ring-green-200'
							: 'bg-red-50 text-red-800 ring-1 ring-red-200'}`}
						role="status"
						aria-live="polite"
					>
						{statusMessage}
					</div>
				{/if}

				<div class="mt-5 space-y-4">
					<div>
						<label for="email" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Email</label>
						<div class="mt-2">
							<input
								type="email"
								name="email"
								id="email"
								value={user?.email || ''}
								readonly
								class="block w-full rounded-md border-0 bg-white dark:bg-slate-700 px-3 py-2 text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-200 dark:ring-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:text-sm"
							/>
						</div>
					</div>
					<div>
						<label for="role" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Role</label>
						<div class="mt-2">
							<input
								type="text"
								name="role"
								id="role"
								value={user?.role || 'user'}
								readonly
								class="block w-full rounded-md border-0 bg-white dark:bg-slate-700 px-3 py-2 text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-200 dark:ring-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:text-sm"
							/>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

{#if showPasswordModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 px-4 py-8">
		<div class="relative w-full max-w-lg rounded-lg bg-white px-6 py-6 shadow-2xl ring-1 ring-slate-900/10 dark:bg-slate-800 dark:ring-slate-700">
			<div class="flex items-start justify-between gap-4">
				<div>
					<h3 class="text-lg font-semibold leading-6 text-slate-900 dark:text-white">Change password</h3>
					<p class="mt-1 text-sm text-slate-500 dark:text-slate-400">Enter your current password and set a new one.</p>
				</div>
				<button
					type="button"
					class="rounded-md p-2 text-slate-500 hover:text-slate-700 hover:bg-slate-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 dark:text-slate-300 dark:hover:text-white dark:hover:bg-slate-700"
					on:click={() => {
						showPasswordModal = false;
						resetStatus();
					}}
					aria-label="Close"
				>
					✕
				</button>
			</div>

			{#if statusMessage}
				<div
					class={`mt-4 rounded-md px-4 py-3 text-sm ${statusType === 'success'
						? 'bg-green-50 text-green-800 ring-1 ring-green-200'
						: 'bg-red-50 text-red-800 ring-1 ring-red-200'}`}
					role="status"
					aria-live="polite"
				>
					{statusMessage}
				</div>
			{/if}

			<form class="mt-6 space-y-4" on:submit|preventDefault={handleChangePassword}>
				<div>
					<label for="current-password" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Current password</label>
					<input
						type="password"
						id="current-password"
						autocomplete="current-password"
						class="mt-2 block w-full rounded-md border-0 bg-white dark:bg-slate-700 px-3 py-2 text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:text-sm"
						bind:value={currentPassword}
						required
					/>
				</div>

				<div class="grid gap-4 sm:grid-cols-2">
					<div>
						<label for="new-password" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">New password</label>
						<input
							type="password"
							id="new-password"
							autocomplete="new-password"
							class="mt-2 block w-full rounded-md border-0 bg-white dark:bg-slate-700 px-3 py-2 text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:text-sm"
							bind:value={newPassword}
							required
						/>
					</div>
					<div>
						<label for="confirm-password" class="block text-sm font-medium leading-6 text-slate-900 dark:text-white">Confirm new password</label>
						<input
							type="password"
							id="confirm-password"
							autocomplete="new-password"
							class="mt-2 block w-full rounded-md border-0 bg-white dark:bg-slate-700 px-3 py-2 text-slate-900 dark:text-white shadow-sm ring-1 ring-inset ring-slate-300 dark:ring-slate-600 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:ring-2 focus:ring-inset focus:ring-blue-500 sm:text-sm"
							bind:value={confirmPassword}
							required
						/>
					</div>
				</div>

				<div class="flex items-center justify-end gap-3 pt-2">
					<button
						type="button"
						class="text-sm font-medium text-slate-500 hover:text-slate-700 dark:text-slate-300 dark:hover:text-white"
						on:click={() => {
							showPasswordModal = false;
							currentPassword = '';
							newPassword = '';
							confirmPassword = '';
							resetStatus();
						}}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="inline-flex items-center justify-center rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-60"
						disabled={isSaving}
					>
						{isSaving ? 'Saving...' : 'Update password'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}