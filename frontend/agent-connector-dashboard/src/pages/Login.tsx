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

  // If already logged in, redirect to home page
  useEffect(() => {
    if (state.isAuthenticated) {
      const from = (location.state as any)?.from?.pathname || '/';
      navigate(from, { replace: true });
    }
  }, [state.isAuthenticated, navigate, location]);

  // Clear error - only executed when component is unmounted
  useEffect(() => {
    return () => {
      clearError();
    };
  }, []); // Remove clearError dependency to avoid infinite loop

  // Handle login
  const handleLogin = async (values: LoginRequest) => {
    try {
      await login(values);
      // After login, it will automatically redirect (through the useEffect above)
    } catch (error) {
      // Error is already handled in AuthContext
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
            {/* Header title */}
            <div style={{ textAlign: 'center', marginBottom: '32px' }}>
              <Title level={2} style={{ color: '#1890ff', marginBottom: '8px' }}>
                Agent-Connector
              </Title>
              <Text type="secondary" style={{ fontSize: '16px' }}>
                Unified agent access platform
              </Text>
            </div>

            {/* Error message */}
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

            {/* Login form */}
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Title level={3} style={{ marginBottom: '8px' }}>
                <LoginOutlined style={{ marginRight: '8px' }} />
                User login
              </Title>
              <Text type="secondary">Please use the account assigned by the administrator to login</Text>
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
                  { required: true, message: 'Please enter username or email' },
                  { min: 3, message: 'Username must be at least 3 characters' },
                ]}
              >
                <Input
                  prefix={<UserOutlined />}
                  placeholder="Username or email"
                  autoComplete="username"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: 'Please enter password' },
                  { min: 6, message: 'Password must be at least 6 characters' },
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="Password"
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
                  Login
                </Button>
              </Form.Item>
            </Form>

            <Divider plain>
              <Text type="secondary">Default admin account</Text>
            </Divider>

            <div style={{ textAlign: 'center' }}>
              <Space direction="vertical" size="small">
                <Text type="secondary">
                  Username: <Text code>admin</Text>
                </Text>
                <Text type="secondary">
                  Password: <Text code>admin123</Text>
                </Text>
                <Divider type="vertical" />
                <Text type="secondary" style={{ fontSize: '12px' }}>
                  If you need to create a new user, please contact the administrator
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