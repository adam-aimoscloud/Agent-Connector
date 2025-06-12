package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue implements PriorityQueue interface using Redis for distributed queue
type RedisQueue struct {
	client *redis.Client
	config *QueueConfig

	// Lua scripts for atomic operations
	enqueueScript        *redis.Script
	dequeueScript        *redis.Script
	updatePriorityScript *redis.Script
	cleanupExpiredScript *redis.Script
}

// Lua script for atomic enqueue operation
const enqueueLuaScript = `
local queue_key = KEYS[1]
local data_key = KEYS[2]
local request_id = ARGV[1]
local priority = tonumber(ARGV[2])
local request_data = ARGV[3]
local max_size = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

-- Check queue size limit
if max_size > 0 then
    local current_size = redis.call('ZCARD', queue_key)
    if current_size >= max_size then
        return {0, "queue_full"}
    end
end

-- Use negative priority for max-heap behavior (Redis ZSET is min-heap by default)
-- Also use current timestamp as tie-breaker to maintain FIFO for same priority
local score = -priority + (redis.call('TIME')[1] + redis.call('TIME')[2] / 1000000) / 1000000000

-- Add to sorted set (priority queue)
redis.call('ZADD', queue_key, score, request_id)

-- Store request data
redis.call('HSET', data_key, request_id, request_data)

-- Set TTL if specified
if ttl > 0 then
    redis.call('EXPIRE', queue_key, ttl)
    redis.call('EXPIRE', data_key, ttl)
end

return {1, "success"}
`

// Lua script for atomic dequeue operation
const dequeueLuaScript = `
local queue_key = KEYS[1]
local data_key = KEYS[2]

-- Get highest priority item (lowest score due to negative priority)
local items = redis.call('ZRANGE', queue_key, 0, 0, 'WITHSCORES')
if #items == 0 then
    return nil
end

local request_id = items[1]

-- Remove from queue
redis.call('ZREM', queue_key, request_id)

-- Get request data
local request_data = redis.call('HGET', data_key, request_id)

-- Remove request data
redis.call('HDEL', data_key, request_id)

return {request_id, request_data}
`

// Lua script for updating priority
const updatePriorityLuaScript = `
local queue_key = KEYS[1]
local request_id = ARGV[1]
local new_priority = tonumber(ARGV[2])

-- Check if request exists in queue
local score = redis.call('ZSCORE', queue_key, request_id)
if not score then
    return 0
end

-- Calculate new score with tie-breaker
local new_score = -new_priority + (redis.call('TIME')[1] + redis.call('TIME')[2] / 1000000) / 1000000000

-- Update priority
redis.call('ZADD', queue_key, new_score, request_id)

return 1
`

// Lua script for cleaning up expired requests
const cleanupExpiredLuaScript = `
local queue_key = KEYS[1]
local data_key = KEYS[2]
local current_time = tonumber(ARGV[1])

-- Get all request IDs
local request_ids = redis.call('ZRANGE', queue_key, 0, -1)
local expired_count = 0

for i = 1, #request_ids do
    local request_id = request_ids[i]
    local request_data = redis.call('HGET', data_key, request_id)
    
    if request_data then
        -- Parse request data to check expiration
        -- This is a simplified check - in practice, you might want to store
        -- expiration time separately for better performance
        local data = cjson.decode(request_data)
        if data.expires_at and data.expires_at < current_time then
            redis.call('ZREM', queue_key, request_id)
            redis.call('HDEL', data_key, request_id)
            expired_count = expired_count + 1
        end
    end
end

return expired_count
`

