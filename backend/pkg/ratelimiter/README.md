# Rate Limiter Package

A high-performance, distributed rate limiter package for Go applications using Redis as the backend storage.

## Features

- **Distributed Rate Limiting**: Uses Redis for distributed rate limiting across multiple application instances
- **Token Bucket Algorithm**: Implements the token bucket algorithm for smooth rate limiting
- **Per-Key Rate Limiting**: Independent rate limits for different keys (users, IPs, API endpoints, etc.)
- **Flexible Configuration**: Configurable rate and burst parameters
- **Context Support**: Full context.Context support for cancellation and timeouts
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Lua Script Optimization**: Uses Redis Lua scripts for atomic operations

## Installation

```bash
go get agent-connector/pkg/ratelimiter
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "agent-connector/pkg/ratelimiter"
)

func main() {
    // Create configuration
    config := &ratelimiter.Config{
        Rate:  10.0, // 10 requests per second
        Burst: 20,   // Allow burst of 20 requests
        Redis: &ratelimiter.RedisConfig{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
        },
    }

    // Create rate limiter
    limiter, err := ratelimiter.NewRateLimiter(ratelimiter.RedisType, config)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    ctx := context.Background()
    key := "user:123"

    // Check if request is allowed
    allowed, err := limiter.Allow(ctx, key)
    if err != nil {
        log.Fatal(err)
    }

    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request rate limited")
    }
}
```

## Configuration

### Config Structure

```go
type Config struct {
    Rate  float64      // Requests per second
    Burst int          // Maximum burst size
    Redis *RedisConfig // Redis configuration (required)
}

type RedisConfig struct {
    Addr            string        // Redis server address
    Password        string        // Redis password (optional)
    DB              int           // Redis database number
    PoolSize        int           // Connection pool size
    MinIdleConns    int           // Minimum idle connections
    ConnMaxIdleTime time.Duration // Connection max idle time
}
```

### Default Configuration

```go
// Get default configuration
config := ratelimiter.DefaultConfig("localhost:6379")

// Or create custom configuration
config := &ratelimiter.Config{
    Rate:  100.0, // 100 requests per second
    Burst: 200,   // Allow burst of 200 requests
    Redis: &ratelimiter.RedisConfig{
        Addr:            "localhost:6379",
        Password:        "",
        DB:              0,
        PoolSize:        10,
        MinIdleConns:    2,
        ConnMaxIdleTime: 30 * time.Minute,
    },
}
```

## Usage Examples

### Basic Usage

```go
// Create rate limiter
limiter, err := ratelimiter.NewRateLimiter(ratelimiter.RedisType, config)
if err != nil {
    log.Fatal(err)
}
defer limiter.Close()

ctx := context.Background()

// Check if request is allowed
allowed, err := limiter.Allow(ctx, "user:123")
if err != nil {
    log.Fatal(err)
}

if !allowed {
    // Handle rate limit exceeded
    return
}

// Process request
```

### Waiting for Rate Limit

```go
// Wait until request is allowed (blocks if necessary)
err := limiter.Wait(ctx, "user:123")
if err != nil {
    log.Fatal(err)
}

// Request is now allowed, proceed
```

### Reservation Pattern

```go
// Make a reservation
reservation, err := limiter.Reserve(ctx, "user:123")
if err != nil {
    log.Fatal(err)
}

if !reservation.OK {
    // Rate limit would be exceeded
    return
}

if reservation.Delay > 0 {
    // Wait for the required delay
    time.Sleep(reservation.Delay)
}

// Proceed with request
```

### Multiple Tokens

```go
// Request multiple tokens at once
allowed, err := limiter.AllowN(ctx, "user:123", 5)
if err != nil {
    log.Fatal(err)
}

if !allowed {
    // Not enough tokens available
    return
}
```

## Rate Limiting Strategies

### Per-User Rate Limiting

```go
userID := "user:123"
allowed, err := limiter.Allow(ctx, userID)
```

### Per-IP Rate Limiting

```go
clientIP := "192.168.1.100"
allowed, err := limiter.Allow(ctx, clientIP)
```

### Per-API Endpoint Rate Limiting

```go
endpoint := "api:/users"
allowed, err := limiter.Allow(ctx, endpoint)
```

### Combined Rate Limiting

```go
// Rate limit per user per endpoint
key := fmt.Sprintf("user:%s:endpoint:%s", userID, endpoint)
allowed, err := limiter.Allow(ctx, key)
```

## Error Handling

```go
allowed, err := limiter.Allow(ctx, "user:123")
if err != nil {
    // Handle different types of errors
    switch {
    case errors.Is(err, context.Canceled):
        // Context was canceled
    case errors.Is(err, context.DeadlineExceeded):
        // Context deadline exceeded
    default:
        // Other errors (Redis connection issues, etc.)
        log.Printf("Rate limiter error: %v", err)
    }
    return
}
```

## Best Practices

1. **Use Meaningful Keys**: Choose descriptive keys that clearly identify what is being rate limited
2. **Handle Errors Gracefully**: Always check for errors and have fallback behavior
3. **Set Appropriate Timeouts**: Use context with timeouts for Redis operations
4. **Monitor Redis Health**: Ensure Redis is healthy and accessible
5. **Choose Appropriate Rates**: Set rate limits based on your application's capacity
6. **Use Connection Pooling**: Configure appropriate pool sizes for your load

## Performance Considerations

- **Redis Performance**: Rate limiter performance depends on Redis performance
- **Network Latency**: Consider network latency between your application and Redis
- **Connection Pooling**: Use appropriate connection pool settings
- **Lua Scripts**: The package uses optimized Lua scripts for atomic operations

## Testing

Run the tests:

```bash
# Run all tests (requires Redis)
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks (requires Redis)
go test -bench=. ./...
```

## Redis Requirements

- Redis 2.6.0 or later (for Lua script support)
- Stable network connection between application and Redis
- Sufficient Redis memory for storing rate limit data

## Thread Safety

The rate limiter is thread-safe and can be used concurrently from multiple goroutines.

## License

This package is part of the Agent-Connector project. 