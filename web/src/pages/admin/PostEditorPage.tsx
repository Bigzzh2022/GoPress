import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, App, Button, Card, DatePicker, Form, Input, Modal, Select, Spin } from 'antd';
import { ArrowLeftOutlined, EyeOutlined, SaveOutlined } from '@ant-design/icons';
import { Link, useNavigate, useParams } from 'react-router-dom';
import dayjs, { type Dayjs } from 'dayjs';
import { adminApi } from '../../api/admin';
import { getErrorMessage } from '../../api/client';
import { useDocumentTitle } from '../../app/hooks';
import { VditorEditor } from '../../components/VditorEditor';
import type { PostPayload, PostStatus } from '../../types';

type PostFormValues = Omit<PostPayload, 'published_at'> & { published_at?: Dayjs | null };

export default function PostEditorPage() {
  const { id } = useParams();
  const postId = id ? Number(id) : null;
  const isEdit = postId !== null && Number.isFinite(postId);
  const navigate = useNavigate();
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<PostFormValues>();
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewHtml, setPreviewHtml] = useState('');
  useDocumentTitle(isEdit ? '编辑文章' : '写新文章');

  const postQuery = useQuery({ queryKey: ['admin', 'post', postId], queryFn: () => adminApi.post(postId!), enabled: isEdit });
  const categoriesQuery = useQuery({ queryKey: ['admin', 'categories'], queryFn: adminApi.categories });
  const tagsQuery = useQuery({ queryKey: ['admin', 'tags'], queryFn: adminApi.tags });
  const save = useMutation({
    mutationFn: (payload: PostPayload) => isEdit ? adminApi.updatePost(postId, payload) : adminApi.createPost(payload),
    onSuccess: async (post) => {
      message.success(isEdit ? '文章已保存' : '文章已创建');
      queryClient.setQueryData(['admin', 'post', post.id], post);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['admin', 'posts'] }),
        queryClient.invalidateQueries({ queryKey: ['admin', 'dashboard'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'posts'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'post'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'related'] }),
        queryClient.invalidateQueries({ queryKey: ['admin', 'categories'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'categories'] }),
        queryClient.invalidateQueries({ queryKey: ['admin', 'tags'] }),
        queryClient.invalidateQueries({ queryKey: ['public', 'tags'] }),
      ]);
      navigate('/admin/posts');
    },
    onError: (error) => message.error(getErrorMessage(error)),
  });
  const preview = useMutation({
    mutationFn: adminApi.preview,
    onSuccess: ({ html }) => { setPreviewHtml(html); setPreviewOpen(true); },
    onError: (error) => message.error(getErrorMessage(error)),
  });

  const submit = (values: PostFormValues) => {
    save.mutate({ ...values, excerpt: values.excerpt || '', cover_image: values.cover_image || '', category_id: values.category_id ?? null, tag_ids: values.tag_ids || [], published_at: values.published_at?.toISOString() || null });
  };
  const saveWithStatus = (status: PostStatus) => { form.setFieldValue('status', status); form.submit(); };
  const initialValues: PostFormValues | undefined = postQuery.data ? {
    title: postQuery.data.title,
    slug: postQuery.data.slug,
    excerpt: postQuery.data.excerpt || '',
    content_md: postQuery.data.content_md || '',
    cover_image: postQuery.data.cover_image || '',
    category_id: postQuery.data.category_id,
    tag_ids: postQuery.data.tags?.map((tag) => tag.id) || [],
    status: postQuery.data.status,
    published_at: postQuery.data.published_at ? dayjs(postQuery.data.published_at) : null,
  } : undefined;

  if (isEdit && postQuery.isLoading) return <div className="editor-loading"><Spin size="large" tip="正在载入文章" /></div>;
  if (isEdit && (postQuery.isError || !postQuery.data)) return <Alert type="error" showIcon message="文章加载失败" description={getErrorMessage(postQuery.error)} action={<Button onClick={() => postQuery.refetch()}>重试</Button>} />;

  return (
    <div className="admin-page post-editor-page">
      <div className="editor-topbar"><div><Link to="/admin/posts" className="editor-back"><ArrowLeftOutlined /> 返回文章管理</Link><h1>{isEdit ? '编辑文章' : '写新文章'}</h1></div><div><Button icon={<EyeOutlined />} loading={preview.isPending} onClick={() => { const content = form.getFieldValue('content_md') as string; if (!content?.trim()) { message.info('请先写一些内容再预览'); return; } preview.mutate(content); }}>预览</Button><Button type="primary" icon={<SaveOutlined />} loading={save.isPending} onClick={() => form.submit()}>保存文章</Button></div></div>
      <Form<PostFormValues> key={postId || 'new'} form={form} layout="vertical" requiredMark={false} initialValues={initialValues || { title: '', slug: '', excerpt: '', content_md: '', cover_image: '', category_id: null, tag_ids: [], status: 'draft', published_at: null }} onFinish={submit} className="post-form">
        <div className="editor-main">
          <Card className="editor-writing-card">
            <Form.Item name="title" rules={[{ required: true, message: '请输入文章标题' }, { max: 120 }]}><Input className="editor-title-input" placeholder="输入一个清晰、准确的标题…" bordered={false} /></Form.Item>
            <Form.Item name="slug" label="文章 URL（可选）" rules={[{ pattern: /^[a-z0-9-]+$/, message: '仅使用小写字母、数字和连字符' }]}><Input addonBefore="/posts/" placeholder="留空由后端自动生成，或输入 my-new-post" /></Form.Item>
            <Form.Item name="content_md" label="正文内容" rules={[{ required: true, message: '请填写文章正文' }, { validator: (_, value: string) => value?.trim() ? Promise.resolve() : Promise.reject(new Error('请填写文章正文')) }]}><VditorEditor /></Form.Item>
          </Card>
        </div>
        <aside className="editor-settings">
          <Card title="发布设置">
            <Form.Item name="status" label="文章状态" rules={[{ required: true }]}><Select<PostStatus> options={[{ value: 'draft', label: '草稿 — 仅后台可见' }, { value: 'published', label: '发布 — 公开可见' }]} /></Form.Item>
            <Form.Item name="published_at" label="发布时间"><DatePicker showTime className="w-full" placeholder="选择发布时间" /></Form.Item>
            <div className="editor-publish-actions"><Button block onClick={() => saveWithStatus('draft')}>存为草稿</Button><Button block type="primary" onClick={() => saveWithStatus('published')}>发布文章</Button></div>
          </Card>
          <Card title="内容组织">
            <Form.Item name="category_id" label="所属分类（可选）"><Select allowClear loading={categoriesQuery.isLoading} placeholder="不选择分类" options={categoriesQuery.data?.map((item) => ({ value: item.id, label: item.name }))} /></Form.Item>
            <Form.Item name="tag_ids" label="文章标签"><Select mode="multiple" loading={tagsQuery.isLoading} placeholder="选择一个或多个标签" optionFilterProp="label" options={tagsQuery.data?.map((item) => ({ value: item.id, label: item.name }))} /></Form.Item>
            <Form.Item name="excerpt" label="文章摘要" rules={[{ max: 300 }]}><Input.TextArea rows={5} maxLength={300} showCount placeholder="用于文章列表和搜索结果，建议 80–120 字" /></Form.Item>
            <Form.Item name="cover_image" label="封面地址（可选）"><Input placeholder="留空时使用内置图形封面" /></Form.Item>
          </Card>
          <div className="editor-tip"><strong>Markdown 小提示</strong><p>支持 GFM 表格、任务列表和代码块。右侧预览可随时检查排版。</p></div>
        </aside>
      </Form>
      <Modal title="文章预览" open={previewOpen} onCancel={() => setPreviewOpen(false)} footer={<Button type="primary" onClick={() => setPreviewOpen(false)}>完成</Button>} width={920} className="preview-modal" destroyOnHidden>
        <iframe title="经过安全隔离的文章 HTML 预览" sandbox="" className="preview-frame" srcDoc={`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><style>body{font:16px/1.8 system-ui,sans-serif;color:#262626;max-width:760px;margin:40px auto;padding:0 24px}pre{overflow:auto;background:#f5f5f5;padding:16px}code{font-family:monospace}img{max-width:100%}table{border-collapse:collapse}td,th{border:1px solid #ddd;padding:8px}</style></head><body>${previewHtml}</body></html>`} />
      </Modal>
    </div>
  );
}
