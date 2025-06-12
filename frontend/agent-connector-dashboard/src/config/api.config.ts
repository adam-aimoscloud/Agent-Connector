// API配置文件 - 集中管理所有后端服务配置

/**
 * 后端服务配置接口
 */
export interface BackendServiceConfig {
  baseURL: string;
  timeout: number;
  headers: Record<string, string>;
  retryAttempts: number;
  retryDelay: number;
}

/**
 * 环境配置接口
 */
export interface ApiConfig {
  // 服务配置
  services: {
    auth: BackendServiceConfig;
    controlFlow: BackendServiceConfig;
    dataFlow: BackendServiceConfig;
  };
  
  // 全局配置
  global: {
    // API请求超时时间（毫秒）
    defaultTimeout: number;
    // 默认重试次数
    defaultRetryAttempts: number;
    // 重试间隔（毫秒）
    defaultRetryDelay: number;
    // 是否启用请求日志
    enableRequestLogging: boolean;
    // 是否启用响应日志
    enableResponseLogging: boolean;
  };
  
  // 认证配置
  auth: {
    // Token存储键名
    tokenKey: string;
    // 用户信息存储键名
    userInfoKey: string;
    // Token过期时间（小时）
    tokenExpirationHours: number;
    // 是否自动刷新Token
    autoRefreshToken: boolean;
  };
  
  // 分页配置
  pagination: {
    // 默认每页数量
    defaultPageSize: number;
    // 最大每页数量
    maxPageSize: number;
    // 分页大小选项
    pageSizeOptions: number[];
  };
}

/**
 * 获取环境变量，支持默认值
 */
const getEnvVar = (key: string, defaultValue: string): string => {
  return process.env[key] || defaultValue;
};

/**
 * 获取环境变量数字值
 */
const getEnvNumber = (key: string, defaultValue: number): number => {
  const value = process.env[key];
  const parsed = value ? parseInt(value, 10) : defaultValue;
  return isNaN(parsed) ? defaultValue : parsed;
};

/**
 * 获取环境变量布尔值
 */
const getEnvBoolean = (key: string, defaultValue: boolean): boolean => {
  const value = process.env[key];
  if (!value) return defaultValue;
  return value.toLowerCase() === 'true';
};

/**
 * 默认请求头
 */
const defaultHeaders = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
};

/**
 * API配置
 */
export const apiConfig: ApiConfig = {
  services: {
    // 认证服务配置
    auth: {
      baseURL: getEnvVar('REACT_APP_AUTH_API_URL', 'http://localhost:8083'),
      timeout: getEnvNumber('REACT_APP_AUTH_API_TIMEOUT', 10000),
      headers: defaultHeaders,
      retryAttempts: getEnvNumber('REACT_APP_AUTH_API_RETRY_ATTEMPTS', 3),
      retryDelay: getEnvNumber('REACT_APP_AUTH_API_RETRY_DELAY', 1000),
    },
    
    // 控制流服务配置
    controlFlow: {
      baseURL: getEnvVar('REACT_APP_CONTROL_FLOW_API_URL', 'http://localhost:8081'),
      timeout: getEnvNumber('REACT_APP_CONTROL_FLOW_API_TIMEOUT', 10000),
      headers: defaultHeaders,
      retryAttempts: getEnvNumber('REACT_APP_CONTROL_FLOW_API_RETRY_ATTEMPTS', 3),
      retryDelay: getEnvNumber('REACT_APP_CONTROL_FLOW_API_RETRY_DELAY', 1000),
    },
    
    // 数据流服务配置
    dataFlow: {
      baseURL: getEnvVar('REACT_APP_DATA_FLOW_API_URL', 'http://localhost:8082'),
      timeout: getEnvNumber('REACT_APP_DATA_FLOW_API_TIMEOUT', 10000),
      headers: defaultHeaders,
      retryAttempts: getEnvNumber('REACT_APP_DATA_FLOW_API_RETRY_ATTEMPTS', 3),
      retryDelay: getEnvNumber('REACT_APP_DATA_FLOW_API_RETRY_DELAY', 1000),
    },
  },
  
  // 全局配置
  global: {
    defaultTimeout: getEnvNumber('REACT_APP_API_DEFAULT_TIMEOUT', 10000),
    defaultRetryAttempts: getEnvNumber('REACT_APP_API_DEFAULT_RETRY_ATTEMPTS', 3),
    defaultRetryDelay: getEnvNumber('REACT_APP_API_DEFAULT_RETRY_DELAY', 1000),
    enableRequestLogging: getEnvBoolean('REACT_APP_ENABLE_REQUEST_LOGGING', false),
    enableResponseLogging: getEnvBoolean('REACT_APP_ENABLE_RESPONSE_LOGGING', false),
  },
  
  // 认证配置
  auth: {
    tokenKey: getEnvVar('REACT_APP_AUTH_TOKEN_KEY', 'auth_token'),
    userInfoKey: getEnvVar('REACT_APP_AUTH_USER_INFO_KEY', 'user_info'),
    tokenExpirationHours: getEnvNumber('REACT_APP_AUTH_TOKEN_EXPIRATION_HOURS', 24),
    autoRefreshToken: getEnvBoolean('REACT_APP_AUTH_AUTO_REFRESH_TOKEN', true),
  },
  
  // 分页配置
  pagination: {
    defaultPageSize: getEnvNumber('REACT_APP_PAGINATION_DEFAULT_PAGE_SIZE', 10),
    maxPageSize: getEnvNumber('REACT_APP_PAGINATION_MAX_PAGE_SIZE', 100),
    pageSizeOptions: [10, 20, 50, 100],
  },
};

/**
 * 获取服务配置
 */
export const getServiceConfig = (serviceName: keyof ApiConfig['services']): BackendServiceConfig => {
  return apiConfig.services[serviceName];
};

/**
 * 获取全局配置
 */
export const getGlobalConfig = () => {
  return apiConfig.global;
};

/**
 * 获取认证配置
 */
export const getAuthConfig = () => {
  return apiConfig.auth;
};

/**
 * 获取分页配置
 */
export const getPaginationConfig = () => {
  return apiConfig.pagination;
};

/**
 * 打印当前配置（用于调试）
 */
export const printCurrentConfig = () => {
  if (process.env.NODE_ENV === 'development') {
    console.group('🔧 API Configuration');
    console.log('Auth Service:', apiConfig.services.auth.baseURL);
    console.log('Control Flow Service:', apiConfig.services.controlFlow.baseURL);
    console.log('Data Flow Service:', apiConfig.services.dataFlow.baseURL);
    console.log('Global Timeout:', apiConfig.global.defaultTimeout);
    console.log('Request Logging:', apiConfig.global.enableRequestLogging);
    console.log('Response Logging:', apiConfig.global.enableResponseLogging);
    console.groupEnd();
  }
};

export default apiConfig; 