<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Go Common Scripts - Copilot Instructions

This workspace contains a collection of utility scripts written in Go for common development and system administration tasks.

## Project Context
- **Language**: Go (Golang)
- **Purpose**: Utility scripts for file operations, web scraping, data processing, HTTP tools, and system administration
- **Architecture**: Collection of standalone scripts, each with their own main function
- **Dependencies**: Minimal external dependencies, primarily using Go standard library

## Code Style Guidelines
- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add proper error handling with descriptive error messages
- Include comments for complex logic
- Use struct types for complex data structures
- Implement proper CLI argument parsing for command-line tools

## Script Categories
1. **File Operations**: File copying, watching, and manipulation
2. **HTTP Tools**: HTTP servers and clients for API testing
3. **Data Processing**: JSON and CSV processing utilities
4. **Web Scraping**: Website data extraction tools
5. **Concurrency**: Go concurrency pattern examples
6. **CLI Tools**: Command-line utilities (calculator, password generator, etc.)
7. **System Info**: System information gathering tools

## Development Preferences
- Each script should be self-contained and runnable with `go run`
- Provide both interactive and command-line interfaces where appropriate
- Include comprehensive help text and usage examples
- Handle edge cases and provide meaningful error messages
- Use channels and goroutines for concurrent operations
- Follow the single responsibility principle for functions

## Testing Approach
- Include example data files in the `examples/` directory
- Test scripts with various input scenarios
- Verify error handling with invalid inputs
- Test concurrent operations under load

## Documentation
- Each script should have clear usage instructions
- Include examples of common use cases
- Document any external dependencies
- Provide troubleshooting information for common issues

When generating code for this workspace, prioritize:
1. **Reliability**: Robust error handling and graceful failures
2. **Usability**: Clear interfaces and helpful error messages
3. **Performance**: Efficient algorithms and proper resource management
4. **Maintainability**: Clean, readable code with good separation of concerns
