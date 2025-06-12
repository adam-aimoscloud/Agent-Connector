package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OpenAIAgent implements the Agent interface for OpenAI compatible APIs
type OpenAIAgent struct {
	config     *OpenAIConfig
	httpClient *http.Client
	status     *AgentStatus
	statusMu   sync.RWMutex // Mutex to protect status field
}

// OpenAIConfig represents configuration for OpenAI compatible agents
type OpenAIConfig struct {
	AgentConfig

	// BaseURL is the base URL for the OpenAI API
	BaseURL string `json:"base_url"`

	// APIKey is the API key for authentication
	APIKey string `json:"api_key"`

	// Organization is the organization ID (optional)
	Organization string `json:"organization,omitempty"`

	// DefaultModel is the default model to use
	DefaultModel string `json:"default_model"`

	// SupportedModels lists available models
	SupportedModels []string `json:"supported_models"`

	// MaxTokens is the default maximum tokens
	MaxTokens int `json:"max_tokens"`

	// Temperature is the default temperature
	Temperature float32 `json:"temperature"`

	// CustomHeaders for additional HTTP headers
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// NewOpenAIAgent creates a new OpenAI compatible agent
func NewOpenAIAgent(config *OpenAIConfig) (*OpenAIAgent, error) {
	if err := validateOpenAIConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set defaults
	setOpenAIDefaults(config)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	agent := &OpenAIAgent{
		config:     config,
		httpClient: httpClient,
		status: &AgentStatus{
			AgentID:     config.ID,
			Status:      "initializing",
			Health:      false,
			LastChecked: time.Now(),
		},
	}

	return agent, nil
}

// validateOpenAIConfig validates the OpenAI configuration
func validateOpenAIConfig(config *OpenAIConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.ID == "" {
		return fmt.Errorf("agent ID is required")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if !config.Type.IsValid() {
		return fmt.Errorf("invalid agent type: %s", config.Type)
	}

	return nil
}

// setOpenAIDefaults sets default values for OpenAI configuration
func setOpenAIDefaults(config *OpenAIConfig) {
	if config.Name == "" {
		config.Name = "OpenAI Agent"
	}

	if config.Type == "" {
		config.Type = AgentTypeOpenAI
	}

	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = DefaultMaxConcurrentRequests
	}

	if config.DefaultModel == "" {
		config.DefaultModel = "gpt-3.5-turbo"
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if len(config.SupportedModels) == 0 {
		config.SupportedModels = []string{
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-16k",
			"gpt-4",
			"gpt-4-32k",
		}
	}
}

// GetID returns the unique identifier of the agent
func (a *OpenAIAgent) GetID() string {
	return a.config.ID
}

// GetName returns the display name of the agent
func (a *OpenAIAgent) GetName() string {
	return a.config.Name
}

// GetType returns the type of the agent source
func (a *OpenAIAgent) GetType() AgentType {
	return AgentTypeOpenAI
}

// GetCapabilities returns the capabilities of the agent
func (a *OpenAIAgent) GetCapabilities() AgentCapabilities {
	return AgentCapabilities{
		SupportsChatCompletion:  true,
		SupportsStreaming:       true,
		SupportsImages:          true,
		SupportsFiles:           false,
		SupportsFunctionCalling: true,
		MaxTokens:               a.config.MaxTokens,
		SupportedLanguages:      []string{"en", "zh", "es", "fr", "de", "ja", "ko"},
	}
}

// Chat sends a chat message and returns the response
func (a *OpenAIAgent) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	// Prepare OpenAI request
	openaiReq := a.prepareOpenAIRequest(request)

	// Make HTTP request
	resp, err := a.makeRequest(ctx, "/v1/chat/completions", openaiReq)
	if err != nil {
		a.updateStatus(false, err)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var openaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		a.updateStatus(false, fmt.Errorf("failed to decode response: %w", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API errors
	if openaiResp.Error != nil {
		agentErr := &AgentError{
			Code:    openaiResp.Error.Code,
			Message: openaiResp.Error.Message,
			Type:    openaiResp.Error.Type,
		}
		a.updateStatus(false, agentErr)
		return nil, agentErr
	}

	// Convert to standard response
	response := a.convertToStandardResponse(&openaiResp)
	a.updateStatus(true, nil)

	return response, nil
}

// ChatStream sends a chat message and returns a streaming response
func (a *OpenAIAgent) ChatStream(ctx context.Context, request *ChatRequest) (*ChatStreamResponse, error) {
	// Set stream to true
	streamReq := *request
	streamReq.Stream = true

	// Prepare OpenAI request
	openaiReq := a.prepareOpenAIRequest(&streamReq)

	// Make streaming HTTP request
	resp, err := a.makeRequest(ctx, "/v1/chat/completions", openaiReq)
	if err != nil {
		a.updateStatus(false, err)
		return nil, err
	}

	// Create channels for streaming
	events := make(chan StreamEvent, 100)
	errors := make(chan error, 1)

	// Start streaming goroutine
	go a.handleStreamResponse(resp.Body, events, errors)

	return &ChatStreamResponse{
		Stream: resp.Body,
		Events: events,
		Errors: errors,
	}, nil
}

// GetModels returns available models for this agent
func (a *OpenAIAgent) GetModels(ctx context.Context) ([]Model, error) {
	// Make request to models endpoint
	resp, err := a.makeRequest(ctx, "/v1/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var modelsResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	// Convert to standard models
	models := make([]Model, len(modelsResp.Data))
	for i, model := range modelsResp.Data {
		models[i] = Model{
			ID:           model.ID,
			Name:         model.ID,
			Description:  fmt.Sprintf("OpenAI model: %s", model.ID),
			Created:      model.Created,
			OwnedBy:      model.OwnedBy,
			Capabilities: a.GetCapabilities(),
		}
	}

	return models, nil
}

// ValidateConfig validates the agent configuration
func (a *OpenAIAgent) ValidateConfig() error {
	return validateOpenAIConfig(a.config)
}

// GetStatus returns the current status of the agent
func (a *OpenAIAgent) GetStatus(ctx context.Context) (*AgentStatus, error) {
	// Check if agent is closed (read-only check first)
	a.statusMu.RLock()
	isClosed := a.httpClient == nil
	a.statusMu.RUnlock()

	if isClosed {
		a.statusMu.Lock()
		a.status.Health = false
		a.status.Status = "inactive"
		a.status.LastChecked = time.Now()
		statusCopy := *a.status
		a.statusMu.Unlock()
		return &statusCopy, nil
	}

	// Perform health check without holding the lock
	healthErr := a.healthCheck(ctx)

	// Update status based on health check result
	a.statusMu.Lock()
	defer a.statusMu.Unlock()

	if healthErr != nil {
		a.status.Health = false
		a.status.Status = "error"
		a.status.Details = map[string]interface{}{
			"error": healthErr.Error(),
		}
	} else {
		a.status.Health = true
		a.status.Status = "active"
	}

	a.status.LastChecked = time.Now()
	statusCopy := *a.status
	return &statusCopy, nil
}

// Close cleans up resources used by the agent
func (a *OpenAIAgent) Close() error {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()

	// Close HTTP client if needed
	if a.httpClient != nil {
		// HTTP client doesn't need explicit closing
		a.httpClient = nil
	}

	a.status.Status = "inactive"
	return nil
}

// prepareOpenAIRequest converts a ChatRequest to OpenAI format
func (a *OpenAIAgent) prepareOpenAIRequest(request *ChatRequest) map[string]interface{} {
	req := map[string]interface{}{
		"messages": request.Messages,
		"model":    a.getModel(request.Model),
		"stream":   request.Stream,
	}

	// Add optional parameters
	if request.Temperature != nil {
		req["temperature"] = *request.Temperature
	} else {
		req["temperature"] = a.config.Temperature
	}

	if request.MaxTokens != nil {
		req["max_tokens"] = *request.MaxTokens
	} else if a.config.MaxTokens > 0 {
		req["max_tokens"] = a.config.MaxTokens
	}

	// Add functions if provided
	if len(request.Functions) > 0 {
		req["functions"] = request.Functions
	}

	// Add tools if provided
	if len(request.Tools) > 0 {
		req["tools"] = request.Tools
	}

	// Add user ID if provided
	if request.UserID != "" {
		req["user"] = request.UserID
	}

	return req
}

// getModel returns the model to use, with fallback to default
func (a *OpenAIAgent) getModel(model string) string {
	if model != "" {
		return model
	}
	return a.config.DefaultModel
}

// makeRequest makes an HTTP request to the OpenAI API
func (a *OpenAIAgent) makeRequest(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	// Get httpClient safely
	a.statusMu.RLock()
	client := a.httpClient
	a.statusMu.RUnlock()

	// Check if agent is closed
	if client == nil {
		return nil, fmt.Errorf("agent is closed")
	}

	url := strings.TrimSuffix(a.config.BaseURL, "/") + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	// Add organization header if provided
	if a.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", a.config.Organization)
	}

	// Add custom headers
	for key, value := range a.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Make request
	startTime := time.Now()
	resp, err := client.Do(req)
	responseTime := time.Since(startTime).Milliseconds()

	// Update response time in status (thread-safe)
	a.statusMu.Lock()
	a.status.ResponseTime = responseTime
	a.statusMu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var errorResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, &AgentError{
				Code:    errorResp.Error.Code,
				Message: errorResp.Error.Message,
				Type:    errorResp.Error.Type,
			}
		}

		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return resp, nil
}

// handleStreamResponse handles streaming response
func (a *OpenAIAgent) handleStreamResponse(body io.ReadCloser, events chan<- StreamEvent, errors chan<- error) {
	defer close(events)
	defer close(errors)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Check for data prefix
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				events <- StreamEvent{
					Type:         "done",
					FinishReason: stringPtr("stop"),
				}
				return
			}

			// Parse JSON data
			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errors <- fmt.Errorf("failed to parse stream chunk: %w", err)
				return
			}

			// Convert to standard event
			event := a.convertStreamChunk(&chunk)
			if event != nil {
				events <- *event
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errors <- fmt.Errorf("error reading stream: %w", err)
	}
}

// convertToStandardResponse converts OpenAI response to standard format
func (a *OpenAIAgent) convertToStandardResponse(resp *openAIResponse) *ChatResponse {
	return &ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: resp.Choices,
		Usage:   resp.Usage,
	}
}

