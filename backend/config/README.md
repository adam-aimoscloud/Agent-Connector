# Agent-Connector Backend Configuration System

## Overview

Agent-Connector backend adopts a unified configuration management system, centrally managing MySQL, Redis, and various service configurations, supporting environment variable override for easy deployment and maintenance.

## Configuration Structure

### Configuration Files
- `config/config.go` - Main configuration manager
- `.env` - Environment variable configuration file
- `.env.example` - Environment variable template

### Configuration Categories

#### 1. Application Configuration (App)
```yaml
app:
  name: "Agent-Connector"
  version: "1.0.0"
  environment: "development"  # development, production, staging
  debug: true
```

#### 2. Database Configuration (Database)
```yaml
database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "123456"
  database: "agent_connector"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "10m"
  ssl_mode: "disable"
  timezone: "Asia/Shanghai"
```

#### 3. Redis Configuration (Redis)
```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 100
  min_idle_conns: 10
  conn_max_idle_time: "5m"
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  key_prefix: "agent_connector"
```

#### 4. Service Configuration (Services)
```yaml
services:
  auth_api:
    host: "localhost"
    port: 8083
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    enable_tls: false
  
  control_flow_api:
    host: "localhost"
    port: 8081
    read_timeout: "30s"
    write_timeout: "30s"
    idle_timeout: "60s"
    enable_tls: false
  
  data_flow_api:
    host: "localhost"
    port: 8082
    read_timeout: "10m"
    write_timeout: "10m"
    idle_timeout: "2m"
    enable_tls: false
```

#### 5. Security Configuration (Security)
```yaml
security:
  jwt_secret: "your-secret-key-change-in-production"
  jwt_expiration: "24h"
  password_min_length: 6
  enable_rate_limit: true
  default_rate_limit: 1000
  bcrypt_cost: 12
  session_timeout: "24h"
  max_login_attempts: 5
  lockout_duration: "15m"
```

#### 6. Logging Configuration (Logging)
```yaml
logging:
  level: "info"          # debug, info, warn, error
  format: "text"         # json, text
  output: "stdout"       # stdout, file
  file_path: "./logs/app.log"
  max_size: 100          # MB
  max_age: 30           # days
  max_backups: 10
  compress: true
```

#### 7. API Configuration (API)
```yaml
api:
  enable_cors: true
  allowed_origins: "*"
  allowed_methods: "GET,POST,PUT,DELETE,OPTIONS"
  allowed_headers: "Origin,Content-Type,Accept,Authorization,X-API-Key"
  max_request_body_size: 10485760  # 10MB
  request_timeout: "30s"
  enable_metrics: true
  metrics_path: "/metrics"
```

## Environment Variables

### Basic Configuration
```bash
# Application configuration
APP_NAME=Agent-Connector
APP_VERSION=1.0.0
APP_ENVIRONMENT=development
APP_DEBUG=true

# Database configuration
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=123456
DB_NAME=agent_connector

# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Service port configuration
AUTH_API_PORT=8083
CONTROL_FLOW_API_PORT=8081
DATA_FLOW_API_PORT=8082

# Security configuration
JWT_SECRET=your-secret-key-change-in-production
```

### Production Environment Configuration Example
```bash
# Production environment
APP_ENVIRONMENT=production
APP_DEBUG=false

# Database
DB_HOST=mysql-server.example.com
DB_PASSWORD=your-strong-password

# Redis
REDIS_ADDR=redis-server.example.com:6379
REDIS_PASSWORD=your-redis-password

# Security
JWT_SECRET=your-very-secure-jwt-secret-key-at-least-32-characters
```

## Usage

### 1. Basic Usage
```go
import "agent-connector/config"

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Use configuration
dsn := cfg.GetDSN()
serviceAddr := cfg.GetServiceAddr("auth")
```

### 2. Get Database Connection String
```go
// Automatically generate DSN based on driver type
dsn := cfg.GetDSN()
```

### 3. Get Service Address
```go
// Get authentication service address
authAddr := cfg.GetServiceAddr("auth")

// Get control flow service address
controlFlowAddr := cfg.GetServiceAddr("control_flow")

// Get data flow service address
dataFlowAddr := cfg.GetServiceAddr("data_flow")
```

