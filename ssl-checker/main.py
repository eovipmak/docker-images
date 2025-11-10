from fastapi import FastAPI, HTTPException
import requests
import ssl
import socket
import dns.resolver
from datetime import datetime
import time

app = FastAPI()

def resolve_domain_to_ip(domain: str) -> str:
    try:
        # Try DNS resolution
        answers = dns.resolver.resolve(domain, 'A')
        return answers[0].address
    except Exception:
        # Fallback to socket
        try:
            return socket.gethostbyname(domain)
        except socket.gaierror:
            raise ValueError("Invalid domain")

def get_ssl_info(domain: str, ip: str) -> dict:
    try:
        context = ssl.create_default_context()
        with socket.create_connection((ip, 443)) as sock:
            with context.wrap_socket(sock, server_hostname=domain) as ssock:
                cert = ssock.getpeercert()
                tls_version = ssock.version
                cipher_suite = ssock.cipher()[0]
                san = cert.get('subjectAltName', [])
                exp_seconds = ssl.cert_time_to_seconds(cert['notAfter'])
                now_seconds = time.time()
                days_until_expiration = int((exp_seconds - now_seconds) / 86400)
                return {
                    "subject": dict(x[0] for x in cert.get('subject', [])),
                    "issuer": dict(x[0] for x in cert.get('issuer', [])),
                    "version": cert.get('version'),
                    "serialNumber": cert.get('serialNumber'),
                    "notBefore": cert.get('notBefore'),
                    "notAfter": cert.get('notAfter'),
                    "signatureAlgorithm": cert.get('signatureAlgorithm'),
                    "tlsVersion": tls_version,
                    "cipherSuite": cipher_suite,
                    "subjectAltNames": san,
                    "daysUntilExpiration": days_until_expiration,
                }
>>>>>>> ba03a7490e277221dadde581adeb9a50281e188f
    except Exception as e:
        return {"error": str(e)}

def get_server_header(domain: str, ip: str) -> str:
<<<<<<< HEAD
    # Try HTTPS first for domain
    if domain:
        try:
            url = f"https://{domain}"
            response = requests.get(url, timeout=10)
            server = response.headers.get('Server')
            if server:
                return server
        except Exception:
            pass
    # Fallback to HTTP for domain or IP
    try:
        url = f"http://{domain or ip}"
        response = requests.get(url, timeout=10)
        server = response.headers.get('Server')
        if server:
            return server
    except Exception:
        pass
    # Banner grabbing fallback
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(5)
        sock.connect((ip, 80))
        sock.send(b"HEAD / HTTP/1.0\r\n\r\n")
        banner = sock.recv(1024).decode('utf-8', errors='ignore')
        sock.close()
        if 'Server:' in banner:
            return banner.split('Server:')[1].split('\r\n')[0].strip()
    except Exception:
        pass
    return 'Unknown'

def get_ip_info(ip: str) -> dict:
    try:
        response = requests.get(f"https://ipinfo.io/{ip}/json", timeout=10)
        return response.json()
    except Exception as e:
        return {"error": str(e)}

@app.get("/check")
def check_ssl(domain: str = None, ip: str = None):
    checked_at = datetime.now().isoformat()
    try:
        if not domain and not ip:
            return {
                "status": "error",
                "timestamp": checked_at,
                "error": "Provide either domain or ip"
            }
        
        if domain:
            ip = resolve_domain_to_ip(domain)
        else:
            # Assume ip is valid, but could add validation
            pass
        
        ssl_info = get_ssl_info(domain or ip, ip)
        server = get_server_header(domain or ip, ip)
        ip_info = get_ip_info(ip)
        
        # Standardize fields
        if 'error' in ssl_info:
            ssl_info = None
        if 'error' in ip_info:
            ip_info = None
        
        # Recommendations
        recommendations = []
        if ssl_info and ssl_info.get('daysUntilExpiration', 0) < 30:
            recommendations.append("Renew certificate soon")
        if ssl_info and ssl_info.get('tlsVersion') in ['TLSv1', 'TLSv1.1']:
            recommendations.append("Upgrade to TLS 1.2 or higher")
        
        return {
            "status": "success",
            "timestamp": checked_at,
            "data": {
                "domain": domain,
                "ip": ip,
                "ssl": ssl_info,
                "server": server,
                "ip_info": ip_info,
                "checkedAt": checked_at,
                "recommendations": recommendations
            }
        }
    except Exception as e:
        return {
            "status": "error",
            "timestamp": checked_at,
            "error": str(e)
        }

def get_server_header(domain: str, ip: str) -> str:
    try:
        url = f"https://{domain}"
        response = requests.get(url, timeout=10)
        return response.headers.get('Server', 'Unknown')
    except Exception:
=======
    # Try HTTPS first for domain
    if domain:
>>>>>>> ba03a7490e277221dadde581adeb9a50281e188f
        try:
            url = f"https://{domain}"
            response = requests.get(url, timeout=10)
            return response.headers.get('Server', 'Unknown')
        except Exception:
            pass
    # Fallback to HTTP for domain or IP
    try:
        url = f"http://{domain or ip}"
        response = requests.get(url, timeout=10)
        return response.headers.get('Server', 'Unknown')
    except Exception:
        return 'Unknown'

def get_ip_info(ip: str) -> dict:
    try:
        response = requests.get(f"https://ipinfo.io/{ip}/json", timeout=10)
        return response.json()
    except Exception as e:
        return {"error": str(e)}

@app.get("/check")
def check_ssl(domain: str = None, ip: str = None):
    timestamp = datetime.now().isoformat()
    try:
        if not domain and not ip:
            return {
                "status": "error",
                "timestamp": timestamp,
                "error": "Provide either domain or ip"
            }
        
        if domain:
            ip = resolve_domain_to_ip(domain)
        else:
            # Assume ip is valid, but could add validation
            pass
        
        ssl_info = get_ssl_info(domain or ip, ip)
        server = get_server_header(domain or ip, ip)
        ip_info = get_ip_info(ip)
        
        return {
            "status": "success",
            "timestamp": timestamp,
            "data": {
                "domain": domain,
                "ip": ip,
                "ssl": ssl_info,
                "server": server,
                "ip_info": ip_info
            }
        }
    except Exception as e:
        return {
            "status": "error",
            "timestamp": timestamp,
            "error": str(e)
        }

