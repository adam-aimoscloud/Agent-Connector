package ratelimiter

// Note: Redis benchmarks require a running Redis instance
// These benchmarks are commented out by default to avoid test failures
// Uncomment and run with a Redis instance for performance testing

/*
func BenchmarkRedisRateLimiter_Allow(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		b.Fatalf("Failed to create Redis rate limiter: %v", err)
	}
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

func BenchmarkRedisRateLimiter_AllowN(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		b.Fatalf("Failed to create Redis rate limiter: %v", err)
	}
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

func BenchmarkRedisRateLimiter_MultipleKeys(b *testing.B) {
	config := &Config{
		Rate:  1000.0, // 1000 requests per second
		Burst: 1000,   // Large burst for benchmarking
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}

	limiter, err := NewRateLimiter(RedisType, config)
	if err != nil {
		b.Fatalf("Failed to create Redis rate limiter: %v", err)
	}
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
*/