### 4. Configuration Validation
```go
// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal("Configuration validation failed:", err)
}
```

### 5. Environment-specific Configuration
```go
// Check if running in production
if cfg.App.Environment == "production" {
    // Production-specific logic
}

// Check if debug mode is enabled
if cfg.App.Debug {
    // Debug-specific logic
}
```

## Configuration Loading Priority

1. **Default Values**: Built-in default configuration
2. **Configuration File**: Values from config files
3. **Environment Variables**: Override with environment variables
4. **Command Line Arguments**: Highest priority (if implemented)

## Environment Variable Mapping

| Configuration Path | Environment Variable | Default Value |
|-------------------|---------------------|---------------|
| `app.name` | `APP_NAME` | "Agent-Connector" |
| `app.environment` | `APP_ENVIRONMENT` | "development" |
| `database.host` | `DB_HOST` | "localhost" |
| `database.port` | `DB_PORT` | 3306 |
| `database.username` | `DB_USER` | "root" |
| `database.password` | `DB_PASSWORD` | "" |
| `redis.addr` | `REDIS_ADDR` | "localhost:6379" |
| `redis.password` | `REDIS_PASSWORD` | "" |
| `security.jwt_secret` | `JWT_SECRET` | "" |

## Configuration Validation

The system automatically validates configuration on startup:

### Required Fields
- Database connection parameters
- JWT secret (in production)
- Service ports

### Validation Rules
- Port numbers must be between 1-65535
- JWT secret must be at least 32 characters in production
- Database connection must be testable
- Redis connection must be available

## Best Practices

### 1. Environment Separation
```bash
# Development
APP_ENVIRONMENT=development
APP_DEBUG=true

# Production
APP_ENVIRONMENT=production
APP_DEBUG=false
```

### 2. Security
```bash
# Use strong passwords
DB_PASSWORD=your-very-strong-database-password

# Use secure JWT secret
JWT_SECRET=your-very-secure-jwt-secret-key-at-least-32-characters-long

# Enable TLS in production
ENABLE_TLS=true
```

### 3. Performance Tuning
```bash
# Database connection pool
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10

# Redis connection pool
REDIS_POOL_SIZE=100
REDIS_MIN_IDLE_CONNS=10
```

### 4. Monitoring
```bash
# Enable metrics
ENABLE_METRICS=true
METRICS_PATH=/metrics

# Logging configuration
LOG_LEVEL=info
LOG_FORMAT=json
```

## Docker Configuration

### Environment Variables in Docker
```dockerfile
ENV APP_ENVIRONMENT=production
ENV DB_HOST=mysql-container
ENV REDIS_ADDR=redis-container:6379
ENV JWT_SECRET=your-production-jwt-secret
```

### Docker Compose
```yaml
version: '3.8'
services:
  backend:
    image: agent-connector:latest
    environment:
      - APP_ENVIRONMENT=production
      - DB_HOST=mysql
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mysql
      - redis
```

## Kubernetes Configuration

### ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-connector-config
data:
  APP_ENVIRONMENT: "production"
  DB_HOST: "mysql-service"
  REDIS_ADDR: "redis-service:6379"
```

### Secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: agent-connector-secrets
type: Opaque
data:
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check database host and port
   - Verify credentials
   - Ensure database exists

2. **Redis Connection Failed**
   - Check Redis server status
   - Verify connection string
   - Check network connectivity

3. **Configuration Validation Failed**
   - Check required environment variables
   - Validate configuration format
   - Review error messages

### Debug Mode

Enable debug mode for detailed logging:
```bash
APP_DEBUG=true
LOG_LEVEL=debug
```

### Health Checks

The system provides health check endpoints:
- `/health` - Basic health check
- `/health/db` - Database connectivity
- `/health/redis` - Redis connectivity

## Migration Guide

### From v1.0 to v2.0
1. Update environment variable names
2. Add new required configuration
3. Update Docker/Kubernetes manifests

### Configuration Schema Changes
- Check CHANGELOG.md for breaking changes
- Update configuration files accordingly
- Test in staging environment first 