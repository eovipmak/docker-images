# SSE Configuration Guide

## Overview

The SSE (Server-Sent Events) feature requires proper configuration to connect the frontend to the backend API.

## Configuration

### Environment Variable: PUBLIC_API_URL

The `PUBLIC_API_URL` environment variable tells the frontend where to connect for SSE events. This is the URL that the **browser** uses to connect to the backend API.

### Local Development

For local development, the default configuration works automatically:

```bash
# frontend/.env (or use .env.example)
PUBLIC_API_URL=http://localhost:8080
```

The frontend will auto-detect and connect to `http://localhost:8080` for SSE.

### Production / Staging Deployment

When deploying to a VPS or server, you **MUST** set the PUBLIC_API_URL to match your backend URL:

#### Option 1: Using Environment Variables (Recommended)

Set the environment variable before starting:

```bash
# For VPS deployment with IP address
export PUBLIC_API_URL=http://YOUR_VPS_IP:8081

# Start services
docker compose -f docker-compose.staging.yml up -d
```

#### Option 2: In docker-compose file

Edit `docker-compose.staging.yml`:

```yaml
frontend:
  environment:
    - PUBLIC_API_URL=http://YOUR_VPS_IP:8081  # Replace with your actual IP
```

#### Option 3: Using .env file

Create a `.env` file in the project root:

```bash
PUBLIC_API_URL=http://YOUR_VPS_IP:8081
```

### With Reverse Proxy / Domain

If you're using a reverse proxy (nginx, traefik, etc.) or have a domain:

```bash
# With subdomain for API
PUBLIC_API_URL=https://api.yourdomain.com

# Or with same domain, different path
PUBLIC_API_URL=https://yourdomain.com/api
```

**Important**: The PUBLIC_API_URL must be accessible from the user's browser, not just from inside the Docker network.

## Port Configuration

### Default Ports

- **Development**:
  - Frontend: `3000`
  - Backend: `8080`
  - Worker: `8081`

- **Staging** (docker-compose.staging.yml):
  - Frontend: `3001` (exposed)
  - Backend: `8081` (exposed)
  - Worker: `8082` (exposed)

- **Production** (docker-compose.prod.yml):
  - Frontend: `3000` (exposed)
  - Backend: `8080` (exposed)
  - Worker: `8081` (exposed)

### SSE Connection Flow

```
Browser → PUBLIC_API_URL/api/v1/stream/events → Backend SSE Handler
```

Example for staging on VPS with IP `123.45.67.89`:

```
Browser → http://123.45.67.89:8081/api/v1/stream/events → Backend
```

## Troubleshooting

### SSE not connecting

1. **Check browser console** for connection errors
2. **Verify PUBLIC_API_URL** is set correctly:
   ```javascript
   // In browser console
   console.log(window.__PUBLIC_API_URL__);
   ```
3. **Check network tab** - you should see a request to `/api/v1/stream/events`
4. **Verify backend is accessible** from your browser:
   ```bash
   curl http://YOUR_VPS_IP:8081/health
   ```

### Connection refused / CORS errors

- Make sure PUBLIC_API_URL points to the externally accessible backend URL
- Check firewall rules allow access to the backend port
- Verify the backend service is running: `docker compose ps`

### Auto-detection fallback

If PUBLIC_API_URL is not set, the frontend will attempt to auto-detect:

- **Localhost**: Connects to `http://localhost:8080`
- **Production (port 3001)**: Connects to `http://hostname:8081`
- **Production (other port)**: Uses current origin

This may not work in all deployment scenarios, so it's recommended to explicitly set PUBLIC_API_URL.

## Examples

### Staging Deployment

```bash
# Set environment variable
export PUBLIC_API_URL=http://123.45.67.89:8081

# Deploy
docker compose -f docker-compose.staging.yml up -d

# Verify
curl http://123.45.67.89:8081/health
# Open http://123.45.67.89:3001 in browser
# Check console for SSE connection logs
```

### Production with Domain

```bash
# In .env file
PUBLIC_API_URL=https://api.example.com

# Or with nginx reverse proxy to same domain
PUBLIC_API_URL=https://example.com

# Deploy
docker compose -f docker-compose.prod.yml up -d
```

### Production with IP (no domain)

```bash
# In .env file
PUBLIC_API_URL=http://YOUR_SERVER_IP:8080

# Deploy
docker compose -f docker-compose.prod.yml up -d
```

## Security Notes

1. **HTTPS in Production**: Always use HTTPS in production for secure token transmission
2. **Firewall**: Ensure backend port is accessible from expected client IPs
3. **Token in URL**: SSE passes the auth token in the URL query parameter (EventSource limitation)
4. **CORS**: Not needed since PUBLIC_API_URL should point directly to the backend

## Verification

After deployment, verify SSE is working:

1. Open the frontend in a browser
2. Login to the dashboard
3. Open browser DevTools Console (F12)
4. Look for logs:
   ```
   [SSE] Connecting to event stream...
   [SSE] Connecting to: http://YOUR_URL/api/v1/stream/events?token=***
   [SSE] Connected to event stream
   [SSE] Connection established: ...
   ```
5. Create a monitor and wait for health checks (30-60 seconds)
6. Check console for: `[SSE] Monitor check event: ...`
