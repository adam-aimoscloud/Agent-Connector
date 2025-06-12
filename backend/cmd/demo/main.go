package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"agent-connector/pkg/ratelimiter"
)

func main() {
	fmt.Println("=== Rate Limiter Demo ===")

	// Demo 1: Basic usage
	fmt.Println("\n1. Basic Rate Limiting Demo")
	basicDemo()

	// Demo 2: Multiple keys
	fmt.Println("\n2. Multiple Keys Demo")
	multipleKeysDemo()

	// Demo 3: Burst behavior
	fmt.Println("\n3. Burst Behavior Demo")
	burstDemo()

	// Demo 4: Wait behavior
	fmt.Println("\n4. Wait Behavior Demo")
	waitDemo()

	// Demo 5: Reservation pattern
	fmt.Println("\n5. Reservation Pattern Demo")
	reservationDemo()
}

func basicDemo() {
	config := &ratelimiter.Config{
		Rate:  5.0, // 5 requests per second
		Burst: 3,   // Allow burst of 3 requests
	}

	limiter := ratelimiter.NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "user:123"

	fmt.Printf("Rate: %.1f req/sec, Burst: %d\n", config.Rate, config.Burst)

	// Make 6 requests quickly
	for i := 1; i <= 6; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		status := "ALLOWED"
		if !allowed {
			status = "RATE LIMITED"
		}

		fmt.Printf("Request %d: %s\n", i, status)
	}
}

func multipleKeysDemo() {
	config := &ratelimiter.Config{
		Rate:  10.0, // 10 requests per second
		Burst: 2,    // Allow burst of 2 requests
	}

	limiter := ratelimiter.NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	users := []string{"alice", "bob", "charlie"}

	fmt.Printf("Rate: %.1f req/sec, Burst: %d per user\n", config.Rate, config.Burst)

	// Each user makes 3 requests
	for _, user := range users {
		key := fmt.Sprintf("user:%s", user)
		fmt.Printf("\nUser %s:\n", user)

		for i := 1; i <= 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}

			status := "ALLOWED"
			if !allowed {
				status = "RATE LIMITED"
			}

			fmt.Printf("  Request %d: %s\n", i, status)
		}
	}
}

func burstDemo() {
	config := &ratelimiter.Config{
		Rate:  2.0, // 2 requests per second (slow)
		Burst: 5,   // Allow burst of 5 requests
	}

	limiter := ratelimiter.NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "burst-test"

	fmt.Printf("Rate: %.1f req/sec, Burst: %d\n", config.Rate, config.Burst)
	fmt.Println("Making 8 requests immediately:")

	// Make 8 requests immediately
	for i := 1; i <= 8; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		status := "ALLOWED"
		if !allowed {
			status = "RATE LIMITED"
		}

		fmt.Printf("Request %d: %s\n", i, status)
	}

	// Wait for tokens to refill
	fmt.Println("\nWaiting 3 seconds for tokens to refill...")
	time.Sleep(3 * time.Second)

	fmt.Println("Making 3 more requests:")
	for i := 9; i <= 11; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		status := "ALLOWED"
		if !allowed {
			status = "RATE LIMITED"
		}

		fmt.Printf("Request %d: %s\n", i, status)
	}
}

func waitDemo() {
	config := &ratelimiter.Config{
		Rate:  3.0, // 3 requests per second
		Burst: 1,   // Allow burst of 1 request
	}

	limiter := ratelimiter.NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "wait-test"

	fmt.Printf("Rate: %.1f req/sec, Burst: %d\n", config.Rate, config.Burst)
	fmt.Println("Making 3 requests with Wait():")

	start := time.Now()

	for i := 1; i <= 3; i++ {
		requestStart := time.Now()

		err := limiter.Wait(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		elapsed := time.Since(requestStart)
		totalElapsed := time.Since(start)

		fmt.Printf("Request %d: completed after %v (total: %v)\n",
			i, elapsed.Round(time.Millisecond), totalElapsed.Round(time.Millisecond))
	}
}

func reservationDemo() {
	config := &ratelimiter.Config{
		Rate:  4.0, // 4 requests per second
		Burst: 2,   // Allow burst of 2 requests
	}

	limiter := ratelimiter.NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "reservation-test"

	fmt.Printf("Rate: %.1f req/sec, Burst: %d\n", config.Rate, config.Burst)
	fmt.Println("Making reservations:")

	for i := 1; i <= 4; i++ {
		reservation, err := limiter.Reserve(ctx, key)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if reservation.OK {
			if reservation.Delay > 0 {
				fmt.Printf("Reservation %d: OK, delay %v\n", i, reservation.Delay.Round(time.Millisecond))
				time.Sleep(reservation.Delay)
			} else {
				fmt.Printf("Reservation %d: OK, no delay\n", i)
			}

			fmt.Printf("  Processing request %d...\n", i)
		} else {
			fmt.Printf("Reservation %d: FAILED\n", i)
		}
	}
}
