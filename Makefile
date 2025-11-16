.PHONY: up down logs rebuild clean help

# Default target
.DEFAULT_GOAL := help

## up: Start all services
up:
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Services started successfully!"
	@echo "Backend API: http://localhost:8080"
	@echo "Worker: http://localhost:8081"
	@echo "Frontend: http://localhost:3000"
	@echo "PostgreSQL: localhost:5432"

## down: Stop all services
down:
	@echo "Stopping all services..."
	docker-compose down
	@echo "Services stopped successfully!"

## logs: View logs from all services
logs:
	docker-compose logs -f

## logs-backend: View logs from backend service
logs-backend:
	docker-compose logs -f backend

## logs-worker: View logs from worker service
logs-worker:
	docker-compose logs -f worker

## logs-frontend: View logs from frontend service
logs-frontend:
	docker-compose logs -f frontend

## logs-postgres: View logs from postgres service
logs-postgres:
	docker-compose logs -f postgres

## rebuild: Rebuild and restart all services
rebuild:
	@echo "Rebuilding all services..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "Services rebuilt and started successfully!"

## clean: Remove all containers, volumes, and images
clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v --rmi all
	@echo "Cleanup completed!"

## ps: Show status of all services
ps:
	docker-compose ps

## restart: Restart all services
restart:
	@echo "Restarting all services..."
	docker-compose restart
	@echo "Services restarted successfully!"

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
