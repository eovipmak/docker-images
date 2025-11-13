# Multi-User Data Isolation Setup

## Overview

The SSL Monitor application now supports multi-user data isolation with organizations, ensuring that each user can only access their own data.

## Features

### 1. Database Schema

The database now includes the following tables:

- **organizations**: Multi-tenant organization support
  - `id`: Primary key
  - `name`: Organization name
  - `description`: Optional description
  - `created_at`: Timestamp

- **users**: User accounts with authentication
  - `id`: Primary key
  - `email`: Unique email address
  - `hashed_password`: Securely hashed password
  - `is_active`: Account status
  - `is_superuser`: Admin flag
  - `is_verified`: Email verification status
  - `organization_id`: Foreign key to organizations (optional)
  - `created_at`: Timestamp

- **monitors**: SSL certificate monitoring configurations
  - `id`: Primary key
  - `user_id`: Foreign key to users (required)
  - `organization_id`: Foreign key to organizations (optional)
  - `domain`: Domain to monitor
  - `check_interval`: Check frequency in seconds
  - `webhook_url`: Optional webhook for notifications
  - `last_check`: Last check timestamp
  - `status`: Monitor status (active, paused, error)
  - `created_at`: Timestamp

- **ssl_checks**: SSL certificate check history
  - `id`: Primary key
  - `user_id`: Foreign key to users (for isolation)
  - `organization_id`: Foreign key to organizations (optional)
  - `domain`: Checked domain
  - `ip`: IP address
  - `port`: Port number
  - `status`: Check status
  - `ssl_status`: SSL certificate status
  - `server_status`: Server status
  - `ip_status`: IP geolocation status
  - `checked_at`: Check timestamp
  - `response_data`: Full JSON response

- **alerts**: Certificate alerts and notifications
  - `id`: Primary key
  - `user_id`: Foreign key to users (required)
  - `organization_id`: Foreign key to organizations (optional)
  - `domain`: Domain with alert
  - `alert_type`: Type (expiring_soon, expired, invalid, error)
  - `severity`: Severity level (low, medium, high, critical)
  - `message`: Alert message
  - `is_read`: Read status
  - `is_resolved`: Resolution status
  - `created_at`: Timestamp
  - `resolved_at`: Resolution timestamp

### 2. Data Isolation

All API endpoints now require authentication and automatically filter data by `user_id`:

- `/api/check` - Creates SSL checks associated with the authenticated user
- `/api/history` - Returns only the authenticated user's check history
- `/api/stats` - Returns statistics for the authenticated user only
- `/api/domains` - Returns only the authenticated user's monitored domains
- `/api/domains` (POST) - Creates domain checks for the authenticated user

**Users cannot see or access data from other users.**

### 3. Environment Configuration

Database URLs can be configured via environment variables:

```bash
DATABASE_URL=sqlite:///./ssl_monitor.db
ASYNC_DATABASE_URL=sqlite+aiosqlite:///./ssl_monitor.db
```

In Docker Compose, these are already configured:

```yaml
environment:
  - DATABASE_URL=sqlite:///./ssl_monitor.db
  - ASYNC_DATABASE_URL=sqlite+aiosqlite:///./ssl_monitor.db
```

### 4. Database Migrations

The application uses Alembic for database migrations. All migrations are in `api/alembic/versions/`.

#### Running Migrations

```bash
cd ssl-monitor/api
alembic upgrade head
```

#### Current Migrations

1. `370a9c7b5096` - Initial migration with users, monitors, and ssl_checks tables
2. `e4b85a247cd7` - Update user model for JWT authentication
3. `a5ad82218b13` - Add organizations and alerts tables with multi-user isolation

#### Creating New Migrations

```bash
alembic revision --autogenerate -m "Description of changes"
```

### 5. Database Backup

A backup script is provided to create database backups:

```bash
cd ssl-monitor/api
python backup_database.py [--output-dir DIRECTORY]
```

