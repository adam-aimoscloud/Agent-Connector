package agent

import (
	"fmt"
	"time"
)

// AgentFactory provides methods to create different types of agents
type AgentFactory struct{}

// NewAgentFactory creates a new agent factory
func NewAgentFactory() *AgentFactory {
	return &AgentFactory{}
}

// CreateAgent creates an agent based on the provided configuration
func (f *AgentFactory) CreateAgent(agentType AgentType, config interface{}) (Agent, error) {
	if !agentType.IsValid() {
		return nil, fmt.Errorf("invalid agent type: %s", agentType)
	}

	switch agentType {
	case AgentTypeOpenAI:
		openaiConfig, ok := config.(*OpenAIConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for OpenAI agent, expected *OpenAIConfig")
		}
		return NewOpenAIAgent(openaiConfig)

	case AgentTypeDify:
		difyConfig, ok := config.(*DifyConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Dify agent, expected *DifyConfig")
		}
		return NewDifyAgent(difyConfig)

	default:
		return nil, fmt.Errorf("unsupported agent type: %s", agentType)
	}
}

// CreateOpenAIAgent creates an OpenAI compatible agent
func (f *AgentFactory) CreateOpenAIAgent(config *OpenAIConfig) (*OpenAIAgent, error) {
	return NewOpenAIAgent(config)
}

// CreateDifyAgent creates a Dify agent
func (f *AgentFactory) CreateDifyAgent(config *DifyConfig) (*DifyAgent, error) {
	return NewDifyAgent(config)
}

// OpenAIConfigBuilder provides a fluent interface for building OpenAI configurations
type OpenAIConfigBuilder struct {
	config *OpenAIConfig
}

// NewOpenAIConfigBuilder creates a new OpenAI config builder
func NewOpenAIConfigBuilder() *OpenAIConfigBuilder {
	return &OpenAIConfigBuilder{
		config: &OpenAIConfig{
			AgentConfig: AgentConfig{
				Type:                  AgentTypeOpenAI,
				Enabled:               true,
				Priority:              50,
				Timeout:               DefaultTimeout,
				MaxConcurrentRequests: DefaultMaxConcurrentRequests,
			},
			Temperature: 0.7,
			MaxTokens:   4096,
		},
	}
}

// WithID sets the agent ID
func (b *OpenAIConfigBuilder) WithID(id string) *OpenAIConfigBuilder {
	b.config.ID = id
	return b
}

// WithName sets the agent name
func (b *OpenAIConfigBuilder) WithName(name string) *OpenAIConfigBuilder {
	b.config.Name = name
	return b
}

// WithBaseURL sets the base URL
func (b *OpenAIConfigBuilder) WithBaseURL(baseURL string) *OpenAIConfigBuilder {
	b.config.BaseURL = baseURL
	return b
}

// WithAPIKey sets the API key
func (b *OpenAIConfigBuilder) WithAPIKey(apiKey string) *OpenAIConfigBuilder {
	b.config.APIKey = apiKey
	return b
}

// WithOrganization sets the organization ID
func (b *OpenAIConfigBuilder) WithOrganization(organization string) *OpenAIConfigBuilder {
	b.config.Organization = organization
	return b
}

// WithDefaultModel sets the default model
func (b *OpenAIConfigBuilder) WithDefaultModel(model string) *OpenAIConfigBuilder {
	b.config.DefaultModel = model
	return b
}

// WithSupportedModels sets the supported models
func (b *OpenAIConfigBuilder) WithSupportedModels(models []string) *OpenAIConfigBuilder {
	b.config.SupportedModels = models
	return b
}

// WithMaxTokens sets the maximum tokens
func (b *OpenAIConfigBuilder) WithMaxTokens(maxTokens int) *OpenAIConfigBuilder {
	b.config.MaxTokens = maxTokens
	return b
}

// WithTemperature sets the temperature
func (b *OpenAIConfigBuilder) WithTemperature(temperature float32) *OpenAIConfigBuilder {
	b.config.Temperature = temperature
	return b
}

// WithTimeout sets the request timeout
func (b *OpenAIConfigBuilder) WithTimeout(timeout time.Duration) *OpenAIConfigBuilder {
	b.config.Timeout = timeout
	return b
}

// WithPriority sets the agent priority
func (b *OpenAIConfigBuilder) WithPriority(priority int) *OpenAIConfigBuilder {
	b.config.Priority = priority
	return b
}

// WithMaxConcurrentRequests sets the maximum concurrent requests
func (b *OpenAIConfigBuilder) WithMaxConcurrentRequests(maxRequests int) *OpenAIConfigBuilder {
	b.config.MaxConcurrentRequests = maxRequests
	return b
}

// WithCustomHeaders sets custom HTTP headers
func (b *OpenAIConfigBuilder) WithCustomHeaders(headers map[string]string) *OpenAIConfigBuilder {
	b.config.CustomHeaders = headers
	return b
}

