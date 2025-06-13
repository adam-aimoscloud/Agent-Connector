#!/usr/bin/env python3
"""
Data Flow API Client for Agent-Connector
Supports OpenAI and Dify compatible API requests with streaming capabilities
"""

import json
import time
import uuid
import requests
import argparse
import sys
from typing import Dict, List, Optional, Union, Iterator
from dataclasses import dataclass, asdict
from enum import Enum


class AgentType(Enum):
    """Agent type enumeration"""
    OPENAI = "openai"
    DIFY_CHAT = "dify-chat"
    DIFY_WORKFLOW = "dify-workflow"


class ResponseMode(Enum):
    """Response mode enumeration"""
    BLOCKING = "blocking"
    STREAMING = "streaming"


@dataclass
class ChatMessage:
    """Chat message structure"""
    role: str  # "system", "user", "assistant"
    content: str


@dataclass
class OpenAIRequest:
    """OpenAI compatible request structure"""
    messages: List[ChatMessage]
    model: str = "gpt-3.5-turbo"
    max_tokens: Optional[int] = None
    temperature: float = 0.7
    stream: bool = False


@dataclass
class DifyRequest:
    """Dify compatible request structure"""
    query: str
    user: str
    conversation_id: str = ""
    inputs: Dict = None
    response_mode: str = "blocking"

    def __post_init__(self):
        if self.inputs is None:
            self.inputs = {}


class DataFlowClient:
    """Data Flow API Client"""
    
    def __init__(self, base_url: str = "http://localhost:8082", api_key: str = None, user_id: str = None):
        """
        Initialize the client
        
        Args:
            base_url: Base URL of the data flow API
            api_key: API key for authentication
            user_id: User ID for rate limiting (optional)
        """
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.user_id = user_id
        self.session = requests.Session()
        
        # Disable proxies for localhost connections
        self.session.proxies = {'http': None, 'https': None}
        
        # Set default headers
        if self.api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {self.api_key}'
            })
        
        # Set user ID header for rate limiting
        if self.user_id:
            self.session.headers.update({
                'X-User-ID': self.user_id
            })
        
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': 'DataFlow-Python-Client/1.0'
        })
    
    def health_check(self) -> Dict:
        """Check API health status"""
        try:
            response = self.session.get(f"{self.base_url}/api/v1/health")
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            return {"error": str(e), "status": "unhealthy"}
    
    def get_service_info(self) -> Dict:
        """Get service information"""
        try:
            response = self.session.get(f"{self.base_url}/")
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            return {"error": str(e)}
    
    def chat_openai(self, agent_id: str, request: OpenAIRequest) -> Union[Dict, Iterator[Dict]]:
        """
        Send OpenAI compatible chat request
        
        Args:
            agent_id: Agent ID
            request: OpenAI request object
            
        Returns:
            Response dict or iterator for streaming
        """
        url = f"{self.base_url}/api/v1/openai/chat/completions?agent_id={agent_id}"
        
        # Convert dataclass to dict and handle nested objects
        data = asdict(request)
        
        if request.stream:
            return self._stream_request(url, data)
        else:
            return self._blocking_request(url, data)
    
    def chat_dify(self, agent_id: str, request: DifyRequest) -> Union[Dict, Iterator[Dict]]:
        """
        Send Dify compatible chat request
        
        Args:
            agent_id: Agent ID
            request: Dify request object
            
        Returns:
            Response dict or iterator for streaming
        """
        url = f"{self.base_url}/api/v1/dify/chat-messages?agent_id={agent_id}"
        
        # Convert dataclass to dict
        data = asdict(request)
        
        if request.response_mode == "streaming":
            return self._stream_request(url, data)
        else:
            return self._blocking_request(url, data)
    
    def chat_dify_workflow(self, agent_id: str, request: DifyRequest) -> Union[Dict, Iterator[Dict]]:
        """
        Send Dify workflow request
        
        Args:
            agent_id: Agent ID
            request: Dify request object
            
        Returns:
            Response dict or iterator for streaming
        """
        url = f"{self.base_url}/api/v1/dify/workflows/run"
        
        # Convert dataclass to dict
        data = asdict(request)
        # Add agent_id to the request data
        data['agent_id'] = agent_id
        
        if request.response_mode == "streaming":
            return self._stream_request(url, data)
        else:
            return self._blocking_request(url, data)
    
    def chat_universal(self, agent_id: str, data: Dict) -> Union[Dict, Iterator[Dict]]:
        """
        Send universal chat request (auto-detects format)
        
        Args:
            agent_id: Agent ID
            data: Request data dict
            
        Returns:
            Response dict or iterator for streaming
        """
        url = f"{self.base_url}/api/v1/chat"
        
        # Add agent_id to the request data
        data['agent_id'] = agent_id
        
        # Check if streaming is requested
        is_streaming = data.get('stream', False) or data.get('response_mode') == 'streaming'
        
        if is_streaming:
            return self._stream_request(url, data)
        else:
            return self._blocking_request(url, data)
    
    def _blocking_request(self, url: str, data: Dict) -> Dict:
        """Send blocking request"""
        try:
            response = self.session.post(url, json=data)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            return {
                "error": {
                    "type": "request_failed",
                    "message": str(e),
                    "status_code": getattr(e.response, 'status_code', None) if hasattr(e, 'response') else None
                }
            }
    
    def _stream_request(self, url: str, data: Dict) -> Iterator[Dict]:
        """Send streaming request"""
        try:
            response = self.session.post(url, json=data, stream=True)
            response.raise_for_status()
            
            for line in response.iter_lines():
                if line:
                    line = line.decode('utf-8')
                    if line.startswith('data: '):
                        try:
                            data_str = line[6:]  # Remove 'data: ' prefix
                            if data_str.strip() == '[DONE]':
                                break
                            yield json.loads(data_str)
                        except json.JSONDecodeError:
                            continue
                            
        except requests.exceptions.RequestException as e:
            yield {
                "error": {
                    "type": "stream_failed",
                    "message": str(e),
                    "status_code": getattr(e.response, 'status_code', None) if hasattr(e, 'response') else None
                }
            }


