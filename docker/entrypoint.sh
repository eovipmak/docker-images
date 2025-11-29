#!/bin/sh
set -e
set -u
set -o pipefail

# Construct DATABASE_URL from environment variables
DATABASE_URL="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable"

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER"; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is up - running migrations"

# Change to the app directory
cd /app

# Run migrations with error handling
if ! migrate -path /app/migrations -database "$DATABASE_URL" up; then
  echo "Migration failed"
  exit 1
fi

echo "Migrations completed - starting application"

# Execute the main command
exec "$@"
