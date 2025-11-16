<script lang="ts">
let email = '';
let password = '';

function handleSubmit() {
console.log('Login attempt:', email);
// TODO: Implement actual login logic
}
</script>

<svelte:head>
<title>Login - V-Insight</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
<div class="max-w-md mx-auto">
<div class="bg-white rounded-lg shadow-md p-8">
<h1 class="text-3xl font-bold text-gray-900 mb-2">Login</h1>
<p class="text-gray-600 mb-6">Sign in to access your monitoring dashboard</p>

<form on:submit|preventDefault={handleSubmit} class="space-y-4">
<div>
<label for="email" class="block text-sm font-medium text-gray-700 mb-1">
Email Address
</label>
<input
type="email"
id="email"
bind:value={email}
class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
placeholder="you@example.com"
required
/>
</div>

<div>
<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
Password
</label>
<input
type="password"
id="password"
bind:value={password}
class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
placeholder="••••••••"
required
/>
</div>

<button
type="submit"
class="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors font-medium"
>
Sign In
</button>
</form>

<p class="mt-4 text-sm text-gray-600 text-center">
Don't have an account? <a href="/register" class="text-blue-600 hover:underline">Sign up</a>
</p>
</div>
</div>
</div>
