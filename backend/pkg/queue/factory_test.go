package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPriorityQueue(t *testing.T) {
	tests := []struct {
		name        string
		queueType   QueueType
		config      *QueueConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			queueType:   RedisType,
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:      "invalid config - no redis",
			queueType: RedisType,
			config: &QueueConfig{
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "redis configuration is required",
		},
		{
			name:      "invalid config - empty redis addr",
			queueType: RedisType,
			config: &QueueConfig{
				Redis:        &RedisConfig{},
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "redis address cannot be empty",
		},
		{
			name:      "unsupported queue type",
			queueType: "invalid",
			config: &QueueConfig{
				Redis:        DefaultRedisQueueConfig("localhost:6379"),
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "unsupported queue type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add Redis config for valid cases
			if tt.config != nil && tt.config.Redis == nil && !tt.expectError {
				tt.config.Redis = DefaultRedisQueueConfig("localhost:6379")
			}

			queue, err := NewPriorityQueue(tt.queueType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, queue)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, queue)
				if queue != nil {
					queue.Close()
				}
			}
		})
	}
}

func TestValidateQueueConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *QueueConfig
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
			name: "no redis config",
			config: &QueueConfig{
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "redis configuration is required",
		},
		{
			name: "empty redis addr",
			config: &QueueConfig{
				Redis:        &RedisConfig{},
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "redis address cannot be empty",
		},
		{
			name: "negative MinIdleConns",
			config: &QueueConfig{
				Redis: &RedisConfig{
					Addr:         "localhost:6379",
					MinIdleConns: -1,
				},
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "MinIdleConns cannot be negative",
		},
		{
			name: "negative DefaultTTL",
			config: &QueueConfig{
				Redis: &RedisConfig{
					Addr: "localhost:6379",
				},
				DefaultTTL:   -1,
				MaxQueueSize: 100,
			},
			expectError: true,
			errorMsg:    "DefaultTTL cannot be negative",
		},
		{
			name: "negative MaxQueueSize",
			config: &QueueConfig{
				Redis: &RedisConfig{
					Addr: "localhost:6379",
				},
				DefaultTTL:   3600,
				MaxQueueSize: -1,
			},
			expectError: true,
			errorMsg:    "MaxQueueSize cannot be negative",
		},
		{
			name: "valid config",
			config: &QueueConfig{
				Redis: &RedisConfig{
					Addr:         "localhost:6379",
					PoolSize:     10,
					MinIdleConns: 2,
				},
				DefaultTTL:   3600,
				MaxQueueSize: 100,
			},
			expectError: false,
		},
		{
			name: "auto-set defaults",
			config: &QueueConfig{
				Redis: &RedisConfig{
					Addr:     "localhost:6379",
					PoolSize: 0, // Should be auto-set
				},
				DefaultTTL:   0,
				MaxQueueSize: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueueConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Check that defaults were set
				if tt.config != nil && tt.config.Redis != nil {
					if tt.name == "auto-set defaults" {
						assert.Equal(t, 10, tt.config.Redis.PoolSize)
						assert.Equal(t, "agent-connector", tt.config.Redis.KeyPrefix)
					}
				}
			}
		})
	}
}

func TestDefaultQueueConfig(t *testing.T) {
	config := DefaultQueueConfig()

	require.NotNil(t, config)
	assert.Equal(t, int64(3600), config.DefaultTTL)
	assert.Equal(t, int64(0), config.MaxQueueSize)
	assert.True(t, config.EnableMetrics)
	assert.Nil(t, config.Redis)
}

func TestDefaultRedisQueueConfig(t *testing.T) {
	addr := "localhost:6379"
	config := DefaultRedisQueueConfig(addr)

	require.NotNil(t, config)
	assert.Equal(t, addr, config.Addr)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 2, config.MinIdleConns)
	assert.Equal(t, 30*time.Minute, config.ConnMaxIdleTime)
	assert.Equal(t, "agent-connector", config.KeyPrefix)
}

