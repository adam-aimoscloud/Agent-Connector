package auth

import (
	"agent-connector/internal"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler authentication handler
type AuthHandler struct {
	userService *internal.UserService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userService: internal.NewUserService(),
	}
}

// Register user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	user := ConvertToInternalUser(&req)
	if err := h.userService.CreateUser(user); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}

		response := AuthResponse{
			Code:    statusCode,
			Message: "Failed to create user",
			Error: &APIError{
				Type:    "registration_error",
				Code:    strconv.Itoa(statusCode),
				Message: err.Error(),
			},
		}
		c.JSON(statusCode, response)
		return
	}

	// Clean up password field
	user.Sanitize()

	response := AuthResponse{
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data:    ConvertFromInternalUser(user),
	}
	c.JSON(http.StatusCreated, response)
}

// Login user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	// Authenticate user
	user, err := h.userService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		// Record login failure log
		if user != nil {
			h.userService.LogUserLogin(user.ID, c.ClientIP(), c.GetHeader("User-Agent"), false, err.Error())
		}

		response := AuthResponse{
			Code:    http.StatusUnauthorized,
			Message: "Login failed",
			Error: &APIError{
				Type:    "authentication_error",
				Code:    "401",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Create session
	session, err := h.userService.CreateSession(user.ID)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create session",
			Error: &APIError{
				Type:    "session_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Record login success log
	h.userService.LogUserLogin(user.ID, c.ClientIP(), c.GetHeader("User-Agent"), true, "Login successful")

	// Clean up password field
	user.Sanitize()

	loginResponse := LoginResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		User:      *ConvertFromInternalUser(user),
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Login successful",
		Data:    loginResponse,
	}
	c.JSON(http.StatusOK, response)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token != "" {
		h.userService.DeleteSession(token)
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Logout successful",
	}
	c.JSON(http.StatusOK, response)
}

// GetProfile get user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		response := AuthResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
			Error: &APIError{
				Type:    "authentication_error",
				Code:    "401",
				Message: "User not found in context",
			},
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user statistics
	loginLogs, _, _ := h.userService.GetUserLoginLogs(user.ID, 1, 1)
	totalLogins, _, _ := h.userService.GetUserLoginLogs(user.ID, 1, 1000)

	stats := UserStatsResponse{
		TotalLogins:   int64(len(totalLogins)),
		AccountAge:    int(time.Since(user.CreatedAt).Hours() / 24),
		LastLoginTime: user.LastLogin,
	}

	if len(loginLogs) > 0 {
		stats.LastLoginIP = loginLogs[0].IP
	}

	// Get session information
	token := extractToken(c)
	sessionInfo := SessionInfoResponse{}
	if token != "" {
		if session, err := h.userService.GetSessionByToken(token); err == nil {
			sessionInfo = *ConvertFromInternalSession(session)
		}
	}

	profileResponse := UserProfileResponse{
		User:    *ConvertFromInternalUser(user),
		Stats:   stats,
		Session: sessionInfo,
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Profile retrieved successfully",
		Data:    profileResponse,
	}
	c.JSON(http.StatusOK, response)
}

// UpdateProfile update user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		response := AuthResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
			Error: &APIError{
				Type:    "authentication_error",
				Code:    "401",
				Message: "User not found in context",
			},
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	// Update user information
	UpdateInternalUserFromProfileRequest(user, &req)

	if err := h.userService.UpdateUser(user); err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update profile",
			Error: &APIError{
				Type:    "update_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Profile updated successfully",
		Data:    ConvertFromInternalUser(user),
	}
	c.JSON(http.StatusOK, response)
}

// ChangePassword change password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		response := AuthResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
			Error: &APIError{
				Type:    "authentication_error",
				Code:    "401",
				Message: "User not found in context",
			},
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	if err := h.userService.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid old password" {
			statusCode = http.StatusBadRequest
		}

		response := AuthResponse{
			Code:    statusCode,
			Message: "Failed to change password",
			Error: &APIError{
				Type:    "password_error",
				Code:    strconv.Itoa(statusCode),
				Message: err.Error(),
			},
		}
		c.JSON(statusCode, response)
		return
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "Password changed successfully",
	}
	c.JSON(http.StatusOK, response)
}

// GetLoginLogs get login logs
func (h *AuthHandler) GetLoginLogs(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		response := AuthResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
			Error: &APIError{
				Type:    "authentication_error",
				Code:    "401",
				Message: "User not found in context",
			},
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logs, total, err := h.userService.GetUserLoginLogs(user.ID, page, pageSize)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get login logs",
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

	response := AuthPaginationResponse{
		Code:    http.StatusOK,
		Message: "Login logs retrieved successfully",
		Data:    ConvertFromInternalLoginLogList(logs),
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	c.JSON(http.StatusOK, response)
}

// -- Admin functions --

// ListUsers get user list (admin function)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	users, total, err := h.userService.ListUsers(page, pageSize, search)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list users",
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

	response := AuthPaginationResponse{
		Code:    http.StatusOK,
		Message: "Users retrieved successfully",
		Data:    ConvertFromInternalUserList(users),
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	c.JSON(http.StatusOK, response)
}

// CreateUser create user (admin function)
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	user := ConvertToInternalUserFromCreateRequest(&req)
	if err := h.userService.CreateUser(user); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}

		response := AuthResponse{
			Code:    statusCode,
			Message: "Failed to create user",
			Error: &APIError{
				Type:    "creation_error",
				Code:    strconv.Itoa(statusCode),
				Message: err.Error(),
			},
		}
		c.JSON(statusCode, response)
		return
	}

	user.Sanitize()

	response := AuthResponse{
		Code:    http.StatusCreated,
		Message: "User created successfully",
		Data:    ConvertFromInternalUser(user),
	}
	c.JSON(http.StatusCreated, response)
}

// GetUser get user information (admin function)
func (h *AuthHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "User ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusNotFound,
			Message: "User not found",
			Error: &APIError{
				Type:    "not_found",
				Code:    "404",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	user.Sanitize()

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "User retrieved successfully",
		Data:    ConvertFromInternalUser(user),
	}
	c.JSON(http.StatusOK, response)
}

// UpdateUser update user information (admin function)
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "User ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusNotFound,
			Message: "User not found",
			Error: &APIError{
				Type:    "not_found",
				Code:    "404",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	UpdateInternalUserFromRequest(user, &req)

	if err := h.userService.UpdateUser(user); err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update user",
			Error: &APIError{
				Type:    "update_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	user.Sanitize()

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "User updated successfully",
		Data:    ConvertFromInternalUser(user),
	}
	c.JSON(http.StatusOK, response)
}

// DeleteUser delete user (admin function)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "User ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete user",
			Error: &APIError{
				Type:    "deletion_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "User deleted successfully",
	}
	c.JSON(http.StatusOK, response)
}

// UpdateUserStatus update user status (admin function)
func (h *AuthHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := AuthResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid user ID",
			Error: &APIError{
				Type:    "validation_error",
				Code:    "400",
				Message: "User ID must be a valid number",
			},
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response := AuthResponse{
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

	if err := h.userService.UpdateUserStatus(uint(id), internal.UserStatus(req.Status)); err != nil {
		response := AuthResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update user status",
			Error: &APIError{
				Type:    "update_error",
				Code:    "500",
				Message: err.Error(),
			},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := AuthResponse{
		Code:    http.StatusOK,
		Message: "User status updated successfully",
	}
	c.JSON(http.StatusOK, response)
}
