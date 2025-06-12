package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Example demonstrates basic usage of the local rate limiter
func ExampleLocalRateLimiter() {
	// Create a rate limiter config
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 5,    // Allow burst of 5 requests
	}

	// Create a local rate limiter
	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "user:123"

	// Check if request is allowed
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		log.Printf("Error checking rate limit: %v", err)
		return
	}

	if allowed {
		fmt.Println("Request allowed")
	} else {
		fmt.Println("Request rate limited")
	}

	// Output: Request allowed
}

// Example demonstrates usage with different rate limiter types
func ExampleNewRateLimiter() {
	// Create a local rate limiter
	config := &Config{
		Rate:  100.0, // 100 requests per second
		Burst: 10,    // Allow burst of 10 requests
	}

	limiter, err := NewRateLimiter(LocalType, config)
	if err != nil {
		log.Printf("Error creating rate limiter: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	key := "api:endpoint:/users"

	// Check multiple requests
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if allowed {
			fmt.Printf("Request %d: allowed\n", i+1)
		} else {
			fmt.Printf("Request %d: rate limited\n", i+1)
		}
	}

	// Output:
	// Request 1: allowed
	// Request 2: allowed
	// Request 3: allowed
}

// Example demonstrates waiting for rate limit
func ExampleRateLimiter_Wait() {
	config := &Config{
		Rate:  2.0, // 2 requests per second
		Burst: 1,   // Allow burst of 1 request
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "slow-api"

	fmt.Println("Starting requests...")
	start := time.Now()

	// First request should be immediate
	err := limiter.Wait(ctx, key)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Request 1 completed after %v\n", time.Since(start))

	// Second request should wait
	err = limiter.Wait(ctx, key)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Request 2 completed after %v\n", time.Since(start))

	// Output will show the second request taking longer due to rate limiting
}

// Example demonstrates reservation pattern
func ExampleRateLimiter_Reserve() {
	config := &Config{
		Rate:  5.0, // 5 requests per second
		Burst: 2,   // Allow burst of 2 requests
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "reservation-example"

	// Make a reservation
	reservation, err := limiter.Reserve(ctx, key)
	if err != nil {
		log.Printf("Error making reservation: %v", err)
		return
	}

	if reservation.OK {
		if reservation.Delay > 0 {
			fmt.Printf("Reservation made, need to wait %v\n", reservation.Delay)
			time.Sleep(reservation.Delay)
		}
		fmt.Println("Proceeding with request")
	} else {
		fmt.Println("Reservation failed")
	}

	// Output: Proceeding with request
}

// Example demonstrates multiple keys
func ExampleRateLimiter_MultipleKeys() {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 3,    // Allow burst of 3 requests
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()

	// Different keys have independent rate limits
	keys := []string{"user:alice", "user:bob", "user:charlie"}

	for _, key := range keys {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			log.Printf("Error for %s: %v", key, err)
			continue
		}

		if allowed {
			fmt.Printf("%s: allowed\n", key)
		} else {
			fmt.Printf("%s: rate limited\n", key)
		}
	}

	// Output:
	// user:alice: allowed
	// user:bob: allowed
	// user:charlie: allowed
}

// Example demonstrates validation and default configuration
func ExampleValidateConfig() {
	// Valid configuration
	config := &Config{
		Rate:  50.0,
		Burst: 100,
	}

	err := ValidateConfig(config)
	if err != nil {
		fmt.Printf("Config validation failed: %v\n", err)
	} else {
		fmt.Println("Config is valid")
	}

	// Invalid configuration
	invalidConfig := &Config{
		Rate:  -1.0, // Invalid rate
		Burst: 10,
	}

	err = ValidateConfig(invalidConfig)
	if err != nil {
		fmt.Printf("Invalid config detected: %v\n", err)
	}

	// Using default config
	defaultConfig := DefaultConfig()
	fmt.Printf("Default config: Rate=%.1f, Burst=%d\n", defaultConfig.Rate, defaultConfig.Burst)

	// Output:
	// Config is valid
	// Invalid config detected: rate must be positive, got: -1.000000
	// Default config: Rate=10.0, Burst=20
}
