package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"agent-connector/pkg/agent"
)

func main() {
	fmt.Println("=== Agent Management Demo ===")

	// Check for required environment variables
	openaiKey := os.Getenv("OPENAI_API_KEY")
	difyKey := os.Getenv("DIFY_API_KEY")
	difyBaseURL := os.Getenv("DIFY_BASE_URL")
	difyAppID := os.Getenv("DIFY_APP_ID")

	if openaiKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set, OpenAI demos will be skipped")
	}

	if difyKey == "" || difyBaseURL == "" || difyAppID == "" {
		log.Println("Warning: Dify environment variables not set, Dify demos will be skipped")
	}

	ctx := context.Background()

	// Demo 1: Basic Agent Creation
	fmt.Println("\n1. Basic Agent Creation")
	basicAgentCreationDemo(openaiKey, difyKey, difyBaseURL, difyAppID)

	// Demo 2: Agent Manager
	fmt.Println("\n2. Agent Manager Demo")
	agentManagerDemo(ctx, openaiKey, difyKey, difyBaseURL, difyAppID)

	// Demo 3: Configuration Builders
	fmt.Println("\n3. Configuration Builders Demo")
	configurationBuildersDemo()

	// Demo 4: Preset Configurations
	fmt.Println("\n4. Preset Configurations Demo")
	presetConfigurationsDemo(openaiKey, difyKey, difyBaseURL, difyAppID)

	// Demo 5: Load Balancing Strategies
	fmt.Println("\n5. Load Balancing Demo")
	loadBalancingDemo(ctx, openaiKey, difyKey, difyBaseURL, difyAppID)

	fmt.Println("\n=== Demo Completed ===")
}

func basicAgentCreationDemo(openaiKey, difyKey, difyBaseURL, difyAppID string) {
	factory := agent.NewAgentFactory()

	// Create OpenAI agent if key is available
	if openaiKey != "" {
		fmt.Println("Creating OpenAI agent...")
		openaiConfig := &agent.OpenAIConfig{
			AgentConfig: agent.AgentConfig{
				ID:   "demo-openai",
				Name: "Demo OpenAI Agent",
				Type: agent.AgentTypeOpenAI,
			},
			BaseURL:      "https://api.openai.com",
			APIKey:       openaiKey,
			DefaultModel: "gpt-3.5-turbo",
		}

		openaiAgent, err := factory.CreateOpenAIAgent(openaiConfig)
		if err != nil {
			log.Printf("Failed to create OpenAI agent: %v", err)
		} else {
			fmt.Printf("✓ Created OpenAI agent: %s (%s)\n",
				openaiAgent.GetName(), openaiAgent.GetID())

			// Get capabilities
			capabilities := openaiAgent.GetCapabilities()
			fmt.Printf("  Capabilities: Chat=%v, Streaming=%v, Functions=%v\n",
				capabilities.SupportsChatCompletion,
				capabilities.SupportsStreaming,
				capabilities.SupportsFunctionCalling)

			openaiAgent.Close()
		}
	}

	// Create Dify agent if configuration is available
	if difyKey != "" && difyBaseURL != "" && difyAppID != "" {
		fmt.Println("Creating Dify agent...")
		difyConfig := &agent.DifyConfig{
			AgentConfig: agent.AgentConfig{
				ID:   "demo-dify",
				Name: "Demo Dify Agent",
				Type: agent.AgentTypeDify,
			},
			BaseURL: difyBaseURL,
			APIKey:  difyKey,
			AppID:   difyAppID,
			AppType: "chatbot",
		}

		difyAgent, err := factory.CreateDifyAgent(difyConfig)
		if err != nil {
			log.Printf("Failed to create Dify agent: %v", err)
		} else {
			fmt.Printf("✓ Created Dify agent: %s (%s)\n",
				difyAgent.GetName(), difyAgent.GetID())

			// Get capabilities
			capabilities := difyAgent.GetCapabilities()
			fmt.Printf("  Capabilities: Chat=%v, Streaming=%v, Files=%v\n",
				capabilities.SupportsChatCompletion,
				capabilities.SupportsStreaming,
				capabilities.SupportsFiles)

			difyAgent.Close()
		}
	}
}

