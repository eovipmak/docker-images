<script lang="ts">
	import { isAuthenticated } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let confirmPassword = '';
	let showPassword = false;
	let showConfirmPassword = false;
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
				window.location.href = '/user/dashboard';
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
					<label for="password" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Password
					</label>
					<div class="mt-1 relative">
						<input
							id="password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="new-password"
							required
                            value={password}
							on:input={(e) => password = e.currentTarget.value}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner pr-10"
							placeholder="••••••••"
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-white transition-colors focus:outline-none focus:ring-2 focus:ring-brand-blue focus:ring-offset-1 focus:ring-offset-dark-900 rounded-md"
							on:click={() => (showPassword = !showPassword)}
							aria-label={showPassword ? 'Hide password' : 'Show password'}
						>
							{#if showPassword}
								<!-- Eye Slash Icon -->
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
									<path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
								</svg>
							{:else}
								<!-- Eye Icon -->
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
									<path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
									<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
							{/if}
						</button>
					</div>
				</div>

				<div>
					<label for="confirmPassword" class="block text-xs font-bold uppercase tracking-wider text-brand-blue mb-2">
						Confirm Password
					</label>
					<div class="mt-1 relative">
						<input
							id="confirmPassword"
							name="confirmPassword"
							type={showConfirmPassword ? 'text' : 'password'}
							autocomplete="new-password"
							required
                            value={confirmPassword}
							on:input={(e) => confirmPassword = e.currentTarget.value}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-blue focus:border-brand-blue sm:text-sm transition-all shadow-inner pr-10"
							placeholder="••••••••"
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-white transition-colors focus:outline-none focus:ring-2 focus:ring-brand-blue focus:ring-offset-1 focus:ring-offset-dark-900 rounded-md"
							on:click={() => (showConfirmPassword = !showConfirmPassword)}
							aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
						>
							{#if showConfirmPassword}
								<!-- Eye Slash Icon -->
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
									<path stroke-linecap="round" stroke-linejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
								</svg>
							{:else}
								<!-- Eye Icon -->
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
									<path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
									<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
							{/if}
						</button>
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
