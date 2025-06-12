package main

import (
	"agent-connector/api/dataflow"
	"agent-connector/config"
	"agent-connector/internal"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("🚀 Starting Data Flow API Server...")
	fmt.Println("===============================================")
	fmt.Printf("📊 Service: %s Data Flow API\n", cfg.App.Name)
	fmt.Printf("🌐 Purpose: Unified agent access for downstream applications\n")
	fmt.Printf("🔗 Server: %s\n", cfg.GetServiceAddr("data"))
	fmt.Printf("📝 Environment: %s\n", cfg.App.Environment)
	fmt.Printf("💾 Database: %s://%s:%d/%s\n", cfg.Database.Driver, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	fmt.Printf("📦 Redis: %s (DB: %d)\n", cfg.Redis.Addr, cfg.Redis.DB)
	fmt.Println("===============================================")

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database
	if err := internal.InitDatabase(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	fmt.Println("✅ Database initialized successfully")

	// Create Gin router
	router := gin.New()

	// Set data flow API routes and middlewares
	routerConfig := &dataflow.DataFlowRouterConfig{
		EnableCORS:         cfg.API.EnableCORS,
		EnableLogging:      true,
		EnableRecovery:     true,
		MaxRequestBodySize: cfg.API.MaxRequestBodySize,
		APIRateLimit:       cfg.Security.DefaultRateLimit,
	}

	dataflow.SetupDataFlowRoutesWithConfig(router, routerConfig)

	// Add root path information
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.App.Name + " Data Flow API",
			"version":     cfg.App.Version,
			"description": "Unified agent access platform for downstream applications",
			"environment": cfg.App.Environment,
			"endpoints": map[string]interface{}{
				"health":        "/api/v1/dataflow/health",
				"chat":          "/api/v1/dataflow/chat/:agent_id",
				"openai":        "/api/v1/dataflow/openai/chat/completions/:agent_id",
				"dify":          "/api/v1/dataflow/dify/chat-messages/:agent_id",
				"documentation": "https://docs.agent-connector.com/dataflow-api",
			},
			"authentication": map[string]string{
				"method":      "API Key + Agent ID",
				"header":      "Authorization: Bearer <api_key> or X-API-Key: <api_key>",
				"agent_id":    "Provided in URL path parameter",
				"description": "API keys are generated and managed by Agent-Connector platform",
			},
			"features": []string{
				"Multi-platform agent support (OpenAI, Dify, etc.)",
				"Streaming and blocking response modes",
				"Automatic format conversion",
				"Rate limiting and priority queuing",
				"Real-time request monitoring",
			},
			"status":    "running",
			"timestamp": time.Now().Unix(),
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServiceAddr("data"),
		Handler:      router,
		ReadTimeout:  cfg.Services.DataFlowAPI.ReadTimeout,
		WriteTimeout: cfg.Services.DataFlowAPI.WriteTimeout,
		IdleTimeout:  cfg.Services.DataFlowAPI.IdleTimeout,
	}

	// Gracefully shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		fmt.Println("\n🛑 Shutting down Data Flow API server...")

		// Give server 5 seconds to complete existing requests
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("❌ Server forced to shutdown: %v\n", err)
		} else {
			fmt.Println("✅ Data Flow API server gracefully stopped")
		}
	}()

	// Print API endpoints information
	printAPIEndpoints(cfg)

	// Start server
	fmt.Printf("🎯 Data Flow API server is running on http://%s\n", cfg.GetServiceAddr("data"))
	fmt.Println("📋 Ready to handle agent requests from downstream applications")
	fmt.Println("💡 Use Ctrl+C to gracefully shutdown the server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// printAPIEndpoints print API endpoints information
func printAPIEndpoints(cfg *config.Config) {
	fmt.Println("\n📡 Available API Endpoints:")
	fmt.Println("├── GET  /                                          - Service information")
	fmt.Println("├── GET  /api/v1/dataflow/health                    - Health check")
	fmt.Println("├── POST /api/v1/dataflow/chat/:agent_id            - Universal chat interface")
	fmt.Println("├── POST /api/v1/dataflow/openai/chat/completions/:agent_id  - OpenAI compatible")
	fmt.Println("└── POST /api/v1/dataflow/dify/chat-messages/:agent_id       - Dify compatible")

	fmt.Println("\n🔐 Authentication:")
	fmt.Println("├── Header: Authorization: Bearer <api_key>")
	fmt.Println("├── Header: X-API-Key: <api_key>")
	fmt.Println("└── Path Parameter: agent_id (generated by Agent-Connector)")

	fmt.Println("\n🌟 Features:")
	fmt.Println("├── ✨ Multi-platform agent support")
	fmt.Println("├── 🔄 Streaming and blocking responses")
	fmt.Println("├── 🔄 Automatic format conversion")
	fmt.Println("├── ⚡ Rate limiting and priority queuing")
	fmt.Println("└── 📊 Real-time monitoring")

	fmt.Println("\n📖 Usage Examples:")
	fmt.Println("# OpenAI-style request:")
	fmt.Printf("curl -X POST http://%s/api/v1/dataflow/openai/chat/completions/your-agent-id \\\n", cfg.GetServiceAddr("data"))
	fmt.Println("  -H \"Authorization: Bearer your-api-key\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}], \"model\": \"gpt-3.5-turbo\"}'")

	fmt.Println("\n# Dify-style request:")
	fmt.Printf("curl -X POST http://%s/api/v1/dataflow/dify/chat-messages/your-agent-id \\\n", cfg.GetServiceAddr("data"))
	fmt.Println("  -H \"Authorization: Bearer your-api-key\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"query\": \"Hello!\", \"user\": \"user123\"}'")
	fmt.Println()
}
