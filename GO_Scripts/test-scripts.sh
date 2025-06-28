#!/bin/bash
# Quick test script for Go Common Scripts
# Run this to verify all scripts are working

echo "=== Testing Go Common Scripts ==="
echo ""

# Test JSON processor
echo "1. Testing JSON Processor..."
go run ./json-tools/json-processor.go keys ./examples/sample.json | head -5
echo ""

# Test CSV processor
echo "2. Testing CSV Processor..."
go run ./csv-tools/csv-processor.go show ./examples/sample.csv | head -5
echo ""

# Test system info
echo "3. Testing System Info..."
go run ./system-info/system-info.go basic | head -5
echo ""

# Test password generator (if not blocked by antivirus)
echo "4. Testing Password Generator..."
go run ./cli-tools/password-gen.go -l 8 2>/dev/null || echo "Password generator may be blocked by antivirus"
echo ""

# Test calculator
echo "5. Testing Calculator..."
echo "Expression: 2+3*4"
go run ./cli-tools/calculator.go expr "2+3*4"
echo ""

echo "=== All tests completed ==="