func TestRequestBuilder(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		expTime := time.Now().Add(time.Hour)

		request, err := NewRequestBuilder().
			WithID("req-123").
			WithUserID("user-456").
			WithAgentID("agent-789").
			WithPriority(PriorityHigh).
			WithPayload(map[string]string{"message": "hello"}).
			WithMetadata("source", "test").
			WithExpiration(expTime).
			Build()

		require.NoError(t, err)
		require.NotNil(t, request)

		assert.Equal(t, "req-123", request.ID)
		assert.Equal(t, "user-456", request.UserID)
		assert.Equal(t, "agent-789", request.AgentID)
		assert.Equal(t, PriorityHigh, request.Priority)
		assert.NotNil(t, request.Payload)
		assert.Equal(t, "test", request.Metadata["source"])
		assert.Equal(t, expTime.Unix(), request.ExpiresAt.Unix())
	})

	t.Run("with TTL", func(t *testing.T) {
		beforeBuild := time.Now()

		request, err := NewRequestBuilder().
			WithID("req-123").
			WithUserID("user-456").
			WithAgentID("agent-789").
			WithPriority(PriorityNormal).
			WithTTL(time.Hour).
			Build()

		require.NoError(t, err)
		require.NotNil(t, request)
		require.NotNil(t, request.ExpiresAt)

		// Check that expiration is approximately 1 hour from now
		expectedExpiry := beforeBuild.Add(time.Hour)
		assert.WithinDuration(t, expectedExpiry, *request.ExpiresAt, time.Second)
	})

	t.Run("missing ID", func(t *testing.T) {
		_, err := NewRequestBuilder().
			WithUserID("user-456").
			WithAgentID("agent-789").
			WithPriority(PriorityNormal).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request ID is required")
	})

	t.Run("missing UserID", func(t *testing.T) {
		_, err := NewRequestBuilder().
			WithID("req-123").
			WithAgentID("agent-789").
			WithPriority(PriorityNormal).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("missing AgentID", func(t *testing.T) {
		_, err := NewRequestBuilder().
			WithID("req-123").
			WithUserID("user-456").
			WithPriority(PriorityNormal).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent ID is required")
	})

	t.Run("invalid priority", func(t *testing.T) {
		_, err := NewRequestBuilder().
			WithID("req-123").
			WithUserID("user-456").
			WithAgentID("agent-789").
			WithPriority(Priority(-1)).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid priority")
	})
}

func TestPriorityFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
		hasError bool
	}{
		{"lowest", PriorityLowest, false},
		{"low", PriorityLow, false},
		{"normal", PriorityNormal, false},
		{"high", PriorityHigh, false},
		{"highest", PriorityHighest, false},
		{"critical", PriorityCritical, false},
		{"invalid", PriorityLowest, true},
		{"", PriorityLowest, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			priority, err := PriorityFromString(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, priority)
			}
		})
	}
}

func TestPriorityMethods(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		tests := []struct {
			priority Priority
			expected string
		}{
			{PriorityLowest, "Lowest"},
			{PriorityLow, "Low"},
			{PriorityNormal, "Normal"},
			{PriorityHigh, "High"},
			{PriorityHighest, "Highest"},
			{PriorityCritical, "Critical"},
			{Priority(2000), "Critical"}, // Above critical threshold
			{Priority(-1), "Lowest"},     // Below lowest
		}

		for _, tt := range tests {
			assert.Equal(t, tt.expected, tt.priority.String())
		}
	})

	t.Run("IsValid method", func(t *testing.T) {
		tests := []struct {
			priority Priority
			valid    bool
		}{
			{PriorityLowest, true},
			{PriorityLow, true},
			{PriorityNormal, true},
			{PriorityHigh, true},
			{PriorityHighest, true},
			{PriorityCritical, true},
			{Priority(-1), false},
			{Priority(2000), false},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.valid, tt.priority.IsValid())
		}
	})
}

func TestQueueNameBuilder(t *testing.T) {
	t.Run("empty builder", func(t *testing.T) {
		name := NewQueueNameBuilder().Build()
		assert.Equal(t, "default", name)
	})

	t.Run("with agent", func(t *testing.T) {
		name := NewQueueNameBuilder().
			WithAgent("agent-123").
			Build()
		assert.Equal(t, "agent:agent-123", name)
	})

	t.Run("complex queue name", func(t *testing.T) {
		name := NewQueueNameBuilder().
			WithService("chat").
			WithRegion("us-west-2").
			WithAgent("agent-123").
			WithCustom("priority", "high").
			Build()
		assert.Equal(t, "service:chat:region:us-west-2:agent:agent-123:priority:high", name)
	})

	t.Run("with custom parts", func(t *testing.T) {
		name := NewQueueNameBuilder().
			WithCustom("custom", "part1", "part2").
			Build()
		assert.Equal(t, "custom:part1:part2", name)
	})
}
