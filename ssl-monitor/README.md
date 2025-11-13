# SSL Monitor

A full-stack web application for SSL certificate monitoring with history tracking. This application combines a modern React frontend with a FastAPI backend and SQLite database for persistent storage.

## Features

- **Modern React UI**: Built with React 19, TypeScript, and Material-UI
- **SSL Certificate Checking**: Check SSL certificates for domains and IP addresses
- **History Tracking**: Store and view history of all SSL checks in SQLite database
- **Statistics Dashboard**: View overall monitoring statistics
- **Real-time Updates**: Dynamic UI updates
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **API Integration**: Makes API calls to the ssl-checker service for SSL validation

## Technology Stack

**Frontend:**
- React 19 with TypeScript
- Material-UI (MUI) for UI components
- Vite for fast development and optimized builds
- React Router for client-side routing
- Axios for API calls

**Backend:**
- FastAPI (Python web framework)
- SQLAlchemy ORM
- SQLite database
- Uvicorn ASGI server

## Project Structure

```plaintext
ssl-monitor/
├── frontend/              # React frontend application
│   ├── src/              # React source code
│   │   ├── components/   # Reusable components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API services
│   │   └── ...
│   ├── package.json      # Frontend dependencies
│   └── vite.config.ts    # Vite configuration
├── api/                  # Backend API
│   ├── main.py          # FastAPI application and endpoints
│   ├── database.py      # Database models and configuration
│   ├── alembic/         # Database migrations
│   └── requirements.txt # Python dependencies
├── Dockerfile           # Multi-stage Docker build
└── README.md            # This file
```

## Quick Start

### Using Docker

