package main

import (
	"agent-connector/api/controlflow"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Mock data structures for testing
var (
	// Mock system config
	mockSystemConfig = map[string]interface{}{
		"rate_limit_mode":  "priority",
		"default_priority": 5,
		"default_qps":      10,
	}

	// Mock user rate limits
	mockUserRateLimits = map[string]map[string]interface{}{
		"user1": {
			"user_id":  "user1",
			"priority": 8,
			"qps":      20,
		},
		"user2": {
			"user_id":  "user2",
			"priority": 3,
			"qps":      5,
		},
	}

	// Mock agents
	mockAgents = map[uint]map[string]interface{}{
		1: {
			"id":          1,
			"name":        "Test Agent 1",
			"type":        "chat",
			"status":      "active",
			"description": "A test chat agent",
		},
		2: {
			"id":          2,
			"name":        "Test Agent 2",
			"type":        "completion",
			"status":      "inactive",
			"description": "A test completion agent",
		},
	}
	nextAgentID uint = 3
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Add request log middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Set routes
	controlflow.SetupControlFlowRoutes(r)

	// Get port, default 8081
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Control Flow API Mock server starting on port %s", port)
	log.Printf("Using in-memory data structures (no database required)")
	log.Printf("Health check: http://localhost:%s/api/v1/health", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// System config handlers
func getSystemConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockSystemConfig,
	})
}

func updateSystemConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	// Validate required fields
	if mode, ok := req["rate_limit_mode"]; !ok || (mode != "priority" && mode != "qps") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid rate limit mode",
		})
		return
	}

	// Update mock config
	for k, v := range req {
		mockSystemConfig[k] = v
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockSystemConfig,
	})
}

// User rate limit handlers
func getUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")
	if rateLimit, exists := mockUserRateLimits[userID]; exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    rateLimit,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User rate limit not found",
		})
	}
}

func listUserRateLimits(c *gin.Context) {
	rateLimits := make([]map[string]interface{}, 0, len(mockUserRateLimits))
	for _, rateLimit := range mockUserRateLimits {
		rateLimits = append(rateLimits, rateLimit)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": rateLimits,
			"total": len(rateLimits),
			"page":  1,
			"size":  len(rateLimits),
		},
	})
}

func createUserRateLimit(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	userID, ok := req["user_id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user_id is required",
		})
		return
	}

	if _, exists := mockUserRateLimits[userID]; exists {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "User rate limit already exists",
		})
		return
	}

	mockUserRateLimits[userID] = req

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    req,
	})
}

func updateUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	req["user_id"] = userID
	mockUserRateLimits[userID] = req

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
	})
}

func deleteUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")

	if _, exists := mockUserRateLimits[userID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User rate limit not found",
		})
		return
	}

	delete(mockUserRateLimits, userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User rate limit deleted",
	})
}

// Agent handlers
func getAgent(c *gin.Context) {
	idStr := c.Param("id")
	id := parseAgentID(idStr)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}

	if agent, exists := mockAgents[id]; exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    agent,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found",
		})
	}
}

// Helper function to parse agent ID
func parseAgentID(idStr string) uint {
	switch idStr {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	case "5":
		return 5
	default:
		return 0
	}
}

func listAgents(c *gin.Context) {
	agents := make([]map[string]interface{}, 0, len(mockAgents))
	agentType := c.Query("type")

	for _, agent := range mockAgents {
		if agentType == "" || agent["type"] == agentType {
			agents = append(agents, agent)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": agents,
			"total": len(agents),
			"page":  1,
			"size":  len(agents),
		},
	})
}

func createAgent(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	// Validate required fields
	if req["name"] == "" || req["type"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name and type are required",
		})
		return
	}

	req["id"] = nextAgentID
	req["status"] = "active" // default status
	mockAgents[nextAgentID] = req
	nextAgentID++

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    req,
	})
}

func updateAgent(c *gin.Context) {
	idStr := c.Param("id")
	id := parseAgentID(idStr)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}

	if _, exists := mockAgents[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found",
		})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	req["id"] = id
	mockAgents[id] = req

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    req,
	})
}

func deleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id := parseAgentID(idStr)
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid agent ID",
		})
		return
	}

	if _, exists := mockAgents[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Agent not found",
		})
		return
	}

	delete(mockAgents, id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent deleted",
	})
}