// WithRetryPolicy sets the retry policy
func (b *OpenAIConfigBuilder) WithRetryPolicy(policy *RetryPolicy) *OpenAIConfigBuilder {
	b.config.RetryPolicy = policy
	return b
}

// WithHealthCheck sets the health check configuration
func (b *OpenAIConfigBuilder) WithHealthCheck(healthCheck *HealthCheckConfig) *OpenAIConfigBuilder {
	b.config.HealthCheck = healthCheck
	return b
}

// Enabled sets whether the agent is enabled
func (b *OpenAIConfigBuilder) Enabled(enabled bool) *OpenAIConfigBuilder {
	b.config.Enabled = enabled
	return b
}

// Build builds the OpenAI configuration
func (b *OpenAIConfigBuilder) Build() *OpenAIConfig {
	return b.config
}

// DifyConfigBuilder provides a fluent interface for building Dify configurations
type DifyConfigBuilder struct {
	config *DifyConfig
}

// NewDifyConfigBuilder creates a new Dify config builder
func NewDifyConfigBuilder() *DifyConfigBuilder {
	return &DifyConfigBuilder{
		config: &DifyConfig{
			AgentConfig: AgentConfig{
				Type:                  AgentTypeDify,
				Enabled:               true,
				Priority:              50,
				Timeout:               DefaultTimeout,
				MaxConcurrentRequests: DefaultMaxConcurrentRequests,
			},
			AppType:           "chatbot",
			Version:           "v1",
			EnableLogging:     true,
			AutoGenerateTitle: true,
		},
	}
}

// WithID sets the agent ID
func (b *DifyConfigBuilder) WithID(id string) *DifyConfigBuilder {
	b.config.ID = id
	return b
}

// WithName sets the agent name
func (b *DifyConfigBuilder) WithName(name string) *DifyConfigBuilder {
	b.config.Name = name
	return b
}

// WithBaseURL sets the base URL
func (b *DifyConfigBuilder) WithBaseURL(baseURL string) *DifyConfigBuilder {
	b.config.BaseURL = baseURL
	return b
}

// WithAPIKey sets the API key
func (b *DifyConfigBuilder) WithAPIKey(apiKey string) *DifyConfigBuilder {
	b.config.APIKey = apiKey
	return b
}

// WithAppID sets the app ID
func (b *DifyConfigBuilder) WithAppID(appID string) *DifyConfigBuilder {
	b.config.AppID = appID
	return b
}

// WithAppType sets the app type
func (b *DifyConfigBuilder) WithAppType(appType string) *DifyConfigBuilder {
	b.config.AppType = appType
	return b
}

// WithVersion sets the API version
func (b *DifyConfigBuilder) WithVersion(version string) *DifyConfigBuilder {
	b.config.Version = version
	return b
}

// WithLogging sets logging enabled/disabled
func (b *DifyConfigBuilder) WithLogging(enabled bool) *DifyConfigBuilder {
	b.config.EnableLogging = enabled
	return b
}

// WithAutoGenerateTitle sets auto title generation
func (b *DifyConfigBuilder) WithAutoGenerateTitle(enabled bool) *DifyConfigBuilder {
	b.config.AutoGenerateTitle = enabled
	return b
}

// WithTimeout sets the request timeout
func (b *DifyConfigBuilder) WithTimeout(timeout time.Duration) *DifyConfigBuilder {
	b.config.Timeout = timeout
	return b
}

// WithPriority sets the agent priority
func (b *DifyConfigBuilder) WithPriority(priority int) *DifyConfigBuilder {
	b.config.Priority = priority
	return b
}

// WithMaxConcurrentRequests sets the maximum concurrent requests
func (b *DifyConfigBuilder) WithMaxConcurrentRequests(maxRequests int) *DifyConfigBuilder {
	b.config.MaxConcurrentRequests = maxRequests
	return b
}

// WithCustomHeaders sets custom HTTP headers
func (b *DifyConfigBuilder) WithCustomHeaders(headers map[string]string) *DifyConfigBuilder {
	b.config.CustomHeaders = headers
	return b
}

// WithRetryPolicy sets the retry policy
func (b *DifyConfigBuilder) WithRetryPolicy(policy *RetryPolicy) *DifyConfigBuilder {
	b.config.RetryPolicy = policy
	return b
}

// WithHealthCheck sets the health check configuration
func (b *DifyConfigBuilder) WithHealthCheck(healthCheck *HealthCheckConfig) *DifyConfigBuilder {
	b.config.HealthCheck = healthCheck
	return b
}

// Enabled sets whether the agent is enabled
func (b *DifyConfigBuilder) Enabled(enabled bool) *DifyConfigBuilder {
	b.config.Enabled = enabled
	return b
}

