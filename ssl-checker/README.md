# SSL Checker API

A FastAPI-based application that checks SSL certificate information, web server technology, and IP geolocation details for domains and IP addresses. The application provides REST API endpoints for easy integration with scripts, monitoring tools, or AI agents.

## Features

- **SSL Certificate Validation**: Retrieve detailed SSL/TLS certificate information including subject, issuer, validity period, and expiration warnings
- **Security Alerts**: Identify weak TLS versions, insecure cipher suites, and certificate issues
- **Domain Resolution**: Automatically resolve domain names to IP addresses
- **Server Detection**: Identify web server technology via Server HTTP header
- **IP Geolocation**: Fetch IP geolocation and ISP information from ipinfo.io
- **Batch Processing**: Check multiple domains/IPs in a single request
- **Smart Recommendations**: Get actionable recommendations for SSL/TLS improvements

## Architecture

The application is structured into modular components for better maintainability:

- `main.py`: FastAPI application and API endpoints
- `ssl_checker.py`: SSL certificate checking and validation logic
- `cert_utils.py`: Certificate parsing and utility functions
- `network_utils.py`: Network operations (DNS, HTTP, socket connections)
- `constants.py`: Application constants and configuration

## API Endpoints

### GET /check

Check SSL certificate information for a single domain or IP address.

**Query Parameters:**
- `domain` (string, optional): The domain name to check (e.g., example.com)
- `ip` (string, optional): The IP address to check (e.g., 93.184.216.34)
- `port` (integer, optional): Port number to check (default: 443)

**Note:** Provide either `domain` or `ip`, but not both.

**Example Request:**

```bash
curl "http://localhost:8000/check?domain=example.com"
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

### POST /batch_check

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
curl -X POST "http://localhost:8000/batch_check" \
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

## Docker Usage

### Build the Image

```bash
docker build -t ssl-checker .
```

### Run the Container

```bash
docker run -p 8000:8000 ssl-checker
```

The API will be available at `http://localhost:8000`.

## Local Development

If you want to run without Docker:

1. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

2. Run the app:
   ```bash
   uvicorn main:app --reload
   ```

## Dependencies

- FastAPI: Web framework
- Uvicorn: ASGI server
- Requests: HTTP library
- dnspython: DNS resolution