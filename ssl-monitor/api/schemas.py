"""
User schemas for FastAPI Users
"""
from typing import Optional
from datetime import datetime
from fastapi_users import schemas
from pydantic import BaseModel, EmailStr, Field


class UserRead(schemas.BaseUser[int]):
    """Schema for reading user data"""
    id: int
    email: EmailStr
    is_active: bool = True
    is_superuser: bool = False
    is_verified: bool = False

    class Config:
        from_attributes = True


class UserCreate(schemas.BaseUserCreate):
    """Schema for creating a new user"""
    email: EmailStr
    password: str
    is_active: Optional[bool] = True
    is_superuser: Optional[bool] = False
    is_verified: Optional[bool] = False


class UserUpdate(schemas.BaseUserUpdate):
    """Schema for updating user data"""
    password: Optional[str] = None
    email: Optional[EmailStr] = None
    is_active: Optional[bool] = None
    is_superuser: Optional[bool] = None
    is_verified: Optional[bool] = None


class AlertConfigCreate(BaseModel):
    """Schema for creating alert configuration"""
    enabled: bool = True
    webhook_url: Optional[str] = None
    alert_30_days: bool = True
    alert_7_days: bool = True
    alert_1_day: bool = True
    alert_ssl_errors: bool = True
    alert_geo_changes: bool = False
    alert_cert_expired: bool = True
    email_notifications: bool = False
    email_address: Optional[EmailStr] = None

    class Config:
        from_attributes = True


class AlertConfigUpdate(BaseModel):
    """Schema for updating alert configuration"""
    enabled: Optional[bool] = None
    webhook_url: Optional[str] = None
    alert_30_days: Optional[bool] = None
    alert_7_days: Optional[bool] = None
    alert_1_day: Optional[bool] = None
    alert_ssl_errors: Optional[bool] = None
    alert_geo_changes: Optional[bool] = None
    alert_cert_expired: Optional[bool] = None
    email_notifications: Optional[bool] = None
    email_address: Optional[EmailStr] = None

    class Config:
        from_attributes = True


class AlertConfigRead(BaseModel):
    """Schema for reading alert configuration"""
    id: int
    user_id: int
    organization_id: Optional[int] = None
    enabled: bool
    webhook_url: Optional[str] = None
    alert_30_days: bool
    alert_7_days: bool
    alert_1_day: bool
    alert_ssl_errors: bool
    alert_geo_changes: bool
    alert_cert_expired: bool
    email_notifications: bool
    email_address: Optional[EmailStr] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AlertRead(BaseModel):
    """Schema for reading alert"""
    id: int
    user_id: int
    organization_id: Optional[int] = None
    domain: str
    alert_type: str
    severity: str
    message: str
    is_read: bool
    is_resolved: bool
    created_at: datetime
    resolved_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class MonitorCreate(BaseModel):
    """Schema for creating a monitor"""
    domain: str = Field(..., description="Domain name to monitor", min_length=1, max_length=255)
    port: int = Field(default=443, description="Port number to check", ge=1, le=65535)
    check_interval: int = Field(default=3600, description="Check interval in seconds", ge=60)
    alerts_enabled: bool = Field(default=True, description="Enable alerts for this domain")
    webhook_url: Optional[str] = None

    class Config:
        from_attributes = True


class MonitorUpdate(BaseModel):
    """Schema for updating a monitor"""
    port: Optional[int] = Field(None, ge=1, le=65535)
    check_interval: Optional[int] = Field(None, ge=60)
    alerts_enabled: Optional[bool] = None
    webhook_url: Optional[str] = None
    status: Optional[str] = None

    class Config:
        from_attributes = True


class MonitorRead(BaseModel):
    """Schema for reading a monitor"""
    id: int
    user_id: int
    organization_id: Optional[int] = None
    domain: str
    port: int
    check_interval: int
    alerts_enabled: bool
    webhook_url: Optional[str] = None
    last_check: Optional[datetime] = None
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
