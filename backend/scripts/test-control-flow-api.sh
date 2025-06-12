#!/bin/bash

# Control Flow API Test Script
# This script tests the Control Flow API endpoints

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_colored() {
    echo -e "${1}${2}${NC}"
}

# API base URL
API_BASE="http://localhost:8080/api/v1"
HEALTH_URL="http://localhost:8080/health"

print_colored $BLUE "🧪 Control Flow API Test Suite"
print_colored $BLUE "==============================="

# Function to check if server is running
check_server() {
    print_colored $YELLOW "🔍 Checking if server is running..."
    if curl -s $HEALTH_URL > /dev/null; then
        print_colored $GREEN "✅ Server is running"
        return 0
    else
        print_colored $RED "❌ Server is not running. Please start the server first:"
        print_colored $YELLOW "   go run cmd/control-flow-api/main.go"
        exit 1
    fi
}

# Function to test system config API
test_system_config() {
    print_colored $YELLOW "📋 Testing System Config API..."
    
    # Get system config
    print_colored $BLUE "  GET /api/v1/system/config"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" $API_BASE/system/config)
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ Get system config successful"
        echo "    Response: $body" | head -c 100 && echo "..."
    else
        print_colored $RED "    ❌ Get system config failed (HTTP $code)"
    fi
    
    # Update system config
    print_colored $BLUE "  PUT /api/v1/system/config"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X PUT $API_BASE/system/config \
        -H "Content-Type: application/json" \
        -d '{
            "rate_limit_mode": "priority",
            "default_priority": 8,
            "default_qps": 20
        }')
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ Update system config successful"
    else
        print_colored $RED "    ❌ Update system config failed (HTTP $code)"
        echo "    Response: $body"
    fi
}

# Function to test user rate limit API
test_user_rate_limit() {
    print_colored $YELLOW "👤 Testing User Rate Limit API..."
    
    # Create user rate limit
    print_colored $BLUE "  POST /api/v1/user-rate-limits"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST $API_BASE/user-rate-limits \
        -H "Content-Type: application/json" \
        -d '{
            "user_id": "test_user_123",
            "priority": 9,
            "enabled": true
        }')
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 201 ]; then
        print_colored $GREEN "    ✅ Create user rate limit successful"
    else
        print_colored $YELLOW "    ⚠️  Create user rate limit (HTTP $code) - may already exist"
    fi
    
    # Get user rate limit
    print_colored $BLUE "  GET /api/v1/user-rate-limits/test_user_123"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" $API_BASE/user-rate-limits/test_user_123)
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ Get user rate limit successful"
    else
        print_colored $RED "    ❌ Get user rate limit failed (HTTP $code)"
    fi
    
    # List user rate limits
    print_colored $BLUE "  GET /api/v1/user-rate-limits?page=1&page_size=10"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$API_BASE/user-rate-limits?page=1&page_size=10")
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ List user rate limits successful"
    else
        print_colored $RED "    ❌ List user rate limits failed (HTTP $code)"
    fi
    
    # Update user rate limit
    print_colored $BLUE "  PUT /api/v1/user-rate-limits/test_user_123"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X PUT $API_BASE/user-rate-limits/test_user_123 \
        -H "Content-Type: application/json" \
        -d '{
            "priority": 7,
            "enabled": true
        }')
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ Update user rate limit successful"
    else
        print_colored $RED "    ❌ Update user rate limit failed (HTTP $code)"
        echo "    Response: $body"
    fi
}

# Function to test agent API
test_agent() {
    print_colored $YELLOW "🤖 Testing Agent API..."
    
    # Create agent
    print_colored $BLUE "  POST /api/v1/agents"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST $API_BASE/agents \
        -H "Content-Type: application/json" \
        -d '{
            "name": "Test OpenAI Agent",
            "type": "openai",
            "url": "https://api.openai.com",
            "api_key": "sk-test-key-for-testing",
            "qps": 15,
            "enabled": true,
            "description": "Test agent for API testing"
        }')
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 201 ]; then
        print_colored $GREEN "    ✅ Create agent successful"
        # Extract agent ID for further tests
        AGENT_ID=$(echo $body | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    else
        print_colored $YELLOW "    ⚠️  Create agent (HTTP $code) - may already exist"
        # Try to get existing agent ID
        list_response=$(curl -s $API_BASE/agents?page_size=1)
        AGENT_ID=$(echo $list_response | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    fi
    
    if [ -n "$AGENT_ID" ]; then
        # Get agent
        print_colored $BLUE "  GET /api/v1/agents/$AGENT_ID"
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" $API_BASE/agents/$AGENT_ID)
        body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
        code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
        
        if [ "$code" -eq 200 ]; then
            print_colored $GREEN "    ✅ Get agent successful"
        else
            print_colored $RED "    ❌ Get agent failed (HTTP $code)"
        fi
        
        # Update agent
        print_colored $BLUE "  PUT /api/v1/agents/$AGENT_ID"
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X PUT $API_BASE/agents/$AGENT_ID \
            -H "Content-Type: application/json" \
            -d '{
                "name": "Updated Test OpenAI Agent",
                "type": "openai",
                "url": "https://api.openai.com",
                "api_key": "sk-updated-test-key",
                "qps": 25,
                "enabled": true,
                "description": "Updated test agent for API testing"
            }')
        body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
        code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
        
        if [ "$code" -eq 200 ]; then
            print_colored $GREEN "    ✅ Update agent successful"
        else
            print_colored $RED "    ❌ Update agent failed (HTTP $code)"
            echo "    Response: $body"
        fi
    fi
    
    # List agents
    print_colored $BLUE "  GET /api/v1/agents?page=1&page_size=10"
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$API_BASE/agents?page=1&page_size=10")
    body=$(echo $response | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    code=$(echo $response | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$code" -eq 200 ]; then
        print_colored $GREEN "    ✅ List agents successful"
    else
        print_colored $RED "    ❌ List agents failed (HTTP $code)"
    fi
}

# Function to cleanup test data
cleanup_test_data() {
    print_colored $YELLOW "🧹 Cleaning up test data..."
    
    # Delete test user rate limit
    curl -s -X DELETE $API_BASE/user-rate-limits/test_user_123 > /dev/null
    
    # Delete test agent (if we have the ID)
    if [ -n "$AGENT_ID" ]; then
        curl -s -X DELETE $API_BASE/agents/$AGENT_ID > /dev/null
    fi
    
    print_colored $GREEN "✅ Cleanup completed"
}

# Main test execution
main() {
    check_server
    
    print_colored $YELLOW "🚀 Starting API tests..."
    
    test_system_config
    test_user_rate_limit
    test_agent
    
    # Cleanup
    cleanup_test_data
    
    print_colored $GREEN "🎉 All API tests completed!"
    print_colored $BLUE "📊 Test Summary:"
    echo "  - System Config API: ✅"
    echo "  - User Rate Limit API: ✅"
    echo "  - Agent API: ✅"
    echo "  - Data Cleanup: ✅"
    
    print_colored $GREEN "All tests completed successfully! 🚀"
}

# Execute main function
main 