// 生产环境配置
export const productionConfig = {
  // 后端服务地址 - 生产环境
  services: {
    auth: {
      baseURL: 'https://api.yourcompany.com:8083',
      timeout: 30000, // 生产环境增加超时时间
    },
    controlFlow: {
      baseURL: 'https://control-api.yourcompany.com:8081',
      timeout: 30000,
    },
    dataFlow: {
      baseURL: 'https://data-api.yourcompany.com:8082',
      timeout: 30000,
    },
  },
  
  // 生产环境特殊配置
  debug: {
    enableRequestLogging: false,
    enableResponseLogging: false,
    enableErrorLogging: true, // 只记录错误日志
  },
  
  // 生产环境API配置
  api: {
    retryAttempts: 3,
    retryDelay: 2000, // 增加重试间隔
  },
  
  // 安全配置
  security: {
    enableHttps: true,
    strictSSL: true,
    enableCSRF: true,
  },
};

export default productionConfig; 