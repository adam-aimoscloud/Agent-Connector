# Agent-Connector Dashboard

## Project Overview

Agent-Connector Dashboard is a modern management platform built with React + TypeScript + Ant Design, designed for managing third-party AI service integration, user permission control, and API rate limiting configuration.

## 🚀 Main Features

### 1. User Authentication System
- **Login/Registration**: Support for username/password login with form validation and error handling
- **Permission Management**: Role-based access control system
  - `admin`: Administrator - Full permissions
  - `operator`: Operator - Partial management permissions  
  - `user`: Regular User - Basic functionality permissions
  - `readonly`: Read-only User - View-only permissions
- **Session Management**: JWT Token authentication with automatic refresh mechanism

### 2. User Management 📊
- **User List**: Paginated display, search filtering, status management
- **CRUD Operations**: Create, edit, delete users
- **Status Control**: Real-time user status switching (Active/Inactive/Blocked/Pending)
- **Detailed Information**: User details view, login history records

### 3. Agent Configuration Management 🤖
- **Multi-type Support**: OpenAI, Dify, Custom third-party service integration
- **Configuration Management**: API endpoints, key management, model selection
- **Connector Keys**: Automatic generation of unified access keys
- **Streaming Response**: Support for real-time streaming output configuration
- **Status Monitoring**: Real-time Agent running status monitoring

### 4. Rate Limiting Configuration ⚡
- **Multi-dimensional Rate Limiting**: 
  - Requests per minute (requests_per_minute)
  - Tokens per minute (tokens_per_minute)  
  - Requests per hour (requests_per_hour)
  - Daily requests/tokens (requests_per_day/tokens_per_day)
- **Application Scope**: Global, user, Agent, IP address level restrictions
- **Usage Monitoring**: Real-time usage statistics and progress display
- **Priority Management**: Automatic rule priority management

### 5. System Settings ⚙️
- **System Overview**: CPU, memory, disk usage monitoring
- **Service Status**: Microservice running status monitoring
- **Statistics**: User count, request statistics, success rate analysis
- **System Maintenance**: Expired session cleanup, data backup management

### 6. Personal Profile 👤
- **Information Management**: Personal information editing, avatar upload
- **Security Settings**: Password change, security alerts
- **Login History**: Login history record viewing

## 🛠 Technology Stack

### Frontend Framework
- **React 18**: Latest version React framework
- **TypeScript**: Type-safe JavaScript
- **Ant Design**: Enterprise-class UI design language

### State Management
- **React Context + useReducer**: Global state management
- **React Hooks**: Modern state and lifecycle management

### Network Requests
- **Axios**: HTTP client with request/response interceptors
- **Multi-service Endpoints**: Support for authentication(8083), control flow(8081), data flow(8082) three services

### Development Tools
- **Create React App**: React application scaffolding
- **ESLint + TypeScript**: Code standards and type checking
- **dayjs**: Modern date processing library

## 📁 Project Structure

```
src/
├── components/           # Shared components
│   └── Layout/          # Layout components
├── contexts/            # React contexts
│   └── AuthContext.tsx  # Authentication context
├── pages/               # Page components
│   ├── Login.tsx        # Login/Registration page
│   ├── Dashboard.tsx    # Dashboard homepage
│   ├── Users.tsx        # User management
│   ├── Agents.tsx       # Agent configuration
│   ├── RateLimits.tsx   # Rate limiting configuration
│   ├── Profile.tsx      # Personal profile
│   └── SystemSettings.tsx # System settings
├── services/            # API services
│   └── api.ts          # API interface definitions
└── App.tsx             # Main application component
```

## 🚦 API Interfaces

### Authentication Service (Port 8083)
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/auth/profile` - Get personal profile
- `PUT /api/v1/auth/profile` - Update personal profile
- `POST /api/v1/auth/change-password` - Change password

### User Management (Port 8083)
- `GET /api/v1/users` - Get user list
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Control Flow Service (Port 8081)
- `GET /api/v1/controlflow/agents` - Get Agent list
- `POST /api/v1/controlflow/agents` - Create Agent
- `PUT /api/v1/controlflow/agents/:id` - Update Agent
- `DELETE /api/v1/controlflow/agents/:id` - Delete Agent

### Data Flow Service (Port 8082)
- `GET /api/v1/dataflow/rate-limits` - Get rate limiting rules
- `POST /api/v1/dataflow/rate-limits` - Create rate limiting rules
- `PUT /api/v1/dataflow/rate-limits/:id` - Update rate limiting rules
- `DELETE /api/v1/dataflow/rate-limits/:id` - Delete rate limiting rules

## 🔧 Development Guide

### Environment Requirements
- Node.js 16+ 
- npm 8+ or yarn 1.22+

### Install Dependencies
```bash
npm install
# Or use Tencent Cloud mirror
npm install --registry https://mirrors.cloud.tencent.com/npm/
```

### Development Run
```bash
npm start
# Application will start at http://localhost:3000
```

### Production Build
```bash
npm run build
# Build files will be output to build/ directory
```

### Type Check
```bash
npm run type-check
```

## 🔐 Default Login Account

- **Username**: admin
- **Password**: admin123
- **Role**: Administrator

## 🎨 UI/UX Features

### Responsive Design
- Mobile-friendly responsive layout
- Breakpoint adaptation: xs(≤576px), sm(≥576px), md(≥768px), lg(≥992px), xl(≥1200px)

### User Experience
- Loading status indicators
- Error boundaries and error handling
- Form validation and friendly prompts
- Data pagination and search
- Real-time status updates

### Internationalization
- English interface
- Date formatting
- Number thousand separator formatting

## 🔄 State Management

### Authentication State
```typescript
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}
```

### Permission Control
```typescript
// Permission check
hasPermission(permission: string): boolean
hasRole(role: string | string[]): boolean

// Permission component
<PermissionGuard permission="user_management">
  <AdminOnlyComponent />
</PermissionGuard>
```

## 🚀 Deployment Recommendations

### Environment Variables
```bash
# Backend service API address configuration
REACT_APP_AUTH_API_URL=http://localhost:8083
REACT_APP_CONTROL_FLOW_API_URL=http://localhost:8081
REACT_APP_DATA_FLOW_API_URL=http://localhost:8082
```

### Production Deployment
```bash
# Build production version
npm run build

# Deploy to static file server
# Copy build/ directory contents to web server
```

### Docker Deployment
```dockerfile
FROM nginx:alpine
COPY build/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 📝 Development Notes

### Code Standards
- Use TypeScript for type safety
- Follow ESLint configuration
- Use Ant Design components consistently
- Implement proper error handling

### Performance Optimization
- Use React.memo for component optimization
- Implement virtual scrolling for large lists
- Use lazy loading for route components
- Optimize bundle size with code splitting

### Security Considerations
- Validate all user inputs
- Implement proper authentication checks
- Use HTTPS in production
- Sanitize data before display

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
