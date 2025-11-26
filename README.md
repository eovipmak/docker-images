# V-Insight - Multi-tenant Monitoring SaaS

A Docker-based multi-tenant monitoring SaaS platform built with Go and SvelteKit. Monitor your websites, APIs, and services with automated health checks, SSL monitoring, and intelligent alerting.

## Architecture

- **Backend API** (Go, Gin): REST API service on port 8080
- **Worker** (Go): Background job processing service on port 8081
- **Frontend** (SvelteKit): Web interface on port 3000 with built-in API proxy
- **PostgreSQL 15**: Database on port 5432

# v-insight

v-insight is a Docker-based, multi-tenant monitoring SaaS for automated health checks, SSL monitoring, and alerting. It helps teams monitor websites, APIs, and services with scalable background workers and a developer-friendly API + web UI.

## ✨ Key features

- HTTP/HTTPS health checks (uptime & response time)
- SSL certificate expiry monitoring
- Flexible alert rules (down, slow_response, ssl_expiry) and incident management
- Multiple notification channels (webhook, Discord, email-ready)
- Multi-tenant data isolation with background worker jobs

## 🚀 Quick Start

```bash
git clone https://github.com/eovipmak/v-insight.git
cd v-insight
cp .env.example .env
make up
```

Or without Make:

```bash
# v-insight

v-insight is a Docker-based, multi-tenant monitoring SaaS for automated health checks, SSL monitoring, and alerting. It helps teams monitor websites, APIs, and services with scalable background workers and a developer-friendly API + web UI.

## ✨ Key features

- HTTP/HTTPS health checks (uptime & response time)
- SSL certificate expiry monitoring
- Flexible alert rules (down, slow_response, ssl_expiry) and incident management
- Multiple notification channels (webhook, Discord, email-ready)
- Multi-tenant data isolation with background worker jobs

## 🚀 Quick Start

```bash
git clone https://github.com/eovipmak/v-insight.git
cd v-insight
cp .env.example .env
make up
```

Or without Make:

```bash
docker-compose up -d
```

Database migrations run automatically on startup.

## 📚 Full documentation

Detailed guides (installation, usage, configuration, architecture, contributing, troubleshooting) are in the `docs/` folder:

- Installation: `./docs/installation.md`
- Usage & examples: `./docs/usage.md`
- Configuration: `./docs/configuration.md`
- Architecture: `./docs/architecture.md`
- Contributing: `./docs/contributing.md`
- Troubleshooting: `./docs/troubleshooting.md`

## 📄 License

This project is licensed under the MIT License — see the `LICENSE` file for details.
