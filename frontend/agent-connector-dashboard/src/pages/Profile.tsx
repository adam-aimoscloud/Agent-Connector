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

  // 初始化表单数据
  useEffect(() => {
    if (user) {
      form.setFieldsValue({
        username: user.username,
        email: user.email,
        full_name: user.full_name,
      });
    }
  }, [user, form]);

  // 加载登录日志
  const loadLoginLogs = async () => {
    setLogsLoading(true);
    try {
      const response = await authApi.getLoginLogs(1, 10);
      if (response.data.code === 200) {
        setLoginLogs(response.data.data || []);
      } else {
        throw new Error(response.data.message || '获取登录日志失败');
      }
    } catch (error: any) {
      console.error('Failed to load login logs:', error);
      message.error(error.response?.data?.message || '加载登录日志失败');
      setLoginLogs([]);
    } finally {
      setLogsLoading(false);
    }
  };

  useEffect(() => {
    loadLoginLogs();
  }, []);

  // 更新个人信息
  const handleUpdateProfile = async (values: any) => {
    setLoading(true);
    try {
      await updateProfile(values);
      message.success('个人信息更新成功');
    } catch (error) {
      console.error('Update profile failed:', error);
      message.error('个人信息更新失败');
    } finally {
      setLoading(false);
    }
  };

  // 修改密码
  const handleChangePassword = async (values: ChangePasswordRequest) => {
    setPasswordLoading(true);
    try {
      await authApi.changePassword(values);
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error: any) {
      console.error('Change password failed:', error);
      message.error(error.response?.data?.message || '密码修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  // 头像上传处理
  const handleAvatarUpload = (info: any) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功');
      // 这里应该更新用户头像URL
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  // 获取角色显示
  const getRoleDisplay = (role: string) => {
    const roleMap = {
      admin: { label: '管理员', color: 'red' },
      operator: { label: '操作员', color: 'orange' },
      user: { label: '用户', color: 'blue' },
      readonly: { label: '只读用户', color: 'gray' },
    };
    return roleMap[role as keyof typeof roleMap] || { label: role, color: 'default' };
  };

  // 登录日志表格列
  const logColumns = [
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'success' ? 'green' : 'red'}>
          {status === 'success' ? '成功' : '失败'}
        </Tag>
      ),
    },
    {
      title: '浏览器',
      dataIndex: 'user_agent',
      key: 'user_agent',
      render: (userAgent: string) => (
        <Text ellipsis style={{ maxWidth: 300 }} title={userAgent}>
          {userAgent}
        </Text>
      ),
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
  ];

  if (!user) {
    return <div>加载中...</div>;
  }

  return (
    <div>
      <Title level={2}>个人资料</Title>
      
      <Tabs defaultActiveKey="profile">
        <TabPane tab="基本信息" key="profile" icon={<UserOutlined />}>
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
                      <Button icon={<UploadOutlined />}>更换头像</Button>
                    </Upload>
                  </div>
                  <Divider />
                  <div style={{ textAlign: 'left' }}>
                    <p><strong>用户名:</strong> {user.username}</p>
                    <p><strong>邮箱:</strong> {user.email}</p>
                    <p>
                      <strong>角色:</strong> 
                      <Tag color={getRoleDisplay(user.role).color} style={{ marginLeft: '8px' }}>
                        {getRoleDisplay(user.role).label}
                      </Tag>
                    </p>
                    <p><strong>创建时间:</strong> {dayjs(user.created_at).format('YYYY-MM-DD')}</p>
                    <p>
                      <strong>最后登录:</strong> 
                      {user.last_login ? dayjs(user.last_login).format('YYYY-MM-DD HH:mm') : '从未登录'}
                    </p>
                  </div>
                </div>
              </Card>
            </Col>
            
            <Col xs={24} lg={16}>
              <Card title="编辑个人信息" extra={<EditOutlined />}>
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
                        label="用户名"
                        rules={[
                          { required: true, message: '请输入用户名' },
                          { min: 3, message: '用户名至少3个字符' },
                        ]}
                      >
                        <Input placeholder="请输入用户名" disabled />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name="email"
                        label="邮箱"
                        rules={[
                          { required: true, message: '请输入邮箱' },
                          { type: 'email', message: '请输入有效的邮箱地址' },
                        ]}
                      >
                        <Input placeholder="请输入邮箱" />
                      </Form.Item>
                    </Col>
                  </Row>
                  
                  <Form.Item
                    name="full_name"
                    label="姓名"
                    rules={[
                      { required: true, message: '请输入姓名' },
                      { max: 100, message: '姓名最多100个字符' },
                    ]}
                  >
                    <Input placeholder="请输入姓名" />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={loading}
                      icon={<SaveOutlined />}
                    >
                      保存更改
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          </Row>
        </TabPane>
        
        <TabPane tab="修改密码" key="password" icon={<LockOutlined />}>
          <Row justify="center">
            <Col xs={24} md={12}>
              <Card title="修改密码" extra={<LockOutlined />}>
                <Alert
                  message="安全提醒"
                  description="为了账户安全，建议定期更换密码。密码应包含字母、数字，长度至少6位。"
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
                    label="当前密码"
                    rules={[{ required: true, message: '请输入当前密码' }]}
                  >
                    <Input.Password placeholder="请输入当前密码" />
                  </Form.Item>
                  
                  <Form.Item
                    name="new_password"
                    label="新密码"
                    rules={[
                      { required: true, message: '请输入新密码' },
                      { min: 6, message: '密码至少6个字符' },
                    ]}
                  >
                    <Input.Password placeholder="请输入新密码" />
                  </Form.Item>
                  
                  <Form.Item
                    name="confirm_password"
                    label="确认新密码"
                    dependencies={['new_password']}
                    rules={[
                      { required: true, message: '请确认新密码' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('两次输入的密码不一致'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password placeholder="请再次输入新密码" />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      htmlType="submit"
                      loading={passwordLoading}
                      icon={<LockOutlined />}
                      block
                    >
                      修改密码
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          </Row>
        </TabPane>
        
        <TabPane tab="登录日志" key="logs" icon={<HistoryOutlined />}>
          <Card title="登录日志" extra={<HistoryOutlined />}>
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
                  `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
              }}
            />
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default Profile; 