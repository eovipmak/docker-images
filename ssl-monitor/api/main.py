"""SSL Monitor API - FastAPI application with SQLite database."""
from fastapi import FastAPI, HTTPException, Depends, Query
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session
from pydantic import BaseModel
from typing import Optional, List
from datetime import datetime
import json
import os

from database import init_db, get_db, SSLCheck
from ssl_checker import get_ssl_certificate_info
from network_utils import resolve_domain_to_ip, get_server_header, get_ip_geolocation
from constants import DEFAULT_SSL_PORT, STATUS_SUCCESS, STATUS_ERROR, STATUS_WARNING, UNKNOWN_SERVER

app = FastAPI(
    title="SSL Monitor API",
    description="SSL Certificate monitoring service with database storage",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize database
init_db()


class CheckRequest(BaseModel):
    """Request model for single check."""
    domain: Optional[str] = None
    ip: Optional[str] = None
    port: int = 443


class BatchCheckRequest(BaseModel):
    """Request model for batch check."""
    domains: Optional[List[str]] = []
    ips: Optional[List[str]] = []
    port: int = 443


def save_check_result(db: Session, result: dict):
    """Save SSL check result to database."""
    try:
        data = result.get("data", {})
        ssl_info = data.get("ssl", {})
        
        db_check = SSLCheck(
            domain=data.get("domain"),
            ip=data.get("ip"),
            port=data.get("port", 443),
            ssl_valid=result.get("status") == "success" and data.get("sslStatus") == "success",
            certificate_info=json.dumps(ssl_info),
            server_info=data.get("server"),
            ip_info=json.dumps(data.get("ip_info", {})),
            alerts=json.dumps(ssl_info.get("alerts", [])),
            recommendations=json.dumps(data.get("recommendations", [])),
            checked_at=datetime.fromisoformat(result.get("timestamp").replace("Z", "+00:00")),
            error_message=result.get("error") if result.get("status") == "error" else None
        )
        db.add(db_check)
        db.commit()
        db.refresh(db_check)
        return db_check.id
    except Exception as e:
        db.rollback()
        print(f"Error saving check result: {e}")
        return None


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
    checked_at = datetime.utcnow().isoformat()
    
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


@app.get("/")
async def read_root():
    """Serve the main UI page."""
    ui_path = os.path.join(os.path.dirname(__file__), "..", "ui", "index.html")
    if os.path.exists(ui_path):
        return FileResponse(ui_path)
    return {"message": "SSL Monitor API", "version": "1.0.0"}


@app.get("/api/check")
async def check_certificate(
    domain: Optional[str] = Query(None),
    ip: Optional[str] = Query(None),
    port: int = Query(443),
    db: Session = Depends(get_db)
):
    """Check SSL certificate for a domain or IP address and save to database."""
    if not domain and not ip:
        raise HTTPException(status_code=400, detail="Either domain or ip must be provided")
    
    if domain and ip:
        raise HTTPException(status_code=400, detail="Provide either domain or ip, not both")
    
    try:
        result = check_single_target(domain=domain, ip=ip, port=port)
        
        # Save to database
        check_id = save_check_result(db, result)
        if check_id:
            result["check_id"] = check_id
        
        return result
    except Exception as e:
        # Don't expose internal error details to external users
        raise HTTPException(status_code=500, detail="An error occurred while checking the certificate")


@app.post("/api/batch_check")
async def batch_check_certificates(
    request: BatchCheckRequest,
    db: Session = Depends(get_db)
):
    """Check SSL certificates for multiple domains/IPs and save to database."""
    results = []
    
    # Check domains
    for domain in request.domains or []:
        try:
            result = check_single_target(domain=domain, port=request.port)
            check_id = save_check_result(db, result)
            if check_id:
                result["check_id"] = check_id
            results.append(result)
        except Exception as e:
            # Don't expose internal error details to external users
            results.append({
                "status": STATUS_ERROR,
                "timestamp": datetime.utcnow().isoformat(),
                "error": "An error occurred while checking the certificate",
                "domain": domain
            })
    
    # Check IPs
    for ip in request.ips or []:
        try:
            result = check_single_target(ip=ip, port=request.port)
            check_id = save_check_result(db, result)
            if check_id:
                result["check_id"] = check_id
            results.append(result)
        except Exception as e:
            # Don't expose internal error details to external users
            results.append({
                "status": STATUS_ERROR,
                "timestamp": datetime.utcnow().isoformat(),
                "error": "An error occurred while checking the certificate",
                "ip": ip
            })
    
    return {
        "status": STATUS_SUCCESS,
        "timestamp": datetime.utcnow().isoformat(),
        "results": results
    }


@app.get("/api/history")
async def get_check_history(
    domain: Optional[str] = Query(None),
    limit: int = Query(50, le=500),
    db: Session = Depends(get_db)
):
    """Get SSL check history from database."""
    query = db.query(SSLCheck)
    
    if domain:
        query = query.filter(SSLCheck.domain == domain)
    
    checks = query.order_by(SSLCheck.checked_at.desc()).limit(limit).all()
    
    return {
        "status": "success",
        "count": len(checks),
        "data": [
            {
                "id": check.id,
                "domain": check.domain,
                "ip": check.ip,
                "port": check.port,
                "ssl_valid": check.ssl_valid,
                "checked_at": check.checked_at.isoformat(),
                "server_info": check.server_info,
                "error_message": check.error_message
            }
            for check in checks
        ]
    }


@app.get("/api/history/{check_id}")
async def get_check_detail(check_id: int, db: Session = Depends(get_db)):
    """Get detailed information for a specific check."""
    check = db.query(SSLCheck).filter(SSLCheck.id == check_id).first()
    
    if not check:
        raise HTTPException(status_code=404, detail="Check not found")
    
    return {
        "status": "success",
        "data": {
            "id": check.id,
            "domain": check.domain,
            "ip": check.ip,
            "port": check.port,
            "ssl_valid": check.ssl_valid,
            "certificate_info": json.loads(check.certificate_info) if check.certificate_info else {},
            "server_info": check.server_info,
            "ip_info": json.loads(check.ip_info) if check.ip_info else {},
            "alerts": json.loads(check.alerts) if check.alerts else [],
            "recommendations": json.loads(check.recommendations) if check.recommendations else [],
            "checked_at": check.checked_at.isoformat(),
            "error_message": check.error_message
        }
    }


@app.get("/api/stats")
async def get_statistics(db: Session = Depends(get_db)):
    """Get overall statistics."""
    total_checks = db.query(SSLCheck).count()
    valid_checks = db.query(SSLCheck).filter(SSLCheck.ssl_valid == True).count()
    invalid_checks = db.query(SSLCheck).filter(SSLCheck.ssl_valid == False).count()
    
    return {
        "status": "success",
        "data": {
            "total_checks": total_checks,
            "valid_certificates": valid_checks,
            "invalid_certificates": invalid_checks,
            "success_rate": round(valid_checks / total_checks * 100, 2) if total_checks > 0 else 0
        }
    }


# Mount static files
static_path = os.path.join(os.path.dirname(__file__), "..", "ui")
if os.path.exists(static_path):
    app.mount("/static", StaticFiles(directory=static_path), name="static")


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy", "service": "ssl-monitor"}
