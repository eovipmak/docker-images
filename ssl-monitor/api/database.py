"""Database configuration and models for SSL Monitor."""
from sqlalchemy import create_engine, Column, Integer, String, DateTime, Text, Boolean
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from datetime import datetime
import os

# Database configuration
DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./ssl_monitor.db")

engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False} if "sqlite" in DATABASE_URL else {}
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

Base = declarative_base()


class SSLCheck(Base):
    """Model for storing SSL check results."""
    __tablename__ = "ssl_checks"

    id = Column(Integer, primary_key=True, index=True)
    domain = Column(String, index=True)
    ip = Column(String)
    port = Column(Integer, default=443)
    ssl_valid = Column(Boolean)
    certificate_info = Column(Text)  # JSON string
    server_info = Column(String)
    ip_info = Column(Text)  # JSON string
    alerts = Column(Text)  # JSON string
    recommendations = Column(Text)  # JSON string
    checked_at = Column(DateTime, default=lambda: datetime.now(timezone.utc))
    error_message = Column(Text, nullable=True)


def init_db():
    """Initialize database tables."""
    Base.metadata.create_all(bind=engine)


def get_db():
    """Get database session."""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
