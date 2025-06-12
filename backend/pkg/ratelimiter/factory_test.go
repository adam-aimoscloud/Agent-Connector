package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name        string
		limiterType RateLimiterType
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			limiterType: LocalType,
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "invalid rate",
			limiterType: LocalType,
			config:      &Config{Rate: -1, Burst: 10},
			expectError: true,
			errorMsg:    "rate must be positive",
		},
		{
			name:        "invalid burst",
			limiterType: LocalType,
			config:      &Config{Rate: 10, Burst: -1},
			expectError: true,
			errorMsg:    "burst must be positive",
		},
		{
			name:        "valid local config",
			limiterType: LocalType,
			config:      &Config{Rate: 10, Burst: 20},
			expectError: false,
		},
		{
			name:        "unsupported type",
			limiterType: "invalid",
			config:      &Config{Rate: 10, Burst: 20},
			expectError: true,
			errorMsg:    "unsupported rate limiter type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := NewRateLimiter(tt.limiterType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, limiter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, limiter)

				// Clean up
				if limiter != nil {
					limiter.Close()
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "negative rate",
			config:      &Config{Rate: -1, Burst: 10},
			expectError: true,
			errorMsg:    "rate must be positive",
		},
		{
			name:        "zero rate",
			config:      &Config{Rate: 0, Burst: 10},
			expectError: true,
			errorMsg:    "rate must be positive",
		},
		{
			name:        "negative burst",
			config:      &Config{Rate: 10, Burst: -1},
			expectError: true,
			errorMsg:    "burst must be positive",
		},
		{
			name:        "zero burst",
			config:      &Config{Rate: 10, Burst: 0},
			expectError: true,
			errorMsg:    "burst must be positive",
		},
		{
			name:        "valid config without redis",
			config:      &Config{Rate: 10, Burst: 20},
			expectError: false,
		},
		{
			name: "redis config with empty addr",
			config: &Config{
				Rate:  10,
				Burst: 20,
				Redis: &RedisConfig{Addr: ""},
			},
			expectError: true,
			errorMsg:    "Redis address cannot be empty",
		},
		{
			name: "redis config with negative MinIdleConns",
			config: &Config{
				Rate:  10,
				Burst: 20,
				Redis: &RedisConfig{
					Addr:         "localhost:6379",
					MinIdleConns: -1,
				},
			},
			expectError: true,
			errorMsg:    "MinIdleConns cannot be negative",
		},
		{
			name: "valid redis config",
			config: &Config{
				Rate:  10,
				Burst: 20,
				Redis: &RedisConfig{
					Addr:         "localhost:6379",
					PoolSize:     10,
					MinIdleConns: 2,
				},
			},
			expectError: false,
		},
		{
			name: "redis config auto-sets default pool size",
			config: &Config{
				Rate:  10,
				Burst: 20,
				Redis: &RedisConfig{
					Addr:     "localhost:6379",
					PoolSize: 0, // Should be auto-set to 10
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Check that default pool size was set
				if tt.config != nil && tt.config.Redis != nil && tt.name == "redis config auto-sets default pool size" {
					assert.Equal(t, 10, tt.config.Redis.PoolSize)
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, 10.0, config.Rate)
	assert.Equal(t, 20, config.Burst)
	assert.Nil(t, config.Redis)
}

func TestDefaultRedisConfig(t *testing.T) {
	addr := "localhost:6379"
	config := DefaultRedisConfig(addr)

	require.NotNil(t, config)
	assert.Equal(t, addr, config.Addr)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 2, config.MinIdleConns)
	assert.Equal(t, time.Duration(30*60*1000*1000*1000), config.ConnMaxIdleTime)
}
