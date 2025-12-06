<script lang="ts">
	import { onMount } from 'svelte';
    import { fade, fly } from 'svelte/transition';

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

    const features = [
        {
            title: 'Real-time Health Monitoring',
            desc: 'Experience unprecedented precision with our Go-based worker infrastructure. Monitor your services at the speed of thought.',
            icon: '⚡',
            items: ['Sub-millisecond checks', '99.99% uptime guarantee', 'Global edge network']
        },
        {
            title: 'SSL Security Guard',
            desc: 'Never let a certificate expire again. We strictly monitor your SSL certificates and notify you way ahead of time.',
            icon: '🔒',
            items: ['Automatic expiration tracking', 'Chain validation', 'Security score analysis']
        },
        {
            title: 'Instant Alerting',
            desc: 'Get notified the second something goes wrong via Email, Slack, Discord, or Webhooks.',
            icon: '🔔',
            items: ['Multi-channel integrated', 'Customizable thresholds', '0-delay dispatch']
        },
        {
            title: 'Deep Analytics',
            desc: 'Analyze response times and uptime history with our beautiful, data-rich dashboards.',
            icon: '📊',
            items: ['Response time graphs', 'Uptime heatmaps', 'Historical data retention']
        }
    ];

    let activeFeature = 0;
</script>

<svelte:head>
	<title>V-Insight - Beyond Limits</title>
</svelte:head>

