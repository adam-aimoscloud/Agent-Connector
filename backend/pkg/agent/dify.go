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

// DifyAgent implements the Agent interface for Dify platform
type DifyAgent struct {
	config     *DifyConfig
	httpClient *http.Client
	status     *AgentStatus
	statusMu   sync.RWMutex // Mutex to protect status field
}

// DifyConfig represents configuration for Dify agents
type DifyConfig struct {
	AgentConfig

	// BaseURL is the base URL for the Dify API
	BaseURL string `json:"base_url"`

	// APIKey is the API key for authentication
	APIKey string `json:"api_key"`

	// AppID is the Dify app identifier
	AppID string `json:"app_id"`

	// AppType is the type of Dify app (chatbot, agent, workflow)
	AppType string `json:"app_type"`

	// Version is the API version
	Version string `json:"version"`

	// EnableLogging enables conversation logging
	EnableLogging bool `json:"enable_logging"`

	// AutoGenerateTitle enables auto title generation
	AutoGenerateTitle bool `json:"auto_generate_title"`

	// CustomHeaders for additional HTTP headers
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

// NewDifyAgent creates a new Dify agent
func NewDifyAgent(config *DifyConfig) (*DifyAgent, error) {
	if err := validateDifyConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set defaults
	setDifyDefaults(config)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	agent := &DifyAgent{
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

// validateDifyConfig validates the Dify configuration
func validateDifyConfig(config *DifyConfig) error {
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

	if config.AppID == "" {
		return fmt.Errorf("app ID is required")
	}

	if !config.Type.IsValid() {
		return fmt.Errorf("invalid agent type: %s", config.Type)
	}

	return nil
}

// setDifyDefaults sets default values for Dify configuration
func setDifyDefaults(config *DifyConfig) {
	if config.Name == "" {
		config.Name = "Dify Agent"
	}

	if config.Type == "" {
		config.Type = AgentTypeDify
	}

	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = DefaultMaxConcurrentRequests
	}

	if config.AppType == "" {
		config.AppType = "chatbot"
	}

	if config.Version == "" {
		config.Version = "v1"
	}
}

// GetID returns the unique identifier of the agent
func (d *DifyAgent) GetID() string {
	return d.config.ID
}

// GetName returns the display name of the agent
func (d *DifyAgent) GetName() string {
	return d.config.Name
}

// GetType returns the type of the agent source
func (d *DifyAgent) GetType() AgentType {
	return AgentTypeDify
}

// GetCapabilities returns the capabilities of the agent
func (d *DifyAgent) GetCapabilities() AgentCapabilities {
	capabilities := AgentCapabilities{
		SupportsChatCompletion:  true,
		SupportsStreaming:       true,
		SupportsImages:          false,
		SupportsFiles:           true,
		SupportsFunctionCalling: false,
		MaxTokens:               4096,
		SupportedLanguages:      []string{"en", "zh", "es", "fr", "de", "ja", "ko"},
	}

	// Adjust capabilities based on app type
	if d.config.AppType == "agent" {
		capabilities.SupportsFunctionCalling = true
	}

	return capabilities
}

// Chat sends a chat message and returns the response
func (d *DifyAgent) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	// Prepare Dify request
	difyReq := d.prepareDifyRequest(request)

	// Make HTTP request
	resp, err := d.makeRequest(ctx, "/chat-messages", difyReq)
	if err != nil {
		d.updateStatus(false, err)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var difyResp difyResponse
	if err := json.NewDecoder(resp.Body).Decode(&difyResp); err != nil {
		d.updateStatus(false, fmt.Errorf("failed to decode response: %w", err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API errors
	if difyResp.Code != "" && difyResp.Code != "success" {
		agentErr := &AgentError{
			Code:    difyResp.Code,
			Message: difyResp.Message,
			Type:    "dify_error",
		}
		d.updateStatus(false, agentErr)
		return nil, agentErr
	}

	// Convert to standard response
	response := d.convertToStandardResponse(&difyResp)
	d.updateStatus(true, nil)

	return response, nil
}

// ChatStream sends a chat message and returns a streaming response
func (d *DifyAgent) ChatStream(ctx context.Context, request *ChatRequest) (*ChatStreamResponse, error) {
	// Prepare Dify streaming request
	difyReq := d.prepareDifyRequest(request)
	difyReq["response_mode"] = "streaming"

	// Make streaming HTTP request
	resp, err := d.makeRequest(ctx, "/chat-messages", difyReq)
	if err != nil {
		d.updateStatus(false, err)
		return nil, err
	}

	// Create channels for streaming
	events := make(chan StreamEvent, 100)
	errors := make(chan error, 1)

	// Start streaming goroutine
	go d.handleStreamResponse(resp.Body, events, errors)

	return &ChatStreamResponse{
		Stream: resp.Body,
		Events: events,
		Errors: errors,
	}, nil
}

// GetModels returns available models for this agent
func (d *DifyAgent) GetModels(ctx context.Context) ([]Model, error) {
	// Dify doesn't have a models endpoint, return app info as model
	model := Model{
		ID:           d.config.AppID,
		Name:         d.config.Name,
		Description:  fmt.Sprintf("Dify %s: %s", d.config.AppType, d.config.Name),
		Created:      time.Now().Unix(),
		OwnedBy:      "dify",
		Capabilities: d.GetCapabilities(),
	}

	return []Model{model}, nil
}

// ValidateConfig validates the agent configuration
func (d *DifyAgent) ValidateConfig() error {
	return validateDifyConfig(d.config)
}

// GetStatus returns the current status of the agent
func (d *DifyAgent) GetStatus(ctx context.Context) (*AgentStatus, error) {
	// Check if agent is closed (read-only check first)
	d.statusMu.RLock()
	isClosed := d.httpClient == nil
	d.statusMu.RUnlock()

	if isClosed {
		d.statusMu.Lock()
		d.status.Health = false
		d.status.Status = "offline"
		d.status.LastChecked = time.Now()
		statusCopy := *d.status
		d.statusMu.Unlock()
		return &statusCopy, nil
	}

	// Perform health check without holding the lock
	healthErr := d.healthCheck(ctx)

	// Update status based on health check result
	d.statusMu.Lock()
	defer d.statusMu.Unlock()

	if healthErr != nil {
		d.status.Health = false
		d.status.Status = "error"
		d.status.Details = map[string]interface{}{
			"error": healthErr.Error(),
		}
	} else {
		d.status.Health = true
		d.status.Status = "active"
	}

	d.status.LastChecked = time.Now()
	statusCopy := *d.status
	return &statusCopy, nil
}

// Close cleans up resources used by the agent
func (d *DifyAgent) Close() error {
	d.statusMu.Lock()
	defer d.statusMu.Unlock()

	// Close HTTP client if needed
	if d.httpClient != nil {
		// HTTP client doesn't need explicit closing
		d.httpClient = nil
	}

	d.status.Status = "offline"
	return nil
}

// prepareDifyRequest converts a ChatRequest to Dify format
func (d *DifyAgent) prepareDifyRequest(request *ChatRequest) map[string]interface{} {
	req := map[string]interface{}{
		"inputs":             map[string]interface{}{},
		"response_mode":      "blocking",
		"user":               d.getUserID(request.UserID),
		"auto_generate_name": d.config.AutoGenerateTitle,
	}

	// Extract the latest user message as query
	var query string
	for i := len(request.Messages) - 1; i >= 0; i-- {
		if request.Messages[i].Role == "user" {
			query = request.Messages[i].Content
			break
		}
	}
	req["query"] = query

	// Add conversation history if available
	if len(request.Messages) > 1 {
		var conversationHistory []map[string]interface{}
		for _, msg := range request.Messages[:len(request.Messages)-1] {
			if msg.Role == "user" || msg.Role == "assistant" {
				conversationHistory = append(conversationHistory, map[string]interface{}{
					"role":    msg.Role,
					"content": msg.Content,
				})
			}
		}
		req["conversation_history"] = conversationHistory
	}

	// Add session ID if provided
	if request.SessionID != "" {
		req["conversation_id"] = request.SessionID
	}

	// Add metadata as inputs
	if request.Metadata != nil {
		for key, value := range request.Metadata {
			req["inputs"].(map[string]interface{})[key] = value
		}
	}

	return req
}

// getUserID returns the user ID, with fallback to anonymous
func (d *DifyAgent) getUserID(userID string) string {
	if userID != "" {
		return userID
	}
	return "anonymous"
}

// makeRequest makes an HTTP request to the Dify API
func (d *DifyAgent) makeRequest(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	// Get httpClient safely
	d.statusMu.RLock()
	client := d.httpClient
	d.statusMu.RUnlock()

	// Check if agent is closed
	if client == nil {
		return nil, fmt.Errorf("agent is closed")
	}

	url := strings.TrimSuffix(d.config.BaseURL, "/") + "/" + d.config.Version + endpoint

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
	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)

	// Add custom headers
	for key, value := range d.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Make request
	startTime := time.Now()
	resp, err := client.Do(req)
	responseTime := time.Since(startTime).Milliseconds()

	// Update response time in status (thread-safe)
	d.statusMu.Lock()
	d.status.ResponseTime = responseTime
	d.statusMu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var errorResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, &AgentError{
				Code:    errorResp.Code,
				Message: errorResp.Message,
				Type:    "dify_error",
			}
		}

		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return resp, nil
}

// handleStreamResponse handles streaming response
func (d *DifyAgent) handleStreamResponse(body io.ReadCloser, events chan<- StreamEvent, errors chan<- error) {
	defer close(events)
	defer close(errors)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
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
			var streamData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &streamData); err != nil {
				errors <- fmt.Errorf("failed to parse stream data: %w", err)
				return
			}

			// Convert to standard event
			event := d.convertStreamData(streamData)
			if event != nil {
				events <- *event
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errors <- fmt.Errorf("error reading stream: %w", err)
	}
}

// convertToStandardResponse converts Dify response to standard format
func (d *DifyAgent) convertToStandardResponse(resp *difyResponse) *ChatResponse {
	response := &ChatResponse{
		ID:      resp.MessageID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   d.config.AppID,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: resp.Answer,
				},
				FinishReason: stringPtr("stop"),
			},
		},
	}

	// Add usage information if available
	if resp.Metadata != nil {
		if tokens, ok := resp.Metadata["usage"].(map[string]interface{}); ok {
			usage := &Usage{}
			if promptTokens, ok := tokens["prompt_tokens"].(float64); ok {
				usage.PromptTokens = int(promptTokens)
			}
			if completionTokens, ok := tokens["completion_tokens"].(float64); ok {
				usage.CompletionTokens = int(completionTokens)
			}
			usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
			response.Usage = usage
		}
	}

	return response
}

