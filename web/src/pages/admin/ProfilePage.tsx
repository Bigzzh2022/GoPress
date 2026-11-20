import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Avatar, Button, Card, Form, Input, Skeleton } from 'antd';
import { LockOutlined, MailOutlined, SaveOutlined, UserOutlined } from '@ant-design/icons';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { useAuthStore } from '../../stores/authStore';
import type { User } from '../../types';

type ProfileValues = Pick<User, 'name' | 'email' | 'avatar' | 'bio'>;
type PasswordValues = { current_password: string; new_password: string; confirm_password: string };

function passwordIssue(value: string) {
  if (!value || value.length < 12) return '密码至少需要 12 位';
  const classes = [/[a-z]/, /[A-Z]/, /\d/, /[^A-Za-z0-9]/].filter((rule) => rule.test(value)).length;
  if (classes < 3) return '需包含小写字母、大写字母、数字、符号中的至少三类';
  return '';
}

export default function ProfilePage() {
  useDocumentTitle('个人资料');
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [passwordForm] = Form.useForm<PasswordValues>();
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
  const changePassword = useMutation({
    mutationFn: (values: PasswordValues) => adminApi.changePassword({
      current_password: values.current_password,
      new_password: values.new_password,
    }),
    onSuccess: () => {
      message.success('密码已更新，请在下次登录时使用新密码');
      passwordForm.resetFields();
    },
    onError: (error) => {
      const text = getErrorMessage(error);
      message.error(text === 'current password is incorrect' ? '当前密码不正确' : text);
    },
  });

  return (
    <div className="admin-page profile-page">
      <div className="admin-page-heading"><div><p>YOUR PUBLIC IDENTITY</p><h1>个人资料</h1><span>这些信息会展示在博客的作者介绍中。</span></div></div>
      {profileQuery.isLoading ? <Card><Skeleton active avatar paragraph={{ rows: 6 }} /></Card> : (
        <div className="profile-layout">
          <Card className="profile-preview"><Avatar size={88}>{profileQuery.data?.name?.slice(0, 1) || 'G'}</Avatar><h2>{profileQuery.data?.name}</h2><p>{profileQuery.data?.bio || '还没有填写个人简介。'}</p><span><MailOutlined />{profileQuery.data?.email}</span><small>公开作者名片</small></Card>
          <div className="profile-forms">
            <Card className="profile-form-card" title="基本资料">
              <Form<ProfileValues> key={profileQuery.data?.id} layout="vertical" requiredMark={false} initialValues={profileQuery.data} onFinish={(values) => save.mutate(values)}>
                <Form.Item name="name" label="显示名称" rules={[{ required: true, message: '请输入显示名称' }, { max: 50 }]}><Input prefix={<UserOutlined />} placeholder="你的名字" /></Form.Item>
                <Form.Item name="email" label="联系邮箱" rules={[{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]}><Input prefix={<MailOutlined />} placeholder="you@example.com" /></Form.Item>
                <Form.Item name="avatar" label="头像地址（可选）"><Input placeholder="留空时使用文字头像" /></Form.Item>
                <Form.Item name="bio" label="个人简介" rules={[{ max: 300 }]}><Input.TextArea rows={5} maxLength={300} showCount placeholder="介绍你的专业方向、兴趣和正在关注的事情" /></Form.Item>
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={save.isPending}>保存资料</Button>
              </Form>
            </Card>
            <Card className="profile-form-card" title="安全设置">
              <p className="security-note">修改后请使用新密码登录。建议至少 12 位，并包含三类不同字符。</p>
              <Form<PasswordValues> form={passwordForm} layout="vertical" requiredMark={false} onFinish={(values) => changePassword.mutate(values)}>
                <Form.Item name="current_password" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}><Input.Password prefix={<LockOutlined />} autoComplete="current-password" placeholder="输入当前使用的密码" /></Form.Item>
                <Form.Item name="new_password" label="新密码" rules={[{ required: true, message: '请输入新密码' }, { validator: (_, value: string) => (!value || !passwordIssue(value) ? Promise.resolve() : Promise.reject(new Error(passwordIssue(value)))) }]}><Input.Password prefix={<LockOutlined />} autoComplete="new-password" placeholder="至少 12 位，包含至少三类字符" /></Form.Item>
                <Form.Item name="confirm_password" label="确认新密码" dependencies={['new_password']} rules={[{ required: true, message: '请再次输入新密码' }, ({ getFieldValue }) => ({ validator(_, value: string) { if (!value || value === getFieldValue('new_password')) return Promise.resolve(); return Promise.reject(new Error('两次输入的密码不一致')); } })]}><Input.Password prefix={<LockOutlined />} autoComplete="new-password" placeholder="再次输入新密码" /></Form.Item>
                <Button type="primary" htmlType="submit" icon={<LockOutlined />} loading={changePassword.isPending}>更新密码</Button>
              </Form>
            </Card>
          </div>
        </div>
      )}
    </div>
  );
}
