# SSL Checker App

This application checks SSL certificate information, web server technology, and IP details for a given domain or IP address. It provides a REST API endpoint for easy integration with scripts or AI agents.

## Features

- Resolve domain to IP address
- Retrieve SSL certificate details (subject, issuer, validity, etc.)
- Identify web server technology via Server header
- Fetch IP geolocation and ISP information from ipinfo.io

## API Endpoint

### GET /check

Query parameters:
- `domain`: The domain name to check (e.g., example.com)
- `ip`: The IP address to check (e.g., 8.8.8.8)

Provide either `domain` or `ip`, but not both.

#### Example Request

```bash
curl "http://localhost:8000/check?domain=example.com"
```

#### Example Response

```json
{
  "domain": "example.com",
  "ip": "93.184.216.34",
  "ssl": {
    "subject": {"commonName": "example.com"},
    "issuer": {"commonName": "DigiCert SHA2 Secure Server CA"},
    "version": 3,
    "serialNumber": "123456789",
    "notBefore": "2023-01-01T00:00:00Z",
    "notAfter": "2024-01-01T00:00:00Z",
    "signatureAlgorithm": "sha256WithRSAEncryption"
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
  }
}
```

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