"""
Alert service for detecting SSL certificate issues and sending notifications
"""
import json
import socket
import ipaddress
import requests
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List
from urllib.parse import urlparse
from sqlalchemy.orm import Session

from database import Alert, AlertConfig, SSLCheck


def parse_expiry_date(expiry_str: str) -> Optional[datetime]:
    """Parse expiry date from SSL response"""
    if not expiry_str:
        return None
    
    try:
        # Try multiple date formats
        for fmt in ["%Y-%m-%d %H:%M:%S", "%Y-%m-%dT%H:%M:%S", "%Y-%m-%d"]:
            try:
                return datetime.strptime(expiry_str.split('.')[0], fmt)
            except ValueError:
                continue
        return None
    except Exception:
        return None


def check_certificate_expiry(ssl_data: Dict[str, Any], alert_config: AlertConfig) -> Optional[tuple]:
    """
    Check if certificate is expiring soon or expired
    
    Returns:
        tuple of (alert_type, severity, message, days_remaining) or None
    """
    if not ssl_data:
        return None
    
    # Get SSL info from the data structure
    ssl_info = ssl_data.get('ssl', {})
    
    # Try to get daysUntilExpiration directly (from ssl-checker API)
    days_remaining = ssl_info.get('daysUntilExpiration')
    
    if days_remaining is None:
        # Fallback: try to parse from notAfter date
        expiry_str = ssl_info.get('notAfter')
        if not expiry_str:
            return None
        
        expiry_date = parse_expiry_date(expiry_str)
        if not expiry_date:
            return None
        
        now = datetime.utcnow()
        days_remaining = (expiry_date - now).days
    
    # Check if certificate is expired
    if days_remaining < 0 and alert_config.alert_cert_expired:
        return (
            'expired',
            'critical',
            f'SSL certificate has expired {abs(days_remaining)} days ago',
            days_remaining
        )
    
    # Check thresholds
    if days_remaining <= 1 and alert_config.alert_1_day:
        return (
            'expiring_soon',
            'critical',
            f'SSL certificate expiring in {days_remaining} day(s)',
            days_remaining
        )
    elif days_remaining <= 7 and alert_config.alert_7_days:
        return (
            'expiring_soon',
            'high',
            f'SSL certificate expiring in {days_remaining} days',
            days_remaining
        )
    elif days_remaining <= 30 and alert_config.alert_30_days:
        return (
            'expiring_soon',
            'medium',
            f'SSL certificate expiring in {days_remaining} days',
            days_remaining
        )
    
    return None


def check_ssl_errors(ssl_data: Dict[str, Any], alert_config: AlertConfig) -> Optional[tuple]:
    """
    Check for SSL validation errors
    
    Returns:
        tuple of (alert_type, severity, message) or None
    """
    if not alert_config.alert_ssl_errors:
        return None
    
    ssl_status = ssl_data.get('sslStatus', 'unknown')
    
    if ssl_status in ['error', 'invalid', 'untrusted']:
        alerts = ssl_data.get('alerts', [])
        error_messages = [alert.get('message', '') for alert in alerts if alert.get('type') == 'error']
        
        message = 'SSL validation error'
        if error_messages:
            message = f"SSL validation error: {', '.join(error_messages[:2])}"
        
        return ('ssl_error', 'high', message)
    
    return None


def check_geo_changes(
    domain: str,
    current_ip: str,
    current_geo: Dict[str, Any],
    db: Session,
    user_id: int,
    alert_config: AlertConfig
) -> Optional[tuple]:
    """
    Check for geolocation changes
    
    Returns:
        tuple of (alert_type, severity, message) or None
    """
    if not alert_config.alert_geo_changes or not current_ip or not current_geo:
        return None
    
    # Get the previous check for this domain
    previous_check = (
        db.query(SSLCheck)
        .filter(SSLCheck.user_id == user_id, SSLCheck.domain == domain)
        .order_by(SSLCheck.checked_at.desc())
        .offset(1)  # Skip the current check
        .first()
    )
    
    if not previous_check or not previous_check.response_data:
        return None
    
    try:
        previous_data = json.loads(previous_check.response_data)
        previous_geo = previous_data.get('data', {}).get('geolocation', {})
        
        if not previous_geo:
            return None
        
        # Check if country changed
        current_country = current_geo.get('country', '')
        previous_country = previous_geo.get('country', '')
        
        if current_country and previous_country and current_country != previous_country:
            return (
                'geo_change',
                'medium',
                f'Geolocation changed from {previous_country} to {current_country}'
            )
    except (json.JSONDecodeError, KeyError):
        pass
    
    return None


