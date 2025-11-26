# V-Insight Frontend

SvelteKit-based web interface for V-Insight monitoring platform, providing a modern, responsive UI for managing monitors, alerts, and incidents.

## Architecture

### Tech Stack

- **Framework**: SvelteKit v2
- **Language**: TypeScript
- **Build Tool**: Vite v5
- **Styling**: Tailwind CSS
- **Testing**: Playwright (E2E)
- **Adapter**: adapter-node (SSR + API proxy)
- **Package Manager**: npm

### Project Structure

```
frontend/
├── src/
│   ├── routes/                    # SvelteKit routes
│   │   ├── +layout.svelte        # Root layout (navigation, auth)
│   │   ├── +page.svelte          # Landing page
│   │   ├── api/                  # Server-side API proxy
│   │   │   └── [...path]/+server.ts   # Forwards all /api/* to backend
│   │   ├── login/                # Login page
│   │   ├── register/             # Registration page
│   │   ├── dashboard/            # Dashboard (overview)
│   │   ├── monitors/             # Monitor management
│   │   │   ├── +page.svelte     # Monitor list
│   │   │   ├── [id]/            # Monitor details
│   │   │   └── create/          # Create monitor form
│   │   ├── alerts/               # Alert rules and channels
│   │   ├── incidents/            # Incident management
│   │   ├── settings/             # User settings
│   │   └── docs/                 # API documentation (Swagger UI)
│   ├── lib/                      # Shared code
│   │   ├── components/          # Reusable Svelte components
│   │   │   ├── MonitorCard.svelte
│   │   │   ├── AlertForm.svelte
│   │   │   ├── IncidentList.svelte
│   │   │   └── ...
│   │   ├── stores/              # Svelte stores (state management)
│   │   │   ├── auth.ts          # Authentication state
│   │   │   └── monitors.ts      # Monitor state
│   │   ├── api/                 # API client functions
│   │   │   ├── client.ts        # Base HTTP client
│   │   │   ├── monitors.ts      # Monitor API calls
│   │   │   ├── alerts.ts        # Alert API calls
│   │   │   └── incidents.ts     # Incident API calls
│   │   ├── types/               # TypeScript type definitions
│   │   │   ├── monitor.ts
│   │   │   ├── alert.ts
│   │   │   └── incident.ts
│   │   └── utils/               # Helper functions
│   ├── app.html                 # HTML template
│   └── app.css                  # Global styles (Tailwind imports)
├── static/                      # Static assets (favicon, images)
├── tests/                       # Playwright E2E tests
├── svelte.config.js            # SvelteKit configuration
├── vite.config.js              # Vite configuration
├── tailwind.config.js          # Tailwind CSS configuration
├── tsconfig.json               # TypeScript configuration
├── package.json                # Dependencies and scripts
└── playwright.config.ts        # Playwright test configuration
```

## Key Features

### CORS-Free API Proxy

**CRITICAL**: This frontend uses a server-side proxy architecture that completely eliminates CORS issues.

- All API requests from browser go to `/api/*` on the same origin (e.g., `localhost:3000/api/*`)
- SvelteKit's `src/routes/api/[...path]/+server.ts` forwards these to the backend (port 8080)
- **NEVER add CORS middleware** - it's unnecessary and will break the architecture

**How it works:**

```
Browser (localhost:3000) 
  → GET /api/v1/monitors
    → SvelteKit Proxy (+server.ts)
      → Backend API (localhost:8080)
        → Response
      → Proxy forwards response
  → Browser receives response
```

### Authentication Flow

1. User logs in via `/login`
2. Backend returns JWT token
3. Token stored in localStorage
4. Token included in Authorization header for all API requests
5. `+layout.svelte` checks auth state on route changes

### Real-Time Updates

- Server-Sent Events (SSE) for live incident updates
- Automatic reconnection on connection loss
- Event listeners in dashboard and monitor detail pages

### Responsive Design

- Mobile-first approach with Tailwind CSS
- Breakpoints: sm (640px), md (768px), lg (1024px), xl (1280px)
- Touch-friendly UI elements

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Backend API running (see `/backend/README.md`)

### Installation

```bash
cd frontend

# Install dependencies
npm install
```

### Environment Variables

Create `.env` in the project root (or use existing `.env.example`):

