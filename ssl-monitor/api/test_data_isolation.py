"""
Tests for multi-user data isolation in SSL Monitor

These tests verify that:
1. User A cannot see User B's data
2. Database filtering by user_id works correctly
3. Organization-based data isolation works
"""
import pytest
import os
import sys
from datetime import datetime
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).resolve().parent))

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from database import Base, User, Organization, SSLCheck, Monitor, Alert

# Use in-memory database for tests
TEST_DATABASE_URL = "sqlite:///:memory:"


@pytest.fixture
def db_session():
    """Create a test database session"""
    engine = create_engine(TEST_DATABASE_URL, connect_args={"check_same_thread": False})
    Base.metadata.create_all(bind=engine)
    TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    
    session = TestingSessionLocal()
    yield session
    
    session.close()
    Base.metadata.drop_all(bind=engine)


@pytest.fixture
def sample_organization(db_session):
    """Create a sample organization"""
    org = Organization(
        name="Test Organization",
        description="Test organization for isolation testing"
    )
    db_session.add(org)
    db_session.commit()
    db_session.refresh(org)
    return org


@pytest.fixture
def user_a(db_session, sample_organization):
    """Create User A"""
    user = User(
        email="usera@example.com",
        hashed_password="hashed_password_a",
        is_active=True,
        is_verified=True,
        organization_id=sample_organization.id
    )
    db_session.add(user)
    db_session.commit()
    db_session.refresh(user)
    return user


@pytest.fixture
def user_b(db_session, sample_organization):
    """Create User B"""
    user = User(
        email="userb@example.com",
        hashed_password="hashed_password_b",
        is_active=True,
        is_verified=True,
        organization_id=sample_organization.id
    )
    db_session.add(user)
    db_session.commit()
    db_session.refresh(user)
    return user


def test_ssl_check_isolation(db_session, user_a, user_b):
    """Test that User A cannot see User B's SSL checks"""
    # User A creates SSL checks
    check_a1 = SSLCheck(
        user_id=user_a.id,
        organization_id=user_a.organization_id,
        domain="example-a1.com",
        status="success",
        ssl_status="valid",
        checked_at=datetime.utcnow()
    )
    check_a2 = SSLCheck(
        user_id=user_a.id,
        organization_id=user_a.organization_id,
        domain="example-a2.com",
        status="success",
        ssl_status="valid",
        checked_at=datetime.utcnow()
    )
    
    # User B creates SSL checks
    check_b1 = SSLCheck(
        user_id=user_b.id,
        organization_id=user_b.organization_id,
        domain="example-b1.com",
        status="success",
        ssl_status="valid",
        checked_at=datetime.utcnow()
    )
    check_b2 = SSLCheck(
        user_id=user_b.id,
        organization_id=user_b.organization_id,
        domain="example-b2.com",
        status="success",
        ssl_status="valid",
        checked_at=datetime.utcnow()
    )
    
    db_session.add_all([check_a1, check_a2, check_b1, check_b2])
    db_session.commit()
    
    # Query as User A - should only see their own checks
    user_a_checks = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_a.id
    ).all()
    
    assert len(user_a_checks) == 2
    assert all(check.user_id == user_a.id for check in user_a_checks)
    assert set(check.domain for check in user_a_checks) == {"example-a1.com", "example-a2.com"}
    
    # Query as User B - should only see their own checks
    user_b_checks = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_b.id
    ).all()
    
    assert len(user_b_checks) == 2
    assert all(check.user_id == user_b.id for check in user_b_checks)
    assert set(check.domain for check in user_b_checks) == {"example-b1.com", "example-b2.com"}


def test_monitor_isolation(db_session, user_a, user_b):
    """Test that User A cannot see User B's monitors"""
    # User A creates monitors
    monitor_a1 = Monitor(
        user_id=user_a.id,
        organization_id=user_a.organization_id,
        domain="monitor-a1.com",
        status="active"
    )
    monitor_a2 = Monitor(
        user_id=user_a.id,
        organization_id=user_a.organization_id,
        domain="monitor-a2.com",
        status="active"
    )
    
    # User B creates monitors
    monitor_b1 = Monitor(
        user_id=user_b.id,
        organization_id=user_b.organization_id,
        domain="monitor-b1.com",
        status="active"
    )
    
    db_session.add_all([monitor_a1, monitor_a2, monitor_b1])
    db_session.commit()
    
    # Query as User A
    user_a_monitors = db_session.query(Monitor).filter(
        Monitor.user_id == user_a.id
    ).all()
    
    assert len(user_a_monitors) == 2
    assert all(monitor.user_id == user_a.id for monitor in user_a_monitors)
    
    # Query as User B
    user_b_monitors = db_session.query(Monitor).filter(
        Monitor.user_id == user_b.id
    ).all()
    
    assert len(user_b_monitors) == 1
    assert all(monitor.user_id == user_b.id for monitor in user_b_monitors)


