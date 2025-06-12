#!/bin/bash

# Authentication API start script
# Usage: ./start_auth_api.sh [options]

# Default configuration
HOST="localhost"
PORT="8083"
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASS=""
DB_NAME="agent_connector"

# Color definitions
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "Agent-Connector Authentication API"
echo "=========================================="

# Check if compiled
if [ ! -f "bin/auth-api" ]; then
    echo -e "${YELLOW}Compiling Authentication API...${NC}"
    go build -o bin/auth-api ./cmd/auth-api
    if [ $? -ne 0 ]; then
        echo -e "${RED}Compilation failed!${NC}"
        exit 1
    fi
    echo -e "${GREEN}Compilation successful!${NC}"
fi

# Check MySQL connection
echo -e "${YELLOW}Checking database connection...${NC}"
if command -v mysql &> /dev/null; then
    if [ -n "$DB_PASS" ]; then
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME;" 2>/dev/null
    else
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -e "USE $DB_NAME;" 2>/dev/null
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Database connection successful${NC}"
    else
        echo -e "${YELLOW}Database connection failed, please ensure MySQL is running and the database exists${NC}"
        echo "Create database command: CREATE DATABASE $DB_NAME;"
    fi
else
    echo -e "${YELLOW}mysql client not found, skipping database check${NC}"
fi

# Display configuration information
echo ""
echo "Service configuration:"
echo "  Listening address: $HOST:$PORT"
echo "  Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo ""
echo "API endpoints:"
echo "  Service info: http://$HOST:$PORT/"
echo "  Health check: http://$HOST:$PORT/api/v1/auth/health"
echo "  User register: POST http://$HOST:$PORT/api/v1/auth/register"
echo "  User login: POST http://$HOST:$PORT/api/v1/auth/login"
echo "  API documentation: see AUTH_API_SUMMARY.md"
echo ""

# Start service
echo -e "${GREEN}Starting Authentication API server...${NC}"
echo "Press Ctrl+C to stop the service"
echo ""

exec ./bin/auth-api \
    -host="$HOST" \
    -port="$PORT" \
    -db-host="$DB_HOST" \
    -db-port="$DB_PORT" \
    -db-user="$DB_USER" \
    -db-pass="$DB_PASS" \
    -db-name="$DB_NAME" 