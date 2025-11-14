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
from typing import Optional, List

import requests
from fastapi import FastAPI, Depends, HTTPException
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse, HTMLResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field, field_validator
from sqlalchemy.orm import Session

from database import init_db, get_db, SSLCheck, User, Alert, AlertConfig
from auth import fastapi_users, auth_backend, current_active_user, get_refresh_jwt_strategy
from schemas import UserRead, UserCreate, AlertConfigCreate, AlertConfigUpdate, AlertConfigRead, AlertRead
from alert_service import process_ssl_check_alerts, get_or_create_alert_config

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


# Pydantic models for request validation
class DomainCreate(BaseModel):
    """Request model for creating a new domain to monitor"""
    domain: str = Field(..., description="Domain name to monitor", min_length=1, max_length=255)
    port: int = Field(default=443, description="Port number to check", ge=1, le=65535)
    
    @field_validator('domain')
    @classmethod
    def validate_domain_field(cls, v: str) -> str:
        """Validate domain name format"""
        v = v.strip().lower()
        if '://' in v or '/' in v or '@' in v:
            raise ValueError("Invalid domain format")
        if not DOMAIN_PATTERN.match(v):
            raise ValueError("Invalid domain name")
        return v


app = FastAPI(
    title="SSL Monitor",
    description="Monitor SSL certificates with history tracking and JWT authentication",
    version="1.0.0"
)

# Add CORS middleware to allow frontend to access the API
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, replace with specific origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include authentication routes
app.include_router(
    fastapi_users.get_auth_router(auth_backend),
    prefix="/auth/jwt",
    tags=["auth"],
)

app.include_router(
    fastapi_users.get_register_router(UserRead, UserCreate),
    prefix="/auth",
    tags=["auth"],
)

app.include_router(
    fastapi_users.get_reset_password_router(),
    prefix="/auth",
    tags=["auth"],
)

app.include_router(
    fastapi_users.get_verify_router(UserRead),
    prefix="/auth",
    tags=["auth"],
)

app.include_router(
    fastapi_users.get_users_router(UserRead, UserCreate),
    prefix="/users",
    tags=["users"],
)

# Mount static files for the React app
# Only mount the assets directory, which contains all the built JS/CSS files
if FRONTEND_DIST_DIR.exists() and (FRONTEND_DIST_DIR / "assets").exists():
    app.mount("/assets", StaticFiles(directory=str(FRONTEND_DIST_DIR / "assets")), name="assets")

# Initialize database on startup
@app.on_event("startup")
def startup_event():
    init_db()


# Custom /auth/me endpoint for getting current user info
@app.get("/auth/me", response_model=UserRead, tags=["auth"])
async def get_current_user(user: User = Depends(current_active_user)):
    """
    Get current authenticated user information.
    
    Returns:
        Current user data
    """
    return user


