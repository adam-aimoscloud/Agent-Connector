package agent

import (
	"context"
	"testing"
	"time"
)

func TestAgentFactory_CreateAgent(t *testing.T) {
	factory := NewAgentFactory()

	tests := []struct {
		name      string
		agentType AgentType
		config    interface{}
		wantErr   bool
	}{
		{
			name:      "Create OpenAI agent",
			agentType: AgentTypeOpenAI,
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
			name:      "Create Dify agent",
			agentType: AgentTypeDify,
			config: &DifyConfig{
				AgentConfig: AgentConfig{
					ID:   "test-dify",
					Name: "Test Dify Agent",
					Type: AgentTypeDify,
				},
				BaseURL: "https://api.dify.ai",
				APIKey:  "test-key",
				AppID:   "test-app",
			},
			wantErr: false,
		},
		{
			name:      "Invalid agent type",
			agentType: AgentType("invalid"),
			config:    &OpenAIConfig{},
			wantErr:   true,
		},
		{
			name:      "Wrong config type for OpenAI",
			agentType: AgentTypeOpenAI,
			config:    &DifyConfig{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := factory.CreateAgent(tt.agentType, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && agent == nil {
				t.Error("CreateAgent() returned nil agent without error")
			}
		})
	}
}

func TestOpenAIConfigBuilder(t *testing.T) {
	config := NewOpenAIConfigBuilder().
		WithID("test-agent").
		WithName("Test Agent").
		WithBaseURL("https://api.openai.com").
		WithAPIKey("test-key").
		WithDefaultModel("gpt-3.5-turbo").
		WithMaxTokens(2048).
		WithTemperature(0.8).
		WithPriority(100).
		Build()

	if config.ID != "test-agent" {
		t.Errorf("Expected ID to be 'test-agent', got %s", config.ID)
	}

	if config.Name != "Test Agent" {
		t.Errorf("Expected Name to be 'Test Agent', got %s", config.Name)
	}

	if config.BaseURL != "https://api.openai.com" {
		t.Errorf("Expected BaseURL to be 'https://api.openai.com', got %s", config.BaseURL)
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey to be 'test-key', got %s", config.APIKey)
	}

	if config.DefaultModel != "gpt-3.5-turbo" {
		t.Errorf("Expected DefaultModel to be 'gpt-3.5-turbo', got %s", config.DefaultModel)
	}

	if config.MaxTokens != 2048 {
		t.Errorf("Expected MaxTokens to be 2048, got %d", config.MaxTokens)
	}

	if config.Temperature != 0.8 {
		t.Errorf("Expected Temperature to be 0.8, got %f", config.Temperature)
	}

	if config.Priority != 100 {
		t.Errorf("Expected Priority to be 100, got %d", config.Priority)
	}
}

func TestDifyConfigBuilder(t *testing.T) {
	config := NewDifyConfigBuilder().
		WithID("test-dify").
		WithName("Test Dify Agent").
		WithBaseURL("https://api.dify.ai").
		WithAPIKey("dify-key").
		WithAppID("app-123").
		WithAppType("agent").
		WithVersion("v1").
		WithPriority(80).
		Build()

	if config.ID != "test-dify" {
		t.Errorf("Expected ID to be 'test-dify', got %s", config.ID)
	}

	if config.Name != "Test Dify Agent" {
		t.Errorf("Expected Name to be 'Test Dify Agent', got %s", config.Name)
	}

	if config.BaseURL != "https://api.dify.ai" {
		t.Errorf("Expected BaseURL to be 'https://api.dify.ai', got %s", config.BaseURL)
	}

	if config.APIKey != "dify-key" {
		t.Errorf("Expected APIKey to be 'dify-key', got %s", config.APIKey)
	}

	if config.AppID != "app-123" {
		t.Errorf("Expected AppID to be 'app-123', got %s", config.AppID)
	}

	if config.AppType != "agent" {
		t.Errorf("Expected AppType to be 'agent', got %s", config.AppType)
	}

	if config.Version != "v1" {
		t.Errorf("Expected Version to be 'v1', got %s", config.Version)
	}

	if config.Priority != 80 {
		t.Errorf("Expected Priority to be 80, got %d", config.Priority)
	}
}

func TestRetryPolicyBuilder(t *testing.T) {
	policy := NewRetryPolicyBuilder().
		WithMaxRetries(5).
		WithInitialDelay(2 * time.Second).
		WithMaxDelay(60 * time.Second).
		WithMultiplier(2.5).
		WithRetryableErrors([]string{"timeout", "rate_limit"}).
		Build()

	if policy.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 2*time.Second {
		t.Errorf("Expected InitialDelay to be 2s, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 60*time.Second {
		t.Errorf("Expected MaxDelay to be 60s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.5 {
		t.Errorf("Expected Multiplier to be 2.5, got %f", policy.Multiplier)
	}

	if len(policy.RetryableErrors) != 2 {
		t.Errorf("Expected 2 retryable errors, got %d", len(policy.RetryableErrors))
	}
}

func TestHealthCheckConfigBuilder(t *testing.T) {
	config := NewHealthCheckConfigBuilder().
		WithEnabled(true).
		WithInterval(30 * time.Second).
		WithTimeout(5 * time.Second).
		WithFailureThreshold(5).
		WithSuccessThreshold(2).
		Build()

	if !config.Enabled {
		t.Error("Expected health check to be enabled")
	}

	if config.Interval != 30*time.Second {
		t.Errorf("Expected Interval to be 30s, got %v", config.Interval)
	}

	if config.Timeout != 5*time.Second {
		t.Errorf("Expected Timeout to be 5s, got %v", config.Timeout)
	}

	if config.FailureThreshold != 5 {
		t.Errorf("Expected FailureThreshold to be 5, got %d", config.FailureThreshold)
	}

	if config.SuccessThreshold != 2 {
		t.Errorf("Expected SuccessThreshold to be 2, got %d", config.SuccessThreshold)
	}
}

func TestPresetConfigs(t *testing.T) {
	presets := NewPresetConfigs()

	// Test OpenAI GPT-3.5-turbo preset
	openaiConfig := presets.OpenAIGPT35Turbo("gpt35", "GPT-3.5 Agent", "sk-test123")
	if openaiConfig.ID != "gpt35" {
		t.Errorf("Expected ID to be 'gpt35', got %s", openaiConfig.ID)
	}
	if openaiConfig.DefaultModel != "gpt-3.5-turbo" {
		t.Errorf("Expected model to be 'gpt-3.5-turbo', got %s", openaiConfig.DefaultModel)
	}

	// Test OpenAI GPT-4 preset
	gpt4Config := presets.OpenAIGPT4("gpt4", "GPT-4 Agent", "sk-test456")
	if gpt4Config.DefaultModel != "gpt-4" {
		t.Errorf("Expected model to be 'gpt-4', got %s", gpt4Config.DefaultModel)
	}
	if gpt4Config.MaxTokens != 8192 {
		t.Errorf("Expected MaxTokens to be 8192, got %d", gpt4Config.MaxTokens)
	}

	// Test Azure OpenAI preset
	azureConfig := presets.AzureOpenAI("azure", "Azure OpenAI", "https://test.openai.azure.com", "key123", "gpt-35-turbo")
	if azureConfig.BaseURL != "https://test.openai.azure.com" {
		t.Errorf("Expected BaseURL to be Azure endpoint, got %s", azureConfig.BaseURL)
	}
	if azureConfig.CustomHeaders["api-version"] != "2023-12-01-preview" {
		t.Error("Expected Azure API version header")
	}

	// Test Dify chatbot preset
	difyConfig := presets.DifyChatbot("dify1", "Dify Bot", "https://api.dify.ai", "dify-key", "app-123")
	if difyConfig.AppType != "chatbot" {
		t.Errorf("Expected AppType to be 'chatbot', got %s", difyConfig.AppType)
	}

	// Test Dify agent preset
	difyAgentConfig := presets.DifyAgent("dify2", "Dify Agent", "https://api.dify.ai", "dify-key", "app-456")
	if difyAgentConfig.AppType != "agent" {
		t.Errorf("Expected AppType to be 'agent', got %s", difyAgentConfig.AppType)
	}
}

func TestAgentManager(t *testing.T) {
	manager, err := NewAgentManager(nil)
	if err != nil {
		t.Fatalf("Failed to create agent manager: %v", err)
	}
	defer manager.Close()

	// Create test agents
	openaiConfig := NewOpenAIConfigBuilder().
		WithID("test-openai").
		WithName("Test OpenAI").
		WithBaseURL("https://api.openai.com").
		WithAPIKey("test-key").
		Build()

	openaiAgent, err := NewOpenAIAgent(openaiConfig)
	if err != nil {
		t.Fatalf("Failed to create OpenAI agent: %v", err)
	}

	difyConfig := NewDifyConfigBuilder().
		WithID("test-dify").
		WithName("Test Dify").
		WithBaseURL("https://api.dify.ai").
		WithAPIKey("test-key").
		WithAppID("test-app").
		Build()

	difyAgent, err := NewDifyAgent(difyConfig)
	if err != nil {
		t.Fatalf("Failed to create Dify agent: %v", err)
	}

	// Test agent registration
	if err := manager.RegisterAgent(openaiAgent); err != nil {
		t.Errorf("Failed to register OpenAI agent: %v", err)
	}

	if err := manager.RegisterAgent(difyAgent); err != nil {
		t.Errorf("Failed to register Dify agent: %v", err)
	}

	// Test duplicate registration
	if err := manager.RegisterAgent(openaiAgent); err == nil {
		t.Error("Expected error when registering duplicate agent")
	}

	// Test agent retrieval
	retrievedAgent, err := manager.GetAgent("test-openai")
	if err != nil {
		t.Errorf("Failed to get agent: %v", err)
	}
	if retrievedAgent.GetID() != "test-openai" {
		t.Errorf("Retrieved wrong agent: %s", retrievedAgent.GetID())
	}

	// Test listing agents
	agents := manager.ListAgents()
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
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

	// Test agent info
	ctx := context.Background()
	info, err := manager.GetAgentInfo(ctx, "test-openai")
	if err != nil {
		t.Errorf("Failed to get agent info: %v", err)
	}
	if info == nil || info.ID != "test-openai" {
		t.Error("Invalid agent info")
	}

	// Test unregistration
	if err := manager.UnregisterAgent("test-openai"); err != nil {
		t.Errorf("Failed to unregister agent: %v", err)
	}

	// Verify agent is removed
	if _, err := manager.GetAgent("test-openai"); err == nil {
		t.Error("Expected error when getting unregistered agent")
	}

	remainingAgents := manager.ListAgents()
	if len(remainingAgents) != 1 {
		t.Errorf("Expected 1 remaining agent, got %d", len(remainingAgents))
	}
}

func TestAgentType(t *testing.T) {
	tests := []struct {
		agentType AgentType
		valid     bool
		str       string
	}{
		{AgentTypeOpenAI, true, "openai"},
		{AgentTypeDify, true, "dify"},
		{AgentType("invalid"), false, "invalid"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agentType), func(t *testing.T) {
			if tt.agentType.IsValid() != tt.valid {
				t.Errorf("IsValid() = %v, want %v", tt.agentType.IsValid(), tt.valid)
			}

			if tt.agentType.String() != tt.str {
				t.Errorf("String() = %v, want %v", tt.agentType.String(), tt.str)
			}
		})
	}
}

func TestAgentError(t *testing.T) {
	err := &AgentError{
		Code:    "invalid_request",
		Message: "Invalid request format",
		Type:    "validation_error",
	}

	if err.Error() != "Invalid request format" {
		t.Errorf("Error() = %v, want %v", err.Error(), "Invalid request format")
	}
}

func BenchmarkAgentFactory_CreateAgent(b *testing.B) {
	factory := NewAgentFactory()
	config := &OpenAIConfig{
		AgentConfig: AgentConfig{
			ID:   "bench-test",
			Name: "Benchmark Agent",
			Type: AgentTypeOpenAI,
		},
		BaseURL: "https://api.openai.com",
		APIKey:  "test-key",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent, err := factory.CreateAgent(AgentTypeOpenAI, config)
		if err != nil {
			b.Fatalf("Failed to create agent: %v", err)
		}
		agent.Close()
	}
}

func BenchmarkOpenAIConfigBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOpenAIConfigBuilder().
			WithID("bench-test").
			WithName("Benchmark Agent").
			WithBaseURL("https://api.openai.com").
			WithAPIKey("test-key").
			WithDefaultModel("gpt-3.5-turbo").
			Build()
	}
}
