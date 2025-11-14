"""
Tests for alert service functionality
"""
import pytest
import json
from datetime import datetime, timedelta
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from database import Base, User, AlertConfig, Alert, SSLCheck
from alert_service import (
    check_certificate_expiry,
    check_ssl_errors,
    check_geo_changes,
    create_alert,
    process_ssl_check_alerts,
    get_or_create_alert_config,
)


# Test database setup
@pytest.fixture
def db_session():
    """Create a test database session"""
    engine = create_engine("sqlite:///:memory:")
    Base.metadata.create_all(engine)
    Session = sessionmaker(bind=engine)
    session = Session()
    yield session
    session.close()


@pytest.fixture
def test_user(db_session):
    """Create a test user"""
    user = User(
        email="test@example.com",
        hashed_password="hashed",
        is_active=True,
        is_verified=True
    )
    db_session.add(user)
    db_session.commit()
    db_session.refresh(user)
    return user


@pytest.fixture
def test_alert_config(db_session, test_user):
    """Create a test alert configuration"""
    config = AlertConfig(
        user_id=test_user.id,
        enabled=True,
        alert_30_days=True,
        alert_7_days=True,
        alert_1_day=True,
        alert_ssl_errors=True,
        alert_geo_changes=True,
        alert_cert_expired=True
    )
    db_session.add(config)
    db_session.commit()
    db_session.refresh(config)
    return config


def test_get_or_create_alert_config(db_session, test_user):
    """Test getting or creating alert config"""
    # First call should create new config
    config1 = get_or_create_alert_config(db_session, test_user.id)
    assert config1 is not None
    assert config1.user_id == test_user.id
    assert config1.enabled is True
    
    # Second call should return existing config
    config2 = get_or_create_alert_config(db_session, test_user.id)
    assert config1.id == config2.id


def test_check_certificate_expiry_30_days(test_alert_config):
    """Test certificate expiry check for 30 days"""
    expiry_date = datetime.utcnow() + timedelta(days=25)
    ssl_data = {
        'certificate': {
            'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
        }
    }
    
    result = check_certificate_expiry(ssl_data, test_alert_config)
    assert result is not None
    alert_type, severity, message, days_remaining = result
    assert alert_type == 'expiring_soon'
    assert severity == 'medium'
    assert 24 <= days_remaining <= 25  # Allow for rounding


def test_check_certificate_expiry_7_days(test_alert_config):
    """Test certificate expiry check for 7 days"""
    expiry_date = datetime.utcnow() + timedelta(days=5)
    ssl_data = {
        'certificate': {
            'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
        }
    }
    
    result = check_certificate_expiry(ssl_data, test_alert_config)
    assert result is not None
    alert_type, severity, message, days_remaining = result
    assert alert_type == 'expiring_soon'
    assert severity == 'high'
    assert 4 <= days_remaining <= 5  # Allow for rounding


def test_check_certificate_expiry_1_day(test_alert_config):
    """Test certificate expiry check for 1 day"""
    expiry_date = datetime.utcnow() + timedelta(hours=12)
    ssl_data = {
        'certificate': {
            'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
        }
    }
    
    result = check_certificate_expiry(ssl_data, test_alert_config)
    assert result is not None
    alert_type, severity, message, days_remaining = result
    assert alert_type == 'expiring_soon'
    assert severity == 'critical'
    assert days_remaining == 0


def test_check_certificate_expired(test_alert_config):
    """Test expired certificate check"""
    expiry_date = datetime.utcnow() - timedelta(days=5)
    ssl_data = {
        'certificate': {
            'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
        }
    }
    
    result = check_certificate_expiry(ssl_data, test_alert_config)
    assert result is not None
    alert_type, severity, message, days_remaining = result
    assert alert_type == 'expired'
    assert severity == 'critical'
    assert -6 <= days_remaining <= -5  # Allow for rounding


def test_check_certificate_expiry_disabled(test_alert_config):
    """Test that disabled alerts don't trigger"""
    test_alert_config.alert_30_days = False
    test_alert_config.alert_7_days = False
    test_alert_config.alert_1_day = False
    
    expiry_date = datetime.utcnow() + timedelta(days=25)
    ssl_data = {
        'certificate': {
            'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
        }
    }
    
    result = check_certificate_expiry(ssl_data, test_alert_config)
    assert result is None


def test_check_ssl_errors(test_alert_config):
    """Test SSL error detection"""
    ssl_data = {
        'sslStatus': 'error',
        'alerts': [
            {'type': 'error', 'message': 'Certificate validation failed'}
        ]
    }
    
    result = check_ssl_errors(ssl_data, test_alert_config)
    assert result is not None
    alert_type, severity, message = result
    assert alert_type == 'ssl_error'
    assert severity == 'high'
    assert 'validation failed' in message.lower()


