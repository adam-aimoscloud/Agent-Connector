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

  // 加载系统统计信息
  const loadSystemStats = async () => {
    try {
      const response = await systemApi.getStats();
      if (response.data.code === 200) {
        setSystemStats(response.data.data || {});
      } else {
        throw new Error(response.data.message || '获取系统统计失败');
      }
    } catch (error: any) {
      console.error('Failed to load system stats:', error);
      message.error(error.response?.data?.message || '加载系统统计失败');
      // 设置空数据状态
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

  // 加载系统配置
  const loadSystemConfig = async () => {
    try {
      const response = await controlFlowApi_endpoints.getSystemConfig();
      if (response.data.code === 200) {
        setSystemConfig(response.data.data || null);
        if (response.data.data) {
          form.setFieldsValue(response.data.data);
        }
      } else {
        throw new Error(response.data.message || '获取系统配置失败');
      }
    } catch (error: any) {
      console.error('Failed to load system config:', error);
      message.error(error.response?.data?.message || '加载系统配置失败');
      setSystemConfig(null);
    }
  };

  useEffect(() => {
    loadSystemStats();
    loadSystemConfig();
    loadServiceStatus();
  }, []);

  // 保存系统配置
  const handleSaveConfig = async (values: any) => {
    setLoading(true);
    try {
      await controlFlowApi_endpoints.updateSystemConfig(values);
      message.success('系统配置保存成功');
      await loadSystemConfig();
    } catch (error) {
      console.error('Save system config failed:', error);
      message.error('系统配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  // 清理过期会话
  const handleCleanupSessions = async () => {
    setLoading(true);
    try {
      await systemApi.cleanupSessions();
      message.success('过期会话清理完成');
      await loadSystemStats();
    } catch (error) {
      console.error('Cleanup sessions failed:', error);
      message.error('清理过期会话失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取成功率
  const getSuccessRate = () => {
    const total = systemStats.total_requests_today || 0;
    const successful = systemStats.successful_requests_today || 0;
    return total > 0 ? ((successful / total) * 100).toFixed(1) : '0.0';
  };

  // 获取状态颜色
  const getStatusColor = (value: number, thresholds: [number, number]) => {
    if (value < thresholds[0]) return '#52c41a';
    if (value < thresholds[1]) return '#faad14';
    return '#ff4d4f';
  };

  // 系统服务状态
  const [serviceStatus, setServiceStatus] = useState<any[]>([]);

  const loadServiceStatus = async () => {
    try {
      const response = await systemApi.getServiceStatus();
      if (response.data.code === 200) {
        setServiceStatus(response.data.data || []);
      } else {
        throw new Error(response.data.message || '获取服务状态失败');
      }
    } catch (error: any) {
      console.error('Failed to load service status:', error);
      message.error(error.response?.data?.message || '加载服务状态失败');
      setServiceStatus([]);
    }
  };

  const serviceColumns = [
    {
      title: '服务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '端口',
      dataIndex: 'port',
      key: 'port',
      render: (port: number) => <Text code>{port}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'running' ? 'green' : 'red'}>
          {status === 'running' ? '运行中' : '已停止'}
        </Tag>
      ),
    },
    {
      title: '运行时间',
      dataIndex: 'uptime',
      key: 'uptime',
    },
  ];

  return (
    <div>
      <Title level={2}>系统设置</Title>
      
      <Tabs defaultActiveKey="overview">
        <TabPane tab="系统概览" key="overview" icon={<BarChartOutlined />}>
          {/* 系统统计卡片 */}
          <Row gutter={16} style={{ marginBottom: '24px' }}>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="总用户数"
                  value={systemStats.total_users || 0}
                  prefix={<DatabaseOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">活跃用户: {systemStats.active_users || 0}</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="Agent数量"
                  value={systemStats.active_agents || 0}
                  suffix={`/ ${systemStats.total_agents || 0}`}
                  prefix={<ThunderboltOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">活跃Agent</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="今日请求"
                  value={systemStats.total_requests_today || 0}
                  prefix={<SecurityScanOutlined />}
                />
                <div style={{ marginTop: '8px' }}>
                  <Text type="secondary">成功率: {getSuccessRate()}%</Text>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Card>
                <Statistic
                  title="系统运行时间"
                  value={systemStats.uptime || '0'}
                  prefix={<SettingOutlined />}
                />
              </Card>
            </Col>
          </Row>

          {/* 系统资源使用情况 */}
          <Row gutter={16} style={{ marginBottom: '24px' }}>
            <Col xs={24} md={8}>
              <Card title="CPU使用率">
                <Progress
                  type="circle"
                  percent={systemStats.cpu_usage || 0}
                  strokeColor={getStatusColor(systemStats.cpu_usage || 0, [70, 85])}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card title="内存使用率">
                <Progress
                  type="circle"
                  percent={systemStats.memory_usage || 0}
                  strokeColor={getStatusColor(systemStats.memory_usage || 0, [80, 90])}
                />
              </Card>
            </Col>
            <Col xs={24} md={8}>
              <Card title="磁盘使用率">
                <Progress
                  type="circle"
                  percent={systemStats.disk_usage || 0}
                  strokeColor={getStatusColor(systemStats.disk_usage || 0, [85, 95])}
                />
              </Card>
            </Col>
          </Row>

          {/* 系统服务状态 */}
          <Card title="系统服务状态">
            <Table
              columns={serviceColumns}
              dataSource={serviceStatus}
              rowKey="name"
              pagination={false}
              size="small"
            />
          </Card>
        </TabPane>

        <TabPane tab="系统配置" key="config" icon={<SettingOutlined />}>
          <PermissionGuard permission="system_management">
            <Card title="限流系统配置" extra={<SettingOutlined />}>
              <Alert
                message="配置说明"
                description="修改系统限流配置将影响所有API请求的处理方式，请谨慎操作。"
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
                      label="限流模式"
                      rules={[{ required: true, message: '请选择限流模式' }]}
                      tooltip="不同限流模式适用于不同的业务场景"
                    >
                      <Select placeholder="请选择限流模式">
                        <Option value="priority">优先级模式</Option>
                        <Option value="fair">公平模式</Option>
                        <Option value="weighted">加权模式</Option>
                      </Select>
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item
                      name="default_priority"
                      label="默认优先级"
                      rules={[
                        { required: true, message: '请输入默认优先级' },
                        { type: 'number', min: 1, max: 10, message: '优先级必须在1-10之间' },
                      ]}
                    >
                      <InputNumber
                        min={1}
                        max={10}
                        style={{ width: '100%' }}
                        placeholder="请输入默认优先级(1-10)"
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Form.Item
                  name="default_qps"
                  label="默认QPS限制"
                  rules={[
                    { required: true, message: '请输入默认QPS限制' },
                    { type: 'number', min: 1, message: 'QPS必须大于0' },
                  ]}
                  tooltip="每秒查询数限制，建议根据系统性能设置"
                >
                  <InputNumber
                    min={1}
                    style={{ width: '100%' }}
                    placeholder="请输入默认QPS限制"
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
                      保存配置
                    </Button>
                    <Button icon={<ReloadOutlined />} onClick={loadSystemConfig}>
                      重新加载
                    </Button>
                  </Space>
                </Form.Item>
              </Form>
            </Card>
          </PermissionGuard>
        </TabPane>

        <TabPane tab="系统维护" key="maintenance" icon={<DatabaseOutlined />}>
          <PermissionGuard permission="system_management">
            <Row gutter={16}>
              <Col xs={24} md={12}>
                <Card title="数据库维护" extra={<DatabaseOutlined />}>
                  <Paragraph>
                    <Text type="secondary">
                      定期清理过期数据和会话，保持系统性能。
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
                      清理过期会话
                    </Button>
                    
                    <Alert
                      message="上次备份时间"
                      description={systemStats.last_backup ? 
                        dayjs(systemStats.last_backup).format('YYYY-MM-DD HH:mm:ss') : 
                        '暂无备份记录'
                      }
                      type="info"
                      showIcon
                    />
                  </Space>
                </Card>
              </Col>

              <Col xs={24} md={12}>
                <Card title="系统监控" extra={<SecurityScanOutlined />}>
                  <Paragraph>
                    <Text type="secondary">
                      监控系统关键指标，确保服务稳定运行。
                    </Text>
                  </Paragraph>

                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Text strong>请求成功率: </Text>
                      <Tag color={parseFloat(getSuccessRate()) >= 95 ? 'green' : 'orange'}>
                        {getSuccessRate()}%
                      </Tag>
                    </div>
                    
                    <div>
                      <Text strong>失败请求数: </Text>
                      <Tag color={systemStats.failed_requests_today > 100 ? 'red' : 'green'}>
                        {systemStats.failed_requests_today || 0}
                      </Tag>
                    </div>

                    <Button
                      icon={<ReloadOutlined />}
                      onClick={loadSystemStats}
                      block
                    >
                      刷新统计数据
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