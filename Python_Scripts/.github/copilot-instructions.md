# Copilot Instructions for Python Cloud Engineering Scripts

This workspace contains comprehensive Python scripts for cloud engineering, DevOps, and data operations. Here are the key guidelines for working with this codebase:

## 🎯 Project Purpose
This collection provides production-ready Python scripts for:
- Data operations (backup, migration, ETL, sync)
- AWS cloud operations
- System monitoring and alerting
- Infrastructure automation
- Security and compliance
- Cost optimization

## 📁 Directory Structure
```
Python_Scripts/
├── data-operations/     # Database backup, migration, ETL, sync
├── aws-operations/      # AWS resource management
├── monitoring/          # System monitoring and alerting
├── automation/          # CI/CD and infrastructure automation
├── security/           # Security scanning and compliance
├── cost-management/    # Cloud cost analysis and optimization
├── network-operations/ # Network configuration and monitoring
└── performance/        # Performance testing and optimization
```

## 🛠️ Code Standards

### Python Style
- Follow PEP 8 style guidelines
- Use Black formatter with 88-character line length
- Use type hints for function parameters and return values
- Include comprehensive docstrings for all modules, classes, and functions
- Use f-strings for string formatting

### Error Handling
- Always include comprehensive exception handling
- Log errors with appropriate context
- Use structured logging with timestamps
- Implement graceful degradation where possible
- Provide meaningful error messages

### Configuration
- Use environment variables for sensitive data
- Support YAML configuration files for complex setups
- Provide example configurations in config/ directories
- Use python-dotenv for environment management
- Include validation for required configuration

### Dependencies
- Keep requirements.txt updated
- Use specific version ranges for stability
- Group dependencies logically with comments
- Consider optional dependencies for advanced features
- Document any system-level dependencies

## 🔧 Development Guidelines

### Adding New Scripts
1. Follow the existing directory structure
2. Include comprehensive docstrings and comments
3. Add corresponding configuration examples
4. Create VS Code tasks for common operations
5. Update the relevant README files
6. Add logging and monitoring capabilities

### Database Operations
- Use SQLAlchemy for database abstraction
- Support multiple database types where applicable
- Include connection pooling for production use
- Implement proper transaction handling
- Add retry logic for transient failures

### AWS Integration
- Use boto3 with proper error handling
- Support multiple authentication methods
- Include resource tagging and cost optimization
- Implement pagination for large result sets
- Add rate limiting and backoff strategies

### Async Programming
- Use asyncio for I/O-bound operations
- Implement proper exception handling in async contexts
- Use aiohttp for HTTP operations
- Consider memory usage with large async operations
- Add proper resource cleanup (async context managers)

## 📊 Data Operations Specifics

### Backup Scripts
- Support multiple database types (PostgreSQL, MySQL, MongoDB, SQLite)
- Include compression and encryption options
- Implement retention policies
- Add backup verification
- Support cloud storage integration

### Migration Scripts
- Handle schema differences between database types
- Implement data transformation during migration
- Support incremental migrations
- Include rollback capabilities
- Add progress tracking and reporting

### ETL Pipelines
- Support multiple data sources and destinations
- Include data validation and quality checks
- Implement parallel processing for performance
- Add monitoring and alerting
- Support schema evolution

## 🔍 Testing and Validation

### Script Testing
- Test with small datasets before production use
- Use dry-run modes where applicable
- Validate configurations before execution
- Test error scenarios and recovery
- Include performance testing for large datasets

### Configuration Validation
- Validate YAML configurations at startup
- Check required environment variables
- Test database connections before operations
- Verify cloud service permissions
- Validate file paths and permissions

## 📚 Documentation

### README Files
- Include comprehensive usage examples
- Document all configuration options
- Provide troubleshooting guides
- Include performance optimization tips
- Add security considerations

### Code Documentation
- Use detailed docstrings with examples
- Document complex algorithms and business logic
- Include parameter and return type documentation
- Add inline comments for complex operations
- Document any external service dependencies

## 🚀 Best Practices

### Security
- Never hardcode sensitive information
- Use secure connection strings
- Implement proper access controls
- Log security-relevant events
- Regular security dependency updates

### Performance
- Use appropriate batch sizes for large operations
- Implement connection pooling
- Add caching where beneficial
- Monitor memory usage
- Use efficient data structures

### Monitoring
- Log all important operations
- Include metrics and performance data
- Set up alerting for failures
- Monitor resource usage
- Track operation success/failure rates

### Cloud Integration
- Use cloud-native services where appropriate
- Implement proper retry and backoff strategies
- Consider regional availability
- Optimize for cost efficiency
- Use infrastructure as code principles

## 🔗 Common Patterns

### Configuration Loading
```python
from dotenv import load_dotenv
import yaml

load_dotenv()
with open('config.yaml', 'r') as f:
    config = yaml.safe_load(f)
```

### Database Connections
```python
from sqlalchemy import create_engine
engine = create_engine(connection_string, pool_size=10, max_overflow=20)
```

### Async Operations
```python
async def process_data():
    async with aiohttp.ClientSession() as session:
        # async operations
        pass
```

### Error Handling
```python
try:
    # operation
    pass
except SpecificException as e:
    logger.error(f"Operation failed: {str(e)}")
    # handle specific error
except Exception as e:
    logger.error(f"Unexpected error: {str(e)}")
    raise
```

This workspace is designed for production use in cloud engineering environments. Always prioritize reliability, security, and maintainability in your implementations.
