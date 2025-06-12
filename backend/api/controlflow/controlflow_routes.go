package controlflow

import (
	"github.com/gin-gonic/gin"
)

// SetupControlFlowRoutes sets up control flow API routes
func SetupControlFlowRoutes(r *gin.Engine) {
	// Create handler instances
	systemConfigHandler := NewDashboardSystemConfigHandler()
	userRateLimitHandler := NewDashboardUserRateLimitHandler()
	agentHandler := NewDashboardAgentHandler()

	// API v1 route group
	v1 := r.Group("/api/v1")
	{
		// System configuration routes
		systemConfig := v1.Group("/system")
		{
			systemConfig.GET("/config", systemConfigHandler.GetSystemConfig)
			systemConfig.PUT("/config", systemConfigHandler.UpdateSystemConfig)
		}

		// User rate limit configuration routes
		userRateLimit := v1.Group("/user-rate-limits")
		{
			userRateLimit.GET("", userRateLimitHandler.ListUserRateLimits)
			userRateLimit.POST("", userRateLimitHandler.CreateUserRateLimit)
			userRateLimit.GET("/:user_id", userRateLimitHandler.GetUserRateLimit)
			userRateLimit.PUT("/:user_id", userRateLimitHandler.UpdateUserRateLimit)
			userRateLimit.DELETE("/:user_id", userRateLimitHandler.DeleteUserRateLimit)
		}

		// Agent configuration routes
		agents := v1.Group("/agents")
		{
			agents.GET("", agentHandler.ListAgents)
			agents.POST("", agentHandler.CreateAgent)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.PUT("/:id", agentHandler.UpdateAgent)
			agents.DELETE("/:id", agentHandler.DeleteAgent)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Control Flow API is running",
		})
	})
}
