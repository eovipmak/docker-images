# Testing Guide for V-Insight

This document provides comprehensive information about testing in the V-Insight project.

## Overview

V-Insight has extensive test coverage across all layers:
- **Backend Unit Tests**: Go tests for services, repositories, and utilities
- **Worker Unit Tests**: Go tests for job executors and checkers
- **Frontend Unit Tests**: Vitest tests for components, stores, and utilities
- **E2E Tests**: Playwright tests for end-to-end user flows

## Quick Start

### Run All Tests

```bash
make test-all
```

### Run Specific Test Suites

```bash
# Backend tests only
make test-backend

# Worker tests only
make test-worker

# Frontend tests only
make test-frontend
```

## Backend Testing (Go)

### Running Backend Tests

```bash
cd backend
go test ./...                    # Run all tests
go test ./... -v                 # Verbose output
go test ./... -cover             # With coverage
go test ./... -coverprofile=coverage.out  # Generate coverage file
go tool cover -html=coverage.out          # View coverage in browser
```

### Test Structure

Backend tests follow Go testing conventions:
- Test files end with `_test.go`
- Test functions start with `Test`
- Use `testify` for assertions and mocking

### Example Test

```go
func TestHashPassword(t *testing.T) {
    password := "testpassword123"
    hash, err := HashPassword(password)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.NotEqual(t, password, hash)
}
```

### Mocking

We use two approaches for mocking:
1. **sqlmock**: For repository tests with database interactions
2. **testify/mock**: For service tests with repository dependencies

Example using sqlmock:
```go
db, mock, err := sqlmock.New()
require.NoError(t, err)
defer db.Close()

sqlxDB := sqlx.NewDb(db, "sqlmock")
repo := NewUserRepository(sqlxDB)

rows := sqlmock.NewRows([]string{"id", "email"}).
    AddRow(1, "test@example.com")

mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
```

### Test Coverage Areas

- ✅ **Auth Package** (`internal/auth/`)
  - Password hashing and verification
  - JWT token generation and validation
  - Demo user password verification

- ✅ **Services** (`internal/domain/service/`)
  - Auth service (register, login, token validation)
  - Alert service (alert evaluation, incident management)
  - Monitor service (basic structure tests)
  - Metrics service (period parsing, data structures)

- ✅ **Repositories** (`internal/repository/postgres/`)
  - User repository (CRUD operations)
  - Tenant repository
  - Alert rule repository
  - Alert channel repository
  - Incident repository

- ✅ **Utilities** (`internal/utils/`)
  - Context utilities
  - Sanitization functions

## Worker Testing (Go)

### Running Worker Tests

```bash
cd worker
go test ./...                    # Run all tests
go test ./... -v                 # Verbose output
go test ./... -cover             # With coverage
```

### Test Coverage Areas

- ✅ **Executor** (`internal/executor/`)
  - HTTP checker (success, errors, timeouts, redirects)
  - SSL checker (valid certs, invalid hosts)
  - Task executor (concurrent task processing, retries)

## Frontend Testing (Vitest)

### Running Frontend Tests

```bash
cd frontend

npm test                         # Run tests once
npm run test:watch              # Watch mode
npm run test:ui                 # Interactive UI mode
npm run test:coverage           # Generate coverage report
```

### Test Structure

Frontend tests use Vitest and Testing Library:
- Test files end with `.test.ts` or `.spec.ts`
- Tests are co-located with source files
- Use `describe` and `it` for test organization

### Example Component Test

```typescript
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StatusBadge from './StatusBadge.svelte';

describe('StatusBadge', () => {
    it('should render "Open" for open status', () => {
        render(StatusBadge, { props: { status: 'open' } });
        expect(screen.getByText('Open')).toBeTruthy();
    });
});
```

### Mocking

Frontend tests mock browser APIs and fetch:

```typescript
// Mock localStorage
global.localStorage = {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn()
};

// Mock fetch
global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({ data: 'test' })
});
```

### Test Coverage Areas

