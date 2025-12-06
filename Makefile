.PHONY: up down logs rebuild clean help migrate-up migrate-down migrate-create migrate-force migrate-version test-backend test-worker test-frontend test-all prod-deploy prod-up prod-down prod-logs prod-status

# Default target
.DEFAULT_GOAL := help

# Load environment variables from .env file
include .env
export

# Construct DATABASE_URL for migrations
DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=disable

## up: Start all services
up:
	@echo "Starting all services..."
	docker compose up -d
	@echo "Services started successfully!"
	@echo "Backend API: http://localhost:8080"
	@echo "Worker: http://localhost:8081"
	@echo "Frontend: http://localhost:3000"
	@echo "PostgreSQL: localhost:5432"

## down: Stop all services
down:
	@echo "Stopping all services..."
	docker compose down
	@echo "Services stopped successfully!"

## logs: View logs from all services
logs:
	docker compose logs -f

## logs-backend: View logs from backend service (last 30s)
logs-backend:
	docker compose logs --since 30s -f backend

## logs-worker: View logs from worker service (last 30s)
logs-worker:
	docker compose logs --since 30s -f worker

## logs-frontend: View logs from frontend service (last 30s)
logs-frontend:
	docker compose logs --since 30s -f frontend

## logs-postgres: View logs from postgres service (last 30s)
logs-postgres:
	docker compose logs --since 30s -f postgres

## rebuild: Rebuild and restart all services
rebuild:
	@echo "Rebuilding all services..."
	docker compose down
	docker compose build --no-cache
	docker compose up -d
	@echo "Services rebuilt and started successfully!"

## clean: Remove all containers, volumes, and images
clean:
	@echo "Cleaning up Docker resources..."
	docker compose down -v --rmi all
	@echo "Cleanup completed!"

## ps: Show status of all services
ps:
	docker compose ps

## restart: Restart all services
restart:
	@echo "Restarting all services..."
	docker compose restart
	@echo "Services restarted successfully!"

## prod-deploy: Deploy to production with backup and health checks
prod-deploy:
	@echo "Running production deployment..."
	@chmod +x deploy.sh
	./deploy.sh

## prod-up: Start production services
prod-up:
	@echo "Starting production services..."
	docker compose -f docker-compose.prod.yml up -d
	@echo "Production services started!"
	@echo "Backend API: http://localhost:8080"
	@echo "Worker: http://localhost:8081"
	@echo "Frontend: http://localhost:3000"

## prod-down: Stop production services
prod-down:
	@echo "Stopping production services..."
	docker compose -f docker-compose.prod.yml down
	@echo "Production services stopped!"

## prod-logs: View production logs
prod-logs:
	docker compose -f docker-compose.prod.yml logs -f

## prod-status: Show production service status
prod-status:
	@echo "Production Service Status:"
	@docker compose -f docker-compose.prod.yml ps
	@echo ""
	@echo "Health Checks:"
	@echo -n "Backend Health: "
	@curl -sf http://localhost:8080/health/ready && echo "✓ Ready" || echo "✗ Not Ready"
	@echo -n "Worker Health: "
	@curl -sf http://localhost:8081/health/ready && echo "✓ Ready" || echo "✗ Not Ready"

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## migrate-up: Run database migrations up
migrate-up:
	@echo "Running database migrations..."
	docker compose exec -e DATABASE_URL="$(DATABASE_URL)" backend migrate -path=/app/migrations -database "$$DATABASE_URL" up
	@echo "Migrations completed!"

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	docker compose exec -e DATABASE_URL="$(DATABASE_URL)" backend migrate -path=/app/migrations -database "$$DATABASE_URL" down
	@echo "Rollback completed!"

## migrate-create: Create a new migration file (usage: make migrate-create name=<migration_name>)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required. Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	docker compose exec backend migrate create -ext sql -dir /app/migrations -seq $(name)
	@echo "Migration files created!"

## migrate-force: Force migration version (usage: make migrate-force version=<version>)
migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Error: version is required. Usage: make migrate-force version=<version>"; \
		exit 1; \
	fi
	@echo "Forcing migration to version $(version)..."
	docker compose exec -e DATABASE_URL="$(DATABASE_URL)" backend migrate -path=/app/migrations -database "$$DATABASE_URL" force $(version)
	@echo "Migration forced to version $(version)!"

## migrate-version: Show current migration version
migrate-version:
	docker compose exec -e DATABASE_URL="$(DATABASE_URL)" backend migrate -path=/app/migrations -database "$$DATABASE_URL" version

## test-backend: Run backend unit tests
test-backend:
	@echo "Running backend tests..."
	cd backend && go test ./... -v -cover
	@echo "Backend tests completed!"

## test-worker: Run worker unit tests
test-worker:
	@echo "Running worker tests..."
	cd worker && go test ./... -v -cover
	@echo "Worker tests completed!"

## test-frontend: Run frontend unit tests
test-frontend:
	@echo "Running frontend tests..."
	cd frontend && npm run test
	@echo "Frontend tests completed!"

## test-all: Run all tests (backend, worker, frontend)
test-all: test-backend test-worker test-frontend
	@echo "All tests completed!"
