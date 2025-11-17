# E2E Tests for V-Insight

This directory contains end-to-end (E2E) tests using Playwright.

## Prerequisites

- Docker and Docker Compose installed
- Node.js and npm installed
- Services must be running (backend, frontend, database)

## Running Tests

### 1. Start the services

```bash
# From the project root
make up
# OR
docker compose up -d
```

Wait for all services to be healthy (~30 seconds).

### 2. Run the tests

```bash
# From the frontend directory
cd frontend

# Run all E2E tests
npm run test:e2e

# Run tests with UI (interactive mode)
npm run test:e2e:ui

# View test report
npm run test:e2e:report
```

## Test Coverage

### Authentication Tests (`auth.spec.ts`)

✅ **User Registration**
- Creates a new user account
- Validates that a JWT token is returned
- Tests email and password validation

✅ **User Login**
- Authenticates existing user
- Receives JWT token on successful login
- Tests error handling for invalid credentials

✅ **Protected Endpoints**
- `/auth/me` endpoint requires authentication
- Rejects requests without token (401)
- Rejects requests with invalid token (401)
- Accepts requests with valid token (200)
- Returns user data for authenticated requests

## Test Structure

```
tests/
└── e2e/
    └── auth.spec.ts  - Authentication flow tests
```

## Configuration

Test configuration is in `playwright.config.ts`:
- Base URL: `http://localhost:3000`
- Browser: Chromium
- Tests run serially (one at a time)
- Timeout: 60 seconds per test

## Troubleshooting

### Tests fail with "ECONNREFUSED"
Make sure all services are running:
```bash
docker compose ps
curl http://localhost:8080/health
```

### Database errors
Reset the database:
```bash
docker compose down -v
docker compose up -d
```

### Permission errors
Fix npm permissions:
```bash
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit
```

## CI/CD Integration

These tests can be integrated into CI/CD pipelines. The tests are designed to:
- Run in headless mode by default
- Retry failed tests (in CI environment)
- Generate HTML reports
- Exit with proper status codes
