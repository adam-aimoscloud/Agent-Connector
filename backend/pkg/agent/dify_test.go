package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewDifyAgent(t *testing.T) {
	tests := []struct {
		name     string
		config   *DifyConfig
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid config",
			config: &DifyConfig{
				AgentConfig: AgentConfig{
					ID:   "test-dify",
					Name: "Test Dify Agent",
					Type: AgentTypeDify,
				},
				BaseURL: "https://api.dify.ai",
				APIKey:  "test-key",
				AppID:   "app-123",
			},
			wantErr: false,
		},
		{
			name: "Missing API key",
			config: &DifyConfig{
				AgentConfig: AgentConfig{
					ID:   "test-dify",
					Name: "Test Dify Agent",
					Type: AgentTypeDify,
				},
				BaseURL: "https://api.dify.ai",
				APIKey:  "",
				AppID:   "app-123",
			},
			wantErr:  true,
			errorMsg: "API key is required",
		},
		{
			name: "Missing App ID",
			config: &DifyConfig{
				AgentConfig: AgentConfig{
					ID:   "test-dify",
					Name: "Test Dify Agent",
					Type: AgentTypeDify,
				},
				BaseURL: "https://api.dify.ai",
				APIKey:  "test-key",
				AppID:   "",
			},
			wantErr:  true,
			errorMsg: "app ID is required",
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
			agent, err := NewDifyAgent(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDifyAgent() error = %v, wantErr %v", err, tt.wantErr)
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

func TestDifyAgent_BasicMethods(t *testing.T) {
	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:       "test-dify",
			Name:     "Test Dify Agent",
			Type:     AgentTypeDify,
			Priority: 80,
		},
		BaseURL: "https://api.dify.ai",
		APIKey:  "test-key",
		AppID:   "app-123",
		AppType: "agent",
	}

	agent, err := NewDifyAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test basic getters
	if agent.GetID() != "test-dify" {
		t.Errorf("Expected ID 'test-dify', got %s", agent.GetID())
	}

	if agent.GetName() != "Test Dify Agent" {
		t.Errorf("Expected name 'Test Dify Agent', got %s", agent.GetName())
	}

	if agent.GetType() != AgentTypeDify {
		t.Errorf("Expected type %s, got %s", AgentTypeDify, agent.GetType())
	}

	// Test capabilities
	capabilities := agent.GetCapabilities()
	if !capabilities.SupportsChatCompletion {
		t.Error("Expected SupportsChatCompletion to be true")
	}
	if !capabilities.SupportsStreaming {
		t.Error("Expected SupportsStreaming to be true")
	}
	if !capabilities.SupportsFiles {
		t.Error("Expected SupportsFiles to be true")
	}
}

func TestDifyAgent_Chat(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/chat-messages") {
			t.Errorf("Expected path to start with /v1/chat-messages, got %s", r.URL.Path)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("Expected Bearer token, got %s", auth)
		}

		// Send mock response
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"id": "msg-123",
			"conversation_id": "conv-456",
			"mode": "chat",
			"answer": "Hello! How can I help you today?",
			"created_at": 1677652288,
			"metadata": {
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 15,
					"total_tokens": 25
				}
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "test-dify",
			Name: "Test Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL,
		APIKey:  "test-key",
		AppID:   "app-123",
		AppType: "chatbot",
	}

	agent, err := NewDifyAgent(config)
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
		UserID: "user-123",
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

func TestDifyAgent_ChatWithError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code": "unauthorized", "message": "Invalid API key", "status": 401}`))
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "test-dify",
			Name: "Test Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL,
		APIKey:  "invalid-key",
		AppID:   "app-123",
	}

	agent, err := NewDifyAgent(config)
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
		UserID: "user-123",
	}

	ctx := context.Background()
	_, err = agent.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	if !strings.Contains(err.Error(), "HTTP error") && !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error message to contain HTTP error or Invalid API key, got: %v", err)
	}
}

func TestDifyAgent_GetModels(t *testing.T) {
	// Create mock server (not actually used by GetModels but needed for agent creation)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "test-dify",
			Name: "Test Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL, // Use mock server instead of real API
		APIKey:  "test-key",
		AppID:   "app-123",
	}

	agent, err := NewDifyAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()
	models, err := agent.GetModels(ctx)
	if err != nil {
		t.Fatalf("GetModels failed: %v", err)
	}

	// Dify typically returns app-specific models
	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	// Check that the app ID is used as model ID
	found := false
	for _, model := range models {
		if model.ID == "app-123" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected app ID to be included as model")
	}
}

func TestDifyAgent_ValidateConfig(t *testing.T) {
	agent := &DifyAgent{}

	// ValidateConfig method doesn't take parameters according to interface
	err := agent.ValidateConfig()
	// This will test the current agent's configuration validation
	// Since we haven't set a config, it should return an error
	if err == nil {
		t.Error("Expected error for uninitialized agent")
	}
}

func TestDifyAgent_Status(t *testing.T) {
	// Create mock server for health check
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/parameters" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"result": "success"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "test-dify",
			Name: "Test Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL, // Use mock server instead of real API
		APIKey:  "test-key",
		AppID:   "app-123",
	}

	agent, err := NewDifyAgent(config)
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

func TestDifyAgent_Close(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "test-dify",
			Name: "Test Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL, // Use mock server instead of real API
		APIKey:  "test-key",
		AppID:   "app-123",
	}

	agent, err := NewDifyAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Test close
	err = agent.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// After close, don't call GetStatus since httpClient is nil
	// Just verify Close() doesn't return error
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}

func TestDifyAgent_AppTypes(t *testing.T) {
	tests := []struct {
		name    string
		appType string
		wantErr bool
	}{
		{
			name:    "Chatbot app type",
			appType: "chatbot",
			wantErr: false,
		},
		{
			name:    "Agent app type",
			appType: "agent",
			wantErr: false,
		},
		{
			name:    "Workflow app type",
			appType: "workflow",
			wantErr: false,
		},
		{
			name:    "Empty app type (should default)",
			appType: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &DifyConfig{
				AgentConfig: AgentConfig{
					ID:   "test-dify",
					Name: "Test Dify Agent",
					Type: AgentTypeDify,
				},
				BaseURL: "https://api.dify.ai",
				APIKey:  "test-key",
				AppID:   "app-123",
				AppType: tt.appType,
			}

			agent, err := NewDifyAgent(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDifyAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && agent == nil {
				t.Error("Expected valid agent, got nil")
			}
		})
	}
}

func BenchmarkDifyAgent_Chat(b *testing.B) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := `{
			"id": "msg-123",
			"conversation_id": "conv-456",
			"mode": "chat",
			"answer": "Hello!",
			"created_at": 1677652288,
			"metadata": {
				"usage": {
					"prompt_tokens": 5,
					"completion_tokens": 2,
					"total_tokens": 7
				}
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := &DifyConfig{
		AgentConfig: AgentConfig{
			ID:   "bench-dify",
			Name: "Benchmark Dify Agent",
			Type: AgentTypeDify,
		},
		BaseURL: server.URL,
		APIKey:  "test-key",
		AppID:   "app-123",
	}

	agent, err := NewDifyAgent(config)
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
		UserID: "user-123",
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
