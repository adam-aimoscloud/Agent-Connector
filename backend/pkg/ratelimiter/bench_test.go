package ratelimiter

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkLocalRateLimiter_Allow(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "bench-key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow(ctx, key)
		}
	})
}

func BenchmarkLocalRateLimiter_AllowN(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "bench-key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.AllowN(ctx, key, 1)
		}
	})
}

func BenchmarkLocalRateLimiter_MultipleKeys(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i%100) // Cycle through 100 keys
			limiter.Allow(ctx, key)
			i++
		}
	})
}

func BenchmarkLocalRateLimiter_Reserve(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "bench-key"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reservation, _ := limiter.Reserve(ctx, key)
			if reservation != nil {
				reservation.Cancel() // Clean up reservation
			}
		}
	})
}

func BenchmarkLocalRateLimiter_DifferentBurstSizes(b *testing.B) {
	burstSizes := []int{1, 10, 100, 1000}

	for _, burst := range burstSizes {
		b.Run(fmt.Sprintf("Burst%d", burst), func(b *testing.B) {
			config := &Config{
				Rate:  1000.0,
				Burst: burst,
			}

			limiter := NewLocalRateLimiter(config)
			defer limiter.Close()

			ctx := context.Background()
			key := "bench-key"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				limiter.Allow(ctx, key)
			}
		})
	}
}

func BenchmarkLocalRateLimiter_HighContention(b *testing.B) {
	config := &Config{
		Rate:  10000.0, // High rate to avoid blocking
		Burst: 10000,   // High burst to avoid blocking
	}

	limiter := NewLocalRateLimiter(config)
	defer limiter.Close()

	ctx := context.Background()
	key := "single-key" // All goroutines use the same key

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow(ctx, key)
		}
	})
}
