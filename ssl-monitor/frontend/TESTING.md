# Testing Guide

This document provides comprehensive information about testing the SSL Monitor frontend application.

## Table of Contents

- [Overview](#overview)
- [Unit Tests](#unit-tests)
- [End-to-End Tests](#end-to-end-tests)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Continuous Integration](#continuous-integration)

## Overview

The SSL Monitor frontend uses two types of tests:

1. **Unit Tests** - Test individual components and functions in isolation using Vitest
2. **E2E Tests** - Test complete user workflows using Playwright

## Unit Tests

Unit tests are located in the `src/tests/` directory and use:
- **Vitest** - Fast unit test framework
- **React Testing Library** - Testing utilities for React components
- **jsdom** - DOM implementation for Node.js

### Running Unit Tests

```bash
# Run all unit tests
npm test

# Run tests in watch mode (development)
npm test -- --watch

# Run tests with UI
npm run test:ui

# Run tests with coverage
npm test -- --coverage
```

### Writing Unit Tests

Example unit test structure:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import MyComponent from '../components/MyComponent';

describe('MyComponent', () => {
  it('should render correctly', () => {
    render(<MyComponent />);
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });
});
```

### Existing Unit Tests

- `Dashboard.test.tsx` - Tests for the main dashboard component
- `AlertsDisplay.test.tsx` - Tests for the alerts display component
- `SSLCheckForm.test.tsx` - Tests for the SSL check form

## End-to-End Tests

E2E tests are located in the `e2e/` directory and use:
- **Playwright** - Modern end-to-end testing framework
- **Chromium** - Browser for test execution

### Running E2E Tests

**Prerequisites:**
- The application must be running locally or specify a different base URL
- Test user credentials must be configured

```bash
# Install Playwright browsers (first time only)
npx playwright install chromium

# Run all e2e tests
npm run test:e2e

# Run e2e tests in UI mode (interactive)
npm run test:e2e:ui

# Run e2e tests in debug mode (step-by-step)
npm run test:e2e:debug

# View test report
npm run test:e2e:report
```

### Configuring E2E Tests

Create a `.env.test.local` file (copy from `.env.test`):

```bash
# Test User Credentials
TEST_USER_EMAIL=your-test-user@example.com
TEST_USER_PASSWORD=your-test-password

# Application URL (default: http://localhost:3000)
PLAYWRIGHT_BASE_URL=http://localhost:3000
```

### Existing E2E Tests

#### Domain Deletion Test (`domain-deletion.spec.ts`)

This test verifies the domain deletion workflow and addresses a critical issue where HTTP 405 errors were occurring after deleting domains and disabling notifications.

**Test Coverage:**
1. User authentication and login
2. Adding a test domain to monitor
3. Disabling notifications/alerts for a domain
4. Deleting a domain
5. Verifying no HTTP 405 errors occur during the workflow
6. Validating uptime bar visibility and rendering

**Network Monitoring:**
- Tracks all HTTP requests and responses
- Logs errors (4xx, 5xx status codes)
- Specifically detects and reports HTTP 405 (Method Not Allowed) errors
- Captures response details for debugging

**Running the Domain Deletion Test:**

```bash
# Run only the domain deletion test
npm run test:e2e -- domain-deletion.spec.ts

# Run in debug mode
npm run test:e2e:debug -- domain-deletion.spec.ts
```

## Running Tests

### Running All Tests

```bash
# Run unit tests
npm test

# Run e2e tests (requires app to be running)
npm run test:e2e
```

### Running Specific Tests

```bash
# Run specific unit test file
npm test -- Dashboard.test.tsx

# Run specific e2e test file
npm run test:e2e -- domain-deletion.spec.ts

# Run tests matching a pattern
npm test -- --grep "dashboard"
```

### Test Output

**Unit Tests:**
- Console output shows pass/fail status
- Coverage reports can be generated
- HTML UI available with `npm run test:ui`

**E2E Tests:**
- Console output shows detailed step-by-step progress
- Screenshots captured on failure
- Videos recorded on failure
- HTML report generated after test run

## Writing Tests

### Best Practices

1. **Isolation** - Each test should be independent
2. **Descriptive Names** - Use clear, descriptive test names
3. **Arrange-Act-Assert** - Follow the AAA pattern
4. **Mock External Dependencies** - Mock API calls and external services
5. **Clean Up** - Clean up resources after tests
6. **Avoid Hardcoded Waits** - Use proper wait conditions instead of `setTimeout`

### Unit Test Best Practices

```typescript
// ✅ Good - descriptive name, clear assertions
it('should display error message when API call fails', async () => {
  vi.mocked(api.getData).mockRejectedValue(new Error('API Error'));
  render(<MyComponent />);
  await waitFor(() => {
    expect(screen.getByText('API Error')).toBeInTheDocument();
  });
});

// ❌ Bad - vague name, no proper waiting
it('test component', () => {
  render(<MyComponent />);
  expect(screen.getByText('Hello')).toBeInTheDocument();
});
```

### E2E Test Best Practices

```typescript
// ✅ Good - uses test.step for clarity, proper waiting
await test.step('Login to application', async () => {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('input[type="email"]', email);
  await page.fill('input[type="password"]', password);
  await page.click('button[type="submit"]');
  await page.waitForURL('/dashboard');
});

// ❌ Bad - no test.step, hardcoded timeout
await page.goto('/login');
await page.fill('input[type="email"]', email);
await page.fill('input[type="password"]', password);
await page.click('button[type="submit"]');
await page.waitForTimeout(5000); // Don't use fixed waits
```

## Continuous Integration

Tests are designed to run in CI/CD pipelines:

### GitHub Actions Example

```yaml
- name: Install dependencies
  run: npm ci

- name: Run unit tests
  run: npm test -- --run

- name: Install Playwright browsers
  run: npx playwright install --with-deps chromium

- name: Start application
  run: npm run dev &
  
- name: Wait for app to be ready
  run: npx wait-on http://localhost:3000

- name: Run E2E tests
  run: npm run test:e2e

- name: Upload test artifacts
  if: always()
  uses: actions/upload-artifact@v3
  with:
    name: playwright-report
    path: playwright-report/
```

## Troubleshooting

### Common Issues

**Unit Tests:**

1. **Tests timeout** - Increase timeout in vitest config
2. **Mock not working** - Ensure mocks are set up before importing components
3. **Async issues** - Use `waitFor` for async operations

**E2E Tests:**

1. **Browser not installed**
   ```bash
   npx playwright install chromium
   ```

2. **Authentication fails**
   - Verify TEST_USER_EMAIL and TEST_USER_PASSWORD in `.env.test.local`
   - Ensure test user exists in the database

3. **Timeout errors**
   - Increase timeout in `playwright.config.ts`
   - Check if the application is running
   - Verify network connectivity

4. **Selector not found**
   - Use Playwright Inspector to debug: `npm run test:e2e:debug`
   - Update selectors to match current UI

## Test Coverage

### Current Coverage

- Dashboard component - ✅ Unit tests
- Alerts display - ✅ Unit tests
- SSL check form - ✅ Unit tests
- Domain deletion workflow - ✅ E2E test
- Notification management - ✅ E2E test
- Uptime bar rendering - ✅ E2E test

### Adding Coverage

To add test coverage for new features:

1. **Unit Tests** - Add tests in `src/tests/` for new components
2. **E2E Tests** - Add tests in `e2e/` for new user workflows
3. **Update Documentation** - Update this guide with new test information

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Playwright Documentation](https://playwright.dev/)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)

## Support

For questions or issues with tests:
1. Check this documentation
2. Review existing tests for examples
3. Check the test framework documentation
4. Open an issue in the repository
