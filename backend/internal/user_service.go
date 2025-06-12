package internal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService user service
type UserService struct{}

// NewUserService create user service instance
func NewUserService() *UserService {
	return &UserService{}
}

// CreateUser create user
func (s *UserService) CreateUser(user *User) error {
	// check if username already exists
	var existingUser User
	if err := DB.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return errors.New("username already exists")
	}

	// check if email already exists
	if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("email already exists")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	// create user
	if err := DB.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// AuthenticateUser user authentication
func (s *UserService) AuthenticateUser(username, password string) (*User, error) {
	var user User
	if err := DB.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// check if user is active
	if !user.IsActive() {
		return nil, errors.New("user account is not active")
	}

	// validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// update last login time
	now := time.Now()
	user.LastLogin = &now
	DB.Model(&user).Update("last_login", now)

	return &user, nil
}

// GetUserByID get user by id
func (s *UserService) GetUserByID(id uint) (*User, error) {
	var user User
	if err := DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &user, nil
}

// GetUserByUsername get user by username
func (s *UserService) GetUserByUsername(username string) (*User, error) {
	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &user, nil
}

// UpdateUser update user info
func (s *UserService) UpdateUser(user *User) error {
	if err := DB.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	return nil
}

// ChangePassword change password
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	// validate old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// update password
	if err := DB.Model(user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return nil
}

// ListUsers get user list
func (s *UserService) ListUsers(page, pageSize int, search string) ([]*User, int64, error) {
	var users []*User
	var total int64

	query := DB.Model(&User{})

	// search conditions
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// get total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %v", err)
	}

	// paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %v", err)
	}

	// clean sensitive info
	for _, user := range users {
		user.Sanitize()
	}

	return users, total, nil
}

// DeleteUser delete user
func (s *UserService) DeleteUser(id uint) error {
	// first get user info
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}

	// in transaction to delete related data and user
	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// delete user related sessions
	if err := tx.Where("user_id = ?", id).Delete(&UserSession{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user sessions: %v", err)
	}

	// delete user related login logs (optional, depending on whether to keep)
	// if err := tx.Where("user_id = ?", id).Delete(&UserLoginLog{}).Error; err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("failed to delete user login logs: %v", err)
	// }

	// hard delete user
	if err := tx.Unscoped().Delete(user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// CreateSession create user session
func (s *UserService) CreateSession(userID uint) (*UserSession, error) {
	// generate random token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// create session, valid for 24 hours
	session := &UserSession{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := DB.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

// GetSessionByToken get session by token
func (s *UserService) GetSessionByToken(token string) (*UserSession, error) {
	var session UserSession
	if err := DB.Preload("User").Where("token = ?", token).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("session not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// check if session is expired
	if session.IsExpired() {
		s.DeleteSession(session.Token)
		return nil, errors.New("session expired")
	}

	return &session, nil
}

// DeleteSession delete session
func (s *UserService) DeleteSession(token string) error {
	if err := DB.Where("token = ?", token).Delete(&UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}
	return nil
}

// CleanExpiredSessions clean expired sessions
func (s *UserService) CleanExpiredSessions() error {
	if err := DB.Where("expires_at < ?", time.Now()).Delete(&UserSession{}).Error; err != nil {
		return fmt.Errorf("failed to clean expired sessions: %v", err)
	}
	return nil
}

// LogUserLogin log user login
func (s *UserService) LogUserLogin(userID uint, ip, userAgent string, success bool, message string) error {
	log := &UserLoginLog{
		UserID:    userID,
		IP:        ip,
		UserAgent: userAgent,
		Success:   success,
		Message:   message,
	}

	if err := DB.Create(log).Error; err != nil {
		return fmt.Errorf("failed to log user login: %v", err)
	}

	return nil
}

// GetUserLoginLogs get user login logs
func (s *UserService) GetUserLoginLogs(userID uint, page, pageSize int) ([]*UserLoginLog, int64, error) {
	var logs []*UserLoginLog
	var total int64

	query := DB.Model(&UserLoginLog{}).Where("user_id = ?", userID)

	// get total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count login logs: %v", err)
	}

	// paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list login logs: %v", err)
	}

	return logs, total, nil
}

// UpdateUserStatus update user status
func (s *UserService) UpdateUserStatus(userID uint, status UserStatus) error {
	if err := DB.Model(&User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update user status: %v", err)
	}
	return nil
}

// generateToken generate random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateDefaultAdmin create default admin account
func (s *UserService) CreateDefaultAdmin() error {
	// check if admin already exists
	var count int64
	DB.Model(&User{}).Where("role = ?", UserRoleAdmin).Count(&count)
	if count > 0 {
		return nil // admin already exists, no need to create
	}

	// create default admin
	admin := &User{
		Username: "admin",
		Email:    "admin@agent-connector.com",
		Password: "admin123",
		FullName: "System Admin",
		Role:     UserRoleAdmin,
		Status:   UserStatusActive,
	}

	return s.CreateUser(admin)
}
