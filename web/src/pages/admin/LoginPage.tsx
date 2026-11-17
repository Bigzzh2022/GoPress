import { useState } from 'react';
import { Button, Form, Input } from 'antd';
import { ArrowRightOutlined, LockOutlined, MailOutlined, CheckCircleFilled } from '@ant-design/icons';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { authApi } from '../../api/auth';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { useAuthStore } from '../../stores/authStore';

interface LoginValues { email: string; password: string }

export default function LoginPage() {
  useDocumentTitle('后台登录');
  const { accessToken, setAuth } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [form] = Form.useForm<LoginValues>();
  const [error, setError] = useState('');
  const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname || '/admin';
  const login = useMutation({
    mutationFn: ({ email, password }: LoginValues) => authApi.login(email, password),
    onSuccess: ({ access_token, user }) => {
      setAuth(access_token, user);
      navigate(from, { replace: true });
    },
    onError: (reason) => setError(getErrorMessage(reason, '登录失败，请检查邮箱和密码。')),
  });

  if (accessToken) return <Navigate to="/admin" replace />;
  return (
    <main className="login-page">
      <section className="login-story">
        <div className="login-story-brand"><span>G</span>GoPress</div>
        <div><p className="login-eyebrow">A QUIET PLACE FOR LOUD IDEAS</p><h1>让每一次思考，<br />都有清晰的归宿。</h1><p>从灵感草稿到正式发布，<br />在一个专注的空间里完成你的技术写作。</p></div>
        <div className="login-quote"><span>“</span><p>写作不是记录已知，<br />而是发现你真正理解了什么。</p></div>
        <small>GOPRESS CONTENT STUDIO · BUILT FOR WRITERS</small>
      </section>
      <section className="login-form-side">
        <div className="login-form-wrap">
          <div className="login-mobile-brand"><span>G</span>GoPress</div>
          <p className="login-form-kicker">WELCOME BACK</p><h2>欢迎回来</h2><p className="login-subtitle">登录内容工作台，继续你的创作。</p>
          <Form form={form} layout="vertical" size="large" onFinish={(values) => { setError(''); login.mutate(values); }} requiredMark={false}>
            <Form.Item name="email" label="邮箱地址" rules={[{ required: true, message: '请输入邮箱地址' }, { type: 'email', message: '请输入有效的邮箱地址' }]}><Input prefix={<MailOutlined />} placeholder="name@example.com" autoComplete="email" /></Form.Item>
            <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}><Input.Password prefix={<LockOutlined />} placeholder="输入你的密码" autoComplete="current-password" /></Form.Item>
            {error && <div className="login-error" role="alert">{error}</div>}
            <Button type="primary" htmlType="submit" block loading={login.isPending} className="login-submit">登录工作台 <ArrowRightOutlined /></Button>
          </Form>
          {import.meta.env.DEV && <div className="demo-account"><CheckCircleFilled /><div><strong>开发环境</strong><p>账号来自 .env 中的 ADMIN_EMAIL 与 ADMIN_PASSWORD</p></div></div>}
          <p className="login-security"><LockOutlined /> 通过安全 Cookie 保持会话，访问令牌不会写入本地存储。</p>
        </div>
      </section>
    </main>
  );
}
