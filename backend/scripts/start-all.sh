#!/bin/bash

# Agent-Connector backend services start script
# Usage: ./start-all.sh

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print info function
print_info() {
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

# Check environment variable file
check_env_file() {
    if [ ! -f ".env" ]; then
        print_warning ".env file not found, using default configuration"
        if [ -f ".env.example" ]; then
            print_info "Copying .env.example to .env"
            cp .env.example .env
        else
            print_error ".env.example file not found, please create .env file manually"
            exit 1
        fi
    fi
    print_success "Environment variable file check completed"
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check MySQL
    if ! command -v mysql &> /dev/null; then
        print_warning "MySQL client not found, cannot verify MySQL connection"
    else
        print_info "MySQL client found"
    fi
    
    # Check Redis
    if ! command -v redis-cli &> /dev/null; then
        print_warning "Redis client not found, cannot verify Redis connection"
    else
        print_info "Redis client found"
    fi
    
    print_success "Dependency check completed"
}

# Build services
build_services() {
    print_info "Building backend services..."
    
    # Build authentication service
    print_info "Building authentication service..."
    go build -o bin/auth-api ./cmd/auth-api/
    
    # Build control flow service
    print_info "Building control flow service..."
    go build -o bin/control-flow-api ./cmd/control-flow-api/
    
    # Build data flow service
    print_info "Building data flow service..."
    go build -o bin/dataflow-api ./cmd/dataflow-api/
    
    print_success "All services built"
}

# Start services
start_services() {
    print_info "Starting backend services..."
    
    # Load environment variables
    if [ -f ".env" ]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Create log directory
    mkdir -p logs
    
    # Start authentication service
    print_info "Starting authentication service (port: ${AUTH_API_PORT:-8083})..."
    nohup ./bin/auth-api > logs/auth-api.log 2>&1 &
    AUTH_PID=$!
    echo $AUTH_PID > logs/auth-api.pid
    sleep 2
    
    # Start control flow service
    print_info "Starting control flow service (port: ${CONTROL_FLOW_API_PORT:-8081})..."
    nohup ./bin/control-flow-api > logs/control-flow-api.log 2>&1 &
    CONTROL_PID=$!
    echo $CONTROL_PID > logs/control-flow-api.pid
    sleep 2
    
    # Start data flow service
    print_info "Starting data flow service (port: ${DATA_FLOW_API_PORT:-8082})..."
    nohup ./bin/dataflow-api > logs/dataflow-api.log 2>&1 &
    DATA_PID=$!
    echo $DATA_PID > logs/dataflow-api.pid
    sleep 2
    
    print_success "All services started"
}

# Check service status
check_services() {
    print_info "Checking service status..."
    
    # Check authentication service
    if curl -s "http://localhost:${AUTH_API_PORT:-8083}/" > /dev/null 2>&1; then
        print_success "Authentication service running (http://localhost:${AUTH_API_PORT:-8083})"
    else
        print_error "Authentication service failed to start"
    fi
    
    # Check control flow service
    if curl -s "http://localhost:${CONTROL_FLOW_API_PORT:-8081}/" > /dev/null 2>&1; then
        print_success "Control flow service running (http://localhost:${CONTROL_FLOW_API_PORT:-8081})"
    else
        print_error "Control flow service failed to start"
    fi
    
    # Check data flow service
    if curl -s "http://localhost:${DATA_FLOW_API_PORT:-8082}/" > /dev/null 2>&1; then
        print_success "Data flow service running (http://localhost:${DATA_FLOW_API_PORT:-8082})"
    else
        print_error "Data flow service failed to start"
    fi
}

# Show
show_service_info() {
    echo ""
    echo "========================================="
    echo "🚀 Agent-Connector backend services started"
    echo "========================================="
    echo "Authentication service:     http://localhost:${AUTH_API_PORT:-8083}"
    echo "Control flow service:   http://localhost:${CONTROL_FLOW_API_PORT:-8081}"
    echo "Data flow service:   http://localhost:${DATA_FLOW_API_PORT:-8082}"
    echo "========================================="
    echo "Log file location: ./logs/"
    echo "PID file location: ./logs/*.pid"
    echo ""
    echo "Stop services: ./scripts/stop-all.sh"
    echo "Check status: ./scripts/status.sh"
    echo "Check logs: tail -f logs/service-name.log"
    echo "========================================="
}

# Main function
main() {
    echo "========================================="
    echo "🚀 Agent-Connector backend services starter"
    echo "========================================="
    
    check_env_file
    check_dependencies
    build_services
    start_services
    sleep 3
    check_services
    show_service_info
}

# 执行主函数
main 