# Custom refresh token endpoint
@app.post("/auth/jwt/refresh", tags=["auth"])
async def refresh_token(user: User = Depends(current_active_user)):
    """
    Refresh access token using a valid access token.
    This endpoint returns a new access token and refresh token.
    
    Returns:
        New access token and refresh token
    """
    from jose import jwt
    from datetime import datetime, timedelta
    from auth import SECRET, REFRESH_SECRET
    
    # Generate new access token (1 hour)
    access_token_payload = {
        "sub": str(user.id),
        "aud": ["fastapi-users:auth"],
        "exp": datetime.utcnow() + timedelta(seconds=3600)
    }
    access_token = jwt.encode(access_token_payload, SECRET, algorithm="HS256")
    
    # Generate new refresh token (7 days)
    refresh_token_payload = {
        "sub": str(user.id),
        "aud": ["fastapi-users:refresh"],
        "exp": datetime.utcnow() + timedelta(seconds=604800)
    }
    refresh_token = jwt.encode(refresh_token_payload, REFRESH_SECRET, algorithm="HS256")
    
    return {
        "access_token": access_token,
        "refresh_token": refresh_token,
        "token_type": "bearer"
    }


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
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Check SSL certificate by calling ssl-checker service and save result to database.
    Results are associated with the authenticated user for data isolation.
    Automatically processes alerts based on user's alert configuration.
    
    Args:
        domain: Domain name to check
        ip: IP address to check
        port: Port number (default: 443)
        db: Database session
        user: Current authenticated user
        
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
        
        # Create database record with user_id for isolation
        ssl_check = SSLCheck(
            user_id=user.id,
            organization_id=user.organization_id,
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
        
        # Process alerts based on user's configuration
        alert_config = get_or_create_alert_config(db, user.id, user.organization_id)
        alerts_created = process_ssl_check_alerts(
            db, user.id, user.organization_id,
            domain or data.get("domain", ""),
            result,
            alert_config
        )
        
        # Add check ID and alerts to response
        result["check_id"] = ssl_check.id
        result["alerts_created"] = len(alerts_created)
        
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
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Get SSL check history from database for the current user.
    Data is filtered by user_id to ensure isolation.
    
    Args:
        domain: Optional domain filter
        limit: Maximum number of results (default: 50)
        db: Database session
        user: Current authenticated user
        
    Returns:
        List of SSL check history records for the current user
    """
    # Base query filtered by user_id for data isolation
    query = db.query(SSLCheck).filter(SSLCheck.user_id == user.id)
    
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
async def get_stats(
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Get statistics from monitoring data for the current user.
    Data is filtered by user_id to ensure isolation.
    
    Args:
        db: Database session
        user: Current authenticated user
        
    Returns:
        Statistics about SSL checks for the current user
    """
    # Filter all queries by user_id
    total_checks = db.query(SSLCheck).filter(SSLCheck.user_id == user.id).count()
    successful_checks = db.query(SSLCheck).filter(
        SSLCheck.user_id == user.id,
        SSLCheck.status == "success"
    ).count()
    error_checks = db.query(SSLCheck).filter(
        SSLCheck.user_id == user.id,
        SSLCheck.status == "error"
    ).count()
    
    # Get unique domains checked by this user
    unique_domains = db.query(SSLCheck.domain).filter(
        SSLCheck.user_id == user.id
    ).distinct().count()
    
    return {
        "status": "success",
        "stats": {
            "total_checks": total_checks,
            "successful_checks": successful_checks,
            "error_checks": error_checks,
            "unique_domains": unique_domains
        }
    }


@app.get("/api/domains", summary="Get list of monitored domains")
async def get_domains(
    limit: int = 100,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Get list of unique domains that have been checked by the current user.
    Data is filtered by user_id to ensure isolation.
    
    Args:
        limit: Maximum number of domains to return (default: 100)
        db: Database session
        user: Current authenticated user
        
    Returns:
        List of unique domains with their latest check information for the current user
    """
    # Get unique domains with their latest check, filtered by user_id
    subquery = (
        db.query(
            SSLCheck.domain,
            db.func.max(SSLCheck.checked_at).label("latest_check")
        )
        .filter(SSLCheck.domain.isnot(None))
        .filter(SSLCheck.user_id == user.id)
        .group_by(SSLCheck.domain)
        .subquery()
    )
    
    # Join to get full details of the latest check for each domain
    domains_data = (
        db.query(SSLCheck)
        .join(
            subquery,
            db.and_(
                SSLCheck.domain == subquery.c.domain,
                SSLCheck.checked_at == subquery.c.latest_check
            )
        )
        .filter(SSLCheck.user_id == user.id)
        .order_by(SSLCheck.checked_at.desc())
        .limit(limit)
        .all()
    )
    
    return {
        "status": "success",
        "count": len(domains_data),
        "domains": [
            {
                "domain": check.domain,
                "ip": check.ip,
                "port": check.port,
                "status": check.status,
                "ssl_status": check.ssl_status,
                "last_checked": check.checked_at.isoformat(),
            }
            for check in domains_data
        ]
    }


@app.post("/api/domains", summary="Add a new domain and check SSL")
async def add_domain(
    domain_data: DomainCreate,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Add a new domain to monitor and perform initial SSL check.
    Domain is associated with the authenticated user for data isolation.
    
    Args:
        domain_data: Domain information with validation
        db: Database session
        user: Current authenticated user
        
    Returns:
        SSL check result for the newly added domain
    """
    try:
        # Validate domain using existing validation function
        validated_domain = validate_domain(domain_data.domain)
        
        # Perform SSL check using ssl-checker service
        params = {
            "domain": validated_domain,
            "port": domain_data.port
        }
        
        response = requests.get(f"{SSL_CHECKER_URL}/api/check", params=params, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        # Extract data for database storage
        status = result.get("status", "error")
        data = result.get("data", {})
        
        # Create database record with user_id for isolation
        ssl_check = SSLCheck(
            user_id=user.id,
            organization_id=user.organization_id,
            domain=validated_domain,
            ip=data.get("ip"),
            port=domain_data.port,
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
        
        # Return the check result
        return {
            "status": "success",
            "message": f"Domain {validated_domain} added and checked successfully",
            "data": {
                "id": ssl_check.id,
                "domain": ssl_check.domain,
                "ip": ssl_check.ip,
                "port": ssl_check.port,
                "ssl_status": ssl_check.ssl_status,
                "checked_at": ssl_check.checked_at.isoformat(),
                "check_result": result
            }
        }
        
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


@app.get("/api/alert-config", response_model=AlertConfigRead, summary="Get user's alert configuration")
async def get_alert_config(
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Get the current user's alert configuration.
    Creates a default configuration if none exists.
    
    Args:
        db: Database session
        user: Current authenticated user
        
    Returns:
        User's alert configuration
    """
    config = get_or_create_alert_config(db, user.id, user.organization_id)
    return config


@app.post("/api/alert-config", response_model=AlertConfigRead, summary="Create or update alert configuration")
async def create_or_update_alert_config(
    config_data: AlertConfigUpdate,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Create or update the user's alert configuration.
    
    Args:
        config_data: Alert configuration data
        db: Database session
        user: Current authenticated user
        
    Returns:
        Updated alert configuration
    """
    # Get or create config
    config = get_or_create_alert_config(db, user.id, user.organization_id)
    
    # Update fields that are provided
    update_data = config_data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(config, field, value)
    
    config.updated_at = datetime.utcnow()
    
    db.commit()
    db.refresh(config)
    
    return config


@app.get("/api/alerts", response_model=List[AlertRead], summary="Get user's alerts")
async def get_alerts(
    unread_only: bool = False,
    unresolved_only: bool = False,
    limit: int = 50,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Get alerts for the current user.
    
    Args:
        unread_only: Filter for unread alerts only
        unresolved_only: Filter for unresolved alerts only
        limit: Maximum number of alerts to return (default: 50)
        db: Database session
        user: Current authenticated user
        
    Returns:
        List of alerts for the current user
    """
    query = db.query(Alert).filter(Alert.user_id == user.id)
    
    if unread_only:
        query = query.filter(Alert.is_read == False)
    
    if unresolved_only:
        query = query.filter(Alert.is_resolved == False)
    
    alerts = query.order_by(Alert.created_at.desc()).limit(limit).all()
    
    return alerts


@app.patch("/api/alerts/{alert_id}/read", response_model=AlertRead, summary="Mark alert as read")
async def mark_alert_read(
    alert_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Mark an alert as read.
    
    Args:
        alert_id: Alert ID
        db: Database session
        user: Current authenticated user
        
    Returns:
        Updated alert
    """
    alert = db.query(Alert).filter(
        Alert.id == alert_id,
        Alert.user_id == user.id
    ).first()
    
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    
    alert.is_read = True
    db.commit()
    db.refresh(alert)
    
    return alert


@app.patch("/api/alerts/{alert_id}/resolve", response_model=AlertRead, summary="Mark alert as resolved")
async def mark_alert_resolved(
    alert_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Mark an alert as resolved.
    
    Args:
        alert_id: Alert ID
        db: Database session
        user: Current authenticated user
        
    Returns:
        Updated alert
    """
    alert = db.query(Alert).filter(
        Alert.id == alert_id,
        Alert.user_id == user.id
    ).first()
    
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    
    alert.is_resolved = True
    alert.resolved_at = datetime.utcnow()
    db.commit()
    db.refresh(alert)
    
    return alert


@app.delete("/api/alerts/{alert_id}", summary="Delete an alert")
async def delete_alert(
    alert_id: int,
    db: Session = Depends(get_db),
    user: User = Depends(current_active_user)
):
    """
    Delete an alert.
    
    Args:
        alert_id: Alert ID
        db: Database session
        user: Current authenticated user
        
    Returns:
        Success message
    """
    alert = db.query(Alert).filter(
        Alert.id == alert_id,
        Alert.user_id == user.id
    ).first()
    
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    
    db.delete(alert)
    db.commit()
    
    return {"status": "success", "message": "Alert deleted"}


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
