# DataFlow API Backend Architecture

## 🏗️ 架构概述

新的DataFlow API采用了基于Backend的架构设计，支持不同类型的AI Agent后端，包括OpenAI兼容接口、Dify Chat和Dify Workflow。

## 📁 文件结构

```
backend/api/dataflow/
├── backends/                    # Backend实现
│   ├── interface.go            # Backend接口定义
│   ├── openai.go              # OpenAI兼容后端
│   ├── dify_chat.go           # Dify Chat后端
│   ├── dify_workflow.go       # Dify Workflow后端
│   └── factory.go             # Backend工厂
├── service.go                  # 核心服务层
├── new_handlers.go            # 新的处理器
├── new_routes.go              # 新的路由配置
├── middleware.go              # 中间件
├── auth_service.go            # 认证服务
├── types.go                   # 类型定义
└── utils.go                   # 工具函数
```

## 🔧 Backend类型

### 1. OpenAI Compatible Backend
- **类型**: `openai`, `openai_compatible`
- **端点**: `/v1/chat/completions`
- **请求格式**: OpenAI Chat Completions API
- **支持**: 流式和非流式响应

### 2. Dify Chat Backend
- **类型**: `dify`, `dify-chat`
- **端点**: `/v1/chat-messages`
- **请求格式**: Dify Chat Messages API
- **支持**: 流式和非流式响应

### 3. Dify Workflow Backend
- **类型**: `dify-workflow`
- **端点**: `/v1/workflows/run`
- **请求格式**: Dify Workflow API
- **支持**: 流式和非流式响应

## 🚀 API端点

### 新的Backend路由

#### OpenAI兼容接口
```
POST /api/v1/openai/chat/completions
```

**请求示例**:
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false
}
```

#### Dify Chat接口
```
POST /api/v1/dify/chat-messages
```

**请求示例**:
```json
{
  "query": "Hello!",
  "user": "user123",
  "inputs": {},
  "response_mode": "blocking"
}
```

#### Dify Workflow接口
```
POST /api/v1/dify/workflows/run
```

**请求示例**:
```json
{
  "inputs": {
    "query": "Hello!"
  },
  "user": "user123",
  "response_mode": "blocking"
}
```

### 传统兼容路由
```
POST /api/v1/chat  # 保持向后兼容
```

## 🔄 请求流程

1. **认证中间件**: 验证Agent ID和API Key
2. **限流中间件**: 检查请求频率限制
3. **请求解析**: 根据端点解析不同格式的请求
4. **Backend选择**: 根据Agent类型和请求内容选择合适的Backend
5. **请求验证**: 验证请求参数的有效性
6. **请求转发**: 构建并发送到实际的Agent服务
7. **响应处理**: 处理Agent响应并返回给客户端

## 🎯 Backend选择逻辑

```go
func DetermineBackendType(agentType string, req *BackendRequest) BackendType {
    switch agentType {
    case "openai", "openai_compatible":
        return BackendTypeOpenAI
    case "dify":
        if req.Query != "" {
            return BackendTypeDifyChat
        } else if req.Data != nil {
            return BackendTypeDifyWorkflow
        }
        return BackendTypeDifyChat
    case "dify-chat":
        return BackendTypeDifyChat
    case "dify-workflow":
        return BackendTypeDifyWorkflow
    default:
        return BackendTypeOpenAI
    }
}
```

## 📊 流式响应处理

新架构统一了流式响应的处理：

1. **SSE格式**: 所有流式响应都使用Server-Sent Events格式
2. **统一解析**: 使用`bufio.Scanner`逐行解析响应
3. **格式转换**: 自动处理不同Backend的响应格式差异
4. **错误处理**: 统一的错误处理和客户端通知

## 🔒 认证和授权

- **Agent认证**: 基于Agent ID和API Key
- **请求验证**: 验证请求格式和必需参数
- **权限检查**: 检查Agent是否启用和支持相应功能

## ⚡ 性能优化

- **连接池**: HTTP客户端使用连接池
- **并发处理**: 支持并发请求处理
- **流式传输**: 减少内存使用和延迟
- **Redis限流**: 分布式限流控制

## 🛠️ 使用示例

### 设置路由
```go
// 使用新的Backend架构
dataflow.SetupBackendRoutes(router, rateLimiter)

// 保持向后兼容
dataflow.SetupLegacyRoutes(router, rateLimiter)
```

### 创建自定义Backend
```go
type CustomBackend struct{}

func (b *CustomBackend) GetType() BackendType {
    return "custom"
}

func (b *CustomBackend) ValidateRequest(req *BackendRequest) error {
    // 实现验证逻辑
    return nil
}

func (b *CustomBackend) BuildForwardRequest(ctx context.Context, req *BackendRequest, agentInfo *AgentInfo) (*http.Request, error) {
    // 实现请求构建逻辑
    return nil, nil
}

// 实现其他接口方法...
```

## 🔍 监控和调试

- **日志记录**: 详细的请求和响应日志
- **错误追踪**: 统一的错误处理和报告
- **性能指标**: 请求延迟和成功率统计
- **健康检查**: `/api/v1/health`端点

## 🚧 迁移指南

### 从旧架构迁移

1. **保持兼容**: 旧的`/api/v1/chat`端点继续工作
2. **逐步迁移**: 可以逐步将客户端迁移到新端点
3. **配置更新**: 更新Agent配置以指定具体的Backend类型

### 客户端更新

- **OpenAI客户端**: 使用`/api/v1/openai/chat/completions`
- **Dify客户端**: 使用`/api/v1/dify/chat-messages`或`/api/v1/dify/workflows/run`
- **自定义客户端**: 根据Agent类型选择合适的端点

## 📈 未来扩展

- **新Backend支持**: 可以轻松添加新的AI服务Backend
- **协议扩展**: 支持更多的API协议和格式
- **插件系统**: 支持自定义处理插件
- **负载均衡**: 支持多实例负载均衡 