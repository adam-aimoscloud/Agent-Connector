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
  Avatar,
  Typography,
  Card,
  Row,
  Col,
  Popconfirm,
  message,
  Drawer,
  Descriptions,
  Divider,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  UserOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useAuth, PermissionGuard } from '../contexts/AuthContext';
import { userApi, User, CreateUserRequest, UpdateUserRequest } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const Users: React.FC = () => {
  const { hasPermission } = useAuth();
  const [modal, contextHolder] = Modal.useModal();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isViewDrawerVisible, setIsViewDrawerVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [form] = Form.useForm();

  // Test Modal.error functionality
  const testModalError = () => {
    modal.error({
      title: 'Test Error Modal',
      content: 'This is a test error modal to verify that Modal.error functionality is working properly.',
      okText: 'OK',
      width: 500,
    });
  };

  // Role options
  const roleOptions = [
    { value: 'admin', label: 'Administrator', color: 'red' },
    { value: 'operator', label: 'Operator', color: 'orange' },
    { value: 'user', label: 'User', color: 'blue' },
    { value: 'readonly', label: 'Read-only User', color: 'gray' },
  ];

  // Status options
  const statusOptions = [
    { value: 'active', label: 'Active', color: 'green' },
    { value: 'inactive', label: 'Inactive', color: 'orange' },
    { value: 'blocked', label: 'Blocked', color: 'red' },
    { value: 'pending', label: 'Pending', color: 'blue' },
  ];

  // Load user list
  const loadUsers = async (page = 1, pageSize = 10, search = '') => {
    setLoading(true);
    try {
      const response = await userApi.getUsers(page, pageSize, search);
      if (response.data.code === 200) {
        setUsers(response.data.data);
        setPagination({
          current: page,
          pageSize,
          total: response.data.pagination.total,
        });
      } else {
        throw new Error(response.data.message || 'Failed to load user list');
      }
    } catch (error: any) {
      console.error('Failed to load users:', error);
      
      // Get detailed error information
      let errorDetail = '';
      if (error.response?.data?.error?.message) {
        errorDetail = error.response.data.error.message;
      } else if (error.response?.data?.message) {
        errorDetail = error.response.data.message;
      } else if (error.message) {
        errorDetail = error.message;
      }
      
      // Show error modal
      modal.error({
        title: 'Failed to Load User List',
        content: errorDetail || 'Unable to load user list. Please check your network connection or contact the system administrator.',
        okText: 'OK',
        width: 500,
      });
      
      setUsers([]);
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
    loadUsers();
  }, []);

  // Search handler
  const handleSearch = (value: string) => {
    setSearchText(value);
    loadUsers(1, pagination.pageSize, value);
  };

  // Pagination handler
  const handleTableChange = (newPagination: any) => {
    loadUsers(newPagination.current, newPagination.pageSize, searchText);
  };

  // Open create/edit modal
  const handleOpenModal = (user?: User) => {
    setEditingUser(user || null);
    if (user) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        full_name: user.full_name,
        role: user.role,
        status: user.status,
      });
    } else {
      form.resetFields();
    }
    setIsModalVisible(true);
  };

  // Save user
  const handleSaveUser = async (values: CreateUserRequest) => {
    try {
      if (editingUser) {
        // Edit user
        await userApi.updateUser(editingUser.id, values);
        message.success('User updated successfully');
      } else {
        // Create user
        await userApi.createUser(values);
        message.success('User created successfully');
      }
      setIsModalVisible(false);
      loadUsers(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Save user failed:', error);
      console.log('Error response:', error.response);
      console.log('Error response data:', error.response?.data);
      
      // Get detailed error information
      let errorMessage = editingUser ? 'Failed to Update User' : 'Failed to Create User';
      let errorDetail = '';
      
      if (error.response?.data?.error?.message) {
        errorDetail = error.response.data.error.message;
        console.log('Using error.response.data.error.message:', errorDetail);
      } else if (error.response?.data?.message) {
        errorDetail = error.response.data.message;
        console.log('Using error.response.data.message:', errorDetail);
      } else if (error.message) {
        errorDetail = error.message;
        console.log('Using error.message:', errorDetail);
      }
      
      console.log('Final error detail:', errorDetail);
      console.log('About to show modal.error');
      
      // Show error modal
      modal.error({
        title: errorMessage,
        content: errorDetail || 'Please check your input information or contact the system administrator.',
        okText: 'OK',
        width: 500,
      });
      
      console.log('modal.error called');
    }
  };

  // Delete user
  const handleDeleteUser = async (userId: number) => {
    try {
      await userApi.deleteUser(userId);
      message.success('User deleted successfully');
      loadUsers(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Delete user failed:', error);
      
      // Get detailed error information
      let errorDetail = '';
      if (error.response?.data?.error?.message) {
        errorDetail = error.response.data.error.message;
      } else if (error.response?.data?.message) {
        errorDetail = error.response.data.message;
      } else if (error.message) {
        errorDetail = error.message;
      }
      
      // Show error modal
      modal.error({
        title: 'Failed to Delete User',
        content: errorDetail || 'An error occurred while deleting the user. Please try again later or contact the system administrator.',
        okText: 'OK',
        width: 500,
      });
    }
  };

  // Update user status
  const handleUpdateStatus = async (userId: number, status: string) => {
    try {
      await userApi.updateUserStatus(userId, status);
      message.success('Status updated successfully');
      loadUsers(pagination.current, pagination.pageSize, searchText);
    } catch (error: any) {
      console.error('Update status failed:', error);
      
      // Get detailed error information
      let errorDetail = '';
      if (error.response?.data?.error?.message) {
        errorDetail = error.response.data.error.message;
      } else if (error.response?.data?.message) {
        errorDetail = error.response.data.message;
      } else if (error.message) {
        errorDetail = error.message;
      }
      
      // Show error modal
      modal.error({
        title: 'Failed to Update Status',
        content: errorDetail || 'An error occurred while updating user status. Please try again later or contact the system administrator.',
        okText: 'OK',
        width: 500,
      });
    }
  };

  // View user details
  const handleViewUser = (user: User) => {
    setSelectedUser(user);
    setIsViewDrawerVisible(true);
  };

  // Get role color
  const getRoleColor = (role: string) => {
    const option = roleOptions.find(opt => opt.value === role);
    return option?.color || 'default';
  };

  // Get status color
  const getStatusColor = (status: string) => {
    const option = statusOptions.find(opt => opt.value === status);
    return option?.color || 'default';
  };

  // Table column definitions
  const columns = [
    {
      title: 'User',
      dataIndex: 'username',
      key: 'username',
      render: (text: string, record: User) => (
        <Space>
          <Avatar icon={<UserOutlined />} src={record.avatar} />
          <div>
            <div style={{ fontWeight: 'bold' }}>{record.full_name}</div>
            <div style={{ color: '#666', fontSize: '12px' }}>@{text}</div>
          </div>
        </Space>
      ),
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => {
        const option = roleOptions.find(opt => opt.value === role);
        return <Tag color={getRoleColor(role)}>{option?.label || role}</Tag>;
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: User) => {
        const option = statusOptions.find(opt => opt.value === status);
        return (
          <Select
            value={status}
            size="small"
            style={{ width: 100 }}
            onChange={(newStatus) => handleUpdateStatus(record.id, newStatus)}
            disabled={!hasPermission('user_management')}
          >
            {statusOptions.map(opt => (
              <Option key={opt.value} value={opt.value}>
                <Tag color={opt.color} style={{ margin: 0 }}>
                  {opt.label}
                </Tag>
              </Option>
            ))}
          </Select>
        );
      },
    },
    {
      title: 'Last Login',
      dataIndex: 'last_login',
      key: 'last_login',
      render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : 'Never logged in',
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 200,
      render: (text: any, record: User) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handleViewUser(record)}
            size="small"
          >
            View
          </Button>
          <PermissionGuard permission="user_management">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleOpenModal(record)}
              size="small"
            >
              Edit
            </Button>
            <Popconfirm
              title="Are you sure you want to delete this user?"
              onConfirm={() => handleDeleteUser(record.id)}
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
      {contextHolder}
      {/* Page title */}
      <Row justify="space-between" align="middle" style={{ marginBottom: '24px' }}>
        <Col>
          <Title level={2} style={{ margin: 0 }}>User Management</Title>
          <Text type="secondary">Manage system user accounts and permissions</Text>
        </Col>
        <Col>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => loadUsers()}>
              Refresh
            </Button>
            <PermissionGuard permission="user_management">
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => handleOpenModal()}
              >
                Create User
              </Button>
            </PermissionGuard>
            <Button
              danger
              onClick={testModalError}
              style={{ backgroundColor: '#ff4d4f', borderColor: '#ff4d4f', color: 'white' }}
            >
              Test Error Modal
            </Button>
          </Space>
        </Col>
      </Row>

      {/* Search bar */}
      <Card style={{ marginBottom: '24px' }}>
        <Row gutter={16} align="middle">
          <Col flex="auto">
            <Search
              placeholder="Search by username, email, or full name"
              allowClear
              enterButton={<SearchOutlined />}
              size="large"
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
        </Row>
      </Card>

      {/* User table */}
      <Card>
        <Table
          columns={columns}
          dataSource={users}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} items`,
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* Create/Edit user modal */}
      <Modal
        title={editingUser ? 'Edit User' : 'Create User'}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSaveUser}
          style={{ marginTop: '24px' }}
        >
          <Form.Item
            name="username"
            label="Username"
            rules={[
              { required: true, message: 'Please enter username' },
              { min: 3, message: 'Username must be at least 3 characters' },
              { max: 50, message: 'Username cannot exceed 50 characters' },
            ]}
          >
            <Input placeholder="Enter username" disabled={!!editingUser} />
          </Form.Item>

          <Form.Item
            name="email"
            label="Email"
            rules={[
              { required: true, message: 'Please enter email' },
              { type: 'email', message: 'Please enter a valid email address' },
            ]}
          >
            <Input placeholder="Enter email address" />
          </Form.Item>

          <Form.Item
            name="full_name"
            label="Full Name"
            rules={[
              { required: true, message: 'Please enter full name' },
              { max: 100, message: 'Full name cannot exceed 100 characters' },
            ]}
          >
            <Input placeholder="Enter full name" />
          </Form.Item>

          {!editingUser && (
            <Form.Item
              name="password"
              label="Password"
              rules={[
                { required: true, message: 'Please enter password' },
                { min: 6, message: 'Password must be at least 6 characters' },
              ]}
            >
              <Input.Password placeholder="Enter password" />
            </Form.Item>
          )}

          <Form.Item
            name="role"
            label="Role"
            rules={[{ required: true, message: 'Please select a role' }]}
          >
            <Select placeholder="Select role">
              {roleOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Tag color={option.color} style={{ margin: 0 }}>
                    {option.label}
                  </Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="status"
            label="Status"
            rules={[{ required: true, message: 'Please select a status' }]}
          >
            <Select placeholder="Select status">
              {statusOptions.map(option => (
                <Option key={option.value} value={option.value}>
                  <Tag color={option.color} style={{ margin: 0 }}>
                    {option.label}
                  </Tag>
                </Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setIsModalVisible(false)}>
                Cancel
              </Button>
              <Button type="primary" htmlType="submit">
                {editingUser ? 'Update' : 'Create'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* User details drawer */}
      <Drawer
        title="User Details"
        placement="right"
        onClose={() => setIsViewDrawerVisible(false)}
        open={isViewDrawerVisible}
        width={500}
      >
        {selectedUser && (
          <div>
            <div style={{ textAlign: 'center', marginBottom: '24px' }}>
              <Avatar
                size={80}
                icon={<UserOutlined />}
                src={selectedUser.avatar}
                style={{ marginBottom: '16px' }}
              />
              <Title level={4} style={{ margin: 0 }}>
                {selectedUser.full_name}
              </Title>
              <Text type="secondary">@{selectedUser.username}</Text>
            </div>

            <Divider />

            <Descriptions column={1} bordered>
              <Descriptions.Item label="Email">
                {selectedUser.email}
              </Descriptions.Item>
              <Descriptions.Item label="Role">
                <Tag color={getRoleColor(selectedUser.role)}>
                  {roleOptions.find(opt => opt.value === selectedUser.role)?.label || selectedUser.role}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={getStatusColor(selectedUser.status)}>
                  {statusOptions.find(opt => opt.value === selectedUser.status)?.label || selectedUser.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Last Login">
                {selectedUser.last_login
                  ? dayjs(selectedUser.last_login).format('YYYY-MM-DD HH:mm:ss')
                  : 'Never logged in'}
              </Descriptions.Item>
              <Descriptions.Item label="Created At">
                {dayjs(selectedUser.created_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
              <Descriptions.Item label="Updated At">
                {dayjs(selectedUser.updated_at).format('YYYY-MM-DD HH:mm:ss')}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default Users; 