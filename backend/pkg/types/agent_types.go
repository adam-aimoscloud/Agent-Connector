package types

// AgentType represents the type of agent for processing
type AgentType string

// Agent type constants - unified type system
const (
	// OpenAI compatible agents
	AgentTypeOpenAI AgentType = "openai"

	// Dify agents
	AgentTypeDifyChat     AgentType = "dify-chat"
	AgentTypeDifyWorkflow AgentType = "dify-workflow"
)

// Response format constants
const (
	ResponseFormatOpenAI = "openai"
	ResponseFormatDify   = "dify"
)

// GetAllAgentTypes returns all supported agent types
func GetAllAgentTypes() []AgentType {
	return []AgentType{
		AgentTypeOpenAI,
		AgentTypeDifyChat,
		AgentTypeDifyWorkflow,
	}
}

// GetAllAgentTypeStrings returns all supported agent types as strings
func GetAllAgentTypeStrings() []string {
	types := GetAllAgentTypes()
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}

// GetAllResponseFormats returns all supported response formats
func GetAllResponseFormats() []string {
	return []string{
		ResponseFormatOpenAI,
		ResponseFormatDify,
	}
}

// IsValidAgentType checks if the given agent type is valid
func IsValidAgentType(agentType string) bool {
	validTypes := GetAllAgentTypeStrings()
	for _, validType := range validTypes {
		if agentType == validType {
			return true
		}
	}
	return false
}

// IsValidResponseFormat checks if the given response format is valid
func IsValidResponseFormat(format string) bool {
	validFormats := GetAllResponseFormats()
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}

// GetDefaultResponseFormat returns the default response format for an agent type
func GetDefaultResponseFormat(agentType AgentType) string {
	switch agentType {
	case AgentTypeOpenAI:
		return ResponseFormatOpenAI
	case AgentTypeDifyChat, AgentTypeDifyWorkflow:
		return ResponseFormatDify
	default:
		return ResponseFormatOpenAI
	}
}

// String returns the string representation of the agent type
func (at AgentType) String() string {
	return string(at)
}

// IsValid checks if the agent type is valid
func (at AgentType) IsValid() bool {
	return IsValidAgentType(string(at))
}
