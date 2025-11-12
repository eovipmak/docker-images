"""
Database models for SSL Monitor
"""
from sqlalchemy import Column, Integer, String, DateTime, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from datetime import datetime

Base = declarative_base()


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
