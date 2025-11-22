# E2E Testing Guide

This document describes the automated end-to-end testing setup for V-Insight.

## Demo User

A demo user is automatically created via database migrations for testing purposes.

### Credentials

- **Email**: `test@gmail.com`
- **Password**: `Password!`
- **Tenant**: Demo Tenant (automatically created)

### Usage

The demo user is created by the migration file `backend/migrations/000008_seed_demo_user.up.sql` and is available immediately after the services start.

You can use this account for:
- Manual testing of the application
- Running automated E2E tests
- Development and debugging
- Demo purposes

### Security Note

⚠️ **Important**: This demo user is created with a well-known password for testing purposes only. In production environments, you should:
- Remove this migration or disable it
- Use proper user registration and authentication
- Never use weak or well-known passwords

## Automated E2E Tests

### Test Suite Location

The main E2E test suite is located at: `tests/e2e-workflow.spec.ts`

### Test Scenarios

The test suite includes the following scenarios:

1. **Login Test**: Validates login with demo user credentials
2. **Add Monitor**: Creates a new monitor for google.com
3. **Edit Monitor**: Updates monitor properties (name, interval, timeout)
4. **Add Alert Rule**: Creates an alert rule for the monitor
5. **Edit Alert Rule**: Updates alert rule properties
6. **Complete Workflow**: Runs all steps in sequence

### Running Tests Locally

#### Prerequisites

```bash
# Install dependencies (if not already done)
npm ci

# Install Playwright browsers
npx playwright install --with-deps chromium
```

#### Run Tests

```bash
# Run all tests
npx playwright test

# Run specific test file
npx playwright test tests/e2e-workflow.spec.ts

# Run tests with UI for debugging
npx playwright test --ui

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run specific test by name
npx playwright test -g "Login with demo user"
```

#### View Test Results

```bash
# Show HTML report
npx playwright show-report

# Test results are also saved in:
# - test-results/ (screenshots, videos, traces)
# - playwright-report/ (HTML report)
```

### GitHub Actions Workflow

The E2E tests run automatically on GitHub Actions via `.github/workflows/e2e-tests.yml`

#### When Tests Run

- On push to `main` or `master` branches
- On pull requests to `main` or `master`
- Manual workflow dispatch
- When a comment containing `/test` is added to a PR

#### Workflow Steps

1. Checkout code
2. Set up Docker Buildx
3. Copy environment file
4. Build and start all services (backend, worker, frontend, PostgreSQL)
5. Wait for services to be healthy (60 seconds)
6. Set up Node.js
7. Install dependencies and Playwright browsers
8. Run E2E tests
9. Upload test results and screenshots as artifacts
10. Show service logs on failure

#### Viewing Results

After a workflow run:
1. Go to the Actions tab in GitHub
2. Click on the workflow run
3. Download artifacts:
   - `playwright-test-results`: Full test results and traces
   - `test-screenshots`: All captured screenshots

### Test Screenshots

The test suite captures screenshots at each major step:

**Individual Tests:**
- `01-login-form.png` - Login page before submission
- `02-dashboard.png` - Dashboard after login
- `03-domains-page.png` - Monitors/domains listing
- `04-add-monitor-form.png` - Add monitor form
- `05-monitor-created.png` - Monitor list after creation
- `06-monitor-detail.png` - Monitor detail page
- `07-edit-monitor-form.png` - Edit monitor form
- `08-monitor-updated.png` - Monitor after update
- `09-alerts-page.png` - Alerts listing
- `10-add-alert-form.png` - Add alert form
- `11-alert-created.png` - Alerts list after creation
- `12-alert-detail.png` - Alert detail page
- `13-edit-alert-form.png` - Edit alert form
- `14-alert-updated.png` - Alert after update

**Complete Workflow Test:**
- `workflow-01-login.png` through `workflow-10-alert-edited.png`

## Test Configuration

### Playwright Configuration

The Playwright configuration is in `playwright.config.ts`:

