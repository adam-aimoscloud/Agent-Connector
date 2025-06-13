package backends

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"agent-connector/pkg/types"
	"io"
	"net/http"
	"strings"
)

// OpenAIBackend implements AgentBackend for OpenAI compatible APIs
type OpenAIBackend struct{}

// NewOpenAIBackend creates a new OpenAI backend
func NewOpenAIBackend() *OpenAIBackend {
	return &OpenAIBackend{}
}

// GetType returns the backend type
func (b *OpenAIBackend) GetType() types.AgentType {
	return types.AgentTypeOpenAI
}

// ValidateRequest validates the request for OpenAI backend
func (b *OpenAIBackend) ValidateRequest(req *BackendRequest) error {
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages field is required for OpenAI backend")
	}

	if req.Model == "" {
		req.Model = "gpt-3.5-turbo" // Set default model
	}

	// Validate messages format
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message[%d].role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message[%d].content is required", i)
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return fmt.Errorf("message[%d].role must be one of: system, user, assistant", i)
		}
	}

	return nil
}

// BuildForwardRequest builds the HTTP request for OpenAI API
func (b *OpenAIBackend) BuildForwardRequest(ctx context.Context, req *BackendRequest, agentInfo *AgentInfo) (*http.Request, error) {
	// Build OpenAI request body
	reqBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	// Add optional fields
	if req.MaxTokens != nil {
		reqBody["max_tokens"] = *req.MaxTokens
	}
	if req.Temperature != nil {
		reqBody["temperature"] = *req.Temperature
	}

	// Serialize request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build full URL
	fullURL := strings.TrimSuffix(agentInfo.URL, "/") + b.GetEndpoint()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+agentInfo.SourceAPIKey)

	return httpReq, nil
}

// ProcessBlockingResponse processes the response for blocking requests
func (b *OpenAIBackend) ProcessBlockingResponse(resp *http.Response) (interface{}, error) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	var response interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// ProcessStreamingResponse processes the response for streaming requests
func (b *OpenAIBackend) ProcessStreamingResponse(resp *http.Response) (io.ReadCloser, error) {
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetEndpoint returns the endpoint path for OpenAI API
func (b *OpenAIBackend) GetEndpoint() string {
	return "/v1/chat/completions"
}
