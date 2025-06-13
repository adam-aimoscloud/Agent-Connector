import React, { useState, useEffect } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Typography,
  Row,
  Col,
  Divider,
  message,
  Upload,
  Table,
  Tag,
  Space,
  Tabs,
  Alert,
} from 'antd';
import {
  UserOutlined,
  EditOutlined,
  LockOutlined,
  HistoryOutlined,
  UploadOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { authApi, ChangePasswordRequest } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TabPane } = Tabs;

const Profile: React.FC = () => {
  const { state, updateProfile } = useAuth();
  const { user } = state;
  const [loading, setLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [loginLogs, setLoginLogs] = useState<any[]>([]);
  const [logsLoading, setLogsLoading] = useState(false);
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();

  // Initialize form data
  useEffect(() => {
    if (user) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        full_name: user.full_name,
      });
    }
  }, [user, form]);

  // Load login logs
  const loadLoginLogs = async () => {
    setLogsLoading(true);
    try {
      const response = await authApi.getLoginLogs(1, 10);
      if (response.data.code === 200) {
        setLoginLogs(response.data.data || []);
      } else {
        throw new Error(response.data.message || 'Failed to get login logs');
      }
    } catch (error: any) {
      console.error('Failed to load login logs:', error);
      message.error(error.response?.data?.message || 'Failed to load login logs');
      setLoginLogs([]);
    } finally {
      setLogsLoading(false);
    }
  };

  useEffect(() => {
    loadLoginLogs();
  }, []);

  // Update personal information
  const handleUpdateProfile = async (values: any) => {
    setLoading(true);
    try {
      await updateProfile(values);
      message.success('Personal information updated successfully');
    } catch (error) {
      console.error('Update profile failed:', error);
      message.error('Personal information update failed');
    } finally {
      setLoading(false);
    }
  };

  // Change password
  const handleChangePassword = async (values: ChangePasswordRequest) => {
    setPasswordLoading(true);
    try {
      await authApi.changePassword(values);
      message.success('Password changed successfully');
      passwordForm.resetFields();
    } catch (error: any) {
      console.error('Change password failed:', error);
      message.error(error.response?.data?.message || 'Password change failed');
    } finally {
      setPasswordLoading(false);
    }
  };

  // Avatar upload processing
  const handleAvatarUpload = (info: any) => {
    if (info.file.status === 'done') {
      message.success('Avatar upload successfully');
      // Here should update user avatar URL
    } else if (info.file.status === 'error') {
      message.error('Avatar upload failed');
    }
  };

  // Get role display
  const getRoleDisplay = (role: string) => {
    const roleMap = {
      admin: { label: 'Admin', color: 'red' },
      operator: { label: 'Operator', color: 'orange' },
      user: { label: 'User', color: 'blue' },
      readonly: { label: 'Readonly user', color: 'gray' },
    };
    return roleMap[role as keyof typeof roleMap] || { label: role, color: 'default' };
  };

  // Login log table columns
  const logColumns = [
    {
      title: 'IP address',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>
          {status === 'success' ? 'Success' : 'Failed'}
        </Tag>
      ),
    },
    {
      title: 'Browser',
      dataIndex: 'user_agent',
      key: 'user_agent',
      render: (userAgent: string) => (
        <Text ellipsis style={{ maxWidth: 300 }} title={userAgent}>
          {userAgent}
        </Text>
      ),
    },
    {
      title: 'Time',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
  ];

  if (!user) {
    return <div>Loading...</div>;
  }

  return (
    <div>
      <Title level={2}>Personal information</Title>
      
      <Tabs defaultActiveKey="profile">
        <TabPane tab="Basic information" key="profile" icon={<UserOutlined />}>
          <Row gutter={24}>
            <Col xs={24} lg={8}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  <Avatar size={120} icon={<UserOutlined />} src={user.avatar} />
                  <div style={{ marginTop: '16px' }}>
                    <Upload
                      name="avatar"
                      showUploadList={false}
                      action="/api/v1/upload/avatar"
                      onChange={handleAvatarUpload}
                    >
                      <Button icon={<UploadOutlined />}>Change avatar</Button>
                    </Upload>
                  </div>
                  <Divider />
                  <div style={{ textAlign: 'left' }}>
                    <p><strong>Username:</strong> {user.username}</p>
                    <p><strong>Email:</strong> {user.email}</p>
                    <p>
                      <strong>Role:</strong> 
                      <Tag color={getRoleDisplay(user.role).color} style={{ marginLeft: '8px' }}>
                        {getRoleDisplay(user.role).label}
                      </Tag>
                    </p>
                    <p><strong>Created time:</strong> {dayjs(user.created_at).format('YYYY-MM-DD')}</p>
                    <p>
                      <strong>Last login:</strong> 
                      {user.last_login ? dayjs(user.last_login).format('YYYY-MM-DD HH:mm') : 'Never logged in'}
                    </p>
                  </div>
                </div>
              </Card>
            </Col>
            
            <Col xs={24} lg={16}>
              <Card title="Edit personal information" extra={<EditOutlined />}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleUpdateProfile}
                  autoComplete="off"
                >
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item
                        name="username"
                        label="Username"
                        rules={[
                          { required: true, message: 'Please enter username' },
                          { min: 3, message: 'Username must be at least 3 characters' },
                        ]}
                      >
                        <Input placeholder="Please enter username" disabled />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="email"
                        label="Email"
                        rules={[
                          { required: true, message: 'Please enter email' },
                          { type: 'email', message: 'Please enter a valid email address' },
                        ]}
                      >
                        <Input placeholder="Please enter email" />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Form.Item
                    name="full_name"
                    label="Name"
                    rules={[
                      { required: true, message: 'Please enter name' },
                      { max: 100, message: 'Name must be at most 100 characters' },
                    ]}
                  >
                    <Input placeholder="Please enter name" />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={loading}
                      icon={<SaveOutlined />}
                    >
                      Save changes
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          </Row>
        </TabPane>
        
        <TabPane tab="Change password" key="password" icon={<LockOutlined />}>
          <Row justify="center">
            <Col xs={24} md={12}>
              <Card title="Change password" extra={<LockOutlined />}>
                <Alert
                  message="Security reminder"
                  description="For account security, it is recommended to change the password regularly. The password should contain letters, numbers, and be at least 6 characters long."
                  type="info"
                  showIcon
                  style={{ marginBottom: '24px' }}
                />
                
                <Form
                  form={passwordForm}
                  layout="vertical"
                  onFinish={handleChangePassword}
                  autoComplete="off"
                >
                  <Form.Item
                    name="old_password"
                    label="Current password"
                    rules={[{ required: true, message: 'Please enter current password' }]}
                  >
                    <Input.Password placeholder="Please enter current password" />
                  </Form.Item>
                  
                  <Form.Item
                    name="new_password"
                    label="New password"
                    rules={[
                      { required: true, message: 'Please enter new password' },
                      { min: 6, message: 'Password must be at least 6 characters' },
                    ]}
                  >
                    <Input.Password placeholder="Please enter new password" />
                  </Form.Item>
                  
                  <Form.Item
                    name="confirm_password"
                    label="Confirm new password"
                    dependencies={['new_password']}
                    rules={[
                      { required: true, message: 'Please confirm new password' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('The two passwords entered are inconsistent'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password placeholder="Please enter new password again" />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={passwordLoading}
                      icon={<LockOutlined />}
                      block
                    >
                      Change password
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          </Row>
        </TabPane>
        
        <TabPane tab="Login logs" key="logs" icon={<HistoryOutlined />}>
          <Card title="Login logs" extra={<HistoryOutlined />}>
            <Table
              columns={logColumns}
              dataSource={loginLogs}
              rowKey="id"
              loading={logsLoading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total, range) =>
                  `${range[0]}-${range[1]} of ${total}`,
              }}
            />
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Profile; 