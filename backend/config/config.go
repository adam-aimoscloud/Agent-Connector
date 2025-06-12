package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config global configuration structure
type Config struct {
	// Application configuration
	App AppConfig `yaml:"app" json:"app"`

	// Database configuration
	Database DatabaseConfig `yaml:"database" json:"database"`

	// Redis configuration
	Redis RedisConfig `yaml:"redis" json:"redis"`

	// Services configuration
	Services ServicesConfig `yaml:"services" json:"services"`

	// Security configuration
	Security SecurityConfig `yaml:"security" json:"security"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// API configuration
	API APIConfig `yaml:"api" json:"api"`
}

// AppConfig application basic configuration
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"` // development, production, staging
	Debug       bool   `yaml:"debug" json:"debug"`
}

// DatabaseConfig database configuration
type DatabaseConfig struct {
	Driver          string        `yaml:"driver" json:"driver"`
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	Username        string        `yaml:"username" json:"username"`
	Password        string        `yaml:"password" json:"password"`
	Database        string        `yaml:"database" json:"database"`
	Charset         string        `yaml:"charset" json:"charset"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	Timezone        string        `yaml:"timezone" json:"timezone"`
}

// RedisConfig Redis configuration
type RedisConfig struct {
	Addr            string        `yaml:"addr" json:"addr"`
	Password        string        `yaml:"password" json:"password"`
	DB              int           `yaml:"db" json:"db"`
	PoolSize        int           `yaml:"pool_size" json:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns" json:"min_idle_conns"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	DialTimeout     time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	KeyPrefix       string        `yaml:"key_prefix" json:"key_prefix"`
}

// ServicesConfig services configuration
type ServicesConfig struct {
	AuthAPI        ServiceConfig `yaml:"auth_api" json:"auth_api"`
	ControlFlowAPI ServiceConfig `yaml:"control_flow_api" json:"control_flow_api"`
	DataFlowAPI    ServiceConfig `yaml:"data_flow_api" json:"data_flow_api"`
}

