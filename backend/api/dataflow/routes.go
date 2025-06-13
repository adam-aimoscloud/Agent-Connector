package dataflow

import (
	"agent-connector/pkg/ratelimiter"

	"github.com/gin-gonic/gin"
)

// SetupBackendRoutes setup routes for backend-based dataflow API
func SetupBackendRoutes(router *gin.Engine, rateLimiter *ratelimiter.RedisRateLimiter) {
	// Create handler
	handler := NewDataFlowAPIHandler(rateLimiter)

	// Create middleware
	middleware := NewDataFlowMiddleware()

	// Create API group
	api := router.Group("/api/v1")

	// Apply middleware
	api.Use(middleware.AuthenticationMiddleware())
	api.Use(middleware.RateLimitMiddleware())

	// OpenAI Compatible Routes
	openai := api.Group("/openai")
	{
		openai.POST("/chat/completions", handler.HandleOpenAIChat)
	}

	// Dify Routes
	dify := api.Group("/dify")
	{
		// Chat Messages API
		dify.POST("/chat-messages", handler.HandleDifyChat)

		// Workflow API
		dify.POST("/workflows/run", handler.HandleDifyWorkflow)
	}

	// Health check
	api.GET("/health", handler.HealthCheck)
}

// SetupLegacyRoutes setup legacy routes for backward compatibility
func SetupLegacyRoutes(router *gin.Engine, rateLimiter *ratelimiter.RedisRateLimiter) {
	// Create legacy handler
	legacyHandler := NewDataFlowAPIHandler(rateLimiter)

	// Create middleware
	middleware := NewDataFlowMiddleware()

	// Create API group
	api := router.Group("/api/v1")

	// Apply middleware
	api.Use(middleware.AuthenticationMiddleware())
	api.Use(middleware.RateLimitMiddleware())

	// Legacy unified endpoint
	api.POST("/chat", legacyHandler.HandleChat)
}
