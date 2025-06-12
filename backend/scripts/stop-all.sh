#!/bin/bash

# Agent-Connector backend services stop script
# Usage: ./stop-all.sh

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Stop service
stop_service() {
    local service_name=$1
    local pid_file="logs/${service_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            print_info "Stopping ${service_name} (PID: $pid)..."
            kill $pid
            sleep 2
            
            # Check if stopped successfully
            if ps -p $pid > /dev/null 2>&1; then
                print_warning "${service_name} failed to stop, force killing..."
                kill -9 $pid
            fi
            
            rm -f "$pid_file"
            print_success "${service_name} stopped"
        else
            print_warning "${service_name} process not found (PID: $pid)"
            rm -f "$pid_file"
        fi
    else
        print_warning "${service_name} PID file not found"
    fi
}

# Main function
main() {
    echo "========================================="
    echo "🛑 Agent-Connector backend services stopper"
    echo "========================================="
    
    print_info "Stopping all backend services..."
    
    stop_service "auth-api"
    stop_service "control-flow-api"
    stop_service "dataflow-api"
    
    print_success "All services stopped"
    echo "========================================="
}

main 