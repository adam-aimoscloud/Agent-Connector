package internal

import (
	"agent-connector/config"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initialize MySQL database
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	return nil
}

// InitDatabase initialize database connection
func InitDatabase() error {
	// load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// print configuration information
	cfg.PrintConfig()

	// get database connection string
	dsn := cfg.GetDSN()

	// create GORM instance
	logLevel := logger.Info
	if cfg.App.Environment == "production" {
		logLevel = logger.Error
	}

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// automatically migrate table structures
	err = DB.AutoMigrate(
		&User{},
		&UserSession{},
		&UserLoginLog{},
		&SystemConfig{},
		&UserRateLimit{},
		&Agent{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// initialize default system configuration
	if err := initDefaultSystemConfig(); err != nil {
		log.Printf("Warning: failed to init default system config: %v", err)
	}

	// create default admin account
	userService := NewUserService()
	if err := userService.CreateDefaultAdmin(); err != nil {
		log.Printf("Warning: Failed to create default admin: %v", err)
	} else {
		log.Println("Default admin account created (username: admin, password: admin123)")
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

// initDefaultSystemConfig initialize default system configuration
func initDefaultSystemConfig() error {
	var count int64
	DB.Model(&SystemConfig{}).Count(&count)

	if count == 0 {
		defaultConfig := &SystemConfig{
			RateLimitMode:   RateLimitModePriority,
			DefaultPriority: 5,
			DefaultQPS:      10,
		}
		return DB.Create(defaultConfig).Error
	}

	return nil
}
