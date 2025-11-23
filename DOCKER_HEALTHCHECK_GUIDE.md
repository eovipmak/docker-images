# Docker Health Check Troubleshooting Guide

## Quick Diagnosis

When E2E tests or deployments fail, follow this checklist:

### 1. Check Container Status
```bash
docker compose ps
```

Look for containers with status other than "healthy" or "running".

### 2. Check Health Check Configuration

Each service that needs monitoring should have a health check defined:

```yaml
healthcheck:
  test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:PORT/health || exit 1"]
  interval: 10s      # How often to check
  timeout: 5s        # Max time for check to complete
  retries: 5         # Number of consecutive failures before unhealthy
  start_period: 30s  # Grace period for service startup
```

### 3. Verify Health Endpoints

Test health endpoints manually:

```bash
# Backend
curl http://localhost:8080/health

# Worker
curl http://localhost:8081/health

# Frontend (basic check)
curl http://localhost:3000
```

Expected responses:
- Backend: `{"status": "ok", "database": "connected"}`
- Worker: `{"status": "ok", "service": "worker", "database": "connected", "jobs": [...]}`
- Frontend: HTML content

### 4. Check Dependencies

Ensure required tools are installed in Docker images:

| Service  | Required Tools | Purpose |
|----------|---------------|---------|
| Backend  | wget, curl, postgresql-client | Health checks and DB access |
| Worker   | wget | Health checks |
| Frontend | None (Node.js built-in) | - |
| Postgres | pg_isready | Health checks |

### 5. Verify Service Dependencies

Services should start in the correct order:

```yaml
depends_on:
  postgres:
    condition: service_healthy  # Wait for postgres to be healthy
```

Dependency chain:
1. PostgreSQL (no dependencies)
2. Backend (depends on PostgreSQL)
3. Worker (depends on PostgreSQL)
4. Frontend (no strict dependency but needs backend to be functional)

### 6. Check Logs

If services are unhealthy, check logs:

```bash
# All services
make logs

# Specific service
make logs-backend
make logs-worker
make logs-frontend
make logs-postgres
```

Common issues to look for:
- Database connection errors
- Port binding failures
- Missing environment variables
- Migration failures
- Compilation errors (in development)

### 7. Test Health Check Command

Test the health check command inside the container:

```bash
# Backend
docker exec v-insight-backend wget --no-verbose --tries=1 --spider http://localhost:8080/health

# Worker
docker exec v-insight-worker wget --no-verbose --tries=1 --spider http://localhost:8081/health
```

If the command fails, the tool might be missing or the endpoint might not be responding.

## Common Issues and Solutions

### Issue: "wget: not found" in health check

**Symptom**: Container shows as unhealthy, logs show `wget: not found`

**Solution**: Add wget to the Dockerfile:
```dockerfile
RUN apk add --no-cache wget
```

### Issue: Health check timeout

**Symptom**: Container never becomes healthy, times out after 30 seconds

**Solution**: 
1. Increase `start_period` to give more time for initialization
2. Check if the service is actually starting (check logs)
3. Verify database migrations are completing

### Issue: Service starts but health check fails

**Symptom**: Service appears to be running but health endpoint returns errors

**Solution**:
1. Check database connectivity (most common cause)
2. Verify environment variables are set correctly
3. Test the health endpoint manually with curl

### Issue: Intermittent failures in CI/CD

**Symptom**: E2E tests pass locally but fail randomly in CI/CD

**Solution**:
1. Increase wait time after health checks pass (migrations may need more time)
2. Add explicit waits in tests for UI elements
3. Check for race conditions in test code
4. Ensure CI runner has sufficient resources

### Issue: Frontend not accessible

**Symptom**: Frontend health check fails with connection refused

**Possible causes**:
1. Node modules not installed
2. Vite dev server failed to start
3. Port binding issue
4. TypeScript compilation errors

**Solution**:
```bash
docker compose logs frontend
# Look for errors in npm install or npm run dev

# Rebuild if needed
docker compose build --no-cache frontend
```

### Issue: Database migrations fail

**Symptom**: Backend fails to start, logs show migration errors

**Solution**:
```bash
# Check migration status
make migrate-version

# Force specific version if needed
make migrate-force version=N

# Reset database (DESTRUCTIVE!)
docker compose down -v
docker compose up -d
```

## Best Practices

### 1. Always Define Health Checks

Every service that provides an HTTP endpoint should have a health check.

### 2. Use Appropriate Intervals

```yaml
interval: 10s      # Frequent checks for critical services
timeout: 5s        # Keep short to avoid blocking
retries: 5         # Allow for temporary glitches
start_period: 30s  # Adjust based on startup time
```

### 3. Implement Comprehensive Health Endpoints

Health endpoints should check:
- Service is running ✓
- Database connectivity ✓
- Critical dependencies ✓

Example (Go with Fiber):
```go
app.Get("/health", func(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
    defer cancel()
    
    if err := db.HealthContext(ctx); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "error",
            "database": "unhealthy",
        })
    }
    
    return c.JSON(fiber.Map{
        "status": "ok",
        "database": "connected",
    })
})
```

### 4. Use Proper Service Dependencies

```yaml
depends_on:
  postgres:
    condition: service_healthy  # Don't just use service_started
```

### 5. Set Restart Policies (Production)

```yaml
restart: unless-stopped  # Auto-restart on failure
```

### 6. Monitor in CI/CD

GitHub Actions workflow should:
1. Wait for health checks to pass
2. Verify each endpoint manually
3. Show logs on failure
4. Provide diagnostic information

## Testing Health Checks Locally

```bash
# 1. Start fresh
make down && make clean

# 2. Copy environment
cp .env.example .env

# 3. Start services
make up

# 4. Monitor startup
watch -n 2 'docker compose ps'

# 5. Check health when all services are up
curl http://localhost:8080/health && \
curl http://localhost:8081/health && \
echo "All services healthy!"

# 6. Run E2E tests
npx playwright test
```

## References

- [Docker Compose Healthcheck](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck)
- [Docker Depends On](https://docs.docker.com/compose/compose-file/compose-file-v3/#depends_on)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)

## Change Log

- **2024-11-23**: Added worker service health check, fixed wget availability, improved CI/CD reliability
