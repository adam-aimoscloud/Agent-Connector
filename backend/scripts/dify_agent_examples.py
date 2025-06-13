#!/usr/bin/env python3
"""
Dify Agent Examples for Agent-Connector
Demonstrates various ways to interact with Dify agents through the Data Flow API
"""

import json
import time
import uuid
from dataflow_client import DataFlowClient, DifyRequest, ChatMessage, OpenAIRequest


def example_dify_simple_chat():
    """Example 1: Simple Dify chat interaction"""
    print("🤖 Example 1: Simple Dify Chat")
    print("=" * 50)
    
    # Initialize client
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="your-api-key-here"  # Replace with actual API key
    )
    
    # Create Dify request
    request = DifyRequest(
        query="Hello! Can you introduce yourself and explain what you can do?",
        user=f"example-user-{int(time.time())}",
        conversation_id="",
        inputs={
            "context": "This is a demonstration of Dify agent integration",
            "language": "en",
            "session_id": f"session-{uuid.uuid4().hex[:8]}"
        },
        response_mode="blocking"
    )
    
    print(f"📝 Request: {request.query}")
    print(f"👤 User: {request.user}")
    print()
    
    # Send request
    response = client.chat_dify("your-agent-id", request)  # Replace with actual agent ID
    
    print("📦 Response:")
    print(json.dumps(response, indent=2))
    print()


def example_dify_streaming_chat():
    """Example 2: Dify streaming chat"""
    print("🌊 Example 2: Dify Streaming Chat")
    print("=" * 50)
    
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="your-api-key-here"
    )
    
    request = DifyRequest(
        query="Please tell me a short story about AI agents working together to solve problems.",
        user=f"streaming-user-{int(time.time())}",
        conversation_id="",
        inputs={
            "story_length": "short",
            "theme": "collaboration",
            "tone": "optimistic"
        },
        response_mode="streaming"
    )
    
    print(f"📝 Streaming Request: {request.query}")
    print("🌊 Streaming Response:")
    print("-" * 30)
    
    try:
        for chunk in client.chat_dify("your-agent-id", request):
            if "error" in chunk:
                print(f"❌ Error: {chunk['error']}")
                break
            else:
                # Extract answer from Dify response
                if "answer" in chunk:
                    print(chunk["answer"], end="", flush=True)
                else:
                    print(f"\n📦 Chunk: {json.dumps(chunk, indent=2)}")
    except KeyboardInterrupt:
        print("\n⏹️ Streaming interrupted")
    
    print("\n" + "-" * 30)
    print()


def example_dify_conversation():
    """Example 3: Multi-turn Dify conversation"""
    print("💬 Example 3: Multi-turn Dify Conversation")
    print("=" * 50)
    
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="your-api-key-here"
    )
    
    # Start conversation
    conversation_id = ""
    user_id = f"conversation-user-{int(time.time())}"
    
    messages = [
        "Hello! I'm interested in learning about machine learning.",
        "Can you explain what supervised learning is?",
        "What about unsupervised learning? How is it different?",
        "Thank you! Can you recommend some beginner-friendly resources?"
    ]
    
    for i, message in enumerate(messages, 1):
        print(f"👤 Turn {i}: {message}")
        
        request = DifyRequest(
            query=message,
            user=user_id,
            conversation_id=conversation_id,
            inputs={
                "turn": i,
                "topic": "machine_learning",
                "level": "beginner"
            },
            response_mode="blocking"
        )
        
        response = client.chat_dify("your-agent-id", request)
        
        if "error" not in response:
            # Extract conversation ID for next turn
            if "conversation_id" in response:
                conversation_id = response["conversation_id"]
            
            print(f"🤖 Response: {response.get('answer', 'No answer field')}")
        else:
            print(f"❌ Error: {response['error']}")
            break
        
        print("-" * 30)
        time.sleep(1)  # Be nice to the API
    
    print()


def example_dify_with_custom_inputs():
    """Example 4: Dify with custom inputs and context"""
    print("⚙️ Example 4: Dify with Custom Inputs")
    print("=" * 50)
    
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="your-api-key-here"
    )
    
    # Complex request with custom inputs
    request = DifyRequest(
        query="Analyze the following business scenario and provide recommendations.",
        user=f"business-analyst-{int(time.time())}",
        conversation_id="",
        inputs={
            "scenario": "A small e-commerce company wants to implement AI chatbots",
            "budget": "limited",
            "timeline": "3 months",
            "team_size": "5 people",
            "technical_expertise": "medium",
            "priorities": ["customer_service", "cost_reduction", "scalability"],
            "constraints": ["budget", "timeline", "technical_resources"],
            "analysis_type": "comprehensive",
            "output_format": "structured_recommendations"
        },
        response_mode="blocking"
    )
    
    print(f"📝 Business Query: {request.query}")
    print(f"📊 Custom Inputs: {json.dumps(request.inputs, indent=2)}")
    print()
    
    response = client.chat_dify("your-agent-id", request)
    
    print("📦 Analysis Response:")
    print(json.dumps(response, indent=2))
    print()


def example_health_and_info():
    """Example 5: Check API health and service info"""
    print("🏥 Example 5: Health Check and Service Info")
    print("=" * 50)
    
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="your-api-key-here"
    )
    
    # Health check
    print("🏥 Checking API health...")
    health = client.health_check()
    print(f"Health Status: {json.dumps(health, indent=2)}")
    print()
    
    # Service info
    print("📊 Getting service information...")
    info = client.get_service_info()
    print(f"Service Info: {json.dumps(info, indent=2)}")
    print()


def example_error_handling():
    """Example 6: Error handling scenarios"""
    print("⚠️ Example 6: Error Handling")
    print("=" * 50)
    
    client = DataFlowClient(
        base_url="http://localhost:8082",
        api_key="invalid-api-key"  # Intentionally invalid
    )
    
    request = DifyRequest(
        query="This should fail due to invalid API key",
        user="error-test-user",
        response_mode="blocking"
    )
    
    print("🔑 Testing with invalid API key...")
    response = client.chat_dify("test-agent", request)
    
    if "error" in response:
        print(f"✅ Expected error caught: {response['error']}")
    else:
        print(f"❓ Unexpected success: {response}")
    
    print()


def main():
    """Run all examples"""
    print("🚀 Dify Agent Examples for Agent-Connector")
    print("=" * 60)
    print()
    
    print("⚠️  IMPORTANT: Before running these examples:")
    print("1. Make sure the Data Flow API is running on port 8082")
    print("2. Replace 'your-api-key-here' with your actual API key")
    print("3. Replace 'your-agent-id' with your actual Dify agent ID")
    print("4. Ensure your Dify agent is properly configured and enabled")
    print()
    
    examples = [
        example_health_and_info,
        example_dify_simple_chat,
        example_dify_streaming_chat,
        example_dify_conversation,
        example_dify_with_custom_inputs,
        example_error_handling
    ]
    
    for i, example_func in enumerate(examples, 1):
        try:
            example_func()
        except Exception as e:
            print(f"❌ Example {i} failed: {e}")
        
        if i < len(examples):
            print("⏳ Waiting 2 seconds before next example...")
            time.sleep(2)
            print()


if __name__ == "__main__":
    main() 