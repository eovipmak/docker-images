import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Function to get initial theme from localStorage or default to light
function getInitialTheme(): boolean {
	if (!browser) return false; // Default to light mode on server
	const stored = localStorage.getItem('theme');
	if (stored === 'dark') return true;
	if (stored === 'light') return false;
	// Default to system preference if no stored preference
	return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

// Create the theme store
export const themeStore = writable<boolean>(getInitialTheme());

// Subscribe to changes and update localStorage and document class
if (browser) {
	themeStore.subscribe((isDark) => {
		localStorage.setItem('theme', isDark ? 'dark' : 'light');
		if (isDark) {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	});
}

// Function to toggle theme
export function toggleTheme(): void {
	themeStore.update((current) => !current);
}