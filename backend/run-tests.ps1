#!/usr/bin/env pwsh
# Test runner script for Typing Master Backend
# This script runs all unit tests with coverage reporting

param(
    [string]$Package = "./...",
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Race,
    [switch]$Short,
    [string]$Run = ""
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Typing Master Backend - Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Build test command
$testCmd = "go test"
$testArgs = @()

# Add package path
$testArgs += $Package

# Add flags
if ($Verbose) {
    $testArgs += "-v"
}

if ($Coverage) {
    $testArgs += "-coverprofile=coverage.out"
    $testArgs += "-covermode=atomic"
}

if ($Race) {
    $testArgs += "-race"
}

if ($Short) {
    $testArgs += "-short"
}

if ($Run -ne "") {
    $testArgs += "-run"
    $testArgs += $Run
}

# Run tests
Write-Host "Running tests..." -ForegroundColor Yellow
Write-Host "Command: $testCmd $($testArgs -join ' ')" -ForegroundColor Gray
Write-Host ""

$env:GO_ENV = "test"

& go test @testArgs

$exitCode = $LASTEXITCODE

if ($exitCode -eq 0) {
    Write-Host ""
    Write-Host "✓ All tests passed!" -ForegroundColor Green
    
    # Generate coverage report if requested
    if ($Coverage -and (Test-Path "coverage.out")) {
        Write-Host ""
        Write-Host "Generating coverage report..." -ForegroundColor Yellow
        
        # Display coverage summary
        go tool cover -func=coverage.out
        
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        Write-Host ""
        Write-Host "✓ Coverage report generated: coverage.html" -ForegroundColor Green
        
        # Calculate total coverage
        $coverageData = go tool cover -func=coverage.out | Select-String "total:"
        if ($coverageData) {
            Write-Host ""
            Write-Host "Coverage Summary:" -ForegroundColor Cyan
            Write-Host $coverageData -ForegroundColor White
        }
    }
} else {
    Write-Host ""
    Write-Host "✗ Tests failed!" -ForegroundColor Red
    exit $exitCode
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test run completed" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
