# API Client Usage Guide

This directory contains the API client for making authenticated requests to the V-Insight backend.

## Quick Start

```typescript
import { fetchAPI } from '$lib/api/client';

// Authenticated request (token automatically added)
const response = await fetchAPI('/api/v1/monitors');
const monitors = await response.json();

// With options
const response = await fetchAPI('/api/v1/monitors', {
  method: 'POST',
  body: JSON.stringify({ name: 'My Monitor', url: 'https://example.com' })
});
```

## Features

- **Automatic Authentication**: JWT token from localStorage automatically added to all requests
- **401 Handling**: Automatically redirects to login if token is invalid/expired
- **CORS-Free**: All requests proxied through SvelteKit server (no CORS issues)
- **Type-Safe**: TypeScript support for better DX

## API

### `fetchAPI(endpoint, options)`

**Parameters:**
- `endpoint` (string): API endpoint path (e.g., '/api/v1/monitors')
- `options` (object): Optional fetch options
  - All standard fetch options supported
  - `skipAuth` (boolean): Skip automatic token injection (for public endpoints)

**Returns:** Promise<Response>

## Examples

### GET Request
```typescript
// List all monitors
const response = await fetchAPI('/api/v1/monitors');
if (response.ok) {
  const data = await response.json();
  console.log('Monitors:', data);
}
```

### POST Request
```typescript
// Create a new monitor
const response = await fetchAPI('/api/v1/monitors', {
  method: 'POST',
  body: JSON.stringify({
    name: 'My Website',
    url: 'https://example.com',
    interval: 60
  })
});

if (response.ok) {
  const monitor = await response.json();
  console.log('Created:', monitor);
} else {
  const error = await response.json();
  console.error('Error:', error);
}
```

### PUT Request
```typescript
// Update a monitor
const response = await fetchAPI(`/api/v1/monitors/${id}`, {
  method: 'PUT',
  body: JSON.stringify({
    name: 'Updated Name',
    interval: 120
  })
});
```

### DELETE Request
```typescript
// Delete a monitor
const response = await fetchAPI(`/api/v1/monitors/${id}`, {
  method: 'DELETE'
});

if (response.ok) {
  console.log('Monitor deleted');
}
```

### Public Endpoint (No Auth)
```typescript
// Login request (public endpoint)
const response = await fetchAPI('/api/v1/auth/login', {
  method: 'POST',
  skipAuth: true,  // Don't add auth token
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});
```

### Error Handling
```typescript
try {
  const response = await fetchAPI('/api/v1/monitors');
  
  if (!response.ok) {
    const error = await response.json();
    console.error('API Error:', error);
    return;
  }
  
  const data = await response.json();
  // Process data
} catch (err) {
  console.error('Network Error:', err);
}
```

### Using in Svelte Components

**In +page.svelte (client-side):**
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { fetchAPI } from '$lib/api/client';
  
  let monitors = [];
  
  onMount(async () => {
    const response = await fetchAPI('/api/v1/monitors');
    if (response.ok) {
      monitors = await response.json();
    }
  });
</script>
```

**In +page.server.ts (server-side):**
```typescript
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
  // Use the fetch from load context (server-side)
  const response = await fetch('/api/v1/monitors');
  
  if (!response.ok) {
    throw error(response.status, 'Failed to load monitors');
  }
  
  const monitors = await response.json();
  return { monitors };
};
```

## Important Notes

1. **Token Management**: The client automatically retrieves the token from localStorage. No need to manually add Authorization headers.

2. **Auto-Redirect on 401**: If the backend returns 401 (Unauthorized), the client will:
   - Clear the invalid token from localStorage
   - Redirect to `/login` (unless already on login/register page)

3. **Proxy Architecture**: All requests go through the SvelteKit API proxy at `/api/*`. This eliminates CORS issues.

4. **Content-Type**: The default Content-Type is `application/json`. Override if needed:
   ```typescript
   await fetchAPI('/api/v1/upload', {
     method: 'POST',
     headers: { 'Content-Type': 'multipart/form-data' },
     body: formData
   });
   ```

5. **Server-Side Usage**: When using in `+page.server.ts` or `+server.ts` files, use the `fetch` from the load context, not this client (tokens are handled differently on the server).

## Backend Endpoints

See backend documentation for available endpoints. Common ones:

**Auth:**
- POST `/api/v1/auth/register` - Register new user
- POST `/api/v1/auth/login` - Login
- GET `/api/v1/auth/me` - Get current user (protected)

**Monitors:**
- GET `/api/v1/monitors` - List monitors (protected)
- POST `/api/v1/monitors` - Create monitor (protected)
- GET `/api/v1/monitors/:id` - Get monitor (protected)
- PUT `/api/v1/monitors/:id` - Update monitor (protected)
- DELETE `/api/v1/monitors/:id` - Delete monitor (protected)

**Alert Rules:**
- GET `/api/v1/alert-rules` - List alert rules (protected)
- POST `/api/v1/alert-rules` - Create alert rule (protected)
- GET `/api/v1/alert-rules/:id` - Get alert rule (protected)
- PUT `/api/v1/alert-rules/:id` - Update alert rule (protected)
- DELETE `/api/v1/alert-rules/:id` - Delete alert rule (protected)

**Alert Channels:**
- GET `/api/v1/alert-channels` - List alert channels (protected)
- POST `/api/v1/alert-channels` - Create alert channel (protected)
- GET `/api/v1/alert-channels/:id` - Get alert channel (protected)
- PUT `/api/v1/alert-channels/:id` - Update alert channel (protected)
- DELETE `/api/v1/alert-channels/:id` - Delete alert channel (protected)
