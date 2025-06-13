package ratelimiter

import (
	"fmt"
)

// RateLimiterType represents the type of rate limiter
type RateLimiterType string

const (
	// RedisType uses Redis for distributed rate limiting
	RedisType RateLimiterType = "redis"
)

// NewRateLimiter creates a new rate limiter based on the configuration
func NewRateLimiter(limiterType RateLimiterType, config *Config) (RateLimiter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Rate <= 0 {
		return nil, fmt.Errorf("rate must be positive")
	}

	if config.Burst <= 0 {
		return nil, fmt.Errorf("burst must be positive")
	}

	switch limiterType {
	case RedisType:
		return NewRedisRateLimiter(config)

	default:
		return nil, fmt.Errorf("unsupported rate limiter type: %s", limiterType)
	}
}

// ValidateConfig validates the rate limiter configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Rate <= 0 {
		return fmt.Errorf("rate must be positive, got: %f", config.Rate)
	}

	if config.Burst <= 0 {
		return fmt.Errorf("burst must be positive, got: %d", config.Burst)
	}

	// Redis configuration is required
	if config.Redis == nil {
		return fmt.Errorf("Redis configuration is required")
	}

	if config.Redis.Addr == "" {
		return fmt.Errorf("Redis address cannot be empty")
	}

	if config.Redis.PoolSize <= 0 {
		config.Redis.PoolSize = 10 // Set default pool size
	}

	if config.Redis.MinIdleConns < 0 {
		return fmt.Errorf("MinIdleConns cannot be negative, got: %d", config.Redis.MinIdleConns)
	}

	return nil
}

// DefaultConfig returns a default rate limiter configuration with Redis
func DefaultConfig(redisAddr string) *Config {
	return &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 20,   // burst of 20 requests
		Redis: DefaultRedisConfig(redisAddr),
	}
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig(addr string) *RedisConfig {
	return &RedisConfig{
		Addr:            addr,
		Password:        "",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnMaxIdleTime: 30 * 60 * 1000 * 1000 * 1000, // 30 minutes in nanoseconds
	}
}
