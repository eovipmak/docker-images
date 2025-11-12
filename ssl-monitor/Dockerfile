FROM python:3.12-slim

LABEL org.opencontainers.image.description="SSL Monitor - SSL Certificate Monitoring with History"

WORKDIR /app

# Copy requirements and install dependencies
COPY api/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY api ./api
COPY ui ./ui
COPY static ./static
COPY entrypoint.sh /app/entrypoint.sh

# Make entrypoint script executable
RUN chmod +x /app/entrypoint.sh

WORKDIR /app/api

EXPOSE 8001

ENTRYPOINT ["/app/entrypoint.sh"]