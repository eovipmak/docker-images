"""
SSL Checker FastAPI Application

This application provides REST API endpoints to check SSL certificate information,
web server technology, and IP geolocation details for domains and IP addresses.
"""

from datetime import datetime
from pathlib import Path
from typing import Optional

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from pydantic import BaseModel, Field

from constants import DEFAULT_SSL_PORT, STATUS_SUCCESS, STATUS_ERROR, STATUS_WARNING, UNKNOWN_SERVER
from network_utils import resolve_domain_to_ip, get_server_header, get_ip_geolocation
from ssl_checker import get_ssl_certificate_info


class BatchRequest(BaseModel):
    """Request model for batch SSL checking."""
    
    domains: list[str] = Field(default_factory=list, description="List of domain names to check")
    ips: list[str] = Field(default_factory=list, description="List of IP addresses to check")
    port: int = Field(default=DEFAULT_SSL_PORT, description="Port number to check")


# Get the absolute path to the UI directory
BASE_DIR = Path(__file__).resolve().parent
UI_DIR = BASE_DIR.parent / "ui"

app = FastAPI(
    title="SSL Checker API",
    description="Check SSL certificate information, server details, and IP geolocation",
    version="2.0.0"
)

# Mount static files for the UI
app.mount("/static", StaticFiles(directory=str(UI_DIR)), name="static")


def check_single_target(
    domain: Optional[str] = None,
    ip: Optional[str] = None,
    port: int = DEFAULT_SSL_PORT
) -> dict:
    """
    Perform SSL and server checks for a single domain or IP address.
    
    Args:
        domain: Domain name to check (optional)
        ip: IP address to check (optional)
        port: Port number (default: 443)
        
    Returns:
        Dictionary containing check results with status, timestamp, and data
    """
    checked_at = datetime.now().isoformat()
    
    try:
        # Validate input
        if not domain and not ip:
            return {
                "status": STATUS_ERROR,
                "timestamp": checked_at,
                "error": "Provide either domain or ip"
            }
        
        # Resolve domain to IP if domain is provided
        if domain:
            ip = resolve_domain_to_ip(domain)
        
        # Get SSL certificate information
        ssl_info, ssl_status, ssl_error_type, recommendations = get_ssl_certificate_info(
            domain, ip, port
        )
        
        # Get server information
        server = get_server_header(domain, ip, port)
        
        # Get IP geolocation information
        ip_info = get_ip_geolocation(ip)
        
        # Determine server status
        if server != UNKNOWN_SERVER:
            server_status = STATUS_SUCCESS
        elif ssl_status == STATUS_ERROR:
            server_status = STATUS_WARNING
        else:
            server_status = STATUS_ERROR
        
        # Determine IP info status
        ip_status = STATUS_SUCCESS if ip_info else STATUS_ERROR
        
        return {
            "status": STATUS_SUCCESS,
            "timestamp": checked_at,
            "data": {
                "domain": domain,
                "ip": ip,
                "port": port,
                "ssl": ssl_info,
                "server": server,
                "ip_info": ip_info,
                "checkedAt": checked_at,
                "recommendations": recommendations,
                "sslStatus": ssl_status,
                "serverStatus": server_status,
                "ipStatus": ip_status,
                "sslErrorType": ssl_error_type
            }
        }
        
    except ValueError as e:
        # Handle validation errors with sanitized messages
        error_msg = str(e)
        # Don't expose internal details if the error contains sensitive info
        if "Cannot resolve domain" in error_msg:
            return {
                "status": STATUS_ERROR,
                "timestamp": checked_at,
                "error": "Unable to resolve domain name"
            }
        return {
            "status": STATUS_ERROR,
            "timestamp": checked_at,
            "error": error_msg
        }
    except Exception as e:
        # Generic error - don't expose internal details
        return {
            "status": STATUS_ERROR,
            "timestamp": checked_at,
            "error": "An error occurred while checking SSL certificate. Please verify the domain/IP and try again."
        }


@app.get("/", summary="Serve the frontend UI")
def serve_ui():
    """
    Serve the main frontend UI page.
    
    Returns:
        HTML response with the UI
    """
    return FileResponse(str(UI_DIR / "index.html"))


@app.get("/api/check", summary="Check SSL certificate for a domain or IP")
def check_ssl(
    domain: Optional[str] = None,
    ip: Optional[str] = None,
    port: int = DEFAULT_SSL_PORT
) -> dict:
    """
    Check SSL certificate information for a single domain or IP address.
    
    Args:
        domain: Domain name to check (e.g., example.com)
        ip: IP address to check (e.g., 8.8.8.8)
        port: Port number (default: 443)
        
    Returns:
        JSON response with SSL certificate details, server info, and IP geolocation
    """
    return check_single_target(domain, ip, port)


@app.post("/api/batch_check", summary="Batch check SSL certificates for multiple domains/IPs")
def batch_check(request: BatchRequest) -> dict:
    """
    Check SSL certificates for multiple domains and/or IP addresses in a single request.
    
    Args:
        request: BatchRequest containing lists of domains and IPs to check
        
    Returns:
        JSON response with results for all checked targets
    """
    results = []
    
    # Check all domains
    for domain in request.domains:
        result = check_single_target(domain=domain, port=request.port)
        results.append(result)
    
    # Check all IPs
    for ip in request.ips:
        result = check_single_target(ip=ip, port=request.port)
        results.append(result)
    
    return {
        "status": STATUS_SUCCESS,
        "timestamp": datetime.now().isoformat(),
        "results": results
    }
