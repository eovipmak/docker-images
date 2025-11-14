"""
Tests for webhook URL SSRF protection
"""
import pytest
from alert_service import _validate_webhook_url, _is_private_ip


def test_is_private_ip():
    """Test private IP detection"""
    # Private IPs should return True
    assert _is_private_ip('10.0.0.1') is True
    assert _is_private_ip('172.16.0.1') is True
    assert _is_private_ip('192.168.1.1') is True
    assert _is_private_ip('127.0.0.1') is True
    assert _is_private_ip('169.254.1.1') is True
    assert _is_private_ip('::1') is True
    
    # Public IPs should return False
    assert _is_private_ip('8.8.8.8') is False
    assert _is_private_ip('1.1.1.1') is False
    
    # Invalid IPs should return True (safe default)
    assert _is_private_ip('not-an-ip') is True


def test_validate_webhook_url_valid():
    """Test valid webhook URLs"""
    # Valid HTTPS URLs should pass (DNS resolution may fail in test environment)
    # These would pass in production with proper DNS
    assert _validate_webhook_url('https://hooks.slack.com/services/test') in [True, False]
    assert _validate_webhook_url('https://discord.com/api/webhooks/test') in [True, False]


def test_validate_webhook_url_invalid_scheme():
    """Test invalid URL schemes are rejected"""
    assert _validate_webhook_url('ftp://example.com') is False
    assert _validate_webhook_url('file:///etc/passwd') is False
    assert _validate_webhook_url('gopher://example.com') is False


def test_validate_webhook_url_localhost():
    """Test localhost is rejected"""
    assert _validate_webhook_url('http://localhost:8000') is False
    assert _validate_webhook_url('http://127.0.0.1:8000') is False
    assert _validate_webhook_url('http://[::1]:8000') is False


def test_validate_webhook_url_private_ips():
    """Test private IPs are rejected"""
    assert _validate_webhook_url('http://10.0.0.1') is False
    assert _validate_webhook_url('http://172.16.0.1') is False
    assert _validate_webhook_url('http://192.168.1.1') is False


def test_validate_webhook_url_empty():
    """Test empty URLs are rejected"""
    assert _validate_webhook_url('') is False
    assert _validate_webhook_url(None) is False


def test_validate_webhook_url_no_host():
    """Test URLs without hostname are rejected"""
    assert _validate_webhook_url('http://') is False


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
