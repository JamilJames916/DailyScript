#!/usr/bin/env python3
"""
ETL Pipeline Script

This script provides a comprehensive ETL (Extract, Transform, Load) pipeline
framework that supports multiple data sources and destinations with configurable
transformation rules, data validation, and error handling.

Features:
- Multiple data sources: CSV, JSON, XML, databases, APIs, cloud storage
- Data transformation: cleaning, validation, type conversion, aggregation
- Multiple destinations: databases, files, cloud storage, APIs
- Pipeline orchestration and scheduling
- Data quality checks and error handling
- Parallel processing and performance optimization

Usage:
    python etl_pipeline.py --config pipeline_config.yaml
    python etl_pipeline.py --source csv --target postgresql --transform-config transforms.yaml
"""

import os
import sys
import json
import yaml
import logging
import argparse
import asyncio
import tempfile
import aiofiles
import aiohttp
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any, Optional, Union
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor
from dataclasses import dataclass
from enum import Enum

import pandas as pd
import numpy as np
import sqlalchemy as sa
from sqlalchemy import create_engine
import requests
import boto3
import paramiko
from pymongo import MongoClient
import redis
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('etl_pipeline.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class DataSourceType(Enum):
    CSV = "csv"
    JSON = "json"
    XML = "xml"
    DATABASE = "database"
    API = "api"
    S3 = "s3"
    FTP = "ftp"
    KAFKA = "kafka"

class DataDestinationType(Enum):
    CSV = "csv"
    JSON = "json"
    DATABASE = "database"
    S3 = "s3"
    API = "api"
    ELASTICSEARCH = "elasticsearch"

@dataclass
class ETLConfig:
    source_type: str
    source_config: Dict
    destination_type: str
    destination_config: Dict
    transformation_config: Dict
    pipeline_config: Dict
    validation_config: Dict

class DataExtractor:
    """Data extraction from various sources"""
    
    def __init__(self, config: Dict):
        self.config = config
        
    async def extract(self, source_type: str, source_config: Dict) -> pd.DataFrame:
        """Extract data from specified source"""
        try:
            if source_type == DataSourceType.CSV.value:
                return await self._extract_csv(source_config)
            elif source_type == DataSourceType.JSON.value:
                return await self._extract_json(source_config)
            elif source_type == DataSourceType.XML.value:
                return await self._extract_xml(source_config)
            elif source_type == DataSourceType.DATABASE.value:
                return await self._extract_database(source_config)
            elif source_type == DataSourceType.API.value:
                return await self._extract_api(source_config)
            elif source_type == DataSourceType.S3.value:
                return await self._extract_s3(source_config)
            elif source_type == DataSourceType.FTP.value:
                return await self._extract_ftp(source_config)
            else:
                raise ValueError(f"Unsupported source type: {source_type}")
                
        except Exception as e:
            logger.error(f"Extraction error: {str(e)}")
            raise

    async def _extract_csv(self, config: Dict) -> pd.DataFrame:
        """Extract data from CSV file"""
        file_path = config['file_path']
        encoding = config.get('encoding', 'utf-8')
        delimiter = config.get('delimiter', ',')
        
        logger.info(f"Extracting CSV data from: {file_path}")
        
        # Handle remote CSV files
        if file_path.startswith(('http://', 'https://')):
            async with aiohttp.ClientSession() as session:
                async with session.get(file_path) as response:
                    content = await response.text()
                    from io import StringIO
                    df = pd.read_csv(StringIO(content), encoding=encoding, delimiter=delimiter)
        else:
            df = pd.read_csv(file_path, encoding=encoding, delimiter=delimiter)
        
        logger.info(f"Extracted {len(df)} rows from CSV")
        return df

    async def _extract_json(self, config: Dict) -> pd.DataFrame:
        """Extract data from JSON file or URL"""
        source = config['source']
        json_path = config.get('json_path', None)  # JSONPath for nested data
        
        logger.info(f"Extracting JSON data from: {source}")
        
        if source.startswith(('http://', 'https://')):
            async with aiohttp.ClientSession() as session:
                async with session.get(source) as response:
                    data = await response.json()
        else:
            async with aiofiles.open(source, 'r') as f:
                content = await f.read()
                data = json.loads(content)
        
        # Handle nested JSON data
        if json_path:
            # Simple JSONPath implementation
            for key in json_path.split('.'):
                if key.isdigit():
                    data = data[int(key)]
                else:
                    data = data[key]
        
        df = pd.json_normalize(data)
        logger.info(f"Extracted {len(df)} rows from JSON")
        return df

    async def _extract_xml(self, config: Dict) -> pd.DataFrame:
        """Extract data from XML file"""
        file_path = config['file_path']
        xpath = config.get('xpath', None)
        
        logger.info(f"Extracting XML data from: {file_path}")
        
        import xml.etree.ElementTree as ET
        
        tree = ET.parse(file_path)
        root = tree.getroot()
        
        # Convert XML to list of dictionaries
        records = []
        elements = root.findall(xpath) if xpath else [root]
        
        for element in elements:
            record = {}
            for child in element:
                record[child.tag] = child.text
            records.append(record)
        
        df = pd.DataFrame(records)
        logger.info(f"Extracted {len(df)} rows from XML")
        return df

    async def _extract_database(self, config: Dict) -> pd.DataFrame:
        """Extract data from database"""
        connection_string = config['connection_string']
        query = config['query']
        
        logger.info("Extracting data from database")
        
        engine = create_engine(connection_string)
        try:
            df = pd.read_sql(query, engine)
            logger.info(f"Extracted {len(df)} rows from database")
            return df
        finally:
            engine.dispose()

    async def _extract_api(self, config: Dict) -> pd.DataFrame:
        """Extract data from REST API"""
        url = config['url']
        method = config.get('method', 'GET')
        headers = config.get('headers', {})
        params = config.get('params', {})
        data_path = config.get('data_path', None)
        
        logger.info(f"Extracting data from API: {url}")
        
        async with aiohttp.ClientSession() as session:
            if method.upper() == 'GET':
                async with session.get(url, headers=headers, params=params) as response:
                    data = await response.json()
            elif method.upper() == 'POST':
                async with session.post(url, headers=headers, json=params) as response:
                    data = await response.json()
        
        # Extract data from nested response
        if data_path:
            for key in data_path.split('.'):
                data = data[key]
        
        df = pd.json_normalize(data)
        logger.info(f"Extracted {len(df)} rows from API")
        return df

    async def _extract_s3(self, config: Dict) -> pd.DataFrame:
        """Extract data from AWS S3"""
        bucket = config['bucket']
        key = config['key']
        file_format = config.get('format', 'csv')
        
        logger.info(f"Extracting data from S3: s3://{bucket}/{key}")
        
        s3_client = boto3.client('s3')
        
        # Download file to temporary location
        temp_file = os.path.join(tempfile.gettempdir(), Path(key).name)
        s3_client.download_file(bucket, key, temp_file)
        
        # Read based on format
        if file_format == 'csv':
            df = pd.read_csv(temp_file)
        elif file_format == 'json':
            df = pd.read_json(temp_file)
        elif file_format == 'parquet':
            df = pd.read_parquet(temp_file)
        else:
            raise ValueError(f"Unsupported S3 file format: {file_format}")
        
        # Clean up temporary file
        os.remove(temp_file)
        
        logger.info(f"Extracted {len(df)} rows from S3")
        return df

    async def _extract_ftp(self, config: Dict) -> pd.DataFrame:
        """Extract data from FTP server"""
        host = config['host']
        username = config['username']
        password = config['password']
        remote_path = config['remote_path']
        file_format = config.get('format', 'csv')
        
        logger.info(f"Extracting data from FTP: {host}/{remote_path}")
        
        # Connect to FTP server
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ssh.connect(host, username=username, password=password)
        
        sftp = ssh.open_sftp()
        
        # Download file
        temp_file = os.path.join(tempfile.gettempdir(), Path(remote_path).name)
        sftp.get(remote_path, temp_file)
        
        # Read based on format
        if file_format == 'csv':
            df = pd.read_csv(temp_file)
        elif file_format == 'json':
            df = pd.read_json(temp_file)
        else:
            raise ValueError(f"Unsupported FTP file format: {file_format}")
        
        # Clean up
        sftp.close()
        ssh.close()
        os.remove(temp_file)
        
        logger.info(f"Extracted {len(df)} rows from FTP")
        return df

class DataTransformer:
    """Data transformation and cleaning"""
    
    def __init__(self, config: Dict):
        self.config = config
        
    def transform(self, df: pd.DataFrame, transformation_config: Dict) -> pd.DataFrame:
        """Apply transformations to data"""
        try:
            logger.info("Starting data transformation")
            
            # Apply transformations in order
            transformations = transformation_config.get('transformations', [])
            
            for transform in transformations:
                transform_type = transform['type']
                
                if transform_type == 'drop_columns':
                    df = self._drop_columns(df, transform)
                elif transform_type == 'rename_columns':
                    df = self._rename_columns(df, transform)
                elif transform_type == 'filter_rows':
                    df = self._filter_rows(df, transform)
                elif transform_type == 'convert_types':
                    df = self._convert_types(df, transform)
                elif transform_type == 'clean_data':
                    df = self._clean_data(df, transform)
                elif transform_type == 'aggregate':
                    df = self._aggregate_data(df, transform)
                elif transform_type == 'join':
                    df = self._join_data(df, transform)
                elif transform_type == 'pivot':
                    df = self._pivot_data(df, transform)
                elif transform_type == 'custom':
                    df = self._apply_custom_function(df, transform)
                else:
                    logger.warning(f"Unknown transformation type: {transform_type}")
            
            logger.info(f"Transformation completed: {len(df)} rows")
            return df
            
        except Exception as e:
            logger.error(f"Transformation error: {str(e)}")
            raise

    def _drop_columns(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Drop specified columns"""
        columns = config['columns']
        return df.drop(columns=columns, errors='ignore')

    def _rename_columns(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Rename columns"""
        mapping = config['mapping']
        return df.rename(columns=mapping)

    def _filter_rows(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Filter rows based on conditions"""
        conditions = config['conditions']
        
        for condition in conditions:
            column = condition['column']
            operator = condition['operator']
            value = condition['value']
            
            if operator == '==':
                df = df[df[column] == value]
            elif operator == '!=':
                df = df[df[column] != value]
            elif operator == '>':
                df = df[df[column] > value]
            elif operator == '<':
                df = df[df[column] < value]
            elif operator == '>=':
                df = df[df[column] >= value]
            elif operator == '<=':
                df = df[df[column] <= value]
            elif operator == 'in':
                df = df[df[column].isin(value)]
            elif operator == 'not_in':
                df = df[~df[column].isin(value)]
        
        return df

    def _convert_types(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Convert column data types"""
        type_mapping = config['mapping']
        
        for column, dtype in type_mapping.items():
            if column in df.columns:
                try:
                    if dtype == 'datetime':
                        df[column] = pd.to_datetime(df[column])
                    elif dtype == 'numeric':
                        df[column] = pd.to_numeric(df[column], errors='coerce')
                    else:
                        df[column] = df[column].astype(dtype)
                except Exception as e:
                    logger.warning(f"Could not convert {column} to {dtype}: {str(e)}")
        
        return df

    def _clean_data(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Clean data (remove duplicates, handle nulls, etc.)"""
        operations = config['operations']
        
        if 'remove_duplicates' in operations:
            df = df.drop_duplicates()
        
        if 'handle_nulls' in operations:
            null_strategy = operations['handle_nulls']
            if null_strategy == 'drop':
                df = df.dropna()
            elif null_strategy == 'fill':
                fill_value = operations.get('fill_value', 0)
                df = df.fillna(fill_value)
        
        if 'trim_strings' in operations:
            string_columns = df.select_dtypes(include=['object']).columns
            for col in string_columns:
                # Only apply to non-null values to avoid converting NaN to 'nan'
                mask = df[col].notna()
                df.loc[mask, col] = df.loc[mask, col].astype(str).str.strip()
        
        return df

    def _aggregate_data(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Aggregate data"""
        group_by = config['group_by']
        aggregations = config['aggregations']
        
        return df.groupby(group_by).agg(aggregations).reset_index()

    def _join_data(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Join with another dataset"""
        # This would require loading the second dataset
        # Implementation depends on specific requirements
        logger.info("Join transformation not implemented in this example")
        return df

    def _pivot_data(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Pivot data"""
        index = config['index']
        columns = config['columns']
        values = config['values']
        
        return df.pivot_table(index=index, columns=columns, values=values).reset_index()

    def _apply_custom_function(self, df: pd.DataFrame, config: Dict) -> pd.DataFrame:
        """Apply custom transformation function"""
        function_code = config['function']
        
        # Execute custom function (be careful with security!)
        local_vars = {'df': df, 'pd': pd, 'np': np}
        exec(function_code, globals(), local_vars)
        
        return local_vars['df']

class DataValidator:
    """Data quality validation"""
    
    def __init__(self, config: Dict):
        self.config = config
        
    def validate(self, df: pd.DataFrame, validation_config: Dict) -> Dict:
        """Validate data quality"""
        results = {
            'passed': True,
            'checks': [],
            'errors': []
        }
        
        try:
            checks = validation_config.get('checks', [])
            
            for check in checks:
                check_type = check['type']
                check_result = True
                
                if check_type == 'not_null':
                    check_result = self._check_not_null(df, check)
                elif check_type == 'unique':
                    check_result = self._check_unique(df, check)
                elif check_type == 'range':
                    check_result = self._check_range(df, check)
                elif check_type == 'format':
                    check_result = self._check_format(df, check)
                elif check_type == 'custom':
                    check_result = self._check_custom(df, check)
                
                results['checks'].append({
                    'type': check_type,
                    'passed': check_result,
                    'config': check
                })
                
                if not check_result:
                    results['passed'] = False
                    results['errors'].append(f"Validation failed: {check}")
            
            logger.info(f"Data validation completed: {len(results['checks'])} checks")
            return results
            
        except Exception as e:
            logger.error(f"Validation error: {str(e)}")
            results['passed'] = False
            results['errors'].append(str(e))
            return results

    def _check_not_null(self, df: pd.DataFrame, config: Dict) -> bool:
        """Check for null values"""
        columns = config['columns']
        for column in columns:
            if df[column].isnull().any():
                logger.error(f"Null values found in column: {column}")
                return False
        return True

    def _check_unique(self, df: pd.DataFrame, config: Dict) -> bool:
        """Check for unique values"""
        columns = config['columns']
        if df[columns].duplicated().any():
            logger.error(f"Duplicate values found in columns: {columns}")
            return False
        return True

    def _check_range(self, df: pd.DataFrame, config: Dict) -> bool:
        """Check value ranges"""
        column = config['column']
        min_val = config.get('min')
        max_val = config.get('max')
        
        if min_val is not None and (df[column] < min_val).any():
            logger.error(f"Values below minimum in column: {column}")
            return False
        
        if max_val is not None and (df[column] > max_val).any():
            logger.error(f"Values above maximum in column: {column}")
            return False
        
        return True

    def _check_format(self, df: pd.DataFrame, config: Dict) -> bool:
        """Check data format using regex"""
        column = config['column']
        pattern = config['pattern']
        
        # Convert to string first to handle non-string values
        if not df[column].astype(str).str.match(pattern).all():
            logger.error(f"Format validation failed for column: {column}")
            return False
        
        return True

    def _check_custom(self, df: pd.DataFrame, config: Dict) -> bool:
        """Custom validation check"""
        function_code = config['function']
        
        local_vars = {'df': df, 'pd': pd, 'np': np, 'result': True}
        exec(function_code, globals(), local_vars)
        
        return local_vars['result']

class DataLoader:
    """Data loading to various destinations"""
    
    def __init__(self, config: Dict):
        self.config = config
        
    async def load(self, df: pd.DataFrame, destination_type: str, destination_config: Dict) -> bool:
        """Load data to specified destination"""
        try:
            if destination_type == DataDestinationType.CSV.value:
                return await self._load_csv(df, destination_config)
            elif destination_type == DataDestinationType.JSON.value:
                return await self._load_json(df, destination_config)
            elif destination_type == DataDestinationType.DATABASE.value:
                return await self._load_database(df, destination_config)
            elif destination_type == DataDestinationType.S3.value:
                return await self._load_s3(df, destination_config)
            elif destination_type == DataDestinationType.API.value:
                return await self._load_api(df, destination_config)
            else:
                raise ValueError(f"Unsupported destination type: {destination_type}")
                
        except Exception as e:
            logger.error(f"Loading error: {str(e)}")
            return False

    async def _load_csv(self, df: pd.DataFrame, config: Dict) -> bool:
        """Load data to CSV file"""
        file_path = config['file_path']
        encoding = config.get('encoding', 'utf-8')
        index = config.get('include_index', False)
        
        logger.info(f"Loading {len(df)} rows to CSV: {file_path}")
        
        # Create directory if it doesn't exist
        Path(file_path).parent.mkdir(parents=True, exist_ok=True)
        
        df.to_csv(file_path, encoding=encoding, index=index)
        logger.info("CSV loading completed")
        return True

    async def _load_json(self, df: pd.DataFrame, config: Dict) -> bool:
        """Load data to JSON file"""
        file_path = config['file_path']
        orient = config.get('orient', 'records')
        
        logger.info(f"Loading {len(df)} rows to JSON: {file_path}")
        
        # Create directory if it doesn't exist
        Path(file_path).parent.mkdir(parents=True, exist_ok=True)
        
        df.to_json(file_path, orient=orient, indent=2)
        logger.info("JSON loading completed")
        return True

    async def _load_database(self, df: pd.DataFrame, config: Dict) -> bool:
        """Load data to database"""
        connection_string = config['connection_string']
        table_name = config['table_name']
        if_exists = config.get('if_exists', 'append')
        
        logger.info(f"Loading {len(df)} rows to database table: {table_name}")
        
        engine = create_engine(connection_string)
        try:
            df.to_sql(table_name, engine, if_exists=if_exists, index=False)
            logger.info("Database loading completed")
            return True
        finally:
            engine.dispose()

    async def _load_s3(self, df: pd.DataFrame, config: Dict) -> bool:
        """Load data to S3"""
        bucket = config['bucket']
        key = config['key']
        file_format = config.get('format', 'csv')
        
        logger.info(f"Loading {len(df)} rows to S3: s3://{bucket}/{key}")
        
        # Save to temporary file
        temp_file = os.path.join(tempfile.gettempdir(), Path(key).name)
        
        if file_format == 'csv':
            df.to_csv(temp_file, index=False)
        elif file_format == 'json':
            df.to_json(temp_file, orient='records')
        elif file_format == 'parquet':
            df.to_parquet(temp_file)
        
        # Upload to S3
        s3_client = boto3.client('s3')
        s3_client.upload_file(temp_file, bucket, key)
        
        # Clean up
        os.remove(temp_file)
        
        logger.info("S3 loading completed")
        return True

    async def _load_api(self, df: pd.DataFrame, config: Dict) -> bool:
        """Load data to API endpoint"""
        url = config['url']
        method = config.get('method', 'POST')
        headers = config.get('headers', {})
        batch_size = config.get('batch_size', 100)
        
        logger.info(f"Loading {len(df)} rows to API: {url}")
        
        # Convert DataFrame to records
        records = df.to_dict('records')
        
        # Send in batches
        async with aiohttp.ClientSession() as session:
            for i in range(0, len(records), batch_size):
                batch = records[i:i + batch_size]
                
                if method.upper() == 'POST':
                    async with session.post(url, headers=headers, json=batch) as response:
                        if response.status != 200:
                            logger.error(f"API loading failed: {response.status}")
                            return False
        
        logger.info("API loading completed")
        return True

class ETLPipeline:
    """Main ETL Pipeline orchestrator"""
    
    def __init__(self, config: ETLConfig):
        self.config = config
        self.extractor = DataExtractor(config.pipeline_config)
        self.transformer = DataTransformer(config.transformation_config)
        self.validator = DataValidator(config.validation_config)
        self.loader = DataLoader(config.pipeline_config)
        
    async def run(self) -> bool:
        """Run the complete ETL pipeline"""
        try:
            pipeline_start = datetime.now()
            logger.info("Starting ETL pipeline")
            
            # Extract
            logger.info("Step 1: Data Extraction")
            df = await self.extractor.extract(
                self.config.source_type,
                self.config.source_config
            )
            
            if df.empty:
                logger.warning("No data extracted, pipeline stopping")
                return False
            
            # Transform
            logger.info("Step 2: Data Transformation")
            df = self.transformer.transform(df, self.config.transformation_config)
            
            # Validate
            logger.info("Step 3: Data Validation")
            validation_results = self.validator.validate(df, self.config.validation_config)
            
            if not validation_results['passed']:
                logger.error("Data validation failed")
                if self.config.pipeline_config.get('stop_on_validation_error', True):
                    return False
            
            # Load
            logger.info("Step 4: Data Loading")
            success = await self.loader.load(
                df,
                self.config.destination_type,
                self.config.destination_config
            )
            
            if not success:
                logger.error("Data loading failed")
                return False
            
            # Pipeline completed
            pipeline_end = datetime.now()
            duration = (pipeline_end - pipeline_start).total_seconds()
            
            logger.info(f"ETL pipeline completed successfully in {duration:.2f} seconds")
            logger.info(f"Processed {len(df)} records")
            
            return True
            
        except Exception as e:
            logger.error(f"ETL pipeline failed: {str(e)}")
            return False

def load_config_from_file(config_file: str) -> ETLConfig:
    """Load ETL configuration from YAML file"""
    with open(config_file, 'r') as f:
        config_data = yaml.safe_load(f)
    
    return ETLConfig(
        source_type=config_data['source']['type'],
        source_config=config_data['source']['config'],
        destination_type=config_data['destination']['type'],
        destination_config=config_data['destination']['config'],
        transformation_config=config_data.get('transformations', {}),
        pipeline_config=config_data.get('pipeline', {}),
        validation_config=config_data.get('validation', {})
    )

async def main():
    parser = argparse.ArgumentParser(description='ETL Pipeline Script')
    parser.add_argument('--config', required=True, help='ETL configuration file')
    parser.add_argument('--source', help='Source type override')
    parser.add_argument('--target', help='Target type override')
    parser.add_argument('--dry-run', action='store_true', help='Validate configuration without running')
    
    args = parser.parse_args()
    
    try:
        # Load configuration
        config = load_config_from_file(args.config)
        
        # Override with command line arguments
        if args.source:
            config.source_type = args.source
        if args.target:
            config.destination_type = args.target
        
        if args.dry_run:
            logger.info("Configuration validation successful")
            return
        
        # Create and run pipeline
        pipeline = ETLPipeline(config)
        success = await pipeline.run()
        
        sys.exit(0 if success else 1)
        
    except Exception as e:
        logger.error(f"Pipeline execution failed: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())
