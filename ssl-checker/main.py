from fastapi import FastAPI, HTTPException
import requests
import ssl
import socket
import dns.resolver
from datetime import datetime

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
            raise HTTPException(status_code=400, detail="Invalid domain")

def get_ssl_info(domain: str, ip: str) -> dict:
    try:
        context = ssl.create_default_context()
        with socket.create_connection((ip, 443)) as sock:
            with context.wrap_socket(sock, server_hostname=domain) as ssock:
                cert = ssock.getpeercert()
                return {
                    "subject": dict(x[0] for x in cert.get('subject', [])),
                    "issuer": dict(x[0] for x in cert.get('issuer', [])),
                    "version": cert.get('version'),
                    "serialNumber": cert.get('serialNumber'),
                    "notBefore": cert.get('notBefore'),
                    "notAfter": cert.get('notAfter'),
                    "signatureAlgorithm": cert.get('signatureAlgorithm'),
                }
    except Exception as e:
        return {"error": str(e)}

def get_server_header(domain: str, ip: str) -> str:
    try:
        url = f"https://{domain}"
        response = requests.get(url, timeout=10)
        return response.headers.get('Server', 'Unknown')
    except Exception:
        try:
            url = f"http://{ip}"
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
    if not domain and not ip:
        raise HTTPException(status_code=400, detail="Provide either domain or ip")
    
    if domain:
        ip = resolve_domain_to_ip(domain)
    else:
        # Assume ip is valid, but could add validation
        pass
    
    ssl_info = get_ssl_info(domain or ip, ip)
    server = get_server_header(domain or ip, ip)
    ip_info = get_ip_info(ip)
    
    return {
        "domain": domain,
        "ip": ip,
        "ssl": ssl_info,
        "server": server,
        "ip_info": ip_info
    }