- ✅ **Stores** (`src/lib/stores/`)
  - Auth store (login, logout, checkAuth, token management)

- ✅ **API Client** (`src/lib/api/`)
  - fetchAPI (auth headers, 401 handling, request options)

- ✅ **Components** (`src/lib/components/`)
  - StatusBadge (status variants, size variants, styling)

## E2E Testing (Playwright)

### Running E2E Tests

```bash
cd frontend

npm run test:e2e                # Run E2E tests
npm run test:e2e:ui             # Interactive UI mode
npm run test:e2e:report         # View test report
```

### E2E Test Guidelines

- E2E tests are in the `tests/` directory
- Test critical user flows end-to-end
- Ensure services are running before E2E tests

## Test Coverage Goals

We aim for the following coverage targets:

- **Backend**: >70% line coverage
- **Worker**: >70% line coverage
- **Frontend**: >70% line coverage
- **Critical Paths**: 100% coverage for auth, alerting, and multi-tenant isolation

## Best Practices

### General

1. **Test Behavior, Not Implementation**: Focus on what the code does, not how it does it
2. **Keep Tests Independent**: Each test should run in isolation
3. **Use Descriptive Names**: Test names should clearly describe what they test
4. **Follow AAA Pattern**: Arrange, Act, Assert
5. **Mock External Dependencies**: Don't make real HTTP calls or database queries in unit tests

### Backend Specific

1. Use `testify/assert` for assertions
2. Use `sqlmock` for database mocking
3. Clean up resources with `defer`
4. Test error cases as well as happy paths
5. Verify mock expectations with `mock.ExpectationsWereMet()`

### Frontend Specific

1. Use Testing Library queries (`getByText`, `getByRole`, etc.)
2. Avoid testing implementation details (internal state, class names unless critical)
3. Mock browser APIs appropriately
4. Clean up after tests with `beforeEach`/`afterEach`
5. Use `vi.fn()` for mocking functions

## Continuous Integration

Tests are automatically run in CI/CD pipeline. All tests must pass before merging.

### CI Test Commands

The CI pipeline runs:
```bash
make test-all
```

This ensures:
- All backend tests pass
- All worker tests pass
- All frontend tests pass
- Code coverage meets minimum thresholds

## Troubleshooting

### Backend Tests Failing

1. **Check Go version**: Ensure you're using Go 1.23+
2. **Update dependencies**: Run `go mod download`
3. **Database issues**: Repository tests use mocks, no real database needed
4. **Timeout errors**: Increase test timeout if needed

### Frontend Tests Failing

1. **Install dependencies**: Run `npm install`
2. **Clear cache**: Delete `node_modules` and run `npm install` again
3. **Check Node version**: Use Node 18+ for best compatibility
4. **Mock issues**: Ensure all browser APIs are properly mocked

### Worker Tests Failing

1. **Network tests**: SSL and HTTP checker tests may fail if no internet connection
2. **Timeout issues**: Increase test timeout for slow connections
3. **Check dependencies**: Run `go mod download`

## Adding New Tests

### Adding Backend Tests

1. Create `*_test.go` file next to the code you're testing
2. Import necessary packages: `testing`, `testify/assert`, `testify/mock`
3. Follow the existing test patterns
4. Run tests with `go test ./...`

### Adding Frontend Tests

1. Create `*.test.ts` file next to the component/module you're testing
2. Import Vitest and Testing Library utilities
3. Follow the existing test patterns
4. Run tests with `npm test`

### Adding E2E Tests

1. Create test file in `frontend/tests/` directory
2. Use Playwright API
3. Test complete user flows
4. Run with `npm run test:e2e`

## Coverage Reports

### View Backend Coverage

```bash
cd backend
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### View Frontend Coverage

```bash
cd frontend
npm run test:coverage
```

Coverage reports will be generated in the console and as HTML files.

## Contributing

When adding new features:

1. Write tests for new functionality
2. Ensure existing tests still pass
3. Maintain or improve coverage percentages
4. Update this documentation if needed

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [Playwright Documentation](https://playwright.dev/)