// NewRedisQueue creates a new Redis-based priority queue
func NewRedisQueue(config *QueueConfig) (*RedisQueue, error) {
	if config.Redis == nil {
		return nil, fmt.Errorf("redis configuration is required for distributed queue")
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:            config.Redis.Addr,
		Password:        config.Redis.Password,
		DB:              config.Redis.DB,
		PoolSize:        config.Redis.PoolSize,
		MinIdleConns:    config.Redis.MinIdleConns,
		ConnMaxIdleTime: config.Redis.ConnMaxIdleTime,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	queue := &RedisQueue{
		client:               client,
		config:               config,
		enqueueScript:        redis.NewScript(enqueueLuaScript),
		dequeueScript:        redis.NewScript(dequeueLuaScript),
		updatePriorityScript: redis.NewScript(updatePriorityLuaScript),
		cleanupExpiredScript: redis.NewScript(cleanupExpiredLuaScript),
	}

	return queue, nil
}

// getQueueKey returns the Redis key for the priority queue
func (q *RedisQueue) getQueueKey(queueName string) string {
	return fmt.Sprintf("%s:queue:%s", q.config.Redis.KeyPrefix, queueName)
}

// getDataKey returns the Redis key for request data storage
func (q *RedisQueue) getDataKey(queueName string) string {
	return fmt.Sprintf("%s:data:%s", q.config.Redis.KeyPrefix, queueName)
}

// Enqueue adds a request to the priority queue
func (q *RedisQueue) Enqueue(ctx context.Context, queueName string, request *Request) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ID == "" {
		return fmt.Errorf("request ID cannot be empty")
	}

	if !request.Priority.IsValid() {
		return fmt.Errorf("invalid priority: %d", request.Priority)
	}

	// Set created time if not set
	if request.CreatedAt.IsZero() {
		request.CreatedAt = time.Now()
	}

	// Serialize request data
	requestData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}

	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Execute enqueue script
	result, err := q.enqueueScript.Run(ctx, q.client, []string{queueKey, dataKey},
		request.ID, int64(request.Priority), string(requestData),
		q.config.MaxQueueSize, q.config.DefaultTTL).Result()

	if err != nil {
		return fmt.Errorf("failed to enqueue request: %w", err)
	}

	// Check result
	if resultSlice, ok := result.([]interface{}); ok && len(resultSlice) == 2 {
		if success, ok := resultSlice[0].(int64); ok && success == 0 {
			if msg, ok := resultSlice[1].(string); ok {
				return fmt.Errorf("enqueue failed: %s", msg)
			}
		}
	}

	return nil
}

// Dequeue removes and returns the highest priority request from the queue
func (q *RedisQueue) Dequeue(ctx context.Context, queueName string) (*Request, error) {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Execute dequeue script
	result, err := q.dequeueScript.Run(ctx, q.client, []string{queueKey, dataKey}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Empty queue
		}
		return nil, fmt.Errorf("failed to dequeue request: %w", err)
	}

	if result == nil {
		return nil, nil // Empty queue
	}

	// Parse result
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		return nil, fmt.Errorf("unexpected dequeue result format")
	}

	requestDataStr, ok := resultSlice[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid request data format")
	}

	// Deserialize request
	var request Request
	if err := json.Unmarshal([]byte(requestDataStr), &request); err != nil {
		return nil, fmt.Errorf("failed to deserialize request: %w", err)
	}

	return &request, nil
}

// DequeueWithTimeout removes and returns the highest priority request with timeout
func (q *RedisQueue) DequeueWithTimeout(ctx context.Context, queueName string, timeout time.Duration) (*Request, error) {
	queueKey := q.getQueueKey(queueName)

	// Use BZPOPMIN for blocking pop with timeout
	result, err := q.client.BZPopMin(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Timeout or empty queue
		}
		return nil, fmt.Errorf("failed to dequeue with timeout: %w", err)
	}

	if result.Member == nil {
		return nil, nil // Empty result
	}

	requestID := result.Member.(string)
	dataKey := q.getDataKey(queueName)

	// Get request data
	requestDataStr, err := q.client.HGet(ctx, dataKey, requestID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("request data not found for ID: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get request data: %w", err)
	}

	// Remove request data
	q.client.HDel(ctx, dataKey, requestID)

	// Deserialize request
	var request Request
	if err := json.Unmarshal([]byte(requestDataStr), &request); err != nil {
		return nil, fmt.Errorf("failed to deserialize request: %w", err)
	}

	return &request, nil
}

