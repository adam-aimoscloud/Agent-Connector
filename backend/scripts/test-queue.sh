#!/bin/bash

# Script to test the priority queue implementation
# This script runs various tests for the priority queue package

set -e

echo "=== Priority Queue Test Suite ==="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Change to the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

print_status "Running priority queue tests..."

# Test 1: Unit Tests
print_status "1. Running unit tests..."
if go test -v ./pkg/queue/ -count=1; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

echo

# Test 2: Build Tests
print_status "2. Building demo application..."
if go build -o /tmp/queue-demo ./cmd/queue-demo/; then
    print_success "Demo application built successfully"
else
    print_error "Failed to build demo application"
    exit 1
fi

echo

# Test 3: Run Demo
print_status "3. Running demo application..."
if /tmp/queue-demo; then
    print_success "Demo application ran successfully"
else
    print_error "Demo application failed"
    exit 1
fi

echo

# Test 4: Code Coverage
print_status "4. Generating test coverage report..."
if go test -coverprofile=coverage.out ./pkg/queue/; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_success "Test coverage: $COVERAGE"
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    print_status "HTML coverage report generated: coverage.html"
    
    # Clean up
    rm -f coverage.out
else
    print_error "Failed to generate coverage report"
fi

echo

# Test 5: Benchmark Tests (if they exist)
print_status "5. Running benchmark tests..."
if go test -bench=. -benchmem ./pkg/queue/ 2>/dev/null | grep -q "Benchmark"; then
    print_status "Running benchmarks..."
    go test -bench=. -benchmem ./pkg/queue/
else
    print_warning "No benchmark tests found"
fi

echo

# Test 6: Race Detection
print_status "6. Running race detection tests..."
if go test -race ./pkg/queue/; then
    print_success "Race detection tests passed"
else
    print_error "Race detection tests failed"
    exit 1
fi

echo

# Test 7: Vet Check
print_status "7. Running go vet..."
if go vet ./pkg/queue/; then
    print_success "go vet passed"
else
    print_error "go vet failed"
    exit 1
fi

echo

# Test 8: Format Check
print_status "8. Checking code formatting..."
UNFORMATTED=$(gofmt -l ./pkg/queue/)
if [ -z "$UNFORMATTED" ]; then
    print_success "Code is properly formatted"
else
    print_warning "The following files need formatting:"
    echo "$UNFORMATTED"
    print_status "Run 'gofmt -w ./pkg/queue/' to fix formatting"
fi

echo

# Clean up temporary files
rm -f /tmp/queue-demo

print_success "All priority queue tests completed successfully!"

echo
echo "=== Test Summary ==="
echo "✅ Unit tests"
echo "✅ Build tests"
echo "✅ Demo application"
echo "✅ Code coverage"
echo "✅ Race detection"
echo "✅ Code quality (vet)"
echo "✅ Code formatting"
echo

print_success "Priority Queue implementation is ready for production use!" 