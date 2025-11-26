# Production Deployment Guide

This guide provides step-by-step instructions for deploying V-Insight to production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Production Deployment](#production-deployment)
4. [Post-Deployment](#post-deployment)
5. [Monitoring](#monitoring)
6. [Backup and Recovery](#backup-and-recovery)
7. [Troubleshooting](#troubleshooting)
8. [Scaling](#scaling)

## Prerequisites

### Server Requirements

- **OS**: Ubuntu 20.04/22.04 LTS or similar Linux distribution
- **CPU**: Minimum 2 cores (4+ recommended for production)
- **RAM**: Minimum 4GB (8GB+ recommended)
- **Storage**: Minimum 20GB free space
- **Network**: Public IP address or domain name

### Software Requirements

- Docker Engine 20.10+
- Docker Compose 2.0+
- Git
- curl or wget

### Installation

```bash
# Install Docker (Ubuntu/Debian)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

## Initial Setup

### 1. Clone Repository

```bash
cd /opt
sudo git clone https://github.com/eovipmak/v-insight.git
sudo chown -R $USER:$USER v-insight
cd v-insight
```

### 2. Configure Environment

```bash
# Copy production environment template
cp .env.production .env

# Edit configuration
nano .env
```

**Required changes in `.env`:**

```bash
# Generate strong password
POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Set your domain or server IP
PUBLIC_API_URL=https://yourdomain.com
# or for non-HTTPS:
PUBLIC_API_URL=http://YOUR_SERVER_IP:8080

# Add your domain
VITE_ALLOWED_HOSTS=localhost,127.0.0.1,yourdomain.com
```

### 3. Configure Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable

# Or firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

## Production Deployment

### Method 1: Automated Deployment (Recommended)

```bash
# Make deployment script executable
chmod +x deploy.sh

# Run deployment
./deploy.sh
```

The script will:
1. ✅ Validate requirements
2. ✅ Create database backup
3. ✅ Build production images
4. ✅ Run database migrations
5. ✅ Start services
6. ✅ Verify health checks
7. ✅ Display status

### Method 2: Manual Deployment

```bash
# Build production images
docker compose -f docker-compose.prod.yml build --no-cache

# Start services
docker compose -f docker-compose.prod.yml up -d

# Check status
docker compose -f docker-compose.prod.yml ps

# Verify health
curl http://localhost:8080/health/ready
curl http://localhost:8081/health/ready
```

### Using Make Commands

```bash
# Deploy with automated script
make prod-deploy

# Or manually
make prod-up          # Start services
make prod-status      # Check health
make prod-logs        # View logs
```

## Post-Deployment

### 1. Verify Services

```bash
# Check service status
make prod-status

# Expected output:
# Backend Health: ✓ Ready
# Worker Health: ✓ Ready
```

### 2. Access Application

- **Frontend**: http://YOUR_SERVER_IP:3000
- **Backend API**: http://YOUR_SERVER_IP:8080 (internal use)
- **Worker**: http://YOUR_SERVER_IP:8081 (internal use)

### 3. Create First Account

1. Navigate to http://YOUR_SERVER_IP:3000
2. Click "Register"
3. Create admin account
4. Start monitoring!

### 4. Setup HTTPS (Recommended)

#### Option 1: Nginx Reverse Proxy with Let's Encrypt

```bash
# Install Nginx and Certbot
sudo apt update
sudo apt install nginx certbot python3-certbot-nginx

# Configure Nginx
sudo nano /etc/nginx/sites-available/v-insight
```

**Nginx configuration:**

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Backend API (optional, usually proxied through frontend)
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/v-insight /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Obtain SSL certificate
sudo certbot --nginx -d yourdomain.com

# Auto-renewal is configured automatically
sudo certbot renew --dry-run
```

#### Option 2: Traefik (Docker-based)

See `docker-compose.prod.yml` and add Traefik service configuration.

### 5. Configure Monitoring

```bash
# Setup cron job for health checks (optional)
crontab -e

# Add:
*/5 * * * * curl -sf http://localhost:8080/health/ready > /dev/null || echo "Backend unhealthy" | mail -s "V-Insight Alert" admin@yourdomain.com
```

## Monitoring

### Health Check Endpoints

```bash
# Backend
curl http://localhost:8080/health        # Legacy
curl http://localhost:8080/health/live   # Liveness
curl http://localhost:8080/health/ready  # Readiness

# Worker
curl http://localhost:8081/health        # Legacy
curl http://localhost:8081/health/live   # Liveness
curl http://localhost:8081/health/ready  # Readiness
```

### View Logs

```bash
# All services
make prod-logs

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f worker
docker compose -f docker-compose.prod.yml logs -f frontend
docker compose -f docker-compose.prod.yml logs -f postgres
```

### Resource Usage

```bash
# Docker stats
docker stats

# Service status
make prod-status
```

## Backup and Recovery

### Automated Backups

Backups are automatically created in `./backups/` during deployment:

- **Location**: `./backups/db_backup_YYYYMMDD_HHMMSS.sql`
- **Retention**: Last 10 backups
- **Trigger**: Before each deployment

### Manual Backup

```bash
# Create backup
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U $POSTGRES_USER $POSTGRES_DB > backups/manual_$(date +%Y%m%d_%H%M%S).sql

# Compress backup
gzip backups/manual_$(date +%Y%m%d_%H%M%S).sql
```

### Restore Backup

```bash
# Stop services
docker compose -f docker-compose.prod.yml down

# Start database only
docker compose -f docker-compose.prod.yml up -d postgres
sleep 10

# Restore backup
cat backups/db_backup_20240115_120000.sql | \
  docker compose -f docker-compose.prod.yml exec -T postgres \
  psql -U $POSTGRES_USER $POSTGRES_DB

# Start all services
docker compose -f docker-compose.prod.yml up -d

# Verify
make prod-status
```

### Offsite Backup Strategy

```bash
# Setup daily backup to S3/remote server
# Add to crontab:
0 2 * * * /opt/v-insight/backup-to-s3.sh
```

**Example backup script (`backup-to-s3.sh`):**

```bash
#!/bin/bash
BACKUP_FILE="backups/db_backup_$(date +%Y%m%d_%H%M%S).sql"
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U $POSTGRES_USER $POSTGRES_DB > $BACKUP_FILE
gzip $BACKUP_FILE
# Upload to S3 (requires AWS CLI)
aws s3 cp ${BACKUP_FILE}.gz s3://your-bucket/v-insight-backups/
```

## Troubleshooting

### Services Not Starting

```bash
# Check logs
make prod-logs

# Check Docker
docker compose -f docker-compose.prod.yml ps

# Restart services
docker compose -f docker-compose.prod.yml restart
```

### Database Connection Failed

```bash
# Check PostgreSQL
docker compose -f docker-compose.prod.yml logs postgres

# Verify credentials in .env
cat .env | grep POSTGRES

# Restart PostgreSQL
docker compose -f docker-compose.prod.yml restart postgres
```

### Health Checks Failing

```bash
# Check backend
curl http://localhost:8080/health/ready

# Check worker
curl http://localhost:8081/health/ready

# Review logs
make prod-logs
```

### Out of Disk Space

```bash
# Check disk usage
df -h

# Clean Docker
docker system prune -a

# Clean old logs
find /var/lib/docker/containers -name "*.log" -exec truncate -s 0 {} \;

# Clean old backups manually
find backups/ -name "*.sql" -mtime +30 -delete
```

### Performance Issues

```bash
# Check resource usage
docker stats

# Adjust resources in docker-compose.prod.yml:
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

## Scaling

### Vertical Scaling

Increase resources in `docker-compose.prod.yml`:

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### Horizontal Scaling

For load balancing, use Docker Swarm or Kubernetes:

```bash
# Docker Swarm example
docker swarm init
docker stack deploy -c docker-compose.prod.yml v-insight
docker service scale v-insight_backend=3
```

### Database Optimization

```yaml
# In docker-compose.prod.yml
postgres:
  environment:
    POSTGRES_SHARED_BUFFERS: 256MB
    POSTGRES_MAX_CONNECTIONS: 200
```

## Maintenance

### Update Application

```bash
# Pull latest code
git pull origin main

# Deploy update
./deploy.sh

# Or manually
make prod-deploy
```

### Database Migrations

Migrations run automatically during deployment. Manual migration:

```bash
# Check current version
docker compose -f docker-compose.prod.yml exec backend \
  migrate -path=/app/migrations -database "$DATABASE_URL" version

# Run migrations
docker compose -f docker-compose.prod.yml exec backend \
  migrate -path=/app/migrations -database "$DATABASE_URL" up
```

### Log Rotation

Logs are automatically rotated (10MB max, 3 files). Manual cleanup:

```bash
# Clean Docker logs
docker compose -f docker-compose.prod.yml down
find /var/lib/docker/containers -name "*.log" -exec truncate -s 0 {} \;
docker compose -f docker-compose.prod.yml up -d
```

## Security Checklist

- [ ] Strong PostgreSQL password set
- [ ] Unique JWT secret generated
- [ ] Firewall configured (ports 22, 80, 443 only)
- [ ] HTTPS enabled with valid SSL certificate
- [ ] Regular backups configured and tested
- [ ] Log monitoring set up
- [ ] OS security updates enabled
- [ ] Docker daemon secured
- [ ] Environment variables not committed to git
- [ ] Database exposed only to Docker network

## Support

For issues or questions:
- GitHub Issues: https://github.com/eovipmak/v-insight/issues
- Documentation: https://github.com/eovipmak/v-insight
