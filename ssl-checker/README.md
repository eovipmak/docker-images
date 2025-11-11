# SSL Checker

A FastAPI-based application that checks SSL certificate information, web server technology, and IP geolocation details for domains and IP addresses. The application provides a web UI and REST API endpoints for easy integration with scripts, monitoring tools, or AI agents.

## Features

- **Web UI**: User-friendly interface for checking SSL certificates
- **SSL Certificate Validation**: Retrieve detailed SSL/TLS certificate information including subject, issuer, validity period, and expiration warnings
- **Security Alerts**: Identify weak TLS versions, insecure cipher suites, and certificate issues
- **Domain Resolution**: Automatically resolve domain names to IP addresses
- **Server Detection**: Identify web server technology via Server HTTP header
- **IP Geolocation**: Fetch IP geolocation and ISP information from ipinfo.io
- **Batch Processing**: Check multiple domains/IPs in a single request
- **Smart Recommendations**: Get actionable recommendations for SSL/TLS improvements

## Project Structure

```plaintext
ssl-checker/
├── api/                    # Backend API
│   ├── main.py            # FastAPI application and API endpoints
│   ├── ssl_checker.py     # SSL certificate checking and validation logic
│   ├── cert_utils.py      # Certificate parsing and utility functions
│   ├── network_utils.py   # Network operations (DNS, HTTP, socket connections)
│   ├── constants.py       # Application constants and configuration
│   └── requirements.txt   # Python dependencies
├── ui/                     # Frontend UI
│   ├── index.html         # Main HTML page
│   ├── styles.css         # Styles
│   └── app.js             # JavaScript application logic
├── Dockerfile             # Docker configuration
└── README.md              # This file
```

## Quick Start

### Using Docker (Recommended)

```bash
# Build the image
docker build -t ssl-checker .

# Run the container
docker run -p 8000:8000 ssl-checker
```

The application will be available at:
- Web UI: `http://localhost:8000`
- API Documentation: `http://localhost:8000/docs`

### Local Development

1. Install dependencies:
   ```bash
   cd api
   pip install -r requirements.txt
   ```

2. Run the app:
   ```bash
   uvicorn main:app --reload
   ```

## Web UI

Access the web interface at `http://localhost:8000` to:
- Check SSL certificates for single domains or IPs
- Perform batch checks on multiple targets
- View detailed certificate information, security alerts, and recommendations
- See server and geolocation information

## API Endpoints

### GET /api/check

Check SSL certificate information for a single domain or IP address.

**Query Parameters:**
- `domain` (string, optional): The domain name to check (e.g., example.com)
- `ip` (string, optional): The IP address to check (e.g., 93.184.216.34)
- `port` (integer, optional): Port number to check (default: 443)

**Note:** Provide either `domain` or `ip`, but not both.

**Example Request:**

```bash
curl "http://localhost:8000/api/check?domain=example.com"
```

**Example Response:**

```json
{
  "status": "success",
  "timestamp": "2025-11-11T12:00:00.000000",
  "data": {
    "domain": "example.com",
    "ip": "93.184.216.34",
    "port": 443,
    "ssl": {
      "subject": {"commonName": "example.com"},
      "issuer": {"commonName": "DigiCert SHA2 Secure Server CA"},
      "version": 3,
      "serialNumber": "123456789",
      "notBefore": "Jan  1 00:00:00 2023 GMT",
      "notAfter": "Jan  1 00:00:00 2024 GMT",
      "signatureAlgorithm": "sha256WithRSAEncryption",
      "tlsVersion": "TLSv1.3",
      "cipherSuite": "TLS_AES_256_GCM_SHA384",
      "subjectAltNames": [["DNS", "example.com"], ["DNS", "www.example.com"]],
      "daysUntilExpiration": 45,
      "alerts": []
    },
    "server": "nginx/1.21.3",
    "ip_info": {
      "ip": "93.184.216.34",
      "hostname": "example.com",
      "city": "Norwell",
      "region": "Massachusetts",
      "country": "US",
      "loc": "42.1596,-70.8217",
      "org": "AS15133 MCI Communications Services, Inc. d/b/a Verizon Business",
      "postal": "02061",
      "timezone": "America/New_York"
    },
    "checkedAt": "2025-11-11T12:00:00.000000",
    "recommendations": [],
    "sslStatus": "success",
    "serverStatus": "success",
    "ipStatus": "success",
    "sslErrorType": null
  }
}
```

### POST /api/batch_check

Check SSL certificates for multiple domains and/or IP addresses in a single request.

**Request Body:**

```json
{
  "domains": ["example.com", "google.com"],
  "ips": ["8.8.8.8"],
  "port": 443
}
```

**Parameters:**
- `domains` (array of strings): List of domain names to check
- `ips` (array of strings): List of IP addresses to check
- `port` (integer, optional): Port number for all checks (default: 443)

**Example Request:**

```bash
curl -X POST "http://localhost:8000/api/batch_check" \
  -H "Content-Type: application/json" \
  -d '{
    "domains": ["example.com", "google.com"],
    "ips": ["8.8.8.8"],
    "port": 443
  }'
```

**Example Response:**

```json
{
  "status": "success",
  "timestamp": "2025-11-11T12:00:00.000000",
  "results": [
    {
      "status": "success",
      "timestamp": "2025-11-11T12:00:00.000000",
      "data": {
        "domain": "example.com",
        "ip": "93.184.216.34",
        ...
      }
    },
    {
      "status": "success",
      "timestamp": "2025-11-11T12:00:01.000000",
      "data": {
        "domain": "google.com",
        ...
      }
    },
    ...
  ]
}
```

### Error Response

When an error occurs, the API returns:

```json
{
  "status": "error",
  "timestamp": "2025-11-11T12:00:00.000000",
  "error": "Error description"
}
```

## Security Alerts and Recommendations

The API provides intelligent security alerts and recommendations:

**Alerts:**
- Certificate expiring soon (< 30 days)
- Insecure TLS version (TLSv1, TLSv1.1)
- Weak signature algorithm (SHA1)
- Certificate CN/SAN mismatch
- Expired certificates
- SSL handshake failures

**Recommendations:**
- Renew certificate soon
- Upgrade to TLS 1.2 or higher
- Use stronger signature algorithm
- Ensure CN/SAN matches domain
- Check certificate/SSL/server configuration

## Technology Stack

**Frontend:**
- Vanilla HTML5, CSS3, and JavaScript (ES6+)
- Modern responsive design
- No framework dependencies (lightweight and fast)
- Fetch API for backend communication

**Backend:**
- FastAPI: High-performance web framework
- Uvicorn: ASGI server
- Python 3.12

**Dependencies:**
- FastAPI: Web framework
- Uvicorn: ASGI server with WebSocket support
- Requests: HTTP library
- dnspython: DNS resolution
- aiofiles: Async file operations for static file serving

## Docker Usage

### Build the Image

```bash
docker build -t ssl-checker .
```

### Run the Container

```bash
docker run -p 8000:8000 ssl-checker
```

The application will be available at `http://localhost:8000`.

## Local Development

If you want to run without Docker:

1. Install dependencies:
   ```bash
   cd api
   pip install -r requirements.txt
   ```

2. Run the app:
   ```bash
   cd api
   uvicorn main:app --reload
   ```

## Version History

### v2.0.0 - Frontend Integration
- Added web UI for easier SSL certificate checking
- Reorganized project structure (separated `api/` and `ui/` folders)
- Updated API endpoints with `/api` prefix
- Enhanced Docker configuration for serving both API and UI
- Improved documentation
