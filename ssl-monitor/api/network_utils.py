"""Network utility functions for DNS resolution and SSL connections."""

import logging
import socket
import ssl
import re
from typing import Optional, Tuple, List, Dict, Any

import dns.resolver
import requests

logger = logging.getLogger(__name__)

from constants import (
    DEFAULT_SSL_PORT,
    DEFAULT_HTTP_PORT,
    CONNECTION_TIMEOUT,
    SOCKET_TIMEOUT,
    IPINFO_API_URL,
    IP_API_COM_URL,
    UNKNOWN_SERVER,
)


def is_valid_domain(domain: str) -> bool:
    """
    Validate domain name format.
    
    Args:
        domain: Domain name to validate
        
    Returns:
        True if domain format is valid, False otherwise
    """
    # Basic domain validation: alphanumeric, dots, hyphens
    # No scheme (http://, etc.) or special chars
    pattern = r'^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$|^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$'
    return bool(re.match(pattern, domain))


def is_valid_ip(ip: str) -> bool:
    """
    Validate IP address format.
    
    Args:
        ip: IP address to validate
        
    Returns:
        True if IP format is valid, False otherwise
    """
    try:
        socket.inet_aton(ip)
        return True
    except socket.error:
        return False


def resolve_domain_to_ip(domain: str) -> str:
    """
    Resolve a domain name to its IP address.
    
    Args:
        domain: Domain name to resolve
        
    Returns:
        IP address string
        
    Raises:
        ValueError: If domain cannot be resolved
    """
    try:
        # Try DNS resolution using dnspython
        answers = dns.resolver.resolve(domain, 'A')
        return answers[0].address
    except Exception:
        # Fallback to standard socket resolution
        try:
            return socket.gethostbyname(domain)
        except socket.gaierror:
            raise ValueError(f"Cannot resolve domain: {domain}")


