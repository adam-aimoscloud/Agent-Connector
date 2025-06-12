package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOpenAIAgent(t *testing.T) {
	tests := []struct {
		name     string
		config   *OpenAIConfig
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid config",
			config: &OpenAIConfig{
				AgentConfig: AgentConfig{
					ID:   "test-openai",
					Name: "Test OpenAI Agent",
					Type: AgentTypeOpenAI,
				},
				BaseURL: "https://api.openai.com",
				APIKey:  "test-key",
			},
			wantErr: false,
		},
		{
			name: "Missing API key",
			config: &OpenAIConfig{
				AgentConfig: AgentConfig{
					ID:   "test-openai",
					Name: "Test OpenAI Agent",
					Type: AgentTypeOpenAI,
				},
				BaseURL: "https://api.openai.com",
				APIKey:  "",
			},
			wantErr:  true,
			errorMsg: "API key is required",
		},
		{
			name: "Missing base URL",
			config: &OpenAIConfig{
				AgentConfig: AgentConfig{
					ID:   "test-openai",
					Name: "Test OpenAI Agent",
					Type: AgentTypeOpenAI,
				},
				BaseURL: "",
				APIKey:  "test-key",
			},
			wantErr:  true,
			errorMsg: "base URL is required",
		},
		{
			name:     "Nil config",
			config:   nil,
			wantErr:  true,
			errorMsg: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewOpenAIAgent(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAIAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got %v", tt.errorMsg, err)
				}
			} else {
				if agent == nil {
					t.Error("Expected valid agent, got nil")
				}
			}
		})
	}
}

func TestOpenAIAgent_BasicMethods(t *testing.T) {
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:       "test-openai",
			Name:     "Test OpenAI Agent",
			Type:     AgentTypeOpenAI,
			Priority: 100,
		},
		BaseURL:      "https://api.openai.com",
		APIKey:       "test-key",
		DefaultModel: "gpt-3.5-turbo",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test basic getters
	if agent.GetID() != "test-openai" {
		t.Errorf("Expected ID 'test-openai', got %s", agent.GetID())
	}

	if agent.GetName() != "Test OpenAI Agent" {
		t.Errorf("Expected name 'Test OpenAI Agent', got %s", agent.GetName())
	}

	if agent.GetType() != AgentTypeOpenAI {
		t.Errorf("Expected type %s, got %s", AgentTypeOpenAI, agent.GetType())
	}

	// Test capabilities
	capabilities := agent.GetCapabilities()
	if !capabilities.SupportsChatCompletion {
		t.Error("Expected SupportsChatCompletion to be true")
	}
	if !capabilities.SupportsStreaming {
		t.Error("Expected SupportsStreaming to be true")
	}
	if !capabilities.SupportsFunctionCalling {
		t.Error("Expected SupportsFunctionCalling to be true")
	}
}

func TestOpenAIAgent_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path /v1/chat/completions, got %s", r.URL.Path)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("Expected Bearer token, got %s", auth)
		}

		// Send mock response
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
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 15,
				"total_tokens": 25
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL:      server.URL,
		APIKey:       "test-key",
		DefaultModel: "gpt-3.5-turbo",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test chat
	req := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Model: "gpt-3.5-turbo",
	}

	ctx := context.Background()
	resp, err := agent.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Error("Expected at least one choice in response")
	}

	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", resp.Choices[0].Message.Role)
	}

	if resp.Choices[0].Message.Content != "Hello! How can I help you today?" {
		t.Errorf("Unexpected response content: %s", resp.Choices[0].Message.Content)
	}

	if resp.Usage.TotalTokens != 25 {
		t.Errorf("Expected total tokens 25, got %d", resp.Usage.TotalTokens)
	}
}

func TestOpenAIAgent_ChatWithError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key", "type": "invalid_request_error"}}`))
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL,
		APIKey:  "invalid-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
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
	_, err = agent.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error message to contain 'Invalid API key', got: %v", err)
	}
}

func TestOpenAIAgent_GetModels(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("Expected path /v1/models, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := `{
			"object": "list",
			"data": [
				{
					"id": "gpt-3.5-turbo",
					"object": "model",
					"created": 1677610602,
					"owned_by": "openai"
				},
				{
					"id": "gpt-4",
					"object": "model",
					"created": 1687882411,
					"owned_by": "openai"
				}
			]
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL,
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()
	models, err := agent.GetModels(ctx)
	if err != nil {
		t.Fatalf("GetModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	expectedModels := []string{"gpt-3.5-turbo", "gpt-4"}
	for i, model := range models {
		if model.ID != expectedModels[i] {
			t.Errorf("Expected model ID %s, got %s", expectedModels[i], model.ID)
		}
	}
}

func TestOpenAIAgent_ValidateConfig(t *testing.T) {
	agent := &OpenAIAgent{}

	// ValidateConfig method doesn't take parameters according to interface
	err := agent.ValidateConfig()
	// This will test the current agent's configuration validation
	// Since we haven't set a config, it should return an error
	if err == nil {
		t.Error("Expected error for uninitialized agent")
	}
}

func TestOpenAIAgent_Status(t *testing.T) {
	// Create mock server for health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"object": "list", "data": []}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server instead of real API
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test initial status
	ctx := context.Background()
	status, err := agent.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	// With mock server, health check should succeed
	if status == nil {
		t.Error("Expected status to be returned")
	}
	if status.Status != "active" {
		t.Errorf("Expected status 'active', got %s", status.Status)
	}
	if !status.Health {
		t.Error("Expected health to be true with mock server")
	}
}

func TestOpenAIAgent_Close(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"object": "list", "data": []}`))
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL, // Use mock server instead of real API
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test close
	err = agent.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Check status after close
	ctx := context.Background()
	status, err := agent.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.Status != "inactive" {
		t.Errorf("Expected status 'inactive' after close, got %s", status.Status)
	}
}

func TestOpenAIAgent_HealthCheck(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"object": "list", "data": []}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "test-openai",
			Name: "Test OpenAI Agent",
			Type: AgentTypeOpenAI,
			HealthCheck: &HealthCheckConfig{
				Enabled:          true,
				Interval:         time.Second,
				Timeout:          5 * time.Second,
				FailureThreshold: 3,
			},
		},
		BaseURL: server.URL,
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	// Wait for health check to run
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	status, err := agent.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.Health {
		t.Error("Expected agent to be healthy after successful health check")
	}
}

func BenchmarkOpenAIAgent_Chat(b *testing.B) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "bench-openai",
			Name: "Benchmark OpenAI Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: server.URL,
		APIKey:  "test-key",
	}

	agent, err := NewOpenAIAgent(config)
	if err != nil {
		b.Fatalf("Failed to create agent: %v", err)
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
		_, err := agent.Chat(ctx, req)
		if err != nil {
			b.Errorf("Chat failed: %v", err)
		}
	}
}
