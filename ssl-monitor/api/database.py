"""
Database models for SSL Monitor
"""
from sqlalchemy import Column, Integer, String, DateTime, Text, ForeignKey, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime

Base = declarative_base()


class User(Base):
    """Model for storing user information"""
    __tablename__ = "users"
    
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String, unique=True, nullable=False, index=True)
    password_hash = Column(String, nullable=False)
    role = Column(String, nullable=False, default="user")  # admin or user
    created_at = Column(DateTime, default=datetime.utcnow)
    
    # Relationship to monitors
    monitors = relationship("Monitor", back_populates="user", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<User(username={self.username}, role={self.role})>"


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


# Database setup
DATABASE_URL = "sqlite:///./ssl_monitor.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def init_db():
    """Initialize database tables"""
    Base.metadata.create_all(bind=engine)


def get_db():
    """Get database session"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
