import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Form, Input, Skeleton } from 'antd';
import { GlobalOutlined, MailOutlined, SaveOutlined } from '@ant-design/icons';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { SiteSettings } from '../../types';

export default function SettingsPage() {
  useDocumentTitle('站点设置');
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<SiteSettings>();
  const settingsQuery = useQuery({ queryKey: ['admin', 'settings'], queryFn: adminApi.settings });

  useEffect(() => {
    if (settingsQuery.data) form.setFieldsValue(settingsQuery.data);
  }, [settingsQuery.data, form]);

  const save = useMutation({
    mutationFn: adminApi.updateSettings,
    onSuccess: async (data) => {
      form.setFieldsValue(data);
      queryClient.setQueryData(['admin', 'settings'], data);
      await queryClient.invalidateQueries({ queryKey: ['public', 'settings'] });
      message.success('站点设置已保存');
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  return (
    <div className="admin-page settings-page">
      <div className="admin-page-heading"><div><p>PUBLIC SITE IDENTITY</p><h1>站点设置</h1><span>调整公开站点展示的基础信息，保存后立即生效。</span></div></div>
      {settingsQuery.isLoading ? <Card className="settings-card"><Skeleton active paragraph={{ rows: 7 }} /></Card> : (
        <Card className="settings-card" title="站点信息">
          <p className="settings-note">这些内容会显示在公开站点的页头、首页导语与页脚中。留空时会使用默认文案。</p>
          <Form<SiteSettings> form={form} layout="vertical" requiredMark={false} onFinish={(values) => save.mutate(values)}>
            <div className="settings-grid">
              <Form.Item name="site_name" label="站点名称" rules={[{ required: true, message: '请输入站点名称' }, { max: 50 }]}>
                <Input prefix={<GlobalOutlined />} placeholder="例如：GoPress" />
              </Form.Item>
              <Form.Item name="site_tagline" label="站点副标题" rules={[{ max: 60 }]}>
                <Input placeholder="例如：CODE & CURIOSITY" />
              </Form.Item>
            </div>
            <Form.Item name="site_description" label="站点简介" extra="展示在首页导语与页脚，可用换行分隔句子" rules={[{ max: 300 }]}>
              <Input.TextArea rows={4} maxLength={300} showCount placeholder="用一两句话介绍这个博客关注什么" />
            </Form.Item>
            <div className="settings-grid">
              <Form.Item name="footer_text" label="页脚标语" rules={[{ max: 100 }]}>
                <Input placeholder="例如：BUILT WITH GO · DESIGNED FOR READING" />
              </Form.Item>
              <Form.Item name="contact_email" label="联系邮箱" rules={[{ type: 'email', message: '邮箱格式不正确' }, { max: 191 }]}>
                <Input prefix={<MailOutlined />} placeholder="hello@example.com" />
              </Form.Item>
            </div>
            <div className="settings-actions">
              <span>站点名称将同步更新浏览器标签页标题。</span>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={save.isPending}>保存设置</Button>
            </div>
          </Form>
        </Card>
      )}
    </div>
  );
}
