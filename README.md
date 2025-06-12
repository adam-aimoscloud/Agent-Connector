# Agent-Connector

A comprehensive AI agent management platform with React frontend and Go backend, supporting multiple AI service providers with unified interfaces, load balancing, and rate limiting.

## Project Structure

```
Agent-Connector/
├── frontend/                    # React TypeScript frontend
│   └── agent-connector-dashboard/
├── backend/                     # Go backend services
│   ├── cmd/                    # Application entry points
│   ├── api/                    # API handlers and routes
│   ├── internal/               # Internal packages
│   ├── pkg/                    # Reusable packages
│   └── config/                 # Configuration management
├── .gitignore                  # Git ignore rules
└── README.md                   # This file
```

## Features

### Frontend (React Dashboard)
- **User Management**: Role-based access control and user administration
- **Agent Configuration**: Multi-provider AI agent setup (OpenAI, Dify, Custom)
- **Rate Limiting**: Comprehensive rate limiting configuration and monitoring
- **System Monitoring**: Real-time system status and performance metrics
- **Responsive Design**: Modern UI with Ant Design components

### Backend (Go Services)
- **Authentication API** (Port 8083): User authentication and management
- **Control Flow API** (Port 8081): Agent configuration and management
- **Data Flow API** (Port 8082): Rate limiting and data flow control
- **Unified Configuration**: Environment-based configuration management
- **Health Monitoring**: Built-in health checks and metrics

## Quick Start

### Prerequisites
- **Frontend**: Node.js 16+, npm 8+
- **Backend**: Go 1.21+, MySQL 8.0+, Redis 6.0+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/adam-aimoscloud/Agent-Connector.git
   cd Agent-Connector
   ```

2. **Database Setup**
   ```bash
   # Start MySQL and Redis (using Docker)
   docker run -d --name mysql-agent \
     -e MYSQL_ROOT_PASSWORD=123456 \
     -e MYSQL_DATABASE=agent_connector \
     -p 3306:3306 mysql:8.0
   
   docker run -d --name redis-agent \
     -p 6379:6379 redis:7-alpine
   ```

3. **Backend Setup**
   ```bash
   cd backend
   
   # Copy and configure environment variables
   cp .env.example .env
   # Edit .env file with your database and Redis configurations
   
   # Download dependencies
   go mod download
   
   # Build all services
   go build -o bin/auth-api ./cmd/auth-api/
   go build -o bin/control-flow-api ./cmd/control-flow-api/
   go build -o bin/dataflow-api ./cmd/dataflow-api/
   ```

4. **Start Backend Services**
   
   **Option 1: Start all services using scripts**
   ```bash
   # Make scripts executable
   chmod +x scripts/*.sh
   
   # Start all services
   ./scripts/start-all.sh
   ```
   
   **Option 2: Start services individually**
   ```bash
   # Terminal 1: Start Authentication API
   go run cmd/auth-api/main.go
   # or
   ./bin/auth-api
   
   # Terminal 2: Start Control Flow API
   go run cmd/control-flow-api/main.go
   # or
   ./bin/control-flow-api
   
   # Terminal 3: Start Data Flow API
   go run cmd/dataflow-api/main.go
   # or
   ./bin/dataflow-api
   ```
   
   **Option 3: Start services in background**
   ```bash
   # Start all services in background
   nohup ./bin/auth-api > logs/auth-api.log 2>&1 &
   nohup ./bin/control-flow-api > logs/control-flow-api.log 2>&1 &
   nohup ./bin/dataflow-api > logs/dataflow-api.log 2>&1 &
   
   # Check if services are running
   ps aux | grep -E "(auth-api|control-flow-api|dataflow-api)"
   ```

5. **Frontend Setup**
   ```bash
   cd frontend/agent-connector-dashboard
   
   # Install dependencies
   npm install
   
   # Copy and configure environment variables
   cp .env.example .env
   # Edit .env file if needed
   
   # Start development server
   npm start
   ```

6. **Verify Services**
   ```bash
   # Check backend services health
   curl http://localhost:8083/health  # Auth API
   curl http://localhost:8081/health  # Control Flow API
   curl http://localhost:8082/health  # Data Flow API
   
   # Frontend should be available at http://localhost:3000
   ```

7. **Access the application**
   - **Frontend Dashboard**: http://localhost:3000
   - **Auth API**: http://localhost:8083
   - **Control Flow API**: http://localhost:8081
   - **Data Flow API**: http://localhost:8082

### Default Login
- **Username**: admin
- **Password**: admin123

### Stop Services
```bash
# Stop all backend services
cd backend
./scripts/stop-all.sh

# Or kill individual processes
pkill -f auth-api
pkill -f control-flow-api
pkill -f dataflow-api

# Stop frontend (Ctrl+C in the terminal where npm start is running)
```

## Configuration

### Environment Variables
The project uses environment variables for configuration. See:
- Backend: `backend/.env.example`
- Frontend: `frontend/agent-connector-dashboard/.env.example`

### Service Ports
- **Auth API**: 8083
- **Control Flow API**: 8081  
- **Data Flow API**: 8082
- **Frontend Dev Server**: 3000

## Documentation

- [Frontend Documentation](frontend/agent-connector-dashboard/README.md)
- [Backend Configuration](backend/config/README.md)
- [API Configuration](frontend/agent-connector-dashboard/src/config/README.md)
- [Git Commit Guidelines](git-commit-guidelines.md)

## Development

### Git Workflow
This project includes a comprehensive `.gitignore` file that handles:
- Go binaries and build artifacts
- Node.js dependencies and build outputs
- Environment variables and secrets
- IDE and editor files
- Operating system specific files
- Log files and temporary data

### Code Structure
- **Frontend**: React + TypeScript + Ant Design
- **Backend**: Go with clean architecture
- **Database**: MySQL for persistent data
- **Cache**: Redis for session and rate limiting
- **Configuration**: Environment-based with validation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 