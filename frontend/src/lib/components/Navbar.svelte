<script lang="ts">
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import { themeStore, toggleTheme } from '$lib/stores/theme';
	import { slide } from 'svelte/transition';

	let isMenuOpen = false;
    let isProfileOpen = false;

	export let navItems = [
		{ name: 'Dashboard', path: '/user/dashboard' },
		{ name: 'Monitors', path: '/user/monitors' },
		{ name: 'Incidents', path: '/user/incidents' },
        { name: 'Settings', path: '/user/settings' }
	];
	export let homeLink = '/user/dashboard';
    export let isAdmin = false;

    function toggleMenu() {
        isMenuOpen = !isMenuOpen;
    }

    function toggleProfile() {
        isProfileOpen = !isProfileOpen;
    }

    // Close menus when clicking outside
    function clickOutside(node: Node) {
        const handleClick = (event: MouseEvent) => {
            if (node && !node.contains(event.target as Node) && !event.defaultPrevented) {
                isProfileOpen = false;
            }
        };

        document.addEventListener('click', handleClick, true);

        return {
            destroy() {
                document.removeEventListener('click', handleClick, true);
            }
        };
    }
</script>

<nav class="sticky top-0 z-50 w-full bg-white/80 dark:bg-[#0f1020]/80 backdrop-blur-md border-b border-gray-200 dark:border-indigo-500/30 transition-colors duration-300">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="flex items-center justify-between h-16 sm:h-20">
			<!-- Logo -->
			<div class="flex-shrink-0 flex items-center">
				<a href={homeLink} class="flex items-center gap-2 group" aria-label="V-Insight Home">
                    <div class="relative w-8 h-8 flex items-center justify-center rounded-lg bg-indigo-600/10 dark:bg-cyan-500/10 group-hover:scale-110 transition-transform duration-300">
                        <svg aria-hidden="true" class="w-5 h-5 text-indigo-600 dark:text-cyan-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                    </div>
					<span class="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-indigo-600 to-violet-600 dark:from-cyan-400 dark:to-purple-400 font-outfit uppercase tracking-wider">
						V-Insight
					</span>
				</a>
			</div>

			<!-- Desktop Nav -->
			<div class="hidden md:block">
				<div class="flex items-baseline space-x-1">
					{#each navItems as item}
						<a
							href={item.path}
							class="px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 
                                {$page.url.pathname === item.path 
                                    ? isAdmin 
                                        ? 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/10 shadow-[0_0_10px_rgba(220,38,38,0.2)]'
                                        : 'text-indigo-600 dark:text-cyan-400 bg-indigo-50 dark:bg-cyan-900/10 shadow-[0_0_10px_rgba(34,211,238,0.2)]'
                                    : 'text-gray-600 dark:text-gray-300 hover:text-indigo-600 dark:hover:text-cyan-300 hover:bg-gray-50 dark:hover:bg-white/5'}"
						>
							{item.name}
						</a>
					{/each}
				</div>
			</div>

			<!-- Right Side -->
			<div class="hidden md:flex items-center gap-4">
                <!-- Theme Toggle -->
                <button 
                    on:click={toggleTheme}
                    class="p-2.5 rounded-full transition-all duration-200 hover:scale-105 active:scale-95
                           {$themeStore 
                               ? 'bg-white/10 text-yellow-400 hover:bg-white/20' 
                               : 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100 shadow-sm'}"
                    aria-label="Toggle Dark Mode"
                >
                    {#if $themeStore}
                        <!-- Moon Icon (Solid) for Dark Mode -->
                        <svg aria-hidden="true" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                            <path fill-rule="evenodd" d="M9.528 1.718a.75.75 0 01.162.819A8.97 8.97 0 009 6a9 9 0 009 9 8.97 8.97 0 003.463-.69.75.75 0 01.981.98 10.503 10.503 0 01-9.694 6.46c-5.799 0-10.5-4.701-10.5-10.5 0-4.368 2.667-8.112 6.46-9.694a.75.75 0 01.818.162z" clip-rule="evenodd" />
                        </svg>
                    {:else}
                        <!-- Sun Icon (Solid) for Light Mode -->
                        <svg aria-hidden="true" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M12 2.25a.75.75 0 01.75.75v2.25a.75.75 0 01-1.5 0V3a.75.75 0 01.75-.75zM7.5 12a4.5 4.5 0 119 0 4.5 4.5 0 01-9 0zM18.894 6.166a.75.75 0 00-1.06-1.06l-1.591 1.59a.75.75 0 101.06 1.061l1.591-1.59zM21.75 12a.75.75 0 01-.75.75h-2.25a.75.75 0 010-1.5H21a.75.75 0 01.75.75zM17.834 18.894a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 10-1.061 1.06l1.59 1.591zM12 18a.75.75 0 01.75.75V21a.75.75 0 01-1.5 0v-2.25A.75.75 0 0112 18zM7.758 17.303a.75.75 0 00-1.061-1.06l-1.591 1.59a.75.75 0 001.06 1.061l1.591-1.59zM6 12a.75.75 0 01-.75.75H3a.75.75 0 010-1.5h2.25A.75.75 0 016 12zM6.697 7.757a.75.75 0 001.06-1.06l-1.59-1.591a.75.75 0 00-1.061 1.06l1.59 1.591z" />
                        </svg>
                    {/if}
                </button>

                <!-- Profile Dropdown -->
                <div class="relative" use:clickOutside>
                    <button 
                        on:click={toggleProfile}
                        aria-label="Open user menu"
                        aria-haspopup="true"
                        aria-expanded={isProfileOpen}
                        class="flex items-center gap-2 p-1 pl-2 pr-1 rounded-full border transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:focus:ring-cyan-500
                               {isAdmin 
                                   ? 'border-red-500 bg-red-50 dark:bg-red-900/20 hover:border-red-400' 
                                   : 'border-gray-200 dark:border-indigo-500/30 bg-gray-50 dark:bg-[#1a1c2e] hover:border-indigo-300 dark:hover:border-cyan-500/50'}"
                    >
                        {#if isAdmin}
                              <span class="text-sm font-bold text-red-600 dark:text-red-400">ADMIN</span>
                              <div class="w-8 h-8 rounded-full bg-gradient-to-br from-red-500 to-orange-600 flex items-center justify-center text-white font-bold text-xs shadow-lg">
                                  A
                              </div>
                        {:else}
                            <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{$authStore.currentUser?.email || 'User'}</span>
                            <div class="w-8 h-8 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white font-bold text-xs shadow-lg">
                                {($authStore.currentUser?.email || 'U').charAt(0).toUpperCase()}
                            </div>
                        {/if}
                    </button>

                    {#if isProfileOpen}
                        <div 
                            transition:slide={{ duration: 200 }}
                            class="absolute right-0 mt-2 w-48 rounded-xl shadow-lg py-1 bg-white dark:bg-[#1a1c2e] ring-1 ring-black ring-opacity-5 focus:outline-none border border-gray-100 dark:border-indigo-500/20"
                        >
                            <div class="py-1" role="menu" aria-orientation="vertical" aria-labelledby="user-menu-button" tabindex="-1">
                                {#if !isAdmin}
                                    <a
                                        href="/user/profile"
                                        class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-white/5 transition-colors"
                                        role="menuitem"
                                        tabindex="-1"
                                        id="user-menu-item-0"
                                        on:click={() => (isProfileOpen = false)}
                                    >
                                        Your Profile
                                    </a>
                                    <a
                                        href="/user/settings"
                                        class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-white/5 transition-colors"
                                        role="menuitem"
                                        tabindex="-1"
                                        id="user-menu-item-1"
                                        on:click={() => (isProfileOpen = false)}
                                    >
                                        Settings
                                    </a>
                                {/if}
                                <button on:click={authStore.logout} class="block w-full text-left px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/10">Sign out</button>
                            </div>
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Mobile menu button -->
            <div class="-mr-2 flex md:hidden">
                <button 
                    on:click={toggleMenu}
                    type="button"
                    class="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-800 focus:ring-white"
                    aria-controls="mobile-menu"
                    aria-expanded={isMenuOpen}
                >
                    <span class="sr-only">{isMenuOpen ? 'Close main menu' : 'Open main menu'}</span>
                    {#if isMenuOpen}
                        <svg aria-hidden="true" class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    {:else}
                        <svg aria-hidden="true" class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    {/if}
                </button>
            </div>
		</div>
	</div>

    <!-- Mobile Menu -->
    {#if isMenuOpen}
        <div class="md:hidden bg-white dark:bg-[#0f1020] border-b border-gray-200 dark:border-indigo-500/30" transition:slide id="mobile-menu">
            <div class="px-2 pt-2 pb-3 space-y-1 sm:px-3">
                {#each navItems as item}
                    <a
                        href={item.path}
                        class="block px-3 py-2 rounded-md text-base font-medium text-gray-700 dark:text-gray-300 hover:text-indigo-600 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-white/5"
                    >
                        {item.name}
                    </a>
                {/each}
                <button 
                    on:click={authStore.logout}
                    class="block w-full text-left px-3 py-2 rounded-md text-base font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/10"
                >
                    Sign out
                </button>
            </div>
        </div>
    {/if}
</nav>
