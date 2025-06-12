package auth

import (
	"agent-connector/internal"
	"time"
)

// AuthResponse authentication API common response structure
type AuthResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError API error structure
type APIError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// AuthPaginationResponse authentication API pagination response structure
type AuthPaginationResponse struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Error      *APIError      `json:"error,omitempty"`
}

// PaginationInfo pagination information
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// RegisterRequest user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	FullName string `json:"full_name" binding:"max=100"`
}

// LoginRequest user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse login successful response
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// ChangePasswordRequest change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// UpdateProfileRequest update personal information request
type UpdateProfileRequest struct {
	FullName string `json:"full_name,omitempty" binding:"max=100"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Avatar   string `json:"avatar,omitempty" binding:"max=255"`
}

// UserResponse user information response
type UserResponse struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	Avatar    string     `json:"avatar"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateUserRequest create user request (admin function)
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	FullName string `json:"full_name" binding:"max=100"`
	Role     string `json:"role" binding:"required,oneof=admin operator user readonly"`
	Status   string `json:"status" binding:"required,oneof=active inactive blocked pending"`
}

// UpdateUserRequest update user request (admin function)
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" binding:"omitempty,email"`
	FullName *string `json:"full_name,omitempty" binding:"omitempty,max=100"`
	Role     *string `json:"role,omitempty" binding:"omitempty,oneof=admin operator user readonly"`
	Status   *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive blocked pending"`
	Avatar   *string `json:"avatar,omitempty" binding:"omitempty,max=255"`
}

// UpdateUserStatusRequest update user status request
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive blocked pending"`
}

// LoginLogResponse login log response
type LoginLogResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// UserProfileResponse user profile response
type UserProfileResponse struct {
	User    UserResponse        `json:"user"`
	Stats   UserStatsResponse   `json:"stats"`
	Session SessionInfoResponse `json:"session"`
}

// UserStatsResponse user statistics information
type UserStatsResponse struct {
	TotalLogins   int64      `json:"total_logins"`
	LastLoginIP   string     `json:"last_login_ip"`
	LoginCount30d int64      `json:"login_count_30d"`
	AccountAge    int        `json:"account_age_days"`
	LastLoginTime *time.Time `json:"last_login_time"`
}

// SessionInfoResponse session information
type SessionInfoResponse struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsExpired bool      `json:"is_expired"`
}

// ConvertFromInternalUser convert from internal user model to response structure
func ConvertFromInternalUser(user *internal.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Avatar:    user.Avatar,
		Role:      string(user.Role),
		Status:    string(user.Status),
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ConvertToInternalUser convert from register request to internal user model
func ConvertToInternalUser(req *RegisterRequest) *internal.User {
	return &internal.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     internal.UserRoleUser,
		Status:   internal.UserStatusActive,
	}
}

// ConvertToInternalUserFromCreateRequest convert from create user request to internal user model
func ConvertToInternalUserFromCreateRequest(req *CreateUserRequest) *internal.User {
	return &internal.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     internal.UserRole(req.Role),
		Status:   internal.UserStatus(req.Status),
	}
}

// UpdateInternalUserFromRequest update internal user model with update request data
func UpdateInternalUserFromRequest(user *internal.User, req *UpdateUserRequest) {
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Role != nil {
		user.Role = internal.UserRole(*req.Role)
	}
	if req.Status != nil {
		user.Status = internal.UserStatus(*req.Status)
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
}

// UpdateInternalUserFromProfileRequest update internal user model with personal information update request data
func UpdateInternalUserFromProfileRequest(user *internal.User, req *UpdateProfileRequest) {
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
}

// ConvertFromInternalUserList convert from internal user model list to response list
func ConvertFromInternalUserList(users []*internal.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, user := range users {
		result[i] = ConvertFromInternalUser(user)
	}
	return result
}

// ConvertFromInternalLoginLog convert from internal login log model to response structure
func ConvertFromInternalLoginLog(log *internal.UserLoginLog) *LoginLogResponse {
	return &LoginLogResponse{
		ID:        log.ID,
		UserID:    log.UserID,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Success:   log.Success,
		Message:   log.Message,
		CreatedAt: log.CreatedAt,
	}
}

// ConvertFromInternalLoginLogList convert from internal login log model list to response list
func ConvertFromInternalLoginLogList(logs []*internal.UserLoginLog) []*LoginLogResponse {
	result := make([]*LoginLogResponse, len(logs))
	for i, log := range logs {
		result[i] = ConvertFromInternalLoginLog(log)
	}
	return result
}

// ConvertFromInternalSession convert from internal session model to session information response
func ConvertFromInternalSession(session *internal.UserSession) *SessionInfoResponse {
	return &SessionInfoResponse{
		Token:     session.Token,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
		IsExpired: session.IsExpired(),
	}
}
