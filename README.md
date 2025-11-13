# Docker Images

This repository contains various Docker-based applications.

## Projects

### SSL Checker
A FastAPI-based application that checks SSL certificate information, web server technology, and IP geolocation details for domains and IP addresses.

**Location:** `ssl-checker/`  
**Port:** 8000  
**Documentation:** [ssl-checker/README.md](ssl-checker/README.md)

### SSL Monitor
A web-based SSL certificate monitoring application with history tracking. Features a modern React frontend with TypeScript and Material-UI, combined with a FastAPI backend and SQLite for persistent storage.

**Location:** `ssl-monitor/`  
**Port:** 8001  
**Documentation:** [ssl-monitor/README.md](ssl-monitor/README.md)

**Technology Stack:**
- **Frontend:** React 19, TypeScript, Material-UI, Vite
- **Backend:** FastAPI, SQLAlchemy, SQLite
- **Features:** Real-time monitoring, history tracking, responsive design

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
┌─────────────────────────────┐         ┌─────────────────┐
│     SSL Monitor             │ ──────> │  SSL Checker    │
│     (Port 8001)             │ API     │  (Port 8000)    │
│                             │ Calls   │                 │
│  ┌───────────────────────┐  │         │  - SSL Validation│
│  │ React Frontend        │  │         │  - Cert Info    │
│  │ (TypeScript, MUI)     │  │         │  - Geo Lookup   │
│  └───────────────────────┘  │         └─────────────────┘
│  ┌───────────────────────┐  │
│  │ FastAPI Backend       │  │
│  │ (Python, SQLAlchemy)  │  │
│  └───────────────────────┘  │
│  ┌───────────────────────┐  │
│  │ SQLite Database       │  │
│  │ (History & Monitors)  │  │
│  └───────────────────────┘  │
└─────────────────────────────┘
```

SSL Monitor is a full-stack application with an integrated React frontend and FastAPI backend. It uses SSL Checker as a backend service for performing SSL certificate checks, while adding its own database layer for history tracking and monitoring statistics.
