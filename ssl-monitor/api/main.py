"""
SSL Monitor FastAPI Application

This application provides a monitoring interface for SSL certificates.
It stores check history in SQLite and makes API calls to the ssl-checker service.
"""
import json
import os
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
UI_DIR = BASE_DIR.parent / "ui"
STATIC_DIR = BASE_DIR.parent / "static"

# SSL Checker service URL - can be configured via environment variable
SSL_CHECKER_URL = os.getenv("SSL_CHECKER_URL", "http://localhost:8000")

app = FastAPI(
    title="SSL Monitor",
    description="Monitor SSL certificates with history tracking",
    version="1.0.0"
)

# Mount static files
app.mount("/static", StaticFiles(directory=str(STATIC_DIR)), name="static")

# Initialize database on startup
@app.on_event("startup")
def startup_event():
    init_db()


@app.get("/", response_class=HTMLResponse, summary="Serve the frontend UI")
async def serve_ui():
    """
    Serve the main frontend UI page.
    
    Returns:
        HTML response with the UI
    """
    return FileResponse(str(UI_DIR / "index.html"))


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
        # Make API call to ssl-checker service
        params = {"port": port}
        if domain:
            params["domain"] = domain
        elif ip:
            params["ip"] = ip
        else:
            raise HTTPException(status_code=400, detail="Provide either domain or ip")
        
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
                "data": json.loads(check.response_data) if check.response_data else None
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
