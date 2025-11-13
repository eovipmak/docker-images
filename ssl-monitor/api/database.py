"""
Database models for SSL Monitor
"""
from sqlalchemy import Column, Integer, String, DateTime, Text, ForeignKey, Boolean, create_engine
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime
from fastapi_users.db import SQLAlchemyBaseUserTable
from typing import AsyncGenerator

Base = declarative_base()


class User(SQLAlchemyBaseUserTable[int], Base):
    """Model for storing user information with authentication fields"""
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, nullable=False, index=True)
    hashed_password = Column(String, nullable=False)
    is_active = Column(Boolean, nullable=False, default=True)
    is_superuser = Column(Boolean, nullable=False, default=False)
    is_verified = Column(Boolean, nullable=False, default=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    # Relationship to monitors
    monitors = relationship("Monitor", back_populates="user", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<User(email={self.email}, is_verified={self.is_verified})>"

class Monitor(Base):
    """Model for storing monitor configurations"""
    __tablename__ = "monitors"
    
    id = Column(Integer, primary_key=True, index=True)
    user_id = Column(Integer, ForeignKey("users.id"), nullable=False)
    domain = Column(String, nullable=False, index=True)
    check_interval = Column(Integer, default=3600)  # in seconds, default 1 hour
    webhook_url = Column(String, nullable=True)
    last_check = Column(DateTime, nullable=True)
    status = Column(String, default="active")  # active, paused, error
    created_at = Column(DateTime, default=datetime.utcnow)
    
    # Relationship to user
    user = relationship("User", back_populates="monitors")
    
    def __repr__(self):
        return f"<Monitor(domain={self.domain}, user_id={self.user_id}, status={self.status})>"


class SSLCheck(Base):
    """Model for storing SSL check history"""
    __tablename__ = "ssl_checks"
    
    id = Column(Integer, primary_key=True, index=True)
    domain = Column(String, index=True)
    ip = Column(String)
    port = Column(Integer, default=443)
    status = Column(String)  # success, error, warning
    ssl_status = Column(String)
    server_status = Column(String)
    ip_status = Column(String)
    checked_at = Column(DateTime, default=datetime.utcnow)
    response_data = Column(Text)  # Store full JSON response
    
    def __repr__(self):
        return f"<SSLCheck(domain={self.domain}, status={self.status}, checked_at={self.checked_at})>"


# Database setup - Synchronous for main app
DATABASE_URL = "sqlite:///./ssl_monitor.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Async database setup for FastAPI Users
ASYNC_DATABASE_URL = "sqlite+aiosqlite:///./ssl_monitor.db"
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
