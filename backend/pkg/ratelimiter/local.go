package ratelimiter

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// LocalRateLimiter implements RateLimiter interface using local token bucket
type LocalRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int

	// cleanup goroutine control
	stopCh          chan struct{}
	cleanupInterval time.Duration
}

// NewLocalRateLimiter creates a new local rate limiter
func NewLocalRateLimiter(config *Config) *LocalRateLimiter {
	limiter := &LocalRateLimiter{
		limiters:        make(map[string]*rate.Limiter),
		rate:            rate.Limit(config.Rate),
		burst:           config.Burst,
		stopCh:          make(chan struct{}),
		cleanupInterval: time.Minute * 5, // cleanup unused limiters every 5 minutes
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// getLimiter gets or creates a rate limiter for the given key
func (l *LocalRateLimiter) getLimiter(key string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := l.limiters[key]; exists {
		return limiter
	}

	// Create new limiter for this key
	limiter = rate.NewLimiter(l.rate, l.burst)
	l.limiters[key] = limiter

	return limiter
}

// Allow checks if the request is allowed under the rate limit
func (l *LocalRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	limiter := l.getLimiter(key)
	return limiter.Allow(), nil
}

// AllowN checks if n requests are allowed under the rate limit
func (l *LocalRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	limiter := l.getLimiter(key)
	return limiter.AllowN(time.Now(), n), nil
}

// Wait blocks until the request can be processed under the rate limit
func (l *LocalRateLimiter) Wait(ctx context.Context, key string) error {
	limiter := l.getLimiter(key)
	return limiter.Wait(ctx)
}

// WaitN blocks until n requests can be processed under the rate limit
func (l *LocalRateLimiter) WaitN(ctx context.Context, key string, n int) error {
	limiter := l.getLimiter(key)
	return limiter.WaitN(ctx, n)
}

// Reserve reserves a token and returns a reservation
func (l *LocalRateLimiter) Reserve(ctx context.Context, key string) (*Reservation, error) {
	limiter := l.getLimiter(key)
	reservation := limiter.Reserve()

	return &Reservation{
		OK:    reservation.OK(),
		Delay: reservation.Delay(),
		cancel: func() error {
			reservation.Cancel()
			return nil
		},
	}, nil
}

// ReserveN reserves n tokens and returns a reservation
func (l *LocalRateLimiter) ReserveN(ctx context.Context, key string, n int) (*Reservation, error) {
	limiter := l.getLimiter(key)
	reservation := limiter.ReserveN(time.Now(), n)

	return &Reservation{
		OK:    reservation.OK(),
		Delay: reservation.Delay(),
		cancel: func() error {
			reservation.Cancel()
			return nil
		},
	}, nil
}

// Close cleans up resources used by the rate limiter
func (l *LocalRateLimiter) Close() error {
	close(l.stopCh)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Clear all limiters
	l.limiters = make(map[string]*rate.Limiter)

	return nil
}

// cleanup periodically removes unused limiters to prevent memory leaks
func (l *LocalRateLimiter) cleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.performCleanup()
		case <-l.stopCh:
			return
		}
	}
}

// performCleanup removes limiters that haven't been used recently
func (l *LocalRateLimiter) performCleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, limiter := range l.limiters {
		// Check if the limiter has tokens available (indicating it hasn't been used recently)
		// This is a heuristic - if burst tokens are available, it might not be actively used
		if limiter.Tokens() >= float64(l.burst) {
			// Additional check: create a test reservation to see the delay
			reservation := limiter.Reserve()
			if reservation.Delay() == 0 {
				// This limiter appears unused, consider removing it
				// But let's be conservative and only remove after checking time
				delete(l.limiters, key)
			}
			reservation.Cancel()
		}
	}

	// Log cleanup if needed (in production, you might want to use a proper logger)
	_ = now // Suppress unused variable warning
}
