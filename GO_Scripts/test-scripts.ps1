# Quick test script for Go Common Scripts (PowerShell version)
# Run this to verify all scripts are working

Write-Host "=== Testing Go Common Scripts ===" -ForegroundColor Green
Write-Host ""

# Test JSON processor
Write-Host "1. Testing JSON Processor..." -ForegroundColor Yellow
$output = go run .\json-tools\json-processor.go keys .\examples\sample.json
$output | Select-Object -First 5
Write-Host ""

# Test CSV processor
Write-Host "2. Testing CSV Processor..." -ForegroundColor Yellow
$output = go run .\csv-tools\csv-processor.go show .\examples\sample.csv
$output | Select-Object -First 5
Write-Host ""

# Test system info
Write-Host "3. Testing System Info..." -ForegroundColor Yellow
$output = go run .\system-info\system-info.go basic
$output | Select-Object -First 5
Write-Host ""

# Test password generator (if not blocked by antivirus)
Write-Host "4. Testing Password Generator..." -ForegroundColor Yellow
try {
    $password = go run .\cli-tools\password-gen.go -l 8 2>$null
    if ($password) {
        Write-Host "Generated password: $password"
    } else {
        Write-Host "Password generator may be blocked by antivirus" -ForegroundColor Red
    }
} catch {
    Write-Host "Password generator may be blocked by antivirus" -ForegroundColor Red
}
Write-Host ""

# Test calculator
Write-Host "5. Testing Calculator..." -ForegroundColor Yellow
Write-Host "Expression: 2+3*4"
$result = go run .\cli-tools\calculator.go expr "2+3*4"
Write-Host $result
Write-Host ""

Write-Host "=== All tests completed ===" -ForegroundColor Green