func agentManagerDemo(ctx context.Context, openaiKey, difyKey, difyBaseURL, difyAppID string) {
	// Create agent manager
	config := &agent.AgentManagerConfig{
		LoadBalancingStrategy: agent.Priority,
		EnableHealthChecks:    true,
		HealthCheckInterval:   30 * time.Second,
		DefaultTimeout:        10 * time.Second,
		MaxRetries:            3,
		EnableMetrics:         true,
	}

	manager, err := agent.NewAgentManager(config)
	if err != nil {
		log.Printf("Failed to create agent manager: %v", err)
		return
	}
	defer manager.Close()

	fmt.Println("✓ Created agent manager")

	// Register agents
	agentCount := 0

	if openaiKey != "" {
		openaiConfig := agent.NewOpenAIConfigBuilder().
			WithID("managed-openai").
			WithName("Managed OpenAI Agent").
			WithBaseURL("https://api.openai.com").
			WithAPIKey(openaiKey).
			WithPriority(100).
			Build()

		openaiAgent, err := agent.NewOpenAIAgent(openaiConfig)
		if err == nil {
			if err := manager.RegisterAgent(openaiAgent); err != nil {
				log.Printf("Failed to register OpenAI agent: %v", err)
			} else {
				fmt.Println("✓ Registered OpenAI agent")
				agentCount++
			}
		}
	}

	if difyKey != "" && difyBaseURL != "" && difyAppID != "" {
		difyConfig := agent.NewDifyConfigBuilder().
			WithID("managed-dify").
			WithName("Managed Dify Agent").
			WithBaseURL(difyBaseURL).
			WithAPIKey(difyKey).
			WithAppID(difyAppID).
			WithPriority(80).
			Build()

		difyAgent, err := agent.NewDifyAgent(difyConfig)
		if err == nil {
			if err := manager.RegisterAgent(difyAgent); err != nil {
				log.Printf("Failed to register Dify agent: %v", err)
			} else {
				fmt.Println("✓ Registered Dify agent")
				agentCount++
			}
		}
	}

	if agentCount == 0 {
		fmt.Println("No agents registered (missing API keys)")
		return
	}

	// List all agents
	agents := manager.ListAgents()
	fmt.Printf("✓ Total agents registered: %d\n", len(agents))

	for _, ag := range agents {
		fmt.Printf("  - %s (%s): %s\n", ag.GetName(), ag.GetID(), ag.GetType())
	}

	// Get agent info
	if len(agents) > 0 {
		firstAgent := agents[0]
		info, err := manager.GetAgentInfo(ctx, firstAgent.GetID())
		if err != nil {
			log.Printf("Failed to get agent info: %v", err)
		} else {
			fmt.Printf("✓ Agent info for %s:\n", info.Name)
			fmt.Printf("  Status: %s (Health: %v)\n", info.Status.Status, info.Status.Health)
			fmt.Printf("  Response Time: %dms\n", info.Status.ResponseTime)
		}
	}

	// Test load balancing
	request := &agent.ChatRequest{
		Messages: []agent.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
	}

	selectedAgent, err := manager.GetAvailableAgent(ctx, request)
	if err != nil {
		log.Printf("Failed to get available agent: %v", err)
	} else {
		fmt.Printf("✓ Load balancer selected: %s\n", selectedAgent.GetName())
	}
}

func configurationBuildersDemo() {
	fmt.Println("Building OpenAI configuration with fluent interface...")

	openaiConfig := agent.NewOpenAIConfigBuilder().
		WithID("builder-openai").
		WithName("Builder OpenAI Agent").
		WithBaseURL("https://api.openai.com").
		WithAPIKey("test-key").
		WithDefaultModel("gpt-4").
		WithMaxTokens(8192).
		WithTemperature(0.5).
		WithPriority(90).
		WithTimeout(60 * time.Second).
		WithCustomHeaders(map[string]string{
			"X-Custom-Header": "demo-value",
		}).
		Build()

	fmt.Printf("✓ Built OpenAI config: %s (Model: %s, MaxTokens: %d)\n",
		openaiConfig.Name, openaiConfig.DefaultModel, openaiConfig.MaxTokens)

	fmt.Println("Building Dify configuration...")

	difyConfig := agent.NewDifyConfigBuilder().
		WithID("builder-dify").
		WithName("Builder Dify Agent").
		WithBaseURL("https://api.dify.ai").
		WithAPIKey("test-key").
		WithAppID("test-app").
		WithAppType("agent").
		WithVersion("v1").
		WithPriority(70).
		WithLogging(true).
		WithAutoGenerateTitle(false).
		Build()

	fmt.Printf("✓ Built Dify config: %s (AppType: %s, Version: %s)\n",
		difyConfig.Name, difyConfig.AppType, difyConfig.Version)

	fmt.Println("Building retry policy...")

	retryPolicy := agent.NewRetryPolicyBuilder().
		WithMaxRetries(5).
		WithInitialDelay(2 * time.Second).
		WithMaxDelay(60 * time.Second).
		WithMultiplier(2.0).
		WithRetryableErrors([]string{"timeout", "rate_limit", "service_unavailable"}).
		Build()

	fmt.Printf("✓ Built retry policy: MaxRetries=%d, InitialDelay=%v\n",
		retryPolicy.MaxRetries, retryPolicy.InitialDelay)

	fmt.Println("Building health check configuration...")

	healthCheck := agent.NewHealthCheckConfigBuilder().
		WithEnabled(true).
		WithInterval(30 * time.Second).
		WithTimeout(10 * time.Second).
		WithFailureThreshold(3).
		WithSuccessThreshold(1).
		Build()

	fmt.Printf("✓ Built health check config: Interval=%v, Timeout=%v\n",
		healthCheck.Interval, healthCheck.Timeout)
}