def create_dify_agent_example():
    """Example: Create a Dify agent request"""
    return DifyRequest(
        query="Hello! Can you help me understand how AI agents work?",
        user=f"user-{uuid.uuid4().hex[:8]}",
        conversation_id="",
        inputs={
            "context": "This is a test conversation with a Dify agent",
            "language": "en"
        },
        response_mode="blocking"
    )


def create_openai_agent_example():
    """Example: Create an OpenAI agent request"""
    return OpenAIRequest(
        messages=[
            ChatMessage(role="system", content="You are a helpful AI assistant."),
            ChatMessage(role="user", content="What are the benefits of using AI agents in business applications?")
        ],
        model="gpt-3.5-turbo",
        max_tokens=500,
        temperature=0.7,
        stream=False
    )


def main():
    """Main function for CLI usage"""
    parser = argparse.ArgumentParser(description="Data Flow API Client")
    parser.add_argument("--base-url", default="http://localhost:8082", help="Base URL of the API")
    parser.add_argument("--api-key", required=True, help="API key for authentication")
    parser.add_argument("--agent-id", help="Agent ID to use")
    parser.add_argument("--user-id", help="User ID for rate limiting")
    parser.add_argument("--type", choices=["openai", "dify-chat", "dify-workflow", "universal"], default="dify-chat", 
                       help="Request type")
    parser.add_argument("--stream", action="store_true", help="Enable streaming mode")
    parser.add_argument("--query", help="Query text (for Dify)")
    parser.add_argument("--message", help="Message content (for OpenAI)")
    parser.add_argument("--health", action="store_true", help="Check API health")
    parser.add_argument("--info", action="store_true", help="Get service info")
    
    args = parser.parse_args()
    
    # Initialize client
    client = DataFlowClient(base_url=args.base_url, api_key=args.api_key, user_id=args.user_id)
    
    # Health check
    if args.health:
        print("🏥 Checking API health...")
        health = client.health_check()
        print(json.dumps(health, indent=2))
        return
    
    # Service info
    if args.info:
        print("📊 Getting service information...")
        info = client.get_service_info()
        print(json.dumps(info, indent=2))
        return
    
    print(f"🤖 Testing {args.type.upper()} agent: {args.agent_id}")
    print(f"🔗 API Base URL: {args.base_url}")
    print(f"🔑 Using API Key: {args.api_key[:10]}...")
    print()
    
    try:
        if args.type == "dify-chat":
            # Dify Chat request
            query = args.query or "Hello! This is a test message from the Python client. Can you respond?"
            request = DifyRequest(
                query=query,
                user=f"python-client-{int(time.time())}",
                conversation_id="",
                inputs={"source": "python-client"},
                response_mode="streaming" if args.stream else "blocking"
            )
            
            print(f"📝 Dify Chat Request:")
            print(f"   Query: {request.query}")
            print(f"   User: {request.user}")
            print(f"   Mode: {request.response_mode}")
            print()
            
            response = client.chat_dify(args.agent_id, request)
            
        elif args.type == "dify-workflow":
            # Dify Workflow request
            query = args.query or "Hello! This is a test message from the Python client. Can you respond?"
            request = DifyRequest(
                query=query,
                user=f"python-client-{int(time.time())}",
                conversation_id="",
                inputs={"source": "python-client"},
                response_mode="streaming" if args.stream else "blocking"
            )
            
            print(f"📝 Dify Workflow Request:")
            print(f"   Query: {request.query}")
            print(f"   User: {request.user}")
            print(f"   Mode: {request.response_mode}")
            print()
            
            response = client.chat_dify_workflow(args.agent_id, request)
            
        elif args.type == "openai":
            # OpenAI request
            message = args.message or "Hello! This is a test message from the Python client. Can you respond?"
            request = OpenAIRequest(
                messages=[
                    ChatMessage(role="system", content="You are a helpful AI assistant."),
                    ChatMessage(role="user", content=message)
                ],
                model="gpt-3.5-turbo",
                max_tokens=500,
                temperature=0.7,
                stream=args.stream
            )
            
            print(f"📝 OpenAI Request:")
            print(f"   Messages: {len(request.messages)} messages")
            print(f"   Model: {request.model}")
            print(f"   Stream: {request.stream}")
            print()
            
            response = client.chat_openai(args.agent_id, request)
            
        else:  # universal
            # Universal request
            data = {
                "messages": [
                    {"role": "user", "content": args.message or "Hello from universal client!"}
                ],
                "model": "gpt-3.5-turbo",
                "stream": args.stream
            }
            
            print(f"📝 Universal Request:")
            print(f"   Data: {json.dumps(data, indent=2)}")
            print()
            
            response = client.chat_universal(args.agent_id, data)
        
        # Handle response
        if args.stream or (args.type in ["dify-chat", "dify-workflow"] and args.stream):
            print("🌊 Streaming Response:")
            print("-" * 50)
            full_answer = ""
            for chunk in response:
                if "error" in chunk:
                    print(f"❌ Error: {chunk['error']}")
                    break
                elif "event" in chunk and chunk["event"] == "done":
                    print("\n" + "-" * 50)
                    print("✅ Stream completed")
                    break
                elif "answer" in chunk:
                    # For Dify responses, extract and display the answer
                    answer_part = chunk["answer"]
                    if answer_part:
                        print(answer_part, end="", flush=True)
                        full_answer += answer_part
                elif "event" in chunk and chunk["event"] == "message_end":
                    # Show usage information if available
                    if "metadata" in chunk and "usage" in chunk["metadata"]:
                        usage = chunk["metadata"]["usage"]
                        print(f"\n\n📊 Usage: {usage.get('total_tokens', 'N/A')} tokens")
                        print(f"💰 Cost: {usage.get('total_price', 'N/A')} {usage.get('currency', '')}")
                else:
                    # For other chunk types or OpenAI format, show full JSON
                    print(f"📦 Chunk: {json.dumps(chunk, indent=2)}")
                    print("-" * 30)
        else:
            print("📦 Response:")
            print(json.dumps(response, indent=2))
            
    except KeyboardInterrupt:
        print("\n⏹️  Interrupted by user")
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main() 