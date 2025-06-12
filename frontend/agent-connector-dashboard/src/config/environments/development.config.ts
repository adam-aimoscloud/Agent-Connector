// Development environment configuration
export const developmentConfig = {
  // Backend service addresses
  services: {
    auth: {
      baseURL: 'http://localhost:8083',
      timeout: 10000,
    },
    controlFlow: {
      baseURL: 'http://localhost:8081',
      timeout: 10000,
    },
    dataFlow: {
      baseURL: 'http://localhost:8082',
      timeout: 10000,
    },
  },
  
  // Development environment specific configuration
  debug: {
    enableRequestLogging: true,
    enableResponseLogging: true,
    enableErrorLogging: true,
  },
  
  // Development environment API configuration
  api: {
    retryAttempts: 1, // Reduce retry attempts in development environment
    retryDelay: 500,
  },
};

export default developmentConfig; 