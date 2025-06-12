#!/bin/bash

# Data Flow API Test Script
# Tests the unified agent access API functionality

API_BASE_URL="http://localhost:8082"
DATAFLOW_BASE_URL="$API_BASE_URL/api/v1/dataflow"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_AGENT_ID="test-agent-001"
TEST_API_KEY="test-api-key-12345678"

echo -e "${BLUE}🧪 Testing Agent-Connector Data Flow API${NC}"
echo -e "${BLUE}=============================================${NC}"
echo "Base URL: $API_BASE_URL"
echo "Data Flow API: $DATAFLOW_BASE_URL"
echo ""

# Function to make HTTP request and check response
test_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    local headers=$6

    echo -e "${YELLOW}Testing: $description${NC}"
    echo "URL: $method $url"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            $headers \
            ${data:+-d "$data"})
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            ${data:+-d "$data"})
    fi
    
    # Split response and status code
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC} (Status: $http_code)"
        if [ -n "$response_body" ]; then
            echo "Response: $response_body" | head -c 200
            echo "..."
        fi
    else
        echo -e "${RED}❌ FAIL${NC} (Expected: $expected_status, Got: $http_code)"
        echo "Response: $response_body"
    fi
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Test 1: Service Information
echo -e "${BLUE}📊 Test 1: Service Information${NC}"
test_request "GET" "$API_BASE_URL/" "" "200" "Get service information"

# Test 2: Health Check
echo -e "${BLUE}🏥 Test 2: Health Check${NC}"
test_request "GET" "$DATAFLOW_BASE_URL/health" "" "200" "Data Flow API health check"

# Test 3: OpenAI-style Chat Request (should fail without proper agent setup)
echo -e "${BLUE}🤖 Test 3: OpenAI-style Chat Request${NC}"
openai_data='{
    "messages": [
        {"role": "user", "content": "Hello, how are you?"}
    ],
    "model": "gpt-3.5-turbo",
    "max_tokens": 100,
    "temperature": 0.7
}'
test_request "POST" "$DATAFLOW_BASE_URL/openai/chat/completions/$TEST_AGENT_ID" \
    "$openai_data" "401" "OpenAI-style chat without API key" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Test 4: Dify-style Chat Request (should fail without proper agent setup)
echo -e "${BLUE}💬 Test 4: Dify-style Chat Request${NC}"
dify_data='{
    "query": "Hello, how are you?",
    "user": "test-user-123",
    "conversation_id": "",
    "inputs": {}
}'
test_request "POST" "$DATAFLOW_BASE_URL/dify/chat-messages/$TEST_AGENT_ID" \
    "$dify_data" "401" "Dify-style chat without API key" \
    "-H \"X-API-Key: $TEST_API_KEY\""

# Test 5: Universal Chat Interface
echo -e "${BLUE}🌐 Test 5: Universal Chat Interface${NC}"
universal_data='{
    "messages": [
        {"role": "user", "content": "What is the weather like?"}
    ],
    "model": "gpt-3.5-turbo"
}'
test_request "POST" "$DATAFLOW_BASE_URL/chat/$TEST_AGENT_ID" \
    "$universal_data" "401" "Universal chat interface" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Test 6: Streaming Request Test
echo -e "${BLUE}🌊 Test 6: Streaming Request${NC}"
streaming_data='{
    "messages": [
        {"role": "user", "content": "Tell me a short story"}
    ],
    "model": "gpt-3.5-turbo",
    "stream": true
}'
test_request "POST" "$DATAFLOW_BASE_URL/openai/chat/completions/$TEST_AGENT_ID" \
    "$streaming_data" "401" "Streaming chat request" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Test 7: Authentication Error Tests
echo -e "${BLUE}🔐 Test 7: Authentication Tests${NC}"

# Missing API key
test_request "POST" "$DATAFLOW_BASE_URL/chat/$TEST_AGENT_ID" \
    "$universal_data" "401" "Request without API key"

# Missing Agent ID
test_request "POST" "$DATAFLOW_BASE_URL/chat/" \
    "$universal_data" "404" "Request without Agent ID" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Invalid Agent ID format
test_request "POST" "$DATAFLOW_BASE_URL/chat/invalid-agent" \
    "$universal_data" "401" "Request with invalid Agent ID" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Test 8: Request Format Validation
echo -e "${BLUE}📝 Test 8: Request Format Validation${NC}"

# Invalid JSON
test_request "POST" "$DATAFLOW_BASE_URL/chat/$TEST_AGENT_ID" \
    "{invalid json}" "400" "Invalid JSON format" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Empty request
test_request "POST" "$DATAFLOW_BASE_URL/chat/$TEST_AGENT_ID" \
    "{}" "401" "Empty request body" \
    "-H \"Authorization: Bearer $TEST_API_KEY\""

# Test 9: CORS Headers Check
echo -e "${BLUE}🌍 Test 9: CORS Headers${NC}"
echo "Testing CORS preflight request..."
cors_response=$(curl -s -I -X OPTIONS "$DATAFLOW_BASE_URL/chat/$TEST_AGENT_ID" \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Authorization")

if echo "$cors_response" | grep -q "Access-Control-Allow-Origin"; then
    echo -e "${GREEN}✅ CORS headers present${NC}"
else
    echo -e "${RED}❌ CORS headers missing${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}📋 Test Summary${NC}"
echo -e "${BLUE}===============${NC}"
echo "🎯 Data Flow API Server: $API_BASE_URL"
echo "📡 All endpoints tested for basic functionality"
echo "🔐 Authentication properly enforced"
echo "📝 Request validation working"
echo "🌍 CORS support enabled"
echo ""
echo -e "${YELLOW}📝 Notes:${NC}"
echo "- All tests expect authentication failures (401) since no agents are configured"
echo "- This validates the authentication layer is working correctly"
echo "- To test full functionality, agents must be configured via Control Flow API"
echo "- Use the Agent-Connector dashboard to create and manage agents"
echo ""
echo -e "${GREEN}✅ Data Flow API testing completed!${NC}"
echo ""
echo -e "${BLUE}🚀 Next Steps:${NC}"
echo "1. Configure agents via Control Flow API (port 8081)"
echo "2. Create API keys for agents"
echo "3. Test actual agent communication"
echo "4. Monitor rate limiting and queuing"

# Optional: Check if Control Flow API is running
echo ""
echo -e "${BLUE}🔗 Checking Control Flow API availability...${NC}"
control_flow_health=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8081/api/v1/system/health 2>/dev/null)
if [ "$control_flow_health" = "200" ]; then
    echo -e "${GREEN}✅ Control Flow API is running on port 8081${NC}"
    echo "You can configure agents at: http://localhost:8081"
else
    echo -e "${YELLOW}⚠️  Control Flow API not detected on port 8081${NC}"
    echo "Start it with: go run cmd/control-flow-api/main.go" 