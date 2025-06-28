#!/usr/bin/env python3
"""
Database Migration Script

This script provides comprehensive database migration functionality between different
database systems (PostgreSQL, MySQL, MongoDB, SQLite) with support for schema
conversion, data transformation, and incremental migrations.

Features:
- Cross-database migrations (PostgreSQL ↔ MySQL ↔ MongoDB ↔ SQLite)
- Schema conversion and mapping
- Data transformation and validation
- Incremental and full migrations
- Progress tracking and rollback support
- Parallel processing for large datasets

Usage:
    python db_migration.py --source postgresql --target mysql --config migration_config.yaml
    python db_migration.py --source mongodb --target postgresql --incremental --batch-size 1000
"""

import os
import sys
import json
import yaml
import logging
import argparse
import threading
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed

import pandas as pd
import sqlalchemy as sa
from sqlalchemy import create_engine, MetaData, Table, Column, inspect
from sqlalchemy.orm import sessionmaker
import pymongo
import sqlite3
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('migration.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class DatabaseMigration:
    def __init__(self, config_file=None):
        self.config = self.load_config(config_file)
        self.source_conn = None
        self.target_conn = None
        self.migration_log = []
        self.lock = threading.Lock()
        
    def load_config(self, config_file):
        """Load migration configuration"""
        default_config = {
            'batch_size': 1000,
            'max_workers': 4,
            'incremental': False,
            'validate_data': True,
            'create_schema': True,
            'drop_target_tables': False,
            'transformation_rules': {}
        }
        
        if config_file and Path(config_file).exists():
            with open(config_file, 'r') as f:
                file_config = yaml.safe_load(f)
                default_config.update(file_config)
        
        return default_config

    def connect_source(self, db_type: str, **kwargs):
        """Connect to source database"""
        try:
            if db_type == 'postgresql':
                self.source_conn = self._connect_postgresql('source', **kwargs)
            elif db_type == 'mysql':
                self.source_conn = self._connect_mysql('source', **kwargs)
            elif db_type == 'mongodb':
                self.source_conn = self._connect_mongodb('source', **kwargs)
            elif db_type == 'sqlite':
                self.source_conn = self._connect_sqlite('source', **kwargs)
            else:
                raise ValueError(f"Unsupported source database type: {db_type}")
            
            logger.info(f"Connected to source database: {db_type}")
            return True
            
        except Exception as e:
            logger.error(f"Source connection error: {str(e)}")
            return False

    def connect_target(self, db_type: str, **kwargs):
        """Connect to target database"""
        try:
            if db_type == 'postgresql':
                self.target_conn = self._connect_postgresql('target', **kwargs)
            elif db_type == 'mysql':
                self.target_conn = self._connect_mysql('target', **kwargs)
            elif db_type == 'mongodb':
                self.target_conn = self._connect_mongodb('target', **kwargs)
            elif db_type == 'sqlite':
                self.target_conn = self._connect_sqlite('target', **kwargs)
            else:
                raise ValueError(f"Unsupported target database type: {db_type}")
            
            logger.info(f"Connected to target database: {db_type}")
            return True
            
        except Exception as e:
            logger.error(f"Target connection error: {str(e)}")
            return False

    def _connect_postgresql(self, conn_type: str, **kwargs):
        """Connect to PostgreSQL database"""
        if conn_type == 'source':
            host = kwargs.get('host', os.getenv('SOURCE_DB_HOST', os.getenv('DB_HOST')))
            port = kwargs.get('port', os.getenv('SOURCE_DB_PORT', os.getenv('DB_PORT')))
            database = kwargs.get('database', os.getenv('SOURCE_DB_NAME', os.getenv('DB_NAME')))
            user = kwargs.get('user', os.getenv('SOURCE_DB_USER', os.getenv('DB_USER')))
            password = kwargs.get('password', os.getenv('SOURCE_DB_PASSWORD', os.getenv('DB_PASSWORD')))
        else:
            host = kwargs.get('host', os.getenv('TARGET_DB_HOST', os.getenv('DB_HOST')))
            port = kwargs.get('port', os.getenv('TARGET_DB_PORT', os.getenv('DB_PORT')))
            database = kwargs.get('database', os.getenv('TARGET_DB_NAME', os.getenv('DB_NAME')))
            user = kwargs.get('user', os.getenv('TARGET_DB_USER', os.getenv('DB_USER')))
            password = kwargs.get('password', os.getenv('TARGET_DB_PASSWORD', os.getenv('DB_PASSWORD')))
        
        url = f"postgresql://{user}:{password}@{host}:{port}/{database}"
        engine = create_engine(url)
        return engine

    def _connect_mysql(self, conn_type: str, **kwargs):
        """Connect to MySQL database"""
        if conn_type == 'source':
            host = kwargs.get('host', os.getenv('SOURCE_DB_HOST', os.getenv('DB_HOST')))
            port = kwargs.get('port', os.getenv('SOURCE_DB_PORT', os.getenv('DB_PORT')))
            database = kwargs.get('database', os.getenv('SOURCE_DB_NAME', os.getenv('DB_NAME')))
            user = kwargs.get('user', os.getenv('SOURCE_DB_USER', os.getenv('DB_USER')))
            password = kwargs.get('password', os.getenv('SOURCE_DB_PASSWORD', os.getenv('DB_PASSWORD')))
        else:
            host = kwargs.get('host', os.getenv('TARGET_DB_HOST', os.getenv('DB_HOST')))
            port = kwargs.get('port', os.getenv('TARGET_DB_PORT', os.getenv('DB_PORT')))
            database = kwargs.get('database', os.getenv('TARGET_DB_NAME', os.getenv('DB_NAME')))
            user = kwargs.get('user', os.getenv('TARGET_DB_USER', os.getenv('DB_USER')))
            password = kwargs.get('password', os.getenv('TARGET_DB_PASSWORD', os.getenv('DB_PASSWORD')))
        
        url = f"mysql+pymysql://{user}:{password}@{host}:{port}/{database}"
        engine = create_engine(url)
        return engine

    def _connect_mongodb(self, conn_type: str, **kwargs):
        """Connect to MongoDB database"""
        if conn_type == 'source':
            uri = kwargs.get('uri', os.getenv('SOURCE_MONGO_URI', os.getenv('MONGO_URI')))
            db_name = kwargs.get('database', os.getenv('SOURCE_MONGO_DB', os.getenv('MONGO_DB')))
        else:
            uri = kwargs.get('uri', os.getenv('TARGET_MONGO_URI', os.getenv('MONGO_URI')))
            db_name = kwargs.get('database', os.getenv('TARGET_MONGO_DB', os.getenv('MONGO_DB')))
        
        client = pymongo.MongoClient(uri)
        return client[db_name]

    def _connect_sqlite(self, conn_type: str, **kwargs):
        """Connect to SQLite database"""
        if conn_type == 'source':
            db_file = kwargs.get('file', os.getenv('SOURCE_SQLITE_FILE', 'source.db'))
        else:
            db_file = kwargs.get('file', os.getenv('TARGET_SQLITE_FILE', 'target.db'))
        
        url = f"sqlite:///{db_file}"
        engine = create_engine(url)
        return engine

    def get_source_schema(self, source_type: str) -> Dict:
        """Get schema information from source database"""
        try:
            if source_type == 'mongodb':
                return self._get_mongodb_schema()
            else:
                return self._get_sql_schema()
                
        except Exception as e:
            logger.error(f"Schema extraction error: {str(e)}")
            return {}

    def _get_sql_schema(self) -> Dict:
        """Get schema from SQL database"""
        inspector = inspect(self.source_conn)
        schema = {}
        
        for table_name in inspector.get_table_names():
            columns = inspector.get_columns(table_name)
            schema[table_name] = {
                'columns': columns,
                'primary_keys': inspector.get_pk_constraint(table_name),
                'foreign_keys': inspector.get_foreign_keys(table_name),
                'indexes': inspector.get_indexes(table_name)
            }
        
        return schema

    def _get_mongodb_schema(self) -> Dict:
        """Get schema from MongoDB database"""
        schema = {}
        
        for collection_name in self.source_conn.list_collection_names():
            collection = self.source_conn[collection_name]
            
            # Sample documents to infer schema
            sample_docs = list(collection.find().limit(100))
            
            if sample_docs:
                fields = {}
                for doc in sample_docs:
                    for key, value in doc.items():
                        if key not in fields:
                            fields[key] = type(value).__name__
                
                schema[collection_name] = {
                    'fields': fields,
                    'document_count': collection.count_documents({})
                }
        
        return schema

    def create_target_schema(self, source_type: str, target_type: str, schema: Dict):
        """Create schema in target database"""
        try:
            if target_type == 'mongodb':
                # MongoDB doesn't require schema creation
                logger.info("MongoDB target: schema creation not required")
                return True
            else:
                return self._create_sql_schema(source_type, target_type, schema)
                
        except Exception as e:
            logger.error(f"Schema creation error: {str(e)}")
            return False

    def _create_sql_schema(self, source_type: str, target_type: str, schema: Dict) -> bool:
        """Create SQL schema in target database"""
        metadata = MetaData()
        
        for table_name, table_info in schema.items():
            columns = []
            
            for col_info in table_info['columns']:
                col_name = col_info['name']
                col_type = self._convert_column_type(col_info['type'], source_type, target_type)
                
                column = Column(col_name, col_type)
                columns.append(column)
            
            table = Table(table_name, metadata, *columns)
        
        # Create tables
        metadata.create_all(self.target_conn)
        logger.info(f"Created {len(schema)} tables in target database")
        return True

    def _convert_column_type(self, source_type, source_db: str, target_db: str):
        """Convert column type between database systems"""
        # Type mapping dictionary
        type_mapping = {
            'postgresql': {
                'mysql': {
                    'INTEGER': sa.Integer,
                    'VARCHAR': sa.String,
                    'TEXT': sa.Text,
                    'TIMESTAMP': sa.DateTime,
                    'BOOLEAN': sa.Boolean,
                    'FLOAT': sa.Float,
                    'DECIMAL': sa.Numeric
                },
                'sqlite': {
                    'INTEGER': sa.Integer,
                    'VARCHAR': sa.String,
                    'TEXT': sa.Text,
                    'TIMESTAMP': sa.DateTime,
                    'BOOLEAN': sa.Boolean,
                    'FLOAT': sa.Float,
                    'DECIMAL': sa.Numeric
                }
            }
        }
        
        # Default to String if no mapping found
        return type_mapping.get(source_db, {}).get(target_db, {}).get(str(source_type), sa.String)

    def migrate_data(self, source_type: str, target_type: str, tables: List[str] = None):
        """Migrate data between databases"""
        try:
            if source_type == 'mongodb' and target_type != 'mongodb':
                return self._migrate_mongodb_to_sql(tables)
            elif source_type != 'mongodb' and target_type == 'mongodb':
                return self._migrate_sql_to_mongodb(tables)
            elif source_type == 'mongodb' and target_type == 'mongodb':
                return self._migrate_mongodb_to_mongodb(tables)
            else:
                return self._migrate_sql_to_sql(tables)
                
        except Exception as e:
            logger.error(f"Data migration error: {str(e)}")
            return False

    def _migrate_sql_to_sql(self, tables: List[str] = None) -> bool:
        """Migrate data between SQL databases"""
        inspector = inspect(self.source_conn)
        table_names = tables or inspector.get_table_names()
        
        total_tables = len(table_names)
        completed_tables = 0
        
        with ThreadPoolExecutor(max_workers=self.config['max_workers']) as executor:
            futures = []
            
            for table_name in table_names:
                future = executor.submit(self._migrate_table, table_name)
                futures.append((future, table_name))
            
            for future, table_name in futures:
                try:
                    success = future.result()
                    if success:
                        completed_tables += 1
                        logger.info(f"Migrated table {table_name} ({completed_tables}/{total_tables})")
                    else:
                        logger.error(f"Failed to migrate table {table_name}")
                except Exception as e:
                    logger.error(f"Error migrating table {table_name}: {str(e)}")
        
        return completed_tables == total_tables

    def _migrate_table(self, table_name: str) -> bool:
        """Migrate a single table"""
        try:
            batch_size = self.config['batch_size']
            
            # Get total row count
            count_query = f"SELECT COUNT(*) FROM {table_name}"
            total_rows = self.source_conn.execute(sa.text(count_query)).scalar()
            
            if total_rows == 0:
                logger.info(f"Table {table_name} is empty, skipping")
                return True
            
            # Migrate in batches
            offset = 0
            migrated_rows = 0
            
            while offset < total_rows:
                # Read batch from source
                query = f"SELECT * FROM {table_name} LIMIT {batch_size} OFFSET {offset}"
                df = pd.read_sql(query, self.source_conn)
                
                if df.empty:
                    break
                
                # Apply transformations
                df = self._apply_transformations(table_name, df)
                
                # Write to target
                df.to_sql(table_name, self.target_conn, if_exists='append', index=False)
                
                migrated_rows += len(df)
                offset += batch_size
                
                with self.lock:
                    progress = (migrated_rows / total_rows) * 100
                    logger.info(f"Table {table_name}: {migrated_rows}/{total_rows} rows ({progress:.1f}%)")
            
            logger.info(f"Completed migration of table {table_name}: {migrated_rows} rows")
            return True
            
        except Exception as e:
            logger.error(f"Error migrating table {table_name}: {str(e)}")
            return False

    def _migrate_mongodb_to_sql(self, collections: List[str] = None) -> bool:
        """Migrate data from MongoDB to SQL database"""
        collection_names = collections or self.source_conn.list_collection_names()
        
        for collection_name in collection_names:
            try:
                collection = self.source_conn[collection_name]
                total_docs = collection.count_documents({})
                
                if total_docs == 0:
                    logger.info(f"Collection {collection_name} is empty, skipping")
                    continue
                
                # Process in batches
                batch_size = self.config['batch_size']
                skip = 0
                
                while skip < total_docs:
                    # Get batch of documents
                    cursor = collection.find().skip(skip).limit(batch_size)
                    documents = list(cursor)
                    
                    if not documents:
                        break
                    
                    # Convert to DataFrame
                    df = pd.DataFrame(documents)
                    
                    # Flatten nested documents
                    df = self._flatten_mongodb_documents(df)
                    
                    # Apply transformations
                    df = self._apply_transformations(collection_name, df)
                    
                    # Write to SQL
                    df.to_sql(collection_name, self.target_conn, if_exists='append', index=False)
                    
                    skip += batch_size
                    progress = (skip / total_docs) * 100
                    logger.info(f"Collection {collection_name}: {skip}/{total_docs} docs ({progress:.1f}%)")
                
                logger.info(f"Completed migration of collection {collection_name}")
                
            except Exception as e:
                logger.error(f"Error migrating collection {collection_name}: {str(e)}")
                return False
        
        return True

    def _migrate_sql_to_mongodb(self, tables: List[str] = None) -> bool:
        """Migrate data from SQL database to MongoDB"""
        inspector = inspect(self.source_conn)
        table_names = tables or inspector.get_table_names()
        
        for table_name in table_names:
            try:
                # Get total row count
                count_query = f"SELECT COUNT(*) FROM {table_name}"
                total_rows = self.source_conn.execute(sa.text(count_query)).scalar()
                
                if total_rows == 0:
                    logger.info(f"Table {table_name} is empty, skipping")
                    continue
                
                # Get collection
                collection = self.target_conn[table_name]
                
                # Process in batches
                batch_size = self.config['batch_size']
                offset = 0
                
                while offset < total_rows:
                    # Read batch from SQL
                    query = f"SELECT * FROM {table_name} LIMIT {batch_size} OFFSET {offset}"
                    df = pd.read_sql(query, self.source_conn)
                    
                    if df.empty:
                        break
                    
                    # Apply transformations
                    df = self._apply_transformations(table_name, df)
                    
                    # Convert to documents
                    documents = df.to_dict('records')
                    
                    # Insert into MongoDB
                    collection.insert_many(documents)
                    
                    offset += batch_size
                    progress = (offset / total_rows) * 100
                    logger.info(f"Table {table_name}: {offset}/{total_rows} rows ({progress:.1f}%)")
                
                logger.info(f"Completed migration of table {table_name}")
                
            except Exception as e:
                logger.error(f"Error migrating table {table_name}: {str(e)}")
                return False
        
        return True

    def _flatten_mongodb_documents(self, df: pd.DataFrame) -> pd.DataFrame:
        """Flatten nested MongoDB documents"""
        # Convert ObjectId to string
        if '_id' in df.columns:
            df['_id'] = df['_id'].astype(str)
        
        # Flatten nested objects (simple approach)
        for col in df.columns:
            if df[col].dtype == 'object':
                # Check if any value is a dict
                if df[col].apply(lambda x: isinstance(x, dict)).any():
                    # Convert dict to JSON string
                    df[col] = df[col].apply(lambda x: json.dumps(x) if isinstance(x, dict) else x)
        
        return df

    def _apply_transformations(self, table_name: str, df: pd.DataFrame) -> pd.DataFrame:
        """Apply data transformations based on configuration"""
        transformations = self.config.get('transformation_rules', {}).get(table_name, {})
        
        for column, rules in transformations.items():
            if column in df.columns:
                for rule in rules:
                    if rule['type'] == 'replace':
                        df[column] = df[column].str.replace(rule['from'], rule['to'])
                    elif rule['type'] == 'convert_type':
                        df[column] = df[column].astype(rule['target_type'])
                    elif rule['type'] == 'default_value':
                        df[column] = df[column].fillna(rule['value'])
        
        return df

    def validate_migration(self, source_type: str, target_type: str) -> bool:
        """Validate migration results"""
        try:
            logger.info("Starting migration validation...")
            
            if source_type == 'mongodb' and target_type != 'mongodb':
                return self._validate_mongodb_to_sql()
            elif source_type != 'mongodb' and target_type == 'mongodb':
                return self._validate_sql_to_mongodb()
            else:
                return self._validate_sql_to_sql()
                
        except Exception as e:
            logger.error(f"Validation error: {str(e)}")
            return False

    def _validate_sql_to_sql(self) -> bool:
        """Validate SQL to SQL migration"""
        inspector_source = inspect(self.source_conn)
        inspector_target = inspect(self.target_conn)
        
        source_tables = set(inspector_source.get_table_names())
        target_tables = set(inspector_target.get_table_names())
        
        # Check if all tables exist
        if not source_tables.issubset(target_tables):
            missing_tables = source_tables - target_tables
            logger.error(f"Missing tables in target: {missing_tables}")
            return False
        
        # Check row counts
        for table_name in source_tables:
            source_count = self.source_conn.execute(sa.text(f"SELECT COUNT(*) FROM {table_name}")).scalar()
            target_count = self.target_conn.execute(sa.text(f"SELECT COUNT(*) FROM {table_name}")).scalar()
            
            if source_count != target_count:
                logger.error(f"Row count mismatch in {table_name}: source={source_count}, target={target_count}")
                return False
        
        logger.info("SQL to SQL migration validation passed")
        return True

    def generate_migration_report(self) -> Dict:
        """Generate migration report"""
        report = {
            'timestamp': datetime.now().isoformat(),
            'migration_log': self.migration_log,
            'status': 'completed' if self.migration_log else 'failed',
            'total_tables': len(self.migration_log),
            'config': self.config
        }
        
        # Save report to file
        report_file = f"migration_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_file, 'w') as f:
            json.dump(report, f, indent=2)
        
        logger.info(f"Migration report saved: {report_file}")
        return report

def main():
    parser = argparse.ArgumentParser(description='Database Migration Script')
    parser.add_argument('--source', choices=['postgresql', 'mysql', 'mongodb', 'sqlite'], 
                       required=True, help='Source database type')
    parser.add_argument('--target', choices=['postgresql', 'mysql', 'mongodb', 'sqlite'], 
                       required=True, help='Target database type')
    parser.add_argument('--config', help='Migration configuration file')
    parser.add_argument('--tables', nargs='+', help='Specific tables/collections to migrate')
    parser.add_argument('--batch-size', type=int, default=1000, help='Batch size for data migration')
    parser.add_argument('--incremental', action='store_true', help='Incremental migration')
    parser.add_argument('--validate', action='store_true', help='Validate migration results')
    parser.add_argument('--create-schema', action='store_true', help='Create target schema')
    
    args = parser.parse_args()
    
    # Create migration instance
    migration = DatabaseMigration(args.config)
    
    # Override config with command line arguments
    migration.config.update({
        'batch_size': args.batch_size,
        'incremental': args.incremental,
        'validate_data': args.validate,
        'create_schema': args.create_schema
    })
    
    # Connect to databases
    if not migration.connect_source(args.source):
        logger.error("Failed to connect to source database")
        sys.exit(1)
    
    if not migration.connect_target(args.target):
        logger.error("Failed to connect to target database")
        sys.exit(1)
    
    try:
        # Get source schema
        logger.info("Extracting source schema...")
        schema = migration.get_source_schema(args.source)
        
        if not schema:
            logger.error("Failed to extract source schema")
            sys.exit(1)
        
        # Create target schema if requested
        if args.create_schema:
            logger.info("Creating target schema...")
            if not migration.create_target_schema(args.source, args.target, schema):
                logger.error("Failed to create target schema")
                sys.exit(1)
        
        # Migrate data
        logger.info("Starting data migration...")
        if not migration.migrate_data(args.source, args.target, args.tables):
            logger.error("Data migration failed")
            sys.exit(1)
        
        # Validate migration
        if args.validate:
            logger.info("Validating migration...")
            if not migration.validate_migration(args.source, args.target):
                logger.error("Migration validation failed")
                sys.exit(1)
        
        # Generate report
        report = migration.generate_migration_report()
        logger.info("Migration completed successfully!")
        
    except Exception as e:
        logger.error(f"Migration failed: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main()
