package dataflow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"agent-connector/api/dataflow/backends"
	"agent-connector/pkg/ratelimiter"
)

// DataflowService handles dataflow operations with different agent backends
type DataflowService struct {
	factory     backends.BackendFactory
	rateLimiter *ratelimiter.RedisRateLimiter
	httpClient  *http.Client
	authService *DataFlowAuthService
}

// NewDataflowService creates a new dataflow service
func NewDataflowService(rateLimiter *ratelimiter.RedisRateLimiter) *DataflowService {
	return &DataflowService{
		factory:     backends.NewDefaultBackendFactory(),
		rateLimiter: rateLimiter,
		authService: NewDataFlowAuthService(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessRequest processes a dataflow request using the appropriate backend
func (s *DataflowService) ProcessRequest(ctx context.Context, req *backends.BackendRequest) (interface{}, error) {
	// Get agent information
	agentInfo, err := s.getAgentInfo(req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}

	// Check if agent is enabled
	if !agentInfo.Enabled {
		return nil, fmt.Errorf("agent %s is disabled", req.AgentID)
	}

	// Determine backend type
	backendType := backends.DetermineAgentType(agentInfo.Type)

	// Create backend instance
	backend, err := s.factory.CreateBackend(backendType)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	// Validate request for this backend
	if err := backend.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Check rate limit
	if err := s.checkRateLimit(ctx, req.AgentID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build forward request
	httpReq, err := backend.BuildForwardRequest(ctx, req, agentInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to build forward request: %w", err)
	}

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Process response based on streaming mode
	if req.Stream || req.ResponseMode == "streaming" {
		return s.processStreamingResponse(backend, resp)
	} else {
		return backend.ProcessBlockingResponse(resp)
	}
}

// ProcessStreamingRequest processes a streaming dataflow request
func (s *DataflowService) ProcessStreamingRequest(ctx context.Context, req *backends.BackendRequest, w http.ResponseWriter) error {
	// Get agent information
	agentInfo, err := s.getAgentInfo(req.AgentID)
	if err != nil {
		return fmt.Errorf("failed to get agent info: %w", err)
	}

	// Check if agent is enabled
	if !agentInfo.Enabled {
		return fmt.Errorf("agent %s is disabled", req.AgentID)
	}

	// Check if agent supports streaming
	if !agentInfo.SupportStreaming {
		return fmt.Errorf("agent %s does not support streaming", req.AgentID)
	}

	// Determine backend type
	backendType := backends.DetermineAgentType(agentInfo.Type)

	// Create backend instance
	backend, err := s.factory.CreateBackend(backendType)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	// Ensure streaming mode
	req.Stream = true
	req.ResponseMode = "streaming"

	// Validate request for this backend
	if err := backend.ValidateRequest(req); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	// Check rate limit
	if err := s.checkRateLimit(ctx, req.AgentID); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build forward request
	httpReq, err := backend.BuildForwardRequest(ctx, req, agentInfo)
	if err != nil {
		return fmt.Errorf("failed to build forward request: %w", err)
	}

	// Execute request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Process streaming response
	streamReader, err := backend.ProcessStreamingResponse(resp)
	if err != nil {
		return fmt.Errorf("failed to process streaming response: %w", err)
	}
	defer streamReader.Close()

	// Set response headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Stream response
	return s.streamResponse(streamReader, w)
}

// getAgentInfo retrieves agent information from database using existing auth service
func (s *DataflowService) getAgentInfo(agentID string) (*backends.AgentInfo, error) {
	// Use existing auth service to authenticate and get agent info
	// We need to pass a dummy API key since we're just getting agent info
	// In a real scenario, this should be refactored to have a separate method
	authInfo, err := s.authService.AuthenticateRequest(agentID, "dummy_key")
	if err != nil {
		// If authentication fails, try to get agent directly
		agent, err := s.authService.agentService.GetAgentByAgentID(agentID)
		if err != nil {
			return nil, fmt.Errorf("agent not found: %w", err)
		}

		return &backends.AgentInfo{
			ID:               agent.ID,
			Name:             agent.Name,
			Type:             string(agent.Type),
			URL:              agent.URL,
			SourceAPIKey:     agent.SourceAPIKey,
			QPS:              agent.QPS,
			Enabled:          agent.Enabled,
			SupportStreaming: agent.SupportStreaming,
			ResponseFormat:   agent.ResponseFormat,
		}, nil
	}

	return &backends.AgentInfo{
		ID:               authInfo.Agent.ID,
		Name:             authInfo.Agent.Name,
		Type:             authInfo.Agent.Type,
		URL:              authInfo.Agent.URL,
		SourceAPIKey:     authInfo.Agent.SourceAPIKey,
		QPS:              authInfo.Agent.QPS,
		Enabled:          authInfo.Agent.Enabled,
		SupportStreaming: authInfo.Agent.SupportStreaming,
		ResponseFormat:   authInfo.Agent.ResponseFormat,
	}, nil
}

// checkRateLimit checks if the request is within rate limits
func (s *DataflowService) checkRateLimit(ctx context.Context, agentID string) error {
	if s.rateLimiter == nil {
		return nil // No rate limiting configured
	}

	allowed, err := s.rateLimiter.Allow(ctx, agentID)
	if err != nil {
		log.Printf("Rate limiter error: %v", err)
		return nil // Allow request if rate limiter fails
	}

	if !allowed {
		return fmt.Errorf("rate limit exceeded for agent %s", agentID)
	}

	return nil
}

// processStreamingResponse processes streaming response for non-HTTP streaming
func (s *DataflowService) processStreamingResponse(backend backends.AgentBackend, resp *http.Response) (io.ReadCloser, error) {
	streamReader, err := backend.ProcessStreamingResponse(resp)
	if err != nil {
		return nil, err
	}
	return streamReader, nil
}

// streamResponse streams the response to the client
func (s *DataflowService) streamResponse(reader io.ReadCloser, w http.ResponseWriter) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle SSE format
		if strings.HasPrefix(line, "data: ") {
			dataContent := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if strings.TrimSpace(dataContent) == "[DONE]" {
				break
			}

			// Try to parse as JSON to validate
			var jsonData interface{}
			if err := json.Unmarshal([]byte(dataContent), &jsonData); err != nil {
				log.Printf("Invalid JSON in stream: %s", dataContent)
				continue
			}

			// Write the line as-is
			if _, err := fmt.Fprintf(w, "%s\n", line); err != nil {
				return fmt.Errorf("failed to write response: %w", err)
			}
		} else {
			// For non-SSE format, assume it's JSON data
			var jsonData interface{}
			if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
				log.Printf("Invalid JSON in stream: %s", line)
				continue
			}

			// Write in SSE format
			if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
				return fmt.Errorf("failed to write response: %w", err)
			}
		}

		flusher.Flush()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
