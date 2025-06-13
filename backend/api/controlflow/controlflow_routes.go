package controlflow

import (
	"github.com/gin-gonic/gin"
)

// SetupControlFlowRoutes setup control flow API routes
func SetupControlFlowRoutes(router *gin.Engine) {
	systemConfigHandler := NewDashboardSystemConfigHandler()
	agentHandler := NewDashboardAgentHandler()

	v1 := router.Group("/api/v1/controlflow")
	{
		// System configuration
		systemConfig := v1.Group("/system-config")
		{
			systemConfig.GET("", systemConfigHandler.GetSystemConfig)
			systemConfig.PUT("", systemConfigHandler.UpdateSystemConfig)
		}

		// Agent configuration
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
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Control Flow API is running",
		})
	})
}
