package dataflow

import (
	"agent-connector/internal"
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
	authService          *DataFlowAuthService
	systemConfigService  *internal.SystemConfigService
	userRateLimitService *internal.UserRateLimitService
}

// NewDataFlowAPIHandler create data flow API handler
func NewDataFlowAPIHandler() *DataFlowAPIHandler {
	return &DataFlowAPIHandler{
		authService:          NewDataFlowAuthService(),
		systemConfigService:  &internal.SystemConfigService{},
		userRateLimitService: &internal.UserRateLimitService{},
	}
}

// HandleChat handle chat request
func (h *DataFlowAPIHandler) HandleChat(c *gin.Context) {
	ctx := c.Request.Context()

	// authenticate request
	authInfo, err := h.authenticateRequest(c)
	if err != nil {
		h.respondWithError(c, http.StatusUnauthorized, "authentication_failed", err.Error())
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

	// check rate limit
	rateLimitInfo, err := h.checkRateLimit(ctx, authInfo, &req)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "rate_limit_error", err.Error())
		return
	}

	if !rateLimitInfo.Allowed {
		h.respondWithRateLimit(c, rateLimitInfo)
		return
	}

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

// authenticateRequest authenticate request
func (h *DataFlowAPIHandler) authenticateRequest(c *gin.Context) (*AuthInfo, error) {
	// get AgentID from URL parameters or JSON body
	agentID := c.Param("agent_id")
	if agentID == "" {
		agentID = c.Query("agent_id")
	}

	// get API Key from header
	apiKey := c.GetHeader("Authorization")
	if apiKey == "" {
		apiKey = c.GetHeader("X-API-Key")
	}

	return h.authService.AuthenticateRequest(agentID, apiKey)
}

// checkRateLimit check rate limit
func (h *DataFlowAPIHandler) checkRateLimit(ctx context.Context, authInfo *AuthInfo, req *DataFlowRequest) (*RateLimitInfo, error) {
	// get system config
	systemConfig, err := h.systemConfigService.GetSystemConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}

	// get user ID
	userID := h.authService.GetUserIDFromAPIKey(authInfo.APIKey)
	authInfo.UserID = userID

	// get user rate limit config
	userRateLimit, err := h.userRateLimitService.GetUserRateLimit(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rate limit: %w", err)
	}

	rateLimitInfo := &RateLimitInfo{
		Mode:     string(systemConfig.RateLimitMode),
		AgentQPS: authInfo.Agent.QPS,
		Allowed:  true,
	}

	// set user rate limit parameters
	if userRateLimit != nil && userRateLimit.Enabled {
		if userRateLimit.Priority != nil {
			rateLimitInfo.UserPriority = *userRateLimit.Priority
		}
		if userRateLimit.QPS != nil {
			rateLimitInfo.UserQPS = *userRateLimit.QPS
		}
	} else {
		// use default config
		rateLimitInfo.UserPriority = systemConfig.DefaultPriority
		rateLimitInfo.UserQPS = systemConfig.DefaultQPS
	}

	// simplified rate limit check (in production environment, a complete rate limit service should be used)
	return rateLimitInfo, nil
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
		reqBody = map[string]interface{}{
			"query":           req.Query,
			"conversation_id": req.ConversationID,
			"user":            req.User,
			"inputs":          req.Inputs,
			"response_mode":   req.ResponseMode,
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

	decoder := json.NewDecoder(stream)
	for {
		var chunk interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			h.writeSSEError(c, "decode_error", err.Error())
			return
		}

		// write SSE data
		h.writeSSEData(c, chunk)
		flusher.Flush()
	}

	// send done signal
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

// respondWithRateLimit return rate limit response
func (h *DataFlowAPIHandler) respondWithRateLimit(c *gin.Context, rateLimitInfo *RateLimitInfo) {
	response := DataFlowResponse{
		Code:    http.StatusTooManyRequests,
		Message: "Rate limit exceeded",
		Error: &APIError{
			Type:    "rate_limit_exceeded",
			Code:    "429",
			Message: fmt.Sprintf("Rate limit exceeded. Wait time: %v", rateLimitInfo.WaitTime),
		},
	}

	// set Rate Limit headers
	c.Header("X-RateLimit-Mode", rateLimitInfo.Mode)
	c.Header("X-RateLimit-User-QPS", strconv.Itoa(rateLimitInfo.UserQPS))
	c.Header("X-RateLimit-Agent-QPS", strconv.Itoa(rateLimitInfo.AgentQPS))
	if rateLimitInfo.WaitTime > 0 {
		c.Header("Retry-After", strconv.Itoa(int(rateLimitInfo.WaitTime.Seconds())))
	}

	c.JSON(http.StatusTooManyRequests, response)
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
