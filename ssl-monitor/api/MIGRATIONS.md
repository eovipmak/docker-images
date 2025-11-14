# Database Migrations Guide

This document provides guidance on managing database migrations for the SSL Monitor application using Alembic.

## Current Migration Status

The application uses Alembic for database schema management. All migrations are located in `alembic/versions/`.

### Migration History

The current migration history includes:
- `370a9c7b5096` - Initial migration with users, monitors, and ssl_checks tables
- `e4b85a247cd7` - Update user model for JWT authentication
- **Branch Point** - The migration tree splits here:
  - Branch A: `a5ad82218b13` → `5f26bc034a6d` (Organizations and alerts)
  - Branch B: `f9c12345abcd` (Monitor fields: port, alerts_enabled)
- `88021bad086a` - **Merge migration** combining both branches

## Running Migrations

### In Development

```bash
cd ssl-monitor/api

# Check current migration status
alembic current

# View migration history
alembic history

# List all head revisions
alembic heads

# Upgrade to latest (handles multiple heads)
alembic upgrade heads
```

### In Docker

Migrations run automatically when the container starts via the `entrypoint.sh` script:
```bash
alembic upgrade heads
```

**Note:** We use `alembic upgrade heads` (plural) instead of `alembic upgrade head` (singular) to gracefully handle scenarios where multiple migration branches exist.

## Handling Multiple Heads

### What are multiple heads?

Multiple heads occur when two or more migration branches exist simultaneously, typically when:
1. Developers create migrations in parallel on different branches
2. Both branches are merged into main without a merge migration

### How to detect multiple heads

```bash
alembic heads
```

If you see output like this, you have multiple heads:
```
5f26bc034a6d (head)
f9c12345abcd (head)
```

### How to resolve multiple heads

#### Option 1: Use `alembic upgrade heads` (Recommended)

This command will upgrade to all head revisions automatically:
```bash
alembic upgrade heads
```

This is the approach used in our `entrypoint.sh` and is the most robust solution.

#### Option 2: Create a merge migration

If you want to merge the branches explicitly:

```bash
# Create a merge migration
alembic merge -m "Merge multiple heads" <revision1> <revision2>

# Example:
alembic merge -m "Merge multiple heads" 5f26bc034a6d f9c12345abcd
```

This creates a new migration file that merges the branches.

## Best Practices

### For Developers

1. **Check for existing migrations** before creating a new one:
   ```bash
   alembic heads
   ```

2. **Coordinate with team** when creating migrations to avoid parallel branches

3. **Always test migrations** locally before committing:
   ```bash
   # Test upgrade
   alembic upgrade heads
   
   # Test downgrade (if applicable)
   alembic downgrade -1
   ```

4. **Use descriptive migration messages**:
   ```bash
   alembic revision --autogenerate -m "Add index to domain column for faster lookups"
   ```

### For Production Deployments

1. **Always backup the database** before running migrations

2. **Review migrations** in the `alembic/versions/` directory before deploying

3. **Use the entrypoint script** which automatically handles migrations with `alembic upgrade heads`

4. **Monitor logs** during container startup to ensure migrations complete successfully

## Troubleshooting

### Error: "Multiple head revisions are present"

**Cause:** Using `alembic upgrade head` (singular) when multiple heads exist.

**Solution:** Use `alembic upgrade heads` (plural) or create a merge migration.

### Error: "Can't locate revision identified by"

**Cause:** The database has a revision that doesn't exist in the current migration files.

**Solution:** 
1. Ensure all migration files are present in `alembic/versions/`
2. Check if you're using the correct database
3. Verify the git branch has all necessary migration files

### Fresh database setup

To start with a clean database:

```bash
# Remove old database
rm ssl_monitor.db

# Run migrations
alembic upgrade heads
```

### Checking database state

To see which migrations have been applied:

```bash
alembic current
```

To see the full migration history:

```bash
alembic history --verbose
```

## Docker Volume Considerations

The docker-compose configuration uses a named volume for persistence:
```yaml
volumes:
  - ssl-monitor-data:/app/api
```

This means the database persists between container restarts. To start fresh:

```bash
# Stop and remove containers, networks, and volumes
docker compose down -v

# Start fresh
docker compose up -d
```

**Warning:** The `-v` flag removes volumes, which will delete your database!

## References

- [Alembic Documentation](https://alembic.sqlalchemy.org/en/latest/)
- [Alembic Branches and Merging](https://alembic.sqlalchemy.org/en/latest/branches.html)
- [SQLAlchemy Documentation](https://docs.sqlalchemy.org/)
