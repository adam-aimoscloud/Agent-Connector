package controlflow

import (
	"agent-connector/internal"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// DashboardSystemConfigHandler Dashboard system configuration handler
type DashboardSystemConfigHandler struct {
	service *internal.SystemConfigService
}

// NewDashboardSystemConfigHandler create Dashboard system configuration handler
func NewDashboardSystemConfigHandler() *DashboardSystemConfigHandler {
	return &DashboardSystemConfigHandler{
		service: &internal.SystemConfigService{},
	}
}

// GetSystemConfig get system configuration
func (h *DashboardSystemConfigHandler) GetSystemConfig(c *gin.Context) {
	config, err := h.service.GetSystemConfig()
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get system config",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "System config retrieved successfully",
		Data:    ConvertFromInternalSystemConfig(config),
	}
	c.JSON(http.StatusOK, response)
}

// UpdateSystemConfig update system configuration
func (h *DashboardSystemConfigHandler) UpdateSystemConfig(c *gin.Context) {
	var req SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	config := ConvertToInternalSystemConfig(&req)
	err := h.service.UpdateSystemConfig(config)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update system config",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// get updated configuration
	updatedConfig, err := h.service.GetSystemConfig()
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get updated system config",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "System config updated successfully",
		Data:    ConvertFromInternalSystemConfig(updatedConfig),
	}
	c.JSON(http.StatusOK, response)
}

// DashboardUserRateLimitHandler Dashboard user rate limit configuration handler
type DashboardUserRateLimitHandler struct {
	service *internal.UserRateLimitService
}

// NewDashboardUserRateLimitHandler create Dashboard user rate limit configuration handler
func NewDashboardUserRateLimitHandler() *DashboardUserRateLimitHandler {
	return &DashboardUserRateLimitHandler{
		service: &internal.UserRateLimitService{},
	}
}

// GetUserRateLimit get user rate limit configuration
func (h *DashboardUserRateLimitHandler) GetUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")

	userRateLimit, err := h.service.GetUserRateLimit(userID)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusNotFound,
			Message: "User rate limit not found",
			Error: &APIError{
				Type:    "not_found",
				Code:    "404",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "User rate limit retrieved successfully",
		Data:    ConvertFromInternalUserRateLimit(userRateLimit),
	}
	c.JSON(http.StatusOK, response)
}

// ListUserRateLimits list user rate limit configurations
func (h *DashboardUserRateLimitHandler) ListUserRateLimits(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	// TODO: implement search functionality
	// search := c.Query("search")

	userRateLimits, total, err := h.service.ListUserRateLimits(page, pageSize)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list user rate limits",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response := ControlFlowPaginationResponse{
		Code:    http.StatusOK,
		Message: "User rate limits retrieved successfully",
		Data:    ConvertFromInternalUserRateLimitList(userRateLimits),
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	c.JSON(http.StatusOK, response)
}

// CreateUserRateLimit create user rate limit configuration
func (h *DashboardUserRateLimitHandler) CreateUserRateLimit(c *gin.Context) {
	var req UserRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	userRateLimit := ConvertToInternalUserRateLimit(&req)
	err := h.service.CreateUserRateLimit(userRateLimit)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user rate limit",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusCreated,
		Message: "User rate limit created successfully",
		Data:    ConvertFromInternalUserRateLimit(userRateLimit),
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateUserRateLimit update user rate limit configuration
func (h *DashboardUserRateLimitHandler) UpdateUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")

	var req UserRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	userRateLimit := ConvertToInternalUserRateLimit(&req)

	err := h.service.UpdateUserRateLimit(userID, userRateLimit)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update user rate limit",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// get updated user rate limit configuration
	updatedUserRateLimit, err := h.service.GetUserRateLimit(userID)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get updated user rate limit",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "User rate limit updated successfully",
		Data:    ConvertFromInternalUserRateLimit(updatedUserRateLimit),
	}
	c.JSON(http.StatusOK, response)
}

