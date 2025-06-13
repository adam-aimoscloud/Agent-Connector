package backends

import (
	"agent-connector/pkg/types"
	"context"
	"io"
	"net/http"
)

// AgentBackend defines the interface for different agent backend implementations
type AgentBackend interface {
	// GetType returns the backend type (openai, dify-chat, dify-workflow)
	GetType() types.AgentType

	// ValidateRequest validates the incoming request for this backend type
	ValidateRequest(req *BackendRequest) error

	// BuildForwardRequest builds the HTTP request to forward to the actual agent
	BuildForwardRequest(ctx context.Context, req *BackendRequest, agentInfo *AgentInfo) (*http.Request, error)

	// ProcessBlockingResponse processes the response from the agent for blocking requests
	ProcessBlockingResponse(resp *http.Response) (interface{}, error)

	// ProcessStreamingResponse processes the response from the agent for streaming requests
	ProcessStreamingResponse(resp *http.Response) (io.ReadCloser, error)

	// GetEndpoint returns the endpoint path for this backend type
	GetEndpoint() string
}

// Import BackendType from unified types package
// BackendType is now defined in pkg/types/backend_types.go

// BackendRequest represents a unified request structure
type BackendRequest struct {
	// Common fields
	AgentID string `json:"agent_id,omitempty"`
	APIKey  string `json:"-"`

	// OpenAI Compatible fields
	Model       string        `json:"model,omitempty"`
	Messages    []ChatMessage `json:"messages,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`

	// Dify Chat fields
	Query          string                 `json:"query,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user,omitempty"`
	Inputs         map[string]interface{} `json:"inputs,omitempty"`
	ResponseMode   string                 `json:"response_mode,omitempty"`

	// Dify Workflow fields
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentInfo represents agent configuration
type AgentInfo struct {
	ID               uint
	Name             string
	Type             string
	URL              string
	SourceAPIKey     string
	QPS              int
	Enabled          bool
	SupportStreaming bool
	ResponseFormat   string
}

// BackendFactory creates backend instances
type BackendFactory interface {
	CreateBackend(backendType types.AgentType) (AgentBackend, error)
	GetSupportedTypes() []types.AgentType
}
