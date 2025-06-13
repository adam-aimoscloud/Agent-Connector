import axios, { AxiosResponse, AxiosInstance } from 'axios';
import { apiConfig, getServiceConfig, getAuthConfig, getGlobalConfig, printCurrentConfig } from '../config/api.config';

// Print current configuration (development environment)
printCurrentConfig();

/**
 * Common function to create axios instance
 */
const createAxiosInstance = (serviceName: keyof typeof apiConfig.services): AxiosInstance => {
  const config = getServiceConfig(serviceName);
  const authConfig = getAuthConfig();
  const globalConfig = getGlobalConfig();
  
  const instance = axios.create({
    baseURL: config.baseURL,
    timeout: config.timeout,
    headers: config.headers,
  });

  // Request interceptor - add authentication token and logging
  instance.interceptors.request.use(
    (requestConfig) => {
      const token = localStorage.getItem(authConfig.tokenKey);
      if (token) {
        requestConfig.headers.Authorization = `Bearer ${token}`;
      }
      
      // Request logging
      if (globalConfig.enableRequestLogging) {
        console.log(`🚀 [${serviceName.toUpperCase()}] Request:`, {
          method: requestConfig.method?.toUpperCase(),
          url: requestConfig.url,
          baseURL: requestConfig.baseURL,
          headers: requestConfig.headers,
          data: requestConfig.data,
        });
      }
      
      return requestConfig;
    },
    (error) => {
      console.error(`❌ [${serviceName.toUpperCase()}] Request Error:`, error);
      return Promise.reject(error);
    }
  );

  // Response interceptor - handle authentication errors and logging
  instance.interceptors.response.use(
    (response) => {
      // Response logging
      if (globalConfig.enableResponseLogging) {
        console.log(`✅ [${serviceName.toUpperCase()}] Response:`, {
          status: response.status,
          statusText: response.statusText,
          data: response.data,
        });
      }
      return response;
    },
    (error) => {
      console.error(`❌ [${serviceName.toUpperCase()}] Response Error:`, {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
      });
      
      if (error.response?.status === 401) {
        localStorage.removeItem(authConfig.tokenKey);
        localStorage.removeItem(authConfig.userInfoKey);
        window.location.href = '/login';
      }
      return Promise.reject(error);
    }
  );

  return instance;
};

// Create axios instances for each service
const api = createAxiosInstance('auth');
const controlFlowApi = createAxiosInstance('controlFlow');

// Type definitions
export interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  avatar: string;
  role: 'admin' | 'operator' | 'user' | 'readonly';
  status: 'active' | 'inactive' | 'blocked' | 'pending';
  last_login?: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: User;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  full_name: string;
  role: string;
  status: string;
}

export interface UpdateUserRequest {
  username?: string;
  email?: string;
  full_name?: string;
  role?: string;
  status?: string;
  avatar?: string;
}

export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

export interface Agent {
  id: number;
  name: string;
  type: 'openai' | 'openai_compatible' | 'dify';
  url: string;
  source_api_key: string;
  connector_api_key: string;
  agent_id: string;
  qps: number;
  enabled: boolean;
  description: string;
  support_streaming: boolean;
  response_format: 'openai' | 'dify';
  created_at: string;
  updated_at: string;
}

export interface CreateAgentRequest {
  name: string;
  type: 'openai' | 'openai_compatible' | 'dify';
  url: string;
  source_api_key: string;
  qps: number;
  enabled: boolean;
  description: string;
  support_streaming: boolean;
  response_format: 'openai' | 'dify';
}

export interface SystemConfig {
  id: number;
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
  error?: {
    type: string;
    code: string;
    message: string;
    details?: string;
  };
}

