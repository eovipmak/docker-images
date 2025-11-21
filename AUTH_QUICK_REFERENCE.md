# Authentication Quick Reference

## For Developers: How to Use Auth in V-Insight

### Quick Start

```typescript
// In any Svelte component
import { authStore } from '$lib/stores/auth';
import { fetchAPI } from '$lib/api/client';

// Check if user is logged in
$authStore.isAuthenticated  // true/false

// Get current user info
$authStore.currentUser      // { id, email, tenant_id } or null

// Make authenticated API request (token added automatically)
const response = await fetchAPI('/api/v1/monitors');
const data = await response.json();
```

### Auth Store

```typescript
import { authStore } from '$lib/stores/auth';

// Login (after getting token from backend)
await authStore.login(token);

// Logout
authStore.logout();

// Check/refresh auth status
await authStore.checkAuth();

// Get current token
const token = authStore.getToken();
```

### Using in Components

```svelte
<script>
  import { authStore } from '$lib/stores/auth';
  
  // Reactive: updates when auth state changes
  $: if ($authStore.isAuthenticated) {
    console.log('User is logged in:', $authStore.currentUser);
  }
</script>

{#if $authStore.isAuthenticated}
  <p>Welcome, {$authStore.currentUser?.email}!</p>
{:else}
  <a href="/login">Login</a>
{/if}
```

### Making API Requests

```typescript
import { fetchAPI } from '$lib/api/client';

// GET request (authenticated)
const monitors = await fetchAPI('/api/v1/monitors')
  .then(r => r.json());

// POST request (authenticated)
const newMonitor = await fetchAPI('/api/v1/monitors', {
  method: 'POST',
  body: JSON.stringify({ name: 'My Site', url: 'https://example.com' })
}).then(r => r.json());

// Public endpoint (no auth)
const response = await fetchAPI('/api/v1/auth/login', {
  method: 'POST',
  skipAuth: true,  // Don't add token
  body: JSON.stringify({ email, password })
});
```

### Route Protection

Routes are automatically protected. To make a route public, add it to the list in `+layout.svelte`:

```typescript
const publicRoutes = ['/', '/login', '/register', '/about'];
```

### Custom Login Form

```svelte
<script lang="ts">
  import { authStore } from '$lib/stores/auth';
  
  let email = '';
  let password = '';
  let error = '';
  
  async function handleLogin() {
    const response = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    
    if (response.ok) {
      const { token } = await response.json();
      await authStore.login(token);
      window.location.href = '/dashboard';
    } else {
      error = 'Login failed';
    }
  }
</script>
```

### Server-Side Data Loading

```typescript
// +page.server.ts
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
  // Use fetch from context - it preserves cookies/headers
  const response = await fetch('/api/v1/monitors');
  const monitors = await response.json();
  
  return { monitors };
};
```

### Checking Auth in Server Hooks

```typescript
// hooks.server.ts (if needed)
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
  // Get token from cookie or header
  const token = event.cookies.get('auth_token');
  
  if (token) {
    // Validate token, attach user to event.locals
    event.locals.user = await validateToken(token);
  }
  
  return resolve(event);
};
```

## Common Patterns

### Protect a Component

```svelte
<script>
  import { authStore } from '$lib/stores/auth';
  import { goto } from '$app/navigation';
  
  $: if (!$authStore.isAuthenticated) {
    goto('/login');
  }
</script>
```

### Show Different Content Based on Auth

```svelte
{#if $authStore.isAuthenticated}
  <DashboardView />
{:else}
  <LandingPage />
{/if}
```

### Display User Email

```svelte
<p>Logged in as: {$authStore.currentUser?.email || 'Guest'}</p>
```

### Logout Button

```svelte
<button on:click={() => {
  authStore.logout();
  window.location.href = '/login';
}}>
  Logout
</button>
```

## API Endpoints Reference

**Public:**
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login

**Protected (requires auth):**
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/monitors` - List monitors
- `POST /api/v1/monitors` - Create monitor
- `GET /api/v1/monitors/:id` - Get monitor
- `PUT /api/v1/monitors/:id` - Update monitor
- `DELETE /api/v1/monitors/:id` - Delete monitor
- `GET /api/v1/alert-rules` - List alert rules
- `POST /api/v1/alert-rules` - Create alert rule
- And more... (see backend documentation)

## Troubleshooting

**Token not being sent:**
- Use `fetchAPI()` instead of `fetch()`
- Check localStorage has `auth_token`

**Getting redirected to login:**
- Token may be expired or invalid
- Check browser console for errors

**401 errors:**
- Token is missing or invalid
- Check if route requires authentication

**Can't access protected route:**
- Make sure you're logged in
- Token must be in localStorage

## Examples by Use Case

### Building a Protected Dashboard Page

```svelte
<!-- routes/dashboard/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchAPI } from '$lib/api/client';
  
  let stats = { monitors: 0, alerts: 0 };
  
  onMount(async () => {
    const monitors = await fetchAPI('/api/v1/monitors').then(r => r.json());
    const alerts = await fetchAPI('/api/v1/alert-rules').then(r => r.json());
    
    stats = {
      monitors: monitors.length,
      alerts: alerts.length
    };
  });
</script>

<h1>Dashboard</h1>
<p>Monitors: {stats.monitors}</p>
<p>Alerts: {stats.alerts}</p>
```

### Building a Settings Page

```svelte
<!-- routes/settings/+page.svelte -->
<script lang="ts">
  import { authStore } from '$lib/stores/auth';
  
  // User info is automatically available
  $: user = $authStore.currentUser;
</script>

<h1>Settings</h1>
<p>Email: {user?.email}</p>
<p>Tenant ID: {user?.tenant_id}</p>
```

### Creating a New Monitor

```svelte
<script lang="ts">
  import { fetchAPI } from '$lib/api/client';
  
  let name = '';
  let url = '';
  
  async function createMonitor() {
    const response = await fetchAPI('/api/v1/monitors', {
      method: 'POST',
      body: JSON.stringify({ name, url, interval: 60 })
    });
    
    if (response.ok) {
      alert('Monitor created!');
    }
  }
</script>

<form on:submit|preventDefault={createMonitor}>
  <input bind:value={name} placeholder="Name" required />
  <input bind:value={url} placeholder="URL" required />
  <button type="submit">Create</button>
</form>
```

## Best Practices

1. **Always use `fetchAPI()`** for API calls (automatic token handling)
2. **Check auth state reactively** with `$authStore.isAuthenticated`
3. **Handle errors** from API calls gracefully
4. **Don't store sensitive data** in component state unnecessarily
5. **Let the layout handle** route protection (don't duplicate logic)
6. **Use TypeScript** for better type safety
7. **Test auth flows** thoroughly before deploying

## Need More Help?

- See `AUTHENTICATION_IMPLEMENTATION.md` for technical details
- See `TESTING_GUIDE.md` for testing instructions
- See `frontend/src/lib/api/README.md` for API client docs
