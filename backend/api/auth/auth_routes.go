package auth

import (
	"agent-connector/internal"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication API routes
func SetupAuthRoutes(r *gin.Engine) {
	// Create handlers
	authHandler := NewAuthHandler()

	// API version grouping
	apiV1 := r.Group("/api/v1")

	// Public routes (no authentication required)
	auth := apiV1.Group("/auth")
	{
		// Basic authentication interfaces
		auth.POST("/register", authHandler.Register) // User registration
		auth.POST("/login", authHandler.Login)       // User login

		// Service information interfaces
		auth.GET("/", getAuthServiceInfo) // Service information
		auth.GET("/health", healthCheck)  // Health check
	}

	// Routes requiring authentication
	authProtected := apiV1.Group("/auth")
	authProtected.Use(AuthMiddleware())
	{
		// User profile management
		authProtected.POST("/logout", authHandler.Logout)                  // User logout
		authProtected.GET("/profile", authHandler.GetProfile)              // Get profile
		authProtected.PUT("/profile", authHandler.UpdateProfile)           // Update profile
		authProtected.POST("/change-password", authHandler.ChangePassword) // Change password
		authProtected.GET("/login-logs", authHandler.GetLoginLogs)         // Get login logs
	}

	// User management routes (admin functionality)
	userManagement := apiV1.Group("/users")
	userManagement.Use(AuthMiddleware())
	userManagement.Use(AdminOnly())
	{
		userManagement.GET("", authHandler.ListUsers)                   // Get user list
		userManagement.POST("", authHandler.CreateUser)                 // Create user
		userManagement.GET("/:id", authHandler.GetUser)                 // Get user information
		userManagement.PUT("/:id", authHandler.UpdateUser)              // Update user information
		userManagement.DELETE("/:id", authHandler.DeleteUser)           // Delete user
		userManagement.PUT("/:id/status", authHandler.UpdateUserStatus) // Update user status
	}

	// System management routes (admin and operator)
	system := apiV1.Group("/system")
	system.Use(AuthMiddleware())
	system.Use(AdminOrOperator())
	{
		system.POST("/cleanup-sessions", cleanupExpiredSessions) // Clean up expired sessions
		system.GET("/stats", getSystemStats)                     // Get system statistics
	}
}

// getAuthServiceInfo gets authentication service information
func getAuthServiceInfo(c *gin.Context) {
	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Authentication service information",
		Data: gin.H{
			"service":     "auth-api",
			"version":     "1.0.0",
			"description": "User authentication and management service",
			"endpoints": gin.H{
				"public": []string{
					"POST /api/v1/auth/register",
					"POST /api/v1/auth/login",
					"GET  /api/v1/auth/health",
				},
				"authenticated": []string{
					"POST /api/v1/auth/logout",
					"GET  /api/v1/auth/profile",
					"PUT  /api/v1/auth/profile",
					"POST /api/v1/auth/change-password",
					"GET  /api/v1/auth/login-logs",
				},
				"admin_only": []string{
					"GET    /api/v1/users",
					"POST   /api/v1/users",
					"GET    /api/v1/users/:id",
					"PUT    /api/v1/users/:id",
					"DELETE /api/v1/users/:id",
					"PUT    /api/v1/users/:id/status",
				},
			},
			"features": []string{
				"User registration and authentication",
				"Session-based authentication with tokens",
				"Role-based access control (RBAC)",
				"Password management",
				"User profile management",
				"Login audit logs",
				"User management (admin)",
			},
		},
	}
	c.JSON(http.StatusOK, response)
}