```bash
# Frontend Configuration
FRONTEND_PORT=3000

# Backend API URL (for server-side proxy)
BACKEND_API_URL=http://backend:8080  # Docker network name
# Or for local development:
# BACKEND_API_URL=http://localhost:8080

# Vite allowed hosts (for reverse proxy setups)
VITE_ALLOWED_HOSTS=localhost,127.0.0.1
```

### Development

#### Option 1: Docker (Recommended)

```bash
# From project root
make up              # Start all services
make logs-frontend   # View frontend logs
```

Frontend available at `http://localhost:3000`

#### Option 2: Local Development

```bash
cd frontend

# Start dev server
npm run dev

# Or with specific host
npm run dev -- --host 0.0.0.0
```

**Note**: Ensure `BACKEND_API_URL` points to your running backend instance.

### Building for Production

```bash
cd frontend

# Build the application
npm run build

# Preview production build
npm run preview

# Start production server
node build
```

## Development Workflow

### Creating New Pages

1. Create route directory: `src/routes/my-page/`
2. Add `+page.svelte` for the page component
3. Optionally add `+page.ts` for data loading
4. Add navigation link in `+layout.svelte`

**Example:**

```svelte
<!-- src/routes/my-page/+page.svelte -->
<script>
  import { onMount } from 'svelte';
  
  let data = [];
  
  onMount(async () => {
    const response = await fetch('/api/v1/my-endpoint');
    data = await response.json();
  });
</script>

<h1>My Page</h1>
{#each data as item}
  <div>{item.name}</div>
{/each}
```

### Adding API Calls

1. Create function in `src/lib/api/`:

```typescript
// src/lib/api/monitors.ts
export async function getMonitors() {
  const token = localStorage.getItem('token');
  const response = await fetch('/api/v1/monitors', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  
  if (!response.ok) {
    throw new Error('Failed to fetch monitors');
  }
  
  return response.json();
}
```

2. Use in component:

```svelte
<script>
  import { getMonitors } from '$lib/api/monitors';
  import { onMount } from 'svelte';
  
  let monitors = [];
  
  onMount(async () => {
    monitors = await getMonitors();
  });
</script>
```

### Creating Components

```svelte
<!-- src/lib/components/MonitorCard.svelte -->
<script lang="ts">
  export let monitor: {
    id: string;
    name: string;
    url: string;
    enabled: boolean;
  };
</script>

<div class="bg-white rounded-lg shadow p-4">
  <h3 class="font-bold text-lg">{monitor.name}</h3>
  <p class="text-gray-600 text-sm">{monitor.url}</p>
  <span class="inline-flex items-center px-2 py-1 rounded text-xs {monitor.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
    {monitor.enabled ? 'Enabled' : 'Disabled'}
  </span>
</div>
```

### State Management with Stores

```typescript
// src/lib/stores/monitors.ts
import { writable } from 'svelte/store';

export const monitors = writable([]);

export function addMonitor(monitor) {
  monitors.update(items => [...items, monitor]);
}

export function removeMonitor(id) {
  monitors.update(items => items.filter(m => m.id !== id));
}
```

Usage in component:

```svelte
<script>
  import { monitors, addMonitor } from '$lib/stores/monitors';
</script>

{#each $monitors as monitor}
  <div>{monitor.name}</div>
{/each}
```

## Testing

### Type Checking

```bash
# Check TypeScript types
npm run check

# Check in watch mode
npm run check:watch
```

### E2E Tests (Playwright)

```bash
# Install Playwright browsers (first time only)
npx playwright install

# Run all tests
npm run test:e2e

# Run tests in UI mode
npm run test:e2e:ui

# Run tests in specific browser
npm run test:e2e -- --project=chromium

# Debug tests
npm run test:e2e -- --debug
```

### Writing Tests

```typescript
// tests/monitors.spec.ts
import { test, expect } from '@playwright/test';

test('can create a monitor', async ({ page }) => {
  await page.goto('/login');
  await page.fill('input[name="email"]', 'test@example.com');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');
  
  await page.goto('/monitors/create');
  await page.fill('input[name="name"]', 'Test Monitor');
  await page.fill('input[name="url"]', 'https://example.com');
  await page.click('button[type="submit"]');
  
  await expect(page.locator('text=Monitor created')).toBeVisible();
});
```

## Styling with Tailwind CSS

### Common Patterns