def create_alert(
    db: Session,
    user_id: int,
    organization_id: Optional[int],
    domain: str,
    alert_type: str,
    severity: str,
    message: str,
    deduplication_window_hours: int = 24
) -> Optional[Alert]:
    """
    Create a new alert in the database with deduplication
    
    Args:
        db: Database session
        user_id: User ID
        organization_id: Organization ID (optional)
        domain: Domain name
        alert_type: Type of alert
        severity: Severity level
        message: Alert message
        deduplication_window_hours: Hours to check for duplicate alerts (default 24)
    
    Returns:
        New or existing alert, or None if skipped due to deduplication
    """
    # Check for existing unresolved alert in the deduplication window
    cutoff_time = datetime.utcnow() - timedelta(hours=deduplication_window_hours)
    existing_alert = (
        db.query(Alert)
        .filter(
            Alert.user_id == user_id,
            Alert.domain == domain,
            Alert.alert_type == alert_type,
            Alert.is_resolved == False,
            Alert.created_at >= cutoff_time
        )
        .first()
    )
    
    if existing_alert:
        # Update the timestamp on existing alert instead of creating duplicate
        existing_alert.created_at = datetime.utcnow()
        existing_alert.message = message  # Update message in case details changed
        db.commit()
        db.refresh(existing_alert)
        return existing_alert
    
    # Create new alert
    alert = Alert(
        user_id=user_id,
        organization_id=organization_id,
        domain=domain,
        alert_type=alert_type,
        severity=severity,
        message=message,
        is_read=False,
        is_resolved=False,
        created_at=datetime.utcnow()
    )
    db.add(alert)
    db.commit()
    db.refresh(alert)
    return alert


def _is_private_ip(ip_str: str) -> bool:
    """Check if IP address is private, loopback, or reserved"""
    try:
        ip = ipaddress.ip_address(ip_str)
        return (
            ip.is_private or
            ip.is_loopback or
            ip.is_link_local or
            ip.is_reserved or
            ip.is_multicast
        )
    except ValueError:
        return True  # Treat invalid IPs as private for safety


def _validate_webhook_url(webhook_url: str) -> bool:
    """
    Validate webhook URL to prevent SSRF attacks
    
    Returns:
        True if URL is safe to use, False otherwise
    """
    if not webhook_url:
        return False
    
    try:
        # Parse URL
        parsed = urlparse(webhook_url)
        
        # Require http or https scheme
        if parsed.scheme not in ['http', 'https']:
            return False
        
        # Reject empty hosts
        if not parsed.hostname:
            return False
        
        # Reject localhost
        if parsed.hostname.lower() in ['localhost', '127.0.0.1', '::1']:
            return False
        
        # If hostname is an IP, validate it directly
        try:
            if _is_private_ip(parsed.hostname):
                return False
        except ValueError:
            pass  # Not an IP, continue with DNS resolution
        
        # Perform DNS resolution and validate all returned IPs
        try:
            addr_info = socket.getaddrinfo(parsed.hostname, None)
            for info in addr_info:
                ip_str = info[4][0]
                if _is_private_ip(ip_str):
                    return False
        except socket.gaierror:
            return False  # DNS resolution failed
        
        return True
        
    except Exception:
        return False


def send_webhook_notification(webhook_url: str, alert_data: Dict[str, Any]) -> bool:
    """
    Send alert notification to webhook URL with SSRF protection
    
    Returns:
        True if successful, False otherwise
    """
    result = send_webhook_notification_detailed(webhook_url, alert_data)
    return result['success']