// convertStreamChunk converts OpenAI stream chunk to standard event
func (a *OpenAIAgent) convertStreamChunk(chunk *openAIStreamChunk) *StreamEvent {
	if len(chunk.Choices) == 0 {
		return nil
	}

	choice := chunk.Choices[0]
	event := &StreamEvent{
		Type: "content",
		Delta: &Delta{
			Role:    choice.Delta.Role,
			Content: choice.Delta.Content,
		},
	}

	if choice.FinishReason != nil {
		event.FinishReason = choice.FinishReason
		event.Type = "finish"
	}

	return event
}

// healthCheck performs a health check on the agent
func (a *OpenAIAgent) healthCheck(ctx context.Context) error {
	// Simple health check by calling the models endpoint
	resp, err := a.makeRequest(ctx, "/v1/models", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// updateStatus updates the agent status based on operation result
func (a *OpenAIAgent) updateStatus(success bool, err error) {
	a.statusMu.Lock()
	defer a.statusMu.Unlock()

	a.status.RequestCount++
	if success {
		a.status.Health = true
		a.status.Status = "online"
	} else {
		a.status.ErrorCount++
		a.status.Health = false
		a.status.Status = "error"
		if err != nil {
			a.status.Details = map[string]interface{}{
				"last_error": err.Error(),
			}
		}
	}

	// Calculate success rate
	if a.status.RequestCount > 0 {
		a.status.SuccessRate = float64(a.status.RequestCount-a.status.ErrorCount) / float64(a.status.RequestCount) * 100
	}

	a.status.LastChecked = time.Now()
}

// OpenAI API response structures
type openAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type openAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Delta        Delta   `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
