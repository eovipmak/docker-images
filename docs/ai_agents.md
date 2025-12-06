# AI Agents & Automation Guidelines

This document is intended to help AI agents (e.g., LLM assistants, CI bots) and automation systems interact with the v-insight repository safely and effectively.

## High-level goals for agents

- Provide accurate code changes that follow project conventions and user isolation.
- Avoid security regressions, secret leaks, or misconfiguration for production environments.
- Ensure tests pass for any code change; run unit tests and lightweight CI steps locally.

## Important files and locations

- API proxy (Frontend): `frontend/src/routes/api/[...path]/+server.ts`
  - The frontend proxies requests to the backend to avoid adding CORS middleware.
- Backend API: `backend/cmd/api` and `backend/internal/api` for handlers and routes.
- Worker: `worker/cmd/worker` and `worker/internal/jobs` for scheduled jobs.
- Migrations: `backend/migrations/` — add both `.up.sql` and `.down.sql` when modifying schema.
- Tests: `backend/`, `worker/` (`go test ./...`), `frontend/` (TypeScript checks and Playwright e2e)

## Rules of engagement

- User safety: All data writes or reads must be scoped by `user_id`. Verify request handlers and repository calls include this filter.
- No CORS middleware: Avoid adding or modifying CORS handling; the proxy is the recommended pattern for browser clients.
- Secrets: Never hardcode secrets (JWT secrets, DB passwords) in code or commit them to the repo.
- Migrations: Create both `.up.sql` and `.down.sql` and use the `make migrate-create` helper. Agents should not remove or modify older migrations.
- Tests: Add or update unit tests for backend/business logic changes. Run `go test` and relevant frontend checks locally before proposing PRs.

## CI and PR requirements

- Run the test suite relevant to your change set. For backend logic changes run: `cd backend && go test ./...`.
- If your change touches the frontend, run `cd frontend && npm run check`.
- If your change impacts E2E or user flows, ensure Playwright tests are updated accordingly and remain deterministic.

## Minimal commands for local validation (for agents)

```bash
# Backend tests
cd backend && go test ./...

# Worker tests
cd worker && go test ./...

# Frontend checks and e2e
cd frontend && npm run check
cd frontend && npx playwright test
```

## Example agent operations

- Safe schema update:
  - Create migration with `make migrate-create name=add_new_field` and edit files for up/down.
  - Run migrations locally: `make migrate-up`.
  - Add repository method and tests.

- Adding a new alert type:
  - Update domain types in `backend/internal/domain/entities` and repository interfaces.
  - Update the evaluator service and add unit tests for expected behavior.
  - Update the OpenAPI documentation in `backend/docs/swagger` if applicable.

## Safety and review prompts

- Agents should create well-scoped, small PRs and include test coverage and thorough descriptions.
- Ask reviewers to validate user isolation behavior and verify no sensitive data is exposed.

## Where to ask for help

- Open an issue with reproduction steps and the expected result.
- Use the repository issue tracker to request architectural input from maintainers.

---

This document should be a starting point for automation to integrate smoothly with repository patterns and expectations.
