# Rate Limiter Module

A distributed token bucket rate limiter implementation for Go, supporting both local and Redis-based distributed rate limiting.

## Features

- **Token Bucket Algorithm**: Implements the token bucket algorithm for smooth rate limiting with burst capability
- **Local Rate Limiting**: In-memory rate limiting for single-instance applications
- **Distributed Rate Limiting**: Redis-based rate limiting for multi-instance applications
- **Multiple Keys**: Support for per-key rate limiting (e.g., per-user, per-API endpoint)
- **Context Support**: All operations support Go context for cancellation and timeouts
- **High Performance**: Optimized for high throughput with minimal allocations
- **Memory Efficient**: Automatic cleanup of unused limiters to prevent memory leaks
- **Thread Safe**: All operations are safe for concurrent use

## Quick Start

### Local Rate Limiter

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
        Burst: 5,    // Allow burst of 5 requests
    }
    
    // Create local rate limiter
    limiter := ratelimiter.NewLocalRateLimiter(config)
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

### Distributed Rate Limiter (Redis)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "agent-connector/pkg/ratelimiter"
)

func main() {
    // Create configuration with Redis
    config := &ratelimiter.Config{
        Rate:  100.0, // 100 requests per second
        Burst: 20,    // Allow burst of 20 requests
        Redis: &ratelimiter.RedisConfig{
            Addr:            "localhost:6379",
            Password:        "",
            DB:              0,
            PoolSize:        10,
            MinIdleConns:    2,
            ConnMaxIdleTime: 30 * time.Minute,
        },
    }
    
    // Create distributed rate limiter
    limiter, err := ratelimiter.NewRedisRateLimiter(config)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()
    
    ctx := context.Background()
    key := "api:/users"
    
    // Check multiple requests
    allowed, err := limiter.AllowN(ctx, key, 5)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("5 requests allowed: %v\n", allowed)
}
```

### Using the Factory Function

```go
package main

import (
    "context"
    "log"
    
    "agent-connector/pkg/ratelimiter"
)

func main() {
    config := ratelimiter.DefaultConfig()
    
    // Create local rate limiter
    limiter, err := ratelimiter.NewRateLimiter(ratelimiter.LocalType, config)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()
    
    // Use the limiter
    ctx := context.Background()
    allowed, _ := limiter.Allow(ctx, "test-key")
    
    if allowed {
        // Process request
    }
}
```

## API Reference

### RateLimiter Interface

The main interface for rate limiting operations:

```go
type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    AllowN(ctx context.Context, key string, n int) (bool, error)
    Wait(ctx context.Context, key string) error
    WaitN(ctx context.Context, key string, n int) error
    Reserve(ctx context.Context, key string) (*Reservation, error)
    ReserveN(ctx context.Context, key string, n int) (*Reservation, error)
    Close() error
}
```

#### Methods

- **Allow(ctx, key)**: Check if a single request is allowed
- **AllowN(ctx, key, n)**: Check if n requests are allowed
- **Wait(ctx, key)**: Block until a request can be processed
- **WaitN(ctx, key, n)**: Block until n requests can be processed
- **Reserve(ctx, key)**: Reserve a token and get a reservation
- **ReserveN(ctx, key, n)**: Reserve n tokens and get a reservation
- **Close()**: Clean up resources

### Configuration

```go
type Config struct {
    Rate  float64      // Requests per second
    Burst int          // Maximum burst size
    Redis *RedisConfig // Redis configuration (optional)
}

type RedisConfig struct {
    Addr            string        // Redis server address
    Password        string        // Redis password
    DB              int           // Redis database number
    PoolSize        int           // Maximum connections in pool
    MinIdleConns    int           // Minimum idle connections
    ConnMaxIdleTime time.Duration // Maximum idle time for connections
}
```

### Factory Functions

```go
// Create a rate limiter with specified type
func NewRateLimiter(limiterType RateLimiterType, config *Config) (RateLimiter, error)

// Create a local rate limiter directly
func NewLocalRateLimiter(config *Config) *LocalRateLimiter

// Create a Redis rate limiter directly
func NewRedisRateLimiter(config *Config) (*RedisRateLimiter, error)

// Validate configuration
func ValidateConfig(config *Config) error

// Get default configuration
func DefaultConfig() *Config

// Get default Redis configuration
func DefaultRedisConfig(addr string) *RedisConfig
```

## Usage Patterns

### Per-User Rate Limiting

```go
// Rate limit per user ID
userID := "user123"
key := fmt.Sprintf("user:%s", userID)
allowed, err := limiter.Allow(ctx, key)
```

### Per-API Endpoint Rate Limiting

```go
// Rate limit per API endpoint
endpoint := "/api/users"
key := fmt.Sprintf("endpoint:%s", endpoint)
allowed, err := limiter.Allow(ctx, key)
```

### Combined Rate Limiting

```go
// Rate limit per user per endpoint
userID := "user123"
endpoint := "/api/users"
key := fmt.Sprintf("user:%s:endpoint:%s", userID, endpoint)
allowed, err := limiter.Allow(ctx, key)
```

### Graceful Degradation

```go
// Wait with timeout for rate limit
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := limiter.Wait(ctx, key)
if err == context.DeadlineExceeded {
    // Handle timeout - maybe return cached data or error
    return handleRateLimitTimeout()
}
```

### Reservation Pattern

```go
// Reserve tokens in advance
reservation, err := limiter.ReserveN(ctx, key, 3)
if err != nil {
    return err
}

if reservation.Delay > 0 {
    // Wait for the reservation to be ready
    time.Sleep(reservation.Delay)
}

// Process the request
if reservation.OK {
    return processRequest()
}
```

## Performance

The rate limiter is designed for high performance:

- **Local**: >1M ops/sec per core for single key, >10M ops/sec for multiple keys
- **Redis**: ~50K ops/sec depending on Redis performance and network latency
- **Memory**: Minimal allocations, automatic cleanup of unused limiters
- **Concurrency**: Lock-free hot path for local limiter, atomic operations for Redis

## Testing

Run the test suite:

```bash
cd backend
go test -v ./pkg/ratelimiter/...
```

Run benchmarks:

```bash
cd backend
go test -bench=. -benchmem ./pkg/ratelimiter/...
```

Generate coverage report:

```bash
cd backend
go test -coverprofile=coverage.out ./pkg/ratelimiter/...
go tool cover -html=coverage.out
```

## Examples

See the `example_test.go` file for complete examples and usage patterns.

## License

This module is part of the Agent Connector project. 