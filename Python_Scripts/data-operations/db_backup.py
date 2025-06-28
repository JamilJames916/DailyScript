#!/usr/bin/env python3
"""
Database Backup Script

This script provides comprehensive database backup functionality for PostgreSQL, 
MySQL, MongoDB, and SQLite databases. It supports local and cloud storage (AWS S3).

Features:
- Multiple database types support
- Compression options
- S3 upload capability
- Retention policy management
- Email notifications
- Backup verification

Usage:
    python db_backup.py --db-type postgresql --config config.yaml
    python db_backup.py --db-type mongodb --s3-upload --retention-days 30
"""

import os
import sys
import subprocess
import gzip
import shutil
import logging
import argparse
import yaml
import boto3
import smtplib
from datetime import datetime, timedelta
from pathlib import Path
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('backup.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class DatabaseBackup:
    def __init__(self, config_file=None):
        self.config = self.load_config(config_file)
        self.backup_dir = Path(self.config.get('backup_dir', './backups'))
        self.backup_dir.mkdir(exist_ok=True)
        
        # AWS S3 setup
        if self.config.get('s3_upload', False):
            self.s3_client = boto3.client(
                's3',
                aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
                aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY'),
                region_name=os.getenv('AWS_DEFAULT_REGION')
            )
            self.s3_bucket = os.getenv('BACKUP_S3_BUCKET')

    def load_config(self, config_file):
        """Load configuration from file or use defaults"""
        default_config = {
            'backup_dir': './backups',
            'compress': True,
            's3_upload': False,
            'retention_days': 30,
            'email_notifications': False,
            'verify_backup': True
        }
        
        if config_file and Path(config_file).exists():
            with open(config_file, 'r') as f:
                file_config = yaml.safe_load(f)
                default_config.update(file_config)
        
        return default_config

    def backup_postgresql(self):
        """Backup PostgreSQL database"""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            db_name = os.getenv('DB_NAME')
            backup_file = self.backup_dir / f"postgresql_{db_name}_{timestamp}.sql"
            
            # pg_dump command
            cmd = [
                'pg_dump',
                f"--host={os.getenv('DB_HOST')}",
                f"--port={os.getenv('DB_PORT')}",
                f"--username={os.getenv('DB_USER')}",
                '--verbose',
                '--no-password',
                '--format=custom',
                '--file', str(backup_file),
                db_name
            ]
            
            env = os.environ.copy()
            env['PGPASSWORD'] = os.getenv('DB_PASSWORD')
            
            logger.info(f"Starting PostgreSQL backup: {backup_file}")
            result = subprocess.run(cmd, env=env, capture_output=True, text=True)
            
            if result.returncode == 0:
                logger.info("PostgreSQL backup completed successfully")
                return self.post_backup_process(backup_file)
            else:
                logger.error(f"PostgreSQL backup failed: {result.stderr}")
                return False
                
        except Exception as e:
            logger.error(f"PostgreSQL backup error: {str(e)}")
            return False

    def backup_mysql(self):
        """Backup MySQL database"""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            db_name = os.getenv('DB_NAME')
            backup_file = self.backup_dir / f"mysql_{db_name}_{timestamp}.sql"
            
            # mysqldump command
            cmd = [
                'mysqldump',
                f"--host={os.getenv('DB_HOST')}",
                f"--port={os.getenv('DB_PORT')}",
                f"--user={os.getenv('DB_USER')}",
                f"--password={os.getenv('DB_PASSWORD')}",
                '--single-transaction',
                '--routines',
                '--triggers',
                db_name
            ]
            
            logger.info(f"Starting MySQL backup: {backup_file}")
            with open(backup_file, 'w') as f:
                result = subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, text=True)
            
            if result.returncode == 0:
                logger.info("MySQL backup completed successfully")
                return self.post_backup_process(backup_file)
            else:
                logger.error(f"MySQL backup failed: {result.stderr}")
                return False
                
        except Exception as e:
            logger.error(f"MySQL backup error: {str(e)}")
            return False

    def backup_mongodb(self):
        """Backup MongoDB database"""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            db_name = os.getenv('MONGO_DB')
            backup_dir = self.backup_dir / f"mongodb_{db_name}_{timestamp}"
            
            # mongodump command
            cmd = [
                'mongodump',
                '--uri', os.getenv('MONGO_URI'),
                '--db', db_name,
                '--out', str(backup_dir)
            ]
            
            logger.info(f"Starting MongoDB backup: {backup_dir}")
            result = subprocess.run(cmd, capture_output=True, text=True)
            
            if result.returncode == 0:
                logger.info("MongoDB backup completed successfully")
                # Create tar.gz archive
                archive_path = f"{backup_dir}.tar.gz"
                shutil.make_archive(str(backup_dir), 'gztar', str(backup_dir))
                shutil.rmtree(backup_dir)  # Remove uncompressed directory
                return self.post_backup_process(Path(archive_path))
            else:
                logger.error(f"MongoDB backup failed: {result.stderr}")
                return False
                
        except Exception as e:
            logger.error(f"MongoDB backup error: {str(e)}")
            return False

    def backup_sqlite(self, db_file):
        """Backup SQLite database"""
        try:
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            db_name = Path(db_file).stem
            backup_file = self.backup_dir / f"sqlite_{db_name}_{timestamp}.db"
            
            logger.info(f"Starting SQLite backup: {backup_file}")
            shutil.copy2(db_file, backup_file)
            logger.info("SQLite backup completed successfully")
            
            return self.post_backup_process(backup_file)
            
        except Exception as e:
            logger.error(f"SQLite backup error: {str(e)}")
            return False

    def post_backup_process(self, backup_file):
        """Post-backup processing: compress, upload, verify"""
        try:
            # Compress if enabled
            if self.config.get('compress', True) and not str(backup_file).endswith('.gz'):
                compressed_file = self.compress_file(backup_file)
                if compressed_file:
                    backup_file = compressed_file
            
            # Verify backup
            if self.config.get('verify_backup', True):
                if not self.verify_backup(backup_file):
                    logger.error("Backup verification failed")
                    return False
            
            # Upload to S3
            if self.config.get('s3_upload', False):
                if not self.upload_to_s3(backup_file):
                    logger.error("S3 upload failed")
                    return False
            
            # Clean up old backups
            self.cleanup_old_backups()
            
            # Send notification
            if self.config.get('email_notifications', False):
                self.send_notification(backup_file, success=True)
            
            logger.info(f"Backup process completed successfully: {backup_file}")
            return True
            
        except Exception as e:
            logger.error(f"Post-backup processing error: {str(e)}")
            if self.config.get('email_notifications', False):
                self.send_notification(backup_file, success=False, error=str(e))
            return False

    def compress_file(self, file_path):
        """Compress backup file using gzip"""
        try:
            compressed_path = Path(f"{file_path}.gz")
            
            with open(file_path, 'rb') as f_in:
                with gzip.open(compressed_path, 'wb') as f_out:
                    shutil.copyfileobj(f_in, f_out)
            
            # Remove original file
            os.remove(file_path)
            logger.info(f"File compressed: {compressed_path}")
            return compressed_path
            
        except Exception as e:
            logger.error(f"Compression error: {str(e)}")
            return None

    def verify_backup(self, backup_file):
        """Verify backup file integrity"""
        try:
            if not backup_file.exists():
                return False
            
            # Check file size
            if backup_file.stat().st_size == 0:
                logger.error("Backup file is empty")
                return False
            
            # For compressed files, try to decompress a small portion
            if str(backup_file).endswith('.gz'):
                try:
                    with gzip.open(backup_file, 'rb') as f:
                        f.read(1024)  # Read first 1KB
                except Exception:
                    logger.error("Compressed backup file is corrupted")
                    return False
            
            logger.info("Backup verification passed")
            return True
            
        except Exception as e:
            logger.error(f"Backup verification error: {str(e)}")
            return False

    def upload_to_s3(self, backup_file):
        """Upload backup file to S3"""
        try:
            if not self.s3_bucket:
                logger.error("S3 bucket not configured")
                return False
            
            s3_key = f"database-backups/{backup_file.name}"
            
            logger.info(f"Uploading to S3: s3://{self.s3_bucket}/{s3_key}")
            self.s3_client.upload_file(
                str(backup_file),
                self.s3_bucket,
                s3_key,
                ExtraArgs={'StorageClass': 'STANDARD_IA'}
            )
            
            logger.info("S3 upload completed successfully")
            return True
            
        except Exception as e:
            logger.error(f"S3 upload error: {str(e)}")
            return False

    def cleanup_old_backups(self):
        """Remove old backup files based on retention policy"""
        try:
            retention_days = self.config.get('retention_days', 30)
            cutoff_date = datetime.now() - timedelta(days=retention_days)
            
            deleted_count = 0
            for backup_file in self.backup_dir.glob('*'):
                if backup_file.is_file():
                    file_time = datetime.fromtimestamp(backup_file.stat().st_mtime)
                    if file_time < cutoff_date:
                        backup_file.unlink()
                        deleted_count += 1
                        logger.info(f"Deleted old backup: {backup_file}")
            
            logger.info(f"Cleanup completed: {deleted_count} old backups deleted")
            
        except Exception as e:
            logger.error(f"Cleanup error: {str(e)}")

    def send_notification(self, backup_file, success=True, error=None):
        """Send email notification about backup status"""
        try:
            smtp_host = os.getenv('SMTP_HOST')
            smtp_port = int(os.getenv('SMTP_PORT', 587))
            smtp_user = os.getenv('SMTP_USER')
            smtp_password = os.getenv('SMTP_PASSWORD')
            
            if not all([smtp_host, smtp_user, smtp_password]):
                logger.warning("Email configuration incomplete, skipping notification")
                return
            
            msg = MIMEMultipart()
            msg['From'] = smtp_user
            msg['To'] = smtp_user  # Send to self
            
            if success:
                msg['Subject'] = f"Database Backup Successful - {backup_file.name}"
                body = f"""
Database backup completed successfully!

Backup File: {backup_file.name}
File Size: {backup_file.stat().st_size / (1024*1024):.2f} MB
Timestamp: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

The backup has been stored locally and uploaded to S3 (if configured).
                """
            else:
                msg['Subject'] = f"Database Backup Failed - {datetime.now().strftime('%Y-%m-%d')}"
                body = f"""
Database backup failed!

Error: {error}
Timestamp: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

Please check the backup logs for more details.
                """
            
            msg.attach(MIMEText(body, 'plain'))
            
            with smtplib.SMTP(smtp_host, smtp_port) as server:
                server.starttls()
                server.login(smtp_user, smtp_password)
                server.send_message(msg)
            
            logger.info("Email notification sent successfully")
            
        except Exception as e:
            logger.error(f"Email notification error: {str(e)}")

