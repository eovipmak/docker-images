# Build stage - Build React frontend
FROM node:20-alpine AS frontend-build

WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build the application
RUN npm run build

# Production stage - Python backend with built frontend
FROM python:3.12-slim

LABEL org.opencontainers.image.description="SSL Monitor - SSL Certificate Monitoring with History"

WORKDIR /app

# Copy requirements and install dependencies
COPY api/requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir --timeout=100 -r requirements.txt

# Copy application files
COPY api ./api
COPY entrypoint.sh /app/entrypoint.sh

# Copy built frontend from build stage
COPY --from=frontend-build /app/frontend/dist ./frontend-dist

# Make entrypoint script executable
RUN chmod +x /app/entrypoint.sh

WORKDIR /app/api

EXPOSE 8001

ENTRYPOINT ["/app/entrypoint.sh"]