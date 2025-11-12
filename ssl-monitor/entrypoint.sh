#!/bin/bash
set -e

echo "Running database migrations..."
cd /app/api
alembic upgrade head

echo "Starting SSL Monitor service..."
exec uvicorn main:app --host 0.0.0.0 --port 8001
