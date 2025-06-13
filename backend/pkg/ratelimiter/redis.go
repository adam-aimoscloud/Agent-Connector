package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements RateLimiter interface using Redis for distributed rate limiting
type RedisRateLimiter struct {
	client *redis.Client
	rate   float64
	burst  int

	// Lua script for atomic token bucket operations
	tokenBucketScript *redis.Script
}

// Lua script for token bucket algorithm
// This script atomically checks and updates token count
const tokenBucketLuaScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

-- Get current state
local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1]) or burst
local last_refill = tonumber(bucket[2]) or now

-- Calculate tokens to add based on time elapsed
local elapsed = math.max(0, now - last_refill)
local tokens_to_add = elapsed * rate / 1000 -- rate is per second, elapsed is in milliseconds
tokens = math.min(burst, tokens + tokens_to_add)

-- Check if we have enough tokens
if tokens >= requested then
    tokens = tokens - requested
    -- Update bucket state
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 3600) -- Expire after 1 hour of inactivity
    return {1, tokens} -- allowed, remaining tokens
else
    -- Update bucket state even if not allowed (for accurate refill time)
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 3600)
    return {0, tokens} -- not allowed, remaining tokens
end
`

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(config *Config) (*RedisRateLimiter, error) {
	if config.Redis == nil {
		return nil, fmt.Errorf("Redis configuration is required")
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:            config.Redis.Addr,
		Password:        config.Redis.Password,
		DB:              config.Redis.DB,
		PoolSize:        config.Redis.PoolSize,
		MinIdleConns:    config.Redis.MinIdleConns,
		ConnMaxIdleTime: config.Redis.ConnMaxIdleTime,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRateLimiter{
		client:            client,
		rate:              config.Rate,
		burst:             config.Burst,
		tokenBucketScript: redis.NewScript(tokenBucketLuaScript),
	}, nil
}

// Allow checks if the request is allowed under the rate limit
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if n requests are allowed under the rate limit
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	now := time.Now().UnixMilli()

	result, err := r.tokenBucketScript.Run(ctx, r.client, []string{key},
		r.rate, r.burst, n, now).Result()

	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit check: %w", err)
	}

	results, ok := result.([]interface{})
	if !ok || len(results) != 2 {
		return false, fmt.Errorf("unexpected result format from Redis script")
	}

	allowed, ok := results[0].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected allowed value type")
	}

	return allowed == 1, nil
}

// Wait blocks until the request can be processed under the rate limit
func (r *RedisRateLimiter) Wait(ctx context.Context, key string) error {
	return r.WaitN(ctx, key, 1)
}

// WaitN blocks until n requests can be processed under the rate limit
func (r *RedisRateLimiter) WaitN(ctx context.Context, key string, n int) error {
	for {
		allowed, err := r.AllowN(ctx, key, n)
		if err != nil {
			return err
		}

		if allowed {
			return nil
		}

		// Calculate wait time based on rate
		waitTime := time.Duration(float64(n)/r.rate*1000) * time.Millisecond

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue to next iteration
		}
	}
}

// Reserve reserves a token and returns a reservation
func (r *RedisRateLimiter) Reserve(ctx context.Context, key string) (*Reservation, error) {
	return r.ReserveN(ctx, key, 1)
}

// ReserveN reserves n tokens and returns a reservation
func (r *RedisRateLimiter) ReserveN(ctx context.Context, key string, n int) (*Reservation, error) {
	allowed, err := r.AllowN(ctx, key, n)
	if err != nil {
		return nil, err
	}

	if allowed {
		return &Reservation{
			OK:    true,
			Delay: 0,
			cancel: func() error {
				// In a real implementation, you might want to return tokens
				// For simplicity, we'll just return nil
				return nil
			},
		}, nil
	}

	// Calculate delay based on rate
	delay := time.Duration(float64(n)/r.rate*1000) * time.Millisecond

	return &Reservation{
		OK:    false,
		Delay: delay,
		cancel: func() error {
			return nil
		},
	}, nil
}

// Close cleans up resources used by the rate limiter
func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

// GetTokens returns the current number of tokens for a key (for monitoring)
func (r *RedisRateLimiter) GetTokens(ctx context.Context, key string) (float64, error) {
	now := time.Now().UnixMilli()

	result, err := r.tokenBucketScript.Run(ctx, r.client, []string{key},
		r.rate, r.burst, 0, now).Result()

	if err != nil {
		return 0, fmt.Errorf("failed to get token count: %w", err)
	}

	results, ok := result.([]interface{})
	if !ok || len(results) != 2 {
		return 0, fmt.Errorf("unexpected result format from Redis script")
	}

	tokens, ok := results[1].(string)
	if !ok {
		// Try int64 format
		if tokensInt, ok := results[1].(int64); ok {
			return float64(tokensInt), nil
		}
		return 0, fmt.Errorf("unexpected tokens value type")
	}

	tokensFloat, err := strconv.ParseFloat(tokens, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse tokens value: %w", err)
	}

	return tokensFloat, nil
}
