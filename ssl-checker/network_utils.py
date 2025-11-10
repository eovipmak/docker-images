"""Network utility functions for DNS resolution and SSL connections."""

import socket
import ssl
from typing import Optional, Tuple, List, Dict, Any

import dns.resolver
import requests

from constants import (
    DEFAULT_SSL_PORT,
    DEFAULT_HTTP_PORT,
    CONNECTION_TIMEOUT,
    SOCKET_TIMEOUT,
    IPINFO_API_URL,
    UNKNOWN_SERVER,
)


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


def get_ip_geolocation(ip: str) -> Optional[Dict[str, Any]]:
    """
    Fetch geolocation and ISP information for an IP address.
    
    Args:
        ip: IP address to lookup
        
    Returns:
        Dictionary with IP information or None on error
    """
    try:
        url = IPINFO_API_URL.format(ip=ip)
        response = requests.get(url, timeout=CONNECTION_TIMEOUT)
        response.raise_for_status()
        data = response.json()
        
        # Return None if there's an error in the response
        if 'error' in data:
            return None
        
        return data
    except Exception:
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
    
    if not verify:
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE
    
    sock = socket.create_connection((ip, port))
    ssock = context.wrap_socket(sock, server_hostname=domain or ip)
    cert = ssock.getpeercert()
    
    return ssock, cert
