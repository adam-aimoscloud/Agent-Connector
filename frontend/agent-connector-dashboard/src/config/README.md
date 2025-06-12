# API Configuration Documentation

## Overview

This directory contains all configuration files for frontend application communication with backend services. The configuration system supports environment variables, different environment configurations, and runtime configuration.

## Configuration File Structure

```
src/config/
├── api.config.ts              # Main configuration file
├── environments/              # Environment-specific configurations
│   ├── development.config.ts  # Development environment configuration
│   └── production.config.ts   # Production environment configuration
└── README.md                  # Configuration documentation
```

## Main Configuration Files

### `api.config.ts`

Main configuration file containing:
- Backend service configuration (authentication, control flow, data flow)
- Global API configuration
- Authentication configuration
- Pagination configuration

### Environment Variables

All configurations support override through environment variables. Create a `.env` file in the project root:

```bash
# Copy example configuration file
cp .env.example .env
```

## Backend Service Configuration

### Authentication Service (Port 8083)
- **Purpose**: User authentication, user management, system statistics
- **Environment Variable**: `REACT_APP_AUTH_API_URL`
- **Default Address**: `http://localhost:8083`

### Control Flow Service (Port 8081)
- **Purpose**: Agent configuration, system configuration
- **Environment Variable**: `REACT_APP_CONTROL_FLOW_API_URL`
- **Default Address**: `http://localhost:8081`

### Data Flow Service (Port 8082)
- **Purpose**: Rate limiting configuration management
- **Environment Variable**: `REACT_APP_DATA_FLOW_API_URL`
- **Default Address**: `http://localhost:8082`

## Environment Variable Configuration

### Service Address Configuration

```bash
# Development environment
REACT_APP_AUTH_API_URL=http://localhost:8083
REACT_APP_CONTROL_FLOW_API_URL=http://localhost:8081
REACT_APP_DATA_FLOW_API_URL=http://localhost:8082

# Production environment
REACT_APP_AUTH_API_URL=https://api.yourcompany.com:8083
REACT_APP_CONTROL_FLOW_API_URL=https://control-api.yourcompany.com:8081
REACT_APP_DATA_FLOW_API_URL=https://data-api.yourcompany.com:8082
```

### Timeout and Retry Configuration

```bash
# API timeout settings (milliseconds)
REACT_APP_AUTH_API_TIMEOUT=10000
REACT_APP_CONTROL_FLOW_API_TIMEOUT=10000
REACT_APP_DATA_FLOW_API_TIMEOUT=10000

# Retry configuration
REACT_APP_API_DEFAULT_RETRY_ATTEMPTS=3
REACT_APP_API_DEFAULT_RETRY_DELAY=1000
```

### Debug Configuration

```bash
# Enable request/response logging (recommended for development)
REACT_APP_ENABLE_REQUEST_LOGGING=true
REACT_APP_ENABLE_RESPONSE_LOGGING=true
```

### Authentication Configuration

```bash
# Local storage key names
REACT_APP_AUTH_TOKEN_KEY=auth_token
REACT_APP_AUTH_USER_INFO_KEY=user_info

# Token expiration time (hours)
REACT_APP_AUTH_TOKEN_EXPIRATION_HOURS=24

# Auto refresh token
REACT_APP_AUTH_AUTO_REFRESH_TOKEN=true
```

### Pagination Configuration

```bash
# Pagination settings
REACT_APP_PAGINATION_DEFAULT_PAGE_SIZE=10
REACT_APP_PAGINATION_MAX_PAGE_SIZE=100
```

## Usage

### Using Configuration in Code

```typescript
import { apiConfig, getServiceConfig, getAuthConfig } from '../config/api.config';

// Get service configuration
const authConfig = getServiceConfig('auth');
console.log('Authentication service address:', authConfig.baseURL);

// Get authentication configuration
const authSettings = getAuthConfig();
console.log('Token key name:', authSettings.tokenKey);

// Get complete configuration
console.log('Complete configuration:', apiConfig);
```

### Debug Configuration

In development environment, configuration is automatically printed to console:

```
🔧 API Configuration
Auth Service: http://localhost:8083
Control Flow Service: http://localhost:8081
Data Flow Service: http://localhost:8082
Global Timeout: 10000
Request Logging: false
Response Logging: false
```

## Deployment Configuration

### Docker Deployment

In Docker environment, configuration can be passed through environment variables:

```dockerfile
ENV REACT_APP_AUTH_API_URL=https://api.production.com:8083
ENV REACT_APP_CONTROL_FLOW_API_URL=https://control-api.production.com:8081
ENV REACT_APP_DATA_FLOW_API_URL=https://data-api.production.com:8082
```

### Kubernetes Deployment

Use ConfigMap or Secret to manage configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-config
data:
  REACT_APP_AUTH_API_URL: "https://api.production.com:8083"
  REACT_APP_CONTROL_FLOW_API_URL: "https://control-api.production.com:8081"
  REACT_APP_DATA_FLOW_API_URL: "https://data-api.production.com:8082"
```

## Troubleshooting

### Connection Issues

1. Check if backend services are running
2. Verify network connection and firewall settings
3. Validate URL and port configuration
4. Check CORS settings

### Configuration Issues

1. Ensure environment variable format is correct
2. Check if `.env` file is in the correct location
3. Restart development server for configuration to take effect
4. Check browser console for configuration logs

### Debug Suggestions

1. Enable request/response logging:
   ```bash
   REACT_APP_ENABLE_REQUEST_LOGGING=true
   REACT_APP_ENABLE_RESPONSE_LOGGING=true
   ```

2. Check Network tab to view actual requests
3. Use browser developer tools to check errors

## Best Practices

1. **Environment Separation**: Use different configurations for different environments
2. **Security**: Never commit sensitive information like API keys to version control
3. **Validation**: Validate configuration on application startup
4. **Documentation**: Keep configuration documentation up to date
5. **Monitoring**: Monitor configuration changes in production

## Configuration Examples

### Development Environment
```bash
# .env.development
REACT_APP_AUTH_API_URL=http://localhost:8083
REACT_APP_CONTROL_FLOW_API_URL=http://localhost:8081
REACT_APP_DATA_FLOW_API_URL=http://localhost:8082
REACT_APP_ENABLE_REQUEST_LOGGING=true
REACT_APP_ENABLE_RESPONSE_LOGGING=true
```

### Production Environment
```bash
# .env.production
REACT_APP_AUTH_API_URL=https://api.production.com:8083
REACT_APP_CONTROL_FLOW_API_URL=https://control-api.production.com:8081
REACT_APP_DATA_FLOW_API_URL=https://data-api.production.com:8082
REACT_APP_ENABLE_REQUEST_LOGGING=false
REACT_APP_ENABLE_RESPONSE_LOGGING=false
```

### Testing Environment
```bash
# .env.test
REACT_APP_AUTH_API_URL=https://api.test.com:8083
REACT_APP_CONTROL_FLOW_API_URL=https://control-api.test.com:8081
REACT_APP_DATA_FLOW_API_URL=https://data-api.test.com:8082
``` 