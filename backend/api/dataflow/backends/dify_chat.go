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

// DifyChatBackend implements AgentBackend for Dify chat-messages API
type DifyChatBackend struct{}

// NewDifyChatBackend creates a new Dify Chat backend
func NewDifyChatBackend() *DifyChatBackend {
	return &DifyChatBackend{}
}

// GetType returns the backend type
func (b *DifyChatBackend) GetType() types.AgentType {
	return types.AgentTypeDifyChat
}

// ValidateRequest validates the request for Dify Chat backend
func (b *DifyChatBackend) ValidateRequest(req *BackendRequest) error {
	if req.Query == "" {
		return fmt.Errorf("query field is required for Dify Chat backend")
	}

	if req.User == "" {
		return fmt.Errorf("user field is required for Dify Chat backend")
	}

	// Set default response mode if not provided
	if req.ResponseMode == "" {
		if req.Stream {
			req.ResponseMode = "streaming"
		} else {
			req.ResponseMode = "blocking"
		}
	}

	// Validate response mode
	if req.ResponseMode != "blocking" && req.ResponseMode != "streaming" {
		return fmt.Errorf("response_mode must be either 'blocking' or 'streaming'")
	}

	// Ensure inputs is not nil
	if req.Inputs == nil {
		req.Inputs = map[string]interface{}{}
	}

	return nil
}

// BuildForwardRequest builds the HTTP request for Dify Chat API
func (b *DifyChatBackend) BuildForwardRequest(ctx context.Context, req *BackendRequest, agentInfo *AgentInfo) (*http.Request, error) {
	// Build Dify request body
	reqBody := map[string]interface{}{
		"query":           req.Query,
		"conversation_id": req.ConversationID,
		"user":            req.User,
		"inputs":          req.Inputs,
		"response_mode":   req.ResponseMode,
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
func (b *DifyChatBackend) ProcessBlockingResponse(resp *http.Response) (interface{}, error) {
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
func (b *DifyChatBackend) ProcessStreamingResponse(resp *http.Response) (io.ReadCloser, error) {
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetEndpoint returns the endpoint path for Dify Chat API
func (b *DifyChatBackend) GetEndpoint() string {
	return "/v1/chat-messages"
}
