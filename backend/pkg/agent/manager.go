package agent

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// DefaultAgentManager implements the AgentManager interface
type DefaultAgentManager struct {
	config *AgentManagerConfig
	agents map[string]Agent
	mutex  sync.RWMutex

	// Load balancing state
	roundRobinCounter int

	// Health check
	healthCheckTicker *time.Ticker
	healthCheckStop   chan struct{}
}

// NewAgentManager creates a new agent manager
func NewAgentManager(config *AgentManagerConfig) (*DefaultAgentManager, error) {
	if config == nil {
		config = DefaultAgentManagerConfig()
	}

	manager := &DefaultAgentManager{
		config: config,
		agents: make(map[string]Agent),
	}

	// Start health checks if enabled
	if config.EnableHealthChecks {
		manager.startHealthChecks()
	}

	return manager, nil
}

// DefaultAgentManagerConfig returns default configuration for agent manager
func DefaultAgentManagerConfig() *AgentManagerConfig {
	return &AgentManagerConfig{
		LoadBalancingStrategy: Priority,
		EnableHealthChecks:    true,
		HealthCheckInterval:   DefaultHealthCheckInterval,
		DefaultTimeout:        DefaultTimeout,
		MaxRetries:            DefaultMaxRetries,
		EnableMetrics:         true,
	}
}

// RegisterAgent registers a new agent
func (m *DefaultAgentManager) RegisterAgent(agent Agent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	agentID := agent.GetID()
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Validate agent configuration
	if err := agent.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid agent configuration: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if agent already exists
	if _, exists := m.agents[agentID]; exists {
		return fmt.Errorf("agent with ID %s already exists", agentID)
	}

	// Register the agent
	m.agents[agentID] = agent

	return nil
}

// UnregisterAgent removes an agent
func (m *DefaultAgentManager) UnregisterAgent(agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Close the agent
	if err := agent.Close(); err != nil {
		return fmt.Errorf("failed to close agent: %w", err)
	}

	// Remove from map
	delete(m.agents, agentID)

	return nil
}

// GetAgent retrieves an agent by ID
func (m *DefaultAgentManager) GetAgent(agentID string) (Agent, error) {
	if agentID == "" {
		return nil, fmt.Errorf("agent ID cannot be empty")
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return agent, nil
}

// ListAgents returns all registered agents
func (m *DefaultAgentManager) ListAgents() []Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	agents := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}

	return agents
}

// ListAgentsByType returns agents of a specific type
func (m *DefaultAgentManager) ListAgentsByType(agentType AgentType) []Agent {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var agents []Agent
	for _, agent := range m.agents {
		if agent.GetType() == agentType {
			agents = append(agents, agent)
		}
	}

	return agents
}

// GetAvailableAgent returns an available agent for the request
func (m *DefaultAgentManager) GetAvailableAgent(ctx context.Context, request *ChatRequest) (Agent, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get healthy agents
	healthyAgents := m.getHealthyAgents(ctx)
	if len(healthyAgents) == 0 {
		return nil, fmt.Errorf("no healthy agents available")
	}

	// Apply load balancing strategy
	switch m.config.LoadBalancingStrategy {
	case RoundRobin:
		return m.roundRobinSelect(healthyAgents), nil
	case Random:
		return m.randomSelect(healthyAgents), nil
	case Priority:
		return m.prioritySelect(healthyAgents), nil
	case LeastConnections:
		return m.leastConnectionsSelect(healthyAgents), nil
	case WeightedRandom:
		return m.weightedRandomSelect(healthyAgents), nil
	default:
		return m.prioritySelect(healthyAgents), nil
	}
}

// Close closes all agents and cleans up resources
func (m *DefaultAgentManager) Close() error {
	// Stop health checks
	if m.healthCheckTicker != nil {
		m.healthCheckTicker.Stop()
		close(m.healthCheckStop)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Close all agents
	var errors []error
	for agentID, agent := range m.agents {
		if err := agent.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close agent %s: %w", agentID, err))
		}
	}

	// Clear agents map
	m.agents = make(map[string]Agent)

	// Return combined error if any
	if len(errors) > 0 {
		return fmt.Errorf("errors closing agents: %v", errors)
	}

	return nil
}

// getHealthyAgents returns a list of healthy agents
func (m *DefaultAgentManager) getHealthyAgents(ctx context.Context) []agentWithConfig {
	var healthyAgents []agentWithConfig

	for _, agent := range m.agents {
		// Check agent status
		status, err := agent.GetStatus(ctx)
		if err != nil || !status.Health {
			continue
		}

		// Get agent config for load balancing
		config := m.getAgentConfig(agent)
		if config != nil && config.Enabled {
			healthyAgents = append(healthyAgents, agentWithConfig{
				agent:  agent,
				config: config,
			})
		}
	}

	return healthyAgents
}

// getAgentConfig extracts configuration from agent (type assertion)
func (m *DefaultAgentManager) getAgentConfig(agent Agent) *AgentConfig {
	switch a := agent.(type) {
	case *OpenAIAgent:
		return &a.config.AgentConfig
	case *DifyAgent:
		return &a.config.AgentConfig
	default:
		// Return default config for unknown agent types
		return &AgentConfig{
			ID:                    agent.GetID(),
			Name:                  agent.GetName(),
			Type:                  agent.GetType(),
			Enabled:               true,
			Priority:              50,
			Timeout:               DefaultTimeout,
			MaxConcurrentRequests: DefaultMaxConcurrentRequests,
		}
	}
}

// Load balancing strategies

