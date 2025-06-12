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

  // 加载统计数据
  const loadStats = async () => {
    setLoading(true);
    try {
      // 从API获取统计数据
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
        throw new Error(statsResponse.data.message || '获取统计数据失败');
      }

      // 获取系统健康状态
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
      // 设置空数据状态
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

  // 系统状态颜色
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

  // 系统状态文本
  const getHealthText = (status: string) => {
    switch (status) {
      case 'healthy':
        return '正常';
      case 'warning':
        return '警告';
      case 'error':
        return '错误';
      default:
        return '未知';
    }
  };

  // 最近活动数据
  const [recentActivities, setRecentActivities] = useState<any[]>([]);

  // 加载最近活动日志
  const loadRecentActivities = async () => {
    try {
      const response = await authApi.getLoginLogs(1, 5);
      if (response.data.code === 200) {
        const activities = response.data.data.map((log: any, index: number) => ({
          key: log.id || index,
          time: new Date(log.created_at).toLocaleString('zh-CN'),
          user: log.username || '未知用户',
          action: log.status === 'success' ? '登录系统' : '登录失败',
          target: `来自 ${log.ip_address}`,
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
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      width: 180,
    },
    {
      title: '用户',
      dataIndex: 'user',
      key: 'user',
      width: 120,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
    },
    {
      title: '对象',
      dataIndex: 'target',
      key: 'target',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>
          {status === 'success' ? '成功' : '失败'}
        </Tag>
      ),
    },
  ];

  return (
    <div>
      {/* 欢迎信息 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col span={24}>
          <Card>
            <Row justify="space-between" align="middle">
              <Col>
                <Space direction="vertical" size="small">
                  <Title level={3} style={{ margin: 0 }}>
                    欢迎回来，{state.user?.full_name || state.user?.username}！
                  </Title>
                  <Text type="secondary">
                    今天是 {new Date().toLocaleDateString('zh-CN', {
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
                  刷新数据
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总用户数"
              value={stats?.totalUsers || 0}
              prefix={<UserOutlined />}
              suffix="人"
              valueStyle={{ color: '#1890ff' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">活跃用户: {stats?.activeUsers || 0}</Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Agent配置"
              value={stats?.totalAgents || 0}
              prefix={<RobotOutlined />}
              suffix="个"
              valueStyle={{ color: '#52c41a' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">活跃配置: {stats?.activeAgents || 0}</Text>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日请求"
              value={stats?.todayRequests || 0}
              prefix={<ThunderboltOutlined />}
              suffix="次"
              valueStyle={{ color: '#fa8c16' }}
            />
            <div style={{ marginTop: '8px' }}>
              <Text type="secondary">错误率: {stats?.errorRate || 0}%</Text>
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
                <Text strong>系统健康度</Text>
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 系统状态和最近活动 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={8}>
          <Card title="系统状态" style={{ height: '400px' }}>
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
                        {service === 'auth_service' && '认证服务'}
                        {service === 'control_service' && '控制服务'}
                        {service === 'data_service' && '数据服务'}
                        {service === 'database' && '数据库'}
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
              message="系统运行正常"
              description="所有核心服务运行良好，数据服务有轻微警告。"
              type="info"
              showIcon
              style={{ marginTop: '16px' }}
            />
          </Card>
        </Col>
        
        <Col xs={24} lg={16}>
          <Card title="最近活动" style={{ height: '400px' }}>
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