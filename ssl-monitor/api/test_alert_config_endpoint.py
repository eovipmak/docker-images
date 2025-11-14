"""
Test for alert-config endpoint to verify GET and POST methods work correctly
"""
import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from main import app
from database import Base, get_db, User, Organization
from auth import get_password_hash


# Create test database
SQLALCHEMY_DATABASE_URL = "sqlite:///./test_alert_endpoint.db"
engine = create_engine(SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False})
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db


@pytest.fixture(scope="function")
def client():
    """Create a test client with a fresh database"""
    # Create tables
    Base.metadata.create_all(bind=engine)
    
    # Create test user
    db = TestingSessionLocal()
    try:
        # Create organization
        org = Organization(name="Test Org")
        db.add(org)
        db.flush()
        
        # Create user
        user = User(
            email="test@example.com",
            hashed_password=get_password_hash("testpassword"),
            is_active=True,
            is_verified=True,
            organization_id=org.id
        )
        db.add(user)
        db.commit()
    finally:
        db.close()
    
    # Create test client
    test_client = TestClient(app)
    
    # Login to get token
    response = test_client.post(
        "/auth/jwt/login",
        data={"username": "test@example.com", "password": "testpassword"}
    )
    assert response.status_code == 200
    token = response.json()["access_token"]
    
    # Set authorization header for all requests
    test_client.headers["Authorization"] = f"Bearer {token}"
    
    yield test_client
    
    # Cleanup
    Base.metadata.drop_all(bind=engine)


def test_get_alert_config(client):
    """Test GET /api/alert-config endpoint"""
    response = client.get("/api/alert-config")
    assert response.status_code == 200
    data = response.json()
    assert "id" in data
    assert "enabled" in data
    assert "webhook_url" in data


def test_post_alert_config(client):
    """Test POST /api/alert-config endpoint"""
    # First get the config
    response = client.get("/api/alert-config")
    assert response.status_code == 200
    
    # Update config with webhook URL
    update_data = {
        "enabled": True,
        "webhook_url": "https://hooks.slack.com/services/test",
        "alert_30_days": True,
        "alert_7_days": True,
        "alert_1_day": False
    }
    
    response = client.post("/api/alert-config", json=update_data)
    assert response.status_code == 200
    data = response.json()
    assert data["enabled"] == True
    assert data["webhook_url"] == "https://hooks.slack.com/services/test"
    assert data["alert_1_day"] == False


def test_options_alert_config(client):
    """Test OPTIONS /api/alert-config endpoint for CORS preflight"""
    response = client.options("/api/alert-config")
    # FastAPI/Starlette should handle OPTIONS automatically
    # It should not return 405
    assert response.status_code != 405


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
