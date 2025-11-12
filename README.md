# Docker Images

This repository contains various Docker-based applications.

## Projects

### SSL Checker
A FastAPI-based application that checks SSL certificate information, web server technology, and IP geolocation details for domains and IP addresses.

**Location:** `ssl-checker/`  
**Port:** 8000  
**Documentation:** [ssl-checker/README.md](ssl-checker/README.md)

### SSL Monitor
A web-based SSL certificate monitoring application with history tracking. Built with Bootstrap 5, HTMX, and FastAPI with SQLite for persistent storage.

**Location:** `ssl-monitor/`  
**Port:** 8001  
**Documentation:** [ssl-monitor/README.md](ssl-monitor/README.md)

### Image Search
Image search application.

**Location:** `image-search/`

## Quick Start with Docker Compose

Run both SSL Checker and SSL Monitor together:

```bash
docker-compose up -d
```

This will start:
- SSL Checker on http://localhost:8000
- SSL Monitor on http://localhost:8001

## Individual Project Setup

Each project can be run independently. See the respective README files for details:
- [SSL Checker README](ssl-checker/README.md)
- [SSL Monitor README](ssl-monitor/README.md)

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│  SSL Monitor    │ ──────> │  SSL Checker    │
│  (Port 8001)    │ API     │  (Port 8000)    │
│                 │ Calls   │                 │
│  - Frontend UI  │         │  - SSL Validation│
│  - SQLite DB    │         │  - Cert Info    │
│  - History      │         │  - Geo Lookup   │
└─────────────────┘         └─────────────────┘
```

SSL Monitor uses SSL Checker as a backend service for performing SSL certificate checks, while adding its own database layer for history tracking and monitoring statistics.
