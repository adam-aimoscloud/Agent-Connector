# Priority Queue Package

A Redis-based distributed priority queue implementation for agent request management in Go.

## Overview

This package provides a robust, scalable priority queue system designed for managing agent requests with different priority levels. It uses Redis as the backing store for distributed operation across multiple service instances.

## Features

- **Priority-based Queueing**: Requests are processed based on priority levels (Critical > Highest > High > Normal > Low > Lowest)
- **Distributed Architecture**: Redis-based implementation for multi-instance deployments
- **Request Expiration**: TTL support for request expiration
- **Atomic Operations**: Lua scripts ensure atomic enqueue/dequeue operations
- **Queue Size Limits**: Configurable maximum queue sizes
- **Flexible Queue Naming**: Hierarchical queue naming for different agents/services
- **Comprehensive Testing**: Thorough unit tests and benchmarks
- **Type Safety**: Strong typing with request builders and validation

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "agent-connector/pkg/queue"
)

func main() {
    // Configure Redis connection
    config := queue.DefaultQueueConfig()
    config.Redis = queue.DefaultRedisQueueConfig("localhost:6379")
    config.MaxQueueSize = 1000
    
    // Create priority queue
    q, err := queue.NewPriorityQueue(queue.RedisType, config)
    if err != nil {
        log.Fatal(err)
    }
    defer q.Close()
    
    ctx := context.Background()
    queueName := "agent:gpt-4"
    
    // Create a request using builder pattern
    request, err := queue.NewRequestBuilder().
        WithID("req-001").
        WithUserID("user-alice").
        WithAgentID("gpt-4").
        WithPriority(queue.PriorityHigh).
        WithPayload("Hello, can you help me?").
        WithTTL(5 * time.Minute).
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Enqueue the request
    if err := q.Enqueue(ctx, queueName, request); err != nil {
        log.Fatal(err)
    }
    
    // Dequeue the highest priority request
    dequeuedRequest, err := q.Dequeue(ctx, queueName)
    if err != nil {
        log.Fatal(err)
    }
    
    if dequeuedRequest != nil {
        fmt.Printf("Processing request: %s from user: %s\n", 
            dequeuedRequest.ID, dequeuedRequest.UserID)
    }
}
```

## Priority Levels

The package supports six priority levels:

| Priority | Value | Description |
|----------|-------|-------------|
| `PriorityCritical` | 1000 | Emergency/system critical requests |
| `PriorityHighest` | 100 | Highest priority user requests |
| `PriorityHigh` | 75 | High priority requests |
| `PriorityNormal` | 50 | Normal priority requests (default) |
| `PriorityLow` | 25 | Low priority requests |
| `PriorityLowest` | 0 | Lowest priority requests |

### Priority String Conversion

```go
// Convert string to priority
priority, err := queue.PriorityFromString("high")
if err != nil {
    log.Fatal(err)
}

// Get priority string representation
fmt.Println(priority.String()) // Output: "High"

// Check if priority is valid
if priority.IsValid() {
    fmt.Println("Priority is valid")
}
```

## Request Builder

Use the fluent request builder for creating well-formed requests:

```go
request, err := queue.NewRequestBuilder().
    WithID("unique-request-id").
    WithUserID("user-123").
    WithAgentID("agent-456").
    WithPriority(queue.PriorityHigh).
    WithPayload(map[string]interface{}{
        "message": "Hello world",
        "context": "chat",
    }).
    WithMetadata("session_id", "sess-789").
    WithMetadata("source", "web-ui").
    WithTTL(10 * time.Minute).
    Build()

if err != nil {
    log.Fatal(err)
}
```

## Queue Naming

Use the queue name builder for consistent naming:

```go
// Simple agent queue
queueName := queue.NewQueueNameBuilder().
    WithAgent("gpt-4").
    Build()
// Result: "agent:gpt-4"

// Complex hierarchical queue
queueName := queue.NewQueueNameBuilder().
    WithService("translation").
    WithRegion("us-west-1").
    WithAgent("translator-v2").
    WithCustom("language", "spanish").
    Build()
