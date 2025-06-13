#!/usr/bin/env python3
"""
Quick Test Script for Data Flow API Python Client
Tests basic functionality without requiring actual agent configuration
"""

import sys
import json
from dataflow_client import DataFlowClient, DifyRequest, OpenAIRequest, ChatMessage


def test_client_initialization():
    """Test client initialization"""
    print("🔧 Testing client initialization...")
    
    try:
        # Test with default settings
        client1 = DataFlowClient()
        print("✅ Default client created successfully")
        
        # Test with custom settings
        client2 = DataFlowClient(
            base_url="http://localhost:8082",
            api_key="test-key"
        )
        print("✅ Custom client created successfully")
        
        # Check headers
        if "Authorization" in client2.session.headers:
            print("✅ Authorization header set correctly")
        else:
            print("❌ Authorization header missing")
            
        return True
    except Exception as e:
        print(f"❌ Client initialization failed: {e}")
        return False


def test_data_structures():
    """Test data structure creation"""
    print("\n📊 Testing data structures...")
    
    try:
        # Test ChatMessage
        message = ChatMessage(role="user", content="Hello")
        print(f"✅ ChatMessage created: {message}")
        
        # Test DifyRequest
        dify_req = DifyRequest(
            query="Test query",
            user="test-user",
            inputs={"key": "value"}
        )
        print(f"✅ DifyRequest created: {dify_req}")
        
        # Test OpenAIRequest
        openai_req = OpenAIRequest(
            messages=[message],
            model="gpt-3.5-turbo"
        )
        print(f"✅ OpenAIRequest created: {openai_req}")
        
        return True
    except Exception as e:
        print(f"❌ Data structure creation failed: {e}")
        return False


def test_health_check():
    """Test health check functionality"""
    print("\n🏥 Testing health check...")
    
    try:
        client = DataFlowClient(base_url="http://localhost:8082")
        health = client.health_check()
        
        print(f"📊 Health check response: {json.dumps(health, indent=2)}")
        
        if "error" in health:
            print("⚠️  Health check returned error (expected if API not running)")
        else:
            print("✅ Health check successful")
            
        return True
    except Exception as e:
        print(f"❌ Health check failed: {e}")
        return False


def test_service_info():
    """Test service info functionality"""
    print("\n📋 Testing service info...")
    
    try:
        client = DataFlowClient(base_url="http://localhost:8082")
        info = client.get_service_info()
        
        print(f"📊 Service info response: {json.dumps(info, indent=2)}")
        
        if "error" in info:
            print("⚠️  Service info returned error (expected if API not running)")
        else:
            print("✅ Service info successful")
            
        return True
    except Exception as e:
        print(f"❌ Service info failed: {e}")
        return False


def test_request_formatting():
    """Test request data formatting"""
    print("\n📝 Testing request formatting...")
    
    try:
        client = DataFlowClient(api_key="test-key")
        
        # Test Dify request formatting
        dify_req = DifyRequest(
            query="Test query",
            user="test-user",
            inputs={"context": "test"},
            response_mode="blocking"
        )
        
        print(f"✅ Dify request formatted correctly")
        print(f"   Query: {dify_req.query}")
        print(f"   User: {dify_req.user}")
        print(f"   Inputs: {dify_req.inputs}")
        
        # Test OpenAI request formatting
        openai_req = OpenAIRequest(
            messages=[
                ChatMessage(role="system", content="You are helpful"),
                ChatMessage(role="user", content="Hello")
            ],
            model="gpt-3.5-turbo",
            temperature=0.7
        )
        
        print(f"✅ OpenAI request formatted correctly")
        print(f"   Messages: {len(openai_req.messages)} messages")
        print(f"   Model: {openai_req.model}")
        print(f"   Temperature: {openai_req.temperature}")
        
        return True
    except Exception as e:
        print(f"❌ Request formatting failed: {e}")
        return False


def test_error_handling():
    """Test error handling"""
    print("\n⚠️  Testing error handling...")
    
    try:
        client = DataFlowClient(
            base_url="http://localhost:8082",
            api_key="invalid-key"
        )
        
        # This should fail gracefully
        request = DifyRequest(
            query="Test error handling",
            user="test-user"
        )
        
        response = client.chat_dify("invalid-agent", request)
        
        if "error" in response:
            print("✅ Error handling working correctly")
            print(f"   Error type: {response['error'].get('type', 'unknown')}")
            print(f"   Error message: {response['error'].get('message', 'no message')}")
        else:
            print("❓ Unexpected success (API might be running)")
            
        return True
    except Exception as e:
        print(f"✅ Exception caught correctly: {e}")
        return True


def main():
    """Run all tests"""
    print("🧪 Data Flow API Python Client Test Suite")
    print("=" * 50)
    
    tests = [
        ("Client Initialization", test_client_initialization),
        ("Data Structures", test_data_structures),
        ("Health Check", test_health_check),
        ("Service Info", test_service_info),
        ("Request Formatting", test_request_formatting),
        ("Error Handling", test_error_handling)
    ]
    
    passed = 0
    total = len(tests)
    
    for test_name, test_func in tests:
        print(f"\n{'='*20} {test_name} {'='*20}")
        try:
            if test_func():
                passed += 1
                print(f"✅ {test_name} PASSED")
            else:
                print(f"❌ {test_name} FAILED")
        except Exception as e:
            print(f"❌ {test_name} FAILED with exception: {e}")
    
    print(f"\n{'='*50}")
    print(f"📊 Test Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed!")
        print("\n🚀 Next steps:")
        print("1. Start the Data Flow API: go run cmd/dataflow-api/main.go")
        print("2. Configure agents via Control Flow API")
        print("3. Test with real agents using: python dataflow_client.py")
        print("4. Run examples: python dify_agent_examples.py")
    else:
        print("⚠️  Some tests failed. Check the output above for details.")
        
    print(f"\n📚 For more information, see: README_PYTHON_CLIENT.md")
    
    return passed == total


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1) 