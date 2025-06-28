# Go Common Scripts Collection

A comprehensive collection of utility scripts written in Go for common development and system administration tasks.

## 📁 Project Structure

```
GO_Scripts/
├── file-operations/          # File manipulation utilities
├── http-server/             # HTTP server implementations
├── http-client/             # HTTP client utilities
├── json-tools/              # JSON processing tools
├── csv-tools/               # CSV processing tools
├── web-scraper/             # Web scraping utilities
├── concurrent/              # Concurrency examples
├── cli-tools/               # Command-line tools
├── system-info/             # System information tools
└── examples/                # Example data files
```

## 🚀 Quick Start

1. **Clone or download** this repository
2. **Navigate** to any script directory
3. **Run** the script with Go:
   ```bash
   go run script-name.go [options]
   ```

## 📝 Available Scripts

### 🗂️ File Operations

#### File Copy (`file-operations/file-copy.go`)
Copy files or directories recursively.

```bash
# Copy a file
go run file-copy.go source.txt destination.txt

# Copy a directory
go run file-copy.go /path/to/source /path/to/destination
```

#### File Watcher (`file-operations/file-watcher.go`)
Monitor files and directories for changes in real-time.

```bash
# Watch current directory
go run file-watcher.go .

# Watch multiple paths
go run file-watcher.go /path/to/dir1 /path/to/file.txt
```

**Features:**
- Real-time file system monitoring
- Multiple path watching
- Event type detection (create, modify, delete, rename, chmod)
- Timestamps for all events

### 🌐 HTTP Tools

#### Simple HTTP Server (`http-server/simple-server.go`)
A feature-rich HTTP server with multiple endpoints.

```bash
# Start server on port 8080 (default)
go run simple-server.go

# Start server on custom port
go run simple-server.go 3000
```

**Available Endpoints:**
- `GET /` - Server information and endpoint list
- `GET /health` - Health check with uptime
- `GET /api/status` - Server status
- `POST /api/echo` - Echo request body
- `GET /api/time` - Current server time
- `GET /static/*` - Static file serving

#### HTTP Client (`http-client/http-client.go`)
Versatile HTTP client for API testing and web requests.

```bash
# GET request
go run http-client.go GET https://httpbin.org/get

# POST request with JSON body
go run http-client.go POST https://httpbin.org/post '{"key":"value"}'

# Download file
go run http-client.go download https://httpbin.org/get response.json
```

**Features:**
- Support for GET, POST, PUT, DELETE methods
- File download capability
- Custom headers support
- Request/response logging

### 📊 Data Processing

#### JSON Processor (`json-tools/json-processor.go`)
Comprehensive JSON manipulation and analysis tool.

```bash
# Pretty print JSON
go run json-processor.go pretty data.json

# List all keys
go run json-processor.go keys data.json

# Get specific value
go run json-processor.go get data.json user.name

# Show statistics
go run json-processor.go stats data.json

# Validate JSON
go run json-processor.go validate data.json
```

**Features:**
- JSON validation and pretty printing
- Key extraction and nested path access
- Statistical analysis (type distribution, depth, etc.)
- Path-based value retrieval

#### CSV Processor (`csv-tools/csv-processor.go`)
Convert between CSV and JSON, filter data, and analyze CSV files.

```bash
# Display CSV content
go run csv-processor.go show data.csv

# Convert CSV to JSON
go run csv-processor.go to-json data.csv output.json

# Convert JSON to CSV
go run csv-processor.go from-json data.json output.csv

# Filter by column
go run csv-processor.go filter data.csv 0 "search_term"

# Get specific column
go run csv-processor.go column data.csv "column_name"

# Show statistics
go run csv-processor.go stats data.csv
```

### 🕷️ Web Scraping

#### Web Scraper (`web-scraper/web-scraper.go`)
Extract data from websites with various scraping capabilities.

```bash
# Full page scrape
go run web-scraper.go scrape https://example.com

# Extract all links
go run web-scraper.go links https://news.ycombinator.com

# Extract images
go run web-scraper.go images https://example.com

# Extract emails
go run web-scraper.go emails https://example.com

# Extract phone numbers
go run web-scraper.go phones https://example.com

# Get page title
go run web-scraper.go title https://example.com
```

