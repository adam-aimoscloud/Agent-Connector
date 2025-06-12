# Agent Management Module

A comprehensive Go module for managing multiple AI agent sources with unified interfaces, load balancing, health checks, and robust configuration management.

## Features

### 🚀 Core Features
- **Unified Agent Interface**: Common interface for different agent types
- **Multiple Agent Sources**: Support for OpenAI Compatible APIs and Dify platform
- **Load Balancing**: Multiple strategies (Priority, Round Robin, Random, Weighted Random, Least Connections)
- **Health Monitoring**: Automated health checks with configurable thresholds
- **Configuration Management**: Fluent builders and preset configurations
- **Error Handling**: Comprehensive error types with retry policies
- **Metrics & Monitoring**: Built-in metrics collection and reporting

### 🛡️ Production Ready
- **Thread Safety**: Concurrent access support with proper synchronization
- **Resource Management**: Automatic cleanup and connection pooling
- **Timeouts & Retries**: Configurable timeout and retry mechanisms
- **Streaming Support**: Real-time streaming responses for compatible agents
- **Validation**: Comprehensive configuration validation

## Supported Agent Types

### OpenAI Compatible
- **OpenAI API**: Official OpenAI GPT models
- **Azure OpenAI**: Microsoft Azure OpenAI Service
- **Custom APIs**: Any OpenAI-compatible API endpoint
- **Features**: Chat completion, streaming, function calling, vision

### Dify Platform
- **Dify Chatbots**: Conversational AI applications
- **Dify Agents**: Advanced AI agents with tools
- **Features**: Chat completion, streaming, file uploads, conversation history

## Quick Start

### Basic Agent Creation

```go
package main

import (
    "context"
    "fmt"
    "agent-connector/pkg/agent"
)

func main() {
    // Create OpenAI agent
    openaiConfig := agent.NewOpenAIConfigBuilder().
        WithID("my-openai").
        WithName("My OpenAI Agent").
        WithBaseURL("https://api.openai.com").
        WithAPIKey("your-api-key").
        WithDefaultModel("gpt-3.5-turbo").
        Build()

    openaiAgent, err := agent.NewOpenAIAgent(openaiConfig)
    if err != nil {
        panic(err)
    }
    defer openaiAgent.Close()

    // Create Dify agent
    difyConfig := agent.NewDifyConfigBuilder().
        WithID("my-dify").
        WithName("My Dify Agent").
        WithBaseURL("https://api.dify.ai").
        WithAPIKey("your-dify-key").
        WithAppID("your-app-id").
        Build()

    difyAgent, err := agent.NewDifyAgent(difyConfig)
    if err != nil {
        panic(err)
    }
    defer difyAgent.Close()

    // Send a chat message
    request := &agent.ChatRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Hello, how are you?"},
        },
    }

    response, err := openaiAgent.Chat(context.Background(), request)
    if err != nil {
        panic(err)
    }

    fmt.Println("Response:", response.Choices[0].Message.Content)
}
```

### Agent Manager with Load Balancing

```go
package main

import (
    "context"
    "agent-connector/pkg/agent"
)

func main() {
    // Create agent manager
    manager, err := agent.NewAgentManager(&agent.AgentManagerConfig{
        LoadBalancingStrategy: agent.Priority,
        EnableHealthChecks:    true,
        HealthCheckInterval:   30 * time.Second,
    })
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    // Register multiple agents
    agent1, _ := agent.NewOpenAIAgent(openaiConfig1)
    agent2, _ := agent.NewOpenAIAgent(openaiConfig2)
    agent3, _ := agent.NewDifyAgent(difyConfig)

    manager.RegisterAgent(agent1)
    manager.RegisterAgent(agent2)
    manager.RegisterAgent(agent3)

    // Get available agent automatically
    request := &agent.ChatRequest{
        Messages: []agent.Message{
            {Role: "user", Content: "Hello!"},
        },
    }

    selectedAgent, err := manager.GetAvailableAgent(context.Background(), request)
    if err != nil {
        panic(err)
    }

    response, err := selectedAgent.Chat(context.Background(), request)
    if err != nil {
        panic(err)
    }

    fmt.Println("Response from", selectedAgent.GetName(), ":", response.Choices[0].Message.Content)
}
```

## Configuration

### OpenAI Configuration

```go
config := agent.NewOpenAIConfigBuilder().
    WithID("my-agent").
    WithName("My Agent").
    WithBaseURL("https://api.openai.com").
    WithAPIKey("sk-...").
    WithDefaultModel("gpt-4").
    WithMaxTokens(8192).
    WithTemperature(0.7).
    WithTimeout(30 * time.Second).
    WithPriority(100).
    WithRetryPolicy(&agent.RetryPolicy{
        MaxRetries:   3,
        InitialDelay: 1 * time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
    }).
    WithHealthCheck(&agent.HealthCheckConfig{
        Enabled:          true,
        Interval:         1 * time.Minute,
        Timeout:          10 * time.Second,
        FailureThreshold: 3,
        SuccessThreshold: 1,
    }).
    Build()
```

