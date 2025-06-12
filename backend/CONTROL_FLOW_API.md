# 控制流 API 接口文档

控制流 API 为 Dashboard 界面提供 HTTP 接口，用于配置用户限流和 Agent 管理。

## 快速开始

### 环境变量配置

创建 `.env` 文件或设置以下环境变量：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=agent_connector

# 服务配置
PORT=8080
GIN_MODE=release  # production 环境设置为 release
```

### 启动服务

```bash
cd backend
go run cmd/control-flow-api/main.go
```

服务启动后访问：
- 健康检查：http://localhost:8080/health
- API 基础路径：http://localhost:8080/api/v1

## API 接口

### 1. 系统配置 API

#### 1.1 获取系统配置

```http
GET /api/v1/system/config
```

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "rate_limit_mode": "priority",
    "default_priority": 5,
    "default_qps": 10,
    "created_at": "2024-06-10T10:00:00Z",
    "updated_at": "2024-06-10T10:00:00Z"
  }
}
```

#### 1.2 更新系统配置

```http
PUT /api/v1/system/config
```

**请求体：**
```json
{
  "rate_limit_mode": "priority",
  "default_priority": 5,
  "default_qps": 10
}
```

**字段说明：**
- `rate_limit_mode`: 限流模式，可选值：`priority`（优先级模式）、`qps`（QPS模式）
- `default_priority`: 优先级模式下用户的默认优先级（1-10）
- `default_qps`: QPS模式下用户的默认QPS限制

### 2. 用户限流配置 API

#### 2.1 获取用户限流配置列表

```http
GET /api/v1/user-rate-limits?page=1&page_size=10
```

**查询参数：**
- `page`: 页码，默认为1
- `page_size`: 每页大小，默认为10，最大为100

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "user_id": "user123",
      "priority": 8,
      "qps": null,
      "enabled": true,
      "created_at": "2024-06-10T10:00:00Z",
      "updated_at": "2024-06-10T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_page": 1
  }
}
```

#### 2.2 获取单个用户限流配置

```http
GET /api/v1/user-rate-limits/:user_id
```

**路径参数：**
- `user_id`: 用户ID

#### 2.3 创建用户限流配置

```http
POST /api/v1/user-rate-limits
```

**请求体：**
```json
{
  "user_id": "user123",
  "priority": 8,
  "enabled": true
}
```

**字段说明：**
- `user_id`: 用户ID（必填）
- `priority`: 优先级（1-10），优先级模式下使用
- `qps`: QPS限制，QPS模式下使用
- `enabled`: 是否启用，默认为true

#### 2.4 更新用户限流配置

```http
PUT /api/v1/user-rate-limits/:user_id
```

**请求体：**
```json
{
  "priority": 9,
  "enabled": true
}
```

#### 2.5 删除用户限流配置

```http
DELETE /api/v1/user-rate-limits/:user_id
```

### 3. Agent 配置 API

#### 3.1 获取 Agent 列表

```http
GET /api/v1/agents?page=1&page_size=10&type=openai
```

**查询参数：**
- `page`: 页码，默认为1
- `page_size`: 每页大小，默认为10，最大为100
- `type`: Agent类型过滤，可选值：`dify`、`openai`、`openai_compatible`

**响应示例：**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "OpenAI GPT-4",
      "type": "openai",
      "url": "https://api.openai.com",
      "api_key": "sk-***",
      "qps": 10,
      "enabled": true,
      "description": "OpenAI GPT-4 模型",
      "created_at": "2024-06-10T10:00:00Z",
      "updated_at": "2024-06-10T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_page": 1
  }
}
```

#### 3.2 获取单个 Agent

```http
GET /api/v1/agents/:id
```

**路径参数：**
- `id`: Agent ID

#### 3.3 创建 Agent

```http
POST /api/v1/agents
```

**请求体：**
```json
{
  "name": "OpenAI GPT-4",
  "type": "openai",
  "url": "https://api.openai.com",
  "api_key": "sk-your-api-key",
  "qps": 10,
  "enabled": true,
  "description": "OpenAI GPT-4 模型"
}
```

**字段说明：**
- `name`: Agent名称（必填）
- `type`: 平台类型，可选值：`dify`、`openai`、`openai_compatible`（必填）
- `url`: 访问URL（必填）
- `api_key`: API密钥（必填）
- `qps`: Agent的QPS限制（必填，大于0）
- `enabled`: 是否启用，默认为true
- `description`: 描述信息

#### 3.4 更新 Agent

```http
PUT /api/v1/agents/:id
```

**请求体：**
```json
{
  "name": "Updated OpenAI GPT-4",
  "type": "openai",
  "url": "https://api.openai.com",
  "api_key": "sk-updated-api-key",
  "qps": 15,
  "enabled": true,
  "description": "更新的 OpenAI GPT-4 模型"
}
```

#### 3.5 删除 Agent

```http
DELETE /api/v1/agents/:id
```

**注意：** 删除操作为软删除，Agent记录不会从数据库中彻底删除。

## 响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 分页响应

```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ],
  "pagination": {  
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_page": 10
  }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "error description"
}
```

## HTTP 状态码

- `200`: 操作成功
- `201`: 创建成功
- `400`: 请求参数错误
- `404`: 资源不存在
- `500`: 服务器内部错误

## 数据库表结构

### system_configs 表
- `id`: 主键
- `rate_limit_mode`: 限流模式（priority/qps）
- `default_priority`: 默认优先级（1-10）
- `default_qps`: 默认QPS限制
- `created_at`: 创建时间
- `updated_at`: 更新时间

### user_rate_limits 表
- `id`: 主键
- `user_id`: 用户ID（唯一索引）
- `priority`: 优先级（可空，优先级模式使用）
- `qps`: QPS限制（可空，QPS模式使用）
- `enabled`: 是否启用
- `created_at`: 创建时间
- `updated_at`: 更新时间

### agents 表
- `id`: 主键
- `name`: Agent名称
- `type`: 平台类型
- `url`: 访问URL
- `api_key`: API密钥
- `qps`: QPS限制
- `enabled`: 是否启用
- `description`: 描述信息
- `created_at`: 创建时间
- `updated_at`: 更新时间
- `deleted_at`: 删除时间（软删除）

## 使用示例

### 配置优先级模式

1. 设置系统为优先级模式：
```bash
curl -X PUT http://localhost:8080/api/v1/system/config \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit_mode": "priority",
    "default_priority": 5,
    "default_qps": 10
  }'
```

2. 为用户设置优先级：
```bash
curl -X POST http://localhost:8080/api/v1/user-rate-limits \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "vip_user",
    "priority": 9,
    "enabled": true
  }'
```

### 配置QPS模式

1. 设置系统为QPS模式：
```bash
curl -X PUT http://localhost:8080/api/v1/system/config \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit_mode": "qps",
    "default_priority": 5,
    "default_qps": 10
  }'
```

2. 为用户设置QPS限制：
```bash
curl -X POST http://localhost:8080/api/v1/user-rate-limits \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "power_user",
    "qps": 50,
    "enabled": true
  }'
```

### 配置Agent

```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI GPT-4",
    "type": "openai",
    "url": "https://api.openai.com",
    "api_key": "sk-your-api-key",
    "qps": 10,
    "enabled": true,
    "description": "OpenAI GPT-4 模型用于高质量对话"
  }'
``` 