**Features:**
- Link and image extraction
- Email and phone number detection
- Respectful crawling with delays
- Custom User-Agent support

### ⚡ Concurrency Examples

#### Concurrent Processor (`concurrent/concurrent-processor.go`)
Demonstrate Go's concurrency patterns and worker pools.

```bash
# CPU-intensive work with worker pool
go run concurrent-processor.go cpu

# I/O-intensive work simulation
go run concurrent-processor.go io

# Web request simulation
go run concurrent-processor.go web

# Map-reduce pattern
go run concurrent-processor.go mapreduce

# Producer-consumer pattern
go run concurrent-processor.go prodcons

# Fan-out fan-in pattern
go run concurrent-processor.go fanout

# Run all examples
go run concurrent-processor.go all
```

**Patterns Demonstrated:**
- Worker pools with job queues
- Map-reduce operations
- Producer-consumer patterns
- Fan-out/fan-in architectures
- Goroutine synchronization

### 🛠️ CLI Tools

#### Calculator (`cli-tools/calculator.go`)
Interactive and command-line calculator with memory functions.

```bash
# Interactive mode
go run calculator.go

# Command-line operations
go run calculator.go add 5 3
go run calculator.go multiply 4 7
go run calculator.go expr "2+3*4"
```

**Features:**
- Basic arithmetic operations
- Memory store/recall/clear
- Expression evaluation
- Calculation history
- Interactive and CLI modes

#### Password Generator (`cli-tools/password-gen.go`)
Generate secure passwords with customizable options.

```bash
# Generate default password
go run password-gen.go

# Custom length and character sets
go run password-gen.go -l 16 --symbols

# Multiple passwords
go run password-gen.go -c 5 -l 12

# Check password strength
go run password-gen.go --check "mypassword123"

# Exclude similar characters
go run password-gen.go --exclude-similar -l 20
```

**Features:**
- Customizable length and character sets
- Symbol inclusion/exclusion
- Similar character filtering
- Batch generation
- Password strength analysis

### 💻 System Information

#### System Info (`system-info/system-info.go`)
Gather comprehensive system and runtime information.

```bash
# Basic system info
go run system-info.go

# Memory statistics
go run system-info.go memory

# Environment variables
go run system-info.go env

# Go runtime info
go run system-info.go runtime

# All information
go run system-info.go all
```

**Information Gathered:**
- OS and architecture details
- CPU and memory information
- Go version and runtime stats
- Environment variables
- Memory usage and GC statistics

## 📋 Requirements

- **Go 1.19+** (tested with Go 1.21)
- **Internet connection** for web-related scripts
- **File system permissions** for file operations

## 🔧 Installation & Dependencies

1. **Initialize Go module** (if not already done):
   ```bash
   go mod init go-common-scripts
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run any script**:
   ```bash
   go run [directory]/[script-name].go [options]
   ```

## 📚 Examples

### Example Data Files

The `examples/` directory contains sample data files for testing:

- `sample.json` - JSON data for testing JSON processor
- `sample.csv` - CSV data for testing CSV processor
- `config.json` - Configuration file example

### Common Use Cases

1. **Monitor log files**:
   ```bash
   go run file-operations/file-watcher.go /var/log/app.log
   ```

2. **API testing**:
   ```bash
   go run http-client/http-client.go POST https://api.example.com/users '{"name":"John"}'
   ```

3. **Data conversion**:
   ```bash
   go run csv-tools/csv-processor.go to-json data.csv | go run json-tools/json-processor.go pretty
   ```

4. **Website analysis**:
   ```bash
   go run web-scraper/web-scraper.go scrape https://news.ycombinator.com
   ```

## 🤝 Contributing

Feel free to contribute by:
- Adding new utility scripts
- Improving existing functionality
- Adding better error handling
- Writing tests
- Improving documentation

## ⚠️ Important Notes

1. **Web Scraping**: Always respect robots.txt and website terms of service
2. **File Operations**: Be careful with file permissions and paths
3. **Network Requests**: Some scripts may be blocked by firewalls or rate limiting
4. **System Info**: Some information may be sensitive in production environments

## 📄 License

This collection is provided as-is for educational and utility purposes. Use responsibly and in accordance with applicable laws and regulations.

---

**Happy Coding! 🚀**
