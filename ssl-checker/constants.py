"""
Constants for SSL Checker application.

This module contains all configuration constants, threshold values,
alert messages, and error types used throughout the SSL checker.
"""

# SSL/TLS Configuration
DEFAULT_SSL_PORT = 443
DEFAULT_HTTP_PORT = 80
CONNECTION_TIMEOUT = 10
SOCKET_TIMEOUT = 5

# Certificate Expiration Thresholds (in days)
EXPIRATION_WARNING_THRESHOLD = 30

# Deprecated/Weak TLS Versions
DEPRECATED_TLS_VERSIONS = ['TLSv1', 'TLSv1.1']

# Weak Cipher and Algorithm Components
WEAK_CIPHER_COMPONENTS = ['RC4', 'DES', '3DES', 'MD5', 'SHA1', 'NULL']

# Alert Messages
ALERT_CERT_EXPIRING_SOON = "Certificate expires soon"
ALERT_INSECURE_TLS = "Insecure TLS version"
ALERT_WEAK_SIGNATURE = "Weak signature algorithm"
ALERT_CN_MISMATCH = "Certificate CN/SAN does not match domain"
ALERT_CERT_EXPIRED = "Certificate has expired or is not yet valid"
ALERT_CERT_INVALID = "Certificate is invalid"
ALERT_SSL_HANDSHAKE_FAILED = "SSL handshake failed"
ALERT_SSL_CHECK_FAILED = "SSL check failed"

# Recommendation Messages
REC_RENEW_CERT_SOON = "Renew certificate soon"
REC_UPGRADE_TLS = "Upgrade to TLS 1.2 or higher"
REC_STRONGER_SIGNATURE = "Use stronger signature algorithm"
REC_MATCH_CN_SAN = "Ensure CN/SAN matches domain"
REC_RENEW_CERT = "Renew the certificate"
REC_CHECK_CERT_CONFIG = "Check certificate configuration"
REC_CHECK_SSL_CONFIG = "Check SSL/TLS configuration"
REC_CHECK_SERVER_CONFIG = "Check server configuration"

# Error Types
ERROR_CERT_INVALID = "CERTIFICATE_INVALID"
ERROR_CN_MISMATCH = "CN_MISMATCH"
ERROR_CERT_EXPIRED = "CERT_EXPIRED"
ERROR_SSL_ERROR = "SSL_ERROR"
ERROR_UNKNOWN = "UNKNOWN_ERROR"

# Status Values
STATUS_SUCCESS = "success"
STATUS_ERROR = "error"
STATUS_WARNING = "warning"

# IP Info API
IPINFO_API_URL = "https://ipinfo.io/{ip}/json"
IP_API_COM_URL = "http://ip-api.com/json/{ip}?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,isp,org,as,asname,reverse,mobile,proxy,hosting,query"

# Server Detection
UNKNOWN_SERVER = "Unknown"
