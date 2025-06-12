package ratelimiter

import (
	"context"
	"time"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if the request is allowed under the rate limit
	// Returns true if allowed, false if rate limited
	Allow(ctx context.Context, key string) (bool, error)

	// AllowN checks if n requests are allowed under the rate limit
	AllowN(ctx context.Context, key string, n int) (bool, error)

	// Wait blocks until the request can be processed under the rate limit
	// Returns error if context is cancelled or deadline exceeded
	Wait(ctx context.Context, key string) error

	// WaitN blocks until n requests can be processed under the rate limit
	WaitN(ctx context.Context, key string, n int) error

	// Reserve reserves a token and returns a reservation
	Reserve(ctx context.Context, key string) (*Reservation, error)

	// ReserveN reserves n tokens and returns a reservation
	ReserveN(ctx context.Context, key string, n int) (*Reservation, error)

	// Close cleans up resources used by the rate limiter
	Close() error
}

// Reservation represents a reserved token
type Reservation struct {
	// OK indicates whether the reservation is valid
	OK bool

	// Delay is the time to wait before the reservation can be used
	Delay time.Duration

	// cancel is used to cancel the reservation
	cancel func() error
}

// Cancel cancels the reservation and returns the token back to the bucket
func (r *Reservation) Cancel() error {
	if r.cancel != nil {
		return r.cancel()
	}
	return nil
}

// Config represents the configuration for rate limiter
type Config struct {
	// Rate is the number of tokens added per second
	Rate float64

	// Burst is the maximum number of tokens in the bucket
	Burst int

	// Redis configuration for distributed rate limiting
	Redis *RedisConfig
}

// RedisConfig represents Redis configuration for distributed rate limiting
type RedisConfig struct {
	// Addr is the Redis server address
	Addr string

	// Password is the Redis password
	Password string

	// DB is the Redis database number
	DB int

	// PoolSize is the maximum number of connections in the pool
	PoolSize int

	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int

	// ConnMaxIdleTime is the maximum idle time for connections
	ConnMaxIdleTime time.Duration
}
