import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Form, Input, Modal, Space, Table, Tag } from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { Tag as TagType } from '../../types';

type TagValues = Pick<TagType, 'name' | 'slug'>;

export default function TagsPage() {
  useDocumentTitle('标签管理');
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<TagValues>();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<TagType | null>(null);
  const tagsQuery = useQuery({ queryKey: ['admin', 'tags'], queryFn: adminApi.tags });
  useEffect(() => { if (open) form.setFieldsValue(editing || { name: '', slug: '' }); else form.resetFields(); }, [open, editing, form]);

  const invalidate = async () => Promise.all([
    ...['admin', 'public'].map((scope) => queryClient.invalidateQueries({ queryKey: [scope, 'tags'] })),
    queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] }),
  ]);
  const save = useMutation({
    mutationFn: (values: TagValues) => editing ? adminApi.updateTag(editing.id, values) : adminApi.createTag(values),
    onSuccess: async () => { message.success(editing ? '标签已更新' : '标签已创建'); setOpen(false); await invalidate(); await queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const remove = useMutation({
    mutationFn: adminApi.deleteTag,
    onSuccess: async () => { message.success('标签已删除'); await invalidate(); await Promise.all([queryClient.invalidateQueries({ queryKey: ['admin', 'posts'] }), queryClient.invalidateQueries({ queryKey: ['public', 'posts'] })]); },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  return (
    <div className="admin-page">
      <div className="admin-page-heading"><div><p>CONNECT THE DOTS</p><h1>标签管理</h1><span>通过标签建立文章之间更细的连接。</span></div><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>新建标签</Button></div>
      <Card className="admin-table-card">
        <Table<TagType> rowKey="id" dataSource={tagsQuery.data} loading={tagsQuery.isLoading} pagination={false} scroll={{ x: 620 }} columns={[
          { title: '标签', dataIndex: 'name', render: (name: string, item) => <div className="tag-table-name"><Tag color="orange">#{name}</Tag><small>/{item.slug}</small></div> },
          { title: '文章数', dataIndex: 'post_count', width: 120, render: (count: number) => `${count} 篇` },
          { title: '创建时间', dataIndex: 'created_at', width: 160, render: (date: string) => dayjs(date).format('YYYY-MM-DD') },
          { title: '操作', key: 'actions', width: 110, render: (_, item) => <Space><Button type="text" icon={<EditOutlined />} onClick={() => { setEditing(item); setOpen(true); }} aria-label={`编辑 ${item.name}`} /><Button type="text" danger icon={<DeleteOutlined />} onClick={() => modal.confirm({ title: `删除标签「${item.name}」？`, content: item.post_count ? `该标签关联了 ${item.post_count} 篇文章。` : '删除后无法恢复。', okText: '删除', okButtonProps: { danger: true }, cancelText: '取消', onOk: () => remove.mutateAsync(item.id).then(() => undefined) })} aria-label={`删除 ${item.name}`} /></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑标签' : '新建标签'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} confirmLoading={save.isPending} okText="保存" cancelText="取消" destroyOnHidden>
        <Form form={form} layout="vertical" onFinish={(values) => save.mutate(values)} requiredMark={false} className="admin-modal-form">
          <Form.Item name="name" label="标签名称" rules={[{ required: true, message: '请输入标签名称' }, { max: 30 }]}><Input placeholder="例如：分布式系统" /></Form.Item>
          <Form.Item name="slug" label="URL 标识" rules={[{ required: true, message: '请输入 URL 标识' }, { pattern: /^[a-z0-9-]+$/, message: '仅使用小写字母、数字和连字符' }]}><Input placeholder="例如：distributed-systems" /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
