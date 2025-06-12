import React, { createContext, useContext, useReducer, useEffect, useCallback, ReactNode } from 'react';
import { message } from 'antd';
import { authApi, User, LoginRequest } from '../services/api';

// 认证状态类型
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  error: string | null;
}

// 认证动作类型
type AuthAction =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: User }
  | { type: 'AUTH_FAILURE'; payload: string }
  | { type: 'AUTH_LOGOUT' }
  | { type: 'UPDATE_USER'; payload: User }
  | { type: 'CLEAR_ERROR' };

// 初始状态
const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  loading: true, // 初始时设置为loading状态
  error: null,
};

// 认证reducer
const authReducer = (state: AuthState, action: AuthAction): AuthState => {
  switch (action.type) {
    case 'AUTH_START':
      return {
        ...state,
        loading: true,
        error: null,
      };
    case 'AUTH_SUCCESS':
      return {
        ...state,
        isAuthenticated: true,
        user: action.payload,
        loading: false,
        error: null,
      };
    case 'AUTH_FAILURE':
      return {
        ...state,
        isAuthenticated: false,
        user: null,
        loading: false,
        error: action.payload,
      };
    case 'AUTH_LOGOUT':
      return {
        ...state,
        isAuthenticated: false,
        user: null,
        loading: false,
        error: null,
      };
    case 'UPDATE_USER':
      return {
        ...state,
        user: action.payload,
      };
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };
    default:
      return state;
  }
};

// 认证上下文类型
interface AuthContextType {
  state: AuthState;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  updateProfile: (data: any) => Promise<void>;
  checkAuth: () => Promise<void>;
  hasPermission: (permission: string) => boolean;
  hasRole: (role: string | string[]) => boolean;
  clearError: () => void;
}

// 创建认证上下文
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 认证提供者组件
interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // 手动检查认证状态（如果需要）
  const checkAuth = async () => {
    // 简化版本，只做基本检查
    console.log('Auth check called manually');
  };

  // 用户登录
  const login = useCallback(async (credentials: LoginRequest) => {
    try {
      dispatch({ type: 'AUTH_START' });
      
      const response = await authApi.login(credentials);
      if (response.data.code === 200 && response.data.data) {
        const { token, user } = response.data.data;
        
        localStorage.setItem('auth_token', token);
        localStorage.setItem('user_info', JSON.stringify(user));
        
        dispatch({ type: 'AUTH_SUCCESS', payload: user });
        message.success('登录成功！');
      } else {
        throw new Error(response.data.message || '登录失败');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '用户名或密码错误';
      dispatch({ type: 'AUTH_FAILURE', payload: errorMessage });
      message.error(errorMessage);
      throw error;
    }
  }, []);

  // 用户登出
  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user_info');
      dispatch({ type: 'AUTH_LOGOUT' });
      message.success('已退出登录');
    }
  }, []);

  // 更新用户资料
  const updateProfile = useCallback(async (data: any) => {
    try {
      const response = await authApi.updateProfile(data);
      if (response.data.code === 200 && response.data.data) {
        const updatedUser = response.data.data;
        dispatch({ type: 'UPDATE_USER', payload: updatedUser });
        localStorage.setItem('user_info', JSON.stringify(updatedUser));
        message.success('资料更新成功！');
      } else {
        throw new Error(response.data.message || 'Update failed');
      }
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || '更新失败';
      message.error(errorMessage);
      throw error;
    }
  }, []);

  // 检查权限
  const hasPermission = useCallback((permission: string): boolean => {
    if (!state.user) return false;

    const { role } = state.user;
    
    switch (role) {
      case 'admin':
        return true; // 管理员拥有所有权限
      case 'operator':
        // 操作员权限：可以管理配置但不能管理用户
        return permission !== 'user_management';
      case 'user':
        // 普通用户权限：只能查看自己的信息
        return permission === 'view_own_profile';
      case 'readonly':
        // 只读用户：只能查看
        return permission.startsWith('view_');
      default:
        return false;
    }
  }, [state.user]);

  // 检查角色
  const hasRole = useCallback((role: string | string[]): boolean => {
    if (!state.user) return false;
    
    const userRole = state.user.role;
    if (Array.isArray(role)) {
      return role.includes(userRole);
    }
    return userRole === role;
  }, [state.user]);

  // 清除错误
  const clearError = useCallback(() => {
    dispatch({ type: 'CLEAR_ERROR' });
  }, []);

  // 初始化时检查本地存储的认证信息
  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    const userInfo = localStorage.getItem('user_info');
    
    if (token && userInfo) {
      try {
        const user = JSON.parse(userInfo);
        dispatch({ type: 'AUTH_SUCCESS', payload: user });
      } catch (error) {
        console.error('Failed to parse user info:', error);
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_info');
        dispatch({ type: 'AUTH_LOGOUT' });
      }
    } else {
      dispatch({ type: 'AUTH_LOGOUT' });
    }
  }, []);

  // 提供认证上下文值
  const contextValue: AuthContextType = {
    state,
    login,
    logout,
    updateProfile,
    checkAuth,
    hasPermission,
    hasRole,
    clearError,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

// 使用认证上下文的Hook
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// 权限守卫组件
interface PermissionGuardProps {
  permission?: string;
  role?: string | string[];
  children: ReactNode;
  fallback?: ReactNode;
}

export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  role,
  children,
  fallback = null,
}) => {
  const { hasPermission, hasRole } = useAuth();

  // 检查权限
  if (permission && !hasPermission(permission)) {
    return <>{fallback}</>;
  }

  // 检查角色
  if (role && !hasRole(role)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

export default AuthContext; 