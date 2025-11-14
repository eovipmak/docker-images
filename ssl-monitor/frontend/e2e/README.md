# E2E Tests with Playwright

This directory contains end-to-end (e2e) tests for the SSL Monitor application using [Playwright](https://playwright.dev/).

## Overview

The e2e tests verify critical user workflows and ensure the application functions correctly from a user's perspective. These tests run in a real browser environment and interact with the application just like a real user would.

## Test Files

- `domain-deletion.spec.ts` - Tests for domain deletion workflow and notification management
  - Verifies no HTTP 405 errors occur when deleting domains
  - Validates notification/alert disabling functionality
  - Checks uptime bar visibility and rendering

## Running the Tests

### Prerequisites

1. Ensure all dependencies are installed:
   ```bash
   npm install
   ```

2. Install Playwright browsers (if not already installed):
   ```bash
   npx playwright install chromium
   ```

### Running Tests

Run all e2e tests:
```bash
npm run test:e2e
```

Run tests in UI mode (interactive):
```bash
npm run test:e2e:ui
```

Run tests in debug mode:
```bash
npm run test:e2e:debug
```

View test report:
```bash
npm run test:e2e:report
```

### Environment Variables

Configure the following environment variables to customize test execution:

- `TEST_USER_EMAIL` - Email for test user authentication (default: `test@example.com`)
- `TEST_USER_PASSWORD` - Password for test user (default: `testpassword123`)
- `PLAYWRIGHT_BASE_URL` - Base URL of the application (default: `http://localhost:3000`)

Example:
```bash
TEST_USER_EMAIL=admin@example.com TEST_USER_PASSWORD=secure123 npm run test:e2e
```

## Test Structure

Each test follows a clear structure:

1. **Setup** - Login and navigate to the relevant page
2. **Action** - Perform user actions (click, type, etc.)
3. **Verification** - Assert expected outcomes
4. **Cleanup** - Reset state for next test (if needed)

## Network Monitoring

The domain deletion tests include comprehensive network monitoring to:
- Track all HTTP requests and responses
- Log errors (4xx, 5xx status codes)
- Specifically detect and report HTTP 405 errors
- Capture response details for debugging

## Test Artifacts

When tests run, the following artifacts are generated:

- **Screenshots** - Captured on test failure
- **Videos** - Recorded when tests fail (retention mode: on-failure)
- **Traces** - Playwright trace files for debugging (on first retry)
- **HTML Report** - Comprehensive test report with results

Artifacts are stored in:
- `test-results/` - Individual test results
- `playwright-report/` - HTML report

## Writing New Tests

1. Create a new `.spec.ts` file in the `e2e/` directory
2. Import required helpers:
   ```typescript
   import { test, expect } from '@playwright/test';
   import { AuthHelper } from './helpers/auth.helper';
   ```
3. Follow the existing test patterns for consistency
4. Use descriptive test names and step descriptions
5. Add console logging for better debugging

## Best Practices

- **Isolation** - Each test should be independent and not rely on other tests
- **Cleanup** - Reset state after tests to avoid side effects
- **Selectors** - Use accessible selectors (role, text, label) when possible
- **Waits** - Use `waitForLoadState`, `waitForSelector` instead of fixed timeouts
- **Error Handling** - Expect and handle error scenarios gracefully
- **Logging** - Add console logs for debugging and test progress tracking

## Debugging Failed Tests

1. **Review Screenshots** - Check `test-results/` for failure screenshots
2. **Watch Videos** - Review recorded videos of test execution
3. **Inspect Traces** - Use `npx playwright show-trace <trace-file>` to debug
4. **Run in Debug Mode** - Use `npm run test:e2e:debug` for step-by-step execution
5. **Check Logs** - Review console output for error messages and request details

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:
- Tests run in headless mode by default
- Automatic retries on failure (2 retries in CI)
- Parallel execution disabled in CI for stability
- All artifacts are preserved for debugging

## Troubleshooting

### Common Issues

**Issue**: Tests fail with "Timeout waiting for page load"
- **Solution**: Increase timeout in `playwright.config.ts` or specific test

**Issue**: Authentication fails
- **Solution**: Verify `TEST_USER_EMAIL` and `TEST_USER_PASSWORD` are correct

**Issue**: Elements not found
- **Solution**: Check if selectors match the current UI, update if needed

**Issue**: Network errors during test
- **Solution**: Ensure the dev server is running and accessible

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Tests](https://playwright.dev/docs/debug)
