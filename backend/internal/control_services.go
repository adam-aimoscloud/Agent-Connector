package internal

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// SystemConfigService system configuration service
type SystemConfigService struct{}

// GetSystemConfig get system configuration
func (s *SystemConfigService) GetSystemConfig() (*SystemConfig, error) {
	var config SystemConfig
	err := DB.First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// if no configuration, return default configuration
			return &SystemConfig{
				RateLimitMode:   RateLimitModePriority,
				DefaultPriority: 5,
				DefaultQPS:      10,
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

// UpdateSystemConfig update system configuration
func (s *SystemConfigService) UpdateSystemConfig(config *SystemConfig) error {
	// validate configuration
	if err := s.validateSystemConfig(config); err != nil {
		return err
	}

	var existingConfig SystemConfig
	err := DB.First(&existingConfig).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// create new configuration
		return DB.Create(config).Error
	} else if err != nil {
		return err
	}

	// update existing configuration
	config.ID = existingConfig.ID
	return DB.Save(config).Error
}

// validateSystemConfig validate system configuration
func (s *SystemConfigService) validateSystemConfig(config *SystemConfig) error {
	if config.RateLimitMode != RateLimitModePriority && config.RateLimitMode != RateLimitModeQPS {
		return errors.New("invalid rate limit mode")
	}

	if config.DefaultPriority < 1 || config.DefaultPriority > 10 {
		return errors.New("default priority must be between 1 and 10")
	}

	if config.DefaultQPS <= 0 {
		return errors.New("default QPS must be greater than 0")
	}

	return nil
}

// UserRateLimitService user rate limit service
type UserRateLimitService struct{}

// GetUserRateLimit get user rate limit configuration
func (s *UserRateLimitService) GetUserRateLimit(userID string) (*UserRateLimit, error) {
	var rateLimit UserRateLimit
	err := DB.Where("user_id = ?", userID).First(&rateLimit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // user has no custom configuration
		}
		return nil, err
	}
	return &rateLimit, nil
}

// ListUserRateLimits get user rate limit configuration list
func (s *UserRateLimitService) ListUserRateLimits(page, pageSize int) ([]*UserRateLimit, int64, error) {
	var rateLimits []*UserRateLimit
	var total int64

	// calculate total
	err := DB.Model(&UserRateLimit{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// paginated query
	offset := (page - 1) * pageSize
	err = DB.Offset(offset).Limit(pageSize).Find(&rateLimits).Error
	if err != nil {
		return nil, 0, err
	}

	return rateLimits, total, nil
}

// CreateUserRateLimit create user rate limit configuration
func (s *UserRateLimitService) CreateUserRateLimit(rateLimit *UserRateLimit) error {
	// validate configuration
	if err := s.validateUserRateLimit(rateLimit); err != nil {
		return err
	}

	// check if user already has configuration
	var existing UserRateLimit
	err := DB.Where("user_id = ?", rateLimit.UserID).First(&existing).Error
	if err == nil {
		return errors.New("user rate limit already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return DB.Create(rateLimit).Error
}

// UpdateUserRateLimit update user rate limit configuration
func (s *UserRateLimitService) UpdateUserRateLimit(userID string, rateLimit *UserRateLimit) error {
	// validate configuration
	if err := s.validateUserRateLimit(rateLimit); err != nil {
		return err
	}

	rateLimit.UserID = userID

	var existing UserRateLimit
	err := DB.Where("user_id = ?", userID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DB.Create(rateLimit).Error
		}
		return err
	}

	rateLimit.ID = existing.ID
	return DB.Save(rateLimit).Error
}

// DeleteUserRateLimit delete user rate limit configuration
func (s *UserRateLimitService) DeleteUserRateLimit(userID string) error {
	result := DB.Where("user_id = ?", userID).Delete(&UserRateLimit{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user rate limit not found")
	}

	return nil
}

// validateUserRateLimit validate user rate limit configuration
func (s *UserRateLimitService) validateUserRateLimit(rateLimit *UserRateLimit) error {
	if rateLimit.UserID == "" {
		return errors.New("user ID is required")
	}

	if rateLimit.Priority != nil && (*rateLimit.Priority < 1 || *rateLimit.Priority > 10) {
		return errors.New("priority must be between 1 and 10")
	}

	if rateLimit.QPS != nil && *rateLimit.QPS <= 0 {
		return errors.New("QPS must be greater than 0")
	}

	return nil
}

// AgentService agent service
type AgentService struct{}

// GetAgentByAgentID get agent by agent ID
func (s *AgentService) GetAgentByAgentID(agentID string) (*Agent, error) {
	var agent Agent
	err := DB.Where("agent_id = ? AND deleted_at IS NULL", agentID).First(&agent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}
	return &agent, nil
}

// generateAgentID generate agent ID
func (s *AgentService) generateAgentID() string {
	return "agent_" + generateRandomString(12)
}

// generateConnectorAPIKey generate connector API key
func (s *AgentService) generateConnectorAPIKey() string {
	return "sk-conn_" + generateRandomString(32)
}

// generateRandomString generate random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1) // prevent timestamp duplication
	}
	return string(result)
}

// GetAgent get agent
func (s *AgentService) GetAgent(id uint) (*Agent, error) {
	var agent Agent
	err := DB.First(&agent, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("agent not found")
		}
		return nil, err
	}
	return &agent, nil
}

// ListAgents get agent list
func (s *AgentService) ListAgents(page, pageSize int, agentType string) ([]*Agent, int64, error) {
	var agents []*Agent
	var total int64

	query := DB.Model(&Agent{})
	if agentType != "" {
		query = query.Where("type = ?", agentType)
	}

	// calculate total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// paginated query
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&agents).Error
	if err != nil {
		return nil, 0, err
	}

	return agents, total, nil
}

// CreateAgent create agent
func (s *AgentService) CreateAgent(agent *Agent) error {
	// validate agent configuration
	if err := s.validateAgent(agent); err != nil {
		return err
	}

	// automatically generate agent ID and connector API key
	agent.AgentID = s.generateAgentID()
	agent.ConnectorAPIKey = s.generateConnectorAPIKey()

	return DB.Create(agent).Error
}

// UpdateAgent update agent
func (s *AgentService) UpdateAgent(id uint, agent *Agent) error {
	// validate agent configuration
	if err := s.validateAgent(agent); err != nil {
		return err
	}

	var existing Agent
	err := DB.First(&existing, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("agent not found")
		}
		return err
	}

	agent.ID = id
	return DB.Save(agent).Error
}

// DeleteAgent delete agent (soft delete)
func (s *AgentService) DeleteAgent(id uint) error {
	result := DB.Delete(&Agent{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("agent not found")
	}

	return nil
}

// validateAgent validate agent configuration
func (s *AgentService) validateAgent(agent *Agent) error {
	if agent.Name == "" {
		return errors.New("agent name is required")
	}

	if agent.Type != AgentTypeDify && agent.Type != AgentTypeOpenAI && agent.Type != AgentTypeOpenAICompatible {
		return errors.New("invalid agent type")
	}

	if agent.URL == "" {
		return errors.New("agent URL is required")
	}

	if agent.SourceAPIKey == "" {
		return errors.New("agent source API key is required")
	}

	if agent.QPS <= 0 {
		return errors.New("agent QPS must be greater than 0")
	}

	return nil
}