```svelte
<!-- Card -->
<div class="bg-white rounded-lg shadow p-6">
  <h2 class="text-xl font-bold mb-4">Title</h2>
  <p class="text-gray-600">Content</p>
</div>

<!-- Button -->
<button class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded">
  Click Me
</button>

<!-- Input -->
<input 
  type="text" 
  class="border border-gray-300 rounded px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500"
  placeholder="Enter value"
/>

<!-- Grid Layout -->
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  <!-- Items -->
</div>
```

### Customizing Tailwind

Edit `tailwind.config.js`:

```javascript
export default {
  theme: {
    extend: {
      colors: {
        'brand': '#1e40af',
      },
    },
  },
}
```

## Configuration

### Vite Configuration

Key settings in `vite.config.js`:

- **Port**: 3000 (configurable via env)
- **Host**: 0.0.0.0 (allows external access)
- **Polling**: Enabled for Docker file watching
- **Proxy**: Not used (handled by SvelteKit server routes)

### SvelteKit Configuration

Key settings in `svelte.config.js`:

- **Adapter**: adapter-node (for SSR and server routes)
- **Prerendering**: Disabled (dynamic authentication)
- **TypeScript**: Enabled with strict mode

### TypeScript Configuration

Key settings in `tsconfig.json`:

- **Strict mode**: Enabled
- **Path aliases**: `$lib` → `src/lib`
- **Target**: ES2020

## API Documentation Page

Access interactive API documentation at `/docs`:

- Full Swagger UI interface
- Try endpoints directly in browser
- Authentication support (click "Authorize" button)
- Automatically loads from backend `/swagger/doc.json`

## Common Tasks

### Adding a New Route

```bash
# Create route directory
mkdir -p src/routes/my-route

# Create page component
cat > src/routes/my-route/+page.svelte << 'EOF'
<h1>My New Route</h1>
EOF
```

### Debugging API Calls

1. Open browser DevTools (F12)
2. Go to Network tab
3. Filter by "Fetch/XHR"
4. Inspect request/response headers and body

### Hot Module Replacement (HMR)

Vite automatically reloads on file changes:
- `.svelte` files: Component-level HMR
- `.ts` files: Full reload
- `.css` files: Injected without reload

## Troubleshooting

### Frontend Not Loading

- Check if Vite dev server is running: `docker compose ps`
- Check frontend logs: `make logs-frontend`
- Verify port 3000 is not in use: `lsof -i :3000`

### API Calls Failing

- Verify backend is running: `curl http://localhost:8080/health`
- Check `BACKEND_API_URL` in `.env`
- Inspect browser console for errors
- Check Network tab in DevTools

### Hot Reload Not Working

- Ensure `server.watch.usePolling: true` in `vite.config.js`
- Restart frontend: `docker compose restart frontend`
- Check file permissions: `ls -la src/`

### TypeScript Errors

```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install

# Run type check
npm run check
```

### Build Errors

```bash
# Clear SvelteKit build cache
rm -rf .svelte-kit build

# Rebuild
npm run build
```

## Production Deployment

### Build Process

```bash
# Install dependencies
npm ci --production=false

# Build application
npm run build

# Result in ./build directory
```

### Environment Variables

Required in production:

```bash
FRONTEND_PORT=3000
BACKEND_API_URL=http://backend-service:8080
NODE_ENV=production
```

### Running Production Server

```bash
# Start Node server
node build

# Or with PM2 (process manager)
pm2 start build/index.js --name v-insight-frontend
```

### Reverse Proxy Setup

Example nginx configuration:

```nginx
server {
    listen 80;
    server_name monitoring.example.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Performance Optimization

### Code Splitting

SvelteKit automatically splits code by route:
- Each route loads only required JavaScript
- Shared components bundled separately

### Image Optimization

- Use WebP format where supported
- Lazy load images: `loading="lazy"`
- Provide multiple sizes for responsive images

### Caching Strategy

- Static assets: Cache for 1 year
- API responses: No cache (always fresh)
- HTML: Cache with revalidation

## Contributing

1. Run type check before committing: `npm run check`
2. Format code with Prettier (automatic on save in VSCode)
3. Follow component naming conventions (PascalCase)
4. Add JSDoc comments for complex functions
5. Test new features with Playwright

## Resources

- [SvelteKit Documentation](https://kit.svelte.dev/docs)
- [Svelte Tutorial](https://svelte.dev/tutorial)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)
