package ratelimiter

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLocalRateLimiter_Allow(t *testing.T) {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 5,    // burst of 5
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	// Should allow burst requests initially
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	// Next request should be rate limited
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request to be rate limited")
	}
}

func TestLocalRateLimiter_AllowN(t *testing.T) {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 10,   // burst of 10
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	// Should allow 5 requests
	allowed, err := limiter.AllowN(ctx, key, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected 5 requests to be allowed")
	}

	// Should allow another 5 requests (using up the burst)
	allowed, err = limiter.AllowN(ctx, key, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected another 5 requests to be allowed")
	}

	// Should not allow 1 more request
	allowed, err = limiter.AllowN(ctx, key, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request to be rate limited")
	}
}

func TestLocalRateLimiter_Wait(t *testing.T) {
	config := &Config{
		Rate:  100.0, // 100 requests per second (fast for testing)
		Burst: 1,     // burst of 1
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	// First request should not wait
	start := time.Now()
	err := limiter.Wait(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed > time.Millisecond*50 { // Allow some tolerance
		t.Fatalf("first request should not wait, but waited %v", elapsed)
	}

	// Second request should wait
	start = time.Now()
	err = limiter.Wait(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed = time.Since(start)
	if elapsed < time.Millisecond*5 { // Should wait at least a few milliseconds
		t.Fatalf("second request should wait, but only waited %v", elapsed)
	}
}

func TestLocalRateLimiter_WaitWithCancel(t *testing.T) {
	config := &Config{
		Rate:  1.0, // 1 request per second (slow for testing cancellation)
		Burst: 1,   // burst of 1
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	key := "test-key"

	// Use up the burst
	ctx := context.Background()
	err := limiter.Wait(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// This should timeout
	err = limiter.Wait(ctx, key)
	if err == nil {
		t.Fatal("expected context timeout error")
	}
	if err != context.DeadlineExceeded && !strings.Contains(err.Error(), "context deadline") {
		t.Fatalf("expected context timeout error, got: %v", err)
	}
}

func TestLocalRateLimiter_Reserve(t *testing.T) {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 3,    // burst of 3
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	// First few reservations should be immediate
	for i := 0; i < 3; i++ {
		reservation, err := limiter.Reserve(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reservation.OK {
			t.Fatalf("expected reservation %d to be OK", i+1)
		}
		if reservation.Delay > 0 {
			t.Fatalf("expected no delay for reservation %d, got %v", i+1, reservation.Delay)
		}
	}

	// Next reservation should have delay
	reservation, err := limiter.Reserve(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reservation.OK {
		t.Fatal("expected reservation to be OK even with delay")
	}
	if reservation.Delay == 0 {
		t.Fatal("expected delay for reservation after burst is exhausted")
	}
}

func TestLocalRateLimiter_MultipleKeys(t *testing.T) {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 2,    // burst of 2
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key1 := "test-key-1"
	key2 := "test-key-2"

	// Each key should have its own bucket
	for i := 0; i < 2; i++ {
		allowed1, err := limiter.Allow(ctx, key1)
		if err != nil {
			t.Fatalf("unexpected error for key1: %v", err)
		}
		if !allowed1 {
			t.Fatalf("expected request %d for key1 to be allowed", i+1)
		}

		allowed2, err := limiter.Allow(ctx, key2)
		if err != nil {
			t.Fatalf("unexpected error for key2: %v", err)
		}
		if !allowed2 {
			t.Fatalf("expected request %d for key2 to be allowed", i+1)
		}
	}

	// Both keys should now be rate limited
	allowed1, err := limiter.Allow(ctx, key1)
	if err != nil {
		t.Fatalf("unexpected error for key1: %v", err)
	}
	if allowed1 {
		t.Fatal("expected key1 to be rate limited")
	}

	allowed2, err := limiter.Allow(ctx, key2)
	if err != nil {
		t.Fatalf("unexpected error for key2: %v", err)
	}
	if allowed2 {
		t.Fatal("expected key2 to be rate limited")
	}
}

func TestLocalRateLimiter_Concurrent(t *testing.T) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 100,    // burst of 100
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "test-key"

	const numGoroutines = 10
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	allowedCount := make([]int, numGoroutines)

	// Launch multiple goroutines making requests concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				allowed, err := limiter.Allow(ctx, key)
				if err != nil {
					t.Errorf("unexpected error in goroutine %d: %v", goroutineID, err)
					return
				}
				if allowed {
					allowedCount[goroutineID]++
				}
			}
		}(i)
	}

	wg.Wait()

	// Count total allowed requests
	totalAllowed := 0
	for _, count := range allowedCount {
		totalAllowed += count
	}

	// Should allow at most burst requests initially
	if totalAllowed > config.Burst {
		t.Fatalf("expected at most %d requests to be allowed, got %d", config.Burst, totalAllowed)
	}

	// Should allow at least some requests (burst capacity)
	if totalAllowed < config.Burst/2 {
		t.Fatalf("expected at least %d requests to be allowed, got %d", config.Burst/2, totalAllowed)
	}
}

func TestLocalRateLimiter_Close(t *testing.T) {
	config := &Config{
		Rate:  10.0,
		Burst: 5,
	}

	limiter := NewLocalRateLimiter(config)

	// Should work before close
	ctx := context.Background()
	allowed, err := limiter.Allow(ctx, "test-key")
	if err != nil {
		t.Fatalf("unexpected error before close: %v", err)
	}
	if !allowed {
		t.Fatal("expected request to be allowed before close")
	}

	// Close the limiter
	err = limiter.Close()
	if err != nil {
		t.Fatalf("unexpected error during close: %v", err)
	}

	// Should still work after close (local limiter doesn't disable after close)
	allowed, err = limiter.Allow(ctx, "test-key")
	if err != nil {
		t.Fatalf("unexpected error after close: %v", err)
	}
}