- **testDir**: `./tests`
- **fullyParallel**: `false` (tests run sequentially for E2E)
- **retries**: 2 on CI, 0 locally
- **workers**: 1 on CI
- **baseURL**: `http://localhost:3000` (or from BASE_URL env var)
- **timeout**: 10 seconds (configurable per test)
- **screenshot**: Captured on failure
- **video**: Retained on failure
- **trace**: On first retry

### Environment Variables

You can customize test behavior with environment variables:

```bash
# Set custom base URL
BASE_URL=http://localhost:3000 npx playwright test

# Run in CI mode
CI=true npx playwright test
```

## Troubleshooting

### Tests Failing Locally

1. **Services not running**: Make sure all services are up
   ```bash
   make up
   docker compose ps
   ```

2. **Database not ready**: Wait for migrations to complete
   ```bash
   make logs-backend | grep -i migration
   ```

3. **Demo user not created**: Check if migration ran
   ```bash
   docker compose exec postgres psql -U postgres -d v_insight -c "SELECT email FROM users WHERE email = 'test@gmail.com';"
   ```

4. **Frontend not accessible**: Verify frontend is running
   ```bash
   curl http://localhost:3000
   ```

### Tests Failing in CI

1. Check the workflow logs in GitHub Actions
2. Download and review the test artifacts
3. Look at the screenshots to see where the test failed
4. Check service logs (shown on failure)

### Common Issues

**Timeout errors**: 
- Services might take longer to start
- Increase wait time in workflow or test
- Check if services are healthy

**Element not found**:
- UI might have changed
- Update selectors in test
- Check screenshots to see actual UI state

**Flaky tests**:
- Add explicit waits for dynamic content
- Use `page.waitForLoadState('networkidle')`
- Increase timeouts for slow operations

## Extending Tests

### Adding New Test Scenarios

1. Create a new test in `tests/e2e-workflow.spec.ts` or a new file:

```typescript
test('Your new test scenario', async ({ page }) => {
  // Login first (if needed)
  await page.goto('http://localhost:3000/login');
  await page.locator('input[name="email"]').fill('test@gmail.com');
  await page.locator('input[name="password"]').fill('Password!');
  await page.locator('button[type="submit"]').click();
  await page.waitForURL('**/dashboard');

  // Your test logic here
  await page.screenshot({ path: 'test-results/my-test.png' });
  
  // Assertions
  await expect(page.locator('...')).toBeVisible();
});
```

2. Run the test to verify it works:
```bash
npx playwright test -g "Your new test scenario"
```

3. Commit and push to trigger CI

### Best Practices

- **Use descriptive test names**: Make it clear what the test validates
- **Take screenshots**: Capture important states for debugging
- **Wait for stability**: Use `waitForLoadState('networkidle')` after navigation
- **Use flexible selectors**: Prefer role-based or text-based selectors
- **Handle dynamic content**: Add appropriate waits and timeouts
- **Clean up**: Each test should be independent
- **Document**: Add comments for complex test logic

## Migration Information

### How the Demo User is Created

The demo user is created by the migration file:
`backend/migrations/000008_seed_demo_user.up.sql`

This migration:
1. Checks if the user already exists
2. Creates the user with a bcrypt-hashed password
3. Creates a tenant named "Demo Tenant"
4. Associates the user with the tenant as owner

### Rollback

To remove the demo user (rollback migration):

```bash
make migrate-down
```

This will execute `000008_seed_demo_user.down.sql` which deletes the demo user and associated tenant (cascade will remove all related data).

### Verifying the Password Hash

A test is included to verify the password hash works correctly:
`backend/internal/auth/demo_user_test.go`

Run it with:
```bash
cd backend
go test ./internal/auth -v -run TestDemoUserPasswordHash
```

## Support

If you encounter issues with the tests:

1. Check the troubleshooting section above
2. Review the test logs and screenshots
3. Check GitHub Actions workflow logs
4. Open an issue with details and screenshots
