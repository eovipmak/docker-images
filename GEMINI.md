# V-Insight Project

This document provides a comprehensive overview of the V-Insight project, its architecture, and key development workflows.

## Project Overview

V-Insight is a web-based monitoring platform designed to provide real-time insights into the health and performance of web services. It consists of three main components:

*   **Backend:** A Go-based REST API that handles data storage, authentication, and business logic.
*   **Frontend:** A SvelteKit-based web interface that provides a user-friendly dashboard for managing monitors, viewing incidents, and configuring alerts.
*   **Worker:** A Go-based background worker that performs health checks, SSL certificate monitoring, and sends notifications.

The entire project is containerized using Docker, which simplifies development and deployment.

## Architecture

*   **Backend:**
    *   **Language:** Go
    *   **Framework:** Gin
    *   **Database:** PostgreSQL
    *   **Authentication:** JWT
*   **Frontend:**
    *   **Framework:** SvelteKit
    *   **Language:** TypeScript
    *   **Styling:** Tailwind CSS
*   **Worker:**
    *   **Language:** Go
    *   **Job Scheduler:** robfig/cron

## Building and Running

The project uses a `Makefile` to streamline common development tasks.

### Prerequisites

*   Docker
*   Docker Compose

### Development

To start all services for development, run the following command from the project root:

```bash
make up
```

This will start the backend, frontend, worker, and a PostgreSQL database in Docker containers.

*   Frontend is available at `http://localhost:3000`
*   Backend is available at `http://localhost:8080`
*   Worker is available at `http://localhost:8081`

To stop all services, run:

```bash
make down
```

### Testing

The project has a suite of tests for each component.

*   **Run all tests:**

    ```bash
    make test-all
    ```

*   **Run backend tests:**

    ```bash
    make test-backend
    ```

*   **Run worker tests:**

    ```bash
    make test-worker
    ```

*   **Run frontend tests:**

    ```bash
    make test-frontend
    ```

## Development Conventions

*   **User Isolation:** The backend enforces user isolation. All persisted objects include a `user_id`.
*   **CORS:** The frontend uses a server-side proxy to handle API requests, so there is no need to configure CORS in the backend.
*   **Database Migrations:** Database migrations are located in the `backend/migrations` directory and are run automatically at startup. New migrations should be created using the `make migrate-create` command.
*   **API Documentation:** The backend API is documented using Swagger. The documentation is available at `/swagger/` in the development environment.
*   **Code Style:** Follow the existing code style and conventions in the codebase.
*   **Commit Messages:** Follow the conventional commit format.
*   **Pull Requests:** Create a pull request for any new features or bug fixes.
