package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"agent-connector/pkg/queue"
)

func main() {
	fmt.Println("=== Priority Queue Demo ===")

	// Demo 1: Basic queue operations
	fmt.Println("\n1. Basic Queue Operations Demo")
	basicQueueDemo()

	// Demo 2: Priority ordering
	fmt.Println("\n2. Priority Ordering Demo")
	priorityOrderingDemo()

	// Demo 3: Request builder
	fmt.Println("\n3. Request Builder Demo")
	requestBuilderDemo()

	// Demo 4: Queue name builder
	fmt.Println("\n4. Queue Name Builder Demo")
	queueNameBuilderDemo()
}

func basicQueueDemo() {
	config := queue.DefaultQueueConfig()
	config.Redis = queue.DefaultRedisQueueConfig("localhost:6379")
	config.MaxQueueSize = 100

	fmt.Printf("Queue Config: MaxSize=%d, TTL=%d seconds\n", config.MaxQueueSize, config.DefaultTTL)

	// Create sample requests
	requests := []*queue.Request{
		{
			ID:       "req-1",
			UserID:   "user-alice",
			AgentID:  "agent-123",
			Priority: queue.PriorityLow,
			Payload:  "Low priority request",
		},
		{
			ID:       "req-2",
			UserID:   "user-bob",
			AgentID:  "agent-123",
			Priority: queue.PriorityHigh,
			Payload:  "High priority request",
		},
		{
			ID:       "req-3",
			UserID:   "user-charlie",
			AgentID:  "agent-123",
			Priority: queue.PriorityNormal,
			Payload:  "Normal priority request",
		},
	}

	ctx := context.Background()
	queueName := "agent:agent-123"

	fmt.Printf("\nEnqueuing %d requests to queue '%s':\n", len(requests), queueName)
	for _, req := range requests {
		fmt.Printf("  - %s: Priority=%s, User=%s\n",
			req.ID, req.Priority.String(), req.UserID)
	}

	fmt.Println("\nSimulated dequeue operations (highest priority first):")
	expectedOrder := []string{"req-2", "req-3", "req-1"} // High, Normal, Low
	for _, reqID := range expectedOrder {
		fmt.Printf("  - Dequeued: %s\n", reqID)
	}

	_ = ctx // Suppress unused variable warning
}

func priorityOrderingDemo() {
	priorities := []queue.Priority{
		queue.PriorityCritical,
		queue.PriorityLowest,
		queue.PriorityHigh,
		queue.PriorityNormal,
		queue.PriorityLow,
		queue.PriorityHighest,
	}

	fmt.Println("Priority levels and their values:")
	for _, p := range priorities {
		fmt.Printf("  - %s: %d (valid: %v)\n", p.String(), int64(p), p.IsValid())
	}

	fmt.Println("\nPriority from string conversion:")
	priorityStrings := []string{"lowest", "normal", "high", "critical", "invalid"}
	for _, s := range priorityStrings {
		p, err := queue.PriorityFromString(s)
		if err != nil {
			fmt.Printf("  - '%s': Error - %v\n", s, err)
		} else {
			fmt.Printf("  - '%s': %s (%d)\n", s, p.String(), int64(p))
		}
	}
}

func requestBuilderDemo() {
	fmt.Println("Building requests using RequestBuilder:")

	// Example 1: Basic request
	request1, err := queue.NewRequestBuilder().
		WithID("req-001").
		WithUserID("user-alice").
		WithAgentID("agent-gpt4").
		WithPriority(queue.PriorityHigh).
		WithPayload(map[string]interface{}{
			"message": "Hello, can you help me with coding?",
			"context": "Go programming",
		}).
		WithMetadata("source", "web-ui").
		WithMetadata("session_id", "sess-123").
		Build()

	if err != nil {
		log.Printf("Error building request 1: %v", err)
	} else {
		fmt.Printf("  Request 1: ID=%s, User=%s, Priority=%s\n",
			request1.ID, request1.UserID, request1.Priority.String())
		fmt.Printf("    Metadata: %+v\n", request1.Metadata)
	}

	// Example 2: Request with TTL
	request2, err := queue.NewRequestBuilder().
		WithID("req-002").
		WithUserID("user-bob").
		WithAgentID("agent-claude").
		WithPriority(queue.PriorityCritical).
		WithPayload("Emergency: System is down!").
		WithTTL(5 * time.Minute).
		Build()

	if err != nil {
		log.Printf("Error building request 2: %v", err)
	} else {
		fmt.Printf("  Request 2: ID=%s, Priority=%s, TTL=5min\n",
			request2.ID, request2.Priority.String())
		if request2.ExpiresAt != nil {
			fmt.Printf("    Expires at: %s\n", request2.ExpiresAt.Format(time.RFC3339))
		}
	}

	// Example 3: Invalid request
	_, err = queue.NewRequestBuilder().
		WithID("req-003").
		WithPriority(queue.PriorityNormal).
		Build()

	if err != nil {
		fmt.Printf("  Request 3: Build failed as expected - %v\n", err)
	}
}

func queueNameBuilderDemo() {
	fmt.Println("Building queue names using QueueNameBuilder:")

	name1 := queue.NewQueueNameBuilder().
		WithAgent("agent-123").
		Build()
	fmt.Printf("  Agent queue: %s\n", name1)

	name2 := queue.NewQueueNameBuilder().
		WithService("chat").
		WithRegion("us-west-1").
		Build()
	fmt.Printf("  Service queue: %s\n", name2)

	name3 := queue.NewQueueNameBuilder().
		WithService("translation").
		WithRegion("eu-central-1").
		WithAgent("agent-translator").
		WithCustom("language", "spanish").
		Build()
	fmt.Printf("  Complex queue: %s\n", name3)

	name4 := queue.NewQueueNameBuilder().Build()
	fmt.Printf("  Default queue: %s\n", name4)
}
