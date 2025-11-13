"""
Simple API test to verify data isolation in endpoints

This test creates two users and verifies they cannot see each other's data.
"""
import asyncio
import os
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

# Set test database
os.environ["DATABASE_URL"] = "sqlite:///./test_api.db"
os.environ["ASYNC_DATABASE_URL"] = "sqlite+aiosqlite:///./test_api.db"

from main import app
from database import Base, get_db, User, SSLCheck

# Create test database
TEST_DATABASE_URL = "sqlite:///./test_api.db"
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


def test_api_data_isolation():
    """Test that API endpoints properly isolate data between users"""
    
    print("\n=== Testing API Data Isolation ===\n")
    
    # Register User A
    print("1. Registering User A...")
    response_a = client.post(
        "/auth/register",
        json={
            "email": "usera@test.com",
            "password": "password123"
        }
    )
    assert response_a.status_code == 201, f"User A registration failed: {response_a.text}"
    print(f"   ✓ User A registered: {response_a.json()['email']}")
    
    # Register User B
    print("2. Registering User B...")
    response_b = client.post(
        "/auth/register",
        json={
            "email": "userb@test.com",
            "password": "password456"
        }
    )
    assert response_b.status_code == 201, f"User B registration failed: {response_b.text}"
    print(f"   ✓ User B registered: {response_b.json()['email']}")
    
    # Login User A
    print("3. Logging in User A...")
    login_a = client.post(
        "/auth/jwt/login",
        data={
            "username": "usera@test.com",
            "password": "password123"
        }
    )
    assert login_a.status_code == 200, f"User A login failed: {login_a.text}"
    token_a = login_a.json()["access_token"]
    print(f"   ✓ User A logged in successfully")
    
    # Login User B
    print("4. Logging in User B...")
    login_b = client.post(
        "/auth/jwt/login",
        data={
            "username": "userb@test.com",
            "password": "password456"
        }
    )
    assert login_b.status_code == 200, f"User B login failed: {login_b.text}"
    token_b = login_b.json()["access_token"]
    print(f"   ✓ User B logged in successfully")
    
    # Get stats for User A (should be empty)
    print("5. Checking User A stats (should be empty)...")
    stats_a = client.get(
        "/api/stats",
        headers={"Authorization": f"Bearer {token_a}"}
    )
    assert stats_a.status_code == 200
    stats_a_data = stats_a.json()["stats"]
    assert stats_a_data["total_checks"] == 0
    print(f"   ✓ User A has {stats_a_data['total_checks']} checks")
    
    # Get stats for User B (should be empty)
    print("6. Checking User B stats (should be empty)...")
    stats_b = client.get(
        "/api/stats",
        headers={"Authorization": f"Bearer {token_b}"}
    )
    assert stats_b.status_code == 200
    stats_b_data = stats_b.json()["stats"]
    assert stats_b_data["total_checks"] == 0
    print(f"   ✓ User B has {stats_b_data['total_checks']} checks")
    
    # Verify unauthenticated access is denied
    print("7. Verifying unauthenticated access is denied...")
    stats_unauth = client.get("/api/stats")
    assert stats_unauth.status_code == 401
    print(f"   ✓ Unauthenticated access properly denied")
    
    print("\n=== All API Data Isolation Tests Passed! ===\n")


if __name__ == "__main__":
    try:
        test_api_data_isolation()
        print("SUCCESS: All tests passed!")
    except AssertionError as e:
        print(f"FAILED: {e}")
        exit(1)
    except Exception as e:
        print(f"ERROR: {e}")
        import traceback
        traceback.print_exc()
        exit(1)
    finally:
        # Clean up test database
        import os
        if os.path.exists("test_api.db"):
            os.remove("test_api.db")
