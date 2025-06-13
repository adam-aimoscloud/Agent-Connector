import React, { useState, useEffect } from 'react';
import {
  Row,
  Col,
  Card,
  Statistic,
  Typography,
  Space,
  Table,
  Tag,
  Progress,
  Alert,
  Button,
  Divider,
} from 'antd';
import {
  UserOutlined,
  RobotOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { systemApi, authApi } from '../services/api';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  const { state } = useAuth();
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<any>(null);
  const [systemHealth, setSystemHealth] = useState<any>(null);

  // Load statistics data
  const loadStats = async () => {
    setLoading(true);
    try {
      // Get statistics data from API
      const [statsResponse, healthResponse] = await Promise.all([
        systemApi.getStats(),
        authApi.healthCheck(),
      ]);
      
      if (statsResponse.data.code === 200) {
        const data = statsResponse.data.data;
        setStats({
          totalUsers: data.total_users || 0,
          activeUsers: data.active_users || 0,
          totalAgents: data.total_agents || 0,
          activeAgents: data.active_agents || 0,
          todayRequests: data.total_requests_today || 0,
          errorRate: ((data.failed_requests_today || 0) / Math.max(data.total_requests_today || 1, 1) * 100).toFixed(1),
        });
      } else {
        throw new Error(statsResponse.data.message || 'Failed to get statistics data');
      }

      // Get system health status
      if (healthResponse.data.code === 200) {
        setSystemHealth(healthResponse.data.data || {
          auth_service: 'unknown',
          control_service: 'unknown',
          data_service: 'unknown',
          database: 'unknown',
          redis: 'unknown',
        });
      }
    } catch (error: any) {
      console.error('Failed to load stats:', error);
      // Set empty data state
      setStats({
        totalUsers: 0,
        activeUsers: 0,
        totalAgents: 0,
        activeAgents: 0,
        todayRequests: 0,
        errorRate: 0,
      });
      setSystemHealth({
        auth_service: 'error',
        control_service: 'error',
        data_service: 'error',
        database: 'error',
        redis: 'error',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
    loadRecentActivities();
  }, []);

  // System status color
  const getHealthColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'green';
      case 'warning':
        return 'orange';
      case 'error':
        return 'red';
      default:
        return 'gray';
    }
  };

  // System status text
  const getHealthText = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'Healthy';
      case 'warning':
        return 'Warning';
      case 'error':
        return 'Error';
      default:
        return 'Unknown';
    }
  };

  // Recent activities data
  const [recentActivities, setRecentActivities] = useState<any[]>([]);

  // Load recent activity logs
  const loadRecentActivities = async () => {
    try {
      const response = await authApi.getLoginLogs(1, 5);
      if (response.data.code === 200) {
        const activities = response.data.data.map((log: any, index: number) => ({
          key: log.id || index,
          time: new Date(log.created_at).toLocaleString('zh-CN'),
          user: log.username || 'Unknown user',
          action: log.status === 'success' ? 'Login system' : 'Login failed',
          target: `From ${log.ip_address}`,
          status: log.status,
        }));
        setRecentActivities(activities);
      }
    } catch (error) {
      console.error('Failed to load recent activities:', error);
      setRecentActivities([]);
    }
  };

  const activityColumns = [
    {
      title: 'Time',
      dataIndex: 'time',
      key: 'time',
      width: 180,
    },
    {
      title: 'User',
      dataIndex: 'user',
      key: 'user',
      width: 120,
    },
    {
      title: 'Action',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: 'Target',
      dataIndex: 'target',
      key: 'target',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>
          {status === 'success' ? 'Success' : 'Failed'}
        </Tag>
      ),
    },
  ];

  return (
    <div>
      {/* Welcome message */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col span={24}>
          <Card>
            <Row justify="space-between" align="middle">
              <Col>
                <Space direction="vertical" size="small">
                  <Title level={3} style={{ margin: 0 }}>
                    Welcome back, {state.user?.full_name || state.user?.username}!
                  </Title>
                  <Text type="secondary">
                    Today is {new Date().toLocaleDateString('zh-CN', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                      weekday: 'long',
                    })}
                  </Text>
                </Space>
              </Col>
              <Col>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={loadStats}
                  loading={loading}
                >
                  Refresh data
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* Statistics cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total users"
              value={stats?.totalUsers || 0}
              prefix={<UserOutlined />}
              suffix="people"
              valueStyle={{ color: '#1890ff' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">Active users: {stats?.activeUsers || 0}</Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Agent configuration"
              value={stats?.totalAgents || 0}
              prefix={<RobotOutlined />}
              suffix="agents"
              valueStyle={{ color: '#52c41a' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">Active agents: {stats?.activeAgents || 0}</Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Today requests"
              value={stats?.todayRequests || 0}
              prefix={<ThunderboltOutlined />}
              suffix="times"
              valueStyle={{ color: '#fa8c16' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">Error rate: {stats?.errorRate || 0}%</Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <div style={{ textAlign: 'center' }}>
              <Progress
                type="dashboard"
                percent={100 - (stats?.errorRate || 0)}
                format={(percent) => `${percent?.toFixed(1)}%`}
                strokeColor={{
                  '0%': '#108ee9',
                  '100%': '#87d068',
                }}
              />
              <div style={{ marginTop: '8px' }}>
                <Text strong>System health</Text>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* System status and recent activities */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title="System status" style={{ height: '400px' }}>
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              {systemHealth && Object.entries(systemHealth).map(([service, status]) => (
                <Row key={service} justify="space-between" align="middle">
                  <Col>
                    <Space>
                      {status === 'healthy' ? (
                        <CheckCircleOutlined style={{ color: 'green' }} />
                      ) : (
                        <ExclamationCircleOutlined style={{ color: 'orange' }} />
                      )}
                      <Text strong>
                        {service === 'auth_service' && 'Authentication service'}
                        {service === 'control_service' && 'Control service'}
                        {service === 'data_service' && 'Data service'}
                        {service === 'database' && 'Database'}
                        {service === 'redis' && 'Redis'}
                      </Text>
                    </Space>
                  </Col>
                  <Col>
                    <Tag color={getHealthColor(status as string)}>
                      {getHealthText(status as string)}
                    </Tag>
                  </Col>
                </Row>
              ))}
            </Space>
            
            <Divider />
            
            <Alert
              message="System is running normally"
              description="All core services are running well, with a slight warning from the data service."
              type="info"
              showIcon
              style={{ marginTop: '16px' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} lg={16}>
          <Card title="Recent activities" style={{ height: '400px' }}>
            <Table
              columns={activityColumns}
              dataSource={recentActivities}
              pagination={false}
              size="small"
              scroll={{ y: 280 }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard; 