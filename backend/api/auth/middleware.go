package auth

import (
	"agent-connector/internal"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserContextKey user context key
const UserContextKey = "current_user"

// AuthMiddleware authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response := AuthResponse{
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &APIError{
					Type:    "authentication_error",
					Code:    "401",
					Message: "Missing or invalid authorization token",
				},
			}
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		userService := internal.NewUserService()
		session, err := userService.GetSessionByToken(token)
		if err != nil {
			response := AuthResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid or expired token",
				Error: &APIError{
					Type:    "authentication_error",
					Code:    "401",
					Message: err.Error(),
				},
			}
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		// Check user status
		if !session.User.IsActive() {
			response := AuthResponse{
				Code:    http.StatusForbidden,
				Message: "User account is not active",
				Error: &APIError{
					Type:    "authorization_error",
					Code:    "403",
					Message: "Your account has been deactivated",
				},
			}
			c.JSON(http.StatusForbidden, response)
			c.Abort()
			return
		}

		// Store user information in context
		c.Set(UserContextKey, &session.User)
		c.Next()
	})
}

// RequireRole role permission middleware
func RequireRole(roles ...internal.UserRole) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			response := AuthResponse{
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &APIError{
					Type:    "authentication_error",
					Code:    "401",
					Message: "User not authenticated",
				},
			}
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		// Check user role
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response := AuthResponse{
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &APIError{
					Type:    "authorization_error",
					Code:    "403",
					Message: "You don't have permission to access this resource",
				},
			}
			c.JSON(http.StatusForbidden, response)
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequirePermission permission check middleware
func RequirePermission(permission string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			response := AuthResponse{
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &APIError{
					Type:    "authentication_error",
					Code:    "401",
					Message: "User not authenticated",
				},
			}
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		if !user.HasPermission(permission) {
			response := AuthResponse{
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &APIError{
					Type:    "authorization_error",
					Code:    "403",
					Message: "You don't have permission to perform this action",
				},
			}
			c.JSON(http.StatusForbidden, response)
			c.Abort()
			return
		}

		c.Next()
	})
}

// AdminOnly only admin middleware
func AdminOnly() gin.HandlerFunc {
	return RequireRole(internal.UserRoleAdmin)
}

// AdminOrOperator admin or operator middleware
func AdminOrOperator() gin.HandlerFunc {
	return RequireRole(internal.UserRoleAdmin, internal.UserRoleOperator)
}

// OptionalAuth optional authentication middleware (not required)
func OptionalAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			userService := internal.NewUserService()
			session, err := userService.GetSessionByToken(token)
			if err == nil && session.User.IsActive() {
				c.Set(UserContextKey, &session.User)
			}
		}
		c.Next()
	})
}

// extractToken extract token from request
func extractToken(c *gin.Context) string {
	// Extract token from Authorization header
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:6]) == "BEARER" {
		return bearerToken[7:]
	}

	// Extract token from query parameters
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Extract token from cookie
	cookie, err := c.Cookie("auth_token")
	if err == nil {
		return cookie
	}

	return ""
}

// GetCurrentUser get current logged in user
func GetCurrentUser(c *gin.Context) *internal.User {
	if user, exists := c.Get(UserContextKey); exists {
		if u, ok := user.(*internal.User); ok {
			return u
		}
	}
	return nil
}

// GetCurrentUserID get current logged in user ID
func GetCurrentUserID(c *gin.Context) uint {
	user := GetCurrentUser(c)
	if user != nil {
		return user.ID
	}
	return 0
}

// IsCurrentUser check if the user is the current user
func IsCurrentUser(c *gin.Context, userID uint) bool {
	return GetCurrentUserID(c) == userID
}

// CanAccessUser check if the user can access the information of the specified user
func CanAccessUser(c *gin.Context, targetUserID uint) bool {
	currentUser := GetCurrentUser(c)
	if currentUser == nil {
		return false
	}

	// Admin can access all users
	if currentUser.CanManageUser() {
		return true
	}

	// Users can only access their own information
	return currentUser.ID == targetUserID
}

// RateLimitMiddleware simple rate limit middleware (based on IP)
func RateLimitMiddleware() gin.HandlerFunc {
	// Here you can integrate more complex rate limiting libraries, such as golang.org/x/time/rate
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip implementation for now, you can add Redis-based rate limiting later
		c.Next()
	})
}

// LoggingMiddleware logging middleware, record API access
func LoggingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Record request start time
		start := c.GetTime("start_time")
		if start.IsZero() {
			c.Set("start_time", c.GetTime("start_time"))
		}

		c.Next()

		// Here you can add more detailed logging
		// Including user ID, request path, response status, etc.
	})
}
