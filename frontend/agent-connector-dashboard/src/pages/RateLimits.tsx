import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Input,
  Space,
  Modal,
  Form,
  Select,
  Tag,
  Typography,
  Card,
  Row,
  Col,
  Popconfirm,
  message,
  Drawer,
  Descriptions,
  Divider,
  InputNumber,
  Switch,
  Alert,
  Progress,
  Statistic,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  EyeOutlined,
  ReloadOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { useAuth, PermissionGuard } from '../contexts/AuthContext';
import { dataFlowApi_endpoints, RateLimit, CreateRateLimitRequest } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Search, TextArea } = Input;
const { Option } = Select;

const RateLimits: React.FC = () => {
  const { hasPermission } = useAuth();
  const [rateLimits, setRateLimits] = useState<RateLimit[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedRateLimit, setSelectedRateLimit] = useState<RateLimit | null>(null);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isViewDrawerVisible, setIsViewDrawerVisible] = useState(false);
  const [editingRateLimit, setEditingRateLimit] = useState<RateLimit | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [form] = Form.useForm();

  // 限制类型选项
  const limitTypeOptions = [
    { value: 'requests_per_minute', label: '每分钟请求数', unit: 'req/min', icon: <ThunderboltOutlined /> },
    { value: 'tokens_per_minute', label: '每分钟Token数', unit: 'tokens/min', icon: <ClockCircleOutlined /> },
    { value: 'requests_per_hour', label: '每小时请求数', unit: 'req/hour', icon: <ThunderboltOutlined /> },
    { value: 'requests_per_day', label: '每日请求数', unit: 'req/day', icon: <ThunderboltOutlined /> },
    { value: 'tokens_per_day', label: '每日Token数', unit: 'tokens/day', icon: <ClockCircleOutlined /> },
  ];

  // 应用范围选项
  const scopeOptions = [
    { value: 'global', label: '全局', color: 'red' },
    { value: 'user', label: '用户', color: 'blue' },
    { value: 'agent', label: 'Agent', color: 'green' },
    { value: 'ip', label: 'IP地址', color: 'orange' },
  ];

  // 状态选项
  const statusOptions = [
    { value: 'active', label: '生效', color: 'green' },
    { value: 'inactive', label: '停用', color: 'red' },
  ];

  // 加载限流配置列表
  const loadRateLimits = async (page = 1, pageSize = 10, search = '') => {
    setLoading(true);
    try {
      const response = await dataFlowApi_endpoints.getRateLimits(page, pageSize);
      if (response.data.code === 200) {
        let rateLimits = response.data.data;
        
        // 如果有搜索条件，在前端进行过滤
        if (search) {
          rateLimits = rateLimits.filter(rateLimit => 
            rateLimit.name.toLowerCase().includes(search.toLowerCase()) ||
            rateLimit.limit_type.toLowerCase().includes(search.toLowerCase()) ||
            rateLimit.scope.toLowerCase().includes(search.toLowerCase()) ||
            rateLimit.scope_value.toLowerCase().includes(search.toLowerCase()) ||
            rateLimit.description.toLowerCase().includes(search.toLowerCase())
          );
        }

        setRateLimits(rateLimits);
        setPagination({
          current: page,
          pageSize,
          total: search ? rateLimits.length : response.data.pagination.total,
        });
      } else {
        throw new Error(response.data.message || '获取限流配置失败');
      }
    } catch (error: any) {
      console.error('Failed to load rate limits:', error);
      message.error(error.response?.data?.message || '加载限流配置失败');
      setRateLimits([]);
      setPagination({
        current: page,
        pageSize,
        total: 0,
      });
    } finally {
      setLoading(false);
    }
  };

  // 初始化加载
  useEffect(() => {
    loadRateLimits();
  }, []);

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchText(value);
    loadRateLimits(1, pagination.pageSize, value);
  };

  // 分页处理
  const handleTableChange = (newPagination: any) => {
    loadRateLimits(newPagination.current, newPagination.pageSize, searchText);
  };

  // 打开创建/编辑模态框
  const handleOpenModal = (rateLimit?: RateLimit) => {
    setEditingRateLimit(rateLimit || null);
    if (rateLimit) {
      form.setFieldsValue({
        name: rateLimit.name,
        limit_type: rateLimit.limit_type,
        limit_value: rateLimit.limit_value,
        scope: rateLimit.scope,
        scope_value: rateLimit.scope_value,
        description: rateLimit.description,
        status: rateLimit.status,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        limit_type: 'requests_per_minute',
        scope: 'global',
        scope_value: '*',
        status: 'active',
      });
    }
    setIsModalVisible(true);
  };

  // 保存限流配置
  const handleSaveRateLimit = async (values: CreateRateLimitRequest) => {
    try {
      if (editingRateLimit) {
        // 编辑限流配置
        await dataFlowApi_endpoints.updateRateLimit(editingRateLimit.id, values);
        message.success('限流配置更新成功');
      } else {
        // 创建限流配置
        await dataFlowApi_endpoints.createRateLimit(values);
        message.success('限流配置创建成功');
      }
      setIsModalVisible(false);
      loadRateLimits(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Save rate limit failed:', error);
      message.error(editingRateLimit ? '限流配置更新失败' : '限流配置创建失败');
    }
  };

  // 删除限流配置
  const handleDeleteRateLimit = async (rateLimitId: number) => {
    try {
      await dataFlowApi_endpoints.deleteRateLimit(rateLimitId);
      message.success('限流配置删除成功');
      loadRateLimits(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Delete rate limit failed:', error);
      message.error('限流配置删除失败');
    }
  };

  // 查看限流配置详情
  const handleViewRateLimit = (rateLimit: RateLimit) => {
    setSelectedRateLimit(rateLimit);
    setIsViewDrawerVisible(true);
  };

  // 获取限制类型信息
  const getLimitTypeInfo = (limitType: string) => {
    return limitTypeOptions.find(opt => opt.value === limitType) || 
           { label: limitType, unit: '', icon: <ThunderboltOutlined /> };
  };

  // 获取使用率
  const getUsagePercentage = (current: number, limit: number) => {
    return limit > 0 ? Math.round((current / limit) * 100) : 0;
  };

  // 获取使用率状态
  const getUsageStatus = (percentage: number) => {
    if (percentage >= 90) return 'exception';
    if (percentage >= 70) return 'active';
    return 'success';
  };

  // 表格列定义
  const columns = [
    {
      title: '规则名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: RateLimit) => (
        <div>
          <div style={{ fontWeight: 'bold' }}>{text}</div>
          <div style={{ color: '#666', fontSize: '12px' }}>
            ID: {record.id}
          </div>
        </div>
      ),
    },
    {
      title: '限制类型',
      dataIndex: 'limit_type',
      key: 'limit_type',
      render: (limitType: string) => {
        const typeInfo = getLimitTypeInfo(limitType);
        return (
          <Space>
            {typeInfo.icon}
            <span>{typeInfo.label}</span>
          </Space>
        );
      },
    },
    {
      title: '限制值',
      dataIndex: 'limit_value',
      key: 'limit_value',
      render: (value: number, record: RateLimit) => {
        const typeInfo = getLimitTypeInfo(record.limit_type);
        return <Text strong>{value.toLocaleString()} {typeInfo.unit}</Text>;
      },
    },
    {
      title: '应用范围',
      key: 'scope',
      render: (text: any, record: RateLimit) => {
        const scopeInfo = scopeOptions.find(opt => opt.value === record.scope);
        return (
          <div>
            <Tag color={scopeInfo?.color}>{scopeInfo?.label}</Tag>
            <div style={{ fontSize: '12px', color: '#666' }}>
              {record.scope_value}
            </div>
          </div>
        );
      },
    },
    {
      title: '使用情况',
      key: 'usage',
      render: (text: any, record: RateLimit) => {
        const percentage = getUsagePercentage(record.current_usage || 0, record.limit_value);
        const status = getUsageStatus(percentage);
        
        return (
          <div style={{ width: 120 }}>
            <Progress
              percent={percentage}
              size="small"
              status={status}
              showInfo={false}
            />
            <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
              {(record.current_usage || 0).toLocaleString()} / {record.limit_value.toLocaleString()}
            </div>
          </div>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const option = statusOptions.find(opt => opt.value === status);
        return (
          <Tag 
            color={option?.color} 
            icon={status === 'active' ? <CheckCircleOutlined /> : <WarningOutlined />}
          >
            {option?.label || status}
          </Tag>
        );
      },
    },
    {
      title: '重置时间',
      dataIndex: 'reset_time',
      key: 'reset_time',
      render: (date: string) => date ? dayjs(date).format('MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (text: any, record: RateLimit) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewRateLimit(record)}
            size="small"
          >
            查看
          </Button>
          <PermissionGuard permission="view">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleOpenModal(record)}
              size="small"
            >
              编辑
            </Button>
            <Popconfirm
              title="确定要删除这个限流规则吗？"
              description="删除后将无法恢复，请谨慎操作。"
              onConfirm={() => handleDeleteRateLimit(record.id)}
              okText="确定"
              cancelText="取消"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                size="small"
              >
                删除
              </Button>
            </Popconfirm>
          </PermissionGuard>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 页面标题 */}
      <Row justify="space-between" align="middle" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2} style={{ margin: 0 }}>限流配置</Title>
          <Text type="secondary">管理API请求频率和使用量限制</Text>
        </Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => loadRateLimits()}>
              刷新
            </Button>
            <PermissionGuard permission="view">
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleOpenModal()}
              >
                新增限流规则
              </Button>
            </PermissionGuard>
          </Space>
        </Col>
      </Row>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="总规则数"
              value={rateLimits.length}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="生效规则"
              value={rateLimits.filter(r => r.status === 'active').length}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="高使用率规则"
              value={rateLimits.filter(r => 
                r.current_usage && getUsagePercentage(r.current_usage, r.limit_value) >= 80
              ).length}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="全局规则"
              value={rateLimits.filter(r => r.scope === 'global').length}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* 搜索和过滤 */}
      <Card style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder="搜索规则名称、类型或范围"
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
        </Row>
      </Card>

      {/* 限流规则表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={rateLimits}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* 创建/编辑限流规则模态框 */}
      <Modal
        title={editingRateLimit ? '编辑限流规则' : '新增限流规则'}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        footer={null}
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveRateLimit}
          autoComplete="off"
        >
          <Alert
            message="配置说明"
            description="限流规则按优先级生效：IP > 用户 > Agent > 全局。相同类型的限制只有最具体的规则会生效。"
            type="info"
            showIcon
            style={{ marginBottom: '24px' }}
          />

          <Form.Item
            name="name"
            label="规则名称"
            rules={[
              { required: true, message: '请输入规则名称' },
              { max: 100, message: '规则名称最多100个字符' },
            ]}
          >
            <Input placeholder="请输入规则名称" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="limit_type"
                label="限制类型"
                rules={[{ required: true, message: '请选择限制类型' }]}
              >
                <Select placeholder="请选择限制类型">
                  {limitTypeOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <Space>
                        {option.icon}
                        <span>{option.label}</span>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="limit_value"
                label="限制值"
                rules={[
                  { required: true, message: '请输入限制值' },
                  { type: 'number', min: 1, message: '限制值必须大于0' },
                ]}
              >
                <InputNumber
                  placeholder="请输入限制值"
                  style={{ width: '100%' }}
                  min={1}
                  formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="scope"
                label="应用范围"
                rules={[{ required: true, message: '请选择应用范围' }]}
              >
                <Select placeholder="请选择应用范围">
                  {scopeOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <Tag color={option.color}>{option.label}</Tag>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="scope_value"
                label="范围值"
                rules={[{ required: true, message: '请输入范围值' }]}
                tooltip={{
                  title: '全局: *, 用户: 用户名, Agent: Agent ID, IP: IP地址',
                  icon: <WarningOutlined />,
                }}
              >
                <Input placeholder="如: *, user123, gpt4-001, 192.168.1.1" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 500, message: '描述最多500个字符' }]}
          >
            <TextArea
              rows={3}
              placeholder="请输入限流规则描述"
              showCount
              maxLength={500}
            />
          </Form.Item>

          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select placeholder="请选择状态">
              {statusOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Tag color={option.color}>{option.label}</Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setIsModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingRateLimit ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* 限流规则详情抽屉 */}
      <Drawer
        title="限流规则详情"
        placement="right"
        onClose={() => setIsViewDrawerVisible(false)}
        open={isViewDrawerVisible}
        width={600}
      >
        {selectedRateLimit && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <div style={{ fontSize: '48px', marginBottom: '16px' }}>
                {getLimitTypeInfo(selectedRateLimit.limit_type).icon}
              </div>
              <Title level={4} style={{ marginBottom: '8px' }}>
                {selectedRateLimit.name}
              </Title>
              <Tag color={selectedRateLimit.status === 'active' ? 'green' : 'red'}>
                {statusOptions.find(opt => opt.value === selectedRateLimit.status)?.label}
              </Tag>
            </div>

            <Divider />

            {/* 使用情况卡片 */}
            <Card style={{ marginBottom: '16px' }}>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic
                    title="当前使用量"
                    value={selectedRateLimit.current_usage || 0}
                    formatter={(value) => value?.toLocaleString() || '0'}
                  />
                </Col>
                <Col span={12}>
                  <Statistic
                    title="使用率"
                    value={getUsagePercentage(
                      selectedRateLimit.current_usage || 0, 
                      selectedRateLimit.limit_value
                    )}
                    suffix="%"
                    valueStyle={{ 
                      color: getUsagePercentage(
                        selectedRateLimit.current_usage || 0, 
                        selectedRateLimit.limit_value
                      ) >= 80 ? '#cf1322' : '#52c41a' 
                    }}
                  />
                </Col>
              </Row>
              <Progress
                percent={getUsagePercentage(
                  selectedRateLimit.current_usage || 0, 
                  selectedRateLimit.limit_value
                )}
                status={getUsageStatus(getUsagePercentage(
                  selectedRateLimit.current_usage || 0, 
                  selectedRateLimit.limit_value
                ))}
                style={{ marginTop: '16px' }}
              />
            </Card>

            <Descriptions column={1} bordered>
              <Descriptions.Item label="规则ID">
                {selectedRateLimit.id}
              </Descriptions.Item>
              <Descriptions.Item label="限制类型">
                <Space>
                  {getLimitTypeInfo(selectedRateLimit.limit_type).icon}
                  {getLimitTypeInfo(selectedRateLimit.limit_type).label}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="限制值">
                <Text strong>
                  {selectedRateLimit.limit_value.toLocaleString()} {getLimitTypeInfo(selectedRateLimit.limit_type).unit}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="应用范围">
                <div>
                  <Tag color={scopeOptions.find(opt => opt.value === selectedRateLimit.scope)?.color}>
                    {scopeOptions.find(opt => opt.value === selectedRateLimit.scope)?.label}
                  </Tag>
                  <br />
                  <Text code>{selectedRateLimit.scope_value}</Text>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="重置时间">
                {selectedRateLimit.reset_time 
                  ? dayjs(selectedRateLimit.reset_time).format('YYYY-MM-DD HH:mm:ss')
                  : '无'
                }
              </Descriptions.Item>
              <Descriptions.Item label="描述">
                {selectedRateLimit.description || '暂无描述'}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {dayjs(selectedRateLimit.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {dayjs(selectedRateLimit.updated_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>

            {selectedRateLimit.current_usage && selectedRateLimit.current_usage > 0 && (
              <>
                <Divider />
                <Alert
                  message="使用提醒"
                  description={
                    getUsagePercentage(selectedRateLimit.current_usage, selectedRateLimit.limit_value) >= 80
                      ? "当前使用量已接近限制值，请注意监控。"
                      : "当前使用量正常。"
                  }
                  type={getUsagePercentage(selectedRateLimit.current_usage, selectedRateLimit.limit_value) >= 80 ? "warning" : "info"}
                  showIcon
                />
              </>
            )}
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default RateLimits; 