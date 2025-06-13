package dataflow

import (
	"encoding/json"
	"net/http"

	"agent-connector/api/dataflow/backends"
	"agent-connector/pkg/ratelimiter"

	"github.com/gin-gonic/gin"
)

// DataFlowAPIHandler new data flow API handler using backend architecture
type DataFlowAPIHandler struct {
	service *DataflowService
}

// NewDataFlowAPIHandler create new data flow API handler
func NewDataFlowAPIHandler(rateLimiter *ratelimiter.RedisRateLimiter) *DataFlowAPIHandler {
	return &DataFlowAPIHandler{
		service: NewDataflowService(rateLimiter),
	}
}

// HandleOpenAIChat handle OpenAI compatible chat request
func (h *DataFlowAPIHandler) HandleOpenAIChat(c *gin.Context) {
	// Get auth info from context (set by middleware)
	authInfo, err := GetAuthInfoFromContext(c)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Parse OpenAI request
	var req struct {
		AgentID  string `json:"agent_id,omitempty"`
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		MaxTokens   *int     `json:"max_tokens,omitempty"`
		Temperature *float64 `json:"temperature,omitempty"`
		Stream      bool     `json:"stream,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Invalid request format: "+err.Error())
		return
	}

	// Use agent_id from request body if provided, otherwise from auth info
	agentID := req.AgentID
	if agentID == "" {
		agentID = authInfo.AgentID
	}

	// Convert messages
	var backendMessages []backends.ChatMessage
	for _, msg := range req.Messages {
		backendMessages = append(backendMessages, backends.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Convert to backend request
	backendReq := &backends.BackendRequest{
		AgentID:     agentID,
		APIKey:      authInfo.APIKey,
		Model:       req.Model,
		Messages:    backendMessages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
	}

	// Process request
	if req.Stream {
		h.handleStreamingRequest(c, backendReq)
	} else {
		h.handleBlockingRequest(c, backendReq)
	}
}

// HandleDifyChat handle Dify chat request
func (h *DataFlowAPIHandler) HandleDifyChat(c *gin.Context) {
	// Get auth info from context (set by middleware)
	authInfo, err := GetAuthInfoFromContext(c)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Parse Dify request
	var req struct {
		AgentID        string                 `json:"agent_id,omitempty"`
		Query          string                 `json:"query"`
		ConversationID string                 `json:"conversation_id,omitempty"`
		User           string                 `json:"user"`
		Inputs         map[string]interface{} `json:"inputs,omitempty"`
		ResponseMode   string                 `json:"response_mode,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Invalid request format: "+err.Error())
		return
	}

	// Use agent_id from request body if provided, otherwise from auth info
	agentID := req.AgentID
	if agentID == "" {
		agentID = authInfo.AgentID
	}

	// Convert to backend request
	backendReq := &backends.BackendRequest{
		AgentID:        agentID,
		APIKey:         authInfo.APIKey,
		Query:          req.Query,
		ConversationID: req.ConversationID,
		User:           req.User,
		Inputs:         req.Inputs,
		ResponseMode:   req.ResponseMode,
		Stream:         req.ResponseMode == "streaming",
	}

	// Process request
	if req.ResponseMode == "streaming" {
		h.handleStreamingRequest(c, backendReq)
	} else {
		h.handleBlockingRequest(c, backendReq)
	}
}

