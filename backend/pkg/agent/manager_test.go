package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// createMockServer creates a mock HTTP server for testing
func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OpenAI API endpoints
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"object": "list", "data": []}`))
			return
		}
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-3.5-turbo",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Hello!"
					},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 5,
					"completion_tokens": 2,
					"total_tokens": 7
				}
			}`
			w.Write([]byte(response))
			return
		}

		// Handle Dify API endpoints (POST requests)
		if r.URL.Path == "/v1/parameters" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"result": "success"}`))
			return
		}

		// Default response for any other requests
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
}

func TestNewAgentManager(t *testing.T) {
	tests := []struct {
		name   string
		config *AgentManagerConfig
	}{
		{
			name:   "Default config",
			config: nil,
		},
		{
			name: "Custom config",
			config: &AgentManagerConfig{
				LoadBalancingStrategy: RoundRobin,
				EnableHealthChecks:    true,
				HealthCheckInterval:   30 * time.Second,
				DefaultTimeout:        10 * time.Second,
				MaxRetries:            3,
				EnableMetrics:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewAgentManager(tt.config)
			if err != nil {
				t.Fatalf("NewAgentManager failed: %v", err)
			}
			if manager == nil {
				t.Error("Expected valid manager, got nil")
			}
		})
	}
}

func TestAgentManager_RegisterAgent(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create test agent
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-agent",
			Name: "Test Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test registration
	err = manager.RegisterAgent(agent)
	if err != nil {
		t.Errorf("RegisterAgent failed: %v", err)
	}

	// Test duplicate registration
	err = manager.RegisterAgent(agent)
	if err == nil {
		t.Error("Expected error for duplicate agent registration")
	}

	// Test nil agent
	err = manager.RegisterAgent(nil)
	if err == nil {
		t.Error("Expected error for nil agent")
	}
}

func TestAgentManager_UnregisterAgent(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create and register test agent
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-agent",
			Name: "Test Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	err = manager.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	// Test unregistration
	err = manager.UnregisterAgent("test-agent")
	if err != nil {
		t.Errorf("UnregisterAgent failed: %v", err)
	}

	// Test unregistering non-existent agent
	err = manager.UnregisterAgent("non-existent")
	if err == nil {
		t.Error("Expected error for unregistering non-existent agent")
	}
}

func TestAgentManager_GetAgent(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create and register test agent
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-agent",
			Name: "Test Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	err = manager.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	// Test getting existing agent
	retrievedAgent, err := manager.GetAgent("test-agent")
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if retrievedAgent.GetID() != "test-agent" {
		t.Errorf("Expected agent ID 'test-agent', got %s", retrievedAgent.GetID())
	}

	// Test getting non-existent agent
	_, err = manager.GetAgent("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent agent")
	}
}

func TestAgentManager_ListAgents(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Initially should be empty
	agents := manager.ListAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents initially, got %d", len(agents))
	}

	// Create and register test agents
	configs := []*OpenAIConfig{
		{
			AgentConfig: AgentConfig{
				ID:   "agent-1",
				Name: "Agent 1",
				Type: AgentTypeOpenAI,
			},
			BaseURL: server.URL, // Use mock server
			APIKey:  "test-key-1",
		},
		{
			AgentConfig: AgentConfig{
				ID:   "agent-2",
				Name: "Agent 2",
				Type: AgentTypeOpenAI,
			},
			BaseURL: server.URL, // Use mock server
			APIKey:  "test-key-2",
		},
	}

	for _, config := range configs {
		agent, err := NewOpenAIAgent(config)
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}

		err = manager.RegisterAgent(agent)
		if err != nil {
			t.Fatalf("RegisterAgent failed: %v", err)
		}
	}

	// Test listing agents
	agents = manager.ListAgents()
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}
}

func TestAgentManager_ListAgentsByType(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create OpenAI agent
	openaiConfig := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "openai-agent",
			Name: "OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	openaiAgent, err := NewOpenAIAgent(openaiConfig)
	if err != nil {
		t.Fatalf("Failed to create OpenAI agent: %v", err)
	}

	// Create Dify agent
	difyConfig := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "dify-agent",
			Name: "Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
		AppID:   "app-123",
	}

	difyAgent, err := NewDifyAgent(difyConfig)
	if err != nil {
		t.Fatalf("Failed to create Dify agent: %v", err)
	}

	// Register agents
	err = manager.RegisterAgent(openaiAgent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	err = manager.RegisterAgent(difyAgent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	// Test listing by type
	openaiAgents := manager.ListAgentsByType(AgentTypeOpenAI)
	if len(openaiAgents) != 1 {
		t.Errorf("Expected 1 OpenAI agent, got %d", len(openaiAgents))
	}

	difyAgents := manager.ListAgentsByType(AgentTypeDify)
	if len(difyAgents) != 1 {
		t.Errorf("Expected 1 Dify agent, got %d", len(difyAgents))
	}

	// Test with non-existent type
	unknownAgents := manager.ListAgentsByType(AgentType("unknown"))
	if len(unknownAgents) != 0 {
		t.Errorf("Expected 0 unknown agents, got %d", len(unknownAgents))
	}
}

func TestAgentManager_GetAvailableAgent(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &AgentManagerConfig{
		LoadBalancingStrategy: Priority,
	}
	manager, err := NewAgentManager(config)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create agents with different priorities
	configs := []*OpenAIConfig{
		{
			AgentConfig: AgentConfig{
				ID:       "high-priority",
				Name:     "High Priority Agent",
				Type:     AgentTypeOpenAI,
				Priority: 100,
				Enabled:  true,
			},
			BaseURL: server.URL, // Use mock server
			APIKey:  "test-key-1",
		},
		{
			AgentConfig: AgentConfig{
				ID:       "low-priority",
				Name:     "Low Priority Agent",
				Type:     AgentTypeOpenAI,
				Priority: 50,
				Enabled:  true,
			},
			BaseURL: server.URL, // Use mock server
			APIKey:  "test-key-2",
		},
	}

	for _, config := range configs {
		agent, err := NewOpenAIAgent(config)
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}

		err = manager.RegisterAgent(agent)
		if err != nil {
			t.Fatalf("RegisterAgent failed: %v", err)
		}
	}

	// Test getting available agent
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()
	agent, err := manager.GetAvailableAgent(ctx, req)
	if err != nil {
		t.Fatalf("GetAvailableAgent failed: %v", err)
	}

	// Should get high priority agent
	if agent == nil {
		t.Error("Expected an agent to be returned")
	}
	if agent != nil && agent.GetID() != "high-priority" {
		t.Errorf("Expected high-priority agent, got %s", agent.GetID())
	}
}

func TestAgentManager_LoadBalancingStrategies(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	strategies := []LoadBalancingStrategy{
		Priority,
		RoundRobin,
		Random,
		WeightedRandom,
		LeastConnections,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			config := &AgentManagerConfig{
				LoadBalancingStrategy: strategy,
			}
			manager, err := NewAgentManager(config)
			if err != nil {
				t.Fatalf("NewAgentManager failed: %v", err)
			}

			// Create test agent
			agentConfig := &OpenAIConfig{
				AgentConfig: AgentConfig{
					ID:      "test-agent",
					Name:    "Test Agent",
					Type:    AgentTypeOpenAI,
					Enabled: true,
				},
				BaseURL: server.URL, // Use mock server
				APIKey:  "test-key",
			}

			agent, err := NewOpenAIAgent(agentConfig)
			if err != nil {
				t.Fatalf("Failed to create agent: %v", err)
			}

			err = manager.RegisterAgent(agent)
			if err != nil {
				t.Fatalf("RegisterAgent failed: %v", err)
			}

			// Test getting available agent with different strategies
			req := &ChatRequest{
				Messages: []Message{
					{
						Role:    "user",
						Content: "Hello",
					},
				},
			}

			ctx := context.Background()
			selectedAgent, err := manager.GetAvailableAgent(ctx, req)
			if err != nil {
				t.Fatalf("GetAvailableAgent failed for strategy %s: %v", strategy, err)
			}

			if selectedAgent == nil {
				t.Errorf("Expected an agent for strategy %s", strategy)
			}
		})
	}
}

func TestAgentManager_Close(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create and register test agent
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-agent",
			Name: "Test Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	err = manager.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	// Test close
	err = manager.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// After close, agents list should be empty (or agents should be closed)
	agents := manager.ListAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents after close, got %d", len(agents))
	}
}

func TestAgentManager_HealthChecks(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	config := &AgentManagerConfig{
		EnableHealthChecks:  true,
		HealthCheckInterval: 100 * time.Millisecond, // Short interval for testing
	}
	manager, err := NewAgentManager(config)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create test agent
	agentConfig := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Type:    AgentTypeOpenAI,
			Enabled: true,
		},
		BaseURL: server.URL, // Use mock server
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(agentConfig)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	err = manager.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	// Wait for health check to run
	time.Sleep(200 * time.Millisecond)

	// Clean up
	manager.Close()
}

func TestAgentManager_EmptyManagerBehavior(t *testing.T) {
	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("NewAgentManager failed: %v", err)
	}

	// Test getting agent from empty manager
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()
	_, err = manager.GetAvailableAgent(ctx, req)
	if err == nil {
		t.Error("Expected error when getting agent from empty manager")
	}

	// Test listing from empty manager
	agents := manager.ListAgents()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents from empty manager, got %d", len(agents))
	}
}

func BenchmarkAgentManager_GetAvailableAgent(b *testing.B) {
	server := createMockServer()
	defer server.Close()

	manager, err := NewAgentManager(nil)
	if err != nil {
		b.Fatalf("NewAgentManager failed: %v", err)
	}

	// Create multiple test agents
	for i := 0; i < 10; i++ {
		config := &OpenAIConfig{
			AgentConfig: AgentConfig{
				ID:      fmt.Sprintf("agent-%d", i),
				Name:    fmt.Sprintf("Agent %d", i),
				Type:    AgentTypeOpenAI,
				Enabled: true,
			},
			BaseURL: server.URL, // Use mock server
			APIKey:  "test-key",
		}

		agent, err := NewOpenAIAgent(config)
		if err != nil {
			b.Fatalf("Failed to create agent: %v", err)
		}

		err = manager.RegisterAgent(agent)
		if err != nil {
			b.Fatalf("RegisterAgent failed: %v", err)
		}
	}

	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetAvailableAgent(ctx, req)
		if err != nil {
			b.Errorf("GetAvailableAgent failed: %v", err)
		}
	}
}