// Result: "service:translation:region:us-west-1:agent:translator-v2:language:spanish"
```

## Configuration

### Default Configuration

```go
config := queue.DefaultQueueConfig()
config.Redis = queue.DefaultRedisQueueConfig("localhost:6379")
```

### Custom Configuration

```go
config := &queue.QueueConfig{
    Redis: &queue.RedisConfig{
        Addr:            "redis.example.com:6379",
        Password:        "secret",
        DB:              2,
        PoolSize:        20,
        MinIdleConns:    5,
        ConnMaxIdleTime: 30 * time.Minute,
        KeyPrefix:       "myapp",
    },
    DefaultTTL:    7200, // 2 hours
    MaxQueueSize:  5000, // Max 5000 requests per queue
    EnableMetrics: true,
}
```

## API Reference

### PriorityQueue Interface

```go
type PriorityQueue interface {
    Enqueue(ctx context.Context, queueName string, request *Request) error
    Dequeue(ctx context.Context, queueName string) (*Request, error)
    DequeueWithTimeout(ctx context.Context, queueName string, timeout time.Duration) (*Request, error)
    Peek(ctx context.Context, queueName string) (*Request, error)
    Size(ctx context.Context, queueName string) (int64, error)
    Remove(ctx context.Context, queueName string, requestID string) error
    UpdatePriority(ctx context.Context, queueName string, requestID string, newPriority Priority) error
    ListByPriority(ctx context.Context, queueName string, offset, limit int64) ([]*Request, error)
    Clear(ctx context.Context, queueName string) error
    Close() error
}
```

### Request Structure

```go
type Request struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id"`
    AgentID   string                 `json:"agent_id"`
    Priority  Priority               `json:"priority"`
    Payload   interface{}            `json:"payload"`
    Metadata  map[string]interface{} `json:"metadata"`
    CreatedAt time.Time              `json:"created_at"`
    ExpiresAt *time.Time             `json:"expires_at,omitempty"`
}
```

## Advanced Usage

### Blocking Dequeue with Timeout

```go
// Wait up to 30 seconds for a request
request, err := q.DequeueWithTimeout(ctx, queueName, 30*time.Second)
if err != nil {
    log.Fatal(err)
}

if request == nil {
    fmt.Println("No requests available within timeout")
} else {
    fmt.Printf("Got request: %s\n", request.ID)
}
```

### Queue Management

```go
// Get queue size
size, err := q.Size(ctx, queueName)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Queue has %d requests\n", size)

// List requests by priority (pagination)
requests, err := q.ListByPriority(ctx, queueName, 0, 10) // First 10 requests
if err != nil {
    log.Fatal(err)
}

for _, req := range requests {
    fmt.Printf("Request %s: Priority=%s, User=%s\n", 
        req.ID, req.Priority.String(), req.UserID)
}

// Update request priority
err = q.UpdatePriority(ctx, queueName, "req-123", queue.PriorityCritical)
if err != nil {
    log.Fatal(err)
}

// Remove specific request
err = q.Remove(ctx, queueName, "req-456")
if err != nil {
    log.Fatal(err)
}

// Clear entire queue
err = q.Clear(ctx, queueName)
if err != nil {
    log.Fatal(err)
}
```

## Testing

Run the test suite:

```bash
# Run all tests
./scripts/test-queue.sh

# Run specific tests
go test -v ./pkg/queue/

# Run with coverage
go test -cover ./pkg/queue/

# Run benchmarks
go test -bench=. ./pkg/queue/

# Race detection
go test -race ./pkg/queue/
```

## Performance Considerations

1. **Redis Connection Pooling**: Configure appropriate pool sizes based on your load
2. **Queue Partitioning**: Use different queue names to distribute load
3. **Batch Operations**: Consider batching multiple operations when possible
4. **TTL Management**: Set appropriate TTLs to prevent queue bloat
5. **Monitoring**: Enable metrics to monitor queue sizes and performance

## Error Handling

The package provides detailed error messages for common scenarios:

- Invalid configurations
- Redis connection issues
- Queue size limits exceeded
- Request validation failures
- Priority out of range

## Integration Example

Here's how to integrate the priority queue with a rate limiter:

```go
// Rate limiter from the ratelimiter package
rateLimiter, _ := ratelimiter.NewRedisRateLimiter(rateLimiterConfig)

// Priority queue
priorityQueue, _ := queue.NewPriorityQueue(queue.RedisType, queueConfig)

// Process requests with rate limiting
func processRequests(ctx context.Context) {
    for {
        // Dequeue next request
        request, err := priorityQueue.DequeueWithTimeout(ctx, "agent:gpt-4", 5*time.Second)
        if err != nil || request == nil {
            continue
        }
        
        // Check rate limit for user
        if !rateLimiter.Allow(ctx, "user:"+request.UserID) {
            // Re-queue with lower priority or handle rate limiting
            request.Priority = queue.PriorityLow
            priorityQueue.Enqueue(ctx, "agent:gpt-4", request)
            continue
        }
        
        // Process the request
        go processRequest(request)
    }
}
```

## License

This package is part of the Agent Connector project and follows the same licensing terms. 