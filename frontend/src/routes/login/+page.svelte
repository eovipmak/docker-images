<script lang="ts">
	import { isAuthenticated } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let error = '';
	let isLoading = false;

	async function handleSubmit() {
		error = '';
		isLoading = true;

		try {
			const response = await fetch('/api/v1/auth/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					email,
					password
				})
			});

			const data = await response.json();

			if (!response.ok) {
				error = data.error || 'Login failed';
				return;
			}

			// Store the token
			if (data.token) {
				await isAuthenticated.login(data.token);
				// Redirect to dashboard
				window.location.href = '/dashboard';
			}
		} catch (err) {
			error = 'An error occurred. Please try again.';
			console.error('Login error:', err);
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
<title>Login - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
<div class="max-w-md mx-auto">
<div class="bg-white rounded-lg shadow-md p-8">
<h1 class="text-3xl font-bold text-gray-900 mb-2">Login</h1>
<p class="text-gray-600 mb-6">Sign in to access your monitoring dashboard</p>

{#if error}
<div class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
{error}
</div>
{/if}

<form on:submit|preventDefault={handleSubmit} class="space-y-4">
<div>
<label for="email" class="block text-sm font-medium text-gray-700 mb-1">
Email Address
</label>
<input
type="email"
id="email"
bind:value={email}
class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
placeholder="you@example.com"
required
disabled={isLoading}
/>
</div>

<div>
<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
Password
</label>
<input
type="password"
id="password"
bind:value={password}
class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
placeholder="••••••••"
required
disabled={isLoading}
/>
</div>

<button
type="submit"
class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
disabled={isLoading}
>
{isLoading ? 'Signing In...' : 'Sign In'}
</button>
</form>

<p class="mt-4 text-sm text-gray-600 text-center">
Don't have an account? <a href="/register" class="text-blue-600 hover:underline">Sign up</a>
</p>
</div>
</div>
</div>