// DeleteUserRateLimit delete user rate limit configuration
func (h *DashboardUserRateLimitHandler) DeleteUserRateLimit(c *gin.Context) {
	userID := c.Param("user_id")

	err := h.service.DeleteUserRateLimit(userID)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete user rate limit",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "User rate limit deleted successfully",
	}
	c.JSON(http.StatusOK, response)
}

// DashboardAgentHandler Dashboard agent configuration handler
type DashboardAgentHandler struct {
	service *internal.AgentService
}

// NewDashboardAgentHandler create Dashboard agent configuration handler
func NewDashboardAgentHandler() *DashboardAgentHandler {
	return &DashboardAgentHandler{
		service: &internal.AgentService{},
	}
}

// GetAgent get agent configuration
func (h *DashboardAgentHandler) GetAgent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid agent ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "Agent ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	agent, err := h.service.GetAgent(uint(id))
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusNotFound,
			Message: "Agent not found",
			Error: &APIError{
				Type:    "not_found",
				Code:    "404",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// default not hide secrets, admin need to see full information
	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "Agent retrieved successfully",
		Data:    ConvertFromInternalAgent(agent, false),
	}
	c.JSON(http.StatusOK, response)
}

// ListAgents list agent configurations
func (h *DashboardAgentHandler) ListAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	agents, total, err := h.service.ListAgents(page, pageSize, search)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list agents",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// in the list, you can choose to hide sensitive information
	hideSecrets := c.Query("hide_secrets") == "true"

	response := ControlFlowPaginationResponse{
		Code:    http.StatusOK,
		Message: "Agents retrieved successfully",
		Data:    ConvertFromInternalAgentList(agents, hideSecrets),
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	c.JSON(http.StatusOK, response)
}

// CreateAgent create agent configuration
func (h *DashboardAgentHandler) CreateAgent(c *gin.Context) {
	var req AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	agent := ConvertToInternalAgent(&req)
	err := h.service.CreateAgent(agent)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create agent",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusCreated,
		Message: "Agent created successfully",
		Data:    ConvertFromInternalAgent(agent, false),
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateAgent update agent configuration
func (h *DashboardAgentHandler) UpdateAgent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid agent ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "Agent ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var req AgentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request format",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// get existing agent
	agent, err := h.service.GetAgent(uint(id))
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusNotFound,
			Message: "Agent not found",
			Error: &APIError{
				Type:    "not_found",
				Code:    "404",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// update agent fields
	UpdateInternalAgentFromRequest(agent, &req)

	err = h.service.UpdateAgent(uint(id), agent)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update agent",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// get updated agent
	updatedAgent, err := h.service.GetAgent(uint(id))
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get updated agent",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "Agent updated successfully",
		Data:    ConvertFromInternalAgent(updatedAgent, false),
	}
	c.JSON(http.StatusOK, response)
}

// DeleteAgent delete agent configuration
func (h *DashboardAgentHandler) DeleteAgent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid agent ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "Agent ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	err = h.service.DeleteAgent(uint(id))
	if err != nil {
		response := ControlFlowResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete agent",
			Error: &APIError{
				Type:    "database_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := ControlFlowResponse{
		Code:    http.StatusOK,
		Message: "Agent deleted successfully",
	}
	c.JSON(http.StatusOK, response)
}

// HealthCheck health check
func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	// check database connection
	dbStatus := DatabaseHealthStatus{Status: "ok"}
	if internal.DB != nil {
		sqlDB, err := internal.DB.DB()
		if err != nil {
			dbStatus.Status = "error"
			dbStatus.Error = err.Error()
		} else {
			stats := sqlDB.Stats()
			dbStatus.Connections = stats.OpenConnections
		}
	} else {
		dbStatus.Status = "not_connected"
	}

	response := HealthCheckResponse{
		Status:    "ok",
		Service:   "control-flow-api",
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
		Uptime:    fmt.Sprintf("%.0f seconds", uptime.Seconds()),
		Database:  dbStatus,
	}

	c.JSON(http.StatusOK, response)
}
