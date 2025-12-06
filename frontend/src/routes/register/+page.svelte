<script lang="ts">
	import { isAuthenticated } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let tenantName = '';
	let error = '';
	let isLoading = false;

	async function handleSubmit(event: Event) {
		event.preventDefault();
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
				await isAuthenticated.login(data.token);
				// Redirect to dashboard
				window.location.href = '/dashboard';
			} else {
				error = 'Registration failed - no token received';
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

<div class="min-h-screen flex items-center justify-center bg-dark-950 py-12 px-4 sm:px-6 lg:px-8 relative overflow-hidden font-sans selection:bg-brand-orange selection:text-white">
    <!-- Background Glows -->
    <div class="absolute top-0 right-1/2 translate-x-1/2 w-[800px] h-[500px] bg-brand-orange/10 rounded-full blur-[120px] -z-10 opacity-20 pointer-events-none"></div>
    <div class="absolute bottom-0 left-0 w-[600px] h-[600px] bg-brand-blue/10 rounded-full blur-[100px] -z-10 opacity-20 pointer-events-none"></div>

	<div class="max-w-md w-full space-y-8 relative z-10">
		<div class="text-center">
            <a href="/" class="mx-auto w-16 h-16 flex items-center justify-center bg-gradient-to-br from-brand-blue to-teal-400 rounded-2xl shadow-[0_0_20px_rgba(0,194,255,0.3)] mb-6 transform hover:scale-105 transition-transform duration-300">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor" class="w-8 h-8 text-white">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zM4 19.235v-.11a6.375 6.375 0 0112.75 0v.109A12.318 12.318 0 0110.374 21c-2.331 0-4.512-.645-6.374-1.766z" />
                </svg>
            </a>
			<h2 class="mt-6 text-3xl font-bold tracking-widest text-white uppercase">Create <span class="text-brand-blue">Account</span></h2>
			<p class="mt-2 text-sm text-gray-400">
				Start monitoring your domains in minutes
			</p>
		</div>

		<div class="bg-dark-950 py-8 px-4 shadow-2xl sm:rounded-2xl sm:px-10 border border-white/10 relative overflow-hidden group">
            <!-- Card Glow -->
            <div class="absolute -top-24 -left-24 w-48 h-48 bg-brand-orange/20 blur-[60px] rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-1000 pointer-events-none"></div>

            {#if error}
                <div class="mb-6 p-4 rounded-lg bg-red-900/20 border border-red-500/50 text-sm text-red-300 flex items-center shadow-[0_0_15px_rgba(220,38,38,0.1)]">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2 flex-shrink-0 text-red-500">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                    </svg>
                    {error}
                </div>
            {/if}

			<form class="space-y-6" on:submit={handleSubmit}>
				<div>
					<label for="email" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Email address
					</label>
					<div class="mt-1">
						<input
							id="email"
							name="email"
							type="email"
							autocomplete="email"
							required
                            bind:value={email}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner"
                            placeholder="you@example.com"
						/>
					</div>
				</div>

				<div>
					<label for="tenantName" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Organization Name
					</label>
					<div class="mt-1">
						<input
							id="tenantName"
							name="tenantName"
							type="text"
							required
                            bind:value={tenantName}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner"
                            placeholder="Acme Corp"
						/>
					</div>
				</div>

				<div>
					<label for="password" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Password
					</label>
					<div class="mt-1">
						<input
							id="password"
							name="password"
							type="password"
							autocomplete="new-password"
							required
                            bind:value={password}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner"
                            placeholder="••••••••"
						/>
					</div>
				</div>

				<div>
					<label for="confirmPassword" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Confirm Password
					</label>
					<div class="mt-1">
						<input
							id="confirmPassword"
							name="confirmPassword"
							type="password"
							autocomplete="new-password"
							required
                            bind:value={confirmPassword}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner"
                            placeholder="••••••••"
						/>
					</div>
				</div>

				<div>
					<button
						type="submit"
                        disabled={isLoading}
						class="w-full flex justify-center py-4 px-4 border border-transparent rounded-lg shadow-lg text-sm font-bold uppercase tracking-widest text-white bg-gradient-to-r from-brand-blue to-teal-500 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand-blue disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 transform hover:-translate-y-0.5"
					>
                        {#if isLoading}
                            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            INITIALIZING...
                        {:else}
						    REGISTER ACCOUNT
                        {/if}
					</button>
				</div>
			</form>

            <div class="mt-8">
                <div class="relative">
                    <div class="absolute inset-0 flex items-center">
                        <div class="w-full border-t border-white/10"></div>
                    </div>
                    <div class="relative flex justify-center text-sm">
                        <span class="px-2 bg-dark-950 text-gray-500 uppercase text-xs tracking-wider font-semibold">
                            Or
                        </span>
                    </div>
                </div>

                <div class="mt-6 text-center">
                    <a href="/login" class="text-sm font-bold text-gray-400 hover:text-white transition-colors uppercase tracking-wide">
                        Sign In To Existing Account
                    </a>
                </div>
            </div>
		</div>
	</div>
</div>