// HandleDifyWorkflow handle Dify workflow request
func (h *DataFlowAPIHandler) HandleDifyWorkflow(c *gin.Context) {
	// Get auth info from context (set by middleware)
	authInfo, err := GetAuthInfoFromContext(c)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Parse Dify workflow request
	var req struct {
		AgentID      string                 `json:"agent_id,omitempty"`
		Inputs       map[string]interface{} `json:"inputs"`
		User         string                 `json:"user"`
		ResponseMode string                 `json:"response_mode,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Invalid request format: "+err.Error())
		return
	}

	// Use agent_id from request body if provided, otherwise from auth info
	agentID := req.AgentID
	if agentID == "" {
		agentID = authInfo.AgentID
	}

	// Convert to backend request
	backendReq := &backends.BackendRequest{
		AgentID:      agentID,
		APIKey:       authInfo.APIKey,
		User:         req.User,
		Data:         req.Inputs,
		ResponseMode: req.ResponseMode,
		Stream:       req.ResponseMode == "streaming",
	}

	// Process request
	if req.ResponseMode == "streaming" {
		h.handleStreamingRequest(c, backendReq)
	} else {
		h.handleBlockingRequest(c, backendReq)
	}
}

// HandleChat handle legacy unified chat request for backward compatibility
func (h *DataFlowAPIHandler) HandleChat(c *gin.Context) {
	// Get auth info from context (set by middleware)
	authInfo, err := GetAuthInfoFromContext(c)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Parse legacy request (try to parse as unified DataFlowRequest)
	var legacyReq map[string]interface{}
	if err := c.ShouldBindJSON(&legacyReq); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Invalid request format: "+err.Error())
		return
	}

	// Convert legacy request to backend request
	backendReq := &backends.BackendRequest{
		AgentID: authInfo.AgentID,
		APIKey:  authInfo.APIKey,
	}

	// Override agent_id if provided in request
	if agentID, ok := legacyReq["agent_id"].(string); ok && agentID != "" {
		backendReq.AgentID = agentID
	}

	// Try to determine the format and convert
	if messages, ok := legacyReq["messages"]; ok {
		// OpenAI format
		if model, ok := legacyReq["model"].(string); ok {
			backendReq.Model = model
		}
		if messagesSlice, ok := messages.([]interface{}); ok {
			for _, msg := range messagesSlice {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					role, _ := msgMap["role"].(string)
					content, _ := msgMap["content"].(string)
					backendReq.Messages = append(backendReq.Messages, backends.ChatMessage{
						Role:    role,
						Content: content,
					})
				}
			}
		}
		if stream, ok := legacyReq["stream"].(bool); ok {
			backendReq.Stream = stream
		}
	} else if query, ok := legacyReq["query"].(string); ok {
		// Dify format
		backendReq.Query = query
		if user, ok := legacyReq["user"].(string); ok {
			backendReq.User = user
		}
		if inputs, ok := legacyReq["inputs"].(map[string]interface{}); ok {
			backendReq.Inputs = inputs
		}
		if responseMode, ok := legacyReq["response_mode"].(string); ok {
			backendReq.ResponseMode = responseMode
			backendReq.Stream = responseMode == "streaming"
		}
	}

	// Process request
	if backendReq.Stream || backendReq.ResponseMode == "streaming" {
		h.handleStreamingRequest(c, backendReq)
	} else {
		h.handleBlockingRequest(c, backendReq)
	}
}

// HealthCheck handle health check request
func (h *DataFlowAPIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "dataflow-backend",
		"timestamp": gin.H{},
	})
}

// handleStreamingRequest handle streaming request
func (h *DataFlowAPIHandler) handleStreamingRequest(c *gin.Context, req *backends.BackendRequest) {
	// Set SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Process streaming request
	err := h.service.ProcessStreamingRequest(c.Request.Context(), req, c.Writer)
	if err != nil {
		h.writeSSEError(c, "processing_error", err.Error())
		return
	}
}

// handleBlockingRequest handle blocking request
func (h *DataFlowAPIHandler) handleBlockingRequest(c *gin.Context, req *backends.BackendRequest) {
	// Process request
	response, err := h.service.ProcessRequest(c.Request.Context(), req)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "processing_error", err.Error())
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// writeSSEError write SSE error
func (h *DataFlowAPIHandler) writeSSEError(c *gin.Context, errorType, message string) {
	errorData := map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
		},
	}

	jsonData, _ := json.Marshal(errorData)
	c.Writer.Write([]byte("data: " + string(jsonData) + "\n\n"))
	c.Writer.Flush()
}

// respondWithError respond with error
func (h *DataFlowAPIHandler) respondWithError(c *gin.Context, statusCode int, errorType, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"type":    errorType,
			"message": message,
		},
	})
}
