"""Utility functions for certificate parsing and validation."""

import ssl
import time
from typing import Dict, List, Optional, Tuple, Any

from constants import (
    EXPIRATION_WARNING_THRESHOLD,
    DEPRECATED_TLS_VERSIONS,
    WEAK_CIPHER_COMPONENTS,
    ALERT_CERT_EXPIRING_SOON,
    ALERT_INSECURE_TLS,
    ALERT_WEAK_SIGNATURE,
    REC_RENEW_CERT_SOON,
    REC_UPGRADE_TLS,
    REC_STRONGER_SIGNATURE,
)


def get_days_until_expiration(cert: Dict[str, Any]) -> int:
    """
    Calculate the number of days until the certificate expires.
    
    Args:
        cert: SSL certificate dictionary
        
    Returns:
        Number of days until expiration (can be negative if expired)
    """
    exp_seconds = ssl.cert_time_to_seconds(cert['notAfter'])
    now_seconds = time.time()
    return int((exp_seconds - now_seconds) / 86400)


def evaluate_cipher_strength(cipher: Optional[str]) -> str:
    """
    Evaluate the strength of a cipher suite.
    
    Args:
        cipher: Cipher suite name
        
    Returns:
        'strong', 'weak', or 'unknown'
    """
    if not cipher:
        return 'unknown'
    
    cipher_upper = cipher.upper()
    if any(weak in cipher_upper for weak in WEAK_CIPHER_COMPONENTS):
        return 'weak'
    
    return 'strong'


def parse_certificate(
    cert: Dict[str, Any],
    ssock: ssl.SSLSocket
) -> Tuple[Dict[str, Any], List[str], List[str]]:
    """
    Parse SSL certificate and extract relevant information.
    
    Args:
        cert: SSL certificate dictionary from getpeercert()
        ssock: SSL socket connection
        
    Returns:
        Tuple of (certificate info dict, alerts list, recommendations list)
    """
    alerts = []
    recommendations = []
    
    # Extract TLS and cipher information
    cipher_info = ssock.cipher()
    tls_version = cipher_info[1] if cipher_info else None
    cipher_suite = cipher_info[0] if cipher_info else None
    
    # Extract Subject Alternative Names
    subject_alt_names = cert.get('subjectAltName', [])
    
    # Calculate expiration
    days_until_expiration = get_days_until_expiration(cert)
    
    # Check for certificate expiration warning
    if days_until_expiration < EXPIRATION_WARNING_THRESHOLD:
        alerts.append(ALERT_CERT_EXPIRING_SOON)
        recommendations.append(REC_RENEW_CERT_SOON)
    
    # Check for deprecated TLS version
    if tls_version and tls_version in DEPRECATED_TLS_VERSIONS:
        alerts.append(ALERT_INSECURE_TLS)
        recommendations.append(REC_UPGRADE_TLS)
    
    # Check for weak signature algorithm
    signature_algorithm = cert.get('signatureAlgorithm')
    sig_alg_str = None
    if signature_algorithm:
        sig_alg_str = signature_algorithm.decode('utf-8', errors='ignore')
        if 'SHA1' in sig_alg_str:
            alerts.append(ALERT_WEAK_SIGNATURE)
            recommendations.append(REC_STRONGER_SIGNATURE)
    
    cert_info = {
        "subject": dict(x[0] for x in cert.get('subject', [])),
        "issuer": dict(x[0] for x in cert.get('issuer', [])),
        "version": cert.get('version'),
        "serialNumber": cert.get('serialNumber'),
        "notBefore": cert.get('notBefore'),
        "notAfter": cert.get('notAfter'),
        "signatureAlgorithm": sig_alg_str,
        "tlsVersion": tls_version,
        "cipherSuite": cipher_suite,
        "subjectAltNames": subject_alt_names,
        "daysUntilExpiration": days_until_expiration,
        "alerts": alerts
    }
    
    return cert_info, alerts, recommendations


def create_empty_cert_info(alerts: List[str]) -> Dict[str, Any]:
    """
    Create an empty certificate info dictionary for error cases.
    
    Args:
        alerts: List of alert messages to include
        
    Returns:
        Certificate info dictionary with None values
    """
    return {
        "subject": None,
        "issuer": None,
        "version": None,
        "serialNumber": None,
        "notBefore": None,
        "notAfter": None,
        "signatureAlgorithm": None,
        "tlsVersion": None,
        "cipherSuite": None,
        "subjectAltNames": [],
        "daysUntilExpiration": None,
        "alerts": alerts
    }