def send_webhook_notification_detailed(webhook_url: str, alert_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Send alert notification to webhook URL with SSRF protection and detailed error reporting
    
    Returns:
        Dict with 'success' (bool), 'error' (str or None), 'status_code' (int or None)
    """
    if not webhook_url:
        return {
            'success': False,
            'error': 'No webhook URL provided',
            'status_code': None
        }
    
    # Validate URL to prevent SSRF
    if not _validate_webhook_url(webhook_url):
        return {
            'success': False,
            'error': 'Invalid or unsafe webhook URL. URL must use http/https scheme and cannot point to private/local IP addresses.',
            'status_code': None
        }
    
    try:
        response = requests.post(
            webhook_url,
            json=alert_data,
            headers={'Content-Type': 'application/json'},
            timeout=10
        )
        success = response.status_code in [200, 201, 202, 204]
        return {
            'success': success,
            'error': None if success else f'Webhook returned status code {response.status_code}',
            'status_code': response.status_code
        }
    except requests.Timeout:
        return {
            'success': False,
            'error': 'Webhook request timed out after 10 seconds',
            'status_code': None
        }
    except requests.ConnectionError as e:
        return {
            'success': False,
            'error': f'Failed to connect to webhook URL: {str(e)}',
            'status_code': None
        }
    except requests.RequestException as e:
        return {
            'success': False,
            'error': f'Failed to send webhook: {str(e)}',
            'status_code': None
        }


def process_ssl_check_alerts(
    db: Session,
    user_id: int,
    organization_id: Optional[int],
    domain: str,
    ssl_check_data: Dict[str, Any],
    alert_config: AlertConfig
) -> List[Alert]:
    """
    Process SSL check results and create alerts based on user configuration
    and per-domain monitor settings.
    
    Returns:
        List of created alerts
    """
    if not alert_config or not alert_config.enabled:
        return []
    
    # Import Monitor here to avoid circular dependency
    from database import Monitor
    
    # Check if domain has monitor with alerts disabled
    monitor = db.query(Monitor).filter(
        Monitor.user_id == user_id,
        Monitor.domain == domain
    ).first()
    
    if monitor and not monitor.alerts_enabled:
        # Alerts disabled for this specific domain
        return []
    
    created_alerts = []
    ssl_data = ssl_check_data.get('data', {})
    
    # Check for certificate expiry
    expiry_check = check_certificate_expiry(ssl_data, alert_config)
    if expiry_check:
        alert_type, severity, message, _ = expiry_check
        alert = create_alert(
            db, user_id, organization_id, domain,
            alert_type, severity, message
        )
        created_alerts.append(alert)
    
    # Check for SSL errors
    error_check = check_ssl_errors(ssl_data, alert_config)
    if error_check:
        alert_type, severity, message = error_check
        alert = create_alert(
            db, user_id, organization_id, domain,
            alert_type, severity, message
        )
        created_alerts.append(alert)
    
    # Check for geo changes
    current_ip = ssl_data.get('ip', '')
    current_geo = ssl_data.get('geolocation', {})
    geo_check = check_geo_changes(
        domain, current_ip, current_geo, db, user_id, alert_config
    )
    if geo_check:
        alert_type, severity, message = geo_check
        alert = create_alert(
            db, user_id, organization_id, domain,
            alert_type, severity, message
        )
        created_alerts.append(alert)
    
    # Send webhook notifications
    # Use domain-specific webhook if available, otherwise use global config
    webhook_url = None
    if monitor and monitor.webhook_url:
        webhook_url = monitor.webhook_url
    elif alert_config.webhook_url:
        webhook_url = alert_config.webhook_url
    
    if created_alerts and webhook_url:
        for alert in created_alerts:
            alert_data = {
                'domain': alert.domain,
                'alert_type': alert.alert_type,
                'severity': alert.severity,
                'message': alert.message,
                'created_at': alert.created_at.isoformat()
            }
            send_webhook_notification(webhook_url, alert_data)
    
    return created_alerts


def get_or_create_alert_config(db: Session, user_id: int, organization_id: Optional[int] = None) -> AlertConfig:
    """Get existing alert config or create default one for user"""
    config = db.query(AlertConfig).filter(AlertConfig.user_id == user_id).first()
    
    if not config:
        config = AlertConfig(
            user_id=user_id,
            organization_id=organization_id,
            enabled=True,
            alert_30_days=True,
            alert_7_days=True,
            alert_1_day=True,
            alert_ssl_errors=True,
            alert_geo_changes=False,
            alert_cert_expired=True,
            email_notifications=False
        )
        db.add(config)
        db.commit()
        db.refresh(config)
    
    return config