// Build builds the Dify configuration
func (b *DifyConfigBuilder) Build() *DifyConfig {
	return b.config
}

// RetryPolicyBuilder provides a fluent interface for building retry policies
type RetryPolicyBuilder struct {
	policy *RetryPolicy
}

// NewRetryPolicyBuilder creates a new retry policy builder
func NewRetryPolicyBuilder() *RetryPolicyBuilder {
	return &RetryPolicyBuilder{
		policy: &RetryPolicy{
			MaxRetries:   3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}
}

// WithMaxRetries sets the maximum number of retries
func (b *RetryPolicyBuilder) WithMaxRetries(maxRetries int) *RetryPolicyBuilder {
	b.policy.MaxRetries = maxRetries
	return b
}

// WithInitialDelay sets the initial delay between retries
func (b *RetryPolicyBuilder) WithInitialDelay(delay time.Duration) *RetryPolicyBuilder {
	b.policy.InitialDelay = delay
	return b
}

// WithMaxDelay sets the maximum delay between retries
func (b *RetryPolicyBuilder) WithMaxDelay(delay time.Duration) *RetryPolicyBuilder {
	b.policy.MaxDelay = delay
	return b
}

// WithMultiplier sets the exponential backoff multiplier
func (b *RetryPolicyBuilder) WithMultiplier(multiplier float64) *RetryPolicyBuilder {
	b.policy.Multiplier = multiplier
	return b
}

// WithRetryableErrors sets the retryable error codes
func (b *RetryPolicyBuilder) WithRetryableErrors(errors []string) *RetryPolicyBuilder {
	b.policy.RetryableErrors = errors
	return b
}

// Build builds the retry policy
func (b *RetryPolicyBuilder) Build() *RetryPolicy {
	return b.policy
}

// HealthCheckConfigBuilder provides a fluent interface for building health check configurations
type HealthCheckConfigBuilder struct {
	config *HealthCheckConfig
}

// NewHealthCheckConfigBuilder creates a new health check config builder
func NewHealthCheckConfigBuilder() *HealthCheckConfigBuilder {
	return &HealthCheckConfigBuilder{
		config: &HealthCheckConfig{
			Enabled:          true,
			Interval:         1 * time.Minute,
			Timeout:          10 * time.Second,
			FailureThreshold: 3,
			SuccessThreshold: 1,
		},
	}
}

// WithEnabled sets whether health checks are enabled
func (b *HealthCheckConfigBuilder) WithEnabled(enabled bool) *HealthCheckConfigBuilder {
	b.config.Enabled = enabled
	return b
}

// WithInterval sets the health check interval
func (b *HealthCheckConfigBuilder) WithInterval(interval time.Duration) *HealthCheckConfigBuilder {
	b.config.Interval = interval
	return b
}

// WithTimeout sets the health check timeout
func (b *HealthCheckConfigBuilder) WithTimeout(timeout time.Duration) *HealthCheckConfigBuilder {
	b.config.Timeout = timeout
	return b
}

// WithFailureThreshold sets the failure threshold
func (b *HealthCheckConfigBuilder) WithFailureThreshold(threshold int) *HealthCheckConfigBuilder {
	b.config.FailureThreshold = threshold
	return b
}

// WithSuccessThreshold sets the success threshold
func (b *HealthCheckConfigBuilder) WithSuccessThreshold(threshold int) *HealthCheckConfigBuilder {
	b.config.SuccessThreshold = threshold
	return b
}

// Build builds the health check configuration
func (b *HealthCheckConfigBuilder) Build() *HealthCheckConfig {
	return b.config
}

// ConfigValidator provides validation for agent configurations
type ConfigValidator struct{}

// NewConfigValidator creates a new config validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateOpenAIConfig validates an OpenAI configuration
func (v *ConfigValidator) ValidateOpenAIConfig(config *OpenAIConfig) error {
	return validateOpenAIConfig(config)
}

// ValidateDifyConfig validates a Dify configuration
func (v *ConfigValidator) ValidateDifyConfig(config *DifyConfig) error {
	return validateDifyConfig(config)
}

// ValidateAgentManagerConfig validates an agent manager configuration
func (v *ConfigValidator) ValidateAgentManagerConfig(config *AgentManagerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	return nil
}

// PresetConfigs provides common preset configurations
type PresetConfigs struct{}

// NewPresetConfigs creates a new preset configs provider
func NewPresetConfigs() *PresetConfigs {
	return &PresetConfigs{}
}

// OpenAIGPT35Turbo returns a preset configuration for OpenAI GPT-3.5-turbo
func (p *PresetConfigs) OpenAIGPT35Turbo(id, name, apiKey string) *OpenAIConfig {
	return NewOpenAIConfigBuilder().
		WithID(id).
		WithName(name).
		WithBaseURL("https://api.openai.com").
		WithAPIKey(apiKey).
		WithDefaultModel("gpt-3.5-turbo").
		WithMaxTokens(4096).
		WithTemperature(0.7).
		Build()
}

// OpenAIGPT4 returns a preset configuration for OpenAI GPT-4
func (p *PresetConfigs) OpenAIGPT4(id, name, apiKey string) *OpenAIConfig {
	return NewOpenAIConfigBuilder().
		WithID(id).
		WithName(name).
		WithBaseURL("https://api.openai.com").
		WithAPIKey(apiKey).
		WithDefaultModel("gpt-4").
		WithMaxTokens(8192).
		WithTemperature(0.7).
		Build()
}

// AzureOpenAI returns a preset configuration for Azure OpenAI
func (p *PresetConfigs) AzureOpenAI(id, name, baseURL, apiKey, deploymentName string) *OpenAIConfig {
	return NewOpenAIConfigBuilder().
		WithID(id).
		WithName(name).
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithDefaultModel(deploymentName).
		WithMaxTokens(4096).
		WithTemperature(0.7).
		WithCustomHeaders(map[string]string{
			"api-version": "2023-12-01-preview",
		}).
		Build()
}

// DifyChatbot returns a preset configuration for Dify chatbot
func (p *PresetConfigs) DifyChatbot(id, name, baseURL, apiKey, appID string) *DifyConfig {
	return NewDifyConfigBuilder().
		WithID(id).
		WithName(name).
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithAppID(appID).
		WithAppType("chatbot").
		WithVersion("v1").
		Build()
}

// DifyAgent returns a preset configuration for Dify agent
func (p *PresetConfigs) DifyAgent(id, name, baseURL, apiKey, appID string) *DifyConfig {
	return NewDifyConfigBuilder().
		WithID(id).
		WithName(name).
		WithBaseURL(baseURL).
		WithAPIKey(apiKey).
		WithAppID(appID).
		WithAppType("agent").
		WithVersion("v1").
		Build()
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        AgentType              `json:"type"`
	Template    map[string]interface{} `json:"template"`
}

// TemplateManager manages configuration templates
type TemplateManager struct {
	templates map[string]*ConfigTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	manager := &TemplateManager{
		templates: make(map[string]*ConfigTemplate),
	}

	// Add default templates
	manager.addDefaultTemplates()

	return manager
}

// addDefaultTemplates adds default configuration templates
func (tm *TemplateManager) addDefaultTemplates() {
	// OpenAI templates
	tm.AddTemplate(&ConfigTemplate{
		Name:        "openai-gpt35",
		Description: "OpenAI GPT-3.5-turbo configuration template",
		Type:        AgentTypeOpenAI,
		Template: map[string]interface{}{
			"base_url":      "https://api.openai.com",
			"default_model": "gpt-3.5-turbo",
			"max_tokens":    4096,
			"temperature":   0.7,
		},
	})

	tm.AddTemplate(&ConfigTemplate{
		Name:        "openai-gpt4",
		Description: "OpenAI GPT-4 configuration template",
		Type:        AgentTypeOpenAI,
		Template: map[string]interface{}{
			"base_url":      "https://api.openai.com",
			"default_model": "gpt-4",
			"max_tokens":    8192,
			"temperature":   0.7,
		},
	})

	// Dify templates
	tm.AddTemplate(&ConfigTemplate{
		Name:        "dify-chatbot",
		Description: "Dify chatbot configuration template",
		Type:        AgentTypeDify,
		Template: map[string]interface{}{
			"app_type":            "chatbot",
			"version":             "v1",
			"enable_logging":      true,
			"auto_generate_title": true,
		},
	})

	tm.AddTemplate(&ConfigTemplate{
		Name:        "dify-agent",
		Description: "Dify agent configuration template",
		Type:        AgentTypeDify,
		Template: map[string]interface{}{
			"app_type":            "agent",
			"version":             "v1",
			"enable_logging":      true,
			"auto_generate_title": true,
		},
	})
}

// AddTemplate adds a configuration template
func (tm *TemplateManager) AddTemplate(template *ConfigTemplate) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if template.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if !template.Type.IsValid() {
		return fmt.Errorf("invalid template type: %s", template.Type)
	}

	tm.templates[template.Name] = template
	return nil
}

// GetTemplate retrieves a configuration template
func (tm *TemplateManager) GetTemplate(name string) (*ConfigTemplate, error) {
	template, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}

	return template, nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() []*ConfigTemplate {
	templates := make([]*ConfigTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}

	return templates
}

// ListTemplatesByType returns templates of a specific type
func (tm *TemplateManager) ListTemplatesByType(agentType AgentType) []*ConfigTemplate {
	var templates []*ConfigTemplate
	for _, template := range tm.templates {
		if template.Type == agentType {
			templates = append(templates, template)
		}
	}

	return templates
}
