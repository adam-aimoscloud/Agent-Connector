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
   git clone <repository-url>
   cd Agent-Connector
   ```

2. **Backend Setup**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your database and Redis configurations
   go mod download
   go run cmd/auth-api/main.go
   ```

3. **Frontend Setup**
   ```bash
   cd frontend/agent-connector-dashboard
   npm install
   npm start
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend APIs: http://localhost:8083, http://localhost:8081, http://localhost:8082

### Default Login
- **Username**: admin
- **Password**: admin123

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