export interface PaginationResponse<T> {
  code: number;
  message: string;
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

// 认证API
export const authApi = {
  // 用户登录
  login: (data: LoginRequest): Promise<AxiosResponse<ApiResponse<LoginResponse>>> =>
    api.post('/api/v1/auth/login', data),

  // 用户登出
  logout: (): Promise<AxiosResponse<ApiResponse<null>>> =>
    api.post('/api/v1/auth/logout'),

  // 获取个人资料
  getProfile: (): Promise<AxiosResponse<ApiResponse<any>>> =>
    api.get('/api/v1/auth/profile'),

  // 更新个人资料
  updateProfile: (data: Partial<UpdateUserRequest>): Promise<AxiosResponse<ApiResponse<User>>> =>
    api.put('/api/v1/auth/profile', data),

  // 修改密码
  changePassword: (data: ChangePasswordRequest): Promise<AxiosResponse<ApiResponse<null>>> =>
    api.post('/api/v1/auth/change-password', data),

  // 获取登录日志
  getLoginLogs: (page = 1, pageSize = 10): Promise<AxiosResponse<PaginationResponse<any>>> =>
    api.get(`/api/v1/auth/login-logs?page=${page}&page_size=${pageSize}`),

  // 健康检查
  healthCheck: (): Promise<AxiosResponse<ApiResponse<any>>> =>
    api.get('/api/v1/auth/health'),
};

// 用户管理API（管理员功能）
export const userApi = {
  // 获取用户列表
  getUsers: (page = 1, pageSize = 10, search = ''): Promise<AxiosResponse<PaginationResponse<User>>> =>
    api.get(`/api/v1/users?page=${page}&page_size=${pageSize}&search=${search}`),

  // 创建用户
  createUser: (data: CreateUserRequest): Promise<AxiosResponse<ApiResponse<User>>> =>
    api.post('/api/v1/users', data),

  // 获取用户详情
  getUser: (id: number): Promise<AxiosResponse<ApiResponse<User>>> =>
    api.get(`/api/v1/users/${id}`),

  // 更新用户信息
  updateUser: (id: number, data: UpdateUserRequest): Promise<AxiosResponse<ApiResponse<User>>> =>
    api.put(`/api/v1/users/${id}`, data),

  // 删除用户
  deleteUser: (id: number): Promise<AxiosResponse<ApiResponse<null>>> =>
    api.delete(`/api/v1/users/${id}`),

  // 更新用户状态
  updateUserStatus: (id: number, status: string): Promise<AxiosResponse<ApiResponse<null>>> =>
    api.put(`/api/v1/users/${id}/status`, { status }),
};

// 系统管理API
export const systemApi = {
  // 获取系统统计
  getStats: (): Promise<AxiosResponse<ApiResponse<any>>> =>
    api.get('/api/v1/system/stats'),

  // 获取服务状态
  getServiceStatus: (): Promise<AxiosResponse<ApiResponse<any[]>>> =>
    api.get('/api/v1/system/services'),

  // 清理过期会话
  cleanupSessions: (): Promise<AxiosResponse<ApiResponse<null>>> =>
    api.post('/api/v1/system/cleanup-sessions'),
};

// 控制流API（Agent配置）
export const controlFlowApi_endpoints = {
  // 系统配置
  getSystemConfig: (): Promise<AxiosResponse<ApiResponse<SystemConfig>>> =>
    controlFlowApi.get('/api/v1/controlflow/system-config'),

  updateSystemConfig: (data: Partial<SystemConfig>): Promise<AxiosResponse<ApiResponse<SystemConfig>>> =>
    controlFlowApi.put('/api/v1/controlflow/system-config', data),

  // Agent管理
  getAgents: (page = 1, pageSize = 10): Promise<AxiosResponse<PaginationResponse<Agent>>> =>
    controlFlowApi.get(`/api/v1/agents?page=${page}&page_size=${pageSize}`),

  createAgent: (data: CreateAgentRequest): Promise<AxiosResponse<ApiResponse<Agent>>> =>
    controlFlowApi.post('/api/v1/agents', data),

  getAgent: (id: number): Promise<AxiosResponse<ApiResponse<Agent>>> =>
    controlFlowApi.get(`/api/v1/agents/${id}`),

  updateAgent: (id: number, data: Partial<CreateAgentRequest>): Promise<AxiosResponse<ApiResponse<Agent>>> =>
    controlFlowApi.put(`/api/v1/agents/${id}`, data),

  deleteAgent: (id: number): Promise<AxiosResponse<ApiResponse<null>>> =>
    controlFlowApi.delete(`/api/v1/agents/${id}`),

  // 健康检查
  healthCheck: (): Promise<AxiosResponse<ApiResponse<any>>> =>
    controlFlowApi.get('/api/v1/controlflow/health'),
};

export default api; 