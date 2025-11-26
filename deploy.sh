#!/bin/bash
# Production Deployment Script for V-Insight
# This script handles safe deployment with database backup and rollback capability

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/db_backup_${TIMESTAMP}.sql"
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if Docker Compose is installed
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if .env file exists
    if [ ! -f "$ENV_FILE" ]; then
        log_error ".env file not found. Please create it from .env.production template."
        exit 1
    fi
    
    log_info "All requirements met."
}

create_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        log_info "Creating backup directory..."
        mkdir -p "$BACKUP_DIR"
    fi
}

backup_database() {
    log_info "Creating database backup..."
    
    # Source environment variables
    set -a
    source "$ENV_FILE"
    set +a
    
    # Check if database container is running
    if ! docker compose -f "$COMPOSE_FILE" ps postgres | grep -q "Up"; then
        log_warn "Database container is not running. Skipping backup."
        return 0
    fi
    
    # Create backup
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > "$BACKUP_FILE"
    
    if [ -f "$BACKUP_FILE" ] && [ -s "$BACKUP_FILE" ]; then
        log_info "Database backup created: $BACKUP_FILE"
        
        # Keep only last 10 backups
        log_info "Cleaning up old backups (keeping last 10)..."
        ls -t ${BACKUP_DIR}/db_backup_*.sql | tail -n +11 | xargs -r rm
    else
        log_error "Database backup failed!"
        exit 1
    fi
}

pull_images() {
    log_info "Pulling latest images..."
    docker compose -f "$COMPOSE_FILE" pull
}

build_images() {
    log_info "Building Docker images..."
    docker compose -f "$COMPOSE_FILE" build --no-cache
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Source environment variables
    set -a
    source "$ENV_FILE"
    set +a
    
    # Construct DATABASE_URL
    DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"
    
    # Run migrations
    docker compose -f "$COMPOSE_FILE" exec -T -e DATABASE_URL="$DATABASE_URL" backend migrate -path=/app/migrations -database "$DATABASE_URL" up
    
    log_info "Migrations completed successfully."
}

restart_services() {
    log_info "Restarting services..."
    
    # Stop services gracefully
    docker compose -f "$COMPOSE_FILE" down
    
    # Start services
    docker compose -f "$COMPOSE_FILE" up -d
    
    log_info "Services started. Waiting for health checks..."
    sleep 10
}

check_health() {
    log_info "Checking service health..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Health check attempt $attempt/$max_attempts..."
        
        # Check backend health
        if curl -sf http://localhost:8080/health/ready > /dev/null 2>&1; then
            log_info "Backend is healthy!"
            
            # Check worker health
            if curl -sf http://localhost:8081/health/ready > /dev/null 2>&1; then
                log_info "Worker is healthy!"
                log_info "All services are healthy and ready!"
                return 0
            fi
        fi
        
        sleep 5
        ((attempt++))
    done
    
    log_error "Health checks failed after $max_attempts attempts!"
    log_error "Deployment may have failed. Check logs with: docker compose -f $COMPOSE_FILE logs"
    return 1
}

rollback() {
    log_error "Deployment failed! Rolling back..."
    
    # Stop current containers
    docker compose -f "$COMPOSE_FILE" down
    
    # If backup exists, offer to restore
    if [ -f "$BACKUP_FILE" ]; then
        log_warn "Database backup available: $BACKUP_FILE"
        log_warn "To restore, run: cat $BACKUP_FILE | docker compose -f $COMPOSE_FILE exec -T postgres psql -U \$POSTGRES_USER \$POSTGRES_DB"
    fi
    
    exit 1
}

show_status() {
    log_info "Deployment Status:"
    echo ""
    docker compose -f "$COMPOSE_FILE" ps
    echo ""
    log_info "Service URLs:"
    echo "  - Backend API: http://localhost:8080"
    echo "  - Backend Health: http://localhost:8080/health"
    echo "  - Backend Ready: http://localhost:8080/health/ready"
    echo "  - Worker: http://localhost:8081"
    echo "  - Worker Health: http://localhost:8081/health"
    echo "  - Worker Ready: http://localhost:8081/health/ready"
    echo "  - Frontend: http://localhost:3000"
    echo "  - Database: localhost:5432"
    echo ""
    log_info "Logs: docker compose -f $COMPOSE_FILE logs -f"
}

# Main deployment flow
main() {
    log_info "Starting V-Insight production deployment..."
    echo ""
    
    # Check requirements
    check_requirements
    
    # Create backup directory
    create_backup_dir
    
    # Backup database
    backup_database
    
    # Pull latest images
    pull_images
    
    # Restart services
    restart_services
    
    # Check health
    if check_health; then
        show_status
        log_info "Deployment completed successfully!"
        exit 0
    else
        rollback
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "V-Insight Production Deployment Script"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --no-backup    Skip database backup (not recommended)"
        echo ""
        echo "Examples:"
        echo "  $0                 # Full deployment with backup and image pull"
        echo "  $0 --no-backup     # Deploy without backup"
        exit 0
        ;;
    --no-backup)
        log_warn "Skipping database backup as requested"
        backup_database() { log_warn "Backup skipped"; }
        main
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