// roundRobinSelect selects agent using round-robin strategy
func (m *DefaultAgentManager) roundRobinSelect(agents []agentWithConfig) Agent {
	if len(agents) == 0 {
		return nil
	}

	m.roundRobinCounter = (m.roundRobinCounter + 1) % len(agents)
	return agents[m.roundRobinCounter].agent
}

// randomSelect selects agent randomly
func (m *DefaultAgentManager) randomSelect(agents []agentWithConfig) Agent {
	if len(agents) == 0 {
		return nil
	}

	return agents[rand.Intn(len(agents))].agent
}

// prioritySelect selects agent with highest priority
func (m *DefaultAgentManager) prioritySelect(agents []agentWithConfig) Agent {
	if len(agents) == 0 {
		return nil
	}

	// Sort by priority (descending)
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].config.Priority > agents[j].config.Priority
	})

	return agents[0].agent
}

// leastConnectionsSelect selects agent with least connections
func (m *DefaultAgentManager) leastConnectionsSelect(agents []agentWithConfig) Agent {
	if len(agents) == 0 {
		return nil
	}

	// For simplicity, we'll use random selection here
	// In a real implementation, you'd track active connections per agent
	return m.randomSelect(agents)
}

// weightedRandomSelect selects agent using weighted random based on priority
func (m *DefaultAgentManager) weightedRandomSelect(agents []agentWithConfig) Agent {
	if len(agents) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, agent := range agents {
		totalWeight += agent.config.Priority
	}

	if totalWeight == 0 {
		return m.randomSelect(agents)
	}

	// Generate random number
	randomNum := rand.Intn(totalWeight)

	// Select agent based on weight
	currentWeight := 0
	for _, agent := range agents {
		currentWeight += agent.config.Priority
		if randomNum < currentWeight {
			return agent.agent
		}
	}

	// Fallback to last agent
	return agents[len(agents)-1].agent
}

// Health check functionality

// startHealthChecks starts periodic health checks
func (m *DefaultAgentManager) startHealthChecks() {
	m.healthCheckTicker = time.NewTicker(m.config.HealthCheckInterval)
	m.healthCheckStop = make(chan struct{})

	go func() {
		for {
			select {
			case <-m.healthCheckTicker.C:
				m.performHealthChecks()
			case <-m.healthCheckStop:
				return
			}
		}
	}()
}

// performHealthChecks performs health checks on all agents
func (m *DefaultAgentManager) performHealthChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.DefaultTimeout)
	defer cancel()

	m.mutex.RLock()
	agents := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	m.mutex.RUnlock()

	// Perform health checks concurrently
	for _, agent := range agents {
		go func(a Agent) {
			_, err := a.GetStatus(ctx)
			if err != nil {
				// Log error or handle unhealthy agent
				// This could trigger alerts, remove from rotation, etc.
			}
		}(agent)
	}
}

// Helper types

// agentWithConfig combines agent with its configuration for load balancing
type agentWithConfig struct {
	agent  Agent
	config *AgentConfig
}

// AgentInfo represents agent information for management
type AgentInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         AgentType         `json:"type"`
	Status       *AgentStatus      `json:"status"`
	Capabilities AgentCapabilities `json:"capabilities"`
	Config       *AgentConfig      `json:"config"`
}

// GetAgentInfo returns detailed information about an agent
func (m *DefaultAgentManager) GetAgentInfo(ctx context.Context, agentID string) (*AgentInfo, error) {
	agent, err := m.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	status, err := agent.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent status: %w", err)
	}

	return &AgentInfo{
		ID:           agent.GetID(),
		Name:         agent.GetName(),
		Type:         agent.GetType(),
		Status:       status,
		Capabilities: agent.GetCapabilities(),
		Config:       m.getAgentConfig(agent),
	}, nil
}

// ListAgentInfos returns detailed information about all agents
func (m *DefaultAgentManager) ListAgentInfos(ctx context.Context) ([]*AgentInfo, error) {
	agents := m.ListAgents()
	infos := make([]*AgentInfo, 0, len(agents))

	for _, agent := range agents {
		info, err := m.GetAgentInfo(ctx, agent.GetID())
		if err != nil {
			// Skip agents with errors but don't fail the entire operation
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// AgentMetrics represents metrics for an agent
type AgentMetrics struct {
	AgentID         string        `json:"agent_id"`
	RequestCount    int           `json:"request_count"`
	ErrorCount      int           `json:"error_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageResponse time.Duration `json:"average_response_time"`
	LastRequest     time.Time     `json:"last_request"`
	Uptime          time.Duration `json:"uptime"`
}

// GetAgentMetrics returns metrics for a specific agent
func (m *DefaultAgentManager) GetAgentMetrics(ctx context.Context, agentID string) (*AgentMetrics, error) {
	agent, err := m.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	status, err := agent.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &AgentMetrics{
		AgentID:         agentID,
		RequestCount:    status.RequestCount,
		ErrorCount:      status.ErrorCount,
		SuccessRate:     status.SuccessRate,
		AverageResponse: time.Duration(status.ResponseTime) * time.Millisecond,
		LastRequest:     status.LastChecked,
		// Uptime calculation would require tracking start time
	}, nil
}

// GetAllAgentMetrics returns metrics for all agents
func (m *DefaultAgentManager) GetAllAgentMetrics(ctx context.Context) (map[string]*AgentMetrics, error) {
	agents := m.ListAgents()
	metrics := make(map[string]*AgentMetrics)

	for _, agent := range agents {
		metric, err := m.GetAgentMetrics(ctx, agent.GetID())
		if err != nil {
			continue // Skip agents with errors
		}
		metrics[agent.GetID()] = metric
	}

	return metrics, nil
}
