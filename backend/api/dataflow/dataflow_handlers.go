package dataflow

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DataFlowAPIHandler data flow API handler
type DataFlowAPIHandler struct {
	authService *DataFlowAuthService
}

// NewDataFlowAPIHandler create data flow API handler
func NewDataFlowAPIHandler() *DataFlowAPIHandler {
	return &DataFlowAPIHandler{
		authService: NewDataFlowAuthService(),
	}
}

// HandleChat handle chat request
// Note: This handler expects authInfo to be set in context by AuthenticationMiddleware
// and rate limiting to be handled by RateLimitMiddleware
func (h *DataFlowAPIHandler) HandleChat(c *gin.Context) {
	// get auth info from context (set by middleware)
	authInfo, err := GetAuthInfoFromContext(c)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// parse request
	var req DataFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Invalid request format: "+err.Error())
		return
	}

	// set authentication information
	req.AgentID = authInfo.AgentID
	req.APIKey = authInfo.APIKey

	// determine request format
	requestFormat := h.authService.DetermineRequestFormat(&req)

	// check agent compatibility
	if !h.authService.IsAgentTypeCompatible(authInfo.Agent.Type, requestFormat) {
		h.respondWithError(c, http.StatusBadRequest, "incompatible_agent",
			fmt.Sprintf("Agent type %s is not compatible with request format %s", authInfo.Agent.Type, requestFormat))
		return
	}

	// handle request based on request type
	if req.Stream || req.ResponseMode == "streaming" {
		h.handleStreamingRequest(c, authInfo, &req, requestFormat)
	} else {
		h.handleBlockingRequest(c, authInfo, &req, requestFormat)
	}
}

// handleStreamingRequest handle streaming request
func (h *DataFlowAPIHandler) handleStreamingRequest(c *gin.Context, authInfo *AuthInfo, req *DataFlowRequest, requestFormat string) {
	// set SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// forward request to actual agent
	responseStream, err := h.forwardStreamingRequest(ctx, authInfo, req, requestFormat)
	if err != nil {
		h.writeSSEError(c, "forward_error", err.Error())
		return
	}
	defer responseStream.Close()

	// stream forward response
	h.streamResponse(c, responseStream, authInfo.Agent.ResponseFormat)
}

// handleBlockingRequest handle blocking request
func (h *DataFlowAPIHandler) handleBlockingRequest(c *gin.Context, authInfo *AuthInfo, req *DataFlowRequest, requestFormat string) {
	// create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// forward request to actual agent
	response, err := h.forwardBlockingRequest(ctx, authInfo, req, requestFormat)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "forward_error", err.Error())
		return
	}

	// return data based on response format
	if authInfo.Agent.ResponseFormat == "dify" {
		c.JSON(http.StatusOK, response)
	} else {
		// OpenAI format
		c.JSON(http.StatusOK, response)
	}
}

