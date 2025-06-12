import React, { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  Button,
  Typography,
  Space,
  Divider,
  Alert,
  Row,
  Col,
} from 'antd';
import {
  UserOutlined,
  LockOutlined,
  LoginOutlined,
} from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { LoginRequest } from '../services/api';

const { Title, Text } = Typography;

const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { state, login, clearError } = useAuth();
  const [loginForm] = Form.useForm();

  // 如果已经登录，重定向到主页
  useEffect(() => {
    if (state.isAuthenticated) {
      const from = (location.state as any)?.from?.pathname || '/';
      navigate(from, { replace: true });
    }
  }, [state.isAuthenticated, navigate, location]);

  // 清除错误 - 只在组件卸载时执行
  useEffect(() => {
    return () => {
      clearError();
    };
  }, []); // 移除clearError依赖，避免无限循环

  // 处理登录
  const handleLogin = async (values: LoginRequest) => {
    try {
      await login(values);
      // 登录成功后会自动重定向（通过上面的useEffect）
    } catch (error) {
      // 错误已经在AuthContext中处理
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
      }}
    >
      <Row justify="center" align="middle" style={{ width: '100%' }}>
        <Col xs={24} sm={20} md={16} lg={12} xl={8}>
          <Card
            style={{
              borderRadius: '12px',
              boxShadow: '0 8px 32px rgba(0, 0, 0, 0.1)',
              border: 'none',
            }}
          >
            {/* 头部标题 */}
            <div style={{ textAlign: 'center', marginBottom: '32px' }}>
              <Title level={2} style={{ color: '#1890ff', marginBottom: '8px' }}>
                Agent-Connector
              </Title>
              <Text type="secondary" style={{ fontSize: '16px' }}>
                统一代理访问平台
              </Text>
            </div>

            {/* 错误提示 */}
            {state.error && (
              <Alert
                message={state.error}
                type="error"
                showIcon
                closable
                onClose={clearError}
                style={{ marginBottom: '24px' }}
              />
            )}

            {/* 登录表单 */}
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Title level={3} style={{ marginBottom: '8px' }}>
                <LoginOutlined style={{ marginRight: '8px' }} />
                用户登录
              </Title>
              <Text type="secondary">请使用管理员分配的账户登录</Text>
            </div>

            <Form
              form={loginForm}
              name="login"
              onFinish={handleLogin}
              autoComplete="off"
              size="large"
              layout="vertical"
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名或邮箱' },
                  { min: 3, message: '用户名至少3个字符' },
                ]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="用户名或邮箱"
                  autoComplete="username"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少6个字符' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码"
                  autoComplete="current-password"
                />
              </Form.Item>

              <Form.Item style={{ marginBottom: '16px' }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={state.loading}
                  block
                  size="large"
                  style={{ borderRadius: '8px' }}
                >
                  登录
                </Button>
              </Form.Item>
            </Form>

            <Divider plain>
              <Text type="secondary">默认管理员账户</Text>
            </Divider>

            <div style={{ textAlign: 'center' }}>
              <Space direction="vertical" size="small">
                <Text type="secondary">
                  用户名: <Text code>admin</Text>
                </Text>
                <Text type="secondary">
                  密码: <Text code>admin123</Text>
                </Text>
                <Divider type="vertical" />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  如需创建新用户，请联系管理员
                </Text>
              </Space>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Login; 