// healthCheck health check
func healthCheck(c *gin.Context) {
	// Check database connection
	dbStatus := "ok"
	dbError := ""
	if internal.DB != nil {
		sqlDB, err := internal.DB.DB()
		if err != nil {
			dbStatus = "error"
			dbError = err.Error()
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "error"
			dbError = err.Error()
		}
	} else {
		dbStatus = "not_connected"
	}

	// Check if there are admin users
	var adminCount int64
	hasAdmin := false
	if internal.DB != nil {
		internal.DB.Model(&internal.User{}).Where("role = ?", internal.UserRoleAdmin).Count(&adminCount)
		hasAdmin = adminCount > 0
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Authentication service health check",
		Data: gin.H{
			"status":    "ok",
			"service":   "auth-api",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"database": gin.H{
				"status": dbStatus,
				"error":  dbError,
			},
			"system": gin.H{
				"has_admin":   hasAdmin,
				"admin_count": adminCount,
				"uptime":      fmt.Sprintf("%.0f seconds", time.Since(time.Now()).Seconds()),
			},
		},
	}

	statusCode := http.StatusOK
	if dbStatus == "error" {
		statusCode = http.StatusServiceUnavailable
		response.Code = statusCode
		response.Message = "Service degraded - database connection issues"
	}

	c.JSON(statusCode, response)
}

// cleanupExpiredSessions cleans up expired sessions
func cleanupExpiredSessions(c *gin.Context) {
	userService := internal.NewUserService()
	if err := userService.CleanExpiredSessions(); err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to cleanup expired sessions",
			Error: &APIError{
				Type:    "cleanup_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Expired sessions cleaned up successfully",
	}
	c.JSON(http.StatusOK, response)
}

// getSystemStats gets system statistics
func getSystemStats(c *gin.Context) {
	if internal.DB == nil {
		response := AuthResponse{
			Code:    http.StatusServiceUnavailable,
			Message: "Database not available",
			Error: &APIError{
				Type:    "database_error",
				Code:    "503",
				Message: "Database connection not established",
			},
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	// Count users
	var userStats struct {
		TotalUsers   int64 `json:"total_users"`
		ActiveUsers  int64 `json:"active_users"`
		AdminUsers   int64 `json:"admin_users"`
		BlockedUsers int64 `json:"blocked_users"`
	}

	internal.DB.Model(&internal.User{}).Count(&userStats.TotalUsers)
	internal.DB.Model(&internal.User{}).Where("status = ?", internal.UserStatusActive).Count(&userStats.ActiveUsers)
	internal.DB.Model(&internal.User{}).Where("role = ?", internal.UserRoleAdmin).Count(&userStats.AdminUsers)
	internal.DB.Model(&internal.User{}).Where("status = ?", internal.UserStatusBlocked).Count(&userStats.BlockedUsers)

	// Count sessions
	var sessionStats struct {
		TotalSessions   int64 `json:"total_sessions"`
		ActiveSessions  int64 `json:"active_sessions"`
		ExpiredSessions int64 `json:"expired_sessions"`
	}

	internal.DB.Model(&internal.UserSession{}).Count(&sessionStats.TotalSessions)
	internal.DB.Model(&internal.UserSession{}).Where("expires_at > ?", time.Now()).Count(&sessionStats.ActiveSessions)
	internal.DB.Model(&internal.UserSession{}).Where("expires_at <= ?", time.Now()).Count(&sessionStats.ExpiredSessions)

	// Count login logs
	var loginStats struct {
		TotalLogins      int64 `json:"total_logins"`
		SuccessfulLogins int64 `json:"successful_logins"`
		FailedLogins     int64 `json:"failed_logins"`
		LoginsToday      int64 `json:"logins_today"`
	}

	today := time.Now().Truncate(24 * time.Hour)
	internal.DB.Model(&internal.UserLoginLog{}).Count(&loginStats.TotalLogins)
	internal.DB.Model(&internal.UserLoginLog{}).Where("success = ?", true).Count(&loginStats.SuccessfulLogins)
	internal.DB.Model(&internal.UserLoginLog{}).Where("success = ?", false).Count(&loginStats.FailedLogins)
	internal.DB.Model(&internal.UserLoginLog{}).Where("created_at >= ?", today).Count(&loginStats.LoginsToday)

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "System statistics retrieved successfully",
		Data: gin.H{
			"users":     userStats,
			"sessions":  sessionStats,
			"logins":    loginStats,
			"timestamp": time.Now().Unix(),
		},
	}
	c.JSON(http.StatusOK, response)
}
