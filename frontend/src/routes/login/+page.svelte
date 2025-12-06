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
				window.location.href = '/user/dashboard';
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

<div class="min-h-screen flex items-center justify-center bg-dark-950 py-12 px-4 sm:px-6 lg:px-8 relative overflow-hidden font-sans selection:bg-brand-orange selection:text-white">
    <!-- Background Glows -->
    <div class="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[500px] bg-brand-blue/10 rounded-full blur-[120px] -z-10 opacity-30 pointer-events-none"></div>
    <div class="absolute bottom-0 right-0 w-[600px] h-[600px] bg-brand-orange/10 rounded-full blur-[100px] -z-10 opacity-20 pointer-events-none"></div>

	<div class="max-w-md w-full space-y-8 relative z-10">
		<div class="text-center">
            <a href="/" class="mx-auto w-16 h-16 flex items-center justify-center bg-gradient-to-br from-brand-orange to-red-600 rounded-2xl shadow-[0_0_20px_rgba(255,107,0,0.3)] mb-6 transform hover:scale-105 transition-transform duration-300">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor" class="w-8 h-8 text-white">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                </svg>
            </a>
			<h2 class="mt-6 text-3xl font-bold tracking-widest text-white uppercase">Welcome <span class="text-brand-orange">Back</span></h2>
			<p class="mt-2 text-sm text-gray-400">
				Sign in to access your monitoring dashboard
			</p>
		</div>

		<div class="bg-dark-950 py-8 px-4 shadow-2xl sm:rounded-2xl sm:px-10 border border-white/10 relative overflow-hidden group">
            <!-- Card Glow -->
            <div class="absolute -top-24 -right-24 w-48 h-48 bg-brand-blue/20 blur-[60px] rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-1000 pointer-events-none"></div>

            {#if error}
                <div class="mb-6 p-4 rounded-lg bg-red-900/20 border border-red-500/50 text-sm text-red-300 flex items-center shadow-[0_0_15px_rgba(220,38,38,0.1)]">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2 flex-shrink-0 text-red-500">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
                    </svg>
                    {error}
                </div>
            {/if}

			<form class="space-y-6" on:submit|preventDefault={handleSubmit}>
				<div>
					<label for="email" class="block text-xs font-bold uppercase tracking-wider text-brand-orange mb-2">
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
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-orange focus:border-brand-orange sm:text-sm transition-all shadow-inner"
                            placeholder="you@example.com"
						/>
					</div>
				</div>

				<div>
					<label for="password" class="block text-xs font-bold uppercase tracking-wider text-brand-orange mb-2">
						Password
					</label>
					<div class="mt-1">
						<input
							id="password"
							name="password"
							type="password"
							autocomplete="current-password"
							required
                            bind:value={password}
                            disabled={isLoading}
							class="appearance-none block w-full px-4 py-3 bg-dark-900 border border-white/10 rounded-lg text-white placeholder-white/20 focus:outline-none focus:ring-1 focus:ring-brand-orange focus:border-brand-orange sm:text-sm transition-all shadow-inner"
                            placeholder="••••••••"
						/>
					</div>
				</div>

				<div>
					<button
						type="submit"
                        disabled={isLoading}
						class="w-full flex justify-center py-4 px-4 border border-transparent rounded-lg shadow-lg text-sm font-bold uppercase tracking-widest text-white bg-gradient-to-r from-brand-orange to-red-600 hover:brightness-110 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand-orange disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 transform hover:-translate-y-0.5"
					>
                        {#if isLoading}
                            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                            </svg>
                            ACCESSING SYSTEM...
                        {:else}
						    ENTER SYSTEM
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
                    <a href="/register" class="text-sm font-bold text-gray-400 hover:text-white transition-colors uppercase tracking-wide">
                        Initialize New Account
                    </a>
                </div>
            </div>
		</div>
	</div>
</div>
