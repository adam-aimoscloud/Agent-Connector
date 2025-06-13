import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Button,
  Typography,
  Row,
  Col,
  message,
  Tabs,
  Select,
  InputNumber,
  Alert,
  Statistic,
  Progress,
  Space,
  Tag,
  Table,
} from 'antd';
import {
  SettingOutlined,
  ThunderboltOutlined,
  DatabaseOutlined,
  SecurityScanOutlined,
  SaveOutlined,
  ReloadOutlined,
  ClearOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import { useAuth, PermissionGuard } from '../contexts/AuthContext';
import { systemApi, controlFlowApi_endpoints, SystemConfig } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

const SystemSettings: React.FC = () => {
  const { hasPermission } = useAuth();
  const [loading, setLoading] = useState(false);
  const [systemStats, setSystemStats] = useState<any>({});
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);
  const [form] = Form.useForm();

  // Load system statistics
  const loadSystemStats = async () => {
    try {
      const response = await systemApi.getStats();
      if (response.data.code === 200) {
        setSystemStats(response.data.data || {});
      } else {
        throw new Error(response.data.message || 'Failed to get system statistics');
      }
    } catch (error: any) {
      console.error('Failed to load system stats:', error);
      message.error(error.response?.data?.message || 'Failed to load system statistics');
      // Set empty data state
      setSystemStats({
        total_users: 0,
        active_users: 0,
        total_agents: 0,
        active_agents: 0,
        total_requests_today: 0,
        successful_requests_today: 0,
        failed_requests_today: 0,
        cpu_usage: 0,
        memory_usage: 0,
        disk_usage: 0,
        uptime: '0',
        last_backup: '',
      });
    }
  };

  // Load system configuration
  const loadSystemConfig = async () => {
    try {
      const response = await controlFlowApi_endpoints.getSystemConfig();
      if (response.data.code === 200) {
        setSystemConfig(response.data.data || null);
        if (response.data.data) {
          form.setFieldsValue(response.data.data);
        }
      } else {
        throw new Error(response.data.message || 'Failed to get system configuration');
      }
    } catch (error: any) {
      console.error('Failed to load system config:', error);
      message.error(error.response?.data?.message || 'Failed to load system configuration');
      setSystemConfig(null);
    }
  };

  useEffect(() => {
    loadSystemStats();
    loadSystemConfig();
    loadServiceStatus();
  }, []);

  // Save system configuration
  const handleSaveConfig = async (values: any) => {
    setLoading(true);
    try {
      await controlFlowApi_endpoints.updateSystemConfig(values);
      message.success('System configuration saved successfully');
      await loadSystemConfig();
    } catch (error) {
      console.error('Save system config failed:', error);
      message.error('System configuration saved failed');
    } finally {
      setLoading(false);
    }
  };

  // Clean up expired sessions
  const handleCleanupSessions = async () => {
    setLoading(true);
    try {
      await systemApi.cleanupSessions();
      message.success('Expired sessions cleaned up');
      await loadSystemStats();
    } catch (error) {
      console.error('Cleanup sessions failed:', error);
      message.error('Failed to clean up expired sessions');
    } finally {
      setLoading(false);
    }
  };

  // Get success rate
  const getSuccessRate = () => {
    const total = systemStats.total_requests_today || 0;
    const successful = systemStats.successful_requests_today || 0;
    return total > 0 ? ((successful / total) * 100).toFixed(1) : '0.0';
  };

  // Get status color
  const getStatusColor = (value: number, thresholds: [number, number]) => {
    if (value < thresholds[0]) return '#52c41a';
    if (value < thresholds[1]) return '#faad14';
    return '#ff4d4f';
  };

  // System service status
  const [serviceStatus, setServiceStatus] = useState<any[]>([]);

  const loadServiceStatus = async () => {
    try {
      const response = await systemApi.getServiceStatus();
      if (response.data.code === 200) {
        setServiceStatus(response.data.data || []);
      } else {
        throw new Error(response.data.message || 'Failed to get service status');
      }
    } catch (error: any) {
      console.error('Failed to load service status:', error);
      message.error(error.response?.data?.message || 'Failed to load service status');
      setServiceStatus([]);
    }
  };

  const serviceColumns = [
    {
      title: 'Service name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Port',
      dataIndex: 'port',
      key: 'port',
      render: (port: number) => <Text code>{port}</Text>,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'running' ? 'green' : 'red'}>
          {status === 'running' ? 'Running' : 'Stopped'}
        </Tag>
      ),
    },
    {
      title: 'Uptime',
      dataIndex: 'uptime',
      key: 'uptime',
    },
  ];

  return (
    <div>
      <Title level={2}>System settings</Title>
      
      <Tabs defaultActiveKey="overview">
        <TabPane tab="System overview" key="overview" icon={<BarChartOutlined />}>
          {/* System statistics card */}
          <Row gutter={16} style={{ marginBottom: '24px' }}>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="Total users"
                  value={systemStats.total_users || 0}
                  prefix={<DatabaseOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">Active users: {systemStats.active_users || 0}</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="Agent number"
                  value={systemStats.active_agents || 0}
                  suffix={`/ ${systemStats.total_agents || 0}`}
                  prefix={<ThunderboltOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">Active Agent</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="Today requests"
                  value={systemStats.total_requests_today || 0}
                  prefix={<SecurityScanOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">Success rate: {getSuccessRate()}%</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="System uptime"
                  value={systemStats.uptime || '0'}
                  prefix={<SettingOutlined />}
                />
              </Card>
            </Col>
          </Row>

          {/* System resource usage */}
          <Row gutter={16} style={{ marginBottom: '24px' }}>
            <Col xs={24} md={8}>
              <Card title="CPU usage">
                <Progress
                  type="circle"
                  percent={systemStats.cpu_usage || 0}
                  strokeColor={getStatusColor(systemStats.cpu_usage || 0, [70, 85])}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card title="Memory usage">
                <Progress
                  type="circle"
                  percent={systemStats.memory_usage || 0}
                  strokeColor={getStatusColor(systemStats.memory_usage || 0, [80, 90])}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card title="Disk usage">
                <Progress
                  type="circle"
                  percent={systemStats.disk_usage || 0}
                  strokeColor={getStatusColor(systemStats.disk_usage || 0, [85, 95])}
                />
              </Card>
            </Col>
          </Row>

          {/* System service status */}
          <Card title="System service status">
            <Table
              columns={serviceColumns}
              dataSource={serviceStatus}
              rowKey="name"
              pagination={false}
              size="small"
            />
          </Card>
        </TabPane>

        <TabPane tab="System configuration" key="config" icon={<SettingOutlined />}>
          <PermissionGuard permission="system_management">
            <Card title="Rate limit system configuration" extra={<SettingOutlined />}>
              <Alert
                message="Configuration instructions"
                description="Modifying the system rate limit configuration will affect the processing of all API requests, please proceed with caution."
                type="warning"
                showIcon
                style={{ marginBottom: '24px' }}
              />

              <Form
                form={form}
                layout="vertical"
                onFinish={handleSaveConfig}
                autoComplete="off"
              >
                <Row gutter={16}>
                  <Col span={12}>
                    <Form.Item
                      name="rate_limit_mode"
                      label="Rate limit mode"
                      rules={[{ required: true, message: 'Please select rate limit mode' }]}
                      tooltip="Different rate limit modes are suitable for different business scenarios"
                    >
                      <Select placeholder="Please select rate limit mode">
                        <Option value="priority">Priority mode</Option>
                        <Option value="fair">Fair mode</Option>
                        <Option value="weighted">Weighted mode</Option>
                      </Select>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      name="default_priority"
                      label="Default priority"
                      rules={[
                        { required: true, message: 'Please enter default priority' },
                        { type: 'number', min: 1, max: 10, message: 'Priority must be between 1 and 10' },
                      ]}
                    >
                      <InputNumber
                        min={1}
                        max={10}
                        style={{ width: '100%' }}
                        placeholder="Please enter default priority (1-10)"
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Form.Item
                  name="default_qps"
                  label="Default QPS limit"
                  rules={[
                    { required: true, message: 'Please enter default QPS limit' },
                    { type: 'number', min: 1, message: 'QPS must be greater than 0' },
                  ]}
                  tooltip="Query per second limit, it is recommended to set it according to the system performance"
                >
                  <InputNumber
                    min={1}
                    style={{ width: '100%' }}
                    placeholder="Please enter default QPS limit"
                    formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  />
                </Form.Item>

                <Form.Item>
                  <Space>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={loading}
                      icon={<SaveOutlined />}
                    >
                      Save configuration
                    </Button>
                    <Button icon={<ReloadOutlined />} onClick={loadSystemConfig}>
                      Reload
                    </Button>
                  </Space>
                </Form.Item>
              </Form>
            </Card>
          </PermissionGuard>
        </TabPane>

        <TabPane tab="System maintenance" key="maintenance" icon={<DatabaseOutlined />}>
          <PermissionGuard permission="system_management">
            <Row gutter={16}>
              <Col xs={24} md={12}>
                <Card title="Database maintenance" extra={<DatabaseOutlined />}>
                  <Paragraph>
                    <Text type="secondary">
                      Periodically clean up expired data and sessions to maintain system performance.
                    </Text>
                  </Paragraph>
                  
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Button
                      type="primary"
                      icon={<ClearOutlined />}
                      onClick={handleCleanupSessions}
                      loading={loading}
                      block
                    >
                      Clean up expired sessions
                    </Button>
                    
                    <Alert
                      message="Last backup time"
                      description={systemStats.last_backup ? 
                        dayjs(systemStats.last_backup).format('YYYY-MM-DD HH:mm:ss') : 
                        'No backup record'
                      }
                      type="info"
                      showIcon
                    />
                  </Space>
                </Card>
              </Col>

              <Col xs={24} md={12}>
                <Card title="System monitoring" extra={<SecurityScanOutlined />}>
                  <Paragraph>
                    <Text type="secondary">
                      Monitor system key indicators to ensure stable service operation.
                    </Text>
                  </Paragraph>

                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Text strong>Request success rate: </Text>
                      <Tag color={parseFloat(getSuccessRate()) >= 95 ? 'green' : 'orange'}>
                        {getSuccessRate()}%
                      </Tag>
                    </div>
                    
                    <div>
                      <Text strong>Failed request number: </Text>
                      <Tag color={systemStats.failed_requests_today > 100 ? 'red' : 'green'}>
                        {systemStats.failed_requests_today || 0}
                      </Tag>
                    </div>

                    <Button
                      icon={<ReloadOutlined />}
                      onClick={loadSystemStats}
                      block
                    >
                      Refresh statistics data
                    </Button>
                  </Space>
                </Card>
              </Col>
            </Row>
          </PermissionGuard>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default SystemSettings; 