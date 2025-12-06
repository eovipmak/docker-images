# V-Insight Project Agents

This directory contains instructions and context for AI agents working on the V-Insight project.

## Agents

- **v-insight-architect**: High-level architecture, design patterns, and project structure.
- **backend-specialist**: Go backend development, API design, and business logic.
- **frontend-specialist**: SvelteKit frontend development, UI/UX, and client-side logic.
- **worker-specialist**: Background job processing, scheduling, and execution.
- **database-specialist**: PostgreSQL schema design, migrations, and query optimization.
- **testing-specialist**: Unit, integration, and E2E testing strategies.
- **docker-devops-specialist**: Docker, CI/CD, and deployment infrastructure.

## Usage

When prompting an AI agent, you can refer to the specific agent role to get more targeted assistance. For example:

"As the database-specialist, how should I design the schema for the new alert logs table?"

## Common Context

All agents share the following context:

- **Project**: V-Insight - Monitoring Platform
- **Architecture**: Monorepo with Go Workspace (backend, worker, shared) and SvelteKit frontend.
- **Database**: PostgreSQL
- **Deployment**: Docker Compose (Dev & Prod)