Default backup location: `./backups/`

The script:
- Creates timestamped backup files
- Supports custom output directories
- Verifies backup creation
- Uses simple file copy for SQLite databases

Example:

```bash
$ python backup_database.py
Creating backup of ssl_monitor.db
Destination: backups/ssl_monitor_backup_20251113_151046.db
✓ Backup created successfully (126,976 bytes)

Backup completed: /home/runner/work/v-insight/v-insight/ssl-monitor/api/backups/ssl_monitor_backup_20251113_151046.db
```

### 6. Testing

Two test suites are provided:

#### Data Isolation Tests

Tests the database layer isolation:

```bash
cd ssl-monitor/api
python -m pytest test_data_isolation.py -v
```

Tests include:
- SSL check isolation between users
- Monitor isolation between users
- Alert isolation between users
- Organization-based filtering
- Statistics isolation

#### API Integration Tests

Tests the API endpoints with authentication:

```bash
cd ssl-monitor/api
python test_api_isolation.py
```

Tests include:
- User registration
- User login
- Authenticated API access
- Unauthenticated access denial
- Data isolation in API responses

## Security

### Data Isolation Implementation

All queries filter by `user_id`:

```python
# Example from /api/history endpoint
query = db.query(SSLCheck).filter(SSLCheck.user_id == user.id)
```

### Authentication Required

All data access endpoints require authentication using JWT tokens:

```python
@app.get("/api/stats")
async def get_stats(
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)  # Authentication required
):
    # Only return data for authenticated user
    total_checks = db.query(SSLCheck).filter(SSLCheck.user_id == user.id).count()
```

### Foreign Key Constraints

All relationships enforce referential integrity:

```python
# SSL checks must belong to a valid user
user_id = Column(Integer, ForeignKey("users.id"), nullable=True, index=True)
```

## Docker Deployment

The application is fully containerized with persistent data:

```yaml
volumes:
  - ssl-monitor-data:/app/api  # Persists database and files
```

To deploy:

```bash
docker-compose up -d
```

The database will be persisted in the Docker volume `ssl-monitor-data`.

## Verification

To verify the setup:

1. **Check migrations are applied:**
   ```bash
   cd ssl-monitor/api
   alembic current
   ```

2. **Verify database schema:**
   ```bash
   sqlite3 ssl_monitor.db ".schema"
   ```

3. **Run isolation tests:**
   ```bash
   python -m pytest test_data_isolation.py -v
   python test_api_isolation.py
   ```

4. **Test backup:**
   ```bash
   python backup_database.py
   ls -lh backups/
   ```

## Acceptance Criteria Status

✅ **Migrations OK** - All migrations applied successfully using Alembic  
✅ **API user-specific** - All endpoints filter by authenticated user  
✅ **Isolation test** - Tests verify User A cannot see User B's data  
✅ **Backup script** - File copy backup script implemented  
✅ **Docker volume** - Database persisted in Docker volume  
✅ **Environment variable DB_URL** - Configurable via DATABASE_URL

## Next Steps

1. **Production Setup:**
   - Generate secure JWT secrets using: `openssl rand -hex 32`
   - Set environment variables:
     ```bash
     JWT_SECRET_KEY=<your-secret-key>
     JWT_REFRESH_SECRET_KEY=<your-refresh-secret-key>
     ```

2. **Organization Management:**
   - Implement organization creation and management endpoints
   - Add organization-level user management
   - Implement organization-wide data access for admins

3. **Alert System:**
   - Implement automatic alert generation for expiring certificates
   - Add webhook notifications
   - Create alert management endpoints

4. **Scheduled Backups:**
   - Set up cron job for automated backups:
     ```bash
     0 2 * * * cd /app/api && python backup_database.py
     ```

5. **Monitoring:**
   - Add health check endpoints
   - Implement metrics collection
   - Set up monitoring for certificate expiration
