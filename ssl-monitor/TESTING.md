# Testing and Verification Guide

## Overview

This guide provides comprehensive instructions for testing the SSL Monitor database setup, migrations, and container communication.

## Prerequisites

- Docker and Docker Compose installed
- Basic knowledge of command line
- (Optional) Python 3.12+ for local testing

## Quick Test

### 1. Start the Services

```bash
cd /home/runner/work/docker-images/docker-images
docker compose up -d
```

Wait for services to be healthy (about 10 seconds).

### 2. Verify Services are Running

```bash
docker compose ps
```

Expected output:
```
NAME          IMAGE                       COMMAND                  SERVICE       CREATED          STATUS
ssl-checker   docker-images-ssl-checker   "uvicorn main:app --…"   ssl-checker   X seconds ago    Up X seconds (healthy)
ssl-monitor   docker-images-ssl-monitor   "/app/entrypoint.sh"     ssl-monitor   X seconds ago    Up X seconds
```

### 3. Test Container Communication

```bash
curl "http://localhost:8001/api/check?domain=google.com&port=443"
```

Expected: JSON response with SSL certificate details and `"status":"success"`.

### 4. Verify Database Tables

```bash
docker compose exec ssl-monitor python -c "
from database import engine
from sqlalchemy import inspect
inspector = inspect(engine)
print('Tables:', ', '.join(inspector.get_table_names()))
"
```

Expected output:
```
Tables: alembic_version, monitors, ssl_checks, users
```

### 5. Verify Sample Data

```bash
docker compose exec ssl-monitor python -c "
from database import SessionLocal, User, Monitor, SSLCheck
db = SessionLocal()
print(f'Users: {db.query(User).count()}')
print(f'Monitors: {db.query(Monitor).count()}')
print(f'SSL Checks: {db.query(SSLCheck).count()}')
db.close()
"
```

Expected output:
```
Users: 2
Monitors: 3
SSL Checks: 1 (or more, depending on how many checks you've made)
```

### 6. Test Data Persistence

```bash
# Restart the container
docker compose restart ssl-monitor
sleep 5

# Check data is still there
docker compose exec ssl-monitor python -c "
from database import SessionLocal, SSLCheck
db = SessionLocal()
print(f'SSL Checks after restart: {db.query(SSLCheck).count()}')
db.close()
"
```

The count should be the same as before restart.

### 7. Clean Up

```bash
# Stop services but keep data
docker compose down

# Or stop and remove all data
docker compose down -v
```

## Detailed Testing

### Local Development Testing (Without Docker)

1. **Install dependencies:**
   ```bash
   cd ssl-monitor/api
   pip install -r requirements.txt
   ```

2. **Run migrations:**
   ```bash
   alembic upgrade head
   ```

3. **Create sample data:**
   ```bash
   python create_sample_data.py
   ```

4. **Start the server:**
   ```bash
   uvicorn main:app --reload --port 8001
   ```

5. **Test endpoints:**
   ```bash
   # Stats endpoint
   curl "http://localhost:8001/api/stats"
   
   # History endpoint
   curl "http://localhost:8001/api/history?limit=10"
   
   # Check endpoint (requires ssl-checker running on port 8000)
   curl "http://localhost:8001/api/check?domain=example.com"
   ```

### Migration Testing

1. **Check current migration status:**
   ```bash
   cd ssl-monitor/api
   alembic current
   ```

2. **View migration history:**
   ```bash
   alembic history
   ```

3. **Test migration idempotency:**
   ```bash
   # Run upgrade multiple times - should be safe
   alembic upgrade head
   alembic upgrade head
   alembic upgrade head
   ```

4. **Create new migration (after model changes):**
   ```bash
   alembic revision --autogenerate -m "Description of changes"
   ```

### Container Communication Testing

1. **Check network configuration:**
   ```bash
   docker network ls
   docker network inspect docker-images_ssl-network
   ```

2. **Test DNS resolution:**
   ```bash
   docker compose exec ssl-monitor ping -c 2 ssl-checker
   docker compose exec ssl-monitor nslookup ssl-checker
   ```

3. **Test direct API call between containers:**
   ```bash
   docker compose exec ssl-monitor curl "http://ssl-checker:8000/api/check?domain=google.com"
   ```

### Database Schema Verification

1. **Inspect database schema:**
   ```bash
   docker compose exec ssl-monitor python -c "
   from database import engine
   from sqlalchemy import inspect
   
   inspector = inspect(engine)
   for table_name in inspector.get_table_names():
       print(f'\nTable: {table_name}')
       for column in inspector.get_columns(table_name):
           print(f\"  - {column['name']}: {column['type']}\")
   "
   ```

2. **Test foreign key relationships:**
   ```bash
   docker compose exec ssl-monitor python -c "
   from database import SessionLocal, Monitor, User
   db = SessionLocal()
   
   # Get a monitor with its user
   monitor = db.query(Monitor).first()
   if monitor:
       print(f'Monitor: {monitor.domain}')
       print(f'Owned by user: {monitor.user.username}')
   db.close()
   "
   ```

## Troubleshooting

### Container Communication Issues

**Problem:** Error about DNS resolution or connection refused.

**Solutions:**
1. Check that both containers are on the same network:
   ```bash
   docker compose ps
   docker network inspect docker-images_ssl-network
   ```

2. Verify ssl-checker is healthy before ssl-monitor starts:
   ```bash
   docker compose logs ssl-checker | grep -i health
   ```

3. Check environment variable:
   ```bash
   docker compose exec ssl-monitor env | grep SSL_CHECKER_URL
   ```

### Migration Issues

**Problem:** Migrations fail or tables don't exist.

**Solutions:**
1. Check migration logs:
   ```bash
   docker compose logs ssl-monitor | grep -i migration
   ```

2. Manually run migrations inside container:
   ```bash
   docker compose exec ssl-monitor alembic upgrade head
   ```

3. Reset database (WARNING: loses all data):
   ```bash
   docker compose down -v
   docker compose up -d
   ```

### Database Locked Issues

**Problem:** "Database is locked" errors.

**Solutions:**
1. SQLite doesn't handle concurrent writes well. This is normal for SQLite.
2. For production, consider PostgreSQL or MySQL.
3. Ensure only one process writes at a time.

## Verification Checklist

- [ ] Docker containers start successfully
- [ ] ssl-checker becomes healthy
- [ ] ssl-monitor starts after ssl-checker is healthy
- [ ] Database migrations run on startup
- [ ] All three tables exist (users, monitors, ssl_checks)
- [ ] Sample data is created (2 users, 3 monitors)
- [ ] Container-to-container communication works
- [ ] API endpoints respond correctly
- [ ] SSL checks are saved to database
- [ ] Data persists across container restarts
- [ ] Foreign key relationships work correctly

## Expected Behavior

### On First Start
1. Containers build
2. ssl-checker starts and becomes healthy
3. ssl-monitor waits for ssl-checker
4. ssl-monitor runs migrations
5. Sample data is created (if not already present)
6. Both services are ready

### On Subsequent Starts
1. Containers start
2. ssl-monitor runs migrations (no changes if schema is current)
3. Existing data is preserved
4. Services are ready

### Data Persistence
- Database file is stored in Docker volume `ssl-monitor-data`
- Data survives container restarts
- Data survives container recreation
- Data is lost only if volume is removed (`docker compose down -v`)

## Performance Notes

- **Startup time:** ~5-10 seconds for both containers
- **Health check interval:** Every 10 seconds
- **Migration time:** < 1 second (for current schema)
- **Database size:** Minimal (~100KB initially, grows with usage)