### Dify Configuration

```go
config := agent.NewDifyConfigBuilder().
    WithID("my-dify-agent").
    WithName("My Dify Agent").
    WithBaseURL("https://api.dify.ai").
    WithAPIKey("app-...").
    WithAppID("your-app-id").
    WithAppType("agent").
    WithVersion("v1").
    WithPriority(80).
    WithLogging(true).
    WithAutoGenerateTitle(true).
    Build()
```

### Preset Configurations

```go
presets := agent.NewPresetConfigs()

// OpenAI presets
gpt35Config := presets.OpenAIGPT35Turbo("gpt35", "GPT-3.5", "your-key")
gpt4Config := presets.OpenAIGPT4("gpt4", "GPT-4", "your-key")
azureConfig := presets.AzureOpenAI("azure", "Azure OpenAI", "https://your-resource.openai.azure.com", "your-key", "deployment-name")

// Dify presets
chatbotConfig := presets.DifyChatbot("chatbot", "My Chatbot", "https://api.dify.ai", "your-key", "app-id")
agentConfig := presets.DifyAgent("agent", "My Agent", "https://api.dify.ai", "your-key", "app-id")
```

## Load Balancing Strategies

### Priority-based (Default)
Selects the agent with the highest priority value.

```go
config := &agent.AgentManagerConfig{
    LoadBalancingStrategy: agent.Priority,
}
```

### Round Robin
Distributes requests evenly across all healthy agents.

```go
config := &agent.AgentManagerConfig{
    LoadBalancingStrategy: agent.RoundRobin,
}
```

### Random
Randomly selects from available healthy agents.

```go
config := &agent.AgentManagerConfig{
    LoadBalancingStrategy: agent.Random,
}
```

### Weighted Random
Randomly selects based on agent priority weights.

```go
config := &agent.AgentManagerConfig{
    LoadBalancingStrategy: agent.WeightedRandom,
}
```

### Least Connections
Selects the agent with the fewest active connections.

```go
config := &agent.AgentManagerConfig{
    LoadBalancingStrategy: agent.LeastConnections,
}
```

## Streaming Support

```go
// Start streaming chat
streamResponse, err := agent.ChatStream(ctx, request)
if err != nil {
    panic(err)
}
defer streamResponse.Stream.Close()

// Read streaming events
for event := range streamResponse.Events {
    switch event.Type {
    case "content":
        fmt.Print(event.Delta.Content)
    case "finish":
        fmt.Println("\nStream finished:", *event.FinishReason)
        return
    case "error":
        fmt.Println("Error:", event.Data)
        return
    }
}

// Check for errors
select {
case err := <-streamResponse.Errors:
    if err != nil {
        fmt.Println("Stream error:", err)
    }
default:
}
```

## Health Monitoring

```go
// Get agent status
status, err := agent.GetStatus(ctx)
if err != nil {
    fmt.Printf("Failed to get status: %v\n", err)
} else {
    fmt.Printf("Agent: %s\n", status.AgentID)
    fmt.Printf("Status: %s\n", status.Status)
    fmt.Printf("Health: %v\n", status.Health)
    fmt.Printf("Response Time: %dms\n", status.ResponseTime)
    fmt.Printf("Success Rate: %.2f%%\n", status.SuccessRate)
}

// Get metrics
metrics, err := manager.GetAgentMetrics(ctx, "agent-id")
if err != nil {
    fmt.Printf("Failed to get metrics: %v\n", err)
} else {
    fmt.Printf("Requests: %d\n", metrics.RequestCount)
    fmt.Printf("Errors: %d\n", metrics.ErrorCount)
    fmt.Printf("Success Rate: %.2f%%\n", metrics.SuccessRate)
}
```

## Error Handling

```go
response, err := agent.Chat(ctx, request)
if err != nil {
    if agentErr, ok := err.(*agent.AgentError); ok {
        fmt.Printf("Agent Error: %s (%s)\n", agentErr.Message, agentErr.Code)
        switch agentErr.Code {
        case "rate_limit_exceeded":
            // Handle rate limiting
        case "invalid_api_key":
            // Handle authentication error
        default:
            // Handle other errors
        }
    } else {
        // Handle other types of errors
        fmt.Printf("General error: %v\n", err)
    }
}
```

## Advanced Usage

### Custom Retry Policy