The SSL Monitor runs on port **8001** (different from ssl-checker's 8000).

```bash
# Build the image
docker build -t ssl-monitor .

# Run the container
docker run -p 8001:8001 ssl-monitor
```

**Note:** The SSL Monitor requires the ssl-checker service to be running. By default, it expects the ssl-checker service at `http://localhost:8000`. You can configure this via the `SSL_CHECKER_URL` environment variable:

```bash
docker run -p 8001:8001 -e SSL_CHECKER_URL=http://ssl-checker:8000 ssl-monitor
```

## Using Docker Compose (Recommended)

For running both ssl-checker and ssl-monitor together with proper service dependency and health checks:

```yaml
services:
  ssl-checker:
    build: ./ssl-checker
    container_name: ssl-checker
    ports:
      - "8000:8000"
    networks:
      - ssl-network
    healthcheck:
      test: ["CMD", "python", "-c", "import urllib.request; urllib.request.urlopen('http://localhost:8000/api/check?domain=google.com&port=443')"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    
  ssl-monitor:
    build: ./ssl-monitor
    container_name: ssl-monitor
    ports:
      - "8001:8001"
    environment:
      - SSL_CHECKER_URL=http://ssl-checker:8000
    depends_on:
      ssl-checker:
        condition: service_healthy
    networks:
      - ssl-network
    volumes:
      - ssl-monitor-data:/app/api

networks:
  ssl-network:
    driver: bridge

volumes:
  ssl-monitor-data:
```

**Key improvements:**
- **Health Check**: Ensures ssl-checker is ready before ssl-monitor starts
- **Service Dependency**: ssl-monitor waits for ssl-checker to be healthy
- **Shared Network**: Both containers communicate via bridge network
- **Volume Persistence**: Database persists across container restarts

Then run:
```bash
docker-compose up
```

This configuration resolves DNS resolution issues between containers and ensures proper startup order.

### Local Development

1. Install dependencies:
   ```bash
   cd api
   pip install -r requirements.txt
   ```

2. Run the application:
   ```bash
   cd api
   uvicorn main:app --reload --port 8001
   ```

3. Make sure ssl-checker is running on port 8000 (or configure SSL_CHECKER_URL environment variable)

## Using the Application

1. **Access the UI**: Open `http://localhost:8001` in your browser

2. **Check SSL Certificate**:
   - Enter a domain (e.g., `example.com`) or IP address
   - Optionally change the port (default: 443)
   - Click "Check SSL Certificate"
   - View the detailed results including certificate info, alerts, and recommendations

3. **View History**:
   - The history panel shows all previous checks
   - Click refresh to update the history
   - History is automatically updated after each new check

4. **Monitor Statistics**:
   - View total checks, successful checks, errors, and unique domains
   - Statistics update automatically every 30 seconds

## API Endpoints

### GET /api/check

Check SSL certificate and save to database.

**Query Parameters:**
- `domain` (string, optional): Domain name to check
- `ip` (string, optional): IP address to check
- `port` (integer, optional): Port number (default: 443)

**Example:**
```bash
curl "http://localhost:8001/api/check?domain=example.com"
```

### GET /api/history

Get SSL check history from database.

**Query Parameters:**
- `domain` (string, optional): Filter by domain
- `limit` (integer, optional): Maximum results (default: 50)

**Example:**
```bash
curl "http://localhost:8001/api/history?limit=10"
```

### GET /api/stats

Get monitoring statistics.

**Example:**
```bash
curl "http://localhost:8001/api/stats"
```

## Environment Variables

- `SSL_CHECKER_URL`: URL of the ssl-checker service (default: `http://localhost:8000`)

## Database

The application uses SQLite with SQLAlchemy ORM and Alembic for database migrations. The database includes three main tables:

### Database Schema

#### 1. Users Table
Stores user account information with role-based access control.

| Column        | Type     | Description                |
|---------------|----------|----------------------------|
| id            | Integer  | Primary key                |
| username      | String   | Unique username            |
| password_hash | String   | Bcrypt hashed password     |
| role          | String   | User role (admin/user)     |
| created_at    | DateTime | Account creation timestamp |

#### 2. Monitors Table
Stores monitoring configurations for automated SSL certificate checking.

| Column         | Type     | Description                      |
|----------------|----------|----------------------------------|
| id             | Integer  | Primary key                      |
| user_id        | Integer  | Foreign key to users.id          |
| domain         | String   | Domain to monitor                |
| check_interval | Integer  | Check interval in seconds        |
| webhook_url    | String   | Optional webhook for alerts      |
| last_check     | DateTime | Timestamp of last check          |
| status         | String   | Monitor status (active/paused)   |
| created_at     | DateTime | Monitor creation timestamp       |

#### 3. SSL Checks Table
Stores history of all SSL certificate checks.

| Column        | Type     | Description                    |
|---------------|----------|--------------------------------|
| id            | Integer  | Primary key                    |
| domain        | String   | Domain checked                 |
| ip            | String   | IP address                     |
| port          | Integer  | Port number (default: 443)     |
| status        | String   | Check status (success/error)   |
| ssl_status    | String   | SSL validation status          |
| server_status | String   | Server status                  |
| ip_status     | String   | IP geolocation status          |
| checked_at    | DateTime | Check timestamp                |
| response_data | Text     | Full JSON response from check  |

### Database Migrations

This project uses **Alembic** for database migrations to ensure smooth schema updates without data loss.

#### Running Migrations

Migrations run automatically when the container starts via the entrypoint script.

Manual migration commands (for development):

```bash
# Upgrade to the latest version
alembic upgrade head

# Check current migration status
alembic current

# View migration history
alembic history

# Create a new migration (after model changes)
alembic revision --autogenerate -m "Description of changes"

# Downgrade to previous version
alembic downgrade -1
```

#### Migration Files

Migrations are stored in `/app/api/alembic/versions/`. The initial migration creates all three tables.

### Sample Data

A sample data script is provided to populate the database with test data:

```bash
cd api
python create_sample_data.py
```

This creates:
- 2 users (admin and regular user)
- 3 monitors (2 active, 1 paused)
- 3 SSL check history records

**Sample Credentials:**
- Admin: `username=admin`, `password=admin123`
- User: `username=user1`, `password=user123`

### Data Persistence

The database file `ssl_monitor.db` is created automatically in the `/app/api` directory inside the container.

#### Using Docker Volumes

To persist the database across container restarts, the docker-compose configuration includes a named volume:

```yaml
volumes:
  - ssl-monitor-data:/app/api
```

This ensures:
- Database survives container restarts and rebuilds
- Migration state is preserved
- No data loss during updates

#### Manual Volume Mount

For development or custom deployments:

```bash
docker run -p 8001:8001 -v $(pwd)/data:/app/api ssl-monitor
```

**Important:** The database and migrations are designed to preserve existing data. When updating code:
1. New migrations are automatically applied on container startup
2. Existing data remains intact
3. Schema changes are handled gracefully through Alembic

## API Documentation

FastAPI provides automatic API documentation:
- Swagger UI: `http://localhost:8001/docs`
- ReDoc: `http://localhost:8001/redoc`

## Development Notes

- The application makes HTTP requests to the ssl-checker service for SSL validation
- All check results are stored in the SQLite database with full response data
- HTMX handles dynamic updates without page refreshes
- Bootstrap 5 provides responsive, mobile-friendly UI

## License

This project is part of the docker-images repository.