def _normalize_ip_api_com_response(data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Normalize ip-api.com response to standard format.
    
    Args:
        data: Raw response from ip-api.com API
        
    Returns:
        Normalized dictionary in standard format with lat/lon as Optional[float]
    """
    # ip-api.com already returns in the desired format
    # Just ensure all expected fields are present
    normalized = {
        "query": data.get("query", ""),
        "status": data.get("status", ""),
        "continent": data.get("continent", ""),
        "continentCode": data.get("continentCode", ""),
        "country": data.get("country", ""),
        "countryCode": data.get("countryCode", ""),
        "region": data.get("region", ""),
        "regionName": data.get("regionName", ""),
        "city": data.get("city", ""),
        "district": data.get("district", ""),
        "zip": data.get("zip", ""),
        "lat": data.get("lat"),  # None if not provided
        "lon": data.get("lon"),  # None if not provided
        "isp": data.get("isp", ""),
        "org": data.get("org", ""),
        "as": data.get("as", ""),
        "asname": data.get("asname", ""),
        "reverse": data.get("reverse", ""),
        "mobile": data.get("mobile", False),
        "proxy": data.get("proxy", False),
        "hosting": data.get("hosting", False),
    }
    return normalized


def _normalize_ipinfo_response(data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Normalize ipinfo.io response to standard format.
    
    The ipinfo.io response has a different structure, so we map it to match
    the ip-api.com format as closely as possible.
    
    Args:
        data: Raw response from ipinfo.io API
        
    Returns:
        Normalized dictionary in standard format with lat/lon as Optional[float]
    """
    # Parse location coordinates if available
    lat, lon = None, None
    if "loc" in data and data["loc"]:
        try:
            parts = data["loc"].split(",")
            if len(parts) == 2:
                lat = float(parts[0])
                lon = float(parts[1])
        except (ValueError, AttributeError):
            pass
    
    # ipinfo.io doesn't provide continent info, we'll leave it empty
    # Map fields from ipinfo.io to standard format
    normalized = {
        "query": data.get("ip", ""),
        "status": "success",  # ipinfo.io doesn't have status field, assume success if we got data
        "continent": "",  # Not provided by ipinfo.io
        "continentCode": "",  # Not provided by ipinfo.io
        "country": data.get("country", ""),
        "countryCode": data.get("country", ""),  # ipinfo.io uses 2-letter code in "country" field
        "region": data.get("region", ""),
        "regionName": data.get("region", ""),  # ipinfo.io doesn't separate region code and name
        "city": data.get("city", ""),
        "district": "",  # Not provided by ipinfo.io
        "zip": data.get("postal", ""),
        "lat": lat,  # None if not available
        "lon": lon,  # None if not available
        "isp": "",  # Not provided by ipinfo.io in free tier
        "org": data.get("org", ""),
        "as": data.get("org", ""),  # ipinfo.io combines AS info in org field
        "asname": "",  # Not separately provided by ipinfo.io
        "reverse": data.get("hostname", ""),
        "mobile": False,  # Not provided by ipinfo.io
        "proxy": False,  # Not provided by ipinfo.io
        "hosting": False,  # Not provided by ipinfo.io
    }
    return normalized


def get_ip_geolocation(ip: str) -> Optional[Dict[str, Any]]:
    """
    Fetch geolocation and ISP information for an IP address.
    
    This function tries ip-api.com first (more detailed information),
    then falls back to ipinfo.io if the first attempt fails.
    Both responses are normalized to a standard format.
    
    Args:
        ip: IP address to lookup
        
    Returns:
        Dictionary with IP information in standardized format or None on error.
        Note: lat/lon fields may be None if coordinates are not available.
    """
    # Validate IP address before making API calls
    if not is_valid_ip(ip):
        logger.warning(f"Invalid IP address provided: {ip}")
        return None
    
    # Try ip-api.com first (primary source)
    try:
        url = IP_API_COM_URL.format(ip=ip)
        response = requests.get(url, timeout=CONNECTION_TIMEOUT)
        response.raise_for_status()
        data = response.json()
        
        # Check if the API returned success status
        if data.get('status') == 'success':
            return _normalize_ip_api_com_response(data)
        # If status is not success (e.g., 'fail'), fall through to ipinfo.io
    except Exception as e:
        # Log the error with details for debugging
        logger.error(f"Failed to fetch IP info from ip-api.com for {ip}: {e}", exc_info=True)
        # Fall through to ipinfo.io on any error
    
    # Fallback to ipinfo.io
    try:
        url = IPINFO_API_URL.format(ip=ip)
        response = requests.get(url, timeout=CONNECTION_TIMEOUT)
        response.raise_for_status()
        data = response.json()
        
        # Return None if there's an error in the response
        if 'error' in data:
            return None
        
        return _normalize_ipinfo_response(data)
    except Exception as e:
        # Log the error with details for debugging
        logger.error(f"Failed to fetch IP info from ipinfo.io for {ip}: {e}", exc_info=True)
        return None


def get_server_header(domain: Optional[str], ip: str, port: int = DEFAULT_SSL_PORT) -> str:
    """
    Attempt to retrieve the Server header from an HTTP/HTTPS request.
    
    This function tries multiple methods to retrieve the server information:
    1. HTTPS request to domain (if domain is provided)
    2. HTTP request to domain or IP
    3. Banner grabbing via socket connection
    
    Args:
        domain: Optional domain name
        ip: IP address
        port: Port number (default: 443)
        
    Returns:
        Server header value or 'Unknown' if not found
    """
    # Validate inputs to prevent SSRF attacks
    if domain and not is_valid_domain(domain):
        return UNKNOWN_SERVER
    if not is_valid_ip(ip):
        return UNKNOWN_SERVER
    
    # Try HTTPS first for domain
    if domain:
        try:
            url = f"https://{domain}:{port}" if port != DEFAULT_SSL_PORT else f"https://{domain}"
            response = requests.get(url, timeout=CONNECTION_TIMEOUT, allow_redirects=True)
            server = response.headers.get('Server')
            if server:
                return server
        except Exception:
            pass
    
    # Fallback to HTTP for domain or IP
    try:
        target = domain or ip
        url = f"http://{target}:{port}" if port != DEFAULT_HTTP_PORT else f"http://{target}"
        response = requests.get(url, timeout=CONNECTION_TIMEOUT, allow_redirects=True)
        server = response.headers.get('Server')
        if server:
            return server
    except Exception:
        pass
    
    # Banner grabbing fallback
    try:
        # Use port 80 for banner if the requested port is 443 (HTTPS default)
        banner_port = port if port != DEFAULT_SSL_PORT else DEFAULT_HTTP_PORT
        
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(SOCKET_TIMEOUT)
        sock.connect((ip, banner_port))
        sock.send(b"HEAD / HTTP/1.0\r\n\r\n")
        banner = sock.recv(1024).decode('utf-8', errors='ignore')
        sock.close()
        
        # Parse Server header from banner
        if 'Server:' in banner:
            server = banner.split('Server:')[1].split('\r\n')[0].strip()
            return server
    except Exception:
        pass
    
    return UNKNOWN_SERVER


def create_ssl_connection(
    domain: Optional[str],
    ip: str,
    port: int,
    verify: bool = True
) -> Tuple[ssl.SSLSocket, Dict[str, Any]]:
    """
    Create an SSL connection and retrieve the certificate.
    
    Args:
        domain: Domain name for hostname verification (can be None)
        ip: IP address to connect to
        port: Port number
        verify: Whether to verify the certificate (default: True)
        
    Returns:
        Tuple of (SSL socket, certificate dictionary)
        
    Raises:
        Various SSL and socket exceptions on connection failure
    """
    context = ssl.create_default_context()
    # Explicitly set minimum TLS version to 1.2 for security
    context.minimum_version = ssl.TLSVersion.TLSv1_2
    
    if not verify:
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE
    
    # Use domain for hostname verification if provided, otherwise use IP
    hostname = domain if domain else ip
    
    sock = socket.create_connection((ip, port), timeout=CONNECTION_TIMEOUT)
    ssock = context.wrap_socket(sock, server_hostname=hostname)
    ssock.settimeout(CONNECTION_TIMEOUT)
    cert = ssock.getpeercert()
    
    return ssock, cert