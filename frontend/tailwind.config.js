/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				dark: {
					50: '#f8f9fa',
					100: '#f3f4f6',
					200: '#e5e7eb',
					300: '#d1d5db',
					400: '#9ca3af',
					500: '#6b7280',
					600: '#4b5563',
					700: '#374151',
					800: '#1f2937',
					900: '#111827',
					950: '#050505' // Darker background
				},
				brand: {
					orange: '#FF6B00',
					blue: '#00C2FF'
				}
			},
			fontFamily: {
				sans: ['Outfit', 'sans-serif'],
			}
		}
	},
	plugins: []
};
