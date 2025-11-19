<script lang="ts">
	import { isAuthenticated } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let tenantName = '';
	let error = '';
	let isLoading = false;

	async function handleSubmit() {
		error = '';
		
		// Validate passwords match
		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		// Validate password length
		if (password.length < 6) {
			error = 'Password must be at least 6 characters';
			return;
		}

		isLoading = true;

		try {
			const response = await fetch('/api/v1/auth/register', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					email,
					password,
					tenant_name: tenantName
				})
			});

			const data = await response.json();

			if (!response.ok) {
				error = data.error || 'Registration failed';
				return;
			}

			// Store the token
			if (data.token) {
				isAuthenticated.login(data.token);
				// Redirect to dashboard
				window.location.href = '/dashboard';
			}
		} catch (err) {
			error = 'An error occurred. Please try again.';
			console.error('Registration error:', err);
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Register - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-md mx-auto">
		<div class="bg-white rounded-lg shadow-md p-8">
			<h1 class="text-3xl font-bold text-gray-900 mb-2">Create Account</h1>
			<p class="text-gray-600 mb-6">Sign up to start monitoring your domains</p>

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
						name="email"
						bind:value={email}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="you@example.com"
						required
						disabled={isLoading}
					/>
				</div>

				<div>
					<label for="tenantName" class="block text-sm font-medium text-gray-700 mb-1">
						Organization Name
					</label>
					<input
						type="text"
						id="tenantName"
						name="tenant_name"
						bind:value={tenantName}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Your Company"
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
						name="password"
						bind:value={password}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="••••••••"
						required
						disabled={isLoading}
						minlength="6"
					/>
					<p class="text-xs text-gray-500 mt-1">Minimum 6 characters</p>
				</div>

				<div>
					<label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-1">
						Confirm Password
					</label>
					<input
						type="password"
						id="confirmPassword"
						bind:value={confirmPassword}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="••••••••"
						required
						disabled={isLoading}
						minlength="6"
					/>
				</div>

				<button
					type="submit"
					class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
					disabled={isLoading}
				>
					{isLoading ? 'Creating Account...' : 'Sign Up'}
				</button>
			</form>

			<p class="mt-4 text-sm text-gray-600 text-center">
				Already have an account? <a href="/login" class="text-blue-600 hover:underline">Sign in</a>
			</p>
		</div>
	</div>
</div>
