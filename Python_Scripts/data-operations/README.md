# Data Operations Scripts

This directory contains comprehensive scripts for data backup, migration, ETL processing, and synchronization tasks commonly used in cloud engineering and DevOps workflows.

## 📊 Scripts Overview

### 1. Database Backup (`db_backup.py`)
Comprehensive database backup solution supporting multiple database types:

**Features:**
- ✅ PostgreSQL, MySQL, MongoDB, SQLite support
- ✅ Compression with gzip
- ✅ AWS S3 upload capability
- ✅ Retention policy management
- ✅ Email notifications
- ✅ Backup verification
- ✅ Automated cleanup

**Usage:**
```bash
# Backup PostgreSQL database
python db_backup.py --db-type postgresql --compress --s3-upload

# Backup MongoDB with custom retention
python db_backup.py --db-type mongodb --retention-days 60 --config backup_config.yaml

# Backup SQLite database
python db_backup.py --db-type sqlite --sqlite-file /path/to/database.db
```

### 2. Database Migration (`db_migration.py`)
Cross-database migration tool with schema conversion:

**Features:**
- ✅ Cross-database migrations (PostgreSQL ↔ MySQL ↔ MongoDB ↔ SQLite)
- ✅ Schema conversion and mapping
- ✅ Data transformation during migration
- ✅ Parallel processing for large datasets
- ✅ Progress tracking and validation
- ✅ Incremental migration support

**Usage:**
```bash
# Migrate PostgreSQL to MySQL
python db_migration.py --source postgresql --target mysql --create-schema --validate

# Incremental MongoDB to PostgreSQL migration
python db_migration.py --source mongodb --target postgresql --incremental --batch-size 1000

# Migrate specific tables
python db_migration.py --source postgresql --target sqlite --tables users orders products
```

### 3. ETL Pipeline (`etl_pipeline.py`)
Comprehensive ETL pipeline framework:

**Features:**
- ✅ Multiple data sources (CSV, JSON, XML, databases, APIs, S3, FTP)
- ✅ Data transformation and validation
- ✅ Multiple destinations (databases, files, cloud storage, APIs)
- ✅ Async processing for performance
- ✅ Data quality checks
- ✅ Error handling and rollback

**Usage:**
```bash
# Run ETL pipeline with configuration
python etl_pipeline.py --config etl_pipeline_config.yaml

# Override source and target
python etl_pipeline.py --config pipeline.yaml --source csv --target postgresql

# Dry run to validate configuration
python etl_pipeline.py --config pipeline.yaml --dry-run
```

### 4. Data Sync (`data_sync.py`)
Real-time data synchronization utility:

**Features:**
- ✅ Real-time change detection
- ✅ Bi-directional sync support
- ✅ Conflict resolution strategies
- ✅ File system monitoring
- ✅ Schema drift handling
- ✅ Redis-based change tracking

**Usage:**
```bash
# Real-time sync between databases
python data_sync.py --config sync_config.yaml --mode realtime

# Batch sync specific tables
python data_sync.py --config sync_config.yaml --mode batch --tables users orders

# Validate sync configuration
python data_sync.py --config sync_config.yaml --dry-run
```

## 🔧 Configuration

Each script supports YAML configuration files for complex setups. Example configurations are provided in the `config/` directory:

- `backup_config.yaml` - Database backup settings
- `migration_config.yaml` - Migration rules and transformations
- `etl_pipeline_config.yaml` - ETL pipeline configuration
- `sync_config.yaml` - Data synchronization settings

## 🌐 Environment Variables

Create a `.env` file in the root directory with the following variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database
DB_USER=your_username
DB_PASSWORD=your_password

# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017/
MONGO_DB=your_database

# AWS Configuration
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_DEFAULT_REGION=us-east-1
BACKUP_S3_BUCKET=your-backup-bucket

# Redis Configuration (for data sync)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Email Notifications
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

## 📦 Dependencies

Install required packages:
```bash
pip install -r requirements.txt
```

Key dependencies include:
- `pandas` - Data manipulation
- `sqlalchemy` - Database toolkit
- `boto3` - AWS SDK
- `pymongo` - MongoDB driver
- `psycopg2-binary` - PostgreSQL adapter
- `redis` - Redis client
- `pyyaml` - YAML configuration
- `aiohttp` - Async HTTP client
- `watchdog` - File system monitoring

## 🚀 Quick Start

1. **Set up environment:**
   ```bash
   cp ../.env.example .env
   # Edit .env with your configuration
   ```

2. **Install dependencies:**
   ```bash
   pip install -r ../requirements.txt
   ```

3. **Test database backup:**
   ```bash
   python db_backup.py --db-type postgresql --config config/backup_config.yaml
   ```

4. **Run ETL pipeline:**
   ```bash
   python etl_pipeline.py --config config/etl_pipeline_config.yaml
   ```

## 📋 Common Use Cases

### Database Backup Automation
```bash
# Daily PostgreSQL backup with S3 upload
python db_backup.py --db-type postgresql --compress --s3-upload --retention-days 30

# Schedule with cron (2 AM daily)
0 2 * * * /usr/bin/python3 /path/to/db_backup.py --db-type postgresql --config /path/to/config.yaml
```

### Data Migration Projects
```bash
# Full migration with validation
python db_migration.py --source mysql --target postgresql --create-schema --validate --batch-size 5000

# Incremental sync after initial migration
python db_migration.py --source mysql --target postgresql --incremental --tables users orders
```

### ETL Data Processing
```bash
# Process CSV files to database
python etl_pipeline.py --source csv --target database --config transform_sales_data.yaml

# API to database ETL
python etl_pipeline.py --source api --target postgresql --config api_ingestion.yaml
```

### Real-time Data Sync
```bash
# Set up real-time sync between environments
python data_sync.py --config prod_to_staging_sync.yaml --mode realtime

# Batch sync for data consistency checks
python data_sync.py --config consistency_check.yaml --mode batch
```

## 🔍 Monitoring and Logging

All scripts provide comprehensive logging:
- Console output for real-time monitoring
- Log files for historical analysis
- Email notifications for critical events
- Redis-based metrics (where applicable)

Log files are created in the script directory:
- `backup.log` - Backup operations
- `migration.log` - Migration activities
- `etl_pipeline.log` - ETL processing
- `data_sync.log` - Synchronization events

## 🚨 Error Handling

Scripts include robust error handling:
- Graceful degradation on non-critical failures
- Detailed error logging with context
- Rollback capabilities where applicable
- Email alerts for critical failures
- Retry mechanisms for transient errors

## 🔐 Security Considerations

- Use environment variables for sensitive data
- Implement least privilege database access
- Encrypt backup files when storing in cloud
- Use SSL/TLS for database connections
- Audit trail for all data operations
- Secure credential storage (consider AWS Secrets Manager, HashiCorp Vault)

## 🧪 Testing

Test scripts with sample data before production use:
```bash
# Test with small datasets first
python db_migration.py --source sqlite --target postgresql --tables test_table --batch-size 100

# Validate configurations without executing
python etl_pipeline.py --config test_config.yaml --dry-run
```

## 📈 Performance Optimization

- Use appropriate batch sizes for large datasets
- Enable parallel processing where supported
- Monitor memory usage during large operations
- Use connection pooling for database operations
- Implement incremental processing for large datasets
- Consider using columnar formats (Parquet) for analytics workloads

## 🤝 Contributing

When adding new features:
1. Follow the existing code structure
2. Add comprehensive error handling
3. Include configuration examples
4. Update documentation
5. Add logging for operations
6. Test with various data sources