func presetConfigurationsDemo(openaiKey, difyKey, difyBaseURL, difyAppID string) {
	presets := agent.NewPresetConfigs()

	if openaiKey != "" {
		fmt.Println("Using OpenAI presets...")

		// GPT-3.5-turbo preset
		gpt35Config := presets.OpenAIGPT35Turbo("preset-gpt35", "Preset GPT-3.5", openaiKey)
		fmt.Printf("✓ GPT-3.5 preset: %s (Model: %s, MaxTokens: %d)\n",
			gpt35Config.Name, gpt35Config.DefaultModel, gpt35Config.MaxTokens)

		// GPT-4 preset
		gpt4Config := presets.OpenAIGPT4("preset-gpt4", "Preset GPT-4", openaiKey)
		fmt.Printf("✓ GPT-4 preset: %s (Model: %s, MaxTokens: %d)\n",
			gpt4Config.Name, gpt4Config.DefaultModel, gpt4Config.MaxTokens)

		// Azure OpenAI preset (example)
		azureConfig := presets.AzureOpenAI(
			"preset-azure",
			"Preset Azure OpenAI",
			"https://your-resource.openai.azure.com",
			"azure-key",
			"gpt-35-turbo",
		)
		fmt.Printf("✓ Azure preset: %s (BaseURL: %s)\n",
			azureConfig.Name, azureConfig.BaseURL)
		fmt.Printf("  Custom headers: %v\n", azureConfig.CustomHeaders)
	}

	if difyKey != "" && difyBaseURL != "" && difyAppID != "" {
		fmt.Println("Using Dify presets...")

		// Dify chatbot preset
		chatbotConfig := presets.DifyChatbot("preset-chatbot", "Preset Dify Chatbot", difyBaseURL, difyKey, difyAppID)
		fmt.Printf("✓ Dify chatbot preset: %s (AppType: %s)\n",
			chatbotConfig.Name, chatbotConfig.AppType)

		// Dify agent preset
		agentConfig := presets.DifyAgent("preset-agent", "Preset Dify Agent", difyBaseURL, difyKey, difyAppID)
		fmt.Printf("✓ Dify agent preset: %s (AppType: %s)\n",
			agentConfig.Name, agentConfig.AppType)
	}
}

func loadBalancingDemo(ctx context.Context, openaiKey, difyKey, difyBaseURL, difyAppID string) {
	strategies := []agent.LoadBalancingStrategy{
		agent.Priority,
		agent.RoundRobin,
		agent.Random,
		agent.WeightedRandom,
	}

	for _, strategy := range strategies {
		fmt.Printf("Testing %s strategy...\n", strategy)

		config := &agent.AgentManagerConfig{
			LoadBalancingStrategy: strategy,
			EnableHealthChecks:    false, // Disable for demo
			DefaultTimeout:        5 * time.Second,
		}

		manager, err := agent.NewAgentManager(config)
		if err != nil {
			log.Printf("Failed to create manager: %v", err)
			continue
		}

		// Register multiple agents with different priorities
		agentCount := 0

		if openaiKey != "" {
			// High priority OpenAI agent
			highPriorityConfig := agent.NewOpenAIConfigBuilder().
				WithID("high-priority").
				WithName("High Priority Agent").
				WithBaseURL("https://api.openai.com").
				WithAPIKey(openaiKey).
				WithPriority(100).
				Build()

			agent1, err := agent.NewOpenAIAgent(highPriorityConfig)
			if err == nil {
				manager.RegisterAgent(agent1)
				agentCount++
			}

			// Medium priority OpenAI agent
			mediumPriorityConfig := agent.NewOpenAIConfigBuilder().
				WithID("medium-priority").
				WithName("Medium Priority Agent").
				WithBaseURL("https://api.openai.com").
				WithAPIKey(openaiKey).
				WithPriority(50).
				Build()

			agent2, err := agent.NewOpenAIAgent(mediumPriorityConfig)
			if err == nil {
				manager.RegisterAgent(agent2)
				agentCount++
			}
		}

		if difyKey != "" && difyBaseURL != "" && difyAppID != "" {
			// Low priority Dify agent
			lowPriorityConfig := agent.NewDifyConfigBuilder().
				WithID("low-priority").
				WithName("Low Priority Agent").
				WithBaseURL(difyBaseURL).
				WithAPIKey(difyKey).
				WithAppID(difyAppID).
				WithPriority(20).
				Build()

			agent3, err := agent.NewDifyAgent(lowPriorityConfig)
			if err == nil {
				manager.RegisterAgent(agent3)
				agentCount++
			}
		}

		if agentCount == 0 {
			fmt.Println("  No agents available for testing")
			manager.Close()
			continue
		}

		// Test multiple selections
		request := &agent.ChatRequest{
			Messages: []agent.Message{
				{Role: "user", Content: "Test message"},
			},
		}

		selections := make(map[string]int)

		for i := 0; i < 10; i++ {
			selectedAgent, err := manager.GetAvailableAgent(ctx, request)
			if err != nil {
				continue
			}
			selections[selectedAgent.GetName()]++
		}

		fmt.Printf("  Selection results: %v\n", selections)
		manager.Close()
	}
}

func init() {
	// Set log flags for better output
	log.SetFlags(log.Ltime | log.Lshortfile)
}