// convertStreamData converts Dify stream data to standard event
func (d *DifyAgent) convertStreamData(data map[string]interface{}) *StreamEvent {
	eventType, ok := data["event"].(string)
	if !ok {
		return nil
	}

	switch eventType {
	case "message", "agent_message":
		if answer, ok := data["answer"].(string); ok {
			return &StreamEvent{
				Type: "content",
				Delta: &Delta{
					Role:    "assistant",
					Content: answer,
				},
			}
		}
	case "message_end":
		return &StreamEvent{
			Type:         "finish",
			FinishReason: stringPtr("stop"),
		}
	case "error":
		return &StreamEvent{
			Type: "error",
			Data: data,
		}
	}

	return nil
}

// healthCheck performs a health check on the agent
func (d *DifyAgent) healthCheck(ctx context.Context) error {
	// Simple health check by making a parameters request
	req := map[string]interface{}{
		"user": "health-check",
	}

	resp, err := d.makeRequest(ctx, "/parameters", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// updateStatus updates the agent status based on operation result
func (d *DifyAgent) updateStatus(success bool, err error) {
	d.statusMu.Lock()
	defer d.statusMu.Unlock()

	d.status.RequestCount++
	if success {
		d.status.Health = true
		d.status.Status = "online"
	} else {
		d.status.ErrorCount++
		d.status.Health = false
		d.status.Status = "error"
		if err != nil {
			d.status.Details = map[string]interface{}{
				"last_error": err.Error(),
			}
		}
	}

	// Calculate success rate
	if d.status.RequestCount > 0 {
		d.status.SuccessRate = float64(d.status.RequestCount-d.status.ErrorCount) / float64(d.status.RequestCount) * 100
	}

	d.status.LastChecked = time.Now()
}

// Dify API response structures
type difyResponse struct {
	MessageID      string                 `json:"message_id"`
	ConversationID string                 `json:"conversation_id"`
	Mode           string                 `json:"mode"`
	Answer         string                 `json:"answer"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      int64                  `json:"created_at"`
	Code           string                 `json:"code,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Status         string                 `json:"status,omitempty"`
}