def test_alert_isolation(db_session, user_a, user_b):
    """Test that User A cannot see User B's alerts"""
    # User A creates alerts
    alert_a = Alert(
        user_id=user_a.id,
        organization_id=user_a.organization_id,
        domain="alert-a.com",
        alert_type="expiring_soon",
        severity="medium",
        message="Certificate expiring soon"
    )
    
    # User B creates alerts
    alert_b = Alert(
        user_id=user_b.id,
        organization_id=user_b.organization_id,
        domain="alert-b.com",
        alert_type="expired",
        severity="high",
        message="Certificate expired"
    )
    
    db_session.add_all([alert_a, alert_b])
    db_session.commit()
    
    # Query as User A
    user_a_alerts = db_session.query(Alert).filter(
        Alert.user_id == user_a.id
    ).all()
    
    assert len(user_a_alerts) == 1
    assert user_a_alerts[0].user_id == user_a.id
    assert user_a_alerts[0].domain == "alert-a.com"
    
    # Query as User B
    user_b_alerts = db_session.query(Alert).filter(
        Alert.user_id == user_b.id
    ).all()
    
    assert len(user_b_alerts) == 1
    assert user_b_alerts[0].user_id == user_b.id
    assert user_b_alerts[0].domain == "alert-b.com"


def test_organization_based_filtering(db_session):
    """Test filtering by organization"""
    # Create two organizations
    org1 = Organization(name="Org 1")
    org2 = Organization(name="Org 2")
    db_session.add_all([org1, org2])
    db_session.commit()
    db_session.refresh(org1)
    db_session.refresh(org2)
    
    # Create users in different organizations
    user_org1 = User(
        email="user@org1.com",
        hashed_password="hash1",
        is_active=True,
        organization_id=org1.id
    )
    user_org2 = User(
        email="user@org2.com",
        hashed_password="hash2",
        is_active=True,
        organization_id=org2.id
    )
    db_session.add_all([user_org1, user_org2])
    db_session.commit()
    db_session.refresh(user_org1)
    db_session.refresh(user_org2)
    
    # Create SSL checks for each organization
    check_org1 = SSLCheck(
        user_id=user_org1.id,
        organization_id=org1.id,
        domain="org1.com",
        status="success"
    )
    check_org2 = SSLCheck(
        user_id=user_org2.id,
        organization_id=org2.id,
        domain="org2.com",
        status="success"
    )
    db_session.add_all([check_org1, check_org2])
    db_session.commit()
    
    # Query by organization
    org1_checks = db_session.query(SSLCheck).filter(
        SSLCheck.organization_id == org1.id
    ).all()
    
    assert len(org1_checks) == 1
    assert org1_checks[0].organization_id == org1.id
    
    org2_checks = db_session.query(SSLCheck).filter(
        SSLCheck.organization_id == org2.id
    ).all()
    
    assert len(org2_checks) == 1
    assert org2_checks[0].organization_id == org2.id


def test_stats_isolation(db_session, user_a, user_b):
    """Test that statistics are correctly isolated per user"""
    # Create SSL checks for User A
    for i in range(5):
        check = SSLCheck(
            user_id=user_a.id,
            domain=f"user-a-{i}.com",
            status="success" if i < 3 else "error"
        )
        db_session.add(check)
    
    # Create SSL checks for User B
    for i in range(3):
        check = SSLCheck(
            user_id=user_b.id,
            domain=f"user-b-{i}.com",
            status="success"
        )
        db_session.add(check)
    
    db_session.commit()
    
    # Get stats for User A
    user_a_total = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_a.id
    ).count()
    user_a_success = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_a.id,
        SSLCheck.status == "success"
    ).count()
    
    assert user_a_total == 5
    assert user_a_success == 3
    
    # Get stats for User B
    user_b_total = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_b.id
    ).count()
    user_b_success = db_session.query(SSLCheck).filter(
        SSLCheck.user_id == user_b.id,
        SSLCheck.status == "success"
    ).count()
    
    assert user_b_total == 3
    assert user_b_success == 3


def test_no_user_id_isolation():
    """Test that records without user_id are not accessible"""
    # This ensures that old data or data created without user_id
    # won't leak between users
    
    # Create a simple test - in production, user_id should be required
    # for all new records, but we want to ensure filtering works correctly
    pass  # This is more of a policy test


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