// Peek returns the highest priority request without removing it
func (q *RedisQueue) Peek(ctx context.Context, queueName string) (*Request, error) {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Get highest priority item without removing
	items, err := q.client.ZRangeWithScores(ctx, queueKey, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to peek queue: %w", err)
	}

	if len(items) == 0 {
		return nil, nil // Empty queue
	}

	requestID := items[0].Member.(string)

	// Get request data
	requestDataStr, err := q.client.HGet(ctx, dataKey, requestID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("request data not found for ID: %s", requestID)
		}
		return nil, fmt.Errorf("failed to get request data: %w", err)
	}

	// Deserialize request
	var request Request
	if err := json.Unmarshal([]byte(requestDataStr), &request); err != nil {
		return nil, fmt.Errorf("failed to deserialize request: %w", err)
	}

	return &request, nil
}

// Size returns the number of requests in the queue
func (q *RedisQueue) Size(ctx context.Context, queueName string) (int64, error) {
	queueKey := q.getQueueKey(queueName)

	size, err := q.client.ZCard(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return size, nil
}

// Remove removes a specific request from the queue by ID
func (q *RedisQueue) Remove(ctx context.Context, queueName string, requestID string) error {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Remove from queue and data storage
	pipe := q.client.Pipeline()
	pipe.ZRem(ctx, queueKey, requestID)
	pipe.HDel(ctx, dataKey, requestID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove request: %w", err)
	}

	return nil
}

// UpdatePriority updates the priority of a request in the queue
func (q *RedisQueue) UpdatePriority(ctx context.Context, queueName string, requestID string, newPriority Priority) error {
	if !newPriority.IsValid() {
		return fmt.Errorf("invalid priority: %d", newPriority)
	}

	queueKey := q.getQueueKey(queueName)

	// Execute update priority script
	result, err := q.updatePriorityScript.Run(ctx, q.client, []string{queueKey},
		requestID, int64(newPriority)).Result()

	if err != nil {
		return fmt.Errorf("failed to update priority: %w", err)
	}

	if updated, ok := result.(int64); ok && updated == 0 {
		return fmt.Errorf("request not found: %s", requestID)
	}

	return nil
}

// ListByPriority returns requests in priority order with pagination
func (q *RedisQueue) ListByPriority(ctx context.Context, queueName string, offset, limit int64) ([]*Request, error) {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Get request IDs in priority order
	requestIDs, err := q.client.ZRange(ctx, queueKey, offset, offset+limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list requests: %w", err)
	}

	if len(requestIDs) == 0 {
		return []*Request{}, nil
	}

	// Get request data for all IDs
	requestDataList, err := q.client.HMGet(ctx, dataKey, requestIDs...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get request data: %w", err)
	}

	requests := make([]*Request, 0, len(requestIDs))
	for _, requestData := range requestDataList {
		if requestData == nil {
			continue // Skip missing data
		}

		requestDataStr, ok := requestData.(string)
		if !ok {
			continue
		}

		var request Request
		if err := json.Unmarshal([]byte(requestDataStr), &request); err != nil {
			continue // Skip invalid data
		}

		requests = append(requests, &request)
	}

	return requests, nil
}

// Clear removes all requests from the queue
func (q *RedisQueue) Clear(ctx context.Context, queueName string) error {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)

	// Remove both queue and data keys
	pipe := q.client.Pipeline()
	pipe.Del(ctx, queueKey)
	pipe.Del(ctx, dataKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	return nil
}

// Close cleans up resources used by the queue
func (q *RedisQueue) Close() error {
	return q.client.Close()
}

// CleanupExpired removes expired requests from the queue
func (q *RedisQueue) CleanupExpired(ctx context.Context, queueName string) (int64, error) {
	queueKey := q.getQueueKey(queueName)
	dataKey := q.getDataKey(queueName)
	currentTime := time.Now().Unix()

	result, err := q.cleanupExpiredScript.Run(ctx, q.client, []string{queueKey, dataKey}, currentTime).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired requests: %w", err)
	}

	expiredCount, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected cleanup result format")
	}

	return expiredCount, nil
}
