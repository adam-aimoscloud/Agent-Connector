package ratelimiter

import (
	"fmt"
)

// Note: The following examples require a running Redis instance
// They are commented out to avoid test failures in CI/CD environments
// Uncomment and run with a Redis instance for testing

/*
// Example demonstrates basic usage of the Redis rate limiter
func Example() {
	// Create a rate limiter config with Redis backend
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 5,    // Allow burst of 5 requests
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	// Create a Redis rate limiter
	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		log.Printf("Error creating rate limiter: %v", err)
		return
	}
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

// Example demonstrates usage with Redis rate limiter
func Example_newRateLimiter() {
	// Create a Redis rate limiter
	config := &Config{
		Rate:  100.0, // 100 requests per second
		Burst: 10,    // Allow burst of 10 requests
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
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

// Example demonstrates waiting for rate limit with Redis
func Example_wait() {
	config := &Config{
		Rate:  2.0, // 2 requests per second
		Burst: 1,   // Allow burst of 1 request
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		log.Printf("Error creating rate limiter: %v", err)
		return
	}
	defer limiter.Close()

	ctx := context.Background()
	key := "slow-api"

	fmt.Println("Starting requests...")
	start := time.Now()

	// First request should be immediate
	err = limiter.Wait(ctx, key)
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

// Example demonstrates reservation pattern with Redis
func Example_reserve() {
	config := &Config{
		Rate:  5.0, // 5 requests per second
		Burst: 2,   // Allow burst of 2 requests
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		log.Printf("Error creating rate limiter: %v", err)
		return
	}
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

// Example demonstrates multiple keys with Redis
func Example_multipleKeys() {
	config := &Config{
		Rate:  10.0, // 10 requests per second
		Burst: 3,    // Allow burst of 3 requests
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		log.Printf("Error creating rate limiter: %v", err)
		return
	}
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
*/

// Example demonstrates validation and default configuration
func ExampleValidateConfig() {
	// Valid configuration with Redis
	config := &Config{
		Rate:  50.0,
		Burst: 100,
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	err := ValidateConfig(config)
	if err != nil {
		fmt.Printf("Config validation failed: %v\n", err)
	} else {
		fmt.Println("Config is valid")
	}

	// Invalid configuration - missing Redis config
	invalidConfig := &Config{
		Rate:  10.0,
		Burst: 20,
		// Redis config is missing - this will fail validation
	}

	err = ValidateConfig(invalidConfig)
	if err != nil {
		fmt.Printf("Invalid config error: %v\n", err)
	}

	// Output:
	// Config is valid
	// Invalid config error: Redis configuration is required
}

// Example demonstrates default configuration
func ExampleDefaultConfig() {
	// Get default configuration
	config := DefaultConfig("localhost:6379")

	fmt.Printf("Default rate: %.1f requests/second\n", config.Rate)
	fmt.Printf("Default burst: %d requests\n", config.Burst)
	fmt.Printf("Redis address: %s\n", config.Redis.Addr)

	// Output:
	// Default rate: 10.0 requests/second
	// Default burst: 20 requests
	// Redis address: localhost:6379
}
