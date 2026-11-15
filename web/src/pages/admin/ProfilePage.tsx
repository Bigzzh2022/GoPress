import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Avatar, Button, Card, Form, Input, Skeleton } from 'antd';
import { MailOutlined, SaveOutlined, UserOutlined } from '@ant-design/icons';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { useAuthStore } from '../../stores/authStore';
import type { User } from '../../types';

type ProfileValues = Pick<User, 'name' | 'email' | 'avatar' | 'bio'>;

export default function ProfilePage() {
  useDocumentTitle('个人资料');
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const setUser = useAuthStore((state) => state.setUser);
  const profileQuery = useQuery({ queryKey: ['admin', 'profile'], queryFn: adminApi.profile });
  const save = useMutation({
    mutationFn: (values: ProfileValues) => adminApi.updateProfile(values),
    onSuccess: (user) => {
      message.success('个人资料已保存');
      setUser(user);
      queryClient.setQueryData(['admin', 'profile'], user);
      queryClient.setQueryData(['public', 'profile'], user);
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  return (
    <div className="admin-page profile-page">
      <div className="admin-page-heading"><div><p>YOUR PUBLIC IDENTITY</p><h1>个人资料</h1><span>这些信息会展示在博客的作者介绍中。</span></div></div>
      {profileQuery.isLoading ? <Card><Skeleton active avatar paragraph={{ rows: 6 }} /></Card> : (
        <div className="profile-layout">
          <Card className="profile-preview"><Avatar size={88}>{profileQuery.data?.name?.slice(0, 1) || 'G'}</Avatar><h2>{profileQuery.data?.name}</h2><p>{profileQuery.data?.bio || '还没有填写个人简介。'}</p><span><MailOutlined />{profileQuery.data?.email}</span><small>公开作者名片</small></Card>
          <Card className="profile-form-card" title="基本资料">
            <Form<ProfileValues> key={profileQuery.data?.id} layout="vertical" requiredMark={false} initialValues={profileQuery.data} onFinish={(values) => save.mutate(values)}>
              <Form.Item name="name" label="显示名称" rules={[{ required: true, message: '请输入显示名称' }, { max: 50 }]}><Input prefix={<UserOutlined />} placeholder="你的名字" /></Form.Item>
              <Form.Item name="email" label="联系邮箱" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}><Input prefix={<MailOutlined />} placeholder="you@example.com" /></Form.Item>
              <Form.Item name="avatar" label="头像地址（可选）"><Input placeholder="留空时使用文字头像" /></Form.Item>
              <Form.Item name="bio" label="个人简介" rules={[{ max: 300 }]}><Input.TextArea rows={5} maxLength={300} showCount placeholder="介绍你的专业方向、兴趣和正在关注的事情" /></Form.Item>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={save.isPending}>保存资料</Button>
            </Form>
          </Card>
        </div>
      )}
    </div>
  );
}
