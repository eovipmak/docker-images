#!/usr/bin/env python3
"""
Database Backup Script for SSL Monitor

This script creates a backup of the SQLite database by copying the file
to a backup location with a timestamp.

Usage:
    python backup_database.py [--output-dir DIRECTORY]
"""
import os
import shutil
import argparse
from datetime import datetime
from pathlib import Path


def get_database_path():
    """Get the database path from environment variable or default"""
    db_url = os.getenv("DATABASE_URL", "sqlite:///./ssl_monitor.db")
    # Extract file path from SQLite URL
    if db_url.startswith("sqlite:///"):
        db_path = db_url.replace("sqlite:///", "")
        # Handle relative paths
        if db_path.startswith("./"):
            db_path = db_path[2:]
        return Path(db_path)
    else:
        raise ValueError(f"Unsupported database URL format: {db_url}")


def backup_database(output_dir: Path = None):
    """
    Create a backup of the database file
    
    Args:
        output_dir: Directory to store backups (default: ./backups)
        
    Returns:
        Path to the backup file
    """
    # Get database path
    db_path = get_database_path()
    
    if not db_path.exists():
        raise FileNotFoundError(f"Database file not found: {db_path}")
    
    # Set default output directory if not provided
    if output_dir is None:
        output_dir = Path("./backups")
    
    # Create output directory if it doesn't exist
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Create backup filename with timestamp
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_filename = f"ssl_monitor_backup_{timestamp}.db"
    backup_path = output_dir / backup_filename
    
    # Copy the database file
    print(f"Creating backup of {db_path}")
    print(f"Destination: {backup_path}")
    
    shutil.copy2(db_path, backup_path)
    
    # Verify backup was created
    if backup_path.exists():
        backup_size = backup_path.stat().st_size
        print(f"✓ Backup created successfully ({backup_size:,} bytes)")
        return backup_path
    else:
        raise RuntimeError("Backup file was not created")


def main():
    """Main function to run the backup script"""
    parser = argparse.ArgumentParser(
        description="Backup SSL Monitor SQLite database"
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("./backups"),
        help="Directory to store backups (default: ./backups)"
    )
    
    args = parser.parse_args()
    
    try:
        backup_path = backup_database(args.output_dir)
        print(f"\nBackup completed: {backup_path.absolute()}")
        return 0
    except Exception as e:
        print(f"\n✗ Backup failed: {str(e)}")
        return 1


if __name__ == "__main__":
    exit(main())