```go
retryPolicy := agent.NewRetryPolicyBuilder().
    WithMaxRetries(5).
    WithInitialDelay(2 * time.Second).
    WithMaxDelay(60 * time.Second).
    WithMultiplier(2.0).
    WithRetryableErrors([]string{
        "timeout",
        "rate_limit_exceeded",
        "service_unavailable",
    }).
    Build()

config := agent.NewOpenAIConfigBuilder().
    // ... other config
    WithRetryPolicy(retryPolicy).
    Build()
```

### Custom Health Check

```go
healthCheck := agent.NewHealthCheckConfigBuilder().
    WithEnabled(true).
    WithInterval(30 * time.Second).
    WithTimeout(10 * time.Second).
    WithFailureThreshold(3).
    WithSuccessThreshold(1).
    Build()

config := agent.NewOpenAIConfigBuilder().
    // ... other config
    WithHealthCheck(healthCheck).
    Build()
```

### Function Calling (OpenAI)

```go
request := &agent.ChatRequest{
    Messages: []agent.Message{
        {Role: "user", Content: "What's the weather like in San Francisco?"},
    },
    Tools: []agent.Tool{
        {
            Type: "function",
            Function: agent.Function{
                Name:        "get_weather",
                Description: "Get current weather for a location",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "location": map[string]interface{}{
                            "type": "string",
                            "description": "The city and state",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
}

response, err := agent.Chat(ctx, request)
if err != nil {
    panic(err)
}

// Check for function calls
for _, choice := range response.Choices {
    if len(choice.Message.ToolCalls) > 0 {
        for _, toolCall := range choice.Message.ToolCalls {
            fmt.Printf("Function: %s\n", toolCall.Function.Name)
            fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)
        }
    }
}
```

## Testing

Run the test suite:

```bash
# Run all tests
./scripts/test-agent.sh

# Run specific tests
go test ./pkg/agent/...

# Run with coverage
go test -cover ./pkg/agent/...

# Run benchmarks
go test -bench=. ./pkg/agent/...

# Run race detection
go test -race ./pkg/agent/...
```

## Demo Application

Build and run the demo:

```bash
# Build demo
go build -o bin/agent-demo ./cmd/agent-demo/

# Set environment variables (optional)
export OPENAI_API_KEY="your-openai-key"
export DIFY_API_KEY="your-dify-key"
export DIFY_BASE_URL="https://api.dify.ai"
export DIFY_APP_ID="your-app-id"

# Run demo
./bin/agent-demo
```

The demo showcases:
1. Basic agent creation
2. Agent manager with load balancing
3. Configuration builders
4. Preset configurations
5. Load balancing strategies

## API Reference

### Interfaces

#### Agent Interface
```go
type Agent interface {
    GetID() string
    GetName() string
    GetType() AgentType
    GetCapabilities() AgentCapabilities
    Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, request *ChatRequest) (*ChatStreamResponse, error)
    GetModels(ctx context.Context) ([]Model, error)
    ValidateConfig() error
    GetStatus(ctx context.Context) (*AgentStatus, error)
    Close() error
}
```

#### AgentManager Interface
```go
type AgentManager interface {
    RegisterAgent(agent Agent) error
    UnregisterAgent(agentID string) error
    GetAgent(agentID string) (Agent, error)
    ListAgents() []Agent
    ListAgentsByType(agentType AgentType) []Agent
    GetAvailableAgent(ctx context.Context, request *ChatRequest) (Agent, error)
    Close() error
}
```

### Types

#### AgentType
```go
const (
    AgentTypeOpenAI AgentType = "openai"
    AgentTypeDify   AgentType = "dify"
)
```

#### ChatRequest
```go
type ChatRequest struct {
    Messages    []Message               `json:"messages"`
    Model       string                  `json:"model,omitempty"`
    Temperature *float32                `json:"temperature,omitempty"`
    MaxTokens   *int                    `json:"max_tokens,omitempty"`
    Stream      bool                    `json:"stream,omitempty"`
    Functions   []Function              `json:"functions,omitempty"`
    Tools       []Tool                  `json:"tools,omitempty"`
    UserID      string                  `json:"user_id,omitempty"`
    SessionID   string                  `json:"session_id,omitempty"`
    Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}
```

#### ChatResponse
```go
type ChatResponse struct {
    ID       string                 `json:"id"`
    Object   string                 `json:"object"`
    Created  int64                  `json:"created"`
    Model    string                 `json:"model"`
    Choices  []Choice               `json:"choices"`
    Usage    *Usage                 `json:"usage,omitempty"`
    Error    *AgentError            `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`./scripts/test-agent.sh`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Roadmap

- [ ] Additional agent sources (Anthropic Claude, Google PaLM, etc.)
- [ ] Advanced metrics and monitoring
- [ ] Circuit breaker pattern implementation
- [ ] Request caching and deduplication
- [ ] WebSocket support for real-time communication
- [ ] Plugin system for custom agent implementations
- [ ] Distributed agent management across multiple nodes 