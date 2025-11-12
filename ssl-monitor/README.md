# SSL Monitor

A web-based SSL certificate monitoring application with history tracking. This application provides a user-friendly interface to check SSL certificates and stores the check history in a SQLite database.

## Features

- **SSL Certificate Checking**: Check SSL certificates for domains and IP addresses
- **History Tracking**: Store and view history of all SSL checks in SQLite database
- **Statistics Dashboard**: View overall monitoring statistics
- **Real-time Updates**: Uses HTMX for dynamic, real-time UI updates
- **Bootstrap 5 UI**: Modern, responsive interface using Bootstrap 5
- **API Integration**: Makes API calls to the ssl-checker service for SSL validation

## Technology Stack

**Frontend:**
- HTML5
- Bootstrap 5 (CDN)
- Vanilla JavaScript
- HTMX for dynamic updates

**Backend:**
- FastAPI (Python web framework)
- SQLAlchemy ORM
- SQLite database
- Uvicorn ASGI server

## Project Structure

```plaintext
ssl-monitor/
├── api/                    # Backend API
│   ├── main.py            # FastAPI application and endpoints
│   ├── database.py        # Database models and configuration
│   └── requirements.txt   # Python dependencies
├── ui/                     # Frontend UI
│   └── index.html         # Main HTML page
├── static/                 # Static files
│   ├── styles.css         # Custom CSS styles
│   └── app.js             # JavaScript application logic
├── Dockerfile             # Docker configuration
└── README.md              # This file
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

### Using Docker Compose (Recommended)

For running both ssl-checker and ssl-monitor together:

```yaml
version: '3.8'

services:
  ssl-checker:
    build: ./ssl-checker
    ports:
      - "8000:8000"
    
  ssl-monitor:
    build: ./ssl-monitor
    ports:
      - "8001:8001"
    environment:
      - SSL_CHECKER_URL=http://ssl-checker:8000
    depends_on:
      - ssl-checker
```

Then run:
```bash
docker-compose up
```

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

The application uses SQLite to store check history. The database file `ssl_monitor.db` is created automatically in the `/app/api` directory inside the container.

To persist the database across container restarts, mount a volume:

```bash
docker run -p 8001:8001 -v $(pwd)/data:/app/api ssl-monitor
```

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
