import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { App, Button, Card, Form, Input, Modal, Space, Table } from 'antd';
import { DeleteOutlined, EditOutlined, FolderOutlined, PlusOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import type { Category } from '../../types';

type CategoryValues = Pick<Category, 'name' | 'slug' | 'description'>;

export default function CategoriesPage() {
  useDocumentTitle('分类管理');
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<CategoryValues>();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const categoriesQuery = useQuery({ queryKey: ['admin', 'categories'], queryFn: adminApi.categories });
  useEffect(() => { if (open) form.setFieldsValue(editing || { name: '', slug: '', description: '' }); else form.resetFields(); }, [open, editing, form]);

  const invalidate = async () => Promise.all([
    ...['admin', 'public'].map((scope) => queryClient.invalidateQueries({ queryKey: [scope, 'categories'] })),
    queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] }),
  ]);
  const save = useMutation({
    mutationFn: (values: CategoryValues) => editing ? adminApi.updateCategory(editing.id, values) : adminApi.createCategory(values),
    onSuccess: async () => { message.success(editing ? '分类已更新' : '分类已创建'); setOpen(false); await invalidate(); await queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }); },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const remove = useMutation({
    mutationFn: adminApi.deleteCategory,
    onSuccess: async () => {
      message.success('分类已删除，相关文章已移为未分类');
      await invalidate();
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['admin', 'posts'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'post'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'related'] }),
      ]);
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const edit = (category: Category) => { setEditing(category); setOpen(true); };

  return (
    <div className="admin-page">
      <div className="admin-page-heading"><div><p>ORGANIZE YOUR IDEAS</p><h1>分类管理</h1><span>用清晰的主题边界组织你的内容。</span></div><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>新建分类</Button></div>
      <Card className="admin-table-card">
        <Table<Category> rowKey="id" dataSource={categoriesQuery.data} loading={categoriesQuery.isLoading} pagination={false} scroll={{ x: 680 }} columns={[
          { title: '分类', dataIndex: 'name', render: (name: string, item) => <div className="taxonomy-name"><span><FolderOutlined /></span><div><strong>{name}</strong><small>/{item.slug}</small></div></div> },
          { title: '描述', dataIndex: 'description', render: (text: string) => text || <span className="muted">暂无描述</span> },
          { title: '文章数', dataIndex: 'post_count', width: 100, render: (count: number) => `${count} 篇` },
          { title: '创建时间', dataIndex: 'created_at', width: 130, render: (date: string) => dayjs(date).format('YYYY-MM-DD') },
          { title: '操作', key: 'actions', width: 110, render: (_, item) => <Space><Button type="text" icon={<EditOutlined />} onClick={() => edit(item)} aria-label={`编辑 ${item.name}`} /><Button type="text" danger icon={<DeleteOutlined />} onClick={() => modal.confirm({ title: `删除分类「${item.name}」？`, content: item.post_count ? `该分类下的 ${item.post_count} 篇文章会保留并移为未分类，删除后无法恢复。` : '分类下的文章会保留并移为未分类，删除后无法恢复。', okText: '删除', okButtonProps: { danger: true }, cancelText: '取消', onOk: () => remove.mutateAsync(item.id).then(() => undefined) })} aria-label={`删除 ${item.name}`} /></Space> },
        ]} />
      </Card>
      <Modal title={editing ? '编辑分类' : '新建分类'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} confirmLoading={save.isPending} okText="保存" cancelText="取消" destroyOnHidden>
        <Form form={form} layout="vertical" onFinish={(values) => save.mutate(values)} requiredMark={false} className="admin-modal-form">
          <Form.Item name="name" label="分类名称" rules={[{ required: true, message: '请输入分类名称' }, { max: 50 }]}><Input placeholder="例如：后端工程" /></Form.Item>
          <Form.Item name="slug" label="URL 标识" rules={[{ required: true, message: '请输入 URL 标识' }, { pattern: /^[a-z0-9-]+$/, message: '仅使用小写字母、数字和连字符' }]}><Input placeholder="例如：backend" /></Form.Item>
          <Form.Item name="description" label="分类描述" rules={[{ max: 200 }]}><Input.TextArea rows={4} placeholder="简要介绍这个分类收录的内容" showCount maxLength={200} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