// ServiceConfig single service configuration
type ServiceConfig struct {
	Host         string        `yaml:"host" json:"host"`
	Port         int           `yaml:"port" json:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	EnableTLS    bool          `yaml:"enable_tls" json:"enable_tls"`
	TLSCertPath  string        `yaml:"tls_cert_path" json:"tls_cert_path"`
	TLSKeyPath   string        `yaml:"tls_key_path" json:"tls_key_path"`
}

// SecurityConfig security configuration
type SecurityConfig struct {
	JWTSecret         string        `yaml:"jwt_secret" json:"jwt_secret"`
	JWTExpiration     time.Duration `yaml:"jwt_expiration" json:"jwt_expiration"`
	PasswordMinLength int           `yaml:"password_min_length" json:"password_min_length"`
	EnableRateLimit   bool          `yaml:"enable_rate_limit" json:"enable_rate_limit"`
	DefaultRateLimit  int           `yaml:"default_rate_limit" json:"default_rate_limit"`
	BcryptCost        int           `yaml:"bcrypt_cost" json:"bcrypt_cost"`
	SessionTimeout    time.Duration `yaml:"session_timeout" json:"session_timeout"`
	MaxLoginAttempts  int           `yaml:"max_login_attempts" json:"max_login_attempts"`
	LockoutDuration   time.Duration `yaml:"lockout_duration" json:"lockout_duration"`
}

// LoggingConfig logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`         // debug, info, warn, error
	Format     string `yaml:"format" json:"format"`       // json, text
	Output     string `yaml:"output" json:"output"`       // stdout, file
	FilePath   string `yaml:"file_path" json:"file_path"` // Log file path
	MaxSize    int    `yaml:"max_size" json:"max_size"`   // MB
	MaxAge     int    `yaml:"max_age" json:"max_age"`     // days
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// APIConfig API related configuration
type APIConfig struct {
	EnableCORS         bool          `yaml:"enable_cors" json:"enable_cors"`
	AllowedOrigins     string        `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods     string        `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders     string        `yaml:"allowed_headers" json:"allowed_headers"`
	MaxRequestBodySize int64         `yaml:"max_request_body_size" json:"max_request_body_size"` // bytes
	RequestTimeout     time.Duration `yaml:"request_timeout" json:"request_timeout"`
	EnableMetrics      bool          `yaml:"enable_metrics" json:"enable_metrics"`
	MetricsPath        string        `yaml:"metrics_path" json:"metrics_path"`
}

// Global configuration instance
var GlobalConfig *Config

// Load loads configuration
func Load() (*Config, error) {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		// .env file not found or failed to load, continue with environment variables
		log.Printf("Warning: .env file not found or failed to load: %v", err)
	} else {
		log.Println("Loaded configuration from .env file")
	}

	// Default configuration
	config := &Config{
		App: AppConfig{
			Name:        "Agent-Connector",
			Version:     "1.0.0",
			Environment: "development",
			Debug:       true,
		},
		Database: DatabaseConfig{
			Driver:          "mysql",
			Host:            "localhost",
			Port:            3306,
			Username:        "root",
			Password:        "",
			Database:        "agent_connector",
			Charset:         "utf8mb4",
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
			SSLMode:         "disable",
			Timezone:        "Asia/Shanghai",
		},
		Redis: RedisConfig{
			Addr:            "localhost:6379",
			Password:        "",
			DB:              0,
			PoolSize:        100,
			MinIdleConns:    10,
			ConnMaxIdleTime: 5 * time.Minute,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			KeyPrefix:       "agent_connector",
		},
		Services: ServicesConfig{
			AuthAPI: ServiceConfig{
				Host:         "localhost",
				Port:         8083,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				EnableTLS:    false,
			},
			ControlFlowAPI: ServiceConfig{
				Host:         "localhost",
				Port:         8081,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				EnableTLS:    false,
			},
			DataFlowAPI: ServiceConfig{
				Host:         "localhost",
				Port:         8082,
				ReadTimeout:  10 * time.Minute,
				WriteTimeout: 10 * time.Minute,
				IdleTimeout:  2 * time.Minute,
				EnableTLS:    false,
			},
		},
		Security: SecurityConfig{
			JWTSecret:         "your-secret-key-please-change-in-production",
			JWTExpiration:     24 * time.Hour,
			PasswordMinLength: 6,
			EnableRateLimit:   true,
			DefaultRateLimit:  1000,
			BcryptCost:        12,
			SessionTimeout:    24 * time.Hour,
			MaxLoginAttempts:  5,
			LockoutDuration:   15 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			FilePath:   "./logs/app.log",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			Compress:   true,
		},
		API: APIConfig{
			EnableCORS:         true,
			AllowedOrigins:     "*",
			AllowedMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowedHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
			MaxRequestBodySize: 10 << 20, // 10MB
			RequestTimeout:     30 * time.Second,
			EnableMetrics:      true,
			MetricsPath:        "/metrics",
		},
	}

	// Load configuration from environment variables
	loadFromEnv(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Application configuration
	if env := os.Getenv("APP_NAME"); env != "" {
		config.App.Name = env
	}
	if env := os.Getenv("APP_VERSION"); env != "" {
		config.App.Version = env
	}
	if env := os.Getenv("APP_ENVIRONMENT"); env != "" {
		config.App.Environment = env
	}
	if env := os.Getenv("APP_DEBUG"); env != "" {
		config.App.Debug = env == "true"
	}

	// Database configuration
	if env := os.Getenv("DB_DRIVER"); env != "" {
		config.Database.Driver = env
	}
	if env := os.Getenv("DB_HOST"); env != "" {
		config.Database.Host = env
	}
	if env := os.Getenv("DB_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			config.Database.Port = port
		}
	}
	if env := os.Getenv("DB_USER"); env != "" {
		config.Database.Username = env
	}
	if env := os.Getenv("DB_PASSWORD"); env != "" {
		config.Database.Password = env
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		config.Database.Database = env
	}

	// Redis configuration
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		config.Redis.Addr = env
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		config.Redis.Password = env
	}
	if env := os.Getenv("REDIS_DB"); env != "" {
		if db, err := strconv.Atoi(env); err == nil {
			config.Redis.DB = db
		}
	}

	// Services configuration
	if env := os.Getenv("AUTH_API_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			config.Services.AuthAPI.Port = port
		}
	}
	if env := os.Getenv("CONTROL_FLOW_API_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			config.Services.ControlFlowAPI.Port = port
		}
	}
	if env := os.Getenv("DATA_FLOW_API_PORT"); env != "" {
		if port, err := strconv.Atoi(env); err == nil {
			config.Services.DataFlowAPI.Port = port
		}
	}

	// Security configuration
	if env := os.Getenv("JWT_SECRET"); env != "" {
		config.Security.JWTSecret = env
	}
}

// validateConfig validates configuration
func validateConfig(config *Config) error {
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if config.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

// GetDSN gets database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.Charset,
	)
}

// GetServiceAddr gets service address
func (c *Config) GetServiceAddr(service string) string {
	var serviceConfig ServiceConfig
	switch strings.ToLower(service) {
	case "auth", "auth-api":
		serviceConfig = c.Services.AuthAPI
	case "control", "control-flow", "control-flow-api":
		serviceConfig = c.Services.ControlFlowAPI
	case "data", "data-flow", "data-flow-api":
		serviceConfig = c.Services.DataFlowAPI
	default:
		return ""
	}

	return fmt.Sprintf("%s:%d", serviceConfig.Host, serviceConfig.Port)
}

// PrintConfig prints configuration information
func (c *Config) PrintConfig() {
	if !c.App.Debug {
		return
	}

	log.Println("=== Agent-Connector Configuration ===")
	log.Printf("App: %s v%s (%s)", c.App.Name, c.App.Version, c.App.Environment)
	log.Printf("Database: %s://%s:%d/%s", c.Database.Driver, c.Database.Host, c.Database.Port, c.Database.Database)
	log.Printf("Redis: %s (DB: %d)", c.Redis.Addr, c.Redis.DB)
	log.Printf("Services:")
	log.Printf("  - Auth API: %s", c.GetServiceAddr("auth"))
	log.Printf("  - Control Flow API: %s", c.GetServiceAddr("control"))
	log.Printf("  - Data Flow API: %s", c.GetServiceAddr("data"))
	log.Println("=====================================")
}
