"""SSL certificate checking and validation logic."""

import ssl
import socket
from typing import Dict, List, Optional, Tuple, Any

from cert_utils import parse_certificate, create_empty_cert_info
from network_utils import create_ssl_connection
from constants import (
    STATUS_SUCCESS,
    STATUS_ERROR,
    ERROR_CERT_INVALID,
    ERROR_CN_MISMATCH,
    ERROR_CERT_EXPIRED,
    ERROR_SSL_ERROR,
    ERROR_UNKNOWN,
    ALERT_CN_MISMATCH,
    ALERT_CERT_EXPIRED,
    ALERT_CERT_INVALID,
    ALERT_SSL_HANDSHAKE_FAILED,
    ALERT_SSL_CHECK_FAILED,
    REC_MATCH_CN_SAN,
    REC_RENEW_CERT,
    REC_CHECK_CERT_CONFIG,
    REC_CHECK_SSL_CONFIG,
    REC_CHECK_SERVER_CONFIG,
)


def get_ssl_certificate_info(
    domain: Optional[str],
    ip: str,
    port: int = 443
) -> Tuple[Optional[Dict[str, Any]], str, Optional[str], List[str]]:
    """
    Retrieve and validate SSL certificate information for a domain/IP.
    
    This function attempts to connect to the specified host and retrieve its
    SSL certificate. It performs validation and provides detailed error
    information if the certificate is invalid.
    
    Args:
        domain: Domain name (optional, used for hostname verification)
        ip: IP address to connect to
        port: Port number (default: 443)
        
    Returns:
        Tuple of:
        - Certificate information dictionary (or None on error)
        - SSL status ('success' or 'error')
        - Error type (or None if successful)
        - List of recommendations
    """
    ssl_status = STATUS_SUCCESS
    ssl_error_type = None
    alerts = []
    recommendations = []
    cert_info = None
    
    try:
        # Attempt secure connection with verification
        ssock, cert = create_ssl_connection(domain, ip, port, verify=True)
        cert_info, alerts, recommendations = parse_certificate(cert, ssock)
        ssock.close()
        
    except ssl.CertificateError as e:
        # Certificate validation failed
        ssl_status = STATUS_ERROR
        ssl_error_type = ERROR_CERT_INVALID
        error_str = str(e).lower()
        
        # Determine specific certificate error type
        if "doesn't match" in error_str or "common name" in error_str or "hostname" in error_str:
            alerts.append(ALERT_CN_MISMATCH)
            recommendations.append(REC_MATCH_CN_SAN)
            ssl_error_type = ERROR_CN_MISMATCH
        elif "expired" in error_str or "not yet valid" in error_str:
            alerts.append(ALERT_CERT_EXPIRED)
            recommendations.append(REC_RENEW_CERT)
            ssl_error_type = ERROR_CERT_EXPIRED
        else:
            alerts.append(ALERT_CERT_INVALID)
            recommendations.append(REC_CHECK_CERT_CONFIG)
        
        # Try to get certificate details without verification
        cert_info = _get_unverified_cert_info(domain, ip, port, alerts)
        
    except ssl.SSLError as e:
        # SSL handshake or protocol error
        ssl_status = STATUS_ERROR
        ssl_error_type = ERROR_SSL_ERROR
        alerts.append(ALERT_SSL_HANDSHAKE_FAILED)
        recommendations.append(REC_CHECK_SSL_CONFIG)
        cert_info = create_empty_cert_info(alerts)
        
    except Exception as e:
        # Unknown error
        ssl_status = STATUS_ERROR
        ssl_error_type = ERROR_UNKNOWN
        alerts.append(ALERT_SSL_CHECK_FAILED)
        recommendations.append(REC_CHECK_SERVER_CONFIG)
        cert_info = create_empty_cert_info(alerts)
    
    return cert_info, ssl_status, ssl_error_type, recommendations


def _get_unverified_cert_info(
    domain: Optional[str],
    ip: str,
    port: int,
    alerts: List[str]
) -> Dict[str, Any]:
    """
    Attempt to retrieve certificate info without verification.
    
    This is used when the certificate fails validation but we still want
    to retrieve and display its details.
    
    Args:
        domain: Domain name (optional)
        ip: IP address
        port: Port number
        alerts: Existing alerts to include in the certificate info
        
    Returns:
        Certificate information dictionary
    """
    try:
        ssock, cert = create_ssl_connection(domain, ip, port, verify=False)
        cert_info, _, _ = parse_certificate(cert, ssock)
        ssock.close()
        
        # Override alerts with the specific validation failure alerts
        cert_info["alerts"] = alerts
        return cert_info
        
    except Exception:
        # Unable to retrieve certificate details even without verification
        return create_empty_cert_info(alerts)
