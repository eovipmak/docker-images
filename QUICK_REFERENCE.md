# Quick Reference: E2E Testing Implementation

## Demo User Credentials
```
Email: test@gmail.com
Password: Password!
Tenant: Demo Tenant
```

## Running Tests

### Locally
```bash
# Prerequisites
npm ci
npx playwright install --with-deps

# Run all tests
npx playwright test

# Run workflow test only
npx playwright test tests/e2e-workflow.spec.ts

# Debug mode with UI
npx playwright test --ui
```

### In GitHub Actions
- Automatically runs on push to main/master
- Automatically runs on PRs to main/master
- Manually trigger via workflow_dispatch
- Comment `/test` on a PR to trigger

## File Structure

### Created Files
- `backend/migrations/000008_seed_demo_user.up.sql` - Demo user migration
- `backend/migrations/000008_seed_demo_user.down.sql` - Rollback
- `backend/internal/auth/demo_user_test.go` - Password hash verification
- `tests/e2e-workflow.spec.ts` - Main E2E test suite
- `.github/workflows/e2e-tests.yml` - Unified test workflow
- `TESTING.md` - Comprehensive testing guide

### Modified Files
- `playwright.config.ts` - Enhanced configuration
- `README.md` - Added testing section
- `.github/workflows/manual-test.yml` - Disabled (commented out)
- `.github/workflows/playwright.yml` - Disabled (commented out)

## Test Scenarios

1. ✅ Login with demo user
2. ✅ Add monitor (domain) for google.com
3. ✅ Edit monitor (interval, timeout, name)
4. ✅ Add alert rule
5. ✅ Edit alert rule
6. ✅ Complete workflow (all steps in sequence)

## Screenshot Naming

Individual tests: `01-login-form.png`, `02-dashboard.png`, etc.
Workflow test: `workflow-01-login.png`, `workflow-02-dashboard.png`, etc.

## Verification

To verify the demo user was created:
```bash
docker compose exec postgres psql -U postgres -d v_insight \
  -c "SELECT email, created_at FROM users WHERE email = 'test@gmail.com';"
```

To verify password hash:
```bash
cd backend
go test ./internal/auth -v -run TestDemoUserPasswordHash
```

## Workflow Triggers

The e2e-tests.yml workflow runs when:
- Code is pushed to main or master branches
- Pull request is opened/updated to main or master
- Manually dispatched from GitHub Actions tab
- Comment `/test` is added to a pull request

## Artifacts

After each workflow run, download:
- `playwright-test-results` - Full test results, traces, videos
- `test-screenshots` - All captured screenshots

## Common Commands

```bash
# Start services
make up

# Check service health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:3000

# View logs
make logs-backend
make logs-worker
make logs-frontend

# Run specific test
npx playwright test -g "Login with demo user"

# View test report
npx playwright show-report
```

## Important Notes

⚠️ **Security**: Demo user is for testing only. Remove in production.
⚠️ **Workflow**: Old manual-test.yml and playwright.yml are disabled.
✅ **Migration**: Demo user is created automatically on database init.
✅ **Screenshots**: Captured at every step for debugging.
✅ **CI/CD**: Fully automated in GitHub Actions.
