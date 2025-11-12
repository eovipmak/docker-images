"""
Script to create sample data for SSL Monitor database.
This helps test the database schema and verify that migrations work correctly.
"""
from passlib.hash import bcrypt
from sqlalchemy.orm import Session
from datetime import datetime, timedelta
import json

from database import SessionLocal, User, Monitor, SSLCheck


def create_sample_data():
    """Create sample users, monitors, and SSL checks."""
    db = SessionLocal()
    
    try:
        # Check if data already exists
        existing_users = db.query(User).count()
        if existing_users > 0:
            print(f"Database already contains {existing_users} users. Skipping sample data creation.")
            return
        
        print("Creating sample users...")
        # Create admin user
        admin_user = User(
            username="admin",
            password_hash=bcrypt.hash("admin123"),
            role="admin",
            created_at=datetime.utcnow()
        )
        db.add(admin_user)
        
        # Create regular user
        regular_user = User(
            username="user1",
            password_hash=bcrypt.hash("user123"),
            role="user",
            created_at=datetime.utcnow()
        )
        db.add(regular_user)
        
        # Commit to get user IDs
        db.commit()
        db.refresh(admin_user)
        db.refresh(regular_user)
        
        print(f"Created users: admin (ID: {admin_user.id}), user1 (ID: {regular_user.id})")
        
        print("Creating sample monitors...")
        # Create monitors
        monitor1 = Monitor(
            user_id=admin_user.id,
            domain="google.com",
            check_interval=3600,  # 1 hour
            webhook_url="https://hooks.slack.com/services/example",
            last_check=datetime.utcnow() - timedelta(minutes=30),
            status="active",
            created_at=datetime.utcnow()
        )
        db.add(monitor1)
        
        monitor2 = Monitor(
            user_id=regular_user.id,
            domain="github.com",
            check_interval=7200,  # 2 hours
            webhook_url=None,
            last_check=datetime.utcnow() - timedelta(hours=1),
            status="active",
            created_at=datetime.utcnow()
        )
        db.add(monitor2)
        
        monitor3 = Monitor(
            user_id=admin_user.id,
            domain="cloudflare.com",
            check_interval=1800,  # 30 minutes
            webhook_url="https://discord.com/api/webhooks/example",
            last_check=None,
            status="paused",
            created_at=datetime.utcnow()
        )
        db.add(monitor3)
        
        db.commit()
        print(f"Created {db.query(Monitor).count()} monitors")
        
        print("Creating sample SSL check history...")
        # Create SSL check history
        ssl_check1 = SSLCheck(
            domain="google.com",
            ip="142.250.185.78",
            port=443,
            status="success",
            ssl_status="success",
            server_status="success",
            ip_status="success",
            checked_at=datetime.utcnow() - timedelta(hours=2),
            response_data=json.dumps({
                "status": "success",
                "timestamp": (datetime.utcnow() - timedelta(hours=2)).isoformat(),
                "data": {
                    "domain": "google.com",
                    "ip": "142.250.185.78",
                    "port": 443,
                    "sslStatus": "success",
                    "serverStatus": "success",
                    "ipStatus": "success"
                }
            })
        )
        db.add(ssl_check1)
        
        ssl_check2 = SSLCheck(
            domain="github.com",
            ip="140.82.121.4",
            port=443,
            status="success",
            ssl_status="success",
            server_status="success",
            ip_status="success",
            checked_at=datetime.utcnow() - timedelta(hours=1),
            response_data=json.dumps({
                "status": "success",
                "timestamp": (datetime.utcnow() - timedelta(hours=1)).isoformat(),
                "data": {
                    "domain": "github.com",
                    "ip": "140.82.121.4",
                    "port": 443,
                    "sslStatus": "success",
                    "serverStatus": "success",
                    "ipStatus": "success"
                }
            })
        )
        db.add(ssl_check2)
        
        ssl_check3 = SSLCheck(
            domain="google.com",
            ip="142.250.185.78",
            port=443,
            status="success",
            ssl_status="success",
            server_status="success",
            ip_status="success",
            checked_at=datetime.utcnow() - timedelta(minutes=30),
            response_data=json.dumps({
                "status": "success",
                "timestamp": (datetime.utcnow() - timedelta(minutes=30)).isoformat(),
                "data": {
                    "domain": "google.com",
                    "ip": "142.250.185.78",
                    "port": 443,
                    "sslStatus": "success",
                    "serverStatus": "success",
                    "ipStatus": "success"
                }
            })
        )
        db.add(ssl_check3)
        
        db.commit()
        print(f"Created {db.query(SSLCheck).count()} SSL check history records")
        
        print("\n=== Sample Data Summary ===")
        print(f"Users: {db.query(User).count()}")
        print(f"  - Admin users: {db.query(User).filter(User.role == 'admin').count()}")
        print(f"  - Regular users: {db.query(User).filter(User.role == 'user').count()}")
        print(f"Monitors: {db.query(Monitor).count()}")
        print(f"  - Active monitors: {db.query(Monitor).filter(Monitor.status == 'active').count()}")
        print(f"  - Paused monitors: {db.query(Monitor).filter(Monitor.status == 'paused').count()}")
        print(f"SSL Checks: {db.query(SSLCheck).count()}")
        
        print("\n=== Credentials ===")
        print("Admin: username='admin', password='admin123'")
        print("User: username='user1', password='user123'")
        
        print("\nSample data created successfully!")
        
    except Exception as e:
        print(f"Error creating sample data: {e}")
        db.rollback()
        raise
    finally:
        db.close()


if __name__ == "__main__":
    create_sample_data()
