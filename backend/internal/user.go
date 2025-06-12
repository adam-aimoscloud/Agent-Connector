package internal

import (
	"time"

	"gorm.io/gorm"
)

// User user model
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password  string         `json:"-" gorm:"not null;size:255"` // not expose password in JSON
	FullName  string         `json:"full_name" gorm:"size:100"`
	Avatar    string         `json:"avatar" gorm:"size:255"`
	Role      UserRole       `json:"role" gorm:"default:'user'"`
	Status    UserStatus     `json:"status" gorm:"default:'active'"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserRole user role enum
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"    // admin
	UserRoleOperator UserRole = "operator" // operator
	UserRoleUser     UserRole = "user"     // user
	UserRoleReadonly UserRole = "readonly" // readonly
)

// UserStatus user status enum
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // active
	UserStatusInactive UserStatus = "inactive" // inactive
	UserStatusBlocked  UserStatus = "blocked"  // blocked
	UserStatusPending  UserStatus = "pending"  // pending
)

// TableName specify table name
func (User) TableName() string {
	return "users"
}

// UserSession user session model
type UserSession struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null;size:255"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserSession) TableName() string {
	return "user_sessions"
}

// IsExpired check if session is expired
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// UserLoginLog user login log
type UserLoginLog struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	IP        string    `json:"ip" gorm:"size:45"`
	UserAgent string    `json:"user_agent" gorm:"size:500"`
	Success   bool      `json:"success" gorm:"default:true"`
	Message   string    `json:"message" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

// TableName specify table name
func (UserLoginLog) TableName() string {
	return "user_login_logs"
}

// HasPermission check if user has specific permission
func (u *User) HasPermission(action string) bool {
	switch u.Role {
	case UserRoleAdmin:
		return true // admin has all permissions
	case UserRoleOperator:
		// operator can manage config but not user
		return action != "user_management"
	case UserRoleUser:
		// user can only view own profile
		return action == "view_own_profile"
	case UserRoleReadonly:
		// readonly user can only view
		return action == "view" || action == "view_own_profile"
	default:
		return false
	}
}

// CanManageUser check if user can manage user
func (u *User) CanManageUser() bool {
	return u.Role == UserRoleAdmin
}

// CanManageSystem check if user can manage system config
func (u *User) CanManageSystem() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOperator
}

// IsActive check if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Sanitize sanitize user data, remove sensitive information
func (u *User) Sanitize() {
	u.Password = ""
}
