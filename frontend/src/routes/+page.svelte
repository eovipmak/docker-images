<script lang="ts">
	import { onMount } from 'svelte';

	let apiStatus = 'Checking...';
	let apiVersion = '';

	onMount(async () => {
		try {
			const response = await fetch('/api/v1');
			if (response.ok) {
				const data = await response.json();
				apiStatus = 'Connected';
				apiVersion = data.version;
			} else {
				apiStatus = 'Error';
			}
		} catch (error) {
			apiStatus = 'Disconnected';
		}
	});
</script>

<svelte:head>
	<title>V-Insight - Modern Monitoring for Modern Teams</title>
</svelte:head>

<div class="bg-white dark:bg-slate-950">
    <!-- Header -->
    <header class="absolute inset-x-0 top-0 z-50">
        <nav class="flex items-center justify-between p-6 lg:px-8" aria-label="Global">
            <div class="flex lg:flex-1">
                <a href="/" class="-m-1.5 p-1.5 flex items-center gap-2">
                    <div class="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-5 h-5 text-white">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                        </svg>
                    </div>
                    <span class="font-bold text-xl tracking-tight text-slate-900 dark:text-white">V-Insight</span>
                </a>
            </div>
            <div class="flex flex-1 justify-end gap-x-6">
                <a href="/login" class="text-sm font-semibold leading-6 text-slate-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors">Log in</a>
                <a href="/register" class="rounded-full bg-blue-600 px-4 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-colors">Sign up</a>
            </div>
        </nav>
    </header>

    <main class="isolate">
        <!-- Hero section -->
        <div class="relative pt-14">
            <div class="absolute inset-x-0 -top-40 -z-10 transform-gpu overflow-hidden blur-3xl sm:-top-80" aria-hidden="true">
                <div class="relative left-[calc(50%-11rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 rotate-[30deg] bg-gradient-to-tr from-[#ff80b5] to-[#9089fc] opacity-30 sm:left-[calc(50%-30rem)] sm:w-[72.1875rem]" style="clip-path: polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)"></div>
            </div>
            
            <div class="py-24 sm:py-32 lg:pb-40">
                <div class="mx-auto max-w-7xl px-6 lg:px-8">
                    <div class="mx-auto max-w-2xl text-center">
                        <h1 class="text-4xl font-bold tracking-tight text-slate-900 dark:text-white sm:text-6xl">
                            Monitor your infrastructure with confidence
                        </h1>
                        <p class="mt-6 text-lg leading-8 text-slate-600 dark:text-slate-300">
                            V-Insight provides real-time monitoring, instant alerts, and detailed analytics for your websites and services. Simple, reliable, and built for developers.
                        </p>
                        <div class="mt-10 flex items-center justify-center gap-x-6">
                            <a href="/register" class="rounded-lg bg-blue-600 px-6 py-3 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-all hover:scale-105">Get started</a>
                            <a href="/login" class="text-sm font-semibold leading-6 text-slate-900 dark:text-white flex items-center gap-1 hover:gap-2 transition-all">
                                Live demo <span aria-hidden="true">→</span>
                            </a>
                        </div>
                    </div>
                    
                    <!-- Status Badge -->
                    <div class="mt-16 flex justify-center">
                        <div class="rounded-full px-4 py-1 text-sm leading-6 text-slate-500 dark:text-slate-400 ring-1 ring-slate-900/10 dark:ring-white/10 hover:ring-slate-900/20 dark:hover:ring-white/20 bg-white/50 dark:bg-slate-900/50 backdrop-blur-sm">
                            System Status: 
                            <span class={apiStatus === 'Connected' ? 'font-semibold text-emerald-600 dark:text-emerald-400' : 'font-semibold text-rose-600 dark:text-rose-400'}>
                                {apiStatus}
                            </span>
                            {#if apiVersion}
                                <span class="mx-2 text-slate-300 dark:text-slate-600">|</span>
                                <span class="text-slate-400 dark:text-slate-500">v{apiVersion}</span>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="absolute inset-x-0 top-[calc(100%-13rem)] -z-10 transform-gpu overflow-hidden blur-3xl sm:top-[calc(100%-30rem)]" aria-hidden="true">
                <div class="relative left-[calc(50%+3rem)] aspect-[1155/678] w-[36.125rem] -translate-x-1/2 bg-gradient-to-tr from-[#ff80b5] to-[#9089fc] opacity-30 sm:left-[calc(50%+36rem)] sm:w-[72.1875rem]" style="clip-path: polygon(74.1% 44.1%, 100% 61.6%, 97.5% 26.9%, 85.5% 0.1%, 80.7% 2%, 72.5% 32.5%, 60.2% 62.4%, 52.4% 68.1%, 47.5% 58.3%, 45.2% 34.5%, 27.5% 76.7%, 0.1% 64.9%, 17.9% 100%, 27.6% 76.8%, 76.1% 97.7%, 74.1% 44.1%)"></div>
            </div>
        </div>
    </main>
</div>

