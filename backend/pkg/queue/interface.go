package queue

import (
	"context"
	"time"
)

// PriorityQueue defines the interface for priority queue operations
type PriorityQueue interface {
	// Enqueue adds a request to the priority queue
	Enqueue(ctx context.Context, queueName string, request *Request) error

	// Dequeue removes and returns the highest priority request from the queue
	Dequeue(ctx context.Context, queueName string) (*Request, error)

	// DequeueWithTimeout removes and returns the highest priority request with timeout
	DequeueWithTimeout(ctx context.Context, queueName string, timeout time.Duration) (*Request, error)

	// Peek returns the highest priority request without removing it
	Peek(ctx context.Context, queueName string) (*Request, error)

	// Size returns the number of requests in the queue
	Size(ctx context.Context, queueName string) (int64, error)

	// Remove removes a specific request from the queue by ID
	Remove(ctx context.Context, queueName string, requestID string) error

	// UpdatePriority updates the priority of a request in the queue
	UpdatePriority(ctx context.Context, queueName string, requestID string, newPriority Priority) error

	// ListByPriority returns requests in priority order with pagination
	ListByPriority(ctx context.Context, queueName string, offset, limit int64) ([]*Request, error)

	// Clear removes all requests from the queue
	Clear(ctx context.Context, queueName string) error

	// Close cleans up resources used by the queue
	Close() error
}

// Request represents a request in the priority queue
type Request struct {
	// ID is the unique identifier for the request
	ID string `json:"id"`

	// UserID is the ID of the user making the request
	UserID string `json:"user_id"`

	// AgentID is the ID of the target agent
	AgentID string `json:"agent_id"`

	// Priority is the priority level of the request
	Priority Priority `json:"priority"`

	// Payload contains the actual request data
	Payload interface{} `json:"payload"`

	// Metadata contains additional request metadata
	Metadata map[string]interface{} `json:"metadata"`

	// CreatedAt is the timestamp when the request was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is the timestamp when the request expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Priority represents the priority level of a request
type Priority int64

const (
	// PriorityLowest represents the lowest priority level
	PriorityLowest Priority = 0

	// PriorityLow represents low priority level
	PriorityLow Priority = 25

	// PriorityNormal represents normal priority level
	PriorityNormal Priority = 50

	// PriorityHigh represents high priority level
	PriorityHigh Priority = 75

	// PriorityHighest represents the highest priority level
	PriorityHighest Priority = 100

	// PriorityCritical represents critical priority level (emergency)
	PriorityCritical Priority = 1000
)

// String returns the string representation of the priority
func (p Priority) String() string {
	switch {
	case p >= PriorityCritical:
		return "Critical"
	case p >= PriorityHighest:
		return "Highest"
	case p >= PriorityHigh:
		return "High"
	case p >= PriorityNormal:
		return "Normal"
	case p >= PriorityLow:
		return "Low"
	default:
		return "Lowest"
	}
}

// IsValid checks if the priority is within valid range
func (p Priority) IsValid() bool {
	return p >= PriorityLowest && p <= PriorityCritical
}

// QueueConfig represents the configuration for priority queue
type QueueConfig struct {
	// Redis configuration for distributed queue
	Redis *RedisConfig

	// DefaultTTL is the default TTL for requests in seconds
	DefaultTTL int64

	// MaxQueueSize is the maximum number of requests per queue (0 = unlimited)
	MaxQueueSize int64

	// EnableMetrics enables metrics collection
	EnableMetrics bool
}

// RedisConfig represents Redis configuration for distributed queue
type RedisConfig struct {
	// Addr is the Redis server address
	Addr string

	// Password is the Redis password
	Password string

	// DB is the Redis database number
	DB int

	// PoolSize is the maximum number of connections in the pool
	PoolSize int

	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int

	// ConnMaxIdleTime is the maximum idle time for connections
	ConnMaxIdleTime time.Duration

	// KeyPrefix is the prefix for all Redis keys
	KeyPrefix string
}

// QueueStats represents queue statistics
type QueueStats struct {
	// TotalRequests is the total number of requests in the queue
	TotalRequests int64 `json:"total_requests"`

	// RequestsByPriority is the breakdown by priority level
	RequestsByPriority map[Priority]int64 `json:"requests_by_priority"`

	// OldestRequest is the timestamp of the oldest request
	OldestRequest *time.Time `json:"oldest_request,omitempty"`

	// AverageWaitTime is the average wait time for requests
	AverageWaitTime time.Duration `json:"average_wait_time"`
}

// QueueMetrics represents queue performance metrics
type QueueMetrics struct {
	// EnqueueCount is the total number of enqueue operations
	EnqueueCount int64 `json:"enqueue_count"`

	// DequeueCount is the total number of dequeue operations
	DequeueCount int64 `json:"dequeue_count"`

	// ExpiredCount is the total number of expired requests
	ExpiredCount int64 `json:"expired_count"`

	// AverageProcessingTime is the average time to process requests
	AverageProcessingTime time.Duration `json:"average_processing_time"`
}
