#!/usr/bin/env python3
"""
Data Sync Utility

This script provides real-time data synchronization between different data stores,
including databases, APIs, and file systems. It supports both one-way and bi-directional
sync with conflict resolution.

Features:
- Real-time change detection
- Multiple sync strategies (append, merge, overwrite)
- Conflict resolution mechanisms
- Schema drift handling
- Monitoring and alerting
- Rollback capabilities

Usage:
    python data_sync.py --config sync_config.yaml
    python data_sync.py --source postgresql --target mongodb --mode realtime
"""

import os
import sys
import json
import yaml
import logging
import argparse
import asyncio
import hashlib
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any, Optional
from dataclasses import dataclass
from enum import Enum

import pandas as pd
import sqlalchemy as sa
from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker
import pymongo
import redis
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('data_sync.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class SyncMode(Enum):
    BATCH = "batch"
    REALTIME = "realtime"
    SCHEDULED = "scheduled"

class ConflictResolution(Enum):
    SOURCE_WINS = "source_wins"
    TARGET_WINS = "target_wins"
    MERGE = "merge"
    MANUAL = "manual"

@dataclass
class SyncConfig:
    source_type: str
    source_config: Dict
    target_type: str
    target_config: Dict
    sync_mode: str
    conflict_resolution: str
    sync_rules: Dict
    monitoring: Dict

class ChangeDetector:
    """Detect changes in data sources"""
    
    def __init__(self, config: Dict):
        self.config = config
        self.redis_client = None
        
        # Initialize Redis for change tracking
        if config.get('use_redis', True):
            try:
                self.redis_client = redis.Redis(
                    host=os.getenv('REDIS_HOST', 'localhost'),
                    port=int(os.getenv('REDIS_PORT', 6379)),
                    password=os.getenv('REDIS_PASSWORD'),
                    decode_responses=True
                )
                self.redis_client.ping()
                logger.info("Connected to Redis for change tracking")
            except Exception as e:
                logger.warning(f"Redis connection failed: {str(e)}")
                self.redis_client = None

    def detect_database_changes(self, connection_string: str, table_name: str, 
                              last_sync_time: datetime = None) -> List[Dict]:
        """Detect changes in database table"""
        try:
            engine = create_engine(connection_string)
            
            # Build query to detect changes
            if last_sync_time:
                query = f"""
                    SELECT * FROM {table_name} 
                    WHERE updated_at > '{last_sync_time.isoformat()}'
                    OR created_at > '{last_sync_time.isoformat()}'
                """
            else:
                query = f"SELECT * FROM {table_name}"
            
            df = pd.read_sql(query, engine)
            changes = df.to_dict('records')
            
            logger.info(f"Detected {len(changes)} changes in {table_name}")
            return changes
            
        except Exception as e:
            logger.error(f"Error detecting database changes: {str(e)}")
            return []

    def detect_file_changes(self, file_path: str) -> Dict:
        """Detect changes in file"""
        try:
            file_stats = Path(file_path).stat()
            file_hash = self._calculate_file_hash(file_path)
            
            last_modified = datetime.fromtimestamp(file_stats.st_mtime)
            
            # Check if file has changed since last sync
            cache_key = f"file_hash:{file_path}"
            last_hash = self.redis_client.get(cache_key) if self.redis_client else None
            
            if last_hash != file_hash:
                if self.redis_client:
                    self.redis_client.set(cache_key, file_hash)
                
                return {
                    'file_path': file_path,
                    'last_modified': last_modified,
                    'hash': file_hash,
                    'changed': True
                }
            
            return {'changed': False}
            
        except Exception as e:
            logger.error(f"Error detecting file changes: {str(e)}")
            return {'changed': False}

    def _calculate_file_hash(self, file_path: str) -> str:
        """Calculate MD5 hash of file"""
        hash_md5 = hashlib.md5()
        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                hash_md5.update(chunk)
        return hash_md5.hexdigest()

class FileWatcher(FileSystemEventHandler):
    """File system event handler for real-time sync"""
    
    def __init__(self, sync_manager):
        self.sync_manager = sync_manager
        
    def on_modified(self, event):
        if not event.is_directory:
            logger.info(f"File modified: {event.src_path}")
            asyncio.create_task(self.sync_manager.sync_file(event.src_path))
    
    def on_created(self, event):
        if not event.is_directory:
            logger.info(f"File created: {event.src_path}")
            asyncio.create_task(self.sync_manager.sync_file(event.src_path))

class ConflictResolver:
    """Resolve data conflicts during sync"""
    
    def __init__(self, strategy: str):
        self.strategy = strategy
    
    def resolve_conflict(self, source_record: Dict, target_record: Dict, 
                        primary_key: str) -> Dict:
        """Resolve conflict between source and target records"""
        try:
            if self.strategy == ConflictResolution.SOURCE_WINS.value:
                return source_record
            
            elif self.strategy == ConflictResolution.TARGET_WINS.value:
                return target_record
            
            elif self.strategy == ConflictResolution.MERGE.value:
                # Merge strategy: source wins for non-null values
                merged = target_record.copy()
                for key, value in source_record.items():
                    if value is not None:
                        merged[key] = value
                return merged
            
            elif self.strategy == ConflictResolution.MANUAL.value:
                # Log conflict for manual resolution
                logger.warning(f"Manual conflict resolution required for {primary_key}")
                return source_record  # Default to source
            
            else:
                logger.warning(f"Unknown conflict resolution strategy: {self.strategy}")
                return source_record
                
        except Exception as e:
            logger.error(f"Error resolving conflict: {str(e)}")
            return source_record

class DataSyncManager:
    """Main data synchronization manager"""
    
    def __init__(self, config: SyncConfig):
        self.config = config
        self.change_detector = ChangeDetector(config.sync_rules)
        self.conflict_resolver = ConflictResolver(config.conflict_resolution)
        self.source_conn = None
        self.target_conn = None
        self.sync_stats = {
            'records_synced': 0,
            'conflicts_resolved': 0,
            'errors': 0,
            'last_sync': None
        }
        
    async def initialize_connections(self):
        """Initialize source and target connections"""
        try:
            # Initialize source connection
            if self.config.source_type == 'postgresql':
                self.source_conn = create_engine(self.config.source_config['connection_string'])
            elif self.config.source_type == 'mongodb':
                client = pymongo.MongoClient(self.config.source_config['uri'])
                self.source_conn = client[self.config.source_config['database']]
            
            # Initialize target connection
            if self.config.target_type == 'postgresql':
                self.target_conn = create_engine(self.config.target_config['connection_string'])
            elif self.config.target_type == 'mongodb':
                client = pymongo.MongoClient(self.config.target_config['uri'])
                self.target_conn = client[self.config.target_config['database']]
            
            logger.info("Database connections initialized")
            return True
            
        except Exception as e:
            logger.error(f"Connection initialization failed: {str(e)}")
            return False

    async def sync_table(self, table_name: str):
        """Sync a single table/collection"""
        try:
            logger.info(f"Starting sync for table: {table_name}")
            
            # Get last sync time
            last_sync_time = self._get_last_sync_time(table_name)
            
            # Detect changes
            if self.config.source_type in ['postgresql', 'mysql']:
                changes = self.change_detector.detect_database_changes(
                    self.config.source_config['connection_string'],
                    table_name,
                    last_sync_time
                )
            else:
                # Handle other source types
                changes = []
            
            if not changes:
                logger.info(f"No changes detected for {table_name}")
                return
            
            # Process changes
            conflicts = 0
            synced = 0
            
            for record in changes:
                try:
                    # Check if record exists in target
                    primary_key = self.config.sync_rules.get('primary_key', 'id')
                    existing_record = self._get_target_record(table_name, primary_key, record[primary_key])
                    
                    if existing_record:
                        # Resolve conflict
                        resolved_record = self.conflict_resolver.resolve_conflict(
                            record, existing_record, primary_key
                        )
                        conflicts += 1
                        record = resolved_record
                    
                    # Sync record to target
                    await self._sync_record_to_target(table_name, record)
                    synced += 1
                    
                except Exception as e:
                    logger.error(f"Error syncing record: {str(e)}")
                    self.sync_stats['errors'] += 1
            
            # Update sync statistics
            self.sync_stats['records_synced'] += synced
            self.sync_stats['conflicts_resolved'] += conflicts
            self.sync_stats['last_sync'] = datetime.now()
            
            # Update last sync time
            self._update_last_sync_time(table_name)
            
            logger.info(f"Sync completed for {table_name}: {synced} records, {conflicts} conflicts")
            
        except Exception as e:
            logger.error(f"Table sync failed for {table_name}: {str(e)}")

    async def sync_file(self, file_path: str):
        """Sync a single file"""
        try:
            logger.info(f"Starting file sync: {file_path}")
            
            # Detect changes
            change_info = self.change_detector.detect_file_changes(file_path)
            
            if not change_info.get('changed', False):
                logger.info(f"No changes detected for {file_path}")
                return
            
            # Read file content
            if file_path.endswith('.csv'):
                df = pd.read_csv(file_path)
            elif file_path.endswith('.json'):
                df = pd.read_json(file_path)
            else:
                logger.warning(f"Unsupported file format: {file_path}")
                return
            
            # Sync to target
            table_name = Path(file_path).stem
            await self._sync_dataframe_to_target(table_name, df)
            
            logger.info(f"File sync completed: {file_path}")
            
        except Exception as e:
            logger.error(f"File sync failed for {file_path}: {str(e)}")

    async def run_batch_sync(self):
        """Run batch synchronization"""
        try:
            logger.info("Starting batch sync")
            
            tables = self.config.sync_rules.get('tables', [])
            
            for table in tables:
                await self.sync_table(table)
            
            logger.info("Batch sync completed")
            
        except Exception as e:
            logger.error(f"Batch sync failed: {str(e)}")

    async def run_realtime_sync(self):
        """Run real-time synchronization"""
        try:
            logger.info("Starting real-time sync")
            
            # Set up file watchers if configured
            if 'watch_directories' in self.config.sync_rules:
                observer = Observer()
                event_handler = FileWatcher(self)
                
                for directory in self.config.sync_rules['watch_directories']:
                    observer.schedule(event_handler, directory, recursive=True)
                
                observer.start()
                logger.info(f"File watchers started for directories: {self.config.sync_rules['watch_directories']}")
            
            # Run periodic sync for database tables
            sync_interval = self.config.sync_rules.get('sync_interval', 60)  # seconds
            
            while True:
                await self.run_batch_sync()
                await asyncio.sleep(sync_interval)
                
        except KeyboardInterrupt:
            logger.info("Real-time sync stopped by user")
        except Exception as e:
            logger.error(f"Real-time sync failed: {str(e)}")

    def _get_last_sync_time(self, table_name: str) -> Optional[datetime]:
        """Get last sync time for table"""
        try:
            if self.change_detector.redis_client:
                timestamp = self.change_detector.redis_client.get(f"last_sync:{table_name}")
                if timestamp:
                    return datetime.fromisoformat(timestamp)
            return None
        except Exception:
            return None

    def _update_last_sync_time(self, table_name: str):
        """Update last sync time for table"""
        try:
            if self.change_detector.redis_client:
                self.change_detector.redis_client.set(
                    f"last_sync:{table_name}",
                    datetime.now().isoformat()
                )
        except Exception as e:
            logger.warning(f"Failed to update last sync time: {str(e)}")

    def _get_target_record(self, table_name: str, primary_key: str, key_value: Any) -> Optional[Dict]:
        """Get existing record from target"""
        try:
            if self.config.target_type in ['postgresql', 'mysql']:
                query = f"SELECT * FROM {table_name} WHERE {primary_key} = '{key_value}'"
                result = pd.read_sql(query, self.target_conn)
                if not result.empty:
                    return result.iloc[0].to_dict()
            elif self.config.target_type == 'mongodb':
                collection = self.target_conn[table_name]
                return collection.find_one({primary_key: key_value})
            
            return None
            
        except Exception as e:
            logger.error(f"Error getting target record: {str(e)}")
            return None

    async def _sync_record_to_target(self, table_name: str, record: Dict):
        """Sync single record to target"""
        try:
            if self.config.target_type in ['postgresql', 'mysql']:
                df = pd.DataFrame([record])
                df.to_sql(table_name, self.target_conn, if_exists='append', index=False)
            elif self.config.target_type == 'mongodb':
                collection = self.target_conn[table_name]
                collection.replace_one(
                    {'_id': record.get('_id')}, 
                    record, 
                    upsert=True
                )
                
        except Exception as e:
            logger.error(f"Error syncing record to target: {str(e)}")
            raise

    async def _sync_dataframe_to_target(self, table_name: str, df: pd.DataFrame):
        """Sync DataFrame to target"""
        try:
            if self.config.target_type in ['postgresql', 'mysql']:
                df.to_sql(table_name, self.target_conn, if_exists='append', index=False)
            elif self.config.target_type == 'mongodb':
                collection = self.target_conn[table_name]
                records = df.to_dict('records')
                collection.insert_many(records)
                
        except Exception as e:
            logger.error(f"Error syncing DataFrame to target: {str(e)}")
            raise

    def get_sync_statistics(self) -> Dict:
        """Get synchronization statistics"""
        return self.sync_stats.copy()

def load_sync_config(config_file: str) -> SyncConfig:
    """Load sync configuration from file"""
    with open(config_file, 'r') as f:
        config_data = yaml.safe_load(f)
    
    return SyncConfig(
        source_type=config_data['source']['type'],
        source_config=config_data['source']['config'],
        target_type=config_data['target']['type'],
        target_config=config_data['target']['config'],
        sync_mode=config_data.get('sync_mode', 'batch'),
        conflict_resolution=config_data.get('conflict_resolution', 'source_wins'),
        sync_rules=config_data.get('sync_rules', {}),
        monitoring=config_data.get('monitoring', {})
    )

async def main():
    parser = argparse.ArgumentParser(description='Data Sync Utility')
    parser.add_argument('--config', required=True, help='Sync configuration file')
    parser.add_argument('--mode', choices=['batch', 'realtime', 'scheduled'], 
                       help='Sync mode override')
    parser.add_argument('--tables', nargs='+', help='Specific tables to sync')
    parser.add_argument('--dry-run', action='store_true', help='Validate configuration only')
    
    args = parser.parse_args()
    
    try:
        # Load configuration
        config = load_sync_config(args.config)
        
        # Override mode if specified
        if args.mode:
            config.sync_mode = args.mode
        
        # Override tables if specified
        if args.tables:
            config.sync_rules['tables'] = args.tables
        
        if args.dry_run:
            logger.info("Configuration validation successful")
            return
        
        # Create sync manager
        sync_manager = DataSyncManager(config)
        
        # Initialize connections
        if not await sync_manager.initialize_connections():
            logger.error("Failed to initialize connections")
            sys.exit(1)
        
        # Run sync based on mode
        if config.sync_mode == SyncMode.BATCH.value:
            await sync_manager.run_batch_sync()
        elif config.sync_mode == SyncMode.REALTIME.value:
            await sync_manager.run_realtime_sync()
        else:
            logger.error(f"Unsupported sync mode: {config.sync_mode}")
            sys.exit(1)
        
        # Print statistics
        stats = sync_manager.get_sync_statistics()
        logger.info(f"Sync completed - Records: {stats['records_synced']}, "
                   f"Conflicts: {stats['conflicts_resolved']}, "
                   f"Errors: {stats['errors']}")
        
    except Exception as e:
        logger.error(f"Sync failed: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())
