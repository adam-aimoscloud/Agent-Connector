package queue

import (
	"fmt"
	"time"
)

// QueueType represents the type of priority queue
type QueueType string

const (
	// RedisType uses Redis for distributed priority queue
	RedisType QueueType = "redis"
)

// NewPriorityQueue creates a new priority queue based on the configuration
func NewPriorityQueue(queueType QueueType, config *QueueConfig) (PriorityQueue, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := ValidateQueueConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch queueType {
	case RedisType:
		return NewRedisQueue(config)

	default:
		return nil, fmt.Errorf("unsupported queue type: %s", queueType)
	}
}

// ValidateQueueConfig validates the queue configuration
func ValidateQueueConfig(config *QueueConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate Redis configuration
	if config.Redis == nil {
		return fmt.Errorf("redis configuration is required")
	}

	if config.Redis.Addr == "" {
		return fmt.Errorf("redis address cannot be empty")
	}

	// Set default values
	if config.Redis.PoolSize <= 0 {
		config.Redis.PoolSize = 10
	}

	if config.Redis.MinIdleConns < 0 {
		return fmt.Errorf("MinIdleConns cannot be negative, got: %d", config.Redis.MinIdleConns)
	}

	if config.Redis.KeyPrefix == "" {
		config.Redis.KeyPrefix = "agent-connector"
	}

	if config.DefaultTTL < 0 {
		return fmt.Errorf("DefaultTTL cannot be negative, got: %d", config.DefaultTTL)
	}

	if config.MaxQueueSize < 0 {
		return fmt.Errorf("MaxQueueSize cannot be negative, got: %d", config.MaxQueueSize)
	}

	return nil
}

// DefaultQueueConfig returns a default queue configuration
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		DefaultTTL:    3600, // 1 hour
		MaxQueueSize:  0,    // unlimited
		EnableMetrics: true,
	}
}

// DefaultRedisQueueConfig returns a default Redis configuration for queues
func DefaultRedisQueueConfig(addr string) *RedisConfig {
	return &RedisConfig{
		Addr:            addr,
		Password:        "",
		DB:              0,
		PoolSize:        10,
		MinIdleConns:    2,
		ConnMaxIdleTime: 30 * time.Minute,
		KeyPrefix:       "agent-connector",
	}
}

// RequestBuilder provides a fluent interface for building requests
type RequestBuilder struct {
	request *Request
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		request: &Request{
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		},
	}
}

// WithID sets the request ID
func (rb *RequestBuilder) WithID(id string) *RequestBuilder {
	rb.request.ID = id
	return rb
}

// WithUserID sets the user ID
func (rb *RequestBuilder) WithUserID(userID string) *RequestBuilder {
	rb.request.UserID = userID
	return rb
}

// WithAgentID sets the agent ID
func (rb *RequestBuilder) WithAgentID(agentID string) *RequestBuilder {
	rb.request.AgentID = agentID
	return rb
}

// WithPriority sets the priority
func (rb *RequestBuilder) WithPriority(priority Priority) *RequestBuilder {
	rb.request.Priority = priority
	return rb
}

// WithPayload sets the payload
func (rb *RequestBuilder) WithPayload(payload interface{}) *RequestBuilder {
	rb.request.Payload = payload
	return rb
}

// WithMetadata adds metadata
func (rb *RequestBuilder) WithMetadata(key string, value interface{}) *RequestBuilder {
	rb.request.Metadata[key] = value
	return rb
}

// WithExpiration sets the expiration time
func (rb *RequestBuilder) WithExpiration(expiresAt time.Time) *RequestBuilder {
	rb.request.ExpiresAt = &expiresAt
	return rb
}

// WithTTL sets the TTL (time to live) from now
func (rb *RequestBuilder) WithTTL(ttl time.Duration) *RequestBuilder {
	expiresAt := time.Now().Add(ttl)
	rb.request.ExpiresAt = &expiresAt
	return rb
}

// Build validates and returns the request
func (rb *RequestBuilder) Build() (*Request, error) {
	if rb.request.ID == "" {
		return nil, fmt.Errorf("request ID is required")
	}

	if rb.request.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if rb.request.AgentID == "" {
		return nil, fmt.Errorf("agent ID is required")
	}

	if !rb.request.Priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %d", rb.request.Priority)
	}

	// Create a copy to avoid sharing the same metadata map
	request := &Request{
		ID:        rb.request.ID,
		UserID:    rb.request.UserID,
		AgentID:   rb.request.AgentID,
		Priority:  rb.request.Priority,
		Payload:   rb.request.Payload,
		Metadata:  make(map[string]interface{}),
		CreatedAt: rb.request.CreatedAt,
		ExpiresAt: rb.request.ExpiresAt,
	}

	// Copy metadata
	for k, v := range rb.request.Metadata {
		request.Metadata[k] = v
	}

	return request, nil
}

// PriorityFromString converts a string to Priority
func PriorityFromString(s string) (Priority, error) {
	switch s {
	case "lowest":
		return PriorityLowest, nil
	case "low":
		return PriorityLow, nil
	case "normal":
		return PriorityNormal, nil
	case "high":
		return PriorityHigh, nil
	case "highest":
		return PriorityHighest, nil
	case "critical":
		return PriorityCritical, nil
	default:
		return PriorityLowest, fmt.Errorf("unknown priority: %s", s)
	}
}

// QueueNameBuilder provides utilities for building queue names
type QueueNameBuilder struct {
	parts []string
}

// NewQueueNameBuilder creates a new queue name builder
func NewQueueNameBuilder() *QueueNameBuilder {
	return &QueueNameBuilder{
		parts: make([]string, 0),
	}
}

// WithAgent adds agent ID to the queue name
func (qb *QueueNameBuilder) WithAgent(agentID string) *QueueNameBuilder {
	qb.parts = append(qb.parts, "agent", agentID)
	return qb
}

// WithService adds service name to the queue name
func (qb *QueueNameBuilder) WithService(serviceName string) *QueueNameBuilder {
	qb.parts = append(qb.parts, "service", serviceName)
	return qb
}

// WithRegion adds region to the queue name
func (qb *QueueNameBuilder) WithRegion(region string) *QueueNameBuilder {
	qb.parts = append(qb.parts, "region", region)
	return qb
}

// WithCustom adds custom parts to the queue name
func (qb *QueueNameBuilder) WithCustom(parts ...string) *QueueNameBuilder {
	qb.parts = append(qb.parts, parts...)
	return qb
}

// Build returns the queue name
func (qb *QueueNameBuilder) Build() string {
	if len(qb.parts) == 0 {
		return "default"
	}

	result := qb.parts[0]
	for i := 1; i < len(qb.parts); i++ {
		result += ":" + qb.parts[i]
	}

	return result
}
