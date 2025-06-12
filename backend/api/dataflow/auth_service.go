package dataflow

import (
	"agent-connector/internal"
	"errors"
	"strings"
	"time"
)

// DataFlowAuthService data flow API authentication service
type DataFlowAuthService struct {
	agentService *internal.AgentService
}

// NewDataFlowAuthService create data flow API authentication service
func NewDataFlowAuthService() *DataFlowAuthService {
	return &DataFlowAuthService{
		agentService: &internal.AgentService{},
	}
}

// AuthenticateRequest authenticate request
func (s *DataFlowAuthService) AuthenticateRequest(agentID, apiKey string) (*AuthInfo, error) {
	// parameter validation
	if agentID == "" {
		return nil, errors.New("agent_id is required")
	}

	if apiKey == "" {
		return nil, errors.New("api_key is required")
	}

	// clean API key format (remove Bearer prefix)
	apiKey = s.cleanAPIKey(apiKey)

	// find agent by agent ID
	agent, err := s.findAgentByAgentID(agentID)
	if err != nil {
		return nil, err
	}

	// validate API key
	if agent.ConnectorAPIKey != apiKey {
		return nil, errors.New("invalid api_key")
	}

	// check if agent is enabled
	if !agent.Enabled {
		return nil, errors.New("agent is disabled")
	}

	// build authentication information
	authInfo := &AuthInfo{
		AgentID:   agentID,
		APIKey:    apiKey,
		Timestamp: time.Now(),
		Agent: &AgentInfo{
			ID:               agent.ID,
			Name:             agent.Name,
			Type:             string(agent.Type),
			URL:              agent.URL,
			SourceAPIKey:     agent.SourceAPIKey,
			QPS:              agent.QPS,
			Enabled:          agent.Enabled,
			SupportStreaming: agent.SupportStreaming,
			ResponseFormat:   agent.ResponseFormat,
		},
	}

	return authInfo, nil
}

// findAgentByAgentID find agent by agent ID
func (s *DataFlowAuthService) findAgentByAgentID(agentID string) (*internal.Agent, error) {
	return s.agentService.GetAgentByAgentID(agentID)
}

// cleanAPIKey clean API key format
func (s *DataFlowAuthService) cleanAPIKey(apiKey string) string {
	// remove Bearer prefix
	if strings.HasPrefix(apiKey, "Bearer ") {
		return strings.TrimPrefix(apiKey, "Bearer ")
	}

	// remove bearer prefix (lowercase)
	if strings.HasPrefix(apiKey, "bearer ") {
		return strings.TrimPrefix(apiKey, "bearer ")
	}

	return strings.TrimSpace(apiKey)
}

// ValidateAgent validate agent configuration
func (s *DataFlowAuthService) ValidateAgent(agent *AgentInfo) error {
	if agent == nil {
		return errors.New("agent info is nil")
	}

	if agent.Name == "" {
		return errors.New("agent name is empty")
	}

	if agent.Type == "" {
		return errors.New("agent type is empty")
	}

	if agent.URL == "" {
		return errors.New("agent URL is empty")
	}

	if agent.SourceAPIKey == "" {
		return errors.New("agent source API key is empty")
	}

	// validate agent type
	validTypes := []string{"openai", "openai_compatible", "dify"}
	isValidType := false
	for _, validType := range validTypes {
		if agent.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return errors.New("invalid agent type: " + agent.Type)
	}

	// validate response format
	if agent.ResponseFormat != "openai" && agent.ResponseFormat != "dify" {
		return errors.New("invalid response format: " + agent.ResponseFormat)
	}

	return nil
}

// GetUserIDFromAPIKey get user ID from API key (simple implementation)
func (s *DataFlowAuthService) GetUserIDFromAPIKey(apiKey string) string {
	// here we use a simple strategy: take the first 8 characters of the API key as the user identifier
	// in a real project, more complex user identification logic may be needed
	if len(apiKey) >= 8 {
		return "user_" + apiKey[:8]
	}
	return "user_default"
}

// IsAgentTypeCompatible check if agent type is compatible with request format
func (s *DataFlowAuthService) IsAgentTypeCompatible(agentType, requestFormat string) bool {
	switch requestFormat {
	case "openai":
		return agentType == "openai" || agentType == "openai_compatible"
	case "dify":
		return agentType == "dify"
	default:
		return false
	}
}

// DetermineRequestFormat determine request format based on request content
func (s *DataFlowAuthService) DetermineRequestFormat(req *DataFlowRequest) string {
	// if there is a messages field, it is usually OpenAI format
	if req.Messages != nil && len(req.Messages) > 0 {
		return "openai"
	}

	// if there is a query field, it is usually Dify format
	if req.Query != "" {
		return "dify"
	}

	// default return OpenAI format
	return "openai"
}

// GenerateRequestID generate request ID
func (s *DataFlowAuthService) GenerateRequestID() string {
	return "req_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generate random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
