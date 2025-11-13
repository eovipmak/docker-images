"""
SSL Monitor FastAPI Application

This application provides a monitoring interface for SSL certificates.
It stores check history in SQLite and makes API calls to the ssl-checker service.
"""
import ipaddress
import json
import os
import re
from datetime import datetime
from pathlib import Path
from typing import Optional

import requests
from fastapi import FastAPI, Depends, HTTPException
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse, HTMLResponse
from sqlalchemy.orm import Session

from database import init_db, get_db, SSLCheck

# Get the absolute path to directories
BASE_DIR = Path(__file__).resolve().parent
FRONTEND_DIST_DIR = BASE_DIR.parent / "frontend-dist"

# SSL Checker service URL - can be configured via environment variable
SSL_CHECKER_URL = os.getenv("SSL_CHECKER_URL", "http://localhost:8000")

# Domain validation regex - conservative pattern
DOMAIN_PATTERN = re.compile(
    r'^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$'
)

# Private/reserved IP ranges to reject
PRIVATE_IP_RANGES = [
    ipaddress.ip_network('10.0.0.0/8'),
    ipaddress.ip_network('172.16.0.0/12'),
    ipaddress.ip_network('192.168.0.0/16'),
    ipaddress.ip_network('127.0.0.0/8'),  # loopback
    ipaddress.ip_network('169.254.0.0/16'),  # link-local
    ipaddress.ip_network('::1/128'),  # IPv6 loopback
    ipaddress.ip_network('fe80::/10'),  # IPv6 link-local
    ipaddress.ip_network('fc00::/7'),  # IPv6 unique local
]


def validate_domain(domain: str) -> str:
    """
    Validate domain name to prevent SSRF attacks.
    
    Args:
        domain: Domain name to validate
        
    Returns:
        Normalized domain name
        
    Raises:
        HTTPException: If domain is invalid
    """
    if not domain:
        raise HTTPException(status_code=400, detail="Domain cannot be empty")
    
    # Strip and normalize
    domain = domain.strip().lower()
    
    # Reject if contains URL schemes or paths
    if '://' in domain or '/' in domain or '@' in domain:
        raise HTTPException(status_code=400, detail="Invalid domain format")
    
    # Validate against domain pattern
    if not DOMAIN_PATTERN.match(domain):
        raise HTTPException(status_code=400, detail="Invalid domain name")
    
    return domain


def validate_ip(ip: str) -> str:
    """
    Validate IP address and reject private/reserved ranges.
    
    Args:
        ip: IP address to validate
        
    Returns:
        Normalized IP address
        
    Raises:
        HTTPException: If IP is invalid or in private/reserved range
    """
    if not ip:
        raise HTTPException(status_code=400, detail="IP address cannot be empty")
    
    # Strip and normalize
    ip = ip.strip()
    
    try:
        ip_obj = ipaddress.ip_address(ip)
        
        # Check against private/reserved ranges
        for network in PRIVATE_IP_RANGES:
            if ip_obj in network:
                raise HTTPException(
                    status_code=400, 
                    detail="Private or reserved IP addresses are not allowed"
                )
        
        return str(ip_obj)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid IP address format")


def _safe_json_loads(data: str):
    """
    Safely load JSON data with error handling.
    
    Args:
        data: JSON string to parse
        
    Returns:
        Parsed JSON object or error sentinel
    """
    if not data:
        return None
    
    try:
        return json.loads(data)
    except (json.JSONDecodeError, TypeError):
        return {"error": "Invalid stored data"}


app = FastAPI(
    title="SSL Monitor",
    description="Monitor SSL certificates with history tracking",
    version="1.0.0"
)

# Mount static files for the React app
# Only mount the assets directory, which contains all the built JS/CSS files
if FRONTEND_DIST_DIR.exists() and (FRONTEND_DIST_DIR / "assets").exists():
    app.mount("/assets", StaticFiles(directory=str(FRONTEND_DIST_DIR / "assets")), name="assets")

# Initialize database on startup
@app.on_event("startup")
def startup_event():
    init_db()


@app.get("/", response_class=HTMLResponse, summary="Serve the frontend UI")
async def serve_ui():
    """
    Serve the main frontend UI page (React app).
    
    Returns:
        HTML response with the React app
    """
    index_path = FRONTEND_DIST_DIR / "index.html"
    if index_path.exists():
        return FileResponse(str(index_path))
    else:
        raise HTTPException(status_code=404, detail="Frontend not found")


