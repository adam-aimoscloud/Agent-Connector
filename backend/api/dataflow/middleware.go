package dataflow

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	"agent-connector/config"
	"agent-connector/pkg/ratelimiter"
)

// AgentRateLimiterManager manages rate limiters for different agents
type AgentRateLimiterManager struct {
	limiters map[string]ratelimiter.RateLimiter
	mutex    sync.RWMutex
}

// NewAgentRateLimiterManager creates a new agent rate limiter manager
func NewAgentRateLimiterManager() *AgentRateLimiterManager {
	return &AgentRateLimiterManager{
		limiters: make(map[string]ratelimiter.RateLimiter),
	}
}

// GetOrCreateLimiter gets or creates a rate limiter for the given agent
func (m *AgentRateLimiterManager) GetOrCreateLimiter(agentID string, qps int) (ratelimiter.RateLimiter, error) {
	m.mutex.RLock()
	limiter, exists := m.limiters[agentID]
	m.mutex.RUnlock()

	if exists {
		return limiter, nil
	}

	// Get Redis address from global config
	redisAddr := config.GlobalConfig.Redis.Addr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // fallback default
	}

	// Create new limiter with Redis backend
	config := &ratelimiter.Config{
		Rate:  float64(qps),
		Burst: qps * 2, // burst is 2x the QPS
		Redis: &ratelimiter.RedisConfig{
			Addr:            redisAddr,
			Password:        config.GlobalConfig.Redis.Password,
			DB:              config.GlobalConfig.Redis.DB,
			PoolSize:        10,
			MinIdleConns:    2,
			ConnMaxIdleTime: 30 * 60 * 1000 * 1000 * 1000, // 30 minutes
		},
	}

	newLimiter, err := ratelimiter.NewRateLimiter(ratelimiter.RedisType, config)
	if err != nil {
		return nil, err
	}

	m.mutex.Lock()
	// Double-check in case another goroutine created it
	if existingLimiter, exists := m.limiters[agentID]; exists {
		m.mutex.Unlock()
		newLimiter.Close() // cleanup the newly created limiter
		return existingLimiter, nil
	}
	m.limiters[agentID] = newLimiter
	m.mutex.Unlock()

	return newLimiter, nil
}

// Close closes all rate limiters
func (m *AgentRateLimiterManager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, limiter := range m.limiters {
		limiter.Close()
	}
	m.limiters = make(map[string]ratelimiter.RateLimiter)
	return nil
}

// DataFlowMiddleware contains middleware dependencies
type DataFlowMiddleware struct {
	authService        *DataFlowAuthService
	rateLimiterManager *AgentRateLimiterManager
}

// NewDataFlowMiddleware creates a new middleware instance
func NewDataFlowMiddleware() *DataFlowMiddleware {
	return &DataFlowMiddleware{
		authService:        NewDataFlowAuthService(),
		rateLimiterManager: NewAgentRateLimiterManager(),
	}
}

// AuthenticationMiddleware handles authentication for dataflow API
func (m *DataFlowMiddleware) AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get AgentID from URL parameters or JSON body
		agentID := c.Param("agent_id")
		if agentID == "" {
			agentID = c.Query("agent_id")
		}

		// get API Key from header
		apiKey := c.GetHeader("Authorization")
		if apiKey == "" {
			apiKey = c.GetHeader("X-API-Key")
		}

		// authenticate request
		authInfo, err := m.authService.AuthenticateRequest(agentID, apiKey)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "authentication_failed", err.Error())
			c.Abort()
			return
		}

		// store auth info in context for later use
		c.Set("authInfo", authInfo)
		c.Next()
	}
}

// RateLimitMiddleware handles rate limiting for dataflow API
func (m *DataFlowMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get auth info from context
		authInfoValue, exists := c.Get("authInfo")
		if !exists {
			m.respondWithError(c, http.StatusInternalServerError, "internal_error", "Authentication info not found")
			c.Abort()
			return
		}

		authInfo, ok := authInfoValue.(*AuthInfo)
		if !ok {
			m.respondWithError(c, http.StatusInternalServerError, "internal_error", "Invalid authentication info")
			c.Abort()
			return
		}

		// agent-level rate limiting
		if m.rateLimiterManager != nil {
			agentLimiter, err := m.rateLimiterManager.GetOrCreateLimiter(authInfo.AgentID, authInfo.Agent.QPS)
			if err != nil {
				m.respondWithError(c, http.StatusInternalServerError, "rate_limit_error", "Failed to get agent rate limiter: "+err.Error())
				c.Abort()
				return
			}

			// Check rate limit
			agentKey := fmt.Sprintf("agent:%s", authInfo.AgentID)
			allowed, err := agentLimiter.Allow(c.Request.Context(), agentKey)
			if err != nil {
				m.respondWithError(c, http.StatusInternalServerError, "rate_limit_error", "Rate limit check failed: "+err.Error())
				c.Abort()
				return
			}

			if !allowed {
				m.respondWithError(c, http.StatusTooManyRequests, "rate_limit_exceeded", "Agent rate limit exceeded")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// respondWithError return error response
func (m *DataFlowMiddleware) respondWithError(c *gin.Context, statusCode int, errorType, message string) {
	response := DataFlowResponse{
		Code:    statusCode,
		Message: "Error",
		Error: &APIError{
			Type:    errorType,
			Code:    strconv.Itoa(statusCode),
			Message: message,
		},
	}
	c.JSON(statusCode, response)
}

// respondWithRateLimit return rate limit response
func (m *DataFlowMiddleware) respondWithRateLimit(c *gin.Context, agentQPS int) {
	response := DataFlowResponse{
		Code:    http.StatusTooManyRequests,
		Message: "Rate limit exceeded",
		Error: &APIError{
			Type:    "rate_limit_exceeded",
			Code:    "429",
			Message: fmt.Sprintf("Agent rate limit exceeded. Agent QPS: %d", agentQPS),
		},
	}

	// set Rate Limit headers
	c.Header("X-RateLimit-Agent-QPS", strconv.Itoa(agentQPS))
	c.Header("Retry-After", "1") // suggest retry after 1 second

	c.JSON(http.StatusTooManyRequests, response)
}

// Close closes the middleware resources
func (m *DataFlowMiddleware) Close() error {
	if m.rateLimiterManager != nil {
		return m.rateLimiterManager.Close()
	}
	return nil
}
