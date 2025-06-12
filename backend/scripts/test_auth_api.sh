#!/bin/bash

# Full script to test the authentication API
# Usage: ./test_auth_api.sh [server_url]

SERVER_URL=${1:-"http://localhost:8083"}
API_BASE="$SERVER_URL/api/v1"

echo "==============================================="
echo "Authentication API test script"
echo "Server: $SERVER_URL"
echo "==============================================="

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0

# Test function
run_test() {
    local test_name="$1"
    local expected_code="$2"
    local actual_code="$3"
    local response="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$actual_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} - $test_name (HTTP $actual_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} - $test_name (Expected: $expected_code, Got: $actual_code)"
        echo -e "${YELLOW}Response:${NC} $response"
        return 1
    fi
}

# 1. Service basic test
echo -e "\n${BLUE}=== 1. Service basic test ===${NC}"

# Root path test
response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$SERVER_URL/")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Root path access" 200 "$http_code" "$content"

# Health check
response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$API_BASE/auth/health")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Health check" 200 "$http_code" "$content"

# Service info
response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$API_BASE/auth/")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Service info" 200 "$http_code" "$content"

# 2. User registration test
echo -e "\n${BLUE}=== 2. User registration test ===${NC}"

# Normal registration
register_data='{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
}'

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$register_data" \
    "$API_BASE/auth/register")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "User registration" 201 "$http_code" "$content"

# Duplicate registration (should fail)
response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$register_data" \
    "$API_BASE/auth/register")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Duplicate username registration (should fail)" 409 "$http_code" "$content"

# Invalid data registration
invalid_register_data='{
    "username": "a",
    "email": "invalid-email",
    "password": "123"
}'

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$invalid_register_data" \
    "$API_BASE/auth/register")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Invalid data registration (should fail)" 400 "$http_code" "$content"

# 3. User login test
echo -e "\n${BLUE}=== 3. User login test ===${NC}"

# Normal login
login_data='{
    "username": "testuser",
    "password": "password123"
}'

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$login_data" \
    "$API_BASE/auth/login")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")

if [ "$http_code" -eq 200 ]; then
    TOKEN=$(echo "$content" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}Login successful, got token: ${TOKEN:0:20}...${NC}"
fi

run_test "User login" 200 "$http_code" "$content"

# Wrong password login
wrong_login_data='{
    "username": "testuser",
    "password": "wrongpassword"
}'

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$wrong_login_data" \
    "$API_BASE/auth/login")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Wrong password login (should fail)" 401 "$http_code" "$content"

# Admin login
admin_login_data='{
    "username": "admin",
    "password": "admin123"
}'

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
    -H "Content-Type: application/json" \
    -d "$admin_login_data" \
    "$API_BASE/auth/login")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")

if [ "$http_code" -eq 200 ]; then
    ADMIN_TOKEN=$(echo "$content" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}Admin login successful, got token: ${ADMIN_TOKEN:0:20}...${NC}"
fi

run_test "Admin login" 200 "$http_code" "$content"

# 4. Protected interface test
echo -e "\n${BLUE}=== 4. Protected interface test ===${NC}"

if [ -n "$TOKEN" ]; then
    # Get personal information
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $TOKEN" \
        "$API_BASE/auth/profile")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Get personal information" 200 "$http_code" "$content"

    # Update personal information
    update_profile_data='{
        "full_name": "Updated Test User",
        "avatar": "https://example.com/avatar.jpg"
    }'

    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X PUT \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$update_profile_data" \
        "$API_BASE/auth/profile")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Update personal information" 200 "$http_code" "$content"

    # Change password
    change_password_data='{
        "old_password": "password123",
        "new_password": "newpassword123"
    }'

    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$change_password_data" \
        "$API_BASE/auth/change-password")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Change password" 200 "$http_code" "$content"

    # Get login logs
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $TOKEN" \
        "$API_BASE/auth/login-logs?page=1&page_size=10")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Get login logs" 200 "$http_code" "$content"

else
    echo -e "${YELLOW}Skip protected interface test (no valid token)${NC}"
fi

# 5. Unauthenticated access to protected interface (should fail)
echo -e "\n${BLUE}=== 5. Unauthenticated access test ===${NC}"

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
    "$API_BASE/auth/profile")
http_code="${response: -3}"
content=$(cat /tmp/response.json 2>/dev/null || echo "")
run_test "Unauthenticated access to personal information (should fail)" 401 "$http_code" "$content"

# 6. Admin functionality test
echo -e "\n${BLUE}=== 6. Admin functionality test ===${NC}"

if [ -n "$ADMIN_TOKEN" ]; then
    # Get user list
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$API_BASE/users?page=1&page_size=10")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Get user list" 200 "$http_code" "$content"

    # Create user
    create_user_data='{
        "username": "newuser",
        "email": "newuser@example.com",
        "password": "password123",
        "full_name": "New User",
        "role": "user",
        "status": "active"
    }'

    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$create_user_data" \
        "$API_BASE/users")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Create user" 201 "$http_code" "$content"

    # Get system stats
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$API_BASE/system/stats")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Get system stats" 200 "$http_code" "$content"

else
    echo -e "${YELLOW}Skip admin functionality test (no admin token)${NC}"
fi

# 7. Normal user access to admin interface (should fail)
echo -e "\n${BLUE}=== 7. Permission control test ===${NC}"

if [ -n "$TOKEN" ]; then
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $TOKEN" \
        "$API_BASE/users")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Normal user access to admin interface (should fail)" 403 "$http_code" "$content"
fi

# 8. Logout test
echo -e "\n${BLUE}=== 8. Logout test ===${NC}"

if [ -n "$TOKEN" ]; then
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST \
        -H "Authorization: Bearer $TOKEN" \
        "$API_BASE/auth/logout")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Logout" 200 "$http_code" "$content"

    # Unauthenticated access to protected interface after logout (should fail)
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X GET \
        -H "Authorization: Bearer $TOKEN" \
        "$API_BASE/auth/profile")
    http_code="${response: -3}"
    content=$(cat /tmp/response.json 2>/dev/null || echo "")
    run_test "Unauthenticated access to protected interface after logout (should fail)" 401 "$http_code" "$content"
fi

# 9. CORS test
echo -e "\n${BLUE}=== 9. CORS test

response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X OPTIONS \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Content-Type,Authorization" \
    "$API_BASE/auth/login")
http_code="${response: -3}"
run_test "CORS preflight request" 200 "$http_code"

# Clean up temporary files
rm -f /tmp/response.json

# Test result summary
echo -e "\n==============================================="
echo -e "${BLUE}Test result summary${NC}"
echo -e "==============================================="
echo -e "Total tests: $TOTAL_TESTS"
echo -e "Passed tests: $PASSED_TESTS"
echo -e "Failed tests: $((TOTAL_TESTS - PASSED_TESTS))"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi 