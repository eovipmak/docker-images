FROM python:3.12-slim

LABEL org.opencontainers.image.description="SSL Monitor - Certificate monitoring service with database storage"

WORKDIR /app

# Copy requirements and install dependencies
COPY api/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY api ./api
COPY ui ./ui

WORKDIR /app/api

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]