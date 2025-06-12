package agent

import (
	"context"
	"io"
	"time"
)

// Agent represents a unified interface for different agent sources
type Agent interface {
	// GetID returns the unique identifier of the agent
	GetID() string

	// GetName returns the display name of the agent
	GetName() string

	// GetType returns the type of the agent source
	GetType() AgentType

	// GetCapabilities returns the capabilities of the agent
	GetCapabilities() AgentCapabilities

	// Chat sends a chat message and returns the response
	Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)

	// ChatStream sends a chat message and returns a streaming response
	ChatStream(ctx context.Context, request *ChatRequest) (*ChatStreamResponse, error)

	// GetModels returns available models for this agent
	GetModels(ctx context.Context) ([]Model, error)

	// ValidateConfig validates the agent configuration
	ValidateConfig() error

	// GetStatus returns the current status of the agent
	GetStatus(ctx context.Context) (*AgentStatus, error)

	// Close cleans up resources used by the agent
	Close() error
}

// AgentManager manages multiple agent sources
type AgentManager interface {
	// RegisterAgent registers a new agent
	RegisterAgent(agent Agent) error

	// UnregisterAgent removes an agent
	UnregisterAgent(agentID string) error

	// GetAgent retrieves an agent by ID
	GetAgent(agentID string) (Agent, error)

	// ListAgents returns all registered agents
	ListAgents() []Agent

	// ListAgentsByType returns agents of a specific type
	ListAgentsByType(agentType AgentType) []Agent

	// GetAvailableAgent returns an available agent for the request
	GetAvailableAgent(ctx context.Context, request *ChatRequest) (Agent, error)

	// Close closes all agents and cleans up resources
	Close() error
}

// AgentType represents the type of agent source
type AgentType string

const (
	// AgentTypeOpenAI represents OpenAI compatible API agents
	AgentTypeOpenAI AgentType = "openai"

	// AgentTypeDify represents Dify platform agents
	AgentTypeDify AgentType = "dify"
)

// String returns the string representation of the agent type
func (at AgentType) String() string {
	return string(at)
}

// IsValid checks if the agent type is valid
func (at AgentType) IsValid() bool {
	switch at {
	case AgentTypeOpenAI, AgentTypeDify:
		return true
	default:
		return false
	}
}

