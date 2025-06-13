package backends

import (
	"fmt"
	"agent-connector/pkg/types"
)

type DefaultBackendFactory struct{}

func NewDefaultBackendFactory() *DefaultBackendFactory {
	return &DefaultBackendFactory{}
}

func (f *DefaultBackendFactory) CreateBackend(agentType types.AgentType) (AgentBackend, error) {
	switch agentType {
	case types.AgentTypeOpenAI:
		return NewOpenAIBackend(), nil
	case types.AgentTypeDifyChat:
		return NewDifyChatBackend(), nil
	case types.AgentTypeDifyWorkflow:
		return NewDifyWorkflowBackend(), nil
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

func (f *DefaultBackendFactory) GetSupportedTypes() []types.AgentType {
	return types.GetAllAgentTypes()
}

func DetermineAgentType(agentType string) types.AgentType {
	switch agentType {
	case "openai":
		return types.AgentTypeOpenAI
	case "dify-chat":
		return types.AgentTypeDifyChat
	case "dify-workflow":
		return types.AgentTypeDifyWorkflow
	default:
		return types.AgentTypeOpenAI
	}
}
