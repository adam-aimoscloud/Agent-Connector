import React, { useState } from 'react';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Typography,
  Space,
  Button,
  Badge,
  Tooltip,
  Divider,
} from 'antd';
import {
  DashboardOutlined,
  UserOutlined,
  TeamOutlined,
  SettingOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  RobotOutlined,
  ThunderboltOutlined,
  BellOutlined,
  ProfileOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const DashboardLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { state, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);

  // Menu items configuration
  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
      path: '/',
    },
    {
      key: '/users',
      icon: <TeamOutlined />,
      label: 'User Management',
      path: '/users',
      permission: 'user_management',
    },
    {
      key: '/agents',
      icon: <RobotOutlined />,
      label: 'Agent Configuration',
      path: '/agents',
      permission: 'view',
    },
    {
      key: '/rate-limits',
      icon: <ThunderboltOutlined />,
      label: 'Rate Limit Configuration',
      path: '/rate-limits',
      permission: 'view',
    },
    {
      key: '/profile',
      icon: <UserOutlined />,
      label: 'Profile',
      path: '/profile',
      permission: 'view_own_profile',
    },
    {
      key: '/system',
      icon: <SettingOutlined />,
      label: 'System Settings',
      path: '/system',
      role: ['admin', 'operator'],
    },
  ];

  // Handle menu click
  const handleMenuClick = (key: string) => {
    const menuItem = menuItems.find(item => item.key === key);
    if (menuItem) {
      navigate(menuItem.path);
    }
  };

  // User dropdown menu
  const userMenuItems = [
    {
      key: 'profile',
      icon: <ProfileOutlined />,
      label: 'Profile',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Logout',
      onClick: logout,
    },
  ];

  // Get user role display text
  const getRoleText = (role: string) => {
    const roleMap = {
      admin: 'Administrator',
      operator: 'Operator',
      user: 'User',
      readonly: 'Read-only User',
    };
    return roleMap[role as keyof typeof roleMap] || role;
  };

  // Get user status color
  const getStatusColor = (status: string) => {
    const colorMap = {
      active: 'green',
      inactive: 'orange',
      blocked: 'red',
      pending: 'blue',
    };
    return colorMap[status as keyof typeof colorMap] || 'default';
  };

  // Filter menu items with permissions
  const getFilteredMenuItems = () => {
    return menuItems.filter(item => {
      if (item.permission && !state.user) return false;
      if (item.permission && state.user) {
        const { role } = state.user;
        switch (role) {
          case 'admin':
            return true;
          case 'operator':
            return item.permission !== 'user_management';
          case 'user':
            return item.permission === 'view_own_profile';
          case 'readonly':
            return item.permission === 'view' || item.permission === 'view_own_profile';
          default:
            return false;
        }
      }
      if (item.role && state.user) {
        return item.role.includes(state.user.role);
      }
      return true;
    });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* Sidebar */}
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          background: '#fff',
          boxShadow: '2px 0 8px rgba(0,0,0,0.05)',
        }}
        width={240}
      >
        {/* Logo area */}
        <div
          style={{
            height: '64px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'flex-start',
            padding: collapsed ? '0' : '0 24px',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <RobotOutlined
            style={{
              fontSize: '24px',
              color: '#1890ff',
            }}
          />
          {!collapsed && (
            <Text
              strong
              style={{
                marginLeft: '12px',
                fontSize: '16px',
                color: '#1890ff',
              }}
            >
              Agent-Connector
            </Text>
          )}
        </div>

        {/* Navigation menu */}
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          style={{
            border: 'none',
            height: 'calc(100vh - 64px)',
          }}
        >
          {getFilteredMenuItems().map(item => (
            <Menu.Item
              key={item.key}
              icon={item.icon}
              onClick={() => handleMenuClick(item.key)}
            >
              {item.label}
            </Menu.Item>
          ))}
        </Menu>
      </Sider>

      {/* Main content area */}
      <Layout>
        {/* Header */}
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          {/* Left side - collapse button */}
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{
              fontSize: '16px',
              width: '40px',
              height: '40px',
            }}
          />

          {/* Right side - user info */}
          <Space size="large">
            {/* Notification button */}
            <Tooltip title="Notifications">
              <Badge count={0} size="small">
                <Button
                  type="text"
                  icon={<BellOutlined />}
                  style={{ fontSize: '16px' }}
                />
              </Badge>
            </Tooltip>

            <Divider type="vertical" />

            {/* User dropdown menu */}
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  cursor: 'pointer',
                  padding: '8px',
                  borderRadius: '8px',
                  transition: 'background-color 0.3s',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = '#f5f5f5';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = 'transparent';
                }}
              >
                <Avatar
                  size="default"
                  icon={<UserOutlined />}
                  src={state.user?.avatar}
                  style={{ marginRight: '12px' }}
                />
                <Space direction="vertical" size={0}>
                  <Text strong style={{ fontSize: '14px' }}>
                    {state.user?.full_name || state.user?.username}
                  </Text>
                  <Space size="small">
                    <Badge
                      color={getStatusColor(state.user?.status || 'active')}
                      text={
                        <Text type="secondary" style={{ fontSize: '12px' }}>
                          {getRoleText(state.user?.role || 'user')}
                        </Text>
                      }
                    />
                  </Space>
                </Space>
              </div>
            </Dropdown>
          </Space>
        </Header>

        {/* Content area */}
        <Content
          style={{
            margin: '24px',
            padding: '24px',
            background: '#fff',
            borderRadius: '8px',
            boxShadow: '0 2px 8px rgba(0,0,0,0.05)',
            overflow: 'auto',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default DashboardLayout; 