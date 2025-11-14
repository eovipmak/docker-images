"""
Database models for SSL Monitor
"""
import os
from sqlalchemy import Column, Integer, String, DateTime, Text, ForeignKey, Boolean, create_engine
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime
from fastapi_users.db import SQLAlchemyBaseUserTable
from typing import AsyncGenerator

Base = declarative_base()


class Organization(Base):
    """Model for storing organization information"""
    __tablename__ = "organizations"
    
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False, index=True)
    description = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    # Relationships
    users = relationship("User", back_populates="organization")
    monitors = relationship("Monitor", back_populates="organization", cascade="all, delete-orphan")
    ssl_checks = relationship("SSLCheck", back_populates="organization", cascade="all, delete-orphan")
    alerts = relationship("Alert", back_populates="organization", cascade="all, delete-orphan")
    alert_configs = relationship("AlertConfig", back_populates="organization", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Organization(name={self.name})>"


class User(SQLAlchemyBaseUserTable[int], Base):
    """Model for storing user information with authentication fields"""
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, nullable=False, index=True)
    hashed_password = Column(String, nullable=False)
    is_active = Column(Boolean, nullable=False, default=True)
    is_superuser = Column(Boolean, nullable=False, default=False)
    is_verified = Column(Boolean, nullable=False, default=False)
    organization_id = Column(Integer, ForeignKey("organizations.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    # Relationships
    organization = relationship("Organization", back_populates="users")
    monitors = relationship("Monitor", back_populates="user", cascade="all, delete-orphan")
    ssl_checks = relationship("SSLCheck", back_populates="user", cascade="all, delete-orphan")
    alerts = relationship("Alert", back_populates="user", cascade="all, delete-orphan")
    alert_configs = relationship("AlertConfig", back_populates="user", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<User(email={self.email}, is_verified={self.is_verified})>"

class Monitor(Base):
    """Model for storing monitor configurations"""
    __tablename__ = "monitors"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    organization_id = Column(Integer, ForeignKey("organizations.id"), nullable=True, index=True)
    domain = Column(String, nullable=False, index=True)
    port = Column(Integer, default=443)  # Port to monitor
    check_interval = Column(Integer, default=3600)  # in seconds, default 1 hour
    webhook_url = Column(String, nullable=True)
    alerts_enabled = Column(Boolean, default=True)  # Enable/disable alerts for this domain
    last_check = Column(DateTime, nullable=True)
    status = Column(String, default="active")  # active, paused, error
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    user = relationship("User", back_populates="monitors")
    organization = relationship("Organization", back_populates="monitors")
    
    def __repr__(self):
        return f"<Monitor(domain={self.domain}, user_id={self.user_id}, status={self.status})>"


class SSLCheck(Base):
    """Model for storing SSL check history"""
    __tablename__ = "ssl_checks"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=True, index=True)
    organization_id = Column(Integer, ForeignKey("organizations.id"), nullable=True, index=True)
    domain = Column(String, index=True)
    ip = Column(String)
    port = Column(Integer, default=443)
    status = Column(String)  # success, error, warning
    ssl_status = Column(String)
    server_status = Column(String)
    ip_status = Column(String)
    checked_at = Column(DateTime, default=datetime.utcnow, index=True)
    response_data = Column(Text)  # Store full JSON response
    
    # Relationships
    user = relationship("User", back_populates="ssl_checks")
    organization = relationship("Organization", back_populates="ssl_checks")
    
    def __repr__(self):
        return f"<SSLCheck(domain={self.domain}, status={self.status}, checked_at={self.checked_at})>"


class Alert(Base):
    """Model for storing alerts and notifications"""
    __tablename__ = "alerts"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    organization_id = Column(Integer, ForeignKey("organizations.id"), nullable=True, index=True)
    domain = Column(String, nullable=False, index=True)
    alert_type = Column(String, nullable=False)  # expiring_soon, expired, invalid, error
    severity = Column(String, default="medium")  # low, medium, high, critical
    message = Column(Text, nullable=False)
    is_read = Column(Boolean, default=False)
    is_resolved = Column(Boolean, default=False)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)
    resolved_at = Column(DateTime, nullable=True)
    
    # Relationships
    user = relationship("User", back_populates="alerts")
    organization = relationship("Organization", back_populates="alerts")
    
    def __repr__(self):
        return f"<Alert(domain={self.domain}, type={self.alert_type}, severity={self.severity})>"


class AlertConfig(Base):
    """Model for storing user-specific alert configuration"""
    __tablename__ = "alert_configs"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    organization_id = Column(Integer, ForeignKey("organizations.id"), nullable=True, index=True)
    
    # Alert settings
    enabled = Column(Boolean, default=True)
    webhook_url = Column(String, nullable=True)
    
    # Certificate expiration thresholds (in days)
    alert_30_days = Column(Boolean, default=True)
    alert_7_days = Column(Boolean, default=True)
    alert_1_day = Column(Boolean, default=True)
    
    # Alert types
    alert_ssl_errors = Column(Boolean, default=True)
    alert_geo_changes = Column(Boolean, default=False)
    alert_cert_expired = Column(Boolean, default=True)
    
    # Notification preferences
    email_notifications = Column(Boolean, default=False)
    email_address = Column(String, nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    user = relationship("User", back_populates="alert_configs")
    organization = relationship("Organization", back_populates="alert_configs")
    
    def __repr__(self):
        return f"<AlertConfig(user_id={self.user_id}, enabled={self.enabled})>"


# Database setup - Synchronous for main app
DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./ssl_monitor.db")
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Async database setup for FastAPI Users
ASYNC_DATABASE_URL = os.getenv("ASYNC_DATABASE_URL", "sqlite+aiosqlite:///./ssl_monitor.db")
async_engine = create_async_engine(ASYNC_DATABASE_URL)
async_session_maker = async_sessionmaker(async_engine, expire_on_commit=False)


def init_db():
    """Initialize database tables"""
    Base.metadata.create_all(bind=engine)


def get_db():
    """Get synchronous database session"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


async def get_async_session() -> AsyncGenerator[AsyncSession, None]:
    """Get async database session for FastAPI Users"""
    async with async_session_maker() as session:
        yield session
