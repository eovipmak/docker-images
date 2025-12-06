# Contributing

Thank you for your interest in contributing to v-insight! This document explains the common workflow and CI/CD notes.

## How to contribute

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Implement changes and run tests locally
   - Backend tests: `cd backend && go test ./...`
   - Frontend checks: `cd frontend && npm run check`
4. Open a pull request against `main` and provide a clear description and testing notes.

### When submitting PRs via automation / AI agents

- Keep changes small and well-scoped — larger changes should be broken up into multiple PRs.
- Include tests that validate the change and the user isolation behavior if relevant.
- Verify the change locally:
   - Backend: `cd backend && go test ./...`
   - Frontend: `cd frontend && npm run check`
   - E2E: `cd frontend && npx playwright test` as needed
- Avoid modifying older migrations or removing historic migration files. Add new migrations instead.
- Never commit secrets or production-environment credentials — use environment variables in CI secrets.

## Testing expectations

- Unit tests and integration tests should pass locally.
- E2E tests are run in CI (Playwright). Provide deterministic tests and avoid flakiness.

## CI/CD (GitHub Actions)

The repository includes workflows to build, test, and deploy:

- Push to `main` runs tests and builds images.
- PRs can deploy to a staging environment (if configured) for UI testing.

If you are using an automated agent to create PRs, attach a note describing the CI and tests you ran to validate the change.

### Secrets

Add these to GitHub secrets if you configure auto-deploy workflows:

- `VPS_HOST`, `VPS_USER`, `SSH_PRIVATE_KEY`, `SSH_PORT` (optional)

## Code style and guidelines

- Keep changes focused and well-tested.
- Avoid committing secrets (.env files).
