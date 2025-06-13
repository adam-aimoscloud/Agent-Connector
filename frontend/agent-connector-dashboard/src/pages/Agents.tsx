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
  InputNumber,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  EyeOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { PermissionGuard } from '../contexts/AuthContext';
import { controlFlowApi_endpoints, Agent, CreateAgentRequest } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Search, TextArea } = Input;
const { Option } = Select;

const Agents: React.FC = () => {
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

  // Agent type options
  const typeOptions = [
    { value: 'openai', label: 'OpenAI', color: 'green', icon: '🤖' },
    { value: 'dify-chat', label: 'Dify Chat', color: 'blue', icon: '💬' },
    { value: 'dify-workflow', label: 'Dify Workflow', color: 'purple', icon: '⚙️' },
  ];

  // Response format options
  const responseFormatOptions = [
    { value: 'openai', label: 'OpenAI Compatible' },
    { value: 'dify', label: 'Dify Compatible' },
  ];



  // Load Agent list
  const loadAgents = async (page = 1, pageSize = 10, search = '') => {
    setLoading(true);
    try {
      const response = await controlFlowApi_endpoints.getAgents(page, pageSize);
      if (response.data.code === 200) {
        let agents = response.data.data;
        
        // If there is a search condition, filter it in the front end
        if (search) {
          agents = agents.filter(agent => 
            agent.name.toLowerCase().includes(search.toLowerCase()) ||
            agent.type.toLowerCase().includes(search.toLowerCase()) ||
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
        throw new Error(response.data.message || 'Failed to get Agent list');
      }
    } catch (error: any) {
      console.error('Failed to load agents:', error);
      message.error(error.response?.data?.message || 'Failed to load Agent list');
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

  // Initialize loading
  useEffect(() => {
    loadAgents();
  }, []);

  // Search processing
  const handleSearch = (value: string) => {
    setSearchText(value);
    loadAgents(1, pagination.pageSize, value);
  };

  // Pagination processing
  const handleTableChange = (newPagination: any) => {
    loadAgents(newPagination.current, newPagination.pageSize, searchText);
  };

  // Open create/edit modal
  const handleOpenModal = (agent?: Agent) => {
    setEditingAgent(agent || null);
    if (agent) {
      form.setFieldsValue({
        name: agent.name,
        type: agent.type,
        url: agent.url,
        source_api_key: agent.source_api_key,
        qps: agent.qps,
        enabled: agent.enabled,
        description: agent.description,
        support_streaming: agent.support_streaming,
        response_format: agent.response_format,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        qps: 10,
        enabled: true,
        support_streaming: true,
        response_format: 'openai',
      });
    }
    setIsModalVisible(true);
  };

  // Save Agent
  const handleSaveAgent = async (values: CreateAgentRequest) => {
    try {
      if (editingAgent) {
        // Edit Agent
        await controlFlowApi_endpoints.updateAgent(editingAgent.id, values);
        message.success('Agent updated successfully');
      } else {
        // Create Agent
        await controlFlowApi_endpoints.createAgent(values);
        message.success('Agent created successfully');
      }
      setIsModalVisible(false);
      loadAgents(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Save agent failed:', error);
      message.error(editingAgent ? 'Agent update failed' : 'Agent create failed');
    }
  };

  // Delete Agent
  const handleDeleteAgent = async (agentId: number) => {
    try {
      await controlFlowApi_endpoints.deleteAgent(agentId);
      message.success('Agent deleted successfully');
      loadAgents(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Delete agent failed:', error);
      message.error('Agent delete failed');
    }
  };

  // View Agent details
  const handleViewAgent = (agent: Agent) => {
    setSelectedAgent(agent);
    setIsViewDrawerVisible(true);
  };

  // Get type information
  const getTypeInfo = (type: string) => {
    return typeOptions.find(opt => opt.value === type) || { label: type, color: 'default', icon: '?' };
  };

  // Hide API Key
  const hideApiKey = (key: string) => {
    if (!key) return '';
    return key.substring(0, 8) + '***hidden***';
  };



  // Table column definition
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
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const typeInfo = getTypeInfo(type);
        return <Tag color={typeInfo.color}>{typeInfo.label}</Tag>;
      },
    },
    {
      title: 'API Endpoint',
      dataIndex: 'url',
      key: 'url',
      render: (url: string) => (
        <Text ellipsis style={{ maxWidth: 200 }} title={url}>
          {url}
        </Text>
      ),
    },
    {
      title: 'QPS Limit',
      dataIndex: 'qps',
      key: 'qps',
      render: (qps: number) => <Text code>{qps}</Text>,
    },
    {
      title: 'Streaming',
      dataIndex: 'support_streaming',
      key: 'support_streaming',
      render: (streaming: boolean) => 
        streaming ? 
          <Tag color="green" icon={<CheckCircleOutlined />}>Supported</Tag> : 
          <Tag color="default">Not supported</Tag>,
    },
    {
      title: 'Status',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'green' : 'red'} icon={enabled ? <CheckCircleOutlined /> : <ExclamationCircleOutlined />}>
          {enabled ? 'Enabled' : 'Disabled'}
        </Tag>
      ),
    },
    {
      title: 'Updated At',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: 'Actions',
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
            View
          </Button>
          <PermissionGuard permission="view">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleOpenModal(record)}
              size="small"
            >
              Edit
            </Button>
            <Popconfirm
              title="Are you sure you want to delete this Agent?"
              description="Once deleted, it cannot be recovered. Please proceed with caution."
              onConfirm={() => handleDeleteAgent(record.id)}
              okText="Yes"
              cancelText="No"
            >
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                size="small"
              >
                Delete
              </Button>
            </Popconfirm>
          </PermissionGuard>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* Page title */}
      <Row justify="space-between" align="middle" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2} style={{ margin: 0 }}>Agent Configuration</Title>
          <Text type="secondary">Manage third-party AI service access configuration</Text>
        </Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => loadAgents()}>
              Refresh
            </Button>
            <PermissionGuard permission="view">
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleOpenModal()}
              >
                Add Agent
              </Button>
            </PermissionGuard>
          </Space>
        </Col>
      </Row>

      {/* Search and filter */}
      <Card style={{ marginBottom: '16px' }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder="Search Agent name, type or description"
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
        </Row>
      </Card>

      {/* Agent table */}
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
              `${range[0]}-${range[1]} of ${total}`, // TODO: translate
          }}
          onChange={handleTableChange}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Create/edit Agent modal */}
      <Modal
        title={editingAgent ? 'Edit Agent' : 'Add Agent'}
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
            message="Configuration instructions"
            description="After configuring the Agent, the connector API key will be automatically generated for unified access management. Please ensure the validity of the source API key."
            type="info"
            showIcon
            style={{ marginBottom: '24px' }}
          />

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="Agent name"
                rules={[
                  { required: true, message: 'Please enter Agent name' },
                  { max: 100, message: 'Agent name can only be up to 100 characters' },
                ]}
              >
                <Input placeholder="Please enter Agent name" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="type"
                label="Agent type"
                rules={[{ required: true, message: 'Please select Agent type' }]}
              >
                <Select placeholder="Please select Agent type">
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
            <Col span={16}>
              <Form.Item
                name="url"
                label="API Endpoint"
                rules={[
                  { required: true, message: 'Please enter API Endpoint' },
                  { type: 'url', message: 'Please enter a valid URL' },
                ]}
              >
                <Input placeholder="https://api.example.com/v1" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="qps"
                label="QPS Limit"
                rules={[
                  { required: true, message: 'Please enter QPS Limit' },
                  { type: 'number', min: 1, message: 'QPS must be greater than 0' },
                ]}
              >
                <InputNumber
                  placeholder="10"
                  min={1}
                  max={1000}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="source_api_key"
            label="Source API Key"
            rules={[{ required: true, message: 'Please enter Source API Key' }]}
          >
            <Input.Password placeholder="Please enter Source API Key" />
          </Form.Item>

          <Form.Item
            name="description"
            label="Description"
            rules={[{ max: 500, message: 'Description can only be up to 500 characters' }]}
          >
            <TextArea
              rows={3}
              placeholder="Please enter Agent description"
              showCount
              maxLength={500}
            />
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="support_streaming"
                label="Streaming"
                valuePropName="checked"
              >
                <Switch checkedChildren="Supported" unCheckedChildren="Not supported" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="response_format"
                label="Response Format"
                rules={[{ required: true, message: 'Please select Response Format' }]}
              >
                <Select placeholder="Please select Response Format">
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
                name="enabled"
                label="Enabled"
                valuePropName="checked"
              >
                <Switch checkedChildren="Enabled" unCheckedChildren="Disabled" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setIsModalVisible(false)}>
                Cancel
              </Button>
              <Button type="primary" htmlType="submit">
                {editingAgent ? 'Update' : 'Create'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Agent details drawer */}
      <Drawer
        title="Agent details"
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
              <Descriptions.Item label="API Endpoint">
                <Text copyable>{selectedAgent.url}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="QPS Limit">
                <Text code>{selectedAgent.qps}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Source API Key">
                <Text code>{hideApiKey(selectedAgent.source_api_key)}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Connector API Key">
                <Text code copyable>{selectedAgent.connector_api_key}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Streaming">
                {selectedAgent.support_streaming ? 
                  <Tag color="green" icon={<CheckCircleOutlined />}>Supported</Tag> : 
                  <Tag color="default">Not supported</Tag>
                }
              </Descriptions.Item>
              <Descriptions.Item label="Response Format">
                <Tag>{selectedAgent.response_format}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={selectedAgent.enabled ? 'green' : 'red'}>
                  {selectedAgent.enabled ? 'Enabled' : 'Disabled'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Description">
                {selectedAgent.description || 'No description'}
              </Descriptions.Item>
              <Descriptions.Item label="Created At">
                {dayjs(selectedAgent.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="Updated At">
                {dayjs(selectedAgent.updated_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>

            <Divider />

            <Alert
              message="Usage instructions"
              description={
                <div>
                  <p>Use the connector API key to access this Agent:</p>
                  <Text code>Authorization: Bearer {selectedAgent.connector_api_key}</Text>
                  <p style={{ marginTop: '8px' }}>
                    Data flow API endpoint: <Text code>http://localhost:8082/api/v1/dataflow/chat/{selectedAgent.agent_id}</Text>
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