def main():
    parser = argparse.ArgumentParser(description='Database Backup Script')
    parser.add_argument('--db-type', choices=['postgresql', 'mysql', 'mongodb', 'sqlite'], 
                       required=True, help='Database type to backup')
    parser.add_argument('--config', help='Configuration file path')
    parser.add_argument('--sqlite-file', help='SQLite database file path (required for SQLite)')
    parser.add_argument('--compress', action='store_true', help='Compress backup files')
    parser.add_argument('--s3-upload', action='store_true', help='Upload to S3')
    parser.add_argument('--retention-days', type=int, default=30, 
                       help='Backup retention period in days')
    
    args = parser.parse_args()
    
    # Override config with command line arguments
    config_override = {
        'compress': args.compress,
        's3_upload': args.s3_upload,
        'retention_days': args.retention_days
    }
    
    backup = DatabaseBackup(args.config)
    backup.config.update(config_override)
    
    success = False
    
    if args.db_type == 'postgresql':
        success = backup.backup_postgresql()
    elif args.db_type == 'mysql':
        success = backup.backup_mysql()
    elif args.db_type == 'mongodb':
        success = backup.backup_mongodb()
    elif args.db_type == 'sqlite':
        if not args.sqlite_file:
            logger.error("SQLite file path required for SQLite backup")
            sys.exit(1)
        success = backup.backup_sqlite(args.sqlite_file)
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
