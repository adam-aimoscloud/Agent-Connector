package dataflow

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupDataFlowRoutes set data flow API routes
func SetupDataFlowRoutes(router *gin.Engine) {
	handler := NewDataFlowAPIHandler()

	// data flow API routes group
	dataFlowAPI := router.Group("/api/v1/dataflow")
	{
		// health check
		dataFlowAPI.GET("/health", handler.HealthCheck)

		// general chat interface - support multiple agent types
		dataFlowAPI.POST("/chat/:agent_id", handler.HandleChat)

		// OpenAI compatible interface
		openaiAPI := dataFlowAPI.Group("/openai")
		{
			// OpenAI format chat completion
			openaiAPI.POST("/chat/completions/:agent_id", handler.HandleChat)
			openaiAPI.POST("/:agent_id/chat/completions", handler.HandleChat)
		}

		// Dify compatible interface
		difyAPI := dataFlowAPI.Group("/dify")
		{
			// Dify format chat message
			difyAPI.POST("/chat-messages/:agent_id", handler.HandleChat)
			difyAPI.POST("/:agent_id/chat-messages", handler.HandleChat)
		}
	}
}

// SetupDataFlowMiddlewares set data flow API middlewares
func SetupDataFlowMiddlewares(router *gin.Engine) {
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

	// request log middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
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

	// error recovery middleware
	router.Use(gin.Recovery())

	// request limit middleware (basic version)
	router.Use(func(c *gin.Context) {
		// limit request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10MB
		c.Next()
	})
}

// DataFlowRouterConfig data flow API router config
type DataFlowRouterConfig struct {
	EnableCORS         bool
	EnableLogging      bool
	EnableRecovery     bool
	MaxRequestBodySize int64
	APIRateLimit       int
}

// DefaultDataFlowRouterConfig default data flow API router config
func DefaultDataFlowRouterConfig() *DataFlowRouterConfig {
	return &DataFlowRouterConfig{
		EnableCORS:         true,
		EnableLogging:      true,
		EnableRecovery:     true,
		MaxRequestBodySize: 10 << 20, // 10MB
		APIRateLimit:       100,      // 100 requests per minute
	}
}

// SetupDataFlowRoutesWithConfig use config to set data flow API routes
func SetupDataFlowRoutesWithConfig(router *gin.Engine, config *DataFlowRouterConfig) {
	if config == nil {
		config = DefaultDataFlowRouterConfig()
	}

	// set configurable middlewares
	if config.EnableCORS {
		router.Use(setupCORSMiddleware())
	}

	if config.EnableLogging {
		router.Use(setupLoggingMiddleware())
	}

	if config.EnableRecovery {
		router.Use(gin.Recovery())
	}

	// request body size limit
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxRequestBodySize)
		c.Next()
	})

	// set routes
	SetupDataFlowRoutes(router)
}

// setupCORSMiddleware set CORS middleware
func setupCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// setupLoggingMiddleware set logging middleware
func setupLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[DataFlow] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
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
	})
}