@app.get("/api/check", summary="Check SSL certificate and save to database")
async def check_ssl(
    domain: Optional[str] = None,
    ip: Optional[str] = None,
    port: int = 443,
    db: Session = Depends(get_db)
):
    """
    Check SSL certificate by calling ssl-checker service and save result to database.
    
    Args:
        domain: Domain name to check
        ip: IP address to check
        port: Port number (default: 443)
        db: Database session
        
    Returns:
        JSON response with SSL certificate details and check history
    """
    try:
        # Validate input - either domain or IP must be provided
        if not domain and not ip:
            raise HTTPException(status_code=400, detail="Provide either domain or ip")
        
        # Validate and normalize inputs
        params = {"port": port}
        if domain:
            validated_domain = validate_domain(domain)
            params["domain"] = validated_domain
            domain = validated_domain
        elif ip:
            validated_ip = validate_ip(ip)
            params["ip"] = validated_ip
            ip = validated_ip
        
        response = requests.get(f"{SSL_CHECKER_URL}/api/check", params=params, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        # Extract data for database storage
        status = result.get("status", "error")
        data = result.get("data", {})
        
        # Create database record
        ssl_check = SSLCheck(
            domain=domain or data.get("domain"),
            ip=data.get("ip", ip),
            port=port,
            status=status,
            ssl_status=data.get("sslStatus", "unknown"),
            server_status=data.get("serverStatus", "unknown"),
            ip_status=data.get("ipStatus", "unknown"),
            checked_at=datetime.utcnow(),
            response_data=json.dumps(result)
        )
        
        db.add(ssl_check)
        db.commit()
        db.refresh(ssl_check)
        
        # Add check ID to response
        result["check_id"] = ssl_check.id
        
        return result
        
    except requests.RequestException as e:
        raise HTTPException(
            status_code=503,
            detail=f"SSL Checker service unavailable: {str(e)}"
        )
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"An error occurred: {str(e)}"
        )


@app.get("/api/history", summary="Get SSL check history")
async def get_history(
    domain: Optional[str] = None,
    limit: int = 50,
    db: Session = Depends(get_db)
):
    """
    Get SSL check history from database.
    
    Args:
        domain: Optional domain filter
        limit: Maximum number of results (default: 50)
        db: Database session
        
    Returns:
        List of SSL check history records
    """
    query = db.query(SSLCheck)
    
    if domain:
        query = query.filter(SSLCheck.domain == domain)
    
    checks = query.order_by(SSLCheck.checked_at.desc()).limit(limit).all()
    
    return {
        "status": "success",
        "count": len(checks),
        "history": [
            {
                "id": check.id,
                "domain": check.domain,
                "ip": check.ip,
                "port": check.port,
                "status": check.status,
                "ssl_status": check.ssl_status,
                "server_status": check.server_status,
                "ip_status": check.ip_status,
                "checked_at": check.checked_at.isoformat(),
                "data": _safe_json_loads(check.response_data)
            }
            for check in checks
        ]
    }


@app.get("/api/stats", summary="Get SSL monitoring statistics")
async def get_stats(db: Session = Depends(get_db)):
    """
    Get overall statistics from monitoring data.
    
    Args:
        db: Database session
        
    Returns:
        Statistics about SSL checks
    """
    total_checks = db.query(SSLCheck).count()
    successful_checks = db.query(SSLCheck).filter(SSLCheck.status == "success").count()
    error_checks = db.query(SSLCheck).filter(SSLCheck.status == "error").count()
    
    # Get unique domains checked
    unique_domains = db.query(SSLCheck.domain).distinct().count()
    
    return {
        "status": "success",
        "stats": {
            "total_checks": total_checks,
            "successful_checks": successful_checks,
            "error_checks": error_checks,
            "unique_domains": unique_domains
        }
    }


# Catch-all route for React Router (must be last)
# This handles client-side routing by serving index.html for all non-API routes
@app.get("/{full_path:path}", response_class=HTMLResponse, summary="Serve React app for client-side routing")
async def catch_all(full_path: str):
    """
    Catch-all route to serve the React app for client-side routing.
    This allows React Router to handle navigation.
    
    Args:
        full_path: The requested path
        
    Returns:
        HTML response with the React app
    """
    index_path = FRONTEND_DIST_DIR / "index.html"
    if index_path.exists():
        return FileResponse(str(index_path))
    else:
        raise HTTPException(status_code=404, detail="Frontend not found")