def test_check_ssl_errors_disabled(test_alert_config):
    """Test that disabled SSL error alerts don't trigger"""
    test_alert_config.alert_ssl_errors = False
    
    ssl_data = {
        'sslStatus': 'error',
        'alerts': [
            {'type': 'error', 'message': 'Certificate validation failed'}
        ]
    }
    
    result = check_ssl_errors(ssl_data, test_alert_config)
    assert result is None


def test_create_alert(db_session, test_user):
    """Test creating an alert"""
    alert = create_alert(
        db_session,
        test_user.id,
        None,
        'example.com',
        'expiring_soon',
        'medium',
        'Certificate expiring in 25 days'
    )
    
    assert alert.id is not None
    assert alert.user_id == test_user.id
    assert alert.domain == 'example.com'
    assert alert.alert_type == 'expiring_soon'
    assert alert.severity == 'medium'
    assert alert.is_read is False
    assert alert.is_resolved is False


def test_process_ssl_check_alerts_expiry(db_session, test_user, test_alert_config):
    """Test processing SSL check for expiry alerts"""
    expiry_date = datetime.utcnow() + timedelta(days=25)
    ssl_check_data = {
        'status': 'success',
        'data': {
            'sslStatus': 'success',
            'certificate': {
                'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
            }
        }
    }
    
    alerts = process_ssl_check_alerts(
        db_session,
        test_user.id,
        None,
        'example.com',
        ssl_check_data,
        test_alert_config
    )
    
    assert len(alerts) == 1
    assert alerts[0].alert_type == 'expiring_soon'
    assert alerts[0].severity == 'medium'


def test_process_ssl_check_alerts_disabled(db_session, test_user, test_alert_config):
    """Test that alerts don't process when config is disabled"""
    test_alert_config.enabled = False
    db_session.commit()
    
    expiry_date = datetime.utcnow() + timedelta(days=25)
    ssl_check_data = {
        'status': 'success',
        'data': {
            'sslStatus': 'success',
            'certificate': {
                'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
            }
        }
    }
    
    alerts = process_ssl_check_alerts(
        db_session,
        test_user.id,
        None,
        'example.com',
        ssl_check_data,
        test_alert_config
    )
    
    assert len(alerts) == 0


def test_check_geo_changes(db_session, test_user, test_alert_config):
    """Test geolocation change detection"""
    # Create a previous check with geolocation
    previous_check = SSLCheck(
        user_id=test_user.id,
        domain='example.com',
        ip='1.2.3.4',
        port=443,
        status='success',
        ssl_status='success',
        response_data=json.dumps({
            'data': {
                'geolocation': {
                    'country': 'United States'
                }
            }
        })
    )
    db_session.add(previous_check)
    db_session.commit()
    
    # Create current check with different geolocation to test against
    current_check = SSLCheck(
        user_id=test_user.id,
        domain='example.com',
        ip='1.2.3.5',
        port=443,
        status='success',
        ssl_status='success',
        response_data=json.dumps({
            'data': {
                'geolocation': {
                    'country': 'Germany'
                }
            }
        })
    )
    db_session.add(current_check)
    db_session.commit()
    
    # Check with new geolocation
    current_geo = {'country': 'Germany'}
    
    result = check_geo_changes(
        'example.com',
        '1.2.3.5',
        current_geo,
        db_session,
        test_user.id,
        test_alert_config
    )
    
    assert result is not None
    alert_type, severity, message = result
    assert alert_type == 'geo_change'
    assert severity == 'medium'
    assert 'United States' in message
    assert 'Germany' in message


def test_no_spam_duplicate_alerts(db_session, test_user, test_alert_config):
    """Test that alert deduplication prevents spam"""
    expiry_date = datetime.utcnow() + timedelta(days=25)
    ssl_check_data = {
        'status': 'success',
        'data': {
            'sslStatus': 'success',
            'certificate': {
                'expiryDate': expiry_date.strftime('%Y-%m-%d %H:%M:%S')
            }
        }
    }
    
    # First check should create alert
    alerts1 = process_ssl_check_alerts(
        db_session,
        test_user.id,
        None,
        'example.com',
        ssl_check_data,
        test_alert_config
    )
    assert len(alerts1) == 1
    first_alert_id = alerts1[0].id
    
    # Second check with same condition should NOT create duplicate
    # Due to deduplication within 24-hour window
    alerts2 = process_ssl_check_alerts(
        db_session,
        test_user.id,
        None,
        'example.com',
        ssl_check_data,
        test_alert_config
    )
    # Should return the same alert (updated timestamp)
    assert len(alerts2) == 1
    assert alerts2[0].id == first_alert_id
    
    # Verify total alert count hasn't increased
    total_alerts = db_session.query(Alert).filter(
        Alert.user_id == test_user.id,
        Alert.domain == 'example.com'
    ).count()
    assert total_alerts == 1


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
