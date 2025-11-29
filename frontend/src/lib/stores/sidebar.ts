import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Create a store for sidebar open/close state
function createSidebarStore() {
	const { subscribe, set, update } = writable(false);

	return {
		subscribe,
		toggle: () => update((n) => !n),
		open: () => set(true),
		close: () => set(false)
	};
}

export const sidebarOpen = createSidebarStore();

export function toggleSidebar() {
	sidebarOpen.toggle();
}

export function openSidebar() {
	sidebarOpen.open();
}

export function closeSidebar() {
	sidebarOpen.close();
}
