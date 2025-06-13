package main

import (
	"agent-connector/api/dataflow"
	"agent-connector/config"
	"agent-connector/internal"
	"agent-connector/pkg/ratelimiter"
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
	fmt.Printf("📊 Service: %s Data Flow API (New Backend Architecture)\n", cfg.App.Name)
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

	// Initialize Redis rate limiter
	rateLimiterConfig := &ratelimiter.Config{
		Rate:  float64(cfg.Security.DefaultRateLimit),
		Burst: cfg.Security.DefaultRateLimit * 2,
		Redis: &ratelimiter.RedisConfig{
			Addr:            cfg.Redis.Addr,
			Password:        cfg.Redis.Password,
			DB:              cfg.Redis.DB,
			PoolSize:        10,
			MinIdleConns:    2,
			ConnMaxIdleTime: 30 * time.Minute,
		},
	}

	redisRateLimiter, err := ratelimiter.NewRedisRateLimiter(rateLimiterConfig)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Redis rate limiter: %v", err)
	}
	fmt.Println("✅ Redis rate limiter initialized successfully")

	// Create Gin router
	router := gin.New()

	// Setup middlewares
	setupMiddlewares(router, cfg)

	// Setup new Backend routes
	dataflow.SetupBackendRoutes(router, redisRateLimiter)
	fmt.Println("✅ New Backend architecture routes initialized")

	// Setup legacy routes for backward compatibility
	dataflow.SetupLegacyRoutes(router, redisRateLimiter)
	fmt.Println("✅ Legacy routes initialized for backward compatibility")

	// Add root path information
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":      cfg.App.Name + " Data Flow API",
			"version":      cfg.App.Version,
			"description":  "Unified agent access platform with Backend architecture",
			"environment":  cfg.App.Environment,
			"architecture": "Backend-based with OpenAI, Dify Chat, and Dify Workflow support",
			"endpoints": map[string]interface{}{
				"health":        "/api/v1/health",
				"openai_chat":   "/api/v1/openai/chat/completions",
				"dify_chat":     "/api/v1/dify/chat-messages",
				"dify_workflow": "/api/v1/dify/workflows/run",
				"legacy_chat":   "/api/v1/chat (deprecated, use specific endpoints)",
				"documentation": "https://docs.agent-connector.com/dataflow-api",
			},
			"authentication": map[string]string{
				"method":      "API Key + Agent ID",
				"header":      "Authorization: Bearer <api_key> or X-API-Key: <api_key>",
				"agent_id":    "Provided in request body or URL parameter",
				"description": "API keys are generated and managed by Agent-Connector platform",
			},
			"features": []string{
				"Backend-based architecture with clear separation",
				"OpenAI compatible interface",
				"Dify Chat and Workflow interfaces",
				"Streaming and blocking response modes",
				"Automatic backend selection",
				"Redis-based distributed rate limiting",
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

		// Close rate limiter
		if redisRateLimiter != nil {
			redisRateLimiter.Close()
		}

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
	fmt.Println("📋 Ready to handle agent requests with new Backend architecture")
	fmt.Println("💡 Use Ctrl+C to gracefully shutdown the server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// setupMiddlewares setup common middlewares
func setupMiddlewares(router *gin.Engine, cfg *config.Config) {
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[DataFlow-Backend] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Recovery middleware
	router.Use(gin.Recovery())

	// Request body size limit
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.API.MaxRequestBodySize)
		c.Next()
	})
}

// printAPIEndpoints print API endpoints information
func printAPIEndpoints(cfg *config.Config) {
	fmt.Println("\n📡 Available API Endpoints (New Backend Architecture):")
	fmt.Println("├── GET  /                                    - Service information")
	fmt.Println("├── GET  /api/v1/health                       - Health check")
	fmt.Println("├── POST /api/v1/openai/chat/completions      - OpenAI compatible interface")
	fmt.Println("├── POST /api/v1/dify/chat-messages           - Dify Chat interface")
	fmt.Println("├── POST /api/v1/dify/workflows/run           - Dify Workflow interface")
	fmt.Println("└── POST /api/v1/chat                         - Legacy unified interface (deprecated)")

	fmt.Println("\n🔐 Authentication:")
	fmt.Println("├── Header: Authorization: Bearer <api_key>")
	fmt.Println("├── Header: X-API-Key: <api_key>")
	fmt.Println("└── Request Body: agent_id field required")

	fmt.Println("\n🌟 New Features:")
	fmt.Println("├── ✨ Backend-based architecture")
	fmt.Println("├── 🔄 Dedicated endpoints for each backend type")
	fmt.Println("├── 🎯 Better type safety and validation")
	fmt.Println("├── ⚡ Redis-based distributed rate limiting")
	fmt.Println("├── 📊 Enhanced monitoring and logging")
	fmt.Println("└── 🔧 Easier to extend and maintain")

	fmt.Println("\n📖 Usage Examples:")
	fmt.Println("# OpenAI-style request:")
	fmt.Printf("curl -X POST http://%s/api/v1/openai/chat/completions \\\n", cfg.GetServiceAddr("data"))
	fmt.Println("  -H \"Authorization: Bearer your-api-key\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"agent_id\": \"your-agent-id\", \"messages\": [{\"role\": \"user\", \"content\": \"Hello!\"}], \"model\": \"gpt-3.5-turbo\"}'")

	fmt.Println("\n# Dify Chat request:")
	fmt.Printf("curl -X POST http://%s/api/v1/dify/chat-messages \\\n", cfg.GetServiceAddr("data"))
	fmt.Println("  -H \"Authorization: Bearer your-api-key\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"agent_id\": \"your-agent-id\", \"query\": \"Hello!\", \"user\": \"user123\"}'")

	fmt.Println("\n# Dify Workflow request:")
	fmt.Printf("curl -X POST http://%s/api/v1/dify/workflows/run \\\n", cfg.GetServiceAddr("data"))
	fmt.Println("  -H \"Authorization: Bearer your-api-key\" \\")
	fmt.Println("  -H \"Content-Type: application/json\" \\")
	fmt.Println("  -d '{\"agent_id\": \"your-agent-id\", \"inputs\": {\"query\": \"Hello!\"}, \"user\": \"user123\"}'")
	fmt.Println()
}
