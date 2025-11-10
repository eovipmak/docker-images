from fastapi import FastAPI, HTTPException
import requests
import ssl
import socket
import dns.resolver
from datetime import datetime
import time
from pydantic import BaseModel

class BatchRequest(BaseModel):
    domains: list[str] = []
    ips: list[str] = []
    port: int = 443

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

def get_ssl_info(domain: str, ip: str, port: int = 443) -> dict:
    try:
        context = ssl.create_default_context()
        with socket.create_connection((ip, port)) as sock:
            with context.wrap_socket(sock, server_hostname=domain) as ssock:
                cert = ssock.getpeercert()
                cipher_info = ssock.cipher()
                tls_version = cipher_info[1] if cipher_info else None
                cipher_suite = cipher_info[0] if cipher_info else None
                san = cert.get('subjectAltName', [])
                exp_seconds = ssl.cert_time_to_seconds(cert['notAfter'])
                now_seconds = time.time()
                days_until_expiration = int((exp_seconds - now_seconds) / 86400)
                
                # Validation and alerts
                alerts = []
                if days_until_expiration < 30:
                    alerts.append("Certificate expires soon")
                if tls_version and tls_version in ['TLSv1', 'TLSv1.1']:
                    alerts.append("Insecure TLS version")
                sig_alg = cert.get('signatureAlgorithm')
                if sig_alg and 'SHA1' in sig_alg.decode('utf-8', errors='ignore'):
                    alerts.append("Weak signature algorithm")
                
                def evaluate_cipher_strength(cipher):
                    if not cipher:
                        return 'unknown'
                    weak_ciphers = ['RC4', 'DES', '3DES', 'MD5', 'SHA1', 'NULL']
                    if any(weak in cipher.upper() for weak in weak_ciphers):
                        return 'weak'
                    else:
                        return 'strong'
                
                return {
                    "subject": dict(x[0] for x in cert.get('subject', [])),
                    "issuer": dict(x[0] for x in cert.get('issuer', [])),
                    "version": cert.get('version'),
                    "serialNumber": cert.get('serialNumber'),
                    "notBefore": cert.get('notBefore'),
                    "notAfter": cert.get('notAfter'),
                    "signatureAlgorithm": sig_alg.decode('utf-8', errors='ignore') if sig_alg else None,
                    "tlsVersion": tls_version,
                    "cipherSuite": cipher_suite,
                    "subjectAltNames": san,
                    "daysUntilExpiration": days_until_expiration,
                    "alerts": alerts
                }
    except Exception as e:
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
            "alerts": ["No SSL certificate installed"]
        }

def get_server_header(domain: str, ip: str, port: int = 443) -> str:
    # Try HTTPS first for domain
    if domain:
        try:
            url = f"https://{domain}:{port}" if port != 443 else f"https://{domain}"
            response = requests.get(url, timeout=10, allow_redirects=True)
            server = response.headers.get('Server')
            if server:
                return server
        except Exception:
            pass
    # Fallback to HTTP for domain or IP
    try:
        url = f"http://{domain or ip}:{port}" if port != 80 else f"http://{domain or ip}"
        response = requests.get(url, timeout=10, allow_redirects=True)
        server = response.headers.get('Server')
        if server:
            return server
    except Exception:
        pass
    # Banner grabbing fallback
    try:
        banner_port = port if port != 443 else 80  # assume 80 for banner if 443
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(5)
        sock.connect((ip, banner_port))
        sock.send(b"HEAD / HTTP/1.0\r\n\r\n")
        banner = sock.recv(1024).decode('utf-8', errors='ignore')
        sock.close()
        if 'Server:' in banner:
            server = banner.split('Server:')[1].split('\r\n')[0].strip()
            return server
    except Exception:
        pass
    return 'Unknown'

def get_ip_info(ip: str) -> dict:
    try:
        response = requests.get(f"https://ipinfo.io/{ip}/json", timeout=10)
        return response.json()
    except Exception as e:
        return {"error": str(e)}

def check_single(domain: str = None, ip: str = None, port: int = 443):
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
            # Assume ip is valid
            pass
        
        ssl_info = get_ssl_info(domain or ip, ip, port)
        server = get_server_header(domain or ip, ip, port)
        ip_info = get_ip_info(ip)
        
        # Status
        ssl_status = "success" if ssl_info.get('subject') is not None else "error"
        server_status = "success" if server != 'Unknown' else "error"
        ip_status = "success" if ip_info and 'error' not in ip_info else "error"
        
        # Standardize fields
        if 'error' in ip_info:
            ip_info = None
        
        # Recommendations
        recommendations = []
        if ssl_status == "error":
            recommendations.append("Install SSL certificate")
        if ssl_info and ssl_info.get('daysUntilExpiration') is not None and ssl_info.get('daysUntilExpiration') < 30:
            recommendations.append("Renew certificate soon")
        if ssl_info and ssl_info.get('tlsVersion') in ['TLSv1', 'TLSv1.1']:
            recommendations.append("Upgrade to TLS 1.2 or higher")
        
        return {
            "status": "success",
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
                "ipStatus": ip_status
            }
        }
    except Exception as e:
        return {
            "status": "error",
            "timestamp": checked_at,
            "error": str(e)
        }

@app.get("/check")
def check_ssl(domain: str = None, ip: str = None, port: int = 443):
    return check_single(domain, ip, port)

@app.post("/batch_check")
def batch_check(request: BatchRequest):
    results = []
    for domain in request.domains:
        result = check_single(domain=domain, port=request.port)
        results.append(result)
    for ip in request.ips:
        result = check_single(ip=ip, port=request.port)
        results.append(result)
    return {
        "status": "success",
        "timestamp": datetime.now().isoformat(),
        "results": results
    }