// AgentCapabilities represents the capabilities of an agent
type AgentCapabilities struct {
	// SupportsChatCompletion indicates if the agent supports chat completion
	SupportsChatCompletion bool `json:"supports_chat_completion"`

	// SupportsStreaming indicates if the agent supports streaming responses
	SupportsStreaming bool `json:"supports_streaming"`

	// SupportsImages indicates if the agent supports image inputs
	SupportsImages bool `json:"supports_images"`

	// SupportsFiles indicates if the agent supports file inputs
	SupportsFiles bool `json:"supports_files"`

	// SupportsFunctionCalling indicates if the agent supports function calling
	SupportsFunctionCalling bool `json:"supports_function_calling"`

	// MaxTokens is the maximum number of tokens supported
	MaxTokens int `json:"max_tokens"`

	// SupportedLanguages is the list of supported languages
	SupportedLanguages []string `json:"supported_languages"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	// Messages is the conversation history
	Messages []Message `json:"messages"`

	// Model is the model to use for completion
	Model string `json:"model,omitempty"`

	// Temperature controls randomness in the response
	Temperature *float32 `json:"temperature,omitempty"`

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`

	// Functions for function calling (OpenAI compatible)
	Functions []Function `json:"functions,omitempty"`

	// Tools for tool calling (newer OpenAI format)
	Tools []Tool `json:"tools,omitempty"`

	// User ID for tracking and rate limiting
	UserID string `json:"user_id,omitempty"`

	// SessionID for conversation tracking
	SessionID string `json:"session_id,omitempty"`

	// Metadata for additional information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	// ID is the unique identifier for the response
	ID string `json:"id"`

	// Object type (e.g., "chat.completion")
	Object string `json:"object"`

	// Created timestamp
	Created int64 `json:"created"`

	// Model used for the response
	Model string `json:"model"`

	// Choices contains the response options
	Choices []Choice `json:"choices"`

	// Usage contains token usage information
	Usage *Usage `json:"usage,omitempty"`

	// Error contains error information if any
	Error *AgentError `json:"error,omitempty"`

	// Metadata for additional information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatStreamResponse represents a streaming chat response
type ChatStreamResponse struct {
	// Stream is the response stream
	Stream io.ReadCloser `json:"-"`

	// Events is a channel for streaming events
	Events <-chan StreamEvent `json:"-"`

	// Error channel for streaming errors
	Errors <-chan error `json:"-"`
}

// StreamEvent represents a streaming response event
type StreamEvent struct {
	// Type of the event (e.g., "content", "function_call", "done")
	Type string `json:"type"`

	// Data contains the event data
	Data interface{} `json:"data"`

	// Delta contains incremental content
	Delta *Delta `json:"delta,omitempty"`

	// FinishReason indicates why the stream finished
	FinishReason *string `json:"finish_reason,omitempty"`
}

// Message represents a chat message
type Message struct {
	// Role of the message sender (system, user, assistant, function)
	Role string `json:"role"`

	// Content of the message
	Content string `json:"content"`

	// Name of the function (for function messages)
	Name string `json:"name,omitempty"`

	// FunctionCall for assistant messages with function calls
	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// ToolCalls for assistant messages with tool calls
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	// Index of the choice
	Index int `json:"index"`

	// Message content
	Message Message `json:"message"`

	// FinishReason indicates why the generation stopped
	FinishReason *string `json:"finish_reason"`
}

// Delta represents incremental content in streaming
type Delta struct {
	// Role of the message sender
	Role string `json:"role,omitempty"`

	// Content delta
	Content string `json:"content,omitempty"`

	// FunctionCall delta
	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// ToolCalls delta
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	// PromptTokens is the number of tokens in the prompt
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the completion
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the total number of tokens used
	TotalTokens int `json:"total_tokens"`
}

// Function represents a function definition for function calling
type Function struct {
	// Name of the function
	Name string `json:"name"`

	// Description of what the function does
	Description string `json:"description"`

	// Parameters schema for the function
	Parameters map[string]interface{} `json:"parameters"`
}

// Tool represents a tool definition
type Tool struct {
	// Type of the tool (e.g., "function")
	Type string `json:"type"`

	// Function definition
	Function Function `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	// Name of the function to call
	Name string `json:"name"`

	// Arguments for the function call (JSON string)
	Arguments string `json:"arguments"`
}

// ToolCall represents a tool call
type ToolCall struct {
	// ID of the tool call
	ID string `json:"id"`

	// Type of the tool call
	Type string `json:"type"`

	// Function call details
	Function FunctionCall `json:"function"`
}

// Model represents an available model
type Model struct {
	// ID of the model
	ID string `json:"id"`

	// Name of the model
	Name string `json:"name"`

	// Description of the model
	Description string `json:"description"`

	// Created timestamp
	Created int64 `json:"created"`

	// OwnedBy indicates who owns the model
	OwnedBy string `json:"owned_by"`

	// Capabilities of the model
	Capabilities AgentCapabilities `json:"capabilities"`
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	// ID of the agent
	AgentID string `json:"agent_id"`

	// Status of the agent (online, offline, error, maintenance)
	Status string `json:"status"`

	// Health indicates if the agent is healthy
	Health bool `json:"health"`

	// LastChecked timestamp
	LastChecked time.Time `json:"last_checked"`

	// ResponseTime in milliseconds
	ResponseTime int64 `json:"response_time_ms"`

	// ErrorCount in the last period
	ErrorCount int `json:"error_count"`

	// RequestCount in the last period
	RequestCount int `json:"request_count"`

	// SuccessRate percentage
	SuccessRate float64 `json:"success_rate"`

	// Additional status information
	Details map[string]interface{} `json:"details,omitempty"`
}

// AgentError represents an error from an agent
type AgentError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Type is the error type
	Type string `json:"type"`

	// Param is the parameter that caused the error
	Param string `json:"param,omitempty"`

	// Details contains additional error details
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AgentError) Error() string {
	return e.Message
}

// AgentConfig represents the base configuration for agents
type AgentConfig struct {
	// ID of the agent
	ID string `json:"id"`

	// Name of the agent
	Name string `json:"name"`

	// Type of the agent
	Type AgentType `json:"type"`

	// Enabled indicates if the agent is enabled
	Enabled bool `json:"enabled"`

	// Priority for agent selection (higher = more preferred)
	Priority int `json:"priority"`

	// Timeout for requests to this agent
	Timeout time.Duration `json:"timeout"`

	// MaxConcurrentRequests limits concurrent requests
	MaxConcurrentRequests int `json:"max_concurrent_requests"`

	// RetryPolicy for failed requests
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`

	// HealthCheck configuration
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty"`

	// Metadata for additional configuration
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RetryPolicy defines retry behavior for failed requests
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"max_retries"`

	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration `json:"initial_delay"`

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration `json:"max_delay"`

	// Multiplier for exponential backoff
	Multiplier float64 `json:"multiplier"`

	// RetryableErrors are error codes that should trigger retries
	RetryableErrors []string `json:"retryable_errors"`
}

// HealthCheckConfig defines health check behavior
type HealthCheckConfig struct {
	// Enabled indicates if health checks are enabled
	Enabled bool `json:"enabled"`

	// Interval between health checks
	Interval time.Duration `json:"interval"`

	// Timeout for health check requests
	Timeout time.Duration `json:"timeout"`

	// FailureThreshold before marking as unhealthy
	FailureThreshold int `json:"failure_threshold"`

	// SuccessThreshold before marking as healthy
	SuccessThreshold int `json:"success_threshold"`
}

// LoadBalancingStrategy defines how to select agents
type LoadBalancingStrategy string

const (
	// RoundRobin strategy
	RoundRobin LoadBalancingStrategy = "round_robin"

	// Random strategy
	Random LoadBalancingStrategy = "random"

	// Priority strategy (use highest priority first)
	Priority LoadBalancingStrategy = "priority"

	// LeastConnections strategy
	LeastConnections LoadBalancingStrategy = "least_connections"

	// WeightedRandom strategy
	WeightedRandom LoadBalancingStrategy = "weighted_random"
)

// AgentManagerConfig represents configuration for the agent manager
type AgentManagerConfig struct {
	// LoadBalancingStrategy for agent selection
	LoadBalancingStrategy LoadBalancingStrategy `json:"load_balancing_strategy"`

	// EnableHealthChecks indicates if health checks should be performed
	EnableHealthChecks bool `json:"enable_health_checks"`

	// HealthCheckInterval for periodic health checks
	HealthCheckInterval time.Duration `json:"health_check_interval"`

	// DefaultTimeout for agent requests
	DefaultTimeout time.Duration `json:"default_timeout"`

	// MaxRetries for failed requests
	MaxRetries int `json:"max_retries"`

	// EnableMetrics indicates if metrics should be collected
	EnableMetrics bool `json:"enable_metrics"`
}

// Default values for configuration
const (
	DefaultTimeout               = 30 * time.Second
	DefaultMaxConcurrentRequests = 10
	DefaultHealthCheckInterval   = 1 * time.Minute
	DefaultMaxRetries            = 3
)
