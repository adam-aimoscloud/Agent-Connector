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
  Switch,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  RobotOutlined,
  EyeOutlined,
  ReloadOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useAuth, PermissionGuard } from '../contexts/AuthContext';
import { controlFlowApi_endpoints, Agent, CreateAgentRequest } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Search, TextArea } = Input;
const { Option } = Select;

const Agents: React.FC = () => {
  const { hasPermission } = useAuth();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isViewDrawerVisible, setIsViewDrawerVisible] = useState(false);
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [form] = Form.useForm();

  // Agent类型选项
  const typeOptions = [
    { value: 'openai', label: 'OpenAI', color: 'green', icon: '🤖' },
    { value: 'dify', label: 'Dify', color: 'blue', icon: '🔧' },
    { value: 'custom', label: 'Custom', color: 'orange', icon: '⚙️' },
  ];

  // 响应格式选项
  const responseFormatOptions = [
    { value: 'openai', label: 'OpenAI Compatible' },
    { value: 'dify', label: 'Dify Compatible' },
  ];

  // 状态选项
  const statusOptions = [
    { value: 'active', label: '活跃', color: 'green' },
    { value: 'inactive', label: '停用', color: 'red' },
  ];

  // 常用模型选项
  const modelOptions = {
    openai: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo', 'gpt-3.5-turbo-16k'],
    dify: ['dify-chat', 'dify-completion', 'dify-workflow'],
    custom: ['custom-model-1', 'custom-model-2'],
  };

  // 加载Agent列表
  const loadAgents = async (page = 1, pageSize = 10, search = '') => {
    setLoading(true);
    try {
      const response = await controlFlowApi_endpoints.getAgents(page, pageSize);
      if (response.data.code === 200) {
        let agents = response.data.data;
        
        // 如果有搜索条件，在前端进行过滤
        if (search) {
          agents = agents.filter(agent => 
            agent.name.toLowerCase().includes(search.toLowerCase()) ||
            agent.type.toLowerCase().includes(search.toLowerCase()) ||
            agent.model.toLowerCase().includes(search.toLowerCase()) ||
            agent.description.toLowerCase().includes(search.toLowerCase())
          );
        }
        
        setAgents(agents);
        setPagination({
          current: page,
          pageSize,
          total: search ? agents.length : response.data.pagination.total,
        });
      } else {
        throw new Error(response.data.message || '获取Agent列表失败');
      }
    } catch (error: any) {
      console.error('Failed to load agents:', error);
      message.error(error.response?.data?.message || '加载Agent列表失败');
      setAgents([]);
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
    loadAgents();
  }, []);

  // 搜索处理
  const handleSearch = (value: string) => {
    setSearchText(value);
    loadAgents(1, pagination.pageSize, value);
  };

  // 分页处理
  const handleTableChange = (newPagination: any) => {
    loadAgents(newPagination.current, newPagination.pageSize, searchText);
  };

  // 打开创建/编辑模态框
  const handleOpenModal = (agent?: Agent) => {
    setEditingAgent(agent || null);
    if (agent) {
      form.setFieldsValue({
        name: agent.name,
        type: agent.type,
        endpoint: agent.endpoint,
        source_api_key: agent.source_api_key,
        model: agent.model,
        description: agent.description,
        support_streaming: agent.support_streaming,
        response_format: agent.response_format,
        status: agent.status,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        support_streaming: true,
        response_format: 'openai',
        status: 'active',
      });
    }
    setIsModalVisible(true);
  };

  // 保存Agent
  const handleSaveAgent = async (values: CreateAgentRequest) => {
    try {
      if (editingAgent) {
        // 编辑Agent
        await controlFlowApi_endpoints.updateAgent(editingAgent.id, values);
        message.success('Agent更新成功');
      } else {
        // 创建Agent
        await controlFlowApi_endpoints.createAgent(values);
        message.success('Agent创建成功');
      }
      setIsModalVisible(false);
      loadAgents(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Save agent failed:', error);
      message.error(editingAgent ? 'Agent更新失败' : 'Agent创建失败');
    }
  };

  // 删除Agent
  const handleDeleteAgent = async (agentId: number) => {
    try {
      await controlFlowApi_endpoints.deleteAgent(agentId);
      message.success('Agent删除成功');
      loadAgents(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Delete agent failed:', error);
      message.error('Agent删除失败');
    }
  };

  // 查看Agent详情
  const handleViewAgent = (agent: Agent) => {
    setSelectedAgent(agent);
    setIsViewDrawerVisible(true);
  };

  // 获取类型信息
  const getTypeInfo = (type: string) => {
    return typeOptions.find(opt => opt.value === type) || { label: type, color: 'default', icon: '?' };
  };

  // 隐藏API Key
  const hideApiKey = (key: string) => {
    if (!key) return '';
    return key.substring(0, 8) + '***hidden***';
  };

  // Agent类型变化时更新模型选项
  const handleTypeChange = (type: string) => {
    const models = modelOptions[type as keyof typeof modelOptions] || [];
    form.setFieldsValue({ model: models[0] || '' });
  };

  // 表格列定义
  const columns = [
    {
      title: 'Agent',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Agent) => {
        const typeInfo = getTypeInfo(record.type);
        return (
          <Space>
            <div style={{ fontSize: '18px' }}>{typeInfo.icon}</div>
            <div>
              <div style={{ fontWeight: 'bold' }}>{text}</div>
              <div style={{ color: '#666', fontSize: '12px' }}>
                ID: {record.agent_id}
              </div>
            </div>
          </Space>
        );
      },
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const typeInfo = getTypeInfo(type);
        return <Tag color={typeInfo.color}>{typeInfo.label}</Tag>;
      },
    },
    {
      title: '模型',
      dataIndex: 'model',
      key: 'model',
      render: (model: string) => <Text code>{model}</Text>,
    },
    {
      title: '端点',
      dataIndex: 'endpoint',
      key: 'endpoint',
      render: (endpoint: string) => (
        <Text ellipsis style={{ maxWidth: 200 }} title={endpoint}>
          {endpoint}
        </Text>
      ),
    },
    {
      title: '流式',
      dataIndex: 'support_streaming',
      key: 'support_streaming',
      render: (streaming: boolean) => 
        streaming ? 
          <Tag color="green" icon={<CheckCircleOutlined />}>支持</Tag> : 
          <Tag color="default">不支持</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const option = statusOptions.find(opt => opt.value === status);
        return (
          <Tag color={option?.color} icon={status === 'active' ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}>
            {option?.label || status}
          </Tag>
        );
      },
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      render: (text: any, record: Agent) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewAgent(record)}
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
              title="确定要删除这个Agent吗？"
              description="删除后将无法恢复，请谨慎操作。"
              onConfirm={() => handleDeleteAgent(record.id)}
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
          <Title level={2} style={{ margin: 0 }}>Agent配置</Title>
          <Text type="secondary">管理第三方AI服务接入配置</Text>
        </Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => loadAgents()}>
              刷新
            </Button>
            <PermissionGuard permission="view">
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleOpenModal()}
              >
                新增Agent
              </Button>
            </PermissionGuard>
          </Space>
        </Col>
      </Row>

      {/* 搜索和过滤 */}
      <Card style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder="搜索Agent名称、类型或模型"
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
        </Row>
      </Card>

      {/* Agent表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={agents}
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
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* 创建/编辑Agent模态框 */}
      <Modal
        title={editingAgent ? '编辑Agent' : '新增Agent'}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveAgent}
          autoComplete="off"
        >
          <Alert
            message="配置说明"
            description="Agent配置后将自动生成连接器API密钥，用于统一访问管理。请确保源API密钥的有效性。"
            type="info"
            showIcon
            style={{ marginBottom: '24px' }}
          />

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="Agent名称"
                rules={[
                  { required: true, message: '请输入Agent名称' },
                  { max: 100, message: 'Agent名称最多100个字符' },
                ]}
              >
                <Input placeholder="请输入Agent名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="Agent类型"
                rules={[{ required: true, message: '请选择Agent类型' }]}
              >
                <Select placeholder="请选择Agent类型" onChange={handleTypeChange}>
                  {typeOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      <Space>
                        <span>{option.icon}</span>
                        <span>{option.label}</span>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="endpoint"
                label="API端点"
                rules={[
                  { required: true, message: '请输入API端点' },
                  { type: 'url', message: '请输入有效的URL' },
                ]}
              >
                <Input placeholder="https://api.example.com/v1" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="model"
                label="模型名称"
                rules={[{ required: true, message: '请输入模型名称' }]}
              >
                <Select placeholder="请选择或输入模型名称" mode="tags" maxTagCount={1}>
                  {form.getFieldValue('type') && 
                    modelOptions[form.getFieldValue('type') as keyof typeof modelOptions]?.map(model => (
                      <Option key={model} value={model}>{model}</Option>
                    ))
                  }
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="source_api_key"
            label="源API密钥"
            rules={[{ required: true, message: '请输入源API密钥' }]}
          >
            <Input.Password placeholder="请输入第三方服务的API密钥" />
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 500, message: '描述最多500个字符' }]}
          >
            <TextArea
              rows={3}
              placeholder="请输入Agent描述信息"
              showCount
              maxLength={500}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="support_streaming"
                label="流式响应"
                valuePropName="checked"
              >
                <Switch checkedChildren="支持" unCheckedChildren="不支持" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="response_format"
                label="响应格式"
                rules={[{ required: true, message: '请选择响应格式' }]}
              >
                <Select placeholder="请选择响应格式">
                  {responseFormatOptions.map(option => (
                    <Option key={option.value} value={option.value}>
                      {option.label}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
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
            </Col>
          </Row>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setIsModalVisible(false)}>
                取消
              </Button>
              <Button type="primary" htmlType="submit">
                {editingAgent ? '更新' : '创建'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Agent详情抽屉 */}
      <Drawer
        title="Agent详情"
        placement="right"
        onClose={() => setIsViewDrawerVisible(false)}
        open={isViewDrawerVisible}
        width={600}
      >
        {selectedAgent && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <div style={{ fontSize: '48px', marginBottom: '16px' }}>
                {getTypeInfo(selectedAgent.type).icon}
              </div>
              <Title level={4} style={{ marginBottom: '8px' }}>
                {selectedAgent.name}
              </Title>
              <Tag color={getTypeInfo(selectedAgent.type).color}>
                {getTypeInfo(selectedAgent.type).label}
              </Tag>
            </div>

            <Divider />

            <Descriptions column={1} bordered>
              <Descriptions.Item label="Agent ID">
                <Text code>{selectedAgent.agent_id}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="模型">
                <Text code>{selectedAgent.model}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="API端点">
                <Text copyable>{selectedAgent.endpoint}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="源API密钥">
                <Text code>{hideApiKey(selectedAgent.source_api_key)}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="连接器密钥">
                <Text code copyable>{selectedAgent.connector_api_key}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="流式响应">
                {selectedAgent.support_streaming ? 
                  <Tag color="green" icon={<CheckCircleOutlined />}>支持</Tag> : 
                  <Tag color="default">不支持</Tag>
                }
              </Descriptions.Item>
              <Descriptions.Item label="响应格式">
                <Tag>{selectedAgent.response_format}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={selectedAgent.status === 'active' ? 'green' : 'red'}>
                  {statusOptions.find(opt => opt.value === selectedAgent.status)?.label || selectedAgent.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="描述">
                {selectedAgent.description || '暂无描述'}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {dayjs(selectedAgent.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {dayjs(selectedAgent.updated_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>

            <Divider />

            <Alert
              message="使用说明"
              description={
                <div>
                  <p>使用连接器API密钥访问此Agent：</p>
                  <Text code>Authorization: Bearer {selectedAgent.connector_api_key}</Text>
                  <p style={{ marginTop: '8px' }}>
                    数据流API端点：<Text code>http://localhost:8082/api/v1/chat</Text>
                  </p>
                </div>
              }
              type="info"
              showIcon
            />
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default Agents; 