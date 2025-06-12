package dataflow

import (
	"encoding/json"
	"time"
)

// DataFlowRequest data flow API common request structure
type DataFlowRequest struct {
	AgentID string `json:"agent_id" binding:"required"`
	APIKey  string `json:"-"` // get from header

	// OpenAI Compatible fields
	Model       string        `json:"model,omitempty"`
	Messages    []ChatMessage `json:"messages,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`

	// Dify Compatible fields
	Query          string                 `json:"query,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user,omitempty"`
	Inputs         map[string]interface{} `json:"inputs,omitempty"`

	// Common fields
	ResponseMode string `json:"response_mode,omitempty"` // "streaming" or "blocking"
}

// ChatMessage OpenAI format message structure
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// DataFlowResponse data flow API common response structure
type DataFlowResponse struct {
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
}

// OpenAI Compatible Response Structures

// OpenAIResponse OpenAI format response
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice OpenAI choice structure
type OpenAIChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason *string      `json:"finish_reason"`
}

// OpenAIUsage OpenAI usage statistics
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamResponse OpenAI streaming response
type OpenAIStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}

// Dify Compatible Response Structures

// DifyResponse Dify format response
type DifyResponse struct {
	Event          string                 `json:"event,omitempty"`
	TaskID         string                 `json:"task_id,omitempty"`
	ID             string                 `json:"id,omitempty"`
	MessageID      string                 `json:"message_id,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Mode           string                 `json:"mode,omitempty"`
	Answer         string                 `json:"answer,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      int64                  `json:"created_at,omitempty"`
}

// DifyStreamResponse Dify streaming response
type DifyStreamResponse struct {
	Event          string                 `json:"event"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	MessageID      string                 `json:"message_id,omitempty"`
	CreatedAt      int64                  `json:"created_at,omitempty"`
	TaskID         string                 `json:"task_id,omitempty"`
	Answer         string                 `json:"answer,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// AuthInfo authentication information
type AuthInfo struct {
	AgentID   string
	UserID    string
	APIKey    string
	Agent     *AgentInfo
	Timestamp time.Time
}

// AgentInfo agent information
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

// RateLimitInfo rate limit information
type RateLimitInfo struct {
	Mode             string // "priority" or "qps"
	UserPriority     int
	UserQPS          int
	AgentQPS         int
	CurrentUserRate  float64
	CurrentAgentRate float64
	Allowed          bool
	WaitTime         time.Duration
}

// RequestContext request context
type RequestContext struct {
	RequestID     string
	AuthInfo      *AuthInfo
	RateLimitInfo *RateLimitInfo
	StartTime     time.Time
	Agent         *AgentInfo
}

// StreamData streaming data wrapper
type StreamData struct {
	Data  interface{} `json:"data"`
	Event string      `json:"event,omitempty"`
	ID    string      `json:"id,omitempty"`
	Retry int         `json:"retry,omitempty"`
}

// ConvertToJSON convert to JSON format
func (s *StreamData) ConvertToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ToSSEFormat convert to SSE format
func (s *StreamData) ToSSEFormat() string {
	var result string

	if s.Event != "" {
		result += "event: " + s.Event + "\n"
	}

	if s.ID != "" {
		result += "id: " + s.ID + "\n"
	}

	if s.Retry > 0 {
		result += "retry: " + string(rune(s.Retry)) + "\n"
	}

	data, _ := json.Marshal(s.Data)
	result += "data: " + string(data) + "\n\n"

	return result
}

// ResponseFormat response format enumeration
type ResponseFormat string

const (
	ResponseFormatOpenAI ResponseFormat = "openai"
	ResponseFormatDify   ResponseFormat = "dify"
)

// StreamMode stream mode enumeration
type StreamMode string

const (
	StreamModeStreaming StreamMode = "streaming"
	StreamModeBlocking  StreamMode = "blocking"
)
