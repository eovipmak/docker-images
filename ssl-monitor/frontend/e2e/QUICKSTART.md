# E2E Testing Quick Start

This is a quick reference for running the Playwright e2e tests.

## Prerequisites

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install chromium
```

## Configuration

Create `.env.test.local` with your test credentials:
```
TEST_USER_EMAIL=your-email@example.com
TEST_USER_PASSWORD=your-password
```

## Running Tests

```bash
# Run all e2e tests
npm run test:e2e

# Run specific test
npm run test:e2e -- domain-deletion.spec.ts

# Interactive mode
npm run test:e2e:ui

# Debug mode
npm run test:e2e:debug
```

## Key Test

**Domain Deletion Test** - Verifies:
- ✅ No HTTP 405 errors when deleting domains
- ✅ No errors when disabling notifications
- ✅ Uptime bar renders correctly

See `e2e/README.md` and `TESTING.md` for full documentation.
