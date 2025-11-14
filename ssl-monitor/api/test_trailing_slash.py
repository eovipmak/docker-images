"""
Test for trailing slash handling in API endpoints.
This ensures that requests with trailing slashes are properly redirected to prevent 405 errors.
"""
import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
import os

# Set test database
os.environ["DATABASE_URL"] = "sqlite:///./test_trailing_slash.db"
os.environ["ASYNC_DATABASE_URL"] = "sqlite+aiosqlite:///./test_trailing_slash.db"

from main import app
from database import Base, get_db

# Create test database
TEST_DATABASE_URL = "sqlite:///./test_trailing_slash.db"
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
    # Register user
    response = client.post(
        "/auth/register",
        json={"email": "testuser@example.com", "password": "testpass123"}
    )
    assert response.status_code == 201
    
    # Login
    response = client.post(
        "/auth/jwt/login",
        data={"username": "testuser@example.com", "password": "testpass123"}
    )
    assert response.status_code == 200
    return response.json()["access_token"]


def test_trailing_slash_redirect_post(auth_token):
    """Test that POST requests with trailing slashes are redirected"""
    # Make POST request with trailing slash
    response = client.post(
        "/api/alert-config/",  # Note trailing slash
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"enabled": True, "webhook_url": "https://hooks.slack.com/test"},
        follow_redirects=True
    )
    
    # Should succeed after redirect
    assert response.status_code == 200
    assert response.json()["webhook_url"] == "https://hooks.slack.com/test"


def test_trailing_slash_redirect_get(auth_token):
    """Test that GET requests with trailing slashes are redirected"""
    # Make GET request with trailing slash
    response = client.get(
        "/api/alert-config/",  # Note trailing slash
        headers={"Authorization": f"Bearer {auth_token}"},
        follow_redirects=True
    )
    
    # Should succeed after redirect
    assert response.status_code == 200
    assert "id" in response.json()
    assert "enabled" in response.json()


def test_no_trailing_slash_works(auth_token):
    """Test that requests without trailing slashes still work normally"""
    # Make POST request without trailing slash
    response = client.post(
        "/api/alert-config",  # No trailing slash
        headers={"Authorization": f"Bearer {auth_token}"},
        json={"enabled": False},
        follow_redirects=False
    )
    
    # Should succeed without redirect
    assert response.status_code == 200
    assert response.json()["enabled"] == False


def test_trailing_slash_returns_307():
    """Test that trailing slash redirect returns 307 status code"""
    # Make request with trailing slash, don't follow redirects
    response = client.get(
        "/api/alert-config/",  # Note trailing slash
        follow_redirects=False
    )
    
    # Should get 307 Temporary Redirect (preserves HTTP method)
    # Even though it will be 401 unauthorized, the redirect should happen first
    # Actually, middleware runs before authentication, so we should get 307
    assert response.status_code == 307
    assert response.headers["location"] == "http://testserver/api/alert-config"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
