package controlflow

import (
	"agent-connector/internal"
	"time"
)

// ControlFlowResponse control flow API common response structure
type ControlFlowResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError API error structure
type APIError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ControlFlowPaginationResponse control flow API pagination response structure
type ControlFlowPaginationResponse struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Error      *APIError      `json:"error,omitempty"`
}

// PaginationInfo pagination information
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SystemConfigRequest system configuration request structure
type SystemConfigRequest struct {
	// Currently no configurable fields, but keeping structure for future use
}

// SystemConfigResponse system configuration response structure
type SystemConfigResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgentRequest agent configuration request structure
type AgentRequest struct {
	Name             string `json:"name" binding:"required"`
	Type             string `json:"type" binding:"required,oneof=openai openai_compatible dify"`
	URL              string `json:"url" binding:"required,url"`
	SourceAPIKey     string `json:"source_api_key" binding:"required"`
	QPS              int    `json:"qps" binding:"min=1"`
	Enabled          bool   `json:"enabled"`
	Description      string `json:"description"`
	SupportStreaming bool   `json:"support_streaming"`
	ResponseFormat   string `json:"response_format" binding:"oneof=openai dify"`
}

// AgentResponse agent configuration response structure
type AgentResponse struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	SourceAPIKey     string    `json:"source_api_key,omitempty"` // in some cases, it may be necessary to hide
	ConnectorAPIKey  string    `json:"connector_api_key"`
	AgentID          string    `json:"agent_id"`
	QPS              int       `json:"qps"`
	Enabled          bool      `json:"enabled"`
	Description      string    `json:"description"`
	SupportStreaming bool      `json:"support_streaming"`
	ResponseFormat   string    `json:"response_format"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AgentUpdateRequest agent update request structure
type AgentUpdateRequest struct {
	Name             *string `json:"name,omitempty"`
	Type             *string `json:"type,omitempty" binding:"omitempty,oneof=openai openai_compatible dify"`
	URL              *string `json:"url,omitempty" binding:"omitempty,url"`
	SourceAPIKey     *string `json:"source_api_key,omitempty"`
	QPS              *int    `json:"qps,omitempty" binding:"omitempty,min=1"`
	Enabled          *bool   `json:"enabled,omitempty"`
	Description      *string `json:"description,omitempty"`
	SupportStreaming *bool   `json:"support_streaming,omitempty"`
	ResponseFormat   *string `json:"response_format,omitempty" binding:"omitempty,oneof=openai dify"`
}

// HealthCheckResponse health check response
type HealthCheckResponse struct {
	Status     string                 `json:"status"`
	Service    string                 `json:"service"`
	Version    string                 `json:"version"`
	Timestamp  int64                  `json:"timestamp"`
	Uptime     string                 `json:"uptime"`
	Database   DatabaseHealthStatus   `json:"database"`
	Components map[string]interface{} `json:"components,omitempty"`
}

// DatabaseHealthStatus database health status
type DatabaseHealthStatus struct {
	Status      string `json:"status"`
	Connections int    `json:"connections,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ConvertFromInternalSystemConfig convert from internal model to response structure
func ConvertFromInternalSystemConfig(config *internal.SystemConfig) *SystemConfigResponse {
	return &SystemConfigResponse{
		ID:        config.ID,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}
}

// ConvertToInternalSystemConfig convert from request structure to internal model
func ConvertToInternalSystemConfig(req *SystemConfigRequest) *internal.SystemConfig {
	return &internal.SystemConfig{}
}

// ConvertFromInternalAgent convert from internal model to response structure
func ConvertFromInternalAgent(agent *internal.Agent, hideSecrets bool) *AgentResponse {
	response := &AgentResponse{
		ID:               agent.ID,
		Name:             agent.Name,
		Type:             string(agent.Type),
		URL:              agent.URL,
		ConnectorAPIKey:  agent.ConnectorAPIKey,
		AgentID:          agent.AgentID,
		QPS:              agent.QPS,
		Enabled:          agent.Enabled,
		Description:      agent.Description,
		SupportStreaming: agent.SupportStreaming,
		ResponseFormat:   agent.ResponseFormat,
		CreatedAt:        agent.CreatedAt,
		UpdatedAt:        agent.UpdatedAt,
	}

	// decide whether to hide sensitive information based on the need
	if !hideSecrets {
		response.SourceAPIKey = agent.SourceAPIKey
	}

	return response
}

// ConvertToInternalAgent convert from request structure to internal model
func ConvertToInternalAgent(req *AgentRequest) *internal.Agent {
	return &internal.Agent{
		Name:             req.Name,
		Type:             internal.AgentType(req.Type),
		URL:              req.URL,
		SourceAPIKey:     req.SourceAPIKey,
		QPS:              req.QPS,
		Enabled:          req.Enabled,
		Description:      req.Description,
		SupportStreaming: req.SupportStreaming,
		ResponseFormat:   req.ResponseFormat,
	}
}

// UpdateInternalAgentFromRequest update internal model with request data
func UpdateInternalAgentFromRequest(agent *internal.Agent, req *AgentUpdateRequest) {
	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Type != nil {
		agent.Type = internal.AgentType(*req.Type)
	}
	if req.URL != nil {
		agent.URL = *req.URL
	}
	if req.SourceAPIKey != nil {
		agent.SourceAPIKey = *req.SourceAPIKey
	}
	if req.QPS != nil {
		agent.QPS = *req.QPS
	}
	if req.Enabled != nil {
		agent.Enabled = *req.Enabled
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.SupportStreaming != nil {
		agent.SupportStreaming = *req.SupportStreaming
	}
	if req.ResponseFormat != nil {
		agent.ResponseFormat = *req.ResponseFormat
	}
}

// ConvertFromInternalAgentList convert from internal model list to response list
func ConvertFromInternalAgentList(agents []*internal.Agent, hideSecrets bool) []*AgentResponse {
	result := make([]*AgentResponse, len(agents))
	for i, agent := range agents {
		result[i] = ConvertFromInternalAgent(agent, hideSecrets)
	}
	return result
}