<div class="bg-dark-950 min-h-screen text-white overflow-x-hidden selection:bg-brand-orange selection:text-white font-sans">
    <!-- Header -->
    <header class="fixed inset-x-0 top-0 z-50 bg-dark-950/80 backdrop-blur-md border-b border-white/5">
        <nav class="container mx-auto px-6 h-20 flex items-center justify-between" aria-label="Global">
            <div class="flex items-center gap-3">
                <div class="relative w-10 h-10 flex items-center justify-center bg-gradient-to-br from-brand-orange to-red-600 rounded-lg shadow-lg shadow-brand-orange/20">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor" class="w-6 h-6 text-white">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
                    </svg>
                </div>
                <span class="font-bold text-2xl tracking-wide text-white">V-INSIGHT</span>
            </div>
            
            <div class="hidden md:flex gap-x-8">
                <a href="#features" class="text-sm font-medium text-gray-300 hover:text-brand-orange transition-colors">FEATURES</a>
                <a href="#about" class="text-sm font-medium text-gray-300 hover:text-brand-orange transition-colors">ABOUT</a>
            </div>

            <div class="flex items-center gap-4">
                <div class="hidden lg:flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border border-white/10 text-xs">
                    <span class="w-2 h-2 rounded-full {apiStatus === 'Connected' ? 'bg-green-500 shadow-[0_0_10px_#22c55e]' : 'bg-red-500'} animate-pulse"></span>
                    <span class="text-gray-400">{apiStatus}</span>
                </div>
                <a href="/user/dashboard" class="text-sm font-medium text-white hover:text-brand-orange transition-colors">LOGIN</a>
            </div>
        </nav>
    </header>

    <main>
        <!-- Hero Section -->
        <div class="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
            <!-- Background Glows -->
            <div class="absolute top-0 left-1/2 -trnsulate-x-1/2 w-[800px] h-[500px] bg-brand-blue/20 rounded-full blur-[120px] -z-10 opacity-30 pointer-events-none"></div>
            <div class="absolute top-20 right-0 w-[600px] h-[600px] bg-brand-orange/20 rounded-full blur-[100px] -z-10 opacity-20 pointer-events-none"></div>

            <div class="container mx-auto px-6 text-center">
                <h2 class="text-brand-orange font-bold tracking-[0.2em] text-sm md:text-base mb-4 uppercase animate-fade-in-up">System Status: Maximum</h2>
                <h1 class="text-5xl md:text-7xl lg:text-8xl font-bold tracking-tight text-white mb-6 leading-tight">
                    BEYOND <span class="bg-clip-text text-transparent bg-gradient-to-r from-brand-orange to-red-500">LIMITS</span>
                </h1>
                <p class="text-xl text-gray-400 max-w-2xl mx-auto mb-10 font-light">
                    Where technology meets infinite possibilities. Monitor your infrastructure with next-generation precision.
                </p>
                <div class="flex flex-col sm:flex-row items-center justify-center gap-4">
                    <a href="/register" class="px-8 py-4 bg-gradient-to-r from-brand-orange to-red-600 rounded-sm font-bold tracking-wider hover:shadow-[0_0_20px_rgba(255,107,0,0.4)] transition-all duration-300 uppercase skew-x-[-10deg]">
                        <span class="block skew-x-[10deg]">Get Started</span>
                    </a>
                    <a href="#features" class="px-8 py-4 border border-brand-orange/50 text-brand-orange rounded-sm font-bold tracking-wider hover:bg-brand-orange/10 transition-all duration-300 uppercase skew-x-[-10deg]">
                        <span class="block skew-x-[10deg]">Learn More</span>
                    </a>
                </div>
            </div>
        </div>

        <!-- Features Section -->
        <div id="features" class="py-24 bg-dark-900 border-y border-white/5 relative">
            <div class="container mx-auto px-6">
                <h2 class="text-3xl md:text-4xl font-bold text-center mb-16 uppercase tracking-wider text-transparent bg-clip-text bg-gradient-to-b from-white to-white/50">
                    Core <span class="text-brand-blue">Features</span>
                </h2>

                <div class="grid lg:grid-cols-12 gap-8 items-start">
                    <!-- Tabs/List -->
                    <div class="lg:col-span-4 flex flex-col gap-2">
                        {#each features as feature, i}
                            <button 
                                class="text-left px-6 py-4 rounded-lg border transition-all duration-300 flex items-center gap-4 group {activeFeature === i ? 'border-brand-orange bg-brand-orange/10 shadow-[0_0_15px_rgba(255,107,0,0.1)]' : 'border-white/5 hover:border-white/10 hover:bg-white/5'}"
                                on:click={() => activeFeature = i}
                            >
                                <span class="text-2xl filter drop-shadow-lg">{feature.icon}</span>
                                <span class="font-bold tracking-wide {activeFeature === i ? 'text-white' : 'text-gray-400 group-hover:text-white'}">{feature.title}</span>
                            </button>
                        {/each}
                    </div>

                    <!-- Content -->
                    <div class="lg:col-span-8 bg-dark-950 border border-white/10 rounded-2xl p-8 lg:p-12 relative overflow-hidden min-h-[400px]">
                        <!-- Decorative glow -->
                        <div class="absolute top-0 right-0 w-64 h-64 bg-brand-blue/10 rounded-full blur-[80px]"></div>
                        
                        {#key activeFeature}
                            <div in:fade={{ duration: 300 }} class="relative z-10">
                                <h3 class="text-3xl font-bold mb-6 text-brand-orange">{features[activeFeature].title}</h3>
                                <p class="text-xl text-gray-300 mb-8 leading-relaxed max-w-2xl">{features[activeFeature].desc}</p>
                                <ul class="space-y-4">
                                    {#each features[activeFeature].items as item}
                                        <li class="flex items-center gap-3 text-gray-400">
                                            <div class="w-1.5 h-1.5 rounded-full bg-brand-blue shadow-[0_0_8px_#00C2FF]"></div>
                                            {item}
                                        </li>
                                    {/each}
                                </ul>
                            </div>
                        {/key}
                    </div>
                </div>
            </div>
        </div>

        <!-- About Section -->
        <div id="about" class="py-24 relative overflow-hidden">
            <div class="absolute inset-0 bg-dark-950"></div>
            <div class="container mx-auto px-6 relative z-10">
                <div class="grid md:grid-cols-2 gap-16 items-center">
                    <div>
                        <h2 class="text-4xl md:text-5xl font-bold mb-8 uppercase">
                            Pioneering the <span class="text-brand-blue">Digital Frontier</span>
                        </h2>
                        <div class="space-y-6 text-gray-400 md:text-lg leading-relaxed">
                            <p>
                                At V-INSIGHT, we don't just monitor technology — we empower it. Our mission is to bridge the gap between complex infrastructure and human understanding.
                            </p>
                            <p>
                                Founded by engineers who demanded better visibility, V-INSIGHT represents a quantum leap in monitoring tools. We combine real-time precision with beautiful analytics to deliver a solution that is both powerful and intuitive.
                            </p>
                            <div class="h-1 w-20 bg-gradient-to-r from-brand-orange to-transparent mt-8"></div>
                        </div>
                    </div>
                    <div class="relative flex justify-center items-center h-[400px]">
                        <!-- Abstract Geometric Shape -->
                        <div class="absolute w-64 h-64 border-2 border-brand-orange/30 transform rotate-45 animate-spin-slow shadow-[0_0_30px_rgba(255,107,0,0.2)]"></div>
                        <div class="absolute w-64 h-64 border-2 border-brand-blue/30 transform -rotate-12 animate-reverse-spin shadow-[0_0_30px_rgba(0,194,255,0.2)]"></div>
                        <div class="absolute w-40 h-40 bg-gradient-to-br from-brand-orange/10 to-brand-blue/10 blur-xl rounded-full"></div>
                    </div>
                </div>
            </div>
        </div>


    </main>

    <footer class="bg-black py-8 text-center text-gray-600 text-sm border-t border-white/5">
        <p>&copy; {new Date().getFullYear()} V-Insight. All rights reserved. <span class="text-brand-orange">Design by V-Team</span></p>
    </footer>
</div>

<style>
    /* Custom animations for the geometric shapes */
    @keyframes spin-slow {
        from { transform: rotate(45deg); }
        to { transform: rotate(405deg); }
    }
    .animate-spin-slow {
        animation: spin-slow 20s linear infinite;
    }
    
    @keyframes reverse-spin {
        from { transform: rotate(-12deg); }
        to { transform: rotate(-372deg); }
    }
    .animate-reverse-spin {
        animation: reverse-spin 25s linear infinite;
    }
</style>
