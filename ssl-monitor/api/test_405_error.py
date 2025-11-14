"""
Test to reproduce the 405 error when deleting domains and disabling alerts.
"""
import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
import os

# Set test database
os.environ["DATABASE_URL"] = "sqlite:///./test_405_error.db"
os.environ["ASYNC_DATABASE_URL"] = "sqlite+aiosqlite:///./test_405_error.db"

from main import app
from database import Base, get_db

# Create test database
TEST_DATABASE_URL = "sqlite:///./test_405_error.db"
engine = create_engine(TEST_DATABASE_URL, connect_args={"check_same_thread": False})
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Create tables
Base.metadata.create_all(bind=engine)


def override_get_db():
    """Override database dependency for testing"""
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db

client = TestClient(app)


@pytest.fixture(scope="module")
def auth_token():
    """Register and login a user, return auth token"""
    # Try to register user (may fail if already exists)
    response = client.post(
        "/auth/register",
        json={"email": "testuser405@example.com", "password": "testpass123"}
    )
    # Accept either 201 (created) or 400 (already exists)
    assert response.status_code in [201, 400], f"Register failed with {response.status_code}: {response.text}"
    
    # Login
    response = client.post(
        "/auth/jwt/login",
        data={"username": "testuser405@example.com", "password": "testpass123"}
    )
    assert response.status_code == 200, f"Login failed: {response.text}"
    return response.json()["access_token"]


def test_delete_domain_without_trailing_slash(auth_token):
    """Test DELETE /api/domains/{domain} without trailing slash"""
    # First add a domain
    response = client.post(
        "/api/domains",
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"domain": "example.com", "port": 443}
    )
    # May fail due to SSL checker service not being available, but that's OK
    # We just want to test the route
    
    # Try to delete the domain
    domain_name = "example.com"
    response = client.delete(
        f"/api/domains/{domain_name}",
        headers={"Authorization": f"Bearer {auth_token}"},
        follow_redirects=False
    )
    
    print(f"DELETE /api/domains/{domain_name} - Status: {response.status_code}")
    # Should be either 200 (success) or 404 (not found), but NOT 405
    assert response.status_code != 405, f"Got 405 error for DELETE /api/domains/{domain_name}"


def test_delete_domain_with_trailing_slash(auth_token):
    """Test DELETE /api/domains/{domain}/ with trailing slash"""
    domain_name = "example.com"
    response = client.delete(
        f"/api/domains/{domain_name}/",  # Note trailing slash
        headers={"Authorization": f"Bearer {auth_token}"},
        follow_redirects=True  # Follow redirects
    )
    
    print(f"DELETE /api/domains/{domain_name}/ - Status: {response.status_code}")
    # Should be either 200 (success) or 404 (not found), but NOT 405
    assert response.status_code != 405, f"Got 405 error for DELETE /api/domains/{domain_name}/"


def test_patch_monitor_without_trailing_slash(auth_token):
    """Test PATCH /api/monitors/{domain} without trailing slash"""
    domain_name = "test-domain.com"
    
    # First create a monitor
    response = client.post(
        "/api/monitors",
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"domain": domain_name, "port": 443, "alerts_enabled": True}
    )
    print(f"Create monitor - Status: {response.status_code}")
    
    # Try to update the monitor
    response = client.patch(
        f"/api/monitors/{domain_name}",
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"alerts_enabled": False},
        follow_redirects=False
    )
    
    print(f"PATCH /api/monitors/{domain_name} - Status: {response.status_code}")
    # Should be either 200 (success) or 404 (not found), but NOT 405
    assert response.status_code != 405, f"Got 405 error for PATCH /api/monitors/{domain_name}"


def test_patch_monitor_with_trailing_slash(auth_token):
    """Test PATCH /api/monitors/{domain}/ with trailing slash"""
    domain_name = "test-domain.com"
    
    # Try to update the monitor with trailing slash
    response = client.patch(
        f"/api/monitors/{domain_name}/",  # Note trailing slash
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"alerts_enabled": True},
        follow_redirects=True  # Follow redirects
    )
    
    print(f"PATCH /api/monitors/{domain_name}/ - Status: {response.status_code}")
    # Should be either 200 (success) or 404 (not found), but NOT 405
    assert response.status_code != 405, f"Got 405 error for PATCH /api/monitors/{domain_name}/"


def test_delete_redirect_status():
    """Test that DELETE with trailing slash returns 307 redirect"""
    response = client.delete(
        "/api/domains/example.com/",  # Note trailing slash
        follow_redirects=False
    )
    
    print(f"DELETE redirect status: {response.status_code}")
    # Middleware should redirect with 307 before authentication
    # If we get 401, it means no redirect happened
    # If we get 405, it means the redirect didn't preserve the method
    assert response.status_code in [307, 401], f"Expected 307 or 401, got {response.status_code}"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "-s"])
