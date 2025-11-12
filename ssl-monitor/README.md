# SSL Monitor

A comprehensive SSL certificate monitoring service with database storage. Built with FastAPI, SQLite, Bootstrap 5, and HTMX.

## Features

- **Real-time SSL Certificate Checking**: Validate SSL/TLS certificates for domains and IP addresses
- **Database Storage**: Store check history in SQLite database for tracking over time
- **Interactive Dashboard**: Modern Bootstrap 5 UI with real-time statistics
- **Check History**: View all previous SSL certificate checks with detailed information
- **Security Alerts**: Identify weak TLS versions, expiring certificates, and security issues
- **Geolocation**: Get IP geolocation information for checked domains
- **REST API**: Full API endpoints for integration with other tools
- **Auto-refresh**: Dashboard auto-updates every 30 seconds

## Technology Stack

**Frontend:**
- HTML5
- Bootstrap 5 (CDN)
- Vanilla JavaScript (ES6+)
- HTMX for dynamic interactions
- Bootstrap Icons

**Backend:**
- FastAPI: High-performance web framework
- SQLAlchemy: Database ORM
- SQLite: Lightweight database
- Uvicorn: ASGI server
- Python 3.12

## Project Structure

```plaintext
ssl-monitor/
├── api/                    # Backend API
│   ├── main.py            # FastAPI application and API endpoints
│   ├── database.py        # Database models and configuration
│   ├── ssl_checker.py     # SSL certificate checking logic
│   ├── cert_utils.py      # Certificate parsing utilities
│   ├── network_utils.py   # Network operations
│   ├── constants.py       # Application constants
│   └── requirements.txt   # Python dependencies
├── ui/                     # Frontend UI
│   ├── index.html         # Main dashboard page
│   └── app.js             # JavaScript application logic
├── Dockerfile             # Docker configuration
└── README.md              # This file
```

## Quick Start

### Using Docker (Recommended)

```bash
# Build the image
docker build -t ssl-monitor .

# Run the container
docker run -p 8000:8000 -v $(pwd)/data:/app/api ssl-monitor
```

The application will be available at:
- Web Dashboard: `http://localhost:8000`
- API Documentation: `http://localhost:8000/docs`

### Local Development

1. Install dependencies:
   ```bash
   cd api
   pip install -r requirements.txt
   ```

2. Run the application:
   ```bash
   cd api
   uvicorn main:app --reload
   ```

## Web Dashboard

Access the web interface at `http://localhost:8000` to:
- Check SSL certificates for domains or IP addresses
- View real-time statistics (total checks, valid/invalid certificates, success rate)
- Browse check history with detailed information
- View certificate details including expiration dates, TLS versions, and cipher suites
- Get security alerts and recommendations

## API Endpoints

### GET /api/check

Check SSL certificate for a single domain or IP address.

**Query Parameters:**
- `domain` (string, optional): Domain name to check
- `ip` (string, optional): IP address to check
- `port` (integer, optional): Port number (default: 443)

**Example:**
```bash
curl "http://localhost:8000/api/check?domain=example.com"
```

### POST /api/batch_check

Check multiple domains/IPs in a single request.

**Request Body:**
```json
{
  "domains": ["example.com", "google.com"],
  "ips": ["8.8.8.8"],
  "port": 443
}
```

### GET /api/history

Get SSL check history from database.

**Query Parameters:**
- `domain` (string, optional): Filter by domain
- `limit` (integer, optional): Number of results (default: 50, max: 500)

**Example:**
```bash
curl "http://localhost:8000/api/history?limit=20"
```

### GET /api/history/{check_id}

Get detailed information for a specific check.

**Example:**
```bash
curl "http://localhost:8000/api/history/1"
```

### GET /api/stats

Get overall statistics.

**Response:**
```json
{
  "status": "success",
  "data": {
    "total_checks": 150,
    "valid_certificates": 140,
    "invalid_certificates": 10,
    "success_rate": 93.33
  }
}
```

### GET /health

Health check endpoint.

## Database

The application uses SQLite for storing check results. The database file `ssl_monitor.db` is created automatically in the `api` directory.

### Database Schema

**ssl_checks table:**
- `id`: Primary key
- `domain`: Domain name
- `ip`: IP address
- `port`: Port number
- `ssl_valid`: Boolean indicating if SSL is valid
- `certificate_info`: JSON string with certificate details
- `server_info`: Server information
- `ip_info`: JSON string with IP geolocation data
- `alerts`: JSON string with security alerts
- `recommendations`: JSON string with recommendations
- `checked_at`: Timestamp of the check
- `error_message`: Error message if check failed

## Docker Usage

### Build the Image

```bash
docker build -t ssl-monitor .
```

### Run with Persistent Storage

```bash
# Create a data directory for database persistence
mkdir -p data

# Run with volume mount
docker run -p 8000:8000 -v $(pwd)/data:/app/api ssl-monitor
```

### Using Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'
services:
  ssl-monitor:
    build: .
    ports:
      - "8000:8000"
    volumes:
      - ./data:/app/api
    restart: unless-stopped
```

Run with:
```bash
docker-compose up -d
```

## Features in Detail

### Real-time Statistics
- Total number of SSL checks performed
- Count of valid and invalid certificates
- Overall success rate percentage
- Auto-updates every 30 seconds

### Check History
- View all previous SSL certificate checks
- Sort by date (newest first)
- Filter by domain
- Detailed view for each check

### Security Alerts
The system provides intelligent security alerts for:
- Certificates expiring soon (< 30 days)
- Insecure TLS versions (TLSv1, TLSv1.1)
- Weak signature algorithms
- Certificate CN/SAN mismatches
- Expired certificates
- SSL handshake failures

### Recommendations
Get actionable recommendations such as:
- Renew certificate before expiration
- Upgrade to TLS 1.2 or higher
- Use stronger signature algorithms
- Verify certificate configuration

## Environment Variables

- `DATABASE_URL`: Database connection string (default: `sqlite:///./ssl_monitor.db`)

## Version History

### v1.0.0 - Initial Release
- SSL certificate monitoring with database storage
- Bootstrap 5 dashboard with statistics
- Check history and detailed views
- REST API endpoints
- Docker support
- SQLite database integration

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
