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

// DifyWorkflowBackend implements AgentBackend for Dify workflows/run API
type DifyWorkflowBackend struct{}

// NewDifyWorkflowBackend creates a new Dify Workflow backend
func NewDifyWorkflowBackend() *DifyWorkflowBackend {
	return &DifyWorkflowBackend{}
}

// GetType returns the backend type
func (b *DifyWorkflowBackend) GetType() types.AgentType {
	return types.AgentTypeDifyWorkflow
}

// ValidateRequest validates the request for Dify Workflow backend
func (b *DifyWorkflowBackend) ValidateRequest(req *BackendRequest) error {
	if req.User == "" {
		return fmt.Errorf("user field is required for Dify Workflow backend")
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

	// Ensure data is not nil
	if req.Data == nil {
		req.Data = map[string]interface{}{}
	}

	return nil
}

// BuildForwardRequest builds the HTTP request for Dify Workflow API
func (b *DifyWorkflowBackend) BuildForwardRequest(ctx context.Context, req *BackendRequest, agentInfo *AgentInfo) (*http.Request, error) {
	// Build Dify workflow request body
	reqBody := map[string]interface{}{
		"inputs":        req.Data,
		"user":          req.User,
		"response_mode": req.ResponseMode,
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
func (b *DifyWorkflowBackend) ProcessBlockingResponse(resp *http.Response) (interface{}, error) {
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
func (b *DifyWorkflowBackend) ProcessStreamingResponse(resp *http.Response) (io.ReadCloser, error) {
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// GetEndpoint returns the endpoint path for Dify Workflow API
func (b *DifyWorkflowBackend) GetEndpoint() string {
	return "/v1/workflows/run"
}