// forwardStreamingRequest forward streaming request
func (h *DataFlowAPIHandler) forwardStreamingRequest(ctx context.Context, authInfo *AuthInfo, req *DataFlowRequest, requestFormat string) (io.ReadCloser, error) {
	// build forward request
	forwardReq, err := h.buildForwardRequest(ctx, authInfo, req, requestFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to build forward request: %w", err)
	}

	// send request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(forwardReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// forwardBlockingRequest forward blocking request
func (h *DataFlowAPIHandler) forwardBlockingRequest(ctx context.Context, authInfo *AuthInfo, req *DataFlowRequest, requestFormat string) (interface{}, error) {
	// build forward request
	forwardReq, err := h.buildForwardRequest(ctx, authInfo, req, requestFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to build forward request: %w", err)
	}

	// send request
	client := &http.Client{
		Timeout: 2 * time.Minute,
	}

	resp, err := client.Do(forwardReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("agent returned error status: %d", resp.StatusCode)
	}

	// parse response
	var response interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// buildForwardRequest build forward request
func (h *DataFlowAPIHandler) buildForwardRequest(ctx context.Context, authInfo *AuthInfo, req *DataFlowRequest, requestFormat string) (*http.Request, error) {
	// build different requests based on agent type
	var reqBody interface{}
	var endpoint string

	switch authInfo.Agent.Type {
	case "openai", "openai_compatible":
		reqBody = map[string]interface{}{
			"model":       req.Model,
			"messages":    req.Messages,
			"max_tokens":  req.MaxTokens,
			"temperature": req.Temperature,
			"stream":      req.Stream,
		}
		endpoint = "/v1/chat/completions"

	case "dify":
		// Ensure inputs is not nil
		inputs := req.Inputs
		if inputs == nil {
			inputs = map[string]interface{}{}
		}

		// Set default response_mode if not provided
		responseMode := req.ResponseMode
		if responseMode == "" {
			if req.Stream {
				responseMode = "streaming"
			} else {
				responseMode = "blocking"
			}
		}

		reqBody = map[string]interface{}{
			"query":           req.Query,
			"conversation_id": req.ConversationID,
			"user":            req.User,
			"inputs":          inputs,
			"response_mode":   responseMode,
		}

		endpoint = "/v1/chat-messages"

	default:
		return nil, fmt.Errorf("unsupported agent type: %s", authInfo.Agent.Type)
	}

	// serialize request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// build full URL
	fullURL := strings.TrimSuffix(authInfo.Agent.URL, "/") + endpoint

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// set request headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+authInfo.Agent.SourceAPIKey)

	// if Dify, set special API Key header
	if authInfo.Agent.Type == "dify" {
		httpReq.Header.Set("Authorization", "Bearer "+authInfo.Agent.SourceAPIKey)
	}

	return httpReq, nil
}

// streamResponse stream response
func (h *DataFlowAPIHandler) streamResponse(c *gin.Context, stream io.ReadCloser, responseFormat string) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		h.writeSSEError(c, "streaming_not_supported", "Streaming not supported")
		return
	}

	// Use bufio.Scanner to read line by line for SSE format
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for SSE data format
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if dataStr == "[DONE]" {
				h.writeSSEData(c, map[string]string{"event": "done"})
				flusher.Flush()
				break
			}

			// Parse JSON data
			var chunk interface{}
			if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
				h.writeSSEError(c, "decode_error", err.Error())
				return
			}

			// Write SSE data
			h.writeSSEData(c, chunk)
			flusher.Flush()
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		h.writeSSEError(c, "stream_read_error", err.Error())
		return
	}

	// Send done signal if not already sent
	h.writeSSEData(c, map[string]string{"event": "done"})
	flusher.Flush()
}

// writeSSEData write SSE data
func (h *DataFlowAPIHandler) writeSSEData(c *gin.Context, data interface{}) {
	jsonData, _ := json.Marshal(data)
	c.Writer.WriteString("data: " + string(jsonData) + "\n\n")
}

// writeSSEError write SSE error
func (h *DataFlowAPIHandler) writeSSEError(c *gin.Context, errorType, message string) {
	errorData := map[string]interface{}{
		"error": map[string]string{
			"type":    errorType,
			"message": message,
		},
	}
	h.writeSSEData(c, errorData)
}

// respondWithError return error response
func (h *DataFlowAPIHandler) respondWithError(c *gin.Context, statusCode int, errorType, message string) {
	response := DataFlowResponse{
		Code:    statusCode,
		Message: "Error",
		Error: &APIError{
			Type:    errorType,
			Code:    strconv.Itoa(statusCode),
			Message: message,
		},
	}
	c.JSON(statusCode, response)
}

// HealthCheck health check
func (h *DataFlowAPIHandler) HealthCheck(c *gin.Context) {
	response := map[string]interface{}{
		"status":    "ok",
		"service":   "dataflow-api",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}
