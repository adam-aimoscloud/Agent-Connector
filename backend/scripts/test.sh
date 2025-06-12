#!/bin/bash

# Test script for agent-connector rate limiter module

set -e

echo "Running Go tests for rate limiter module..."

# Change to the backend directory
cd "$(dirname "$0")/.."

# Download dependencies
echo "Downloading Go dependencies..."
go mod download

# Run tests with verbose output and coverage
echo "Running unit tests..."
go test -v -race -coverprofile=coverage.out ./pkg/ratelimiter/...

# Display coverage report
echo "Coverage report:"
go tool cover -func=coverage.out

# Generate HTML coverage report
echo "Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "HTML coverage report generated: coverage.html"

# Run benchmarks
echo "Running benchmarks..."
go test -bench=. -benchmem ./pkg/ratelimiter/...

echo "All tests completed